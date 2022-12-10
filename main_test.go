package main

import (
	"bytes"
	"io/ioutil"
	"net"
	"testing"

	"github.com/dddpaul/gonc/tcp"
	"github.com/dddpaul/gonc/udp"
	"github.com/stretchr/testify/assert"
)

var Host = "127.0.0.1"
var Port = ":9991"
var Input = "Input from my side, пока, £, 语汉"
var InputFromOtherSide = "Input from other side, пока, £, 语汉, 123"

func TestTransferStreams(t *testing.T) {
	in := bytes.NewReader([]byte(Input))
	out := new(bytes.Buffer)

	ready := make(chan bool, 1)
	done := make(chan bool, 1)

	// Send data from "my" side
	go func() {
		<-ready
		con, err := net.Dial("tcp", Host+Port)
		assert.Nil(t, err)
		tcp.TransferStreams(con, in, out)
		done <- true
	}()

	// Start server on the "other" side
	ln, err := net.Listen("tcp", Port)
	assert.Nil(t, err)
	ready <- true
	con, err := ln.Accept()
	assert.Nil(t, err)

	// Read data on the "other" side
	buf := make([]byte, 1024)
	n, err := con.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, Input, string(buf[0:n]))

	// Send data from the "other" side
	n, err = con.Write([]byte(InputFromOtherSide))
	assert.Nil(t, err)
	err = con.Close()
	assert.Nil(t, err)
	<-done
	assert.Equal(t, InputFromOtherSide, string(out.Bytes()[0:n]))
}

func TestTransferPackets(t *testing.T) {
	in := ioutil.NopCloser(bytes.NewReader([]byte(Input)))
	out := new(bytes.Buffer)

	ready := make(chan bool, 1)
	done := make(chan bool, 1)

	// Send data from "my" side
	go func() {
		<-ready
		con, err := net.Dial("udp", Host+Port)
		assert.Nil(t, err)
		udp.TransferPackets(con, in, out)
		done <- true
	}()

	// Start server on the "other" side
	con, err := net.ListenPacket("udp", Port)
	assert.Nil(t, err)
	ready <- true

	// Read data on the "other" side
	buf := make([]byte, 1024)
	n, a, err := con.ReadFrom(buf)
	assert.Nil(t, err)
	assert.Equal(t, Input, string(buf[0:n]))

	// Send data from the "other" side
	n, err = con.WriteTo([]byte(InputFromOtherSide), a)
	assert.Nil(t, err)
	_, err = con.WriteTo([]byte("~."), a)
	assert.Nil(t, err)
	err = con.Close()
	assert.Nil(t, err)
	// <-done
	assert.Equal(t, InputFromOtherSide, string(out.Bytes()[0:n]))
}
