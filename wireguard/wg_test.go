package wireguard

import (
	"bytes"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func echoServer(t *testing.T, listener net.Listener) {
	conn, err := listener.Accept()
	assert.NoError(t, err)
	defer conn.Close()

	for {
		_, err := io.Copy(conn, conn)
		if err != nil {
			break
		}
	}
	assert.NoError(t, err)
}

func sendAndReceive(t *testing.T, conn net.Conn, input string) string {
	_, err := io.Copy(conn, bytes.NewBufferString(input))
	assert.NoError(t, err)

	received := bytes.Buffer{}
	_, err = io.CopyN(&received, conn, int64(len(input)))
	assert.NoError(t, err)

	err = conn.Close()
	assert.NoError(t, err)

	return received.String()
}

func TestWireguard(t *testing.T) {
	configA, err := FromWgQuick(`
[Interface]
PrivateKey = 2OZeP9sbnTBiyn1+43610zdMHhhE3CpaBJFxRJl5gGI=
Address = 10.0.0.1
ListenPort = 43234

[Peer]
PublicKey = fw2pUc5mHyrSLe43NG+Rb90isqFKnKmK2Et0Ma76CkY=
AllowedIPs = 10.0.0.2/32
`, "tunnelA")
	assert.NoError(t, err)

	configB, err := FromWgQuick(`
[Interface]
PrivateKey = kBXqMKQPlxmJPuxCxsmd+xuoQxZQocKlI2w1sB8zFnI=
Address = 10.0.0.2

[Peer]
PublicKey = h761vZ6TghHSmFuuEsAXRMJj8WLHkGhfyQXLcaXS2Xs=
AllowedIPs = 10.0.0.1/32
Endpoint = localhost:43234
`, "tunnelB")
	assert.NoError(t, err)

	tunnelA, err := CreateTunnel(configA)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	tunnelB, err := CreateTunnel(configB)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	t.Run("explicit port, explicit IP", func(t *testing.T) {
		listener, err := tunnelA.Listen("tcp", "10.0.0.1:43235")
		assert.NoError(t, err)

		go echoServer(t, listener)

		conn, err := tunnelB.Dial("tcp", listener.Addr().String())
		assert.NoError(t, err)

		input := "Test string"
		result := sendAndReceive(t, conn, input)

		assert.Equal(t, input, result)
	})

	t.Run("explicit port, implicit IP", func(t *testing.T) {
		listener, err := tunnelA.Listen("tcp", ":43236")
		assert.NoError(t, err)

		go echoServer(t, listener)

		conn, err := tunnelB.Dial("tcp", listener.Addr().String())
		assert.NoError(t, err)

		input := "Test string"
		result := sendAndReceive(t, conn, input)

		assert.Equal(t, input, result)
	})

	t.Run("implicit port, implicit IP", func(t *testing.T) {
		listener, err := tunnelA.Listen("tcp", "")
		assert.NoError(t, err)

		go echoServer(t, listener)

		// the netstack doesn't include the host name, only the random port
		_, port, err := net.SplitHostPort(listener.Addr().String())
		assert.NoError(t, err)

		conn, err := tunnelB.Dial("tcp", "10.0.0.1:"+port)
		assert.NoError(t, err)

		input := "Test string"
		result := sendAndReceive(t, conn, input)

		assert.Equal(t, input, result)
	})
}
