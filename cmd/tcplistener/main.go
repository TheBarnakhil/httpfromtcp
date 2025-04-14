package main

import (
	"fmt"
	"io"
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
		fmt.Println(line)
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	line := make(chan string)
	go func() {
		defer f.Close()
		defer close(line)
		currLine := ""
		for {
			buff := make([]byte, 8)
			n, err := f.Read(buff)
			splitVal := strings.Split(string(buff[:n]), "\n")
			for i := 0; i < len(splitVal)-1; i++ {
				line <- fmt.Sprintf("%s%s", currLine, splitVal[i])
				currLine = ""
			}
			currLine += splitVal[len(splitVal)-1]
			if err != nil {
				break
			}
		}
		line <- currLine
	}()
	return line
}
