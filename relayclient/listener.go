// Package relayclient
// Implements the net.Listener interface for easy use with the relay server.
// exists here for the sole purpose of being imported by the test(s) to allow for automated testing of the relay server.
package relayclient

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Listener struct {
	RelayConnection       net.Conn
	RelayConnectionString string
	RelayPort             int
	relayReader           *bufio.Reader
	relayedPath           string
	closed                bool
	ReceivedString        string
}

func ListenRelay(host string, port int) (*Listener, error) {
	relayPath := fmt.Sprintf("%v:%v", host, port)
	conn, err := net.Dial("tcp", relayPath)
	if err != nil {
		return nil, err
	}
	relayReader := bufio.NewReader(conn)
	relayAddress, err := relayReader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	relayPort, err := strconv.Atoi(strings.ReplaceAll(relayAddress, "\n", ""))
	if err != nil {
		return nil, err
	}
	fmt.Printf("established relay address: %v:%v", host, relayAddress)
	return &Listener{
		RelayConnection:       conn,
		RelayConnectionString: relayPath,
		RelayPort:             relayPort,
		relayReader:           relayReader,
		relayedPath:           relayAddress,
		closed:                false,
	}, nil
}

func (l *Listener) Accept() (net.Conn, error) {
	if l.closed {
		return nil, errors.New("Relay Connection is Closed")
	}
	_, err := l.relayReader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	conn, err := net.Dial("tcp", l.RelayConnectionString)
	if err != nil {
		return nil, err
	}
	conn.Write([]byte(l.relayedPath + "\n"))
	return conn, nil
}

func (l *Listener) Close() error {
	l.closed = true
	return l.RelayConnection.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.RelayConnection.RemoteAddr()
}
