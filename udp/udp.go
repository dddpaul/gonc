package udp

import (
	"io"
	"log"
	"net"
	"os"
)

const (
	// BufferLimit specifies buffer size that is sufficient to handle full-size UDP datagram or TCP segment in one step
	BufferLimit = 2<<16 - 1
	// DisconnectSequence is used to disconnect UDP sessions
	DisconnectSequence = "~."
)

// Progress indicates transfer status
type Progress struct {
	remoteAddr net.Addr
	bytes      uint64
}

// TransferPackets launches receive goroutine first, wait for address from it (if needed), launches send goroutine then
func TransferPackets(con net.Conn) {
	c := make(chan Progress)

	// Read from Reader and write to Writer until EOF.
	// ra is an address to whom packets must be sent in listen mode.
	copy := func(r io.ReadCloser, w io.WriteCloser, ra net.Addr) {
		defer func() {
			r.Close()
			w.Close()
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
				// In listen mode remote address is unknown until read from connection.
				// So we must inform caller function with received remote address.
				if con.RemoteAddr() == nil && ra == nil {
					ra = addr
					c <- Progress{remoteAddr: ra}
				}
			} else {
				n, err = r.Read(buf)
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("[%s]: ERROR: %s\n", ra, err)
				}
				break
			}
			if string(buf[0:n-1]) == DisconnectSequence {
				break
			}

			// Write
			if con, ok := w.(*net.UDPConn); ok && con.RemoteAddr() == nil {
				// Connection remote address must be nil otherwise "WriteTo with pre-connected connection" will be thrown
				n, err = con.WriteTo(buf[0:n], ra)
			} else {
				n, err = w.Write(buf[0:n])
			}
			if err != nil {
				log.Printf("[%s]: ERROR: %s\n", ra, err)
				break
			}
			bytes += uint64(n)
		}
		c <- Progress{bytes: bytes}
	}

	ra := con.RemoteAddr()
	go copy(con, os.Stdout, ra)
	// If connection hasn't got remote address then wait for it from receiver goroutine
	if ra == nil {
		p := <-c
		ra = p.remoteAddr
		log.Printf("[%s]: Datagram has been received\n", ra)
	}
	go copy(os.Stdin, con, ra)

	p := <-c
	log.Printf("[%s]: Connection has been closed, %d bytes has been received\n", ra, p.bytes)
	p = <-c
	log.Printf("[%s]: Local peer has been stopped, %d bytes has been sent\n", ra, p.bytes)
}

// StartServer starts UDP listener
func StartServer(proto string, port string) {
	addr, err := net.ResolveUDPAddr(proto, port)
	if err != nil {
		log.Fatalln(err)
	}
	con, err := net.ListenUDP(proto, addr)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Listening on", proto+port)
	// This connection doesn't know remote address yet
	TransferPackets(con)
}

// StartClient starts UDP connector
func StartClient(proto string, host string, port string) {
	addr, err := net.ResolveUDPAddr(proto, host+port)
	if err != nil {
		log.Fatalln(err)
	}
	con, err := net.DialUDP(proto, nil, addr)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Sending datagrams to", host+port)
	TransferPackets(con)
}
