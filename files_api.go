package api

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (c *Commander) PostFile(fpath string) string {
	f, err := os.Open(fpath)
	if err != nil {
		return ""
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(f)
	filename := filepath.Base(fpath)
	result := fmt.Sprintf("%s:%s", filename, buf.String())
	return result
}

func (c *Commander) GetFile(decoded string) []byte {
	split := strings.Split(decoded, ":")
	stringified := split[1]
	return []byte(stringified)
}

func (c *Commander) GetFilename(decoded string) string {
	split := strings.Split(decoded, ":")
	filename := split[0]
	return filename
}
