package wireguard

import (
	"bytes"
	"fmt"
	"net"
	"net/netip"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

const (
	DefaultMtu = 1420
)

type Tunnel struct {
	net    *netstack.Net
	dev    *device.Device
	config *Config
}

func resolveEndpoint(e Endpoint) (string, error) {
	if e.Host == "" {
		return "", nil
	}
	resolvedIps, err := net.LookupIP(e.Host)
	if err != nil {
		return "", err
	}
	for _, r := range resolvedIps {
		if r4 := r.To4(); len(r4) == net.IPv4len {
			return fmt.Sprintf("%s:%d", r.String(), e.Port), nil
		} else {
			return fmt.Sprintf("[%s]:%d", r.String(), e.Port), nil
		}
	}
	return "", fmt.Errorf("unable to resolve %s, got: %v", e.Host, resolvedIps)
}

func CreateTunnelFromFile(file string) (*Tunnel, error) {
	config, err := FromWgQuickFile(file, "wg-nc")
	if err != nil {
		return nil, err
	}
	return CreateTunnel(config)
}

func CreateTunnel(config *Config) (*Tunnel, error) {
	localIps := []netip.Addr{}
	for _, p := range config.Interface.Addresses {
		localIps = append(localIps, p.Addr())
	}

	mtu := config.Interface.MTU
	if mtu == 0 {
		mtu = DefaultMtu
	}
	// Create the virtual TUN device
	tunDev, gNet, err := netstack.CreateNetTUN(localIps, config.Interface.DNS, int(mtu))
	if err != nil {
		return nil, err
	}

	// Setup the actual wireguard
	wgDev := device.NewDevice(
		tunDev, conn.NewDefaultBind(),
		device.NewLogger(device.LogLevelError, fmt.Sprintf("(%s)", config.Name)),
	)

	wgConf := bytes.NewBuffer(nil)
	fmt.Fprintf(wgConf, "private_key=%s\n", config.Interface.PrivateKey.ToHex())
	if config.Interface.ListenPort != 0 {
		fmt.Fprintf(wgConf, "listen_port=%d\n", config.Interface.ListenPort)
	}

	fmt.Fprintf(wgConf, "replace_peers=true\n")

	for _, p := range config.Peers {
		endpointAddr, err := resolveEndpoint(p.Endpoint)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(wgConf, "public_key=%s\n", p.PublicKey.ToHex())
		fmt.Fprintf(wgConf, "preshared_key=%s\n", p.PresharedKey.ToHex())
		if endpointAddr != "" {
			fmt.Fprintf(wgConf, "endpoint=%s\n", endpointAddr)
		}
		for _, prefix := range p.AllowedIPs {
			fmt.Fprintf(wgConf, "allowed_ip=%s\n", prefix.String())
		}
		fmt.Fprintf(wgConf, "persistent_keepalive_interval=%d\n", p.PersistentKeepalive)
	}

	if err := wgDev.IpcSetOperation(wgConf); err != nil {
		return nil, err
	}

	wgDev.Up()

	return &Tunnel{
		net:    gNet,
		dev:    wgDev,
		config: config,
	}, nil
}

func (t *Tunnel) Dial(network, addr string) (net.Conn, error) {
	return t.net.Dial(network, addr)
}

func (t *Tunnel) Listen(proto string, address string) (net.Listener, error) {
	if proto != "tcp" {
		return nil, fmt.Errorf("only tcp proto is supported")
	}
	if address == "" {
		// nothing is specified, listen on random port
		return t.net.ListenTCP(nil)
	}
	host, portName, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	var ipAddr netip.Addr

	if host != "" {
		addrs, err := t.net.LookupHost(host)
		if err != nil {
			return nil, err
		}
		ipAddr, err = netip.ParseAddr(addrs[0])
		if err != nil {
			return nil, err
		}
	} else {
		ipAddr = t.config.Interface.Addresses[0].Addr()
	}

	port, err := net.LookupPort(proto, portName)
	if err != nil {
		return nil, err
	}

	addrPort := netip.AddrPortFrom(ipAddr, uint16(port))
	tcpAddr := net.TCPAddrFromAddrPort(addrPort)
	return t.net.ListenTCP(tcpAddr)
}
