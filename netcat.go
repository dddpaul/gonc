package main

import (
	"flag"
	"io"
	"log"
	"net"
	// _ "net/http/pprof" // HTTP profiling
	"os"
	"strconv"
)

const (
	// BufferLimit specifies buffer size that is sufficient to handle full-size UDP datagram or TCP segment in one step
	BufferLimit = 2<<16 - 1
	// UDPDisconnectSequence is used to disconnect UDP sessions
	UDPDisconnectSequence = "~."
)

// Progress indicates transfer status
type Progress struct {
	remoteAddr net.Addr
	bytes      uint64
}

func (p *Progress) String() string {
	return "{" +
		"remoteAddr: " + p.remoteAddr.String() +
		", bytes: " + strconv.Itoa(int(p.bytes)) +
		"}"
}

// TransferStreams launches two read-write goroutines and waits for signal from them
func TransferStreams(con io.ReadWriteCloser) {
	c := make(chan Progress)
	done := make(chan bool)

	// Read from Reader and write to Writer until EOF
	copy := func(r io.ReadCloser, w io.WriteCloser) {
		defer func() {
			r.Close()
			w.Close()
			if <-done {
				close(c)
			}
		}()
		n, err := io.Copy(w, r)
		if err != nil {
			log.Printf("Read/write error: %s\n", err)
		}
		var addr net.Addr
		if con, ok := r.(net.Conn); ok {
			addr = con.RemoteAddr()
		}
		if con, ok := w.(net.Conn); ok {
			addr = con.RemoteAddr()
		}
		c <- Progress{bytes: uint64(n), remoteAddr: addr}
	}

	go copy(con, os.Stdout)
	go copy(os.Stdin, con)

	d := false
	for p := range c {
		log.Printf("One of goroutines has been finished: %s\n", p.String())
		done <- d
		d = !d
	}
}

// TransferPackets launches receive goroutine first, wait for address from it (if needed), launches send goroutine then
func TransferPackets(con net.Conn) {
	c := make(chan Progress)
	done := make(chan bool)

	// Read from Reader and write to Writer until EOF.
	// ra is an address to whom packets must be sent in UDP listen mode.
	copy := func(r io.ReadCloser, w io.WriteCloser, ra net.Addr) {
		defer func() {
			r.Close()
			w.Close()
			if _, ok := w.(*net.UDPConn); ok {
				log.Printf("Stop receiving flow from %v\n", ra)
			}
			if <-done {
				close(c)
			}
		}()

		buf := make([]byte, BufferLimit)
		bytes := uint64(0)
		var n int
		var err error
		var addr net.Addr

		for {
			// Read
			if con, ok := r.(*net.UDPConn); ok {
				n, addr, err = con.ReadFrom(buf)
				// Send remote address to caller function (for UDP in listen mode only)
				if con.RemoteAddr() == nil && ra == nil {
					ra = addr
					c <- Progress{remoteAddr: ra}
				}
			} else {
				n, err = r.Read(buf)
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("Read error: %s\n", err)
				}
				break
			}
			if string(buf[0:n-1]) == UDPDisconnectSequence {
				break
			}

			// Write
			if con, ok := w.(*net.UDPConn); ok && con.RemoteAddr() == nil {
				// Special case for UDP in listen mode otherwise net.ErrWriteToConnected will be thrown
				n, err = con.WriteTo(buf[0:n], ra)
			} else {
				n, err = w.Write(buf[0:n])
			}
			if err != nil {
				log.Printf("Write error: %s\n", err)
				break
			}
			bytes += uint64(n)
		}
		c <- Progress{bytes: bytes, remoteAddr: ra}
	}

	go copy(con, os.Stdout, nil)
	ra := con.RemoteAddr()
	// If connection hasn't got remote address then wait for it from receiver goroutine
	if ra == nil {
		p := <-c
		ra = p.remoteAddr
		log.Printf("Connect from %v\n", ra)
	}
	go copy(os.Stdin, con, ra)

	d := false
	for p := range c {
		log.Printf("One of goroutines has been finished: %s\n", p.String())
		done <- d
		d = !d
	}
}

func main() {
	// go http.ListenAndServe(":6060", nil)
	var host, port, proto string
	var listen bool
	flag.StringVar(&host, "host", "", "Remote host to connect, i.e. 127.0.0.1")
	flag.StringVar(&proto, "proto", "tcp", "TCP/UDP mode")
	flag.BoolVar(&listen, "listen", false, "Listen mode")
	flag.StringVar(&port, "port", ":9999", "Port to listen on or connect to (prepended by colon), i.e. :9999")
	flag.Parse()

	startTCPServer := func() {
		ln, err := net.Listen(proto, port)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Listening on", proto+port)
		con, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Connect from", con.RemoteAddr())
		TransferStreams(con)
	}

	startTCPClient := func() {
		con, err := net.Dial(proto, host+port)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Connected to", host+port)
		TransferStreams(con)
	}

	startUDPServer := func() {
		addr, err := net.ResolveUDPAddr(proto, port)
		if err != nil {
			log.Fatalln(err)
		}
		con, err := net.ListenUDP(proto, addr)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Listening on", proto+port)
		TransferPackets(con)
	}

	startUDPClient := func() {
		addr, err := net.ResolveUDPAddr(proto, host+port)
		if err != nil {
			log.Fatalln(err)
		}
		con, err := net.DialUDP(proto, nil, addr)
		if err != nil {
			log.Fatalln(err)
		}
		TransferPackets(con)
	}

	switch proto {
	case "tcp":
		if listen {
			startTCPServer()
		} else if host != "" {
			startTCPClient()
		} else {
			flag.Usage()
		}
	case "udp":
		if listen {
			startUDPServer()
		} else if host != "" {
			startUDPClient()
		} else {
			flag.Usage()
		}
	default:
		flag.Usage()
	}
}
