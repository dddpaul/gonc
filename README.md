wg-nc
=========
Simple Netcat tool written in Go. Fork of [gonc](https://github.com/dddpaul/gonc) with added userland wireguard support.


Install:

```
go get -u github.com/dnachev/wg-nc
```

```
Usage of wg-nc:
  -host string
        Remote host to connect, i.e. 127.0.0.1
  -listen
        Listen mode
  -port string
        Port to listen on or connect to (prepended by colon), i.e. :9999 (default ":9999")
  -proto string
        TCP/UDP mode (default "tcp")
  -wg string
        Wireguard config file
```

Comments:

* Send `~.` to disconnect in UDP mode.
* Wireguard support was inspired by the [Our User-Mode WireGuard Year](https://fly.io/blog/our-user-mode-wireguard-year/) blog post.