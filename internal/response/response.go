package response

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/TheBarnakhil/httpfromtcp/internal/headers"
)

type StatusCode int

type WriterState int

type ContentType string

const (
	Plain ContentType = "text/plain"
	HTML  ContentType = "text/html"
)

const (
	StatusLineNext WriterState = iota
	HeadersNext
	BodyNext
	Done
)

type Writer struct {
	StatusLine  []byte
	Headers     []byte
	Body        []byte
	writerState WriterState
}

const (
	OK          StatusCode = 200
	BadRequest  StatusCode = 400
	ServerError StatusCode = 500
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != StatusLineNext {
		return errors.New("error: add the status line then the headers and then the body")
	}

	switch statusCode {
	case OK:
		w.StatusLine = []byte("HTTP/1.1 200 OK\r\n")
		w.writerState = HeadersNext
		return nil
	case BadRequest:
		w.StatusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
		w.writerState = HeadersNext
		return nil
	case ServerError:
		w.StatusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
		w.writerState = HeadersNext
		return nil
	default:
		return errors.New("error: Invalid status code")
	}
}

func GetDefaultHeaders(contentLen int, contentType ContentType) headers.Headers {
	h := headers.NewHeaders()
	h["Content-Length"] = strconv.Itoa(contentLen)
	h["Connection"] = "close"
	if contentType != "" {
		h["Content-Type"] = string(contentType)
	} else {
		h["Content-Type"] = "text/plain"
	}
	return h
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != HeadersNext {
		return errors.New("error: add the status line then the headers and then the body")
	}
	headerStr := ""
	for key, val := range headers {
		headerStr += fmt.Sprintf("%s: %s\r\n", key, val)
	}
	w.Headers = []byte(headerStr + "\r\n")
	w.writerState = BodyNext
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != BodyNext {
		return 0, errors.New("error: add the status line then the headers and then the body")
	}
	w.Body = p
	w.writerState = Done
	return len(p), nil
}
