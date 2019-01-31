package api

import(
	"fmt"
	"bufio"
	"os"

	shell "github.com/ipfs/go-ipfs-api"
)

type ipfsResponse struct {
	Name string `json:"Name"`
	Hash string `json:"Hash"`
	Size string `json:"Size"`
}

const infura = "https://ipfs.infura.io:5001"
var cat = fmt.Sprintf("%s/api/v0/cat", infura)

func (c *Commander) AddFileToIPFS(filepath string) string {
	sh := shell.NewShell(infura)
	f, err := os.Open(filepath)
	buf := bufio.NewReader(f)
	cid, err := sh.Add(buf)
	if err != nil {
		return ""
	}
	return cid
}

func (c *Commander) HashFileIPFS(filepath string) string {
	sh := shell.NewShell(infura)
	opts := shell.OnlyHash(true)
	f, err := os.Open(filepath)
	buf := bufio.NewReader(f)
	cid, err := sh.Add(buf, opts)
	if err != nil {
		return ""
	}
	return cid
}

func (c *Commander) CatFileFromIPFS(hash string) ([]byte, error) {
	url := fmt.Sprintf("%s?arg=%s", cat, hash)
	data, err := RequestHTTPS(url)
	if err != nil {
		return []byte{}, err
	}
	return []byte(data), nil
}