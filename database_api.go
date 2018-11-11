package api

import(
	"fmt"
	"database/sql"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type NewMessage struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Date     string `json:"date"`
	Status   string `json:"status"`
	Sender   string `json:"author"`
	Text     string `json:"text"`
}

type Status string

const(
	Sent   Status = "sent"
	Read   Status = "read"
	Fail   Status = "fail"
	Unread Status = "unread"
)

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
	// path := c.ConstantPath + "/history/"
	// if _, err := os.Stat(); os.IsNotExist(err) {

	// }
	db, err := c.openDB("history")
	if err != nil {
		return false
	}
	stmnt := `create table if not exists knownUsers(
	id integer not null primary key,
	username text,
	link text,
	address text,
	hash text);`
	_, err = db.Exec(stmnt)
	if err != nil {
		// fmt.Println(err)
		closeDB(db)
		return false
	}
	closeDB(db)
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
	stmnt := "select link from knownUsers where address = ?"
	st, err := db.Prepare(stmnt)
	if err != nil {
		return ""
	}
	defer st.Close()
	err = st.QueryRow(address).Scan(&link)
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
	stmnt := "select address from knownUsers where link = ?"
	st, err := db.Prepare(stmnt)
	if err != nil {
		return ""
	}
	defer st.Close()
	err = st.QueryRow(link).Scan(&address)
	if err != nil {
		return ""
	}
	return address
}

func (c *Commander) GetCipherByAddress(address string) string {
	var cipher string
	db, err := c.openDB("history")
	if err != nil {
		return ""
	}
	defer closeDB(db)
	stmnt := "select hash from knownUsers where address = ?"
	st, err := db.Prepare(stmnt)
	if err != nil {
		return ""
	}
	defer st.Close()
	err = st.QueryRow(address).Scan(&cipher)
	if err != nil {
		return ""
	}
	return cipher
}

func (c *Commander) CheckExistance(link string) bool {
	db, err := c.openDB("history")
	if err != nil {
		// fmt.Println("cant open db")
		return true
	}
	defer closeDB(db)
	stmnt := "select address from knownUsers where link = ?"
	st, err := db.Prepare(stmnt)
	if err != nil {
		// fmt.Println(err)
		return true
	}
	defer st.Close()
	address := ""
	err = st.QueryRow(link).Scan(&address)
	if err != nil {
		return false
	}
	return true
}

