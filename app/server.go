package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	netType string = "tcp"
	host    string = "0.0.0.0"
	port    string = "4221"
)

type headers map[string]string
type request struct {
	Method  string
	Path    string
	Version string
	Headers headers
	Body    []byte
}

const (
	CRLF             = "\r\n"
	STATUS_OK        = "HTTP/1.1 200 OK"
	STATUS_NOT_FOUND = "HTTP/1.1 404 Not Found"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Bind to port 4221
	listener, err := net.Listen(netType, host+":"+port)
	if err != nil {
		fmt.Println("Failed to bind to port" + port)
		os.Exit(1)
	}

	// Accept a TCP connection
	for true {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatalln("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		if err := handleConnection(connection); err != nil {
			log.Fatalln("Error handling connection", err.Error())
			os.Exit(1)
		}
	}
}

func handleConnection(conn net.Conn) error {
	defer conn.Close()

	fmt.Println("New connection from: ", conn.RemoteAddr().String())

	req, err := buildRequest(conn)
	if err != nil {
		fmt.Println("Error reading data: ", err.Error())
		os.Exit(1)
	}

	if err := handleResponse(conn, req); err != nil {
		fmt.Println("Error writing output: ", err.Error())
		os.Exit(1)
	}

	return nil
}

func buildRequest(conn net.Conn) (req request, err error) {
	buffer := make([]byte, 4096)
	_, err = conn.Read(buffer)
	if err != nil {
		return req, err
	}

	parts := strings.Split(string(buffer), "\r\n")
	if len(parts) == 0 {
		return req, errors.New("HTTP startline missing")
	}
	startLine := parts[0]
	headers := parts[1:]
	err = setRequestPath(startLine, &req)
	if err != nil {
		return req, err
	}
	setHeaders(headers, &req)

	return req, nil
}

//GET /index.html HTTP/1.1
func setRequestPath(line string, req *request) error {
	parts := strings.Split(line, " ")

	if len(parts) != 3 {
		return errors.New("the HTTP startline should include three part like: GET /index.html HTTP/1.1")
	}

	req.Method, req.Path, req.Version = parts[0], parts[1], parts[2]
	fmt.Println(req.Method, " ", req.Path, " ", req.Version)
	return nil
}

//Host: localhost:4221
//User-Agent: curl/7.64.1
func setHeaders(headerLines []string, req *request) error {
	if req.Headers == nil {
		req.Headers = make(headers, len(headerLines))
	}

	for _, line := range headerLines {
		splittedLine := strings.Split(line, "")
		if len(splittedLine) == 2 {
			req.Headers[splittedLine[0]] = splittedLine[1]
			fmt.Println(splittedLine[0], " ", splittedLine[1])
		}
	}

	return nil
}

func handleResponse(conn net.Conn, req request) error {
	switch {
	case req.Path == "/":
		if _, err := conn.Write([]byte(STATUS_OK + CRLF + CRLF)); err != nil {
			return err
		}
	case strings.HasPrefix(req.Path, "/echo/"):
		req.Body = []byte(strings.TrimPrefix(req.Path, "/echo/"))
		var writeBuffer bytes.Buffer

		writeBuffer.Write([]byte(STATUS_OK + CRLF))
		writeBuffer.Write([]byte("Content-Type: text/plain" + CRLF))
		writeBuffer.Write([]byte("Content-Length: " + strconv.Itoa(len(req.Body)) + CRLF + CRLF))
		writeBuffer.Write([]byte(string(req.Body) + CRLF + CRLF))

		if _, err := writeBuffer.WriteTo(conn); err != nil {
			return err
		}

	default:
		if _, err := conn.Write([]byte(STATUS_NOT_FOUND + CRLF + CRLF)); err != nil {
			return err
		}
	}
	return nil
}
