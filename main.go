package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BrandonBentley/relay/server"
)

func main() {
	conf := getConfig()
	if conf.BufferSize != 0 {
		server.BufferSize = conf.BufferSize
	}
	if conf.StartingPort != 0 {
		server.StartingPortNumber = conf.StartingPort
	}
	server.StartServer(conf.Port)
}

type config struct {
	Port         int
	BufferSize   int
	StartingPort int
}

func getConfig() config {
	if len(os.Args) < 2 {
		fmt.Println("Port number is required for execution")
		os.Exit(1)
	}
	var port int
	var err error
	if port, err = strconv.Atoi(os.Args[1]); err != nil {
		fmt.Printf("%v is not a valid port number", os.Args[1])
		os.Exit(1)
	}
	conf := config{
		Port: port,
	}
	if len(os.Args) > 3 {
		skip := false
		flagArgs := os.Args[2:]
		length := len(flagArgs)
		for i, v := range flagArgs {
			if skip {
				skip = false
				continue
			}
			if v == "-b" || v == "--buffersize" {
				if i+1 < length {
					if strings.Contains(flagArgs[i+1], "-") {
						fmt.Println("buffer size (int) missing")
						os.Exit(1)
					}
					if conf.BufferSize, err = strconv.Atoi(flagArgs[i+1]); err != nil {
						fmt.Println(flagArgs[i+1], " is an invalid buffer size")
						os.Exit(1)
					}
				}
				skip = true
			} else if v == "-p" || v == "--startingport" {
				if i+1 < length {
					if strings.Contains(flagArgs[i+1], "-") {
						fmt.Println("starting port (int) missing")
						os.Exit(1)
					}
					if conf.StartingPort, err = strconv.Atoi(flagArgs[i+1]); err != nil {
						fmt.Println(flagArgs[i+1], " is an invalid starting port")
						os.Exit(1)
					}
				}
				skip = true
			}
		}
	}
	fmt.Println(conf)
	return conf
}
