package contentencoder

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

type Encoding string

const (
	GZIP     Encoding = "gzip"
	DEFLATE  Encoding = "deflate"
	BR       Encoding = "br"
	ZSTD     Encoding = "zstd"
	IDENTITY Encoding = "identity"
)

func (enc Encoding) ToExt() (string, bool) {
	switch enc {
	case GZIP, DEFLATE:
		return ".gz", true
	case BR:
		return ".br", true
	case ZSTD:
		return ".zst", true
	case IDENTITY:
		return "", true // No extension
	default:
		return "", false
	}
}

func (enc Encoding) ToStr() string {
	return string(enc)
}

func (enc Encoding) Encode(data []byte) ([]byte, error) {
	switch enc {
	case GZIP:
		return EncodeGzip(data)
	case DEFLATE:
		return EncodeDeflate(data, gzip.BestSpeed)
	case BR:
		return EncodeBrotli(data, 5) // mid-quality default
	case ZSTD:
		return EncodeZstd(data, 3) // reasonable tradeoff
	case IDENTITY:
		return EncodeIdentity(data)
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", enc)
	}
}

func EncodeGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	zw := gzip.NewWriter(&buf)

	_, err := zw.Write(data)

	if err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func EncodeDeflate(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer

	zw, err := flate.NewWriter(&buf, level)

	if err != nil {
		return nil, err
	}

	_, err = zw.Write(data)

	if err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func EncodeBrotli(data []byte, quality int) ([]byte, error) {
	var buf bytes.Buffer
	zw := brotli.NewWriterLevel(&buf, quality)
	_, err := zw.Write(data)
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func EncodeZstd(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer
	enc, err := zstd.NewWriter(&buf, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(level)))
	if err != nil {
		return nil, err
	}
	_, err = enc.Write(data)
	if err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func EncodeIdentity(data []byte) ([]byte, error) {
	return data, nil
}
