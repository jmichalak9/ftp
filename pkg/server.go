// Package ftp implements FTP server (RFC 959 and related).
package ftp

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

// Server represents an FTP server.
type Server struct {
	Addr string
}

// Client represents an FTP client. It is used by a server for communication.
type client struct {
	conn     net.Conn
	username string
	pwd      string
	datatype string
	dataSock net.Listener
	toClose  bool
}

type commandHandler func(*client, string) error

var handlers = map[string]commandHandler{
	"CWD":  handleCWD,
	"FEAT": handleFEAT,
	"LIST": handleLIST,
	"MDTM": handleMDTM,
	"PASS": handlePASS,
	"PASV": handlePASV,
	"PWD":  handlePWD,
	"QUIT": handleQUIT,
	"RETR": handleRETR,
	"SIZE": handleSIZE,
	"SYST": handleSYST,
	"TYPE": handleTYPE,
	"USER": handleUSER,
}

var eol = "\r\n"

func handleCWD(c *client, argv string) error {
	file, err := getItemFromPath(argv)
	if err != nil || reflect.TypeOf(file).String() != "map[string]interface {}" {
		c.conn.Write([]byte(ReplyFileUnavailable + eol))
		return err
	}
	c.pwd = argv
	_, err = c.conn.Write([]byte(ReplyFileActionOK + eol))
	return err
}

func handleFEAT(c *client, argv string) error {
	_, err := c.conn.Write([]byte("211-Features:\r\n" +
		"211 End\r\n"))
	return err
}

func handleLIST(c *client, argv string) error {
	conn, err := c.dataSock.Accept()
	if err != nil {
		fmt.Println(err)
		return err
	}
	c.conn.Write([]byte(ReplyFileStatusOK + eol))
	dir, err := getItemFromPath(c.pwd)
	if reflect.TypeOf(dir).String() != "map[string]interface {}" || err != nil {
		// TODO: write error to c.conn
		return nil
	}
	reply := ""
	for k := range dir.(map[string]interface{}) {
		// TODO: send proper file properties
		reply += ("-rwxr-xr-x  10 kuba wheel 776 Nov 24 15:50 " + k + "\r\n")
	}
	conn.Write([]byte(reply))
	conn.Close()
	c.conn.Write([]byte(ReplyClosingDataConn + eol))
	return nil
}

func handleMDTM(c *client, argv string) error {
	file, err := getItemFromPath(argv)
	if err != nil || reflect.TypeOf(file).String() != "string" {
		c.conn.Write([]byte(ReplyFileUnavailable + eol))
		return err
	}
	_, err = c.conn.Write([]byte("213 207001010000\r\n"))
	return err
}

func handlePASS(c *client, argv string) error {
	_, err := c.conn.Write([]byte(ReplyUserLoggedIn + eol))
	return err
}

func handlePASV(c *client, argv string) error {
	ip := []int{127, 0, 0, 1}
	port := c.dataSock.Addr().(*net.TCPAddr).Port
	_, err := c.conn.Write([]byte(fmt.Sprintf(ReplyEnteringPasv,
		ip[0], ip[1], ip[2], ip[3], port/256, port%256) + eol))
	return err
}

func handlePWD(c *client, argv string) error {
	_, err := c.conn.Write([]byte(fmt.Sprintf(ReplyPathNameOK, c.pwd) + eol))
	return err
}

func handleQUIT(c *client, argv string) error {
	c.toClose = true
	_, err := c.conn.Write([]byte(ReplyClosingConn + eol))
	return err
}

func handleRETR(c *client, argv string) error {
	conn, err := c.dataSock.Accept()
	if err != nil {
		fmt.Println(err)
		return err
	}
	c.conn.Write([]byte(ReplyFileStatusOK + eol))
	file, err := getItemFromPath(argv)
	fmt.Printf("%s %v", argv, file)
	if reflect.TypeOf(file).String() != "string" || err != nil {
		// TODO: write error to c.conn
		return nil
	}
	conn.Write([]byte(file.(string)))
	conn.Close()
	c.conn.Write([]byte(ReplyClosingDataConn + eol))
	return nil
}

func handleSIZE(c *client, argv string) error {
	file, err := getItemFromPath(argv)
	if err != nil || reflect.TypeOf(file).String() != "string" {
		c.conn.Write([]byte(ReplyFileUnavailable + eol))
		return nil
	}
	_, err = c.conn.Write([]byte(
		fmt.Sprintf(ReplyFileStatus, strconv.Itoa(len(file.(string)))) + eol))
	return err
}

func handleSYST(c *client, argv string) error {
	_, err := c.conn.Write([]byte(ReplySystemType + eol))
	return err
}

func handleUSER(c *client, argv string) error {
	c.username = argv
	_, err := c.conn.Write([]byte(ReplyUserNameOK + eol))
	return err
}

func handleTYPE(c *client, argv string) error {
	c.datatype = argv
	_, err := c.conn.Write([]byte(ReplyCmdOK + eol))
	return err
}

// pathToSlice slices a path string to a slice of strings.
// e.g. "/home/example/aaa" -> ["home", "example", "aaa"]
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

// handleConnection handles a connection with an individual client.
func handleConnection(conn net.Conn) {
	// Open additional port for data connection.
	dataSock, err := net.Listen("tcp4", ":0")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dataSock.Close()
	c := client{
		conn:     conn,
		dataSock: dataSock,
		pwd:      "/",
	}
	c.conn.SetDeadline(time.Now().Add(time.Minute))
	defer c.conn.Close()
	c.conn.Write([]byte(ReplyServiceReady + eol))

	for {
		netData, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Print(string(netData))

		line := strings.SplitN(strings.TrimSpace(string(netData)), " ", 2)
		// If the command is without arguments
		if len(line) == 1 {
			line = append(line, "")
		}
		cmd, argv := line[0], line[1]
		if handler, ok := handlers[cmd]; ok {
			handler(&c, argv)
		} else {
			c.conn.Write([]byte(ReplyNotImplemented + eol))
		}

		if c.toClose {
			c.conn.Close()
			return
		}
	}
}

// ListenAndServe listens on a TCP address and handles incoming
// connections.
func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp4", s.Addr)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return err
		}
		go handleConnection(c)
	}
}
