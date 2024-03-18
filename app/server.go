package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
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
	CRLF               = "\r\n"
	STATUS_OK          = "HTTP/1.1 200 OK"
	STATUS_CREATED     = "HTTP/1.1 201 Created"
	STATUS_NOT_FOUND   = "HTTP/1.1 404 Not Found"
	STATUS_BAD_REQUEST = "HTTP/1.1 400 Bad Request"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	directoryPtr := flag.String("directory", ".", "the directory to serve files from")
	flag.Parse()
	baseDir := *directoryPtr

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

		go handleConnection(connection, baseDir)
	}
}

func handleConnection(conn net.Conn, baseDir string) {
	defer conn.Close()

	fmt.Println("New connection from: ", conn.RemoteAddr().String())

	req, err := buildRequest(conn)
	if err != nil {
		fmt.Println("Error reading data: ", err.Error())
		os.Exit(1)
	}

	if err := handleResponse(conn, req, baseDir); err != nil {
		fmt.Println("Error writing output: ", err.Error())
		os.Exit(1)
	}

}

func buildRequest(conn net.Conn) (req request, err error) {
	buffer := make([]byte, 4096)
	_, err = conn.Read(buffer)
	if err != nil {
		return req, err
	}

	parts := strings.Split(string(buffer), CRLF+CRLF)
	if len(parts) == 0 {
		return req, errors.New("HTTP startline missing")
	}
	front := parts[0]
	back := parts[1]

	startLine := strings.Split(front, CRLF)[0]
	headers := strings.Split(front, CRLF)[1:]
	err = setRequestPath(startLine, &req)
	if err != nil {
		return req, err
	}
	setHeaders(headers, &req)
	req.Body = []byte(strings.TrimRight(back, "\x00"))

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
		splittedLine := strings.Split(line, ": ")
		if len(splittedLine) == 2 {
			req.Headers[splittedLine[0]] = splittedLine[1]
			fmt.Println(splittedLine[0], " ", splittedLine[1])
		}
	}

	return nil
}

func handleEcho(req request, conn net.Conn) error {
	resBody := []byte(strings.TrimPrefix(req.Path, "/echo/"))
	var writeBuffer bytes.Buffer

	writeBuffer.Write([]byte(STATUS_OK + CRLF))
	writeBuffer.Write([]byte("Content-Type: text/plain" + CRLF))
	writeBuffer.Write([]byte("Content-Length: " + strconv.Itoa(len(resBody)) + CRLF + CRLF))
	writeBuffer.Write([]byte(string(resBody) + CRLF + CRLF))

	if _, err := writeBuffer.WriteTo(conn); err != nil {
		return err
	}
	return nil
}

func handleUserAgent(req request, conn net.Conn) error {
	resBody := []byte(req.Headers["User-Agent"])
	var writeBuffer bytes.Buffer

	writeBuffer.Write([]byte(STATUS_OK + CRLF))
	writeBuffer.Write([]byte("Content-Type: text/plain" + CRLF))
	writeBuffer.Write([]byte("Content-Length: " + strconv.Itoa(len(resBody)) + CRLF + CRLF))
	writeBuffer.Write([]byte(string(resBody) + CRLF + CRLF))

	if _, err := writeBuffer.WriteTo(conn); err != nil {
		return err
	}
	return nil
}

func handleFile(req request, conn net.Conn, baseDir string) error {
	filePath := baseDir + "/" + strings.TrimPrefix(req.Path, "/files/")
	fmt.Println("Serving file: ", filePath)

	switch req.Method {
	case "GET":
		file, err := os.Open(filePath)
		if err != nil {
			if _, err := conn.Write([]byte(STATUS_NOT_FOUND + CRLF + CRLF)); err != nil {
				return err
			}
			return err
		}
		defer file.Close()

		fileData, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println("Error in reading file: ", err.Error())
			return err
		}

		resBody := fileData
		var writeBuffer bytes.Buffer

		writeBuffer.Write([]byte(STATUS_OK + CRLF))
		writeBuffer.Write([]byte("Content-Type: application/octet-stream" + CRLF))
		writeBuffer.Write([]byte("Content-Length: " + strconv.Itoa(len(resBody)) + CRLF + CRLF))
		writeBuffer.Write([]byte(string(resBody) + CRLF + CRLF))

		if _, err := writeBuffer.WriteTo(conn); err != nil {
			return err
		}
		return nil
	case "POST":
		if _, err := os.Stat(filePath); err == nil {
			if _, err := conn.Write([]byte(STATUS_BAD_REQUEST + CRLF + CRLF)); err != nil {
				return err
			}
			return errors.New("the file" + filePath + "has already exist")
		}

		file, err := os.Create(filePath)
		if err != nil {
			if _, err := conn.Write([]byte(STATUS_BAD_REQUEST + CRLF + CRLF)); err != nil {
				return err
			}
			return err
		}
		defer file.Close()

		if _, err := file.Write(req.Body); err != nil {
			if _, err := conn.Write([]byte(STATUS_BAD_REQUEST + CRLF + CRLF)); err != nil {
				return err
			}
			return err
		}

		if _, err := conn.Write([]byte(STATUS_CREATED + CRLF + CRLF)); err != nil {
			return err
		}

		return nil
	default:
		return errors.New("/files/ only accept method [get, post]")
	}

}

func handleResponse(conn net.Conn, req request, baseDir string) error {
	switch {
	case req.Path == "/":
		if _, err := conn.Write([]byte(STATUS_OK + CRLF + CRLF)); err != nil {
			return err
		}
	case strings.HasPrefix(req.Path, "/echo/"):
		if err := handleEcho(req, conn); err != nil {
			return err
		}
	case req.Path == "/user-agent":
		if err := handleUserAgent(req, conn); err != nil {
			return err
		}
	case strings.HasPrefix(req.Path, "/files/"):
		if err := handleFile(req, conn, baseDir); err != nil {
			return err
		}
	default:
		if _, err := conn.Write([]byte(STATUS_NOT_FOUND + CRLF + CRLF)); err != nil {
			return err
		}
	}
	return nil
}
