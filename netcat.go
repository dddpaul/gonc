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
	flag.StringVar(&host, "host", "", "Remote host to connect, i.e. 127.0.0.1")
	flag.BoolVar(&listen, "listen", false, "Listen mode")
	flag.StringVar(&port, "port", "", "Port to listen on or connect to (prepended by colon), i.e. :9999")
	flag.Parse()

	if listen {
		ln, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Listening on", port)
		con, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Connect from", con.RemoteAddr())
		transferStreams(con)
	} else if host != "" {
		con, err := net.Dial("tcp", host+port)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Connected to", host+port)
		transferStreams(con)
	} else {
		flag.Usage()
	}
}
