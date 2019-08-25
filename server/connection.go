package server

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"
)

var connectionChannels map[int]chan Connection
var BufferSize = 4096
var StartingPortNumber = 8081
var ConnectionChannelSize = 10

func init() {
	connectionChannels = map[int]chan Connection{}
}

func NewConnection(conn net.Conn) Connection {
	return Connection{conn, bufio.NewReader(conn), make([]byte, BufferSize)}
}

type Connection struct {
	Conn   net.Conn
	Reader *bufio.Reader
	Buffer []byte
}

func (c *Connection) ReadAndWriteTo(destination Connection) error {
	read, err := c.Reader.Read(c.Buffer)
	if err != nil {
		return err
	}
	return destination.Write(c.Buffer[:read])
}

func (c *Connection) ReadAll() ([]byte, error) {
	bytes := make([]byte, 1000)
	read, err := c.Reader.Read(bytes)
	return bytes[:read], err
}

func (c *Connection) Write(data []byte) error {
	_, err := c.Conn.Write(data)
	return err
}

func (c *Connection) Close() error {
	return c.Conn.Close()
}

func (c Connection) String() string {
	return fmt.Sprintf("{ %v, %v, %v }", c.Conn, c.Reader)
}

type ConnectionCoupler struct {
	ServerConn Connection
	ClientConn Connection
	wg         *sync.WaitGroup
}

func NewConnectionCoupler(server, client Connection) *ConnectionCoupler {
	return &ConnectionCoupler{
		ServerConn: server,
		ClientConn: client,
		wg:         &sync.WaitGroup{},
	}
}

func (cc *ConnectionCoupler) IsActive() bool {
	timer := time.NewTimer(time.Millisecond)
	ch := make(chan bool, 0)
	go func() {
		cc.wg.Wait()
		ch <- true
	}()
	select {
	case <-ch:
		return false
	case <-timer.C:
	}
	return true
}

func (cc *ConnectionCoupler) Couple() {
	cc.wg.Add(2)
	go func() {
		for {
			err := cc.ServerConn.ReadAndWriteTo(cc.ClientConn)
			if err != nil {
				cc.ServerConn.Close()
				cc.ClientConn.Close()
				break
			}
		}
		cc.wg.Done()
	}()
	go func() {
		for {
			err := cc.ClientConn.ReadAndWriteTo(cc.ServerConn)
			if err != nil {
				cc.ServerConn.Close()
				cc.ClientConn.Close()
				break
			}
		}
		cc.wg.Done()
	}()
}
