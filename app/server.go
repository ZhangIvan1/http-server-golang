package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
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
	for true {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatalln("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		handleConnection(connection)
	}
}

func handleConnection(conn net.Conn) error {
	defer conn.Close()

	fmt.Println("New connection from: ", conn.RemoteAddr().String())

	// Read data from the connection
	readBuffer := make([]byte, 0, 4096)
	_, err := conn.Read(readBuffer)
	if err != nil {
		log.Fatalln("Error reading data: ", err.Error())
	}

	//Respond with HTTP/1.1 200 OK\r\n\r\n
	_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	if err != nil {
		log.Fatalln("Error writing output: ", err.Error())
	}

	readArray := strings.Split(string(readBuffer), " ")
	fmt.Println(readArray)

	// Otherwise, need to respond with a 404 Not Found response.
	write := []byte("HTTP/1.1 404 Not Found\r\n\r\n")
	if readArray[1] == "/" {
		// If the path is '/', need to respond with a 200 OK response
		write = []byte("HTTP/1.1 200 OK\r\n\r\n")
	}

	conn.Write(write)
	return nil
}
