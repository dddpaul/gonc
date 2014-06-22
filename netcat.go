package main

import (
	"net"
	"log"
	"io"
	"os"
)

func handle(reader io.Reader, writer io.Writer) {
	buf := make([]byte, 128)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Read error: %s\n", err)
			}
			break
		}
		writer.Write(buf[0:n])
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
