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

func handleUdpIn(r net.PacketConn, w io.Writer) <-chan net.Addr {
	buf := make([]byte, 1024)
	c := make(chan net.Addr)
	go func() {
		var remoteAddr net.Addr = nil
		defer func() {
			if con, ok := w.(net.PacketConn); ok {
				con.Close();
				log.Printf("Connection is closed\n")
			}
			c <- remoteAddr
		}()

		for {
			n, addr, err := r.ReadFrom(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Read error: %s\n", err)
				}
				break
			}
			if remoteAddr == nil {
				remoteAddr = addr
				c <- remoteAddr
			}
			w.Write(buf[0:n])
		}
	}()
	return c
}

func handleUdpOut(r io.Reader, w net.PacketConn, addr net.Addr) <-chan bool {
	buf := make([]byte, 1024)
	c := make(chan bool)
	go func() {
		defer func() {
			if con, ok := w.(net.PacketConn); ok {
				con.Close();
				log.Printf("Connection is closed\n")
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
			w.WriteTo(buf[0:n], addr)
		}
	}()
	return c
}

func transferPackets(con net.PacketConn) {
	c1 := handleUdpIn(con, os.Stdout)
	remoteAddr := <-c1
	c2 := handleUdpOut(os.Stdin, con, remoteAddr)
	select {
		case <-c1:
			log.Println("Remote connection is closed")
		case <-c2:
			log.Println("Local program is terminated")
	}
}

func main() {
	var host, port, proto string
	var listen bool
	flag.StringVar(&host, "host", "", "Remote host to connect, i.e. 127.0.0.1")
	flag.StringVar(&proto, "proto", "tcp", "TCP/UDP mode")
	flag.BoolVar(&listen, "listen", false, "Listen mode")
	flag.StringVar(&port, "port", "", "Port to listen on or connect to (prepended by colon), i.e. :9999")
	flag.Parse()

	if listen {
		if proto == "tcp" {
			ln, err := net.Listen(proto, port)
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
		} else if proto == "udp" {
			con, err := net.ListenPacket(proto, port)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println("Listening on", port)
			transferPackets(con)
		}
	} else if host != "" {
		con, err := net.Dial(proto, host+port)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Connected to", host+port)
		transferStreams(con)
	} else {
		flag.Usage()
	}
}
