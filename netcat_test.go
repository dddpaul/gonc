package main

import (
	"net"
	"testing"
	"time"
	"strings"
	"errors"
)

var Host = "localhost"
var Port = ":9999"
var Input = "Input from other side, пока, £, 语汉"

func TestTCP(t *testing.T) {
	err := testTCP(t)
	if err != nil {
		t.Error(err)
	}
}

func TestUDP(t *testing.T) {
	err := testUDP(t)
	if err != nil {
		t.Error(err)
	}
}

func testTCP(t *testing.T) error {
	// Send test data to listener from goroutine and wait for potentials errors at the end of the test
	ec := make(chan error)
	go func() {
		// Wait for main thread starts listener
		time.Sleep( 200 * time.Millisecond )
		con, err := net.Dial("tcp", Host+Port)
		if( err != nil ) {
			ec<-err
		}

		// Transfer data
		c1 := handle(strings.NewReader(Input), con)

		// Wait for data will be transferred
		time.Sleep( 200 * time.Millisecond )
		select {
		case <-c1:
		default:
			ec<-errors.New("handle() must write to result channel")
		}
	    ec<-nil
	}()

	ln, err := net.Listen("tcp", Port)
	if err != nil {
		return err
	}

	con, err := ln.Accept()
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	n, err := con.Read(buf)
	if err != nil {
		return err
	}

	if string(buf[0:n]) != Input {
		return errors.New("Received string is invalid")
	}

	// Wait for errors from goroutine
	if err := <-ec; err != nil {
		return err
	}

	return nil
}

func testUDP(t *testing.T) error {
	// Send test data to listener from goroutine and wait for potentials errors at the end of the test
	ec := make(chan error)
	go func() {
		// Wait for main thread starts listener
		time.Sleep( 200 * time.Millisecond )
		con, err := net.Dial("udp", Host+Port)
		if( err != nil ) {
			ec<-err
		}

		// Transfer data
		c1 := handle(strings.NewReader(Input), con)

		// Wait for data will be transferred
		time.Sleep( 200 * time.Millisecond )
		select {
		case <-c1:
		default:
		ec<-errors.New("handle() must write to result channel")
		}
		ec<-nil
	}()

	con, err := net.ListenPacket("udp", Port)
	if err != nil {
		return err
	}

	buf := make([]byte, 1024)
	n, _, err := con.ReadFrom(buf)
	if err != nil {
		return err
	}

	if string(buf[0:n]) != Input {
		return errors.New("Received string is invalid")
	}

	// Wait for errors from goroutine
	if err := <-ec; err != nil {
		return err
	}

	return nil
}
