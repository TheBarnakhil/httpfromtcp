package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	f, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatalf("Error opening the file: %v", err)
	}
	defer f.Close()

	lineChan := getLinesChannel(f)

	for line := range lineChan {
		fmt.Printf("read: %s\n", line)
	}

}

func getLinesChannel(f io.Reader) <-chan string {
	line := make(chan string)

	go func() {
		currLine := ""
		for {
			buff := make([]byte, 8)
			_, err := f.Read(buff)
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
