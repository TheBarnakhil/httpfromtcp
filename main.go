package main

import (
	"fmt"
	"log"
	"net"
	"strings"
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

	lineChan := getLinesChannel(conn)

	for line := range lineChan {
		fmt.Printf("read: %s\n", line)
	}

}

func getLinesChannel(conn net.Conn) <-chan string {
	line := make(chan string)

	go func() {
		currLine := ""
		for {
			buff := make([]byte, 8)
			_, err := conn.Read(buff)
			splitVal := strings.Split(string(buff), "\n")
			if len(splitVal) == 1 {
				currLine += splitVal[0]
			} else if len(splitVal) == 2 {
				line <- (currLine + splitVal[0])
				currLine = splitVal[1]
			}
			if err != nil {
				break
			}
		}
		line <- currLine
		close(line)
	}()
	return line
}
