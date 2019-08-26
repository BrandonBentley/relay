package relayclient

import (
	"bufio"
	"errors"
	"fmt"
	"net"
)

type Listener struct {
	relayConnection       net.Conn
	relayConnectionString string
	relayReader           *bufio.Reader
	relayedPath           string
	closed                bool
}

func ListenRelay(host string, port int) (net.Listener, error) {
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
	fmt.Printf("established relay address: %v:%v", host, relayAddress)
	return &Listener{
		relayConnection:       conn,
		relayConnectionString: relayPath,
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
		return nil, err
	}
	conn, err := net.Dial("tcp", l.relayConnectionString)
	if err != nil {
		return nil, err
	}
	conn.Write([]byte(l.relayedPath))
	return conn, nil
}

func (l *Listener) Close() error {
	l.closed = true
	return l.relayConnection.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.relayConnection.RemoteAddr()
}
