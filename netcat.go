package main

import (
	"net"
	"log"
	"io"
	"os"
)

func handle(r io.Reader, w io.Writer) {
	defer func() {
		if con, ok := w.(net.Conn); ok {
			con.Close();
			log.Printf("Connection from %v is closed\n", con.RemoteAddr())
		}
	}()
	buf := make([]byte, 1024)
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
}

func main() {
	ln, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Println(err)
		return
	}
	for {
		con, err := ln.Accept()
		log.Println("Connect from", con.RemoteAddr())
		if err != nil {
			log.Println(err)
			continue
		}
		go handle(os.Stdin, con)
		go handle(con, os.Stdout)
	}
}
