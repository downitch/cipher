package api

import(
	"fmt"
	"bytes"
	"encoding/base64"
	"errors"
	"os"
	"io/ioutil"
	"strings"
	"image/png"
	"github.com/jakobvarmose/go-qidenticon"
	qrcode "github.com/skip2/go-qrcode"
)

func GetHexColor(hash string) string {
	hash = strings.Split(hash, "x")[1]
	return hash[:6]
}

func (c *Commander) GenAvatar(link string) string {
	address := c.GetAddressByLink(link)
	input := fmt.Sprintf("%s%s", link, address)
	inputHexed := Hexify(input)
	code := qidenticon.Code(inputHexed)
	size := 200
	settings := qidenticon.DefaultSettings()
	img := qidenticon.Render(code, size, settings)
	var buff bytes.Buffer
	png.Encode(&buff, img)
	enc := base64.StdEncoding.EncodeToString(buff.Bytes())
	return enc
}

func (c *Commander) GenQrCode() string {
	var png []byte
	link := c.GetHSLink()
	addr := c.GetSelfAddress()
	split := strings.Split(link, ".")
	res := split[0]
	input := fmt.Sprintf("%s:%s", res, addr)
	png, err := qrcode.Encode(input, qrcode.Medium, 512)
	if err != nil {
		return ""
	}
	enc := base64.StdEncoding.EncodeToString(png)
	return enc
}

func (c *Commander) GetHSLink() string {
	path := c.ConstantPath
	pathToHostname := path + "/hs/hostname"
	data, _ := ioutil.ReadFile(pathToHostname)
	link := strings.Split(string(data), "\n")[0]
	return link
}

func (c *Commander) GetTCPHSLink() string {
	path := c.ConstantPath
	pathToHostname := path + "/tcp/hostname"
	data, _ := ioutil.ReadFile(pathToHostname)
	link := strings.Split(string(data), "\n")[0]
	return link
}

func (c *Commander) GetSelfAddress() string {
	path := c.ConstantPath
	data, _ := ioutil.ReadFile(path + "/hs/address")
	address := string(data)
	formattedAddress := strings.Split(address, "\n")[0]
	return formattedAddress
}

func (c *Commander) UpdateCurrentAddress(address string) error {
	path := c.ConstantPath
	fullPath := path + "/hs/address"
	data, err := ioutil.ReadFile(fullPath)
	lines := strings.Split(string(data), "\n")
	line := lines[0]
	if address == line {
		return nil
	}
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return errors.New("can't open file to append/writeOnly")
	}
	defer f.Close()
	if _, err = f.WriteString(address + "\n"); err != nil {
		return errors.New("can't add string to file")
	}
	return nil
}