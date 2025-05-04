package response

import (
	"errors"
	"fmt"
	"io"
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
	TrailersNext
	Done
)

type Writer struct {
	writerState WriterState
	writer      io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writerState: StatusLineNext,
		writer:      w,
	}
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
		_, err := w.writer.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}
		w.writerState = HeadersNext
		return nil
	case BadRequest:
		_, err := w.writer.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}
		w.writerState = HeadersNext
		return nil
	case ServerError:
		_, err := w.writer.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}
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
	_, err := w.writer.Write([]byte(headerStr + "\r\n"))
	if err != nil {
		return err
	}
	w.writerState = BodyNext
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != BodyNext {
		return 0, errors.New("error: add the status line then the headers and then the body")
	}
	n, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}
	w.writerState = Done
	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writerState != BodyNext {
		return 0, errors.New("error: add the status line then the headers and then the body")
	}
	n1, err := w.writer.Write([]byte(strconv.FormatInt(int64(len(p)), 16) + "\r\n"))
	if err != nil {
		return 0, err
	}
	n2, err := w.writer.Write([]byte(p))
	if err != nil {
		return n1, err
	}
	n3, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return n1 + n2, err
	}
	return n1 + n2 + n3, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != BodyNext {
		return 0, errors.New("error: add the status line then the headers and then the body")
	}
	n, err := w.writer.Write([]byte("0" + "\r\n"))
	if err != nil {
		return 0, err
	}
	w.writerState = TrailersNext
	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.writerState != TrailersNext {
		return errors.New("error: add the status line then the headers and then the body")
	}
	if len(h) != 0 {
		headerStr := ""
		for key, val := range h {
			headerStr += fmt.Sprintf("%s: %s\r\n", key, val)
		}
		_, err := w.writer.Write([]byte(headerStr))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.writerState = Done
	return nil
}
