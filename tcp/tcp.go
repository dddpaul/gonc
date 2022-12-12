package tcp

import (
	"io"
	"log"
	"net"
	"os"
)

// Progress indicates transfer status
type Progress struct {
	direction string
	bytes     uint64
}

// TransferStreams launches two read-write goroutines and waits for signal from them
func TransferStreams(con net.Conn, in io.Reader, out io.Writer) {
	defer func() {
		con.Close()
	}()
	c := make(chan Progress)

	// Read from input and send to connection
	go func() {
		n, err := io.Copy(con, in)
		if err != nil {
			log.Printf("[%s]: ERROR: %s\n", con.RemoteAddr(), err)
		}
		c <- Progress{
			direction: "sent to connection",
			bytes:     uint64(n),
		}
	}()

	// Read from connection and send to output
	go func() {
		n, err := io.Copy(out, con)
		if err != nil {
			log.Printf("[%s]: ERROR: %s\n", con.RemoteAddr(), err)
		}

		c <- Progress{
			direction: "received from connection",
			bytes:     uint64(n),
		}
	}()

	p := <-c
	log.Printf("TCP %s: %d bytes has been %s\n", con.RemoteAddr(), p.bytes, p.direction)
	p = <-c
	log.Printf("TCP %s: %d bytes has been %s\n", con.RemoteAddr(), p.bytes, p.direction)
}

// StartServer starts TCP listener
func StartServer(proto string, port string) {
	ln, err := net.Listen(proto, port)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Listening on", proto+port)
	con, err := ln.Accept()
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("[%s]: Connection has been opened\n", con.RemoteAddr())
	TransferStreams(con, os.Stdin, os.Stdout)
}

// StartClient starts TCP connector
func StartClient(proto string, host string, port string) {
	con, err := net.Dial(proto, host+port)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Connected to", host+port)
	TransferStreams(con, os.Stdin, os.Stdout)
}
