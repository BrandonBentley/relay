package server

import (
	"fmt"
	"net"
)

// StartServer initializes the monitoring endpoint
// and starts the relay service on the provided
// port number.
func StartServer(port int) error {
	ConnectionMonitor = NewMonitor(MonitorPort)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return fmt.Errorf("Failed to Start Relay Server: %v", err)
	}
	defer ln.Close()
	fmt.Printf("Relay Server Successfully Started: %v\n", ln.Addr().String())
	registerNewRelayConnections(ln)
	return nil
}
