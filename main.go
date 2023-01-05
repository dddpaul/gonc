package main

import (
	"flag"
	"log"
	"net"

	"github.com/dnachev/wg-nc/tcp"
	"github.com/dnachev/wg-nc/udp"
	wg "github.com/dnachev/wg-nc/wireguard"
)

func main() {
	var host, port, proto, wgFile string
	var listenMode bool
	flag.StringVar(&host, "host", "", "Remote host to connect, i.e. 127.0.0.1")
	flag.StringVar(&proto, "proto", "tcp", "TCP/UDP mode")
	flag.BoolVar(&listenMode, "listen", false, "Listen mode")
	flag.StringVar(&port, "port", ":9999", "Port to listen on or connect to (prepended by colon), i.e. :9999")
	flag.StringVar(&wgFile, "wg", "", "Wireguard config file")
	flag.Parse()

	dial := net.Dial
	listen := net.Listen

	if wgFile != "" {
		if proto != "tcp" || listenMode {
			log.Fatalln("Wireguard is supported only for TCP connect mode")
		}
		tunnel, err := wg.CreateTunnelFromFile(wgFile)
		if err != nil {
			log.Fatalln(err)
		}
		dial = func(network, addr string) (net.Conn, error) {
			return tunnel.Dial(network, addr)
		}
		listen = func(network, address string) (net.Listener, error) {
			return tunnel.Listen(network, port)
		}
	}

	switch proto {
	case "tcp":
		if listenMode {
			tcp.StartServer(listen, proto, port)
		} else if host != "" {
			tcp.StartClient(dial, proto, host, port)
		} else {
			flag.Usage()
		}
	case "udp":
		if listenMode {
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
