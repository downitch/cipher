package api

import(
	"fmt"
	"database/sql"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type NewMessage struct {
	Type     string `json:"type"`
	Date     string `json:"date"`
	Status   string `json:"status"`
	Sender   string `json:"author"`
	Text     string `json:"text"`
}

type NewUser struct {
	Link string
	Addr string
	Hash string
}

type User struct {
	Username 		 string     `json:"username"`
	Link 		 		 string     `json:"link"`
	Addr 		 		 string     `json:"addr"`
	LastMessage  NewMessage `json:"lastMessage"`
	NewMessages  string     `json:"newMessages"`
}

func (c *Commander) UpdateStorage() bool {
	db, err := c.openDB("history")
	if err != nil {
		return false
	}
	defer closeDB(db)
	stmnt := `create table if not exists knownUsers(
	username text,
	link text,
	address text,
	hash text);`
	_, err = db.Exec(stmnt)
	if err != nil {
		return false
	}
	return true
}

func (c *Commander) openDB(name string) (*sql.DB, error) {
	path := c.ConstantPath
	fullPath := fmt.Sprintf("%s/history/%s.db", path, name)
	db, err := sql.Open("sqlite3", fullPath)
	if err != nil {
		return &sql.DB{}, err
	}
	return db, nil
}

func closeDB(db *sql.DB) bool {
	db.Close()
	return true
}

func (c *Commander) GetLinkByAddress(address string) string {
	var link string
	db, err := c.openDB("history")
	if err != nil {
		return ""
	}
	defer closeDB(db)
	stmnt := fmt.Sprintf(
		`select link from knownUsers where address = %s`, address)
	rows, err := db.Query(stmnt)
	if err != nil {
		return ""
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&link)
		if err != nil {
			return ""
		}
	}
	err = rows.Err()
	if err != nil {
		return ""
	}
	return link
}

func (c *Commander) GetAddressByLink(link string) string {
	var address string
	db, err := c.openDB("history")
	if err != nil {
		return ""
	}
	defer closeDB(db)
	stmnt := fmt.Sprintf(`select address from knownUsers where link = %s`, link)
	rows, err := db.Query(stmnt)
	if err != nil {
		return ""
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&address)
		if err != nil {
			return ""
		}
	}
	err = rows.Err()
	if err != nil {
		return ""
	}
	return address
}

func (c *Commander) CheckExistance(link string) bool {
	amount := 0
	db, err := c.openDB("history")
	if err != nil {
		return true
	}
	defer closeDB(db)
	stmnt := fmt.Sprintf(`select id from knownUsers where link = %s`, link)
	rows, err := db.Query(stmnt)
	if err != nil {
		return true
	}
	defer rows.Close()
	for rows.Next() {
		amount = amount + 1
	}
	err = rows.Err()
	if err != nil {
		return true
	}
	if amount > 0 {
		return true
	}
	return false
}

func (c *Commander) GetChats() []User {
	var users []User
	db, err := c.openDB("history")
	if err != nil {
		return []User{}
	}
	defer closeDB(db)
	stmnt := `select username, link, address from knownUsers;`
	rows, err := db.Query(stmnt)
	if err != nil {
		return []User{}
	}
	defer rows.Close()
	for rows.Next() {
		var username string
		var link string
		var address string
		err = rows.Scan(&username, &link, &address)
		if err != nil {
			return []User{}
		}
		lastMsg, err := c.GetLastMessage(address)
		if err != nil {
			return []User{}
		}
		newMsgs, err := c.GetNewMessages(address)
		if err != nil {
			return []User{}
		}
		newMsgsStringified := strconv.Itoa(newMsgs)
		users = append(
			users, User{
				username,
				link,
				address,
				lastMsg,
				newMsgsStringified})
	}
	err = rows.Err()
	if err != nil {
		return []User{}
	}
	return users
}

