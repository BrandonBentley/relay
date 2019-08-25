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

func registerNewRelayConnections(ln net.Listener) {
	port := StartingPortNumber
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("FATAL: Failed to Accept Connection: %v\n", err)
			break
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

	timer := time.NewTimer(time.Millisecond * 1000)

	ch := make(chan bool, 0)
	var relayPort int

	go func() {
		relayAddress, err := c.Reader.ReadString('\n')
		if err != nil {
			ch <- false
		}
		relayAddress = strings.ReplaceAll(relayAddress, "\n", "")
		relayPort, err = strconv.Atoi(relayAddress)
		if err != nil {
			ch <- false
		}
		ch <- true
	}()

	select {
	case read := <-ch:
		if read {
			if connectionChannel, ok := connectionChannels[relayPort]; ok {
				connectionChannel <- c
			} else {
				fmt.Println("MISSED IT BY THAT MUCH")
				conn.Close()
			}
		} else {
			conn.Close()
		}
		return false
	case <-timer.C:
		channel := make(chan Connection, ConnectionChannelSize)
		connectionChannels[port] = channel
		relay := NewRelayService(c, port)
		err := relay.Start()
		if err != nil {
			fmt.Println("%v :: %v", port, err)
		}
		c.Conn.Write([]byte(fmt.Sprintf("%v\n", port)))
		fmt.Printf("Successfully Accepted Relay Client Connection: %v\n", fmt.Sprintf("%v:%v", relay.address, relay.port))
	}
	return true
}

type relayService struct {
	serviceConn      Connection
	port             int
	address          string
	connectionString string
}

func NewRelayService(conn Connection, port int) relayService {
	address := conn.Conn.LocalAddr().(*net.TCPAddr).IP.String()
	return relayService{
		serviceConn:      conn,
		port:             port,
		address:          address,
		connectionString: fmt.Sprintf("%v:%v", address, port),
	}
}

func (r *relayService) Start() error {
	rln, err := net.Listen("tcp", fmt.Sprintf(":%v", r.port))
	if err != nil {
		fmt.Println("Failed to Start Relay Service: %v", err)
		return err
	}
	go func() {
		bytes := make([]byte, 1)
		for {
			_, err := r.serviceConn.Reader.Read(bytes)
			if err != nil {
				close(connectionChannels[r.port])
				delete(connectionChannels, r.port)
				rln.Close()
				break
			}
		}
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
			coupler := NewConnectionCoupler(serverConnection, clientConnection)
			coupler.Couple()
		}

	}()
	return nil
}

func (r *relayService) RequestNewConnection() (Connection, error) {
	err := r.serviceConn.Write([]byte("\n"))
	if err != nil {
		fmt.Println("-----------------------ERROR")
		os.Exit(3)
	}
	timer := time.NewTimer(time.Millisecond * 1000)

	var conn Connection
	select {
	case <-timer.C:
		return Connection{}, errors.New("Connection Timeout")
	case conn = <-connectionChannels[r.port]:
	}
	return conn, nil
}
