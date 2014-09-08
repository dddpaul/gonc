package main

import (
	"flag"
	"log"
	"io"
	"os"
	"net"
	"strconv"
)

/**
 * Read from Reader and write to Writer until EOF
 */
func readAndWrite(r io.Reader, w io.Writer) <-chan bool {
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

/**
 * Launch two read-write goroutines and waits for signal from them
 */
func transferStreams(con net.Conn) {
	c1 := readAndWrite(os.Stdin, con)
	c2 := readAndWrite(con, os.Stdout)
	select {
	case <-c1:
		log.Println("Local program is terminated")
	case <-c2:
		log.Println("Remote connection is closed")
	}
}

/**
 * Receive UDP datagrams from PacketConn and write it to Writer
 */
func receivePackets(r net.Conn, w io.Writer) <-chan net.Addr {
	buf := make([]byte, 1024)
	c := make(chan net.Addr)
	go func() {
		var remoteAddr net.Addr = nil
		var addr net.Addr
		defer func() {
			c <- remoteAddr
		}()

		for {
			var n int
			var err error
			if con, ok := r.(*net.UDPConn); ok {
				n, addr, err = con.ReadFrom(buf)
			} else {
				n, err = r.Read(buf)
			}
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

/**
 * Read data from Reader and send it as UDP datagram to Addr through PacketConn
 */
func sendPackets(r io.Reader, w net.Conn, addr net.Addr) <-chan bool {
	buf := make([]byte, 1024)
	c := make(chan bool)
	go func() {
		defer func() {
			w.Close();
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
			if con, ok := w.(*net.UDPConn); ok {
				con.WriteTo(buf[0:n], addr)
			} else {
				w.Write(buf[0:n])
			}
		}
	}()
	return c
}

/**
 * Launch receive goroutine first, wait for address from it, launch send goroutine then.
 */
func transferPackets(con net.Conn) {
	c1 := receivePackets(con, os.Stdout)
	remoteAddr := <-c1
	log.Println(remoteAddr)
	c2 := sendPackets(os.Stdin, con, remoteAddr)
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
			p, err := strconv.Atoi(string([]byte(port)[1:]))
			addr := &net.UDPAddr{ IP: net.IPv4(127, 0, 0, 1), Port: p}
			con, err := net.ListenUDP(proto, addr)
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
