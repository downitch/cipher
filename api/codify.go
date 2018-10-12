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

	"github.com/ethereum/go-ethereum/ethclient"
)

func GenRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func appendByte(slice []byte, data ...byte) []byte {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) {
		newSlice := make([]byte, (n+1)*2)
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

func EncodeMessage(path string, client ethclient.Client, receiver string, msg string) (string, error) {
	realPath, _ := parsePath(path)

	bytedMsg := []byte(msg)

	rest := len(bytedMsg) % 32
	for i := 0; i < (32 - rest); i++ {
		bytedMsg = appendByte(bytedMsg, 0)
	}

	destination := make([]byte, len(bytedMsg))

	latestBlock, _ := GetLatestBlock(client)
	parsedLatestBlock, _ := strconv.Atoi(latestBlock)

	randomBlockNum := rand.Intn(int(parsedLatestBlock - 1))
	randomBlockNum += 1

	stringifiedBlockNum := strconv.Itoa(randomBlockNum)
	bytedRandomBlockNum := []byte(stringifiedBlockNum)

	rest = len(bytedRandomBlockNum) % 32
	for i := 0; i < (32 - rest); i++ {
		bytedRandomBlockNum = appendByte(bytedRandomBlockNum, 0)
	}
	blockDestination := make([]byte, len(bytedRandomBlockNum))

	block, _ := GetBlockHash(client, int64(randomBlockNum))

	cipher, _ := parseCurrentCipher(realPath, receiver)
	decodedCipher, _ := hex.DecodeString(cipher)

	newCipher := strings.Split(block, "x")[1]
	newCipherDecoded, _ := hex.DecodeString(newCipher)

	ekey, _ := aes.NewCipher(newCipherDecoded)
	ekey.Encrypt(destination, bytedMsg)

	nkey, _ := aes.NewCipher(decodedCipher)
	nkey.Encrypt(blockDestination, bytedRandomBlockNum)

	result := string(blockDestination) + "*:*" + string(destination)

	return result, nil
}

// func DecodeMessage(path string, sender string, msg string) (string, error) {
// 	cipher, _ := parseCurrentCipher(path, sender)
// 	return cipher, nil
// }