package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/BrandonBentley/relay/server"
)

func main() {
	conf := getConfig()

	server.StartServer(conf.Port)
}

type config struct {
	Port int
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

	return config{
		Port: port,
	}
}