func (c *Commander) GetChatHistory(addr string) ([]NewMessage, error) {
	var messages []NewMessage
	db, err := c.openDB(addr)
	if err != nil {
		return []NewMessage{}, err
	}
	stmnt := `select origin, date, status, sender, input from messages;`
	rows, err := db.Query(stmnt)
	if err != nil {
		return []NewMessage{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var origin string
		var date string
		var status string
		var sender string
		var input string
		err = rows.Scan(&origin, &date, &status, &sender, &input)
		if err != nil {
			return []NewMessage{}, err
		}
		messages = append(messages, NewMessage{origin, date, status, sender, input})
	}
	err = rows.Err()
	if err != nil {
		return []NewMessage{}, err
	}
	return messages, nil
}

func (c *Commander) GetLastMessage(addr string) (NewMessage, error) {
	var msg NewMessage
	db, err := c.openDB(addr)
	if err != nil {
		return NewMessage{}, err
	}
	defer closeDB(db)
	stmnt := `select
	origin,
	date,
	status,
	sender,
	input from messages where id = (select max(id) from messages);`
	rows, err := db.Query(stmnt)
	if err != nil {
		return NewMessage{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var origin string
		var date string
		var status string
		var sender string
		var input string
		err = rows.Scan(&origin, &date, &status, &sender, &input)
		if err != nil {
			return NewMessage{}, err
		}
		msg = NewMessage{origin, date, status, sender, input}
	}
	err = rows.Err()
	if err != nil {
		return NewMessage{}, err
	}
	return msg, nil
}

func (c *Commander) GetNewMessages(addr string) (int, error) {
	amount := 0
	db, err := c.openDB(addr)
	if err != nil {
		return 0, err
	}
	defer closeDB(db)
	stmnt := `select id from messages where status = self;`
	rows, err := db.Query(stmnt)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		amount = amount + 1
	}
	err = rows.Err()
	if err != nil {
		return 0, err
	}
	return amount, nil
}

func (c *Commander) UpdateSelfMessages(address string) {
	db, err := c.openDB(address)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := `update messages set status=down where status=self;`
	_, err = db.Exec(stmnt)
	if err != nil {
		return
	}
	return
}

func (c *Commander) UpdateSentMessages(address string) {
	db, err := c.openDB(address)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := `update messages set status=read where status=sent;`
	_, err = db.Exec(stmnt)
	if err != nil {
		return
	}
	return
}

func (c *Commander) AddNewUser(u *NewUser) error {
	db, err := c.openDB(u.Addr)
	if err != nil {
		return err
	}
	stmnt := `create table messages(
	id integer not null primary key,
	origin text,
	date text,
	status text,
	sender text,
	input text);`
	_, err = db.Exec(stmnt)
	if err != nil {
		closeDB(db)
		return err
	}
	closeDB(db)
	db, err = c.openDB("history")
	if err != nil {
		pathNewUser := fmt.Sprintf("%s/history/%s.db", c.ConstantPath, u.Addr)
		os.Remove(pathNewUser)
		return err
	}
	stmnt = fmt.Sprintf(`insert into knownUsers(
		username,
		link,
		address,
		hash) values('', '%s', '%s', '%s');`, u.Link, u.Addr, u.Hash)
	_, err = db.Exec(stmnt)
	if err != nil {
		closeDB(db)
		pathNewUser := fmt.Sprintf("%s/history/%s.db", c.ConstantPath, u.Addr)
		os.Remove(pathNewUser)
		return err
	}
	closeDB(db)
	return nil
}

func (c *Commander) SaveMessage(addr string, msg NewMessage) error {
	db, err := c.openDB(addr)
	if err != nil {
		return err
	}
	defer closeDB(db)
	stmnt := fmt.Sprintf(
		`insert into messages(
		type,
		date,
		status,
		sender,
		input) values(
		'%s',
		'%s',
		'%s',
		'%s',
		'%s');`, msg.Type, msg.Date, msg.Status, msg.Sender, msg.Text)
	_, err = db.Exec(stmnt)
	if err != nil {
		return err
	}
	return nil
}
