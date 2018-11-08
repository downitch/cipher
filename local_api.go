package api

import(
	"fmt"
	"bufio"
	"errors"
	"os"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type NewMessage struct {
	Type     string `json:"type"`
	Date     int		`json:"date"`
	Text     string `json:"text"`
	Author   string `json:"author"`
	Status   string `json:"status"`
}

type Chat struct {
	Username 		string 		 `json:"username"`
	Address 		string 		 `json:"address"`
	LastMessage NewMessage `json:"lastMessage"`
	NewMessages string 		 `json:"newMessages"`
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

func (c *Commander) GetChats() []Chat {
	var chats []Chat
	path := c.ConstantPath
	fullPath := path + "/history/history"
	file, err := os.Open(fullPath)
	if err != nil {
		return []Chat{}
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		txt := scanner.Text()
		split := strings.Split(txt, "*:*")
		username := strings.Split(split[0], ".")[0]
		address := split[1]
		msgs, _ := c.GetMessages(address, []int{0,0})
		ppath := c.ConstantPath + "/" + address
		lastMessage := NewMessage{}
		if len(msgs) > 0 {
			lastMessage = msgs[0]
		}
		nm := c.GetSelfStatusMessages(address)
		ch := Chat{username, address, lastMessage, nm}
		chats = append(chats, ch)
	}
	if err := scanner.Err(); err != nil {
		return []Chat{}
	}
	return chats
}

func (c *Commander) GetMessages(addr string, pos []int) ([]NewMessage, error) {
	var messages []NewMessage
	var position int
	path := c.ConstantPath
	fullPath := path + "/history/" + addr
	file, err := os.Open(fullPath)
	if err != nil {
	  return []NewMessage{}, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if pos[0] == pos[1] && pos[0] == 0 {
			txt := scanner.Text()
			split := strings.Split(txt, "*:*")
			t := split[0]
			date, _ := strconv.Atoi(split[1])
			author := split[2]
			status := split[3]
			text := split[4]
			messages = append(messages, NewMessage{t, date, text, author, status})
		} else if position >= pos[1] && (position - pos[1]) < pos[0] {
			txt := scanner.Text()
			split := strings.Split(txt, "*:*")
			t := split[0]
			date, _ := strconv.Atoi(split[1])
			author := split[2]
			status := split[3]
			text := split[4]
			messages = append(messages, NewMessage{t, date, text, author, status})
		}
		position = position + 1
	}
	if err := scanner.Err(); err != nil {
		return []NewMessage{}, err
	}
	for i := len(messages) / 2 - 1; i >= 0; i-- {
		opp := len(messages) - 1 - i
		messages[i], messages[opp] = messages[opp], messages[i]
	}
	return messages, nil
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

func (c *Commander) SaveMessage(message string, address string, author string) bool {
	path := c.ConstantPath
	fullPath := path + "/history/" + address
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer f.Close()
	defer func() {
		fmt.Println("back to 888")
		os.Chtimes(fullPath, time.Unix(888, 0), time.Unix(888, 0))
	}()
	os.Chtimes(fullPath, time.Unix(999, 0), time.Unix(999, 0))
	currentTime := strconv.Itoa(int(time.Now().UnixNano() / 1000000))
	text := "text*:*" + currentTime + "*:*" + author + "*:*self*:*" + message + "\n"
	if c.GetSelfAddress() == author {
		text = "text*:*" + currentTime + "*:*" + author + "*:*sent*:*" + message + "\n"
	}
	if _, err = f.WriteString(text); err != nil {
		fmt.Println(err)
		// f.Close()
		// os.Chtimes(fullPath, time.Unix(888, 0), time.Unix(888, 0))
		return false
	}
	// f.Close()
	// os.Chtimes(fullPath, time.Unix(888, 0), time.Unix(888, 0))
	return true
}

func (c *Commander) GetSelfStatusMessages(address string) string {
	amount := 0
	path := c.ConstantPath
	fullPath := path + "/history/" + address
	input, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return strconv.Itoa(amount)
	}
	lines := strings.Split(string(input), "\n")
	for _, line := range lines {
		if strings.Contains(line, "*:*self*:*") {
			amount = amount + 1
		}
	}
	return strconv.Itoa(amount)
}

func (c *Commander) updateFolderChtimes() bool {
	var files []string
	root := c.ConstantPath + "/history"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		return false
	}
	for _, file := range files {
		splt := strings.Split(file, "/")
		if splt[len(splt) - 1][:2] == "0x" {
			os.Chtimes(file, time.Unix(888, 0), time.Unix(888, 0))
			f, _ := os.Stat(file)
			fmt.Println(f.ModTime().UnixNano() / 1000000000)
		}
	}
	return true
}

func (c *Commander) UpdateLocalMessageStatus(address string) bool {
	path := c.ConstantPath
	fullPath := path + "/history/" + address
	f, _ := os.Stat(fullPath)
	modificationTime := f.ModTime()
	tUnix := modificationTime.UnixNano() / 1000000000
	if tUnix != 888 {
		fmt.Println("LOCAL FILE BUSY")
		return false
	}
	os.Chtimes(fullPath, time.Unix(999, 0), time.Unix(999, 0))
	input, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return false
	}
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, "*:*self*:*") {
			split := strings.Split(line, "*:*")
			split[3] = "down"
			l := strings.Join(split[:], "*:*")
			lines[i] = l
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(fullPath, []byte(output), 0666)
	if err != nil {
		return false
	}
	os.Chtimes(fullPath, time.Unix(888, 0), time.Unix(888, 0))
	return true
}

func (c *Commander) UpdateMessageStatus(address string) bool {
	path := c.ConstantPath
	fullPath := path + "/history/" + address
	f, _ := os.Stat(fullPath)
	modificationTime := f.ModTime()
	tUnix := modificationTime.UnixNano() / 1000000000
	if tUnix != 888 {
		fmt.Println("REMOTE FILE BUSY")
		return false
	}
	os.Chtimes(fullPath, time.Unix(999, 0), time.Unix(999, 0))
	input, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return false
	}
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, "*:*sent*:*") {
			split := strings.Split(line, "*:*")
			split[3] = "read"
			l := strings.Join(split[:], "*:*")
			lines[i] = l
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(fullPath, []byte(output), 0666)
	if err != nil {
		return false
	}
	os.Chtimes(fullPath, time.Unix(888, 0), time.Unix(888, 0))
	return true
}

func (c *Commander) WriteDownNewUser(cb string, address string, cipher string) error {
	path := c.ConstantPath
	fullPath := path + "/history/" + address
	f, err := os.Create(fullPath)
	if err != nil {
		return errors.New("can't create history file")
	}
	f.Close()
	os.Chtimes(fullPath, time.Unix(888, 0), time.Unix(888, 0))
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