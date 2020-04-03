package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type commandHandler func(string) error

var handlers = map[string]commandHandler{
	"USER": handleUSER,
}

func handleUSER(argv string) error {
	fmt.Printf("USER handler called with %s\n", argv)
	return nil
}

func handleConnection(c net.Conn) {
	fmt.Println("new connection")
	c.Write([]byte("220 FTP server\r\n"))
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		line := strings.SplitN(strings.TrimSpace(string(netData)), " ", 2)
		if len(line) < 2 {
			break
		}
		cmd, argv := line[0], line[1]
		if handler, ok := handlers[cmd]; ok {
			handler(argv)
		}
		// Only temporary. TODO: handle connection end.
		if cmd == "STOP" {
			break
		}

		fmt.Println(line)
		c.Write([]byte("sent " + string(netData)))
	}
	c.Close()
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