func (c *Commander) GetChats() []User {
	var users []User
	db, err := c.openDB("history")
	if err != nil {
		// fmt.Println(err)
		return []User{}
	}
	defer closeDB(db)
	stmnt := `select username, link, address from knownUsers;`
	rows, err := db.Query(stmnt)
	if err != nil {
		// fmt.Println("Error on query from knownUsers")
		// fmt.Println(err)
		return []User{}
	}
	defer rows.Close()
	for rows.Next() {
		var username string
		var link string
		var address string
		err = rows.Scan(&username, &link, &address)
		if err != nil {
			// fmt.Println("error on Scanning row")
			// fmt.Println(err)
			return []User{}
		}
		lastMsg, err := c.GetLastMessage(address)
		if err != nil {
			// fmt.Println("error getting last message")
			// fmt.Println(err)
			return []User{}
		}
		newMsgs, err := c.GetNewMessages(address)
		if err != nil {
			// fmt.Println("error getting amount of self messages")
			// fmt.Println(err)
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
		// fmt.Println("error at checking rows error")
		// fmt.Println(err)
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
	stmnt := `select id, origin, date, status, sender, input from messages;`
	rows, err := db.Query(stmnt)
	if err != nil {
		return []NewMessage{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var origin string
		var date string
		var status string
		var sender string
		var input string
		err = rows.Scan(&id, &origin, &date, &status, &sender, &input)
		if err != nil {
			return []NewMessage{}, err
		}
		messages = append(messages, NewMessage{id, origin, date, status, sender, input})
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
	id,
	origin,
	date,
	status,
	sender,
	input from messages where id = (select max(id) from messages);`
	st, err := db.Prepare(stmnt)
	if err != nil {
		return NewMessage{}, nil
	}
	defer st.Close()
	var id string
	var origin string
	var date string
	var status string
	var sender string
	var input string
	err = st.QueryRow().Scan(&id, &origin, &date, &status, &sender, &input)
	if err != nil {
		return NewMessage{}, nil
	}
	msg = NewMessage{id, origin, date, status, sender, input}
	// fmt.Println(msg)
	return msg, nil
}

func (c *Commander) GetLastMessageId(addr string) int {
	var id int
	db, err := c.openDB(addr)
	if err != nil {
		return 0
	}
	defer closeDB(db)
	stmnt := `select max(id) from messages;`
	st, err := db.Prepare(stmnt)
	if err != nil {
		return 0
	}
	defer st.Close()
	err = st.QueryRow().Scan(&id)
	if err != nil {
		return 0
	}
	return id
}

func (c *Commander) GetNewMessages(addr string) (int, error) {
	amount := 0
	db, err := c.openDB(addr)
	if err != nil {
		return 0, err
	}
	defer closeDB(db)
	stmnt := `select id from messages where status = ?;`
	rows, err := db.Query(stmnt, "self")
	if err != nil {
		// fmt.Println(err)
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		amount = amount + 1
		// fmt.Println(amount)
	}
	err = rows.Err()
	if err != nil {
		// fmt.Println("KEK")
		// fmt.Println(err)
		return 0, err
	}
	return amount, nil
}

func (c *Commander) GetMessageById(addr string, id int) NewMessage {
	msg := NewMessage{}
	db, err := c.openDB(addr)
	if err != nil {
		return NewMessage{}
	}
	stmnt := "select id, origin, date, status, sender, input from messages where id = ?;"
	st, err := db.Prepare(stmnt)
	if err != nil {
		return NewMessage{}
	}
	defer st.Close()
	var uid string
	var origin string
	var date string
	var status string
	var sender string
	var input string
	err = st.QueryRow(id).Scan(&uid, &origin, &date, &status, &sender, &input)
	if err != nil {
		return NewMessage{}
	}
	msg = NewMessage{uid, origin, date, status, sender, input}
	return msg
}

func (c *Commander) UpdateSelfMessages(address string) {
	db, err := c.openDB(address)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := `update messages set status = ? where status = ?;`
	_, err = db.Exec(stmnt, "down", "self")
	if err != nil {
		// fmt.Println(err)
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
	stmnt := `update messages set status = ? where status = ?;`
	_, err = db.Exec(stmnt, "read", "sent")
	if err != nil {
		// fmt.Println(err)
		return
	}
	return
}

func (c *Commander) UpdateFailedMessage(id int, address string) {
	db, err := c.openDB(address)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := `update messages set status = ? where id = ?;`
	_, err = db.Exec(stmnt, "failed", id)
	if err != nil {
		// fmt.Println(err)
		return
	}
	return
}

func (c *Commander) UpdateUnfailMessage(id int, address string) {
	db, err := c.openDB(address)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := `update messages set status = ? where id = ?;`
	_, err = db.Exec(stmnt, "sent", id)
	if err != nil {
		// fmt.Println(err)
		return
	}
	return
}

func (c *Commander) UpdateUnreadMessage(id int, address string) {
	db, err := c.openDB(address)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := `update messages set status = ? where id = ?;`
	_, err = db.Exec(stmnt, "unread", id)
	if err != nil {
		// fmt.Println(err)
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

func (c *Commander) SetUsername(link string, username string) bool {
	db, err := c.openDB("history")
	if err != nil {
		return false
	}
	defer closeDB(db)
	stmnt := "update knownUsers set username = ? where link = ?"
	_, err = db.Exec(stmnt, username, link)
	if err != nil {
		return false
	}
	return true
}

func (c *Commander) DeleteContact(link string) bool {
	db, err := c.openDB("history")
	if err != nil {
		return false
	}
	defer closeDB(db)
	address := c.GetAddressByLink(link)
	stmnt := fmt.Sprintf(`delete from knownUsers where link = '%s'`, link)
	_, err = db.Exec(stmnt)
	if err != nil {
		return false
	}
	path := c.ConstantPath
	fullPath := fmt.Sprintf("%s/history/%s.db", path, address)
	os.Remove(fullPath)
	return true
}

func (c *Commander) SaveMessage(addr string, rec string, msg string) int {
	status := "sent"
	db, err := c.openDB(rec)
	if err != nil {
		return 0
	}
	defer closeDB(db)
	if addr == rec {
		status = "self"
	}
	date := strconv.Itoa(int(time.Now().Unix()))
	stmnt := fmt.Sprintf(
		`insert into messages(
		origin,
		date,
		status,
		sender,
		input) values(
		'%s',
		'%s',
		'%s',
		'%s',
		'%s');`, "text", date, status, addr, msg)
	_, err = db.Exec(stmnt)
	if err != nil {
		// fmt.Println("Statement broken")
		// fmt.Println(err)
		return 0
	}
	id := c.GetLastMessageId(rec)
	return id
}