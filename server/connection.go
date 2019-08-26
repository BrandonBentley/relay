package server

import (
	"bufio"
	"net"
	"sync"
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

// NewConnection returns a wrapped net.Conn with an
// initialized bufio.Reader.
func NewConnection(conn net.Conn) Connection {
	return Connection{conn, bufio.NewReader(conn), make([]byte, BufferSize)}
}

// Connection is the wrapper for a net.Conn that has an
// initalized bufio.Reader and a buffer to read content
// from the connection.
type Connection struct {
	Conn   net.Conn
	Reader *bufio.Reader
	Buffer []byte
}

// ReadAndWriteTo reads from the attached Connection and
// writes that data directly to the destination Connection.
func (c *Connection) ReadAndWriteTo(destination Connection) error {
	read, err := c.Reader.Read(c.Buffer)
	if err != nil {
		return err
	}
	return destination.Write(c.Buffer[:read])
}

// Write wraps the net.Conn.Write function
func (c *Connection) Write(data []byte) error {
	_, err := c.Conn.Write(data)
	return err
}

// Close wraps the net.Conn.Close() function
func (c *Connection) Close() error {
	return c.Conn.Close()
}

// ConnectionCoupler contains both client and server connections
// to be coupled together to act as a single bi-driectional connection
type ConnectionCoupler struct {
	ServerConn Connection
	ClientConn Connection
	port       int
	wg         *sync.WaitGroup
}

// NewConnectionCoupler returns a new ConnectionCoupler
// takes the server and client connections and the port number that the client
// connected to as arguments.
func NewConnectionCoupler(server, client Connection, port int) *ConnectionCoupler {
	return &ConnectionCoupler{
		ServerConn: server,
		ClientConn: client,
		port:       port,
		wg:         &sync.WaitGroup{},
	}
}

// Couple starts coupling the client and server connection together
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
