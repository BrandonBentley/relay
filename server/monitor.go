package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

// Monitor is used to see what ports are currently listening
// as well as how many active connections are on each port.
type Monitor struct {
	connectionCountChannels map[int]chan int
	connectionCounts        map[int]int
	addChannel              chan int
	countMutex              sync.Mutex
	channelMutex            sync.Mutex
}

// NewMonitor starts an HTTP server that returns a json
// representation of the active relay ports and the
// client connection count for each.
func NewMonitor(port int) *Monitor {
	mon := &Monitor{
		connectionCountChannels: map[int]chan int{},
		connectionCounts:        map[int]int{},
		addChannel:              make(chan int, 10),
		countMutex:              sync.Mutex{},
		channelMutex:            sync.Mutex{},
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mon.countMutex.Lock()
		bytes, err := json.MarshalIndent(mon.connectionCounts, "", "    ")
		mon.countMutex.Unlock()
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
	go mon.addSubRoutine()
	return mon
}

func (m *Monitor) addSubRoutine() {
	var port int
	var wg sync.WaitGroup
	for {
		port = <-m.addChannel
		m.channelMutex.Lock()
		if _, ok := m.connectionCountChannels[port]; !ok {
			m.connectionCountChannels[port] = make(chan int)
		}
		m.channelMutex.Unlock()
		m.countMutex.Lock()
		if _, ok := m.connectionCounts[port]; !ok {
			m.connectionCounts[port] = 0
		}
		m.countMutex.Unlock()
		wg.Add(1)
		go func() {
			localPort := port
			wg.Done()
			for {
				m.channelMutex.Lock()
				if channel, ok := m.connectionCountChannels[localPort]; ok {
					m.channelMutex.Unlock()
					change := <-channel
					if change == 0 {

						break
					}
					m.countMutex.Lock()
					m.connectionCounts[localPort] = m.connectionCounts[localPort] + change
					m.countMutex.Unlock()
				} else {
					m.channelMutex.Unlock()
					break
				}
			}
		}()
		wg.Wait()
	}
}

// Add sets up to monitor the port passed in
func (m *Monitor) Add(port int) {
	m.addChannel <- port
}

// Delete Removes the provided port from being monitored
func (m *Monitor) Delete(port int) {
	m.channelMutex.Lock()
	channel := m.connectionCountChannels[port]
	m.channelMutex.Unlock()

	m.countMutex.Lock()
	left := m.connectionCounts[port] - 6
	m.countMutex.Unlock()

	for i := 0; i < left; i++ {
		<-channel
	}
	m.channelMutex.Lock()
	close(channel)
	m.channelMutex.Unlock()
	for {
		if _, more := <-channel; !more {
			break
		}
	}

	m.channelMutex.Lock()
	delete(m.connectionCountChannels, port)
	m.channelMutex.Unlock()

	m.countMutex.Lock()
	delete(m.connectionCounts, port)
	m.countMutex.Unlock()
}
