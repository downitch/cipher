package api

import (
	"fmt"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	random "math/rand"
	"os"
	"strconv"
	"strings"
)

func GenRandomString(n int) string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
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

func trimNullBytes(slice []byte) []byte {
	l := len(slice)
	l -= 1
	for i := l; i >= 0; i-- {
		if slice[i] != 0 {
			return slice[:i+1]
		}
	}
	return []byte{}
}

func stringifySlice(slice []byte) string {
	result := strconv.Itoa(int(slice[0]))
	for i := 1; i < len(slice); i++ {
		result = fmt.Sprintf("%s %d", result, int(slice[i]))
	}
	return result
}

func bytifyString(str string) []byte {
	var result []byte
	slice := strings.Split(str, " ")
	for i := 0; i < len(slice); i++ {
		numbered, _ := strconv.Atoi(slice[i])
		result = appendByte(result, byte(numbered))
	}
	return result
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

func CipherMessage(receiver string, msg string) []byte {
	realPath, err := parsePath("/history/history")
	// if it happened that the route can not be parsed, returns error
	if err != nil {
		return []byte{}
	}
	// now all the encryption works only with byte slices
	bytedMessage := []byte(msg)
	// b represents message
	b := base64.StdEncoding.EncodeToString(bytedMessage)
	randomCipher, number, _ := GetRandomBlock()
	strNumber := strconv.Itoa(number)
	// n represents blockchain's block number
	n := base64.StdEncoding.EncodeToString([]byte(strNumber))
	decodedRandomCipher, _ := hex.DecodeString(randomCipher)
	randomBlock, _ := aes.NewCipher(decodedRandomCipher)
	// parsing our database to get correct cipher from there
	constCipher, _ := parseCurrentCipher(realPath, receiver)
	decodedCipher, _ := hex.DecodeString(constCipher)
	constBlock, _ := aes.NewCipher(decodedCipher)
	// creating variable that will contain encrypted number
	ciphernumber := make([]byte, aes.BlockSize+len(n))
	// initializing IV
	iv := ciphernumber[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return []byte{}
	}
	cfb := cipher.NewCFBEncrypter(constBlock, iv)
	cfb.XORKeyStream(ciphernumber[aes.BlockSize:], []byte(n))
	// creating variable that will contain encrypted message
	ciphertext := make([]byte, aes.BlockSize+len(b))
	// initializing IV
	iv = ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return []byte{}
	}
	cfb = cipher.NewCFBEncrypter(randomBlock, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	// now concat both slices into one long slice separated with three bytes
	result := append(ciphernumber, byte(42), byte(58), byte(42))
	for i := 0; i < len(ciphertext); i++ {
		result = append(result, ciphertext[i])
	}
	// returning encrypted message
	return result
}

func DecipherMessage(receiver string, msg []byte) []byte {
	strMsg := stringifySlice(msg)
	split := strings.Split(strMsg, " 42 58 42 ")
	num := bytifyString(split[0])
	msg = bytifyString(split[1])
	realPath, err := parsePath("/history/history")
	// if it happened that the route can not be parsed, returns error
	if err != nil {
		return []byte{}
	}
	// parsing our database to get correct cipher from there
	constCipher, _ := parseCurrentCipher(realPath, receiver)
	decodedCipher, _ := hex.DecodeString(constCipher)
	block, _ := aes.NewCipher(decodedCipher)
	if len(num) < aes.BlockSize {
		return []byte{}
	}
	iv := num[:aes.BlockSize]
	num = num[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(num, num)
	data, err := base64.StdEncoding.DecodeString(string(num))
	if err != nil {
		return []byte{}
	}
	blockNumber, _ := strconv.Atoi(string(data))
	hash, _ := GetBlockHash(int64(blockNumber))
	decodedCipher, _ = hex.DecodeString(hash)
	block, _ = aes.NewCipher(decodedCipher)
	if len(msg) < aes.BlockSize {
		return []byte{}
	}
	iv = msg[:aes.BlockSize]
	msg = msg[aes.BlockSize:]
	cfb = cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)
	data, err = base64.StdEncoding.DecodeString(string(msg))
	if err != nil {
		return []byte{}
	}
	return data
}
