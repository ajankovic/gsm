package gsm

import (
	"bytes"

	"golang.org/x/text/transform"

	"testing"
)

var (
	packer   = NewPacker()
	unpacker = NewUnpacker()

	packingTable = []struct {
		unpacked []byte
		packed   []byte
	}{
		{
			unpacked: []byte{0x31, 0x32},
			packed:   []byte{0x31, 0x19},
		},
		{
			unpacked: []byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39},
			packed:   []byte{0x31, 0xD9, 0x8C, 0x56, 0xB3, 0xDD, 0x70, 0x39},
		},
		{
			unpacked: []byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38},
			packed:   []byte{0x31, 0xD9, 0x8C, 0x56, 0xB3, 0xDD, 0x70, 0x31, 0xD9, 0x8C, 0x56, 0xB3, 0xDD, 0x70, 0x31, 0xD9, 0x8C, 0x56, 0xB3, 0xDD, 0x70},
		},
	}
)

func TestPacking7bit(t *testing.T) {
	for i := 0; i < len(packingTable); i++ {
		res, _, err := transform.Bytes(packer, packingTable[i].unpacked)
		if err != nil {
			t.Errorf("packing bytes %v", err)
		}
		if !bytes.Equal(res, packingTable[i].packed) {
			t.Errorf("got %X expected %X", res, packingTable[i].packed)
		}
	}
}

func TestUnpacking7bit(t *testing.T) {
	for i := 0; i < len(packingTable); i++ {
		res, _, err := transform.Bytes(unpacker, packingTable[i].packed)
		if err != nil {
			t.Errorf("unpacking bytes %v", err)
		}
		if !bytes.Equal(res, packingTable[i].unpacked) {
			t.Errorf("got %X expected %X", res, packingTable[i].unpacked)
		}
	}
}

var (
	encoder = NewEncoder(0)
	decoder = NewDecoder()

	transTable = []struct {
		utf8 []byte
		gsm  []byte
	}{
		{
			// aSkIi
			utf8: []byte{0x61, 0x53, 0x6B, 0x49, 0x69},
			gsm:  []byte{0x61, 0x53, 0x6B, 0x49, 0x69},
		},
		{
			// "non-standard" and multibyte characters
			utf8: []byte{0xC3, 0xA8, 0xC3, 0xA9, 0xC3, 0x98, 0xCE, 0xA6},
			gsm:  []byte{0x04, 0x05, 0x0B, 0x12},
		},
		{
			// extension tables
			utf8: []byte{0x0C, 0x5E, 0xE2, 0x82, 0xAC},
			gsm:  []byte{0x1B, 0x0A, 0x1B, 0x14, 0x1B, 0x65},
		},
	}
)

func TestGsmEncode(t *testing.T) {
	for i := 0; i < len(transTable); i++ {
		res, _, err := transform.Bytes(encoder, transTable[i].utf8)
		if err != nil {
			t.Errorf("encoding bytes %v", err)
		}
		if !bytes.Equal(res, transTable[i].gsm) {
			t.Errorf("got %X expected %X", res, transTable[i].gsm)
		}
	}
}

func TestGsmDecode(t *testing.T) {
	for i := 0; i < len(transTable); i++ {
		res, _, err := transform.Bytes(decoder, transTable[i].gsm)
		if err != nil {
			t.Errorf("decoding bytes %v", err)
		}
		if !bytes.Equal(res, transTable[i].utf8) {
			t.Errorf("got %X %q expected %X", res, res, transTable[i].utf8)
		}
	}
}
