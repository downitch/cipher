package api

import(
	"encoding/json"
	"fmt"
	"bytes"
	"os"
	"io"
	"mime/multipart"
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

func (c *Commander) AddFileToIPFS(filename string) (string, error) {
	fullPath := c.ConstantPath + "/" + filename
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return "", err
	}
	fh, err := os.Open(fullPath)
	if err != nil {
		fmt.Println("error opening file")
		return "", err
	}
	defer fh.Close()
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return "", err
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	data, err := RequestPostHTTPS(URLAdd, contentType, bodyBuf)
	if err != nil {
		return "", err
	}
	response := &ipfsResponse{}
	err = json.Unmarshal([]byte(data), response)
	if err != nil {
		return "", err
	}
	return response.Hash, nil
}