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

func (c *Commander) AddFileToIPFS(filepath string) string {
	sh := shell.NewShell("https://ipfs.infura.io:5001")
	f, err := os.Open(filepath)
	buf := bufio.NewReader(f)
	cid, err := sh.Add(buf)
	if err != nil {
		return ""
	}
	return cid
}

func (c *Commander) HashFileIPFS(filepath string) string {
	sh := shell.NewShell("https://ipfs.infura.io:5001")
	opts := shell.OnlyHash(true)
	f, err := os.Open(filepath)
	buf := bufio.NewReader(f)
	cid, err := sh.Add(buf, opts)
	if err != nil {
		return ""
	}
	return cid
}

const URLCat = "ipfs.infura.io:5001/api/v0/cat"

func (c *Commander) CatFileFromIPFS(hash string) ([]byte, error) {
	url := fmt.Sprintf("%s?arg=%s", URLCat, hash)
	data, err := RequestHTTPS(url)
	if err != nil {
		return []byte{}, err
	}
	return []byte(data), nil
}