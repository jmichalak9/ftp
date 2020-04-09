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
	listener net.Listener
}
type commandHandler func(client, string) error

var handlers = map[string]commandHandler{
	"USER": handleUSER,
	"PASS": handlePASS,
	"SYST": handleSYST,
	"PWD":  handlePWD,
	"CWD":  handleCWD,
	"TYPE": handleTYPE,
	"SIZE": handleSIZE,
	"PASV": handlePASV,
	"LIST": handleLIST,
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

func handleCWD(c client, argv string) error {
	file, err := getItemFromPath(argv)
	if err != nil || reflect.TypeOf(file).String() != "map[string]interface {}" {
		c.conn.Write([]byte("550 file not found\r\n"))
		return nil
	}
	c.conn.Write([]byte("250 OK\r\n"))
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
	if nameslice[0] == "" {
		return searchDir, nil
	}
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
		if reflect.TypeOf(item).String() == "map[string]interface {}" {
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

func handlePASV(c client, argv string) error {
	ip := []int{127, 0, 0, 1}
	port := c.listener.Addr().(*net.TCPAddr).Port
	c.conn.Write([]byte(fmt.Sprintf("227 Entering Passive Mode (%d,%d,%d,%d,%d,%d)\r\n",
		ip[0], ip[1], ip[2], ip[3], port/256, port%256)))
	return nil
}

func handleLIST(c client, argv string) error {
	conn, err := c.listener.Accept()
	if err != nil {
		fmt.Println(err)
		return err
	}
	c.conn.Write([]byte("150 File status okay; about to open data connection.\r\n"))
	dir, err := getItemFromPath("/")
	if reflect.TypeOf(dir).String() != "map[string]interface {}" || err != nil {
		// TODO: write error to c.conn
		return nil
	}
	reply := ""
	for k := range dir.(map[string]interface{}) {
		reply += ("kuba wheel 776 Nov 24 15:50 " + k + "\r\n")
	}
	conn.Write([]byte(reply))
	conn.Close()
	c.conn.Write([]byte("226 Closing data connection.\r\n"))
	return nil
}

func handleConnection(conn net.Conn) {
	// create additional port to transfer files
	listener, err := net.Listen("tcp4", ":0")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()
	c := client{
		conn:     conn,
		listener: listener,
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
