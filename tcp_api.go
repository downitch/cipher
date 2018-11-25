package api

import(
	"bufio"
	"net"
	"log"
	"strings"
)

func handleTCP(conn net.Conn) error {
	defer func() {
		log.Printf("closing connection from %v", conn.RemoteAddr())
		conn.Close()
	}()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	scanr := bufio.NewScanner(r)
	for {
		scanned := scanr.Scan()
		if !scanned {
			if err := scanr.Err(); err != nil {
				log.Printf("%v(%v)", err, conn.RemoteAddr())
				return err
			}
			break
		}
		w.WriteString(strings.ToUpper(scanr.Text()) + "\n")
		w.Flush()
	}
	return nil
}

func (c *Commander) RunTCPServer() {
	// Here we define our TCP web-server that will be visible from darkweb
	listener, err := net.Listen("tcp", ":4888")
	if err != nil {
		return
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting connection %v", err)
			continue
		}
		log.Printf("accepted connection from %v", conn.RemoteAddr())
		handleTCP(conn)
	}
}