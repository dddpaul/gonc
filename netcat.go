package main

import (
	"flag"
	"io"
	"log"
	"net"
	"sync/atomic"
	// _ "net/http/pprof" // HTTP profiling
	"os"
)

// Ready to handle full-size UDP datagram or TCP segment in one step
const (
	BUFFERLIMIT = 2<<16 - 1
)

// Progress indicates transfer status
type Progress struct {
	ra     net.Addr
	rBytes uint64
	wBytes uint64
}

// TransferStreams launches two read-write goroutines and waits for signal from them
func TransferStreams(con io.ReadWriteCloser) {
	c := make(chan Progress)
	var i uint64

	// Read from Reader and write to Writer until EOF
	copy := func(r io.ReadCloser, w io.WriteCloser) {
		defer func() {
			r.Close()
			w.Close()
			var toClose bool
			for atomic.CompareAndSwapUint64(&i, 1, 2) {
				toClose = true
			}
			if toClose {
				close(c)
				toClose = false
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
		c <- Progress{rBytes: uint64(n), wBytes: 0, ra: addr}
	}

	go copy(con, os.Stdout)
	go copy(os.Stdin, con)

	for p := range c {
		log.Printf("One of the copy goroutines has been finished: %+v\n", p)
		atomic.AddUint64(&i, 1)
	}
}

// TransferPackets launches receive goroutine first, wait for address from it (if needed), launches send goroutine then
func TransferPackets(con net.Conn) {
	c1 := copyPackets(con, os.Stdout, nil)
	// If connection hasn't got remote address then wait for it from receiver goroutine
	ra := con.RemoteAddr()
	if ra == nil {
		progress := <-c1
		ra = progress.ra
		log.Println("Connect from", ra)
	}
	c2 := copyPackets(os.Stdin, con, ra)
	select {
	case progress := <-c1:
		log.Printf("Remote connection is closed: %+v\n", progress)
	case progress := <-c2:
		log.Printf("Local program is terminated: %+v\n", progress)
	}
}

// Read from Reader and write to Writer until EOF.
// ra is an address to whom packets must be sent in UDP listen mode.
func copyPackets(r io.ReadCloser, w io.WriteCloser, ra net.Addr) <-chan Progress {
	buf := make([]byte, BUFFERLIMIT)
	c := make(chan Progress)
	rBytes, wBytes := uint64(0), uint64(0)
	go func() {
		defer func() {
			r.Close()
			w.Close()
			if _, ok := w.(*net.UDPConn); ok {
				log.Printf("Stop receiving flow from %v\n", ra)
			}
			c <- Progress{rBytes: rBytes, wBytes: wBytes, ra: ra}
		}()

		var n int
		var err error
		var addr net.Addr

		for {
			// Read
			if con, ok := r.(*net.UDPConn); ok {
				n, addr, err = con.ReadFrom(buf)
				// Inform caller function with remote address once
				// (for UDP in listen mode only)
				if con.RemoteAddr() == nil && ra == nil {
					ra = addr
					c <- Progress{rBytes: rBytes, wBytes: wBytes, ra: ra}
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
			rBytes += uint64(n)

			// Write
			if con, ok := w.(*net.UDPConn); ok && con.RemoteAddr() == nil {
				// Special case for UDP in listen mode otherwise
				// net.ErrWriteToConnected will be thrown
				n, err = con.WriteTo(buf[0:n], ra)
			} else {
				n, err = w.Write(buf[0:n])
			}
			if err != nil {
				log.Fatalf("Write error: %s\n", err)
			}
			wBytes += uint64(n)
		}
	}()
	return c
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

	if proto == "tcp" {
		if listen {
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
		} else if host != "" {
			con, err := net.Dial(proto, host+port)
			if err != nil {
				log.Fatalln(err)
			}
			log.Println("Connected to", host+port)
			TransferStreams(con)
		} else {
			flag.Usage()
		}
	} else if proto == "udp" {
		if listen {
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
		} else if host != "" {
			addr, err := net.ResolveUDPAddr(proto, host+port)
			if err != nil {
				log.Fatalln(err)
			}
			con, err := net.DialUDP(proto, nil, addr)
			if err != nil {
				log.Fatalln(err)
			}
			TransferPackets(con)
		} else {
			flag.Usage()
		}
	}
}
