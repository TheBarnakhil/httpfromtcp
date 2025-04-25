package main

import (
	"fmt"
	"log"
	"net"

	"github.com/TheBarnakhil/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("Error opening the file: %v", err)
	}
	defer fmt.Println("Connection has been terminated")
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		log.Fatalf("Error connecting to TCP listener: %v", err)
	}

	log.Println("Connection has been established!")

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Request line:")
	fmt.Println("- Method:", req.RequestLine.Method)
	fmt.Println("- Target:", req.RequestLine.RequestTarget)
	fmt.Println("- Version:", req.RequestLine.HttpVersion)
	fmt.Println("Headers:")
	for key, val := range req.Headers {
		fmt.Printf("- %s: %s\n", key, val)
	}
}
