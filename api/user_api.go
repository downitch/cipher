package api

import(
	"encoding/hex"
	"errors"
	"os"
	"io/ioutil"
	"strings"
)

func GetCallbackLink(address string) string {
	path, err := os.Getwd()
	if err != nil {
		return ""
	}
	data, err := ioutil.ReadFile(path + "/api/history/history")
	lines := strings.Split(string(data), "\n")
	lines = lines[:len(lines)-1]
	for line := range lines {
		step := strings.Split(lines[line], "*:*")[1]
		if address == step {
			return strings.Split(lines[line], "*:*")[0]
		}
	}
	return ""
}

func GetSelfAddress() string {
	path, err := os.Getwd()
	if err != nil {
		return ""
	}
	data, err := ioutil.ReadFile(path + "/api/hs/address")
	address := string(data)
	formattedAddress := strings.Split(address, "\n")[0]
	return formattedAddress
}

func CheckExistance(address string) error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(path + "/api/history/history")
	lines := strings.Split(string(data), "\n")
	lines = lines[:len(lines)-1]
	for line := range lines {
		step := strings.Split(lines[line], "*:*")[1]
		if address == step {
			return errors.New("found user")
		}
	}
	return nil
}

func Hexify(source string) string {
	return hex.EncodeToString([]byte(source))
}

func WriteDownNewUser(cb string, address string, cipher string) error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	fullPath := path + "/api/history/history"
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return errors.New("can't open file to append/writeOnly")
	}
	defer f.Close()
	text := cb + "*:*" + address + "*:*" + cipher + "\n"
	if _, err = f.WriteString(text); err != nil {
		return errors.New("can't add string to file")
	}
	return nil
}