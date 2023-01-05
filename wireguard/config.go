/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 * Original source: https://github.com/WireGuard/wireguard-windows/blob/004c22c5647e5c492daf21d0310cbf575e4e3277/conf/config.go
 */
package wireguard

import (
	"crypto/subtle"
	"encoding/hex"
	"net/netip"
	"time"
)

const KeyLength = 32

type Endpoint struct {
	Host string
	Port uint16
}

type (
	Key           [KeyLength]byte
	HandshakeTime time.Duration
	Bytes         uint64
)

type Config struct {
	Name      string
	Interface Interface
	Peers     []Peer
}

type Interface struct {
	PrivateKey Key
	Addresses  []netip.Prefix
	ListenPort uint16
	MTU        uint16
	DNS        []netip.Addr
	DNSSearch  []string
	PreUp      string
	PostUp     string
	PreDown    string
	PostDown   string
	TableOff   bool
}

type Peer struct {
	PublicKey           Key
	PresharedKey        Key
	AllowedIPs          []netip.Prefix
	Endpoint            Endpoint
	PersistentKeepalive uint16
}

func (k *Key) IsZero() bool {
	var zeros Key
	return subtle.ConstantTimeCompare(zeros[:], k[:]) == 1
}

func (k *Key) ToHex() string {
	return hex.EncodeToString(k[:])
}
