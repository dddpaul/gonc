package main

import (
	"net"
	"fmt"
	"log"
	"io"
	"os"
)

func handleIn(con net.Conn) {
	buf := make([]byte, 128)
	for {
		n, err := con.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Read error: %s\n", err)
			}
			break
		}
		log.Printf("Input buffer: len=%d cap=%d\n", len(buf), cap(buf))
		fmt.Print(string(buf[0:n]))
	}
}

func handleOut(con net.Conn) {
	buf := make([]byte, 128)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Read error: %s\n", err)
			}
			break
		}
		log.Printf("Output buffer: len=%d cap=%d\n", len(buf), cap(buf))
		con.Write(buf[0:n])
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
		go handleIn(con)
		go handleOut(con)
	}
}
