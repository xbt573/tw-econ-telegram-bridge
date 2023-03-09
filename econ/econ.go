package econ

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strings"
	"time"
)

var (
	ErrAlreadyConnected    = errors.New("econ: already connected")
	ErrAlreadyDisconnected = errors.New("econ: already disconnected")
	ErrDisconnected        = errors.New("econ: disconnected")
	ErrWrongPassword       = errors.New("econ: wrong password")
	serverNameRegex        = regexp.MustCompile(`(?U)Value: ([\s\S]+)\n`)
)

type ECON struct {
	Connected  bool
	Completed  chan bool
	ServerName string
	ip         string
	password   string
	port       int
	conn       net.Conn
}

func (e *ECON) Connect() error {
	if e.conn != nil {
		return ErrAlreadyConnected
	}

	conn, err := net.Dial("tcp", fmt.Sprint(e.ip, ":", e.port))
	if err != nil {
		return err
	}

	_, err = conn.Read(make([]byte, 1024))
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte(e.password + "\n"))
	if err != nil {
		return err
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return err
	}

	if !strings.Contains(string(buffer[:n]), "Authentication successful") {
		return ErrWrongPassword
	}

	conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))

	for {
		_, err = conn.Read(make([]byte, 1024))
		if err != nil {
			break
		}
	}

	conn.SetReadDeadline(time.Time{})

	_, err = conn.Write([]byte("sv_name\n"))
	if err != nil {
		return err
	}

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return err
		}

		if strings.Contains(string(buffer[:n]), "Value") {
			break
		}
	}

	match := serverNameRegex.FindStringSubmatch(string(buffer[:n]))
	if len(match) > 0 {
		e.ServerName = match[1]
	}

	e.conn = conn
	e.Connected = true
	return nil
}

func (e *ECON) Disconnect() error {
	if e.conn == nil {
		return ErrAlreadyDisconnected
	}

	e.conn.Close()
	e.conn = nil
	e.Connected = false
	e.Completed <- true
	return nil
}

func (e *ECON) Read() (string, error) {
	if e.conn == nil {
		return "", ErrDisconnected
	}

	buffer := make([]byte, 1024)
	n, err := e.conn.Read(buffer)
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Println("EOF reached, trying to disconnect...")
			err := e.Disconnect()
			if err != nil {
				return "", err
			}
			return "", nil
		}
		return "", err
	}

	return string(buffer[:n]), nil
}

func (e *ECON) Write(message string) error {
	if e.conn == nil {
		return ErrDisconnected
	}

	_, err := e.conn.Write([]byte(message + "\n"))
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Println("EOF reached, trying to disconnect...")
			err := e.Disconnect()
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	return nil
}

func (e *ECON) Send(message string) error {
	return e.Write(fmt.Sprintf("say \"%v\"", message))
}

func NewECON(ip, password string, port int) *ECON {
	return &ECON{
		ip:        ip,
		password:  password,
		port:      port,
		Completed: make(chan bool),
	}
}
