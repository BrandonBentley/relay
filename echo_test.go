package main_test

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/BrandonBentley/relay/relayclient"
	"github.com/BrandonBentley/relay/server"
)

func TestEchoServer(t *testing.T) {
	port := 9000
	relayhost := "localhost"
	server.StartingPortNumber = 9001
	expectedMessage := "Hello, world"
	go func() {
		server.StartServer(port)
	}()
	time.Sleep(time.Millisecond * 200)

	//Start Echo Server Relay Client
	ln, err := relayclient.ListenRelay(relayhost, port)
	if err != nil {
		t.Fatalf("Failed to start Echo server: %v", err)
	}
	if ln.RelayPort != 9001 {
		t.Fatalf("Relay Port Mismatch. Expected %v got %v ", 9001, ln.RelayPort)
	}
	wg := sync.WaitGroup{}
	var tc struct {
		ReceivedString string
	}
	wg.Add(1)
	//Start Echo Server
	go func() {
		runTestEchoServer(ln, t)
		wg.Done()
	}()
	time.Sleep(time.Millisecond * 1000)

	//Start Echo Client
	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", relayhost, 9001))
	if err != nil {
		t.Fatalf("Echo Client Failed to Establish Connection: %v", err)
	}
	reader := bufio.NewReader(conn)
	bytes := make([]byte, 100)
	var bytesRead int

	conn.Write([]byte(expectedMessage))

	bytesRead, err = reader.Read(bytes)
	if err != nil {
		t.Fatalf("Echo Client Failed to Read From Connection: %v", err)
	}
	tc.ReceivedString = string(bytes[:bytesRead])
	conn.Close()
	//end Echo client

	if ln.ReceivedString != expectedMessage {
		t.Errorf(`Echo Server ReceivedString Mismatch. Expected "%v" got "%v"`, expectedMessage, ln.ReceivedString)
	}

	if tc.ReceivedString != expectedMessage {
		t.Errorf(`Echo Server ReceivedString Mismatch. Expected "%v" got "%v"`, expectedMessage, tc.ReceivedString)
	}
}

func runTestEchoServer(ln *relayclient.Listener, t *testing.T) {
	firstTime := true
	for {
		conn, err := ln.Accept()
		if err != nil {
			if firstTime {
				t.Fatalf("Echo Server Failed to Establish Connection: %v", err)
			} else {
				conn.Close()
				ln.Close()
				break
			}
		}
		firstTime = false

		reader := bufio.NewReader(conn)
		bytes := make([]byte, 100)
		bytesRead, err := reader.Read(bytes)
		ln.ReceivedString = string(bytes[:bytesRead])
		conn.Write(bytes[:bytesRead])
	}
}
