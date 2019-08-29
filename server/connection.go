package server

import (
	"bufio"
	"net"
	"sync"
)

var connectionChannels map[int]chan Connection
var connectionMapMutex sync.Mutex
var BufferSize = 4096

var StartingPortNumber = 8081

var RelayPortRetryAttempts = 20
var ConnectionChannelSize = 10

var MonitorPort = 8079
var ConnectionMonitor *Monitor

func getConnectionChannel(port int) (chan Connection, bool) {
	connectionMapMutex.Lock()
	channel, ok := connectionChannels[port]
	connectionMapMutex.Unlock()
	return channel, ok
}
func pushToChannel(port int, con Connection) bool {
	connectionMapMutex.Lock()
	if channel, ok := connectionChannels[port]; ok && channel != nil {
		channel <- con
	} else {
		return false
	}
	connectionMapMutex.Unlock()
	return true
}

func makeConnectionChannel(port int) {
	connectionMapMutex.Lock()
	connectionChannels[port] = make(chan Connection, ConnectionChannelSize)
	connectionMapMutex.Unlock()
}

func deleteConnectionChannel(port int) {
	connectionMapMutex.Lock()
	if channel, ok := connectionChannels[port]; ok && channel != nil {
		close(channel)
	}
	delete(connectionChannels, port)
	connectionMapMutex.Unlock()
}

func init() {
	connectionChannels = map[int]chan Connection{}
	connectionMapMutex = sync.Mutex{}
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
func (cc *ConnectionCoupler) Couple(shutdown chan bool) {
	ConnectionMonitor.channelMutex.Lock()
	channel := ConnectionMonitor.connectionCountChannels[cc.port]
	if channel != nil {
		channel <- 1
	}
	ConnectionMonitor.channelMutex.Unlock()
	cc.wg.Add(2)
	disconnect := make(chan bool, 0)
	go func() {
		select {
		case <-shutdown:
			cc.ServerConn.Close()
		case <-disconnect:
		}

	}()
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
		ConnectionMonitor.channelMutex.Lock()
		channel := ConnectionMonitor.connectionCountChannels[cc.port]
		ConnectionMonitor.channelMutex.Unlock()
		if channel != nil {
			channel <- -1
		}
		close(disconnect)
	}()
}
