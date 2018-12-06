# GSM 03.38 Character Encoding

[![GoDoc Badge]][GoDoc] [![GoReportCard Badge]][GoReportCard] [![Build Status](https://travis-ci.com/ajankovic/gsm.svg?branch=master)](https://travis-ci.com/ajankovic/gsm)

This module provides transformers for encoding/decoding GSM character sets into/from UTF-8. It relies on interfaces defined by golang.org/x/text/transform package.

More details about the interfaces can be found [here](https://godoc.org/golang.org/x/text/transform#Transformer).

Character set mapping table was taken from [here](http://www.unicode.org/Public/MAPPINGS/ETSI/GSM0338.TXT).

[GoDoc]: https://godoc.org/github.com/ajankovic/gsm
[GoDoc Badge]: https://godoc.org/github.com/ajankovic/gsm?status.svg
[GoReportCard]: https://goreportcard.com/report/github.com/ajankovic/gsm
[GoReportCard Badge]: https://goreportcard.com/badge/github.com/ajankovic/gsm