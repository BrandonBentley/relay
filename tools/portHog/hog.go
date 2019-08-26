package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	listeners := make([]net.Listener, 19)
	for i := 0; i < 19; i++ {
		listeners[i], _ = net.Listen("tcp", fmt.Sprintf(":%v", 8081+i))
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Press Enter To Exit...")
	reader.ReadString('\n')
	for _, l := range listeners {
		if l != nil {
			l.Close()
		}
	}
}
