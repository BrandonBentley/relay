package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/BrandonBentley/relay/tools/echoserver/relayclient"
)

func main() {
	conf := getConfig()

	ln, err := relayclient.ListenRelay(conf.relayhost, conf.relayport)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Failed to Accept Connection: %v\n", err)
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		bytes := make([]byte, 100)
		bytesRead, err := reader.Read(bytes)
		if err != nil {
			fmt.Println("---Socket Disconnected---")
			conn.Close()
			break
		}
		str := string(bytes[:bytesRead])
		fmt.Println(str)
		conn.Write([]byte(str))
	}
}

type config struct {
	relayhost string
	relayport int
}

func getConfig() config {
	if len(os.Args) < 3 {
		fmt.Println("Relay host and port number are required for execution")
		os.Exit(1)
	}
	var port int
	var err error
	if port, err = strconv.Atoi(os.Args[2]); err != nil {
		fmt.Printf("%v is not a valid port number", os.Args[1])
		os.Exit(1)
	}

	return config{
		relayhost: os.Args[1],
		relayport: port,
	}
}
