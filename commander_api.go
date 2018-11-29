package api

import(
	"net"
)

type Commander struct {
	ConstantPath string
	Connection net.Conn
}

func NewCommander(path string) *Commander {
	var connection net.Conn
	return &Commander{ 
		ConstantPath: path,
		Connection: connection,
	}
}

func (c *Commander) ChangeCommanderPath(path string) {
	c.ConstantPath = path
}

func (c *Commander) AcceptConnection(connection net.Conn) {
	c.Connection = connection
}