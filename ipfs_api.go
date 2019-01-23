package api

import(
	"encoding/json"
	"fmt"
	"bytes"
	"os"
	"io"
	"mime/multipart"
	"strings"
)

type ipfsResponse struct {
	Name string `json:"Name"`
	Hash string `json:"Hash"`
	Size string `json:"Size"`
}

const URLCat = "ipfs.infura.io:5001/api/v0/cat"
const URLAdd = "ipfs.infura.io:5001/api/v0/add?pin=false"

func (c *Commander) CatFileFromIPFS(hash string) ([]byte, error) {
	url := fmt.Sprintf("%s?arg=%s", URLCat, hash)
	data, err := RequestHTTPS(url)
	if err != nil {
		return []byte{}, err
	}
	return []byte(data), nil
}

func (c *Commander) AddFileToIPFS(filepath string) string {
	split := strings.Split(filepath, "/")
	filename := split[len(split) - 1]
	split = nil
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Println("GOLANG")
		fmt.Println("error writing to buffer")
		fmt.Println(err)
		return ""
	}
	fh, err := os.Open(filepath)
	if err != nil {
		fmt.Println("GOLANG")
		fmt.Println("error opening file")
		fmt.Println(err)
		return ""
	}
	defer fh.Close()
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		fmt.Println("GOLANG")
		fmt.Println(err)
		return ""
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	// this operation might be necessary later on
	// os.Remove(filepath)
	data, err := RequestPostHTTPS(URLAdd, contentType, bodyBuf)
	if err != nil {
		fmt.Println("GOLANG")
		fmt.Println(err)
		return ""
	}
	response := &ipfsResponse{}
	err = json.Unmarshal([]byte(data), response)
	if err != nil {
		fmt.Println("GOLANG")
		fmt.Println(err)
		return ""
	}
	return response.Hash
}