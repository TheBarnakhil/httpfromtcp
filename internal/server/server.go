package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/TheBarnakhil/httpfromtcp/internal/request"
	"github.com/TheBarnakhil/httpfromtcp/internal/response"
)

type Server struct {
	Listener    net.Listener
	Open        atomic.Bool
	HandlerFunc Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, handlerFunc Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return &Server{}, errors.New("error: Error creating a listener")
	}
	server := Server{Listener: listener, HandlerFunc: handlerFunc}
	server.Open.Store(true)
	go server.listen()
	return &server, nil
}

func (s *Server) Close() error {
	s.Open.Store(false)
	if s.Listener != nil {
		return s.Listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for s.Open.Load() {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection %v", err)
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	writer := response.Writer{}

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.BadRequest,
			Message: fmt.Sprintf(`
		<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me. %v</p>
  </body>
</html>
		`, err),
		}
		hErr.writeHandlerErrortoWriter(&writer)
	}
	s.HandlerFunc(&writer, req)

	conn.Write(writer.StatusLine)
	conn.Write(writer.Headers)
	conn.Write(writer.Body)
}

func (h HandlerError) writeHandlerErrortoWriter(w *response.Writer) {
	w.WriteStatusLine(h.StatusCode)
	messageBytes := []byte(h.Message)
	headers := response.GetDefaultHeaders(len(messageBytes), response.HTML)
	w.WriteHeaders(headers)
	w.WriteBody(messageBytes)
}
