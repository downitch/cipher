package api

import(
	"fmt"
	"bufio"
	"net"
	"log"
	"strings"
	// "time"

	"golang.org/x/net/proxy"
)

func (c *Commander) handleTCP(conn net.Conn, inc chan net.Conn) {
	var cbconn net.Conn
	defer func() {
		log.Printf("closing connection from %v now", conn.RemoteAddr())
		conn.Close()
	}()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	scanr := bufio.NewScanner(r)
	for {
		scanned := scanr.Scan()
		if !scanned {
			if err := scanr.Err(); err != nil {
				return
			}
			break
		}
		data := scanr.Text()
		dataParts := strings.Split(data, ":")
		switch dataParts[0] {
		case "handshake":
			callerId := dataParts[1]
			callerIdWithPort := fmt.Sprintf("%s:88", callerId)
			cbconn, err := c.Call(callerIdWithPort, "connected")
			if err != nil {
				fmt.Println(err)
				return
			}
			if cbconn != nil {
				response := fmt.Sprintf("connected:%s\n", c.GetTCPHSLink())
				w.WriteString(response)
				w.Flush()
				inc <- cbconn
			} else {
				fmt.Println("CANT CONNECT BACK")
				w.WriteString("\n")
				w.Flush()
				inc <- nil
				return
			}
		case "endCall":
			if cbconn != nil {
				cbconn.Close()
			}
			return
		default:
			// received bytes transformed to audio
			// 
			// ******* transforming it *******
			// 
			// done, sending back amount of bytes received
			fmt.Println(data)
			w.WriteString(fmt.Sprintf("%d\n", len(data)))
			w.Flush()
		}
	}
	return
}

func (c *Commander) SendBytes(conn net.Conn, input string) bool {
	if conn == nil {
		return false
	}
	inputWithNewLine := fmt.Sprintf("%s\n", input)
	toSend := []byte(inputWithNewLine)
	_, err := conn.Write(toSend)
	if err != nil {
		return false
	}
	return true
}

func (c *Commander) Call(callerId string, status string) (net.Conn, error) {
	var conn net.Conn
	var err error
	dailer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, &net.Dialer{})
	if err != nil {
		return conn, err
	}
	conn, err = dailer.Dial("tcp", callerId)
	if err != nil {
		return conn, err
	}
	if status == "call" {
		toSend := fmt.Sprintf("handshake:%s\n", c.GetTCPHSLink())
		_, err = conn.Write([]byte(toSend))
		if err != nil {
			return conn, err
		}
	}
	return conn, nil
}

func (c *Commander) EndCall(conn net.Conn) {
	c.SendBytes(conn, fmt.Sprintf("endCall:%s", c.GetTCPHSLink()))
	conn.Close()
	return
}

func (c *Commander) RunTCPServer() {
	incomingConnection := make(chan net.Conn)
	// Here we define our TCP web-server that will be visible from darkweb
	listener, err := net.Listen("tcp", ":4888")
	if err != nil {
		return
	}
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("error accepting connection %v", err)
				continue
			}
			log.Printf("accepted connection from %v", conn.RemoteAddr())
			c.handleTCP(conn, incomingConnection)
		}
	}()
	for {
		if connection := <- incomingConnection; connection != nil {
			fmt.Println("HAHAHAHAHAHAH WORKS")
			c.SendBytes(connection, "HAHAHAHAHAHAH WORKS")
		} else {
			fmt.Println("No connection")
		}
	}
}
