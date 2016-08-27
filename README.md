gonc
=========

Simple Netcat tool written in Go.

Install:

```
go get -u github.com/dddpaul/gonc
```

Usage:

```
gonc [OPTIONS]
  -host="": Remote host to connect, i.e. 127.0.0.1
  -listen=false: Listen mode
  -port="": Port to listen on or connect to (prepended by colon), i.e. :9999
  -proto="tcp": TCP/UDP mode
```

Comments:

* Send `~.` to disconnect in UDP mode.

You can grab some binaries from [Bintray](http://dl.bintray.com/dddpaul/generic/gonc/). This is a simplest way to get Netcat for Windows :)
