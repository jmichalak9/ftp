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

type commandHandler func(*client, string) string

// File represents file content in a virtual file system.
type File string

// Directory represents a directory in a virtual file system.
type Directory map[string]interface{}

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

func handleCWD(c *client, argv string) string {
	_, err := getDirFromPath(argv)
	if err != nil {
		return ReplyFileUnavailable + eol
	}
	c.pwd = argv
	return ReplyFileActionOK + eol
}

func handleFEAT(c *client, argv string) string {
	return "211-Features:" + eol +
		"211 End" + eol
}

func handleLIST(c *client, argv string) string {
	dataConn, err := c.dataSock.Accept()
	if err != nil {
		return ReplyCannotOpenDataConn + eol
	}
	defer dataConn.Close()
	c.conn.Write([]byte(ReplyFileStatusOK + eol))
	dir, err := getDirFromPath(c.pwd)
	if err != nil {
		return ReplyActionAborted + eol
	}
	reply := ""
	for k, v := range dir {
		// TODO: send proper file properties
		switch v.(type) {
		case Directory:
			reply += ("drwxr-xr-x  10 kuba wheel 776 Nov 24 15:50 " + k + "\r\n")
		case File:
			reply += ("-rwxr-xr-x  10 kuba wheel 776 Nov 24 15:50 " + k + "\r\n")
		}
	}
	dataConn.Write([]byte(reply))
	return ReplyClosingDataConn + eol
}

func handleMDTM(c *client, argv string) string {
	_, err := getFileFromPath(argv)
	if err != nil {
		return ReplyFileUnavailable + eol
	}
	return "213 207001010000" + eol
}

func handlePASS(c *client, argv string) string {
	return ReplyUserLoggedIn + eol
}

func handlePASV(c *client, argv string) string {
	ip := []int{127, 0, 0, 1}
	port := c.dataSock.Addr().(*net.TCPAddr).Port
	return fmt.Sprintf(ReplyEnteringPasv,
		ip[0], ip[1], ip[2], ip[3], port/256, port%256) + eol
}

func handlePWD(c *client, argv string) string {
	return fmt.Sprintf(ReplyPathNameOK, c.pwd) + eol
}

func handleQUIT(c *client, argv string) string {
	c.toClose = true
	return ReplyClosingConn + eol
}

func handleRETR(c *client, argv string) string {
	dataConn, err := c.dataSock.Accept()
	if err != nil {
		return ReplyCannotOpenDataConn + eol
	}
	defer dataConn.Close()
	c.conn.Write([]byte(ReplyFileStatusOK + eol))
	file, err := getFileFromPath(argv)
	if err != nil {
		return ReplyActionAborted + eol
	}
	dataConn.Write([]byte(file))
	return ReplyClosingDataConn + eol
}

func handleSIZE(c *client, argv string) string {
	file, err := getFileFromPath(argv)
	if err != nil {
		return ReplyFileUnavailable + eol
	}
	return fmt.Sprintf(ReplyFileStatus, strconv.Itoa(len(file))) + eol
}

func handleSYST(c *client, argv string) string {
	return ReplySystemType + eol
}

func handleUSER(c *client, argv string) string {
	c.username = argv
	return ReplyUserNameOK + eol
}

func handleTYPE(c *client, argv string) string {
	c.datatype = argv
	return ReplyCmdOK + eol
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
		if name == "" || name == "." {
			continue
		}
		item, ok := searchDir[name]
		if !ok {
			return "", errors.New("File not found")
		}
		switch t := item.(type) {
		case File:
			return item.(File), nil
		case Directory:
			searchDir = item.(Directory)
			continue
		default:
			return nil, errors.New("unknown type: " + reflect.TypeOf(t).String())
		}
	}
	return searchDir, nil
}

func getFileFromPath(path string) (File, error) {
	item, err := getItemFromPath(path)
	if err != nil {
		return "", err
	}
	if file, ok := item.(File); ok {
		return file, nil
	}
	return "", errors.New("not a file")
}

func getDirFromPath(path string) (Directory, error) {
	item, err := getItemFromPath(path)
	if err != nil {
		return nil, err
	}
	if dir, ok := item.(Directory); ok {
		return dir, nil
	}
	return nil, errors.New("not a directory")
}

// handleConnection handles a connection with an individual client.
func handleConnection(conn net.Conn) {
	// Open additional port for data connection.
	dataSock, err := net.Listen("tcp4", ":0")
	if err != nil {
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
			c.conn.Write([]byte(handler(&c, argv)))
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
