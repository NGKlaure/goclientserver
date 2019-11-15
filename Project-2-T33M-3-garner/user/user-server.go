package main

import (
	"bufio"
	"fmt"
	"net"
)

func handleConnection(c net.Conn) {

	// run loop forever (or until ctrl-c)
	for {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		// output message received
		fmt.Print("The user connected is:", string(message))

	}
	c.Close()

}

func main() {
	fmt.Println("Launching server...")

	//listen on all interfaces
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer ln.Close()
	for {
		// accept connection on port
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(conn)
	}

}
