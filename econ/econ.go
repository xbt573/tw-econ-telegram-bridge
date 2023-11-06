package econ

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrAlreadyConnected    = errors.New("econ: already connected")
	ErrAlreadyDisconnected = errors.New("econ: already disconnected")
	ErrDisconnected        = errors.New("econ: disconnected")
	ErrWrongPassword       = errors.New("econ: wrong password")
)

type ECON struct {
	connected bool
	ip        string
	port      string
	password  string

	conn net.Conn
}

type ECONOpts struct {
	Ip       string
	Port     uint16
	Password string
}

func NewECON(opts ECONOpts) (*ECON, error) {
	return &ECON{
		ip:       opts.Ip,
		password: opts.Password,
		port:     strconv.Itoa(int(opts.Port)),
	}, nil
}

func (e *ECON) Connected() bool {
	return e.connected
}

func (e *ECON) Connect() error {
	if e.connected {
		return ErrAlreadyConnected
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(e.ip, e.port))
	if err != nil {
		return err
	}

	// read out useless info
	conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	_, err = conn.Read(make([]byte, 1024))
	if err != nil && !os.IsTimeout(err) {
		conn.Close()
		return err
	}
	conn.SetReadDeadline(time.Time{})

	_, err = conn.Write([]byte(e.password + "\n"))
	if err != nil {
		conn.Close()
		return err
	}

	conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && !os.IsTimeout(err) {
		conn.Close()
		return err
	}
	conn.SetReadDeadline(time.Time{})

	if !strings.Contains(string(buf[:n]), "Authentication successful") {
		return ErrWrongPassword
	}

	e.conn = conn
	e.connected = true

	return nil
}

func (e *ECON) Disconnect() error {
	if !e.connected {
		return ErrAlreadyDisconnected
	}

	err := e.conn.Close()
	if err != nil {
		return err
	}

	e.conn = nil
	e.connected = false

	return nil
}

func (e *ECON) Write(buf []byte) error {
	if !e.connected {
		return ErrDisconnected
	}

	_, err := e.conn.Write(append(buf, '\n'))
	if err != nil {
		return err
	}

	return nil
}

func (e *ECON) Read() ([]byte, error) {
	// "ping" socket
	if err := e.Write([]byte{}); err != nil {
		return []byte{}, err
	}

	buffer := make([]byte, 8192)
	n, err := e.conn.Read(buffer)
	if err != nil {
		return []byte{}, err
	}

	return buffer[:n], nil
}

func (e *ECON) Message(message string) error {
	if arr := strings.Split(message, "\n"); len(arr) > 1 {
		err := e.Write([]byte(fmt.Sprintf("say \"%v\"", arr[0])))
		if err != nil {
			return err
		}

		for _, x := range arr[1:] {
			err := e.Write([]byte(fmt.Sprintf("say \"> %v\"", x)))
			if err != nil {
				return err
			}
		}

		return nil
	}

	return e.Write([]byte(fmt.Sprintf("say \"%v\"", message)))
}
