package server

import (
	"fmt"
	"net"
	"os"
)

func StartServer(port int) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		fmt.Printf("Failed to Start Relay Server: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Printf("Relay Server Successfully Started: %v\n", ln.Addr().String())
	registerNewRelayConnections(ln)
}
