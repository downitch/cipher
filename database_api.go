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
	Pinned   string `json:"pinned"`
}

type Status string

const(
	Sent   Status = "sent"
	Read   Status = "read"
	Fail   Status = "fail"
	Unread Status = "unread"
	New    Status = "new"
)

type NewUser struct {
	Link string
	Addr string
	Hash string
}

type User struct {
	Username 		  string       `json:"username"`
	Link 		 		  string       `json:"link"`
	Addr 		 		  string       `json:"addr"`
	LastMessage   NewMessage   `json:"lastMessage"`
	NewMessages   string       `json:"newMessages"`
	Notifications []NewMessage `json:"notifications"`
	LastOnline    string       `json:"lastOnline"`
}

func (c *Commander) AddTableColumn(tname string, cname string, ctype string, dvalue string, dbname string) {
	dbnames := []string{dbname}
	for i := range dbnames {
		db, err := c.openDB(dbnames[i])
		if err != nil {
			e := fmt.Sprintf("No such database with name %s found", dbnames[i])
			fmt.Println(e)
			return
		}
		defer closeDB(db)
		stmnt := fmt.Sprintf("select %s from %s limit 1;", cname, tname)
		_, err = db.Prepare(stmnt)
		if err != nil {
			stmnt = fmt.Sprintf("alter table %s add column %s %s;", tname, cname, ctype)
			_, err = db.Exec(stmnt)
			if err != nil {
				fmt.Println(err)
				fmt.Println("Can't alter table from database ", dbnames[i])
				return
			}
			stmnt = fmt.Sprintf("update %s set %s = %s", tname, cname, dvalue)
			_, err = db.Exec(stmnt)
			if err != nil {
				fmt.Println("Can't update to default value")
				return
			}
		}
	}
}

func (c *Commander) UpdateStorage() bool {
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
		closeDB(db)
		return false
	}
	closeDB(db)
	db, err = c.openDB("blocks")
	if err != nil {
		return false
	}
	stmnt = `create table if not exists blocks(
	id integer not null primary key,
	hash text,
	number int);`
	_, err = db.Exec(stmnt)
	if err != nil {
		closeDB(db)
		return false
	}
	closeDB(db)
	return true
}

func (c *Commander) openDB(name string) (*sql.DB, error) {
	// db := c.DbConnection
	// var err error
	// if c.DbFilename != name {
	// 	path := c.ConstantPath
	// 	fullPath := fmt.Sprintf("%s/history/%s.db", path, name)
	// 	db, err = sql.Open("sqlite3", fullPath)
	// 	if err != nil {
	// 		return &sql.DB{}, err
	// 	}
	// }
	// if c.DbConnection == nil {
	// 	path := c.ConstantPath
	// 	fullPath := fmt.Sprintf("%s/history/%s.db", path, name)
	// 	db, err = sql.Open("sqlite3", fullPath)
	// 	if err != nil {
	// 		return &sql.DB{}, err
	// 	}
	// }
	// c.SetDatabaseConnection(name, db)
	// return db, nil
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
		
		return true
	}
	defer closeDB(db)
	stmnt := "select address from knownUsers where link = ?"
	st, err := db.Prepare(stmnt)
	if err != nil {
		
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
		notifications := c.GetNotifications(address)
		newMsgsStringified := strconv.Itoa(newMsgs)
		lastOnline, _ := c.GetLastOnline(address)
		users = append(
			users, User{
				username,
				link,
				address,
				lastMsg,
				newMsgsStringified,
				notifications,
				lastOnline,
			})
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
	defer closeDB(db)
	stmnt := `select id, origin, date, status, sender, input, pinned from messages;`
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
		var pinned string
		err = rows.Scan(&id, &origin, &date, &status, &sender, &input, &pinned)
		if err != nil {
			return []NewMessage{}, err
		}
		messages = append(messages, NewMessage{id, origin, date, status, sender, input, pinned})
	}
	err = rows.Err()
	if err != nil {
		return []NewMessage{}, err
	}
	return messages, nil
}

