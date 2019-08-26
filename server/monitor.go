package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Monitor is used to see what ports are currently listening
// as well as how many active connections are on each port.
type Monitor struct {
	connectionCountChannels map[int]chan int
	connectionCounts        map[int]int
}

// NewMonitor starts an HTTP server that returns a json
// representation of the active relay ports and the
// client connection count for each.
func NewMonitor(port int) *Monitor {
	mon := &Monitor{
		connectionCountChannels: map[int]chan int{},
		connectionCounts:        map[int]int{},
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bytes, err := json.MarshalIndent(mon.connectionCounts, "", "    ")
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("ERROR OCCURED: %v", err)))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	})
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
			fmt.Printf("Failed to Start Monitor Service: %v", err)
			os.Exit(5)
		}
	}()
	return mon
}

// Add sets up to monitor the port passed in
func (m *Monitor) Add(port int) {
	if _, ok := m.connectionCountChannels[port]; !ok {
		m.connectionCountChannels[port] = make(chan int)
	}

	if _, ok := m.connectionCounts[port]; !ok {
		m.connectionCounts[port] = 0
	}

	go func() {
		for {
			if _, ok := m.connectionCountChannels[port]; ok {
				change := <-m.connectionCountChannels[port]
				if change == 0 {
					break
				}
				m.connectionCounts[port] = m.connectionCounts[port] + change

			} else {
				break
			}

		}
	}()
}

// Delete Removes the provided port from being monitored
func (m *Monitor) Delete(port int) {
	if m.connectionCountChannels[port] != nil {
		close(m.connectionCountChannels[port])
	}
	delete(m.connectionCounts, port)
	delete(m.connectionCountChannels, port)
}
