package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"os"
	"testing"

	"github.com/dddpaul/gonc/tcp"
	"github.com/dddpaul/gonc/udp"
	"github.com/stretchr/testify/assert"
)

type MockWriter struct {
	w io.Writer
}

func (mw MockWriter) Write(p []byte) (n int, err error) {
	return mw.Write(p)
}

func (mw MockWriter) Close() error {
	return nil
}

var Host = "127.0.0.1"
var Port = ":9991"
var Input = "Input from other side, пока, £, 语汉"

func TestTransferStreams(t *testing.T) {
	in := ioutil.NopCloser(bytes.NewReader([]byte(Input)))
	out := MockWriter{ioutil.Discard}

	// Send data to server
	go func() {
		con, err := net.Dial("tcp", Host+Port)
		assert.Nil(t, err)
		tcp.TransferStreams(con, in, out)
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
}

func TestTransferPackets(t *testing.T) {
	w, oldStdin := mockStdin(t)

	// Send data to server
	go func() {
		con, err := net.Dial("udp", Host+Port)
		assert.Nil(t, err)
		_, err = w.Write([]byte(Input))
		assert.Nil(t, err)
		udp.TransferPackets(con)
	}()

	con, err := net.ListenPacket("udp", Port)
	assert.Nil(t, err)

	buf := make([]byte, 1024)
	n, _, err := con.ReadFrom(buf)
	assert.Nil(t, err)

	assert.Equal(t, Input, string(buf[0:n]))

	os.Stdin = oldStdin
}

// Bytes written to w are read from os.Stdin
func mockStdin(t *testing.T) (w *os.File, oldStdin *os.File) {
	oldStdin = os.Stdin
	r, w, err := os.Pipe()
	assert.Nil(t, err)
	os.Stdin = r
	return
}
