package compress

import (
	"bytes"
	"io"

	"github.com/andybalholm/brotli"
)

// CompressBrotli compresses data using Brotli
func CompressBrotli(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := brotli.NewWriterLevel(&buf, brotli.DefaultCompression)
	
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// DecompressBrotli decompresses Brotli-compressed data
func DecompressBrotli(data []byte) ([]byte, error) {
	reader := brotli.NewReader(bytes.NewReader(data))
	return io.ReadAll(reader)
}
