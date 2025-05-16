package compression

import (
	"bytes"
	"compress/gzip"
	"io"
	"log/slog"
)

type CompressionType string

const (
	Gzip CompressionType = "gzip"
)

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
