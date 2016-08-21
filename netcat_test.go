package main

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var Host = "127.0.0.1"
var Port = ":9991"
var Input = "Input from other side, пока, £, 语汉"

func TestTransferStreams(t *testing.T) {
	oldStdin := os.Stdin

	// Bytes written to w1 are read from os.Stdin
	r, w, e := os.Pipe()
	assert.Nil(t, e)
	os.Stdin = r

	// Send data to server from goroutine and wait for potentials errors at the end of the test
	go func() {
		// Wait for main thread starts server
		time.Sleep(200 * time.Millisecond)
		con, err := net.Dial("tcp", Host+Port)
		assert.Nil(t, err)

		// Client sends data
		_, err = w.Write([]byte(Input))
		assert.Nil(t, err)
		TransferStreams(con)
	}()

	// Server receives data
	ln, err := net.Listen("tcp", Port)
	assert.Nil(t, err)

	con, err := ln.Accept()
	assert.Nil(t, err)

	buf := make([]byte, 1024)
	n, err := con.Read(buf)
	assert.Nil(t, err)

	assert.Equal(t, Input, string(buf[0:n]))

	os.Stdin = oldStdin
}

func TestTransferPackets(t *testing.T) {
	oldStdin := os.Stdin

	// Bytes written to w1 are read from os.Stdin
	r, w, e := os.Pipe()
	assert.Nil(t, e)
	os.Stdin = r

	// Send test data to server from goroutine and wait for potentials errors at the end of the test
	go func() {
		// Wait for main thread starts server
		time.Sleep(200 * time.Millisecond)
		con, err := net.Dial("udp", Host+Port)
		assert.Nil(t, err)

		// Client sends data
		_, err = w.Write([]byte(Input))
		assert.Nil(t, err)
		TransferStreams(con)
	}()

	con, err := net.ListenPacket("udp", Port)
	assert.Nil(t, err)

	buf := make([]byte, 1024)
	n, _, err := con.ReadFrom(buf)
	assert.Nil(t, err)

	assert.Equal(t, Input, string(buf[0:n]))

	os.Stdin = oldStdin
}
