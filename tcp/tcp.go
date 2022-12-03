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

	// Read from Reader and write to Writer until EOF
	copy := func(r io.Reader, w io.Writer) {
		n, err := io.Copy(w, r)
		if err != nil {
			log.Printf("[%s]: ERROR: %s\n", con.RemoteAddr(), err)
		}

		var direction string
		if _, ok := w.(net.Conn); ok {
			direction = "sent to connection"
		} else {
			direction = "received from connection"
		}

		c <- Progress{
			direction: direction,
			bytes:     uint64(n),
		}
	}

	go copy(in, con)
	go copy(con, out)

	p := <-c
	log.Printf("[%s]: Connection has been closed by remote peer, %d bytes has been %s\n", con.RemoteAddr(), p.bytes, p.direction)
	p = <-c
	log.Printf("[%s]: Local peer has been stopped, %d bytes has been %s\n", con.RemoteAddr(), p.bytes, p.direction)
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
