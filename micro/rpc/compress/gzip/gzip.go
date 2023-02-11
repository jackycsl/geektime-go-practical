package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
)

type Compressor struct{}

func (c *Compressor) Code() byte {
	return 1
}

func (c *Compressor) Compress(data []byte) ([]byte, error) {
	res := &bytes.Buffer{}
	gw := gzip.NewWriter(res)
	_, err := gw.Write(data)
	if err != nil {
		return nil, err
	}
	// Manual close
	if err = gw.Close(); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func (c *Compressor) Uncompress(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	return io.ReadAll(gr)
}
