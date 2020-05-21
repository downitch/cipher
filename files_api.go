package api

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func (c *Commander) PostFile(fpath string) []byte {
	content, err := ioutil.ReadFile(fpath)
	if err != nil {
		fmt.Println(err)
	}
	return content
}

func (c *Commander) PostFilename(fpath string) string {
	return filepath.Base(fpath)
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
