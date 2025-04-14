package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("Error resolving address: %v", err)
	}
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Error establishing connection: %v", err)
	}
	defer fmt.Println("Connection has been terminated")
	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		line, err := reader.ReadString(byte('\n'))
		if err != nil {
			log.Fatalf("Error reading input: %v", err)
		}
		udpConn.Write([]byte(line))
	}

}
