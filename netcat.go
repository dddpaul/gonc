package main

import (
	"net"
	"log"
	"io"
	"os"
	"flag"
)

func handle(r io.Reader, w io.Writer) <-chan bool {
	buf := make([]byte, 1024)
	c := make(chan bool)
	go func() {
		defer func() {
			if con, ok := w.(net.Conn); ok {
				con.Close();
				log.Printf("Connection from %v is closed\n", con.RemoteAddr())
			}
			c <- true
		}()

		for {
			n, err := r.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Read error: %s\n", err)
				}
				break
			}
			w.Write(buf[0:n])
		}
	}()
	return c
}

func transferStreams(con net.Conn) {
	c1 := handle(os.Stdin, con)
	c2 := handle(con, os.Stdout)
	select {
	case <-c1:
		log.Println("Local program is terminated")
	case <-c2:
		log.Println("Remote connection is closed")
	}
}

func main() {
	var host, port string
	var listen bool
	flag.StringVar(&host, "host", "127.0.0.1", "Remote host to connect")
	flag.BoolVar(&listen, "listen", false, "Listen mode")
	flag.StringVar(&port, "port", ":9999", "Port to listen on or connect to")
	flag.Parse()

	if listen {
		log.Println("Listening on", port)
		ln, err := net.Listen("tcp", port)
		if err != nil {
			log.Println(err)
			return
		}
		con, err := ln.Accept()
		log.Println("Connect from", con.RemoteAddr())
		if err != nil {
			log.Println(err)
			return
		}
		transferStreams(con)
	} else {
		log.Println("Connecting to", host+port)
		con, err := net.Dial("tcp", host+port)
		if err != nil {
			log.Println(err)
			return
		}
		transferStreams(con)
	}
}
