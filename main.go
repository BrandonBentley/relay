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
	setup(conf)

	err := server.StartServer(conf.Port)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setup(conf config) {
	if conf.BufferSize != 0 {
		server.BufferSize = conf.BufferSize
	}
	if conf.StartingPort != 0 {
		server.StartingPortNumber = conf.StartingPort
	}
	if conf.RetryAmount != 0 {
		server.RelayPortRetryAttempts = conf.RetryAmount
	}
	if conf.MonitoringPort != 0 {
		server.MonitorPort = conf.MonitoringPort
	}
}

type config struct {
	Port           int
	BufferSize     int
	StartingPort   int
	RetryAmount    int
	MonitoringPort int
}

func getConfig() config {
	if len(os.Args) < 2 {
		fmt.Println("Port number is required for execution")
		printHelp(1)
		os.Exit(1)
	}
	if os.Args[1] == "-h" || os.Args[1] == "help" || os.Args[1] == "--help" {
		printHelp(0)
	}
	var port int
	var err error
	if port, err = strconv.Atoi(os.Args[1]); err != nil {
		fmt.Printf("%v is not a valid port number", os.Args[1])
		printHelp(1)
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
						printHelp(1)
					}
					if conf.BufferSize, err = strconv.Atoi(flagArgs[i+1]); err != nil {
						fmt.Println(flagArgs[i+1], " is an invalid buffer size")
						printHelp(1)
					}
				}
				skip = true
			} else if v == "-p" || v == "--startingport" {
				if i+1 < length {
					if strings.Contains(flagArgs[i+1], "-") {
						fmt.Println("starting port (int) missing")
						printHelp(1)
					}
					if conf.StartingPort, err = strconv.Atoi(flagArgs[i+1]); err != nil {
						fmt.Println(flagArgs[i+1], " is an invalid starting port")
						printHelp(1)
					}
				}
				skip = true
			} else if v == "-r" || v == "--retryattempts" {
				if i+1 < length {
					if strings.Contains(flagArgs[i+1], "-") {
						fmt.Println("retryattempts (int) missing")
						printHelp(1)
					}
					if conf.RetryAmount, err = strconv.Atoi(flagArgs[i+1]); err != nil {
						fmt.Println(flagArgs[i+1], " is an invalid retry amount")
						printHelp(1)
					}
				}
				skip = true
			} else if v == "-m" || v == "--monitorport" {
				if i+1 < length {
					if strings.Contains(flagArgs[i+1], "-") {
						fmt.Println("monitorPort (int) missing")
						printHelp(1)
					}
					if conf.MonitoringPort, err = strconv.Atoi(flagArgs[i+1]); err != nil {
						fmt.Println(flagArgs[i+1], " is an invalid monitor port number")
						printHelp(1)
					}
				}
				skip = true
			} else {
				fmt.Printf("%v is not a valid option", v)
				printHelp(1)
			}
		}
	}
	return conf
}

func printHelp(exit int) {
	fmt.Println("\nRelay Server")
	fmt.Println("-------------------------------------------\n")
	fmt.Println("usage: relay port {optional params}\n")
	formatString := "%-30v %v\n"
	fmt.Printf(formatString, "-b, --buffersize {size}", "set the size of each connection coupler (in bytes).")
	fmt.Printf(formatString, "-p, --startingport {port}", "set first port to be used for relayed clients.")
	fmt.Printf(formatString, "-m, --monitorport {port}", "set port for the monitoring HTTP endpoint")
	fmt.Printf(formatString, "-r, --retryattempts {count}", "set the number of retry attempts allowed when finding")
	fmt.Printf(formatString, "", " an available port for a new relay client.")

	os.Exit(exit)
}
