package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var Host = "127.0.0.1"
var Port = ":9991"
var Input = "Input from other side, пока, £, 语汉"

// ReadCloser is used to wrap strings.Reader for implementing io.ReadCloser interface
type ReadCloser struct {
	r *strings.Reader
}

func (rc ReadCloser) Read(p []byte) (n int, err error) {
	return rc.r.Read(p)
}

func (rc ReadCloser) Close() error {
	return nil
}

func TestTCP(t *testing.T) {
	// Capture STDIN and STDOUT streams
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	// Bytes written to w1 are read from os.Stdin
	r1, w1, e := os.Pipe()
	assert.Nil(t, e)
	os.Stdin = r1

	// Bytes written os.Stdout are read from r2
	// r2, w2, e := os.Pipe()
	// assert.Nil(t, e)
	// os.Stdout = w2

	// Send test data to listener from goroutine and wait for potentials errors at the end of the test
	go func() {
		// Wait for main thread starts listener
		time.Sleep(200 * time.Millisecond)
		con, err := net.Dial("tcp", Host+Port)
		assert.Nil(t, err)

		// Transfer data
		_, err = w1.Write([]byte(Input))
		assert.Nil(t, err)
		TransferStreams(con)

		// Wait for data will be transferred
		time.Sleep(200 * time.Millisecond)
	}()

	ln, err := net.Listen("tcp", Port)
	assert.Nil(t, err)

	con, err := ln.Accept()
	assert.Nil(t, err)

	buf := make([]byte, 1024)
	n, err := con.Read(buf)
	assert.Nil(t, err)

	// Restore STDIN and STDOUT
	os.Stdin = oldStdin
	os.Stdout = oldStdout

	assert.Equal(t, Input, string(buf[0:n]))
}

func TestUDP(t *testing.T) {
	// Send test data to listener from goroutine and wait for potentials errors at the end of the test
	go func() {
		// Wait for main thread starts listener
		time.Sleep(200 * time.Millisecond)
		con, err := net.Dial("udp", Host+Port)
		assert.Nil(t, err)

		// Transfer data
		addr, err := net.ResolveUDPAddr("udp", Host+Port)
		assert.Nil(t, err)
		fmt.Println(con.RemoteAddr())
		c1 := copyPackets(ReadCloser{r: strings.NewReader(Input)}, con, addr)

		// Wait for data will be transferred
		time.Sleep(200 * time.Millisecond)
		select {
		case progress := <-c1:
			t.Logf("Remote connection is closed: %+v\n", progress)
		default:
			t.Fatal("handle() must write to result channel")
		}
	}()

	con, err := net.ListenPacket("udp", Port)
	assert.Nil(t, err)

	buf := make([]byte, 1024)
	n, _, err := con.ReadFrom(buf)
	assert.Nil(t, err)

	assert.Equal(t, Input, string(buf[0:n]))
}
