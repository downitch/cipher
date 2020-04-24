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
	return fmt.Sprintf("%s:%s", filename, buf.String())
}

func (c *Commander) GetFile(decoded string) (string, []byte) {
	split := strings.Split(decoded, ":")
	filename := split[0]
	stringified := split[1]
	return filename, []byte(stringified)
}
