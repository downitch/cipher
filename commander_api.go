package api

import (
	"database/sql"
	"net"
)

type Commander struct {
	ConstantPath      string
	Connection        net.Conn
	DbFilename        string
	DbConnection      *sql.DB
	MessagesCount     int
	OldMessagesCount  int
	OldMessagesObsrvr bool
}

func NewCommander(path string) *Commander {
	var connection net.Conn
	var dbconnection sql.DB
	var dbfilename string
	var messagecount int
	observer := false
	return &Commander{
		ConstantPath:      path,
		Connection:        connection,
		DbFilename:        dbfilename,
		DbConnection:      &dbconnection,
		MessagesCount:     messagecount,
		OldMessagesCount:  messagecount,
		OldMessagesObsrvr: observer,
	}
}

func (c *Commander) ChangeCommanderPath(path string) {
	c.ConstantPath = path
}

func (c *Commander) AcceptConnection(connection net.Conn) {
	c.Connection = connection
}

func (c *Commander) SetDatabaseConnection(name string, conn *sql.DB) {
	c.DbFilename = name
	c.DbConnection = conn
}

func (c *Commander) SetMessageCount(count int) {
	c.MessagesCount = count
}

func (c *Commander) SetOldMessageCount(count int) {
	c.OldMessagesCount = count
}

func (c *Commander) StartObserving() {
	c.OldMessagesObsrvr = true
}

func (c *Commander) StopObserving() {
	c.OldMessagesObsrvr = false
}
