package api

import(
	"fmt"
	"bufio"
	"encoding/hex"
	"errors"
	"os"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	Date int				`json:"date"`
	Text string 		`json:"text"`
	Author string 	`json:"author"`
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

func (c *Commander) GetCallbackLink(address string) string {
	path := c.ConstantPath
	data, _ := ioutil.ReadFile(path + "/history/history")
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

func (c *Commander) GetAddressByLink(link string) string {
	path := c.ConstantPath
	data, _ := ioutil.ReadFile(path + "/history/history")
	lines := strings.Split(string(data), "\n")
	lines = lines[:len(lines)-1]
	for line := range lines {
		step := strings.Split(lines[line], "*:*")[0]
		if link == step {
			return strings.Split(lines[line], "*:*")[1]
		}
	}
	return ""
}

func (c *Commander) GetSelfAddress() string {
	path := c.ConstantPath
	data, _ := ioutil.ReadFile(path + "/hs/address")
	address := string(data)
	formattedAddress := strings.Split(address, "\n")[0]
	return formattedAddress
}

func (c *Commander) CheckExistance(link string) error {
	path := c.ConstantPath
	data, _ := ioutil.ReadFile(path + "/history/history")
	lines := strings.Split(string(data), "\n")
	lines = lines[:len(lines)-1]
	for line := range lines {
		step := strings.Split(lines[line], "*:*")[0]
		if link == step {
			return errors.New("found user")
		}
	}
	return nil
}

func Hexify(source string) string {
	return hex.EncodeToString([]byte(source))
}

func (c *Commander) SaveMessage(message string, address string) bool {
	path := c.ConstantPath
	fullPath := path + "/history/" + address
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return false
	}
	defer f.Close()
	currentTime := strconv.Itoa(int(time.Now().UnixNano() / 1000000))
	text := currentTime + "*:*" + address + "*:*" + message + "\n"
	if _, err = f.WriteString(text); err != nil {
		return false
	}
	return true
}

func (c *Commander) GetMessages(addr string, pos []int) ([]Message, error) {
	var messages []Message
	var position int
	path := c.ConstantPath
	fullPath := path + "/history/" + addr
	file, err := os.Open(fullPath)
	if err != nil {
		fmt.Println(err)
	  return []Message{}, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if position >= pos[1] && (position - pos[1]) < pos[0] {
			txt := scanner.Text()
			split := strings.Split(txt, "*:*")
			date, _ := strconv.Atoi(split[0])
			author := split[1]
			text := split[2]
			messages = append(messages, Message{date, text, author})
		}
		position = position + 1
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return []Message{}, err
	}
	return messages, nil
}

func (c *Commander) WriteDownNewUser(cb string, address string, cipher string) error {
	path := c.ConstantPath
	fullPath := path + "/history/" + address
	f, err := os.Create(fullPath)
	if err != nil {
		return errors.New("can't create history file")
	}
	f.Close()
	fullPath = path + "/history/history"
	f, err = os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY, 0600)
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