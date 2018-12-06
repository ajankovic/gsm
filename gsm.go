// Package gsm provides transformers for encoding/decoding GSM
// character set into/from UTF-8.
// It relies on interfaces defined by golang.org/x/text/transform
//
// More details can be found here https://godoc.org/golang.org/x/text/transform#Transformer
package gsm

import (
	"unicode/utf8"

	"golang.org/x/text/transform"
)

// Decoder implements transform.Transformer interface which
// transforms bytes from GSM to UTF-8 encoding.
type Decoder struct {
}

// NewDecoder creates new GSM decoder.
func NewDecoder() *Decoder {
	return &Decoder{}
}

// Reset implements transform.Transformer interface.
func (u Decoder) Reset() {}

// Transform implements transform.Transformer interface.
func (u Decoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for i := 0; i < len(src); i++ {
		c := src[i]
		dec := &decode[c]
		if c == 0x1B && i+1 < len(src) {
			d, ok := escDecode[src[i+1]]
			if ok {
				dec = d
				i++
			}
		}
		n := int(dec.len)
		if nDst+n > len(dst) {
			err = transform.ErrShortDst
			break
		}
		for j := 0; j < n; j++ {
			dst[nDst] = dec.data[j]
			nDst++
		}
		nSrc = i + 1
	}
	return nDst, nSrc, err
}

// Encoder implements transform.Transformer interface which
// transforms UTF-8 bytes into GSM bytes.
// More details here https://godoc.org/golang.org/x/text/transform#Transformer
type Encoder struct {
	replacement byte
}

// NewEncoder creates new GSM encoder.
func NewEncoder(replacement byte) *Encoder {
	if replacement == 0 {
		replacement = byte(0x3F)
	}
	return &Encoder{
		replacement: replacement,
	}
}

// Reset implements transform.Transformer interface.
func (en Encoder) Reset() {}

// Transform implements transform.Transformer interface.
func (en Encoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	r, size := rune(0), 0
trans:
	for nSrc < len(src) {
		if nDst >= len(dst)-1 {
			err = transform.ErrShortDst
			break
		}

		// Decode a multi-byte rune.
		r, size = utf8.DecodeRune(src[nSrc:])
		nSrc += size
		if r == utf8.RuneError {
			dst[nDst] = en.replacement
			nDst++
			continue
		}

		// Binary search in [low, high) for that rune in the encode table.
	search:
		for low, high := 0x00, 0x80; ; {
			if low >= high {
				// If search ended without results check extension table.
				for i := 0; i < len(escEncode); i++ {
					got := rune(escEncode[i] & (1<<16 - 1))
					if got == r {
						if nDst+2 >= len(dst) {
							// Destination doesn't have enough room to receive bytes
							// it should reattempt transformation starting from the
							// problematic bytes so we deduct previously added size.
							nSrc -= size
							err = transform.ErrShortDst
							break trans
						}
						dst[nDst] = byte(escEncode[i] >> 24)
						nDst++
						dst[nDst] = byte((escEncode[i] << 8) >> 24)
						nDst++
						break search
					}
				}
				dst[nDst] = en.replacement
				nDst++
				break
			}
			mid := (low + high) / 2
			got := encode[mid]
			gotRune := rune(got & (1<<16 - 1))
			if gotRune < r {
				low = mid + 1
			} else if gotRune > r {
				high = mid
			} else {
				dst[nDst] = byte(got >> 24)
				nDst++
				if dst[nDst-1] == 27 {
					dst[nDst] = 27
					nDst++
				}
				break
			}
		}
	}
	return nDst, nSrc, err
}

// SevenBitPacker is used for transforming 8-bit character packing
// into 7-bit character packing.
type SevenBitPacker struct {
	s uint32
}

// NewPacker creates new 7-bit packer.
func NewPacker() *SevenBitPacker {
	return &SevenBitPacker{}
}

// Reset implements transform.Transformer interface.
func (p *SevenBitPacker) Reset() {
	p.s = 1
}

// Transform implements transform.Transformer interface.
func (p *SevenBitPacker) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	i := 1
	for nDst = 0; i <= len(src); nDst++ {
		// Clear high bits and remove borrowed bits.
		if nDst >= len(dst) {
			err = transform.ErrShortDst
			return
		}
		dst[nDst] = src[i-1] >> (p.s - 1)
		next := byte(0)
		if i < len(src) {
			next = src[i] << (8 - p.s)
		}
		// Slap borrowed high bits with low bits.
		dst[nDst] |= next
		p.s++
		i++
		nSrc++
		if p.s == 8 {
			p.s = 1
			i++
			nSrc++
		}
	}
	return nDst, nSrc, err
}

// NewUnpacker creates new SevenBitUnpacker.
func NewUnpacker() transform.Transformer {
	return &SevenBitUnpacker{}
}

// SevenBitUnpacker is used for transforming 7-bit character packing
// into 8-bit character packing.
type SevenBitUnpacker struct {
	borrowed byte
	s        uint32
}

// Reset implements transform.Transformer interface.
func (u *SevenBitUnpacker) Reset() {
	u.borrowed = 0
	u.s = 1
}

// Transform implements transform.Transformer interface.
func (u *SevenBitUnpacker) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	// Handle repeated call to Transform and use borrowed value.
	if u.s == 8 {
		dst[nDst] = u.borrowed
		u.borrowed = 0
		u.s = 1
		nDst++
	}
	for j := 0; j < len(src); j++ {
		if nDst >= len(dst) {
			err = transform.ErrShortDst
			return
		}
		// Clear high bits and slap borrowed bits from next byte.
		dst[nDst] = src[j] << u.s
		dst[nDst] >>= 1
		dst[nDst] |= u.borrowed
		u.borrowed = src[j] >> (8 - u.s)
		u.s++
		nDst++
		nSrc++
		if u.s == 8 {
			dst[nDst] = u.borrowed
			u.borrowed = 0
			u.s = 1
			nDst++
		}
	}

	return nDst, nSrc, err
}
