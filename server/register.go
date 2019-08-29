package server

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// registerNewRelayConnections Handles the connection from
// client application that are registering to have their traffic
// relayed.
func registerNewRelayConnections(ln net.Listener) {
	port := StartingPortNumber
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("FATAL: Failed to Accept Connection: %v\n", err)
			break
		}
		for i := 0; i < RelayPortRetryAttempts; i++ {
			rln, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
			if err == nil {
				rln.Close()
				break
			} else {
				if strings.Contains(err.Error(), "Only one usage of each socket address") {
					port++
					continue
				}
			}
		}
		if handleRelayConnection(conn, port) {
			port++
		}
	}
}

// handleRelayConnection establishes a new connection with relayed service
// returns true if connection was a newly registered service.
func handleRelayConnection(conn net.Conn, port int) bool {
	c := NewConnection(conn)
	var relayPort int

	c.Conn.SetReadDeadline(time.Now().Add(time.Second))
	relayAddress, err := c.Reader.ReadString('\n')
	c.Conn.SetReadDeadline(time.Time{})
	isTimeoutError := false
	if err != nil {
		if strings.Contains(err.Error(), "i/o timeout") {
			isTimeoutError = true
		} else if strings.Contains(err.Error(), "forcibly closed by the remote host") {
			return false
		} else {
			fmt.Printf("++++++++++++++++++++++ %T : %v\n", err, err)
			c.Close()
			return false
		}
	}

	if isTimeoutError {
		makeConnectionChannel(port)
		relay := NewRelayService(c, port)
		err := relay.Start()
		if err != nil {
			fmt.Printf("Failed to Start Relay Service:: %v :: %v", port, err)
			c.Conn.Close()
			return false
		}
		fmt.Printf("Successfully Accepted Relay Client Connection: %v\n", fmt.Sprintf("%v:%v", relay.address, relay.port))
		return true
	} else {
		relayAddress = strings.ReplaceAll(relayAddress, "\n", "")
		relayPort, err = strconv.Atoi(relayAddress)
		if err != nil {
			c.Close()
			return false
		}
		if !pushToChannel(relayPort, c) {
			conn.Close()
		}
	}
	return false
}

// relayService is the struct used to contain the
// information need to start a relayed service.
type relayService struct {
	serviceConn      Connection
	port             int
	address          string
	connectionString string
	shutdown         chan bool
}

// NewRelayService sets up and returns a new relayService
func NewRelayService(conn Connection, port int) relayService {
	address := conn.Conn.LocalAddr().(*net.TCPAddr).IP.String()
	return relayService{
		serviceConn:      conn,
		port:             port,
		address:          address,
		connectionString: fmt.Sprintf("%v:%v", address, port),
		shutdown:         make(chan bool, 0),
	}
}

func (r *relayService) ShutDown() {
	close(r.shutdown)
	time.Sleep(time.Millisecond * 500)
}

// Start sets up the attached relayService to listen and handle
// coupling of client connections with the relayed service
// connections. Returns immediately after setting up listener.
// Also if the connection to the service is disconnected the
// listener shuts down.
func (r *relayService) Start() error {
	ConnectionMonitor.Add(r.port)
	rln, err := net.Listen("tcp", fmt.Sprintf(":%v", r.port))
	if err != nil {
		return err
	}
	r.serviceConn.Write([]byte(fmt.Sprintf("%v\n", r.port)))
	go func() {
		bytes := make([]byte, 1)
		for {
			_, err := r.serviceConn.Reader.Read(bytes)
			if err != nil {
				deleteConnectionChannel(r.port)
				rln.Close()
				break
			}
		}
		r.ShutDown()
		ConnectionMonitor.Delete(r.port)
		fmt.Printf("Relay Client: %v Disconnected\n", fmt.Sprintf("%v:%v", r.address, r.port))
	}()
	go func() {
		for {
			conn, err := rln.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "closed network connection") {
					break
				}
				fmt.Printf("Port %v: Failed to Accept Connection: %v\n", r.port, err)
			}
			clientConnection := NewConnection(conn)
			serverConnection, err := r.RequestNewConnection()
			if err != nil {
				fmt.Printf("Port %v: Failed to Get Server Connection: %v\n", r.port, err)
				clientConnection.Close()
				continue
			}
			coupler := NewConnectionCoupler(serverConnection, clientConnection, r.port)
			coupler.Couple(r.shutdown)
		}
	}()
	return nil
}

// RequestNewConnection signals the relayed client application
// to provide another Connection and returns the connection.
func (r *relayService) RequestNewConnection() (Connection, error) {
	err := r.serviceConn.Write([]byte("\n"))
	if err != nil {
		fmt.Println("-----------------------ERROR")
		os.Exit(3)
	}
	timer := time.NewTimer(time.Millisecond * 1000)

	var conn Connection
	channel, _ := getConnectionChannel(r.port)
	select {
	case <-timer.C:
		return Connection{}, errors.New("Connection Timeout")
	case conn = <-channel:
	}
	return conn, nil
}
