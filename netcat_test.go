package main

import (
	"net"
	"testing"
	"time"
	"strings"
)

var Host = "localhost"
var Port = ":9999"
var Input = "Input from other side, пока, £, 语汉"

func TestListen(t *testing.T) {
	// Send test data to listener from goroutine
	go func() {
		time.Sleep( 500 * time.Millisecond )
		con, err := net.Dial("tcp", Host+Port)
		if( err != nil ) {
			t.Error(err)
			return
		}
		c := handle(strings.NewReader(Input), con)
		if !<-c {
			t.Error("handle() must write to result channel")
		}
	}()

	ln, err := net.Listen("tcp", Port)
	if err != nil {
		t.Error(err)
		return
	}
	con, err := ln.Accept()
	if err != nil {
		t.Error(err)
		return
	}
	buf := make([]byte, 1024)
	n, err := con.Read(buf)
	if err != nil {
		t.Error(err)
		return
	}
	if string(buf[0:n]) != Input {
		t.Error("Received string is invalid")
		return
	}
}
