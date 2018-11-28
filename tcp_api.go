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

func handleTCP(conn net.Conn) error {
	defer func() {
		log.Printf("closing connection from %v now", conn.RemoteAddr())
		conn.Close()
	}()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	scanr := bufio.NewScanner(r)
	for {
		fmt.Println("scanning...")
		scanned := scanr.Scan()
		fmt.Println(scanned)
		fmt.Println("scanned")
		if !scanned {
			if err := scanr.Err(); err != nil {
				log.Printf("%v(%v)", err, conn.RemoteAddr())
				return err
			}
			break
		}
		data := scanr.Text()
		dataParts := strings.Split(data, ":")
		if dataParts[0] == "handshake" {
			callerId := dataParts[1]
			w.WriteString(callerId + "\n")
			w.Flush()
		} else {
			w.WriteString(strings.ToUpper(data) + "\n")
			w.Flush()
		}
	}
	return nil
}

func Call(callerId string) {
	dailer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, &net.Dialer{})
	if err != nil {
		fmt.Println(err)
		return
	}
	c, err := dailer.Dial("tcp", callerId)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("sending hello...")
	_, err = c.Write([]byte("Hello\n"))
	if err != nil {
		println("Write to server failed:", err.Error())
		return
	}
	fmt.Println("reading reply")
	reply := make([]byte, 1024)
	_, err = c.Read(reply)
	if err != nil {
		println("Write to server failed:", err.Error())
		return
	}
	println("reply from server=", string(reply))
	c.Close()
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
