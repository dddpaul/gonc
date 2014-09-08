package main

import (
	"github.com/stretchr/testify/assert"
	"net"
	"strings"
	"testing"
	"time"
)

var Host = "127.0.0.1"
var Port = ":9991"
var Input = "Input from other side, пока, £, 语汉"

func TestTCP(t *testing.T) {
	// Send test data to listener from goroutine and wait for potentials errors at the end of the test
	go func() {
		// Wait for main thread starts listener
		time.Sleep(200 * time.Millisecond)
		con, err := net.Dial("tcp", Host+Port)
		assert.Nil(t, err)

		// Transfer data
		c1 := readAndWrite(strings.NewReader(Input), con)

		// Wait for data will be transferred
		time.Sleep(200 * time.Millisecond)
		select {
		case <-c1:
		default:
			t.Fatal("handle() must write to result channel")
		}
	}()

	ln, err := net.Listen("tcp", Port)
	assert.Nil(t, err)

	con, err := ln.Accept()
	assert.Nil(t, err)

	buf := make([]byte, 1024)
	n, err := con.Read(buf)
	assert.Nil(t, err)

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
		c1 := readAndWrite(strings.NewReader(Input), con)

		// Wait for data will be transferred
		time.Sleep(200 * time.Millisecond)
		select {
		case <-c1:
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
