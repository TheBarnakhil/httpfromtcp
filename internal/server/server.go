package server

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/TheBarnakhil/httpfromtcp/internal/response"
)

type Server struct {
	Listener net.Listener
	Open     atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return &Server{}, errors.New("error: Error creating a listener")
	}
	server := Server{Listener: listener}
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
	req := []byte{}
	buff := bytes.NewBuffer(req)
	response.WriteStatusLine(buff, 200)
	response.WriteHeaders(buff, response.GetDefaultHeaders(0))

	res := buff.String() + "\r\n" + "Hello World!"
	conn.Write([]byte(res))
}
