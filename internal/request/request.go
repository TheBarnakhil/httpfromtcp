package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/TheBarnakhil/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	ParserState internal
	Headers     headers.Headers
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

type internal int

const (
	Initialized internal = iota
	Done
	ParsingHeaders
)

const bufferLen = 8

func RequestFromReader(reader io.Reader) (*Request, error) {

	buffer := make([]byte, bufferLen)
	readToIndex := 0

	req := &Request{
		ParserState: Initialized,
		Headers:     headers.NewHeaders(),
	}

	for req.ParserState != Done {
		if len(buffer) <= readToIndex {
			bufferCopy := make([]byte, len(buffer)*2)
			copy(bufferCopy, buffer)
			buffer = bufferCopy
		}

		numBytesRead, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.ParserState != Done {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", req.ParserState, numBytesRead)
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead
		parsedNum, err := req.parse(buffer[:readToIndex])
		if err != nil {
			return req, errors.New("error: Unable to parse from buffer")
		}
		copy(buffer, buffer[parsedNum:])
		readToIndex -= parsedNum

	}
	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	// Get just request line
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])

	// Get the pointer to the the RequestLine struct
	reqLine, err := splitRequestLineString(requestLineText)
	if err != nil {
		return &RequestLine{}, idx, err
	}

	if !strings.Contains(reqLine.RequestTarget, "/") {
		return &RequestLine{}, idx, errors.New("invalid request target")
	}

	// Check if http version is 1.1
	if reqLine.HttpVersion != "1.1" {
		return &RequestLine{}, idx, errors.New("invalid http version")
	}

	// Check if method contains only cap letters
	for _, char := range reqLine.Method {
		if unicode.IsLower(char) {
			return &RequestLine{}, idx, errors.New("method names are required to be uppercase")
		}
	}

	return reqLine, idx + 2, nil
}

/*
splitRequestLine takes a string and splits into 3 parts.
str is split by whitespaces and a RequestLine struct is constructed.
It returns a pointer to the created struct and an error if the length of the sections is not 3
*/
func splitRequestLineString(str string) (*RequestLine, error) {
	sections := strings.Split(str, " ")
	if len(sections) != 3 {
		return &RequestLine{}, errors.New("there are not enough sections present in the request line")
	}
	httpVersion := strings.Split(sections[2], "/")
	if len(httpVersion) != 2 {
		return &RequestLine{}, errors.New("invalid http version string")
	}
	return &RequestLine{HttpVersion: httpVersion[1], RequestTarget: sections[1], Method: sections[0]}, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.ParserState != Done {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.ParserState {
	case Initialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			// something actually went wrong
			return 0, err
		}
		if n == 0 {
			// just need more data
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.ParserState = ParsingHeaders
		return n, nil
	case ParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.ParserState = Done
		}
		return n, nil
	case Done:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("unknown state")
	}
}
