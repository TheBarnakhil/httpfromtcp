package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/TheBarnakhil/httpfromtcp/internal/headers"
	"github.com/TheBarnakhil/httpfromtcp/internal/request"
	"github.com/TheBarnakhil/httpfromtcp/internal/response"
	"github.com/TheBarnakhil/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		route := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")

		resp, err := http.Get("https://httpbin.org" + route)
		if err != nil {
			handler400(w, req)
		}
		defer resp.Body.Close()

		w.WriteStatusLine(response.OK)
		h := response.GetDefaultHeaders(0, response.Plain)
		delete(h, "Content-Length")
		h["Transfer-Encoding"] = "chunked"
		err = w.WriteHeaders(h)
		if err != nil {
			fmt.Println("error in writing headers: ", err)
			return
		}

		buff := make([]byte, 1024)
		var body []byte
		for {
			n, err := resp.Body.Read(buff)
			body = append(body, buff[:n]...)
			fmt.Println("Read", n, "bytes")
			if n > 0 {
				_, err := w.WriteChunkedBody(buff[:n])
				if err != nil {
					fmt.Println("Error writing chunked body: ", err)
					return
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("Error reading response body:", err)
				break
			}
		}
		_, err = w.WriteChunkedBodyDone()
		if err != nil {
			fmt.Println("Error writing chunked body done: ", err)
		}

		h = headers.NewHeaders()
		h["Trailer"] = "X-Content-SHA256, X-Content-Length"
		h["X-Content-SHA256"] = fmt.Sprintf("%x", sha256.Sum256(body))
		h["X-Content-Length"] = strconv.Itoa(len(body))
		w.WriteTrailers(h)
	}

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		handler400(w, req)
		return
	case "/myproblem":
		handler500(w, req)
	default:
		handler200(w, req)
	}
}

func handler400(w *response.Writer, _ *request.Request) {
	content := `
		<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
		`
	w.WriteStatusLine(response.BadRequest)
	headers := response.GetDefaultHeaders(len(content), response.HTML)
	w.WriteHeaders(headers)
	w.WriteBody([]byte(content))
}

func handler500(w *response.Writer, _ *request.Request) {
	content := `
	<html>
<head>
	<title>500 Internal Server Error</title>
</head>
<body>
	<h1>Internal Server Error</h1>
	<p>Okay, you know what? This one is on me.</p>
</body>
</html>
	`
	w.WriteStatusLine(response.ServerError)
	headers := response.GetDefaultHeaders(len(content), response.HTML)
	w.WriteHeaders(headers)
	w.WriteBody([]byte(content))
}

func handler200(w *response.Writer, _ *request.Request) {
	content := `
	<html>
<head>
	<title>200 OK</title>
</head>
<body>
	<h1>Success!</h1>
	<p>Your request was an absolute banger.</p>
</body>
</html>
	`
	w.WriteStatusLine(response.OK)
	headers := response.GetDefaultHeaders(len(content), response.HTML)
	w.WriteHeaders(headers)
	w.WriteBody([]byte(content))
}
