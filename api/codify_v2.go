package api

import (
	"crypto/aes"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

func GenRandomString(n int) string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func appendByte(slice []byte, data ...byte) []byte {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) {
		newSlice := make([]byte, (n + 1) * 2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

func parsePath(path string) (string, error) {
	fullPath, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return fullPath + "/api" + path, nil
}

func parseCurrentCipher(path string, receiver string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.New("can't parse file")
	}
	for _, line := range strings.Split(string(data), "\n") {
		split := strings.Split(line, "*:*")
		if split[1] == receiver {
			return split[2], nil
		}
	}
	return "", errors.New("receiver not found")
}

func CipherMessage(receiver string, msg string) (string, error) {
	// in order to have quick navigation we parse current path and concatenate it
	// with the directory needed (between both, it adds /api/ since its subdir)
	realPath, err := parsePath("/history/history")
	// if it happened that the route can not be parsed, returns error
	if err != nil {
		return "", err
	}
	// after obtaining realPath we byte the message
	bytedMsg := []byte(msg)
	// byted message may vary slice length, so we check if its length is devidable
	// by 32. we later add empty bytes to the slice in order to be sure
	rest := len(bytedMsg) % 32
	for i := 0; i < (32 - rest); i++ {
		bytedMsg = appendByte(bytedMsg, 0)
	}
	// after bytedMsg have correct length we create one more slice for cipher
	destination := make([]byte, len(bytedMsg))
	// obtaining random block and its number by order
	block, random, _ := GetRandomBlock()
	// converting block's number to string and to byte slice
	stringifiedBlockNum := strconv.Itoa(random)
	bytedRandomBlockNum := []byte(stringifiedBlockNum)
	// byted block number may vary slice length so we check if its length is
	// devidable by 32. we later add empty bytes to the slice in order to be sure
	rest = len(bytedRandomBlockNum) % 32
	for i := 0; i < (32 - rest); i++ {
		bytedRandomBlockNum = appendByte(bytedRandomBlockNum, 0)
	}
	// after bytedRandomBlockNum has correct length we create one more slice
	// for cipher
	blockDestination := make([]byte, len(bytedRandomBlockNum))
	// parsing our database to get correct cipher from there
	cipher, _ := parseCurrentCipher(realPath, receiver)
	decodedCipher, _ := hex.DecodeString(cipher)
	// parsing block hash to get correct cipher from there
	newCipher := strings.Split(block, "x")[1]
	newCipherDecoded, _ := hex.DecodeString(newCipher)
	// encrypting with block hash the message
	ekey, _ := aes.NewCipher(newCipherDecoded)
	ekey.Encrypt(destination, bytedMsg)
	// encrypting block number with constant cipher
	nkey, _ := aes.NewCipher(decodedCipher)
	nkey.Encrypt(blockDestination, bytedRandomBlockNum)
	// concat both encoded parts with *:* as a result of the function
	result := string(blockDestination) + "*:*" + string(destination)
	// returning ideally encoded result back
	return result, nil
}

// func DecodeMessage(path string, sender string, msg string) (string, error) {
// 	cipher, _ := parseCurrentCipher(path, sender)
// 	return cipher, nil
// }