package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"
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
	c.conn.Write([]byte("217 /\r\n"))
	return nil
}

func handleTYPE(c client, argv string) error {
	c.datatype = argv
	c.conn.Write([]byte("200 Command okay.\r\n"))
	return nil
}

func pathToSlice(path string) []string {
	normalised := strings.TrimPrefix(path, "/")
	return strings.Split(normalised, "/")
}

func getItemFromPath(path string) (interface{}, error) {
	nameslice := pathToSlice(path)
	searchDir := files
	for _, name := range nameslice {
		// item is an entry in a filesystem
		item, ok := searchDir[name]
		if !ok {
			return "", errors.New("File not found")
		}
		if reflect.TypeOf(item).String() == "string" {
			// item is a file
			return item, nil
		}
		if reflect.TypeOf(item).String() == "map[string]interface{}" {
			// item is a directory
			searchDir = item.(map[string]interface{})
			continue
		}
		// item is of unknown type
		return "", errors.New("unknown entry type in file system")
	}
	return searchDir, nil
}

func handleSIZE(c client, argv string) error {
	file, err := getItemFromPath(argv)
	if err != nil || reflect.TypeOf(file).String() != "string" {
		c.conn.Write([]byte("550 Cannot read file size.\r\n"))
		return nil
	}
	c.conn.Write([]byte("213 " + strconv.Itoa(len(file.(string))) + "\r\n"))
	return nil
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
	c.conn.SetDeadline(time.Now().Add(time.Minute))
	defer c.conn.Close()
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
	}
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
