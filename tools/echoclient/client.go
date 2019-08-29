package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

var wg sync.WaitGroup
var conf config

func main() {
	conf = getConfig()
	numConnections := 20
	for {
		if true {
			wg.Add(numConnections)
			for i := 0; i < numConnections; i++ {
				go talk()
			}
			wg.Wait()
			time.Sleep(time.Second)
		}
	}

	time.Sleep(time.Second)

	func() {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", conf.Port))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		reader := bufio.NewReader(conn)
		bytes := make([]byte, 100)
		var bytesRead int
		conn.Write([]byte("Hi"))
		time.Sleep(time.Millisecond * 100)
		bytesRead, err = reader.Read(bytes)
		if err != nil {
			fmt.Println("!!!Disconnecting Socket!!!")
			conn.Close()
			return
		}
		fmt.Println(string(bytes[:bytesRead]))
		time.Sleep(time.Millisecond * 100)
		conn.Write([]byte("What Up?"))
		time.Sleep(time.Millisecond * 100)
		bytesRead, err = reader.Read(bytes)
		if err != nil {
			fmt.Println("!!!Disconnecting Socket!!!")
			conn.Close()
			return
		}
		fmt.Println(string(bytes[:bytesRead]))
		time.Sleep(time.Millisecond * 100)
		conn.Write([]byte("The Ceiling"))
		time.Sleep(time.Millisecond * 100)
		bytesRead, err = reader.Read(bytes)
		if err != nil {
			fmt.Println("!!!Disconnecting Socket!!!")
			conn.Close()
			return
		}
		time.Sleep(time.Millisecond * 100)
		fmt.Println(string(bytes[:bytesRead]))
		time.Sleep(time.Millisecond * 1000)
		fmt.Println("---Closing Connection---")
		conn.Close()
	}()

}

func talk() {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", conf.Port))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	reader := bufio.NewReader(conn)
	bytes := make([]byte, 100)
	var bytesRead int
	conn.Write([]byte("Hi"))
	time.Sleep(time.Millisecond * 100)
	bytesRead, err = reader.Read(bytes)
	if err != nil {
		fmt.Println("!!!Disconnecting Socket!!!")
		conn.Close()
		wg.Done()
		return
	}
	fmt.Println(string(bytes[:bytesRead]))
	time.Sleep(time.Millisecond * 100)
	conn.Write([]byte("What Up?"))
	time.Sleep(time.Millisecond * 100)
	bytesRead, err = reader.Read(bytes)
	if err != nil {
		fmt.Println("!!!Disconnecting Socket!!!")
		conn.Close()
		wg.Done()
		return
	}
	fmt.Println(string(bytes[:bytesRead]))
	time.Sleep(time.Millisecond * 100)
	conn.Write([]byte("Not Much"))
	time.Sleep(time.Millisecond * 100)
	bytesRead, err = reader.Read(bytes)
	if err != nil {
		fmt.Println("!!!Disconnecting Socket!!!")
		conn.Close()
		wg.Done()
		return
	}
	time.Sleep(time.Millisecond * 100)
	fmt.Println(string(bytes[:bytesRead]))
	time.Sleep(time.Millisecond * 1000)
	fmt.Println("---Closing Connection---")
	conn.Close()
	wg.Done()
}

type config struct {
	Port int
}

func getConfig() config {
	if len(os.Args) < 2 {
		return config{
			Port: 8081,
		}
	}
	var port int
	var err error
	if port, err = strconv.Atoi(os.Args[1]); err != nil {
		fmt.Printf("%v is not a valid port number", os.Args[1])
		os.Exit(1)
	}

	return config{
		Port: port,
	}
}
