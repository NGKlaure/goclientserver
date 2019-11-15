package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {

	//connect to this socket
	conn, _ := net.Dial("tcp", "3.86.179.34:8081")
	for {
		//read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Text to send: ")
		text, _ := reader.ReadString('\n')
		// send to socket
		fmt.Fprintf(conn, text+"\n")
		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Message from server: " + message)
	}
}

//https://www.linode.com/docs/development/go/developing-udp-and-tcp-clients-and-servers-in-go/
