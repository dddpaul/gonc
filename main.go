package main

import (
	"flag"

	"github.com/dddpaul/gonc/tcp"
	"github.com/dddpaul/gonc/udp"
)

func main() {
	var host, port, proto string
	var listen bool
	flag.StringVar(&host, "host", "", "Remote host to connect, i.e. 127.0.0.1")
	flag.StringVar(&proto, "proto", "tcp", "TCP/UDP mode")
	flag.BoolVar(&listen, "listen", false, "Listen mode")
	flag.StringVar(&port, "port", ":9999", "Port to listen on or connect to (prepended by colon), i.e. :9999")
	flag.Parse()

	switch proto {
	case "tcp":
		if listen {
			tcp.StartServer(proto, port)
		} else if host != "" {
			tcp.StartClient(proto, host, port)
		} else {
			flag.Usage()
		}
	case "udp":
		if listen {
			udp.StartServer(proto, port)
		} else if host != "" {
			udp.StartClient(proto, host, port)
		} else {
			flag.Usage()
		}
	default:
		flag.Usage()
	}
}
