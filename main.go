package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

	dir := flag.String("dir", ".", "The directory to serve files from (default is current directory)")
	flag.Parse()

	if !strings.HasSuffix(*dir, "/") {
		*dir += "/"
	}

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Failed to bind to port 8080")
		os.Exit(1)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		go handleConnection(conn, *dir)
	}
}

func handleConnection(conn net.Conn, dir string) {
	defer conn.Close()

	req := make([]byte, 1024)
	n, err := conn.Read(req)
	if err != nil {
		fmt.Println("Error reading request:", err.Error())
		return
	}

	requestLines := strings.Split(string(req[:n]), "\r\n")
	requestLine := requestLines[0]
	method := strings.Split(requestLine, " ")[0]
	path := strings.Split(requestLine, " ")[1]

	var userAgent string
	for _, line := range requestLines {
		if strings.HasPrefix(line, "User-Agent:") {
			userAgent = strings.TrimSpace(strings.TrimPrefix(line, "User-Agent:"))
			break
		}
	}

	fmt.Println("User-Agent:", userAgent)
	fmt.Println("Path:", path)
	fmt.Println("Method:", method)

	if strings.HasPrefix(path, "/echo/") {
		message := strings.TrimPrefix(path, "/echo/")
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(message), message)))
	} else if strings.HasPrefix(path, "/files/") {
		filename := strings.TrimPrefix(path, "/files/")
		filePath := dir + filename

		if method == "GET" {
			data, err := os.ReadFile(filePath)
			var response string

			if err != nil {
				if os.IsNotExist(err) {
					response = "HTTP/1.1 404 Not Found\r\n\r\n"
				} else {
					response = "HTTP/1.1 500 Internal Server Error\r\n\r\n"
				}
			} else {
				response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(data), data)
			}
			conn.Write([]byte(response))
		} else if method == "POST" {
			body := strings.Join(requestLines[len(requestLines)-1:], "\n")
			err := os.WriteFile(filePath, []byte(body), 0644)
			if err != nil {
				conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
				return
			}
			conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		}
	} else {
		switch path {
		case "/":
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		case "/user-agent":
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
			conn.Write([]byte(response))
		default:
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}
	}
}
