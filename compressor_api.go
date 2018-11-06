package api

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	"github.com/pierrec/lz4"
)

func CompressSmallData(source string) []byte {
	data := []byte(source)
	buf := &bytes.Buffer{}
	gw := gzip.NewWriter(buf)
	gw.Write(data)
	gw.Close()
	return buf.Bytes()
}

func DecompressSmallData(source []byte) string {
	var buf bytes.Buffer
	gr, _ := gzip.NewReader(bytes.NewBuffer(source))
	defer gr.Close()
	source, _ = ioutil.ReadAll(gr)
	buf.Write(source)
	return buf.String()
}

func CompressBigData(ssource string) []byte {
	source := []byte(ssource)
	compressed := make([]byte, len(source))
	_, err := lz4.CompressBlockHC(source, compressed, 0)
	if err != nil {
		return source
	}
	compressed, err = TrimNullBytes(compressed)
	if err != nil {
		return source
	}
	return compressed
}

func DecompressBigData(source []byte) string {
	decompressed := make([]byte, len(source) * 10)
	_, err := lz4.UncompressBlock(source, decompressed)
	if err != nil {
		return string(source)
	}
	decompressedTrimmed, err := TrimNullBytes(decompressed)
	if err != nil {
		return string(decompressed)
	}
	return string(decompressedTrimmed)
}