func (c *Commander) GetLastOnline(addr string) (string, error) {
	db, err := c.openDB(addr)
	if err != nil {
		return "0", err
	}
	defer closeDB(db)
	stmnt := "select date from messages where sender = ? order by id desc limit 1;"
	st, err := db.Prepare(stmnt)
	if err != nil {
		return "0", nil
	}
	defer st.Close()
	var date string
	err = st.QueryRow(addr).Scan(&date)
	if err != nil {
		return "0", nil
	}
	return date, nil
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
	input,
	pinned from messages where id = (select max(id) from messages);`
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
	var pinned string
	err = st.QueryRow().Scan(&id, &origin, &date, &status, &sender, &input, &pinned)
	if err != nil {
		return NewMessage{}, nil
	}
	msg = NewMessage{id, origin, date, status, sender, input, pinned}
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

func (c *Commander) GetNotifications(addr string) []NewMessage {
	var messages []NewMessage
	db, err := c.openDB(addr)
	if err != nil {
		return []NewMessage{}
	}
	defer closeDB(db)
	stmnt := `select id, origin, date, status, sender, input, pinned from messages where status = 'new';`
	rows, err := db.Query(stmnt)
	if err != nil {
		return []NewMessage{}
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var origin string
		var date string
		var status string
		var sender string
		var input string
		var pinned string
		err = rows.Scan(&id, &origin, &date, &status, &sender, &input, &pinned)
		if err != nil {
			return []NewMessage{}
		}
		messages = append(messages, NewMessage{id, origin, date, status, sender, input, pinned})
	}
	err = rows.Err()
	if err != nil {
		return []NewMessage{}
	}
	return messages
}

func (c *Commander) GetNewMessages(addr string) (int, error) {
	amount := 0
	db, err := c.openDB(addr)
	if err != nil {
		return 0, err
	}
	defer closeDB(db)
	stmnt := `select id from messages where status = ? or status = ?;`
	rows, err := db.Query(stmnt, "self", "new")
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

func (c *Commander) GetMessageById(addr string, id int) NewMessage {
	msg := NewMessage{}
	db, err := c.openDB(addr)
	if err != nil {
		return NewMessage{}
	}
	defer closeDB(db)
	stmnt := "select id, origin, date, status, sender, input, pinned from messages where id = ?;"
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
	var pinned string
	err = st.QueryRow(id).Scan(&uid, &origin, &date, &status, &sender, &input, &pinned)
	if err != nil {
		return NewMessage{}
	}
	msg = NewMessage{uid, origin, date, status, sender, input, pinned}
	return msg
}

func (c *Commander) UpdateSelfMessages(address string) {
	db, err := c.openDB(address)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := `update messages set status = ? where status = ?;`
	db.Exec(stmnt, "down", "self")
	return
}

func (c *Commander) UpdatedSelfNewMessages(address string) {
	db, err := c.openDB(address)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := `update messages set status = ? where status = ?;`
	db.Exec(stmnt, "self", "new")
	return
}

func (c *Commander) UpdateSentMessages(address string) {
	db, err := c.openDB(address)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := `update messages set status = ? where status = ?;`
	db.Exec(stmnt, "read", "sent")
	return
}

func (c *Commander) UpdateFailedMessage(id int, address string) {
	db, err := c.openDB(address)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := `update messages set status = ? where id = ?;`
	db.Exec(stmnt, "failed", id)
	return
}

func (c *Commander) UpdateUnfailMessage(id int, address string) {
	db, err := c.openDB(address)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := `update messages set status = ? where id = ?;`
	db.Exec(stmnt, "sent", id)
	return
}

func (c *Commander) UpdateUnreadMessage(id int, address string) {
	db, err := c.openDB(address)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := `update messages set status = ? where id = ?;`
	db.Exec(stmnt, "unread", id)
	return
}

func(c *Commander) SaveBlock(hash string, number int) error {
	db, err := c.openDB("blocks")
	if err != nil {
		return err
	}
	stmnt := "select count(*) from blocks"
	st, err := db.Prepare(stmnt)
	if err != nil {
		return err
	}
	var amount int
	err = st.QueryRow().Scan(&amount)
	if err != nil {
		return err
	}
	st.Close()
	closeDB(db)
	if amount > 25 {
		return nil
	}
	db, err = c.openDB("blocks")
	if err != nil {
		return err
	}
	defer closeDB(db)
	stmnt = fmt.Sprintf(`insert into blocks(hash, number) values('%s', '%d')`, hash, number)
	_, err = db.Exec(stmnt)
	if err != nil {
		return err
	}
	return nil
}

func (c *Commander) GetRandomBlockFromDB() RandomBlock {
	db, err := c.openDB("blocks")
	if err != nil {
		return RandomBlock{}
	}
	defer closeDB(db)
	stmnt := "select id, hash, number from blocks where id >= (abs(random()) % (SELECT max(id) FROM blocks)) limit 1"
	st, err := db.Prepare(stmnt)
	if err != nil {
		return RandomBlock{}
	}
	var id int
	var hash string
	var number int
	err = st.QueryRow().Scan(&id, &hash, &number)
	if err != nil {
		return RandomBlock{}
	}
	st.Close()
	stmnt = fmt.Sprintf(`delete from blocks where id = '%d'`, id)
	_, err = db.Exec(stmnt)
	if err != nil {
		return RandomBlock{}
	}
	return RandomBlock{hash, number}
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

func (c *Commander) GetPinnedMessage(addr string) int {
	db, err := c.openDB(addr)
	if err != nil {
		return 0
	}
	defer closeDB(db)
	stmnt := "select id from messages where pinned = ?"
	st, err := db.Prepare(stmnt)
	if err != nil {
		return 0
	}
	defer st.Close()
	var mid int
	err = st.QueryRow("true").Scan(&mid)
	if err != nil {
		return 0
	}
	return mid
}

func (c *Commander) PinMessage(addr string, mid int) {
	db, err := c.openDB(addr)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := "select id from messages where pinned = ?"
	st, err := db.Prepare(stmnt)
	if err != nil {
		return
	}
	defer st.Close()
	id := 0
	err = st.QueryRow("true").Scan(&id)
	if err != nil {
		stmnt = "update messages set pinned = ? where id = ?"
		_, err = db.Exec(stmnt, "true", mid)
		if err != nil {
			
		}
		return
	}
	if id != 0 {
		c.UnpinMessage(addr)
	}
	stmnt = "update messages set pinned = ? where id = ?"
	_, err = db.Exec(stmnt, "true", mid)
	if err != nil {
		return
	}
	return
}

func (c *Commander) UnpinMessage(addr string) {
	db, err := c.openDB(addr)
	if err != nil {
		return
	}
	defer closeDB(db)
	stmnt := "update messages set pinned = ?"
	db.Exec(stmnt, "false")
	return
}

func (c *Commander) SaveMessage(addr string, rec string, mtype string, msg string) int {
	status := "sent"
	db, err := c.openDB(rec)
	if err != nil {
		return 0
	}
	defer closeDB(db)
	if addr == rec {
		status = "new"
	}
	date := strconv.Itoa(int(time.Now().Unix()))
	stmnt := fmt.Sprintf(
		`insert into messages(
		origin,
		date,
		status,
		sender,
		input,
		pinned) values(
		'%s',
		'%s',
		'%s',
		'%s',
		'%s',
		'false');`, mtype, date, status, addr, msg)
	_, err = db.Exec(stmnt)
	if err != nil {
		return 0
	}
	id := c.GetLastMessageId(rec)
	return id
}