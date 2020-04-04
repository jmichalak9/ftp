package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type client struct {
	conn     net.Conn
	username string
	datatype string
}
type commandHandler func(client, string) error

var handlers = map[string]commandHandler{
	"USER": handleUSER,
	"PASS": handlePASS,
	"SYST": handleSYST,
	"PWD":  handlePWD,
	"TYPE": handleTYPE,
	"SIZE": handleSIZE,
	"QUIT": handleQUIT,
}

func handleUSER(c client, argv string) error {
	c.username = argv
	if argv == "anonymous" {
		fmt.Println("Logging user anonymous")
		c.conn.Write([]byte("331 User name ok, need password\r\n"))
	} else {
		fmt.Println("Logging unknown user")
		c.conn.Write([]byte("331 User name ok, need password\r\n"))
	}
	return nil
}

func handlePASS(c client, argv string) error {
	c.conn.Write([]byte("230 User logged in, proceed\r\n"))
	return nil
}

func handleSYST(c client, argv string) error {
	c.conn.Write([]byte("215 UNIX Type: L8\r\n"))
	return nil
}

func handlePWD(c client, argv string) error {
	//pwd, _ := os.Getwd()
	c.conn.Write([]byte("217 " + " \r\n"))
	return nil
}

func handleTYPE(c client, argv string) error {
	c.datatype = argv
	c.conn.Write([]byte("200 Command okay.\r\n"))
	return nil
}

func handleSIZE(c client, argv string) error {
	c.conn.Write([]byte("502 Command not implemented.\r\n"))
	return c.conn.Close()
}

func handleQUIT(c client, argv string) error {
	c.conn.Write([]byte("221 Service closing control connection.\r\n"))
	return c.conn.Close()
}

func handleConnection(conn net.Conn) {
	fmt.Println("new connection")
	c := client{
		conn: conn,
	}
	c.conn.Write([]byte("220 FTP server\r\n"))
	for {
		netData, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(string(netData))

		line := strings.SplitN(strings.TrimSpace(string(netData)), " ", 2)
		// if the command is without arguments
		if len(line) == 1 {
			line = append(line, "")
		}
		cmd, argv := line[0], line[1]
		if handler, ok := handlers[cmd]; ok {
			handler(c, argv)
		}
		// Only temporary. TODO: handle connection end.
		if cmd == "STOP" {
			break
		}

	}
	c.conn.Close()
}

func main() {
	PORT := ":2137"
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}
