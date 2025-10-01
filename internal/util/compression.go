package util

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"log/slog"
)

type CompressionType string

const (
	Gzip CompressionType = "gzip"
	None CompressionType = "none"
)

// CompressBytes compresses the input data and returns compressed bytes
func CompressBytes(data []byte, ct CompressionType) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	return Compress(data, ct)
}

// DecompressBytes decompresses the byte data
func DecompressBytes(data []byte, ct CompressionType) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	return Decompress(data, ct)
}

// CompressAndEncode compresses the input data and encodes it to base64
func CompressAndEncode(data []byte, ct CompressionType) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	compressed, err := Compress(data, ct)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(compressed), nil
}

// DecodeAndDecompress decodes base64 string and decompresses the data
func DecodeAndDecompress(data string, ct CompressionType) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	return Decompress(decoded, ct)
}

// Compress compresses the input data using the specified compression type
func Compress(data []byte, ct CompressionType) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	switch ct {
	case Gzip:
		return compressGzip(data)
	default:
		return data, nil
	}
}

// Decompress decompresses the input data using the specified compression type
func Decompress(data []byte, ct CompressionType) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	switch ct {
	case Gzip:
		return decompressGzip(data)
	default:
		return data, nil
	}
}

func compressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decompressGzip(data []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func(zr *gzip.Reader) {
		err := zr.Close()
		if err != nil {
			slog.Error("Error closing gzip reader", "errs", err)
		}
	}(zr)
	return io.ReadAll(zr)
}
