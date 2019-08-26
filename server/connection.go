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
var RelayPortRetryAttempts = 20
var ConnectionChannelSize = 10

var MonitorPort = 8079
var ConnectionMonitor *Monitor

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
	read, err := c.Reader.Read(c.Buffer)
	return c.Buffer[:read], err
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
	port       int
	wg         *sync.WaitGroup
}

func NewConnectionCoupler(server, client Connection, port int) *ConnectionCoupler {
	return &ConnectionCoupler{
		ServerConn: server,
		ClientConn: client,
		port:       port,
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
	ConnectionMonitor.connectionCountChannels[cc.port] <- 1
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
	go func() {
		cc.wg.Wait()
		ConnectionMonitor.connectionCountChannels[cc.port] <- -1
	}()
}
