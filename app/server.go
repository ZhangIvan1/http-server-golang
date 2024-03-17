package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Bind to port 4221
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	// Accept a TCP connection
	connection, err := listener.Accept()
	if err != nil {
		log.Fatalln("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	defer connection.Close()

	fmt.Println("New connection from: ", connection.RemoteAddr().String())

	// Read data from the connection
	buffer := make([]byte, 0, 4096)
	_, err = connection.Read(buffer)
	if err != nil {
		log.Fatalln("Error reading data: ", err.Error())
	}

	//Respond with HTTP/1.1 200 OK\r\n\r\n
	_, err = connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	if err != nil {
		log.Fatalln("Error writing output: ", err.Error())
	}
}
