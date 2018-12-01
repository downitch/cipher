package api

import(
	"database/sql"
	"net"
)

type Commander struct {
	ConstantPath string
	Connection   net.Conn
	DbFilename   string
	DbConnection *sql.DB
}

func NewCommander(path string) *Commander {
	var connection   net.Conn
	dbconnection  := &sql.DB{}
	return &Commander{ 
		ConstantPath: path,
		Connection:   connection,
		DbFilename:   "",
		DbConnection: dbconnection,
	}
}

func (c *Commander) ChangeCommanderPath(path string) {
	c.ConstantPath = path
}

func (c *Commander) AcceptConnection(connection net.Conn) {
	c.Connection = connection
}

func (c *Commander) SetDatabaseConnection(name string, conn *sql.DB) {
	c.DbFilename   = name
	c.DbConnection = conn
}