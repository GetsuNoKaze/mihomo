package socks

import (
	"net"
	"testing"
	"time"

	"github.com/metacubex/mihomo/adapter/inbound"
	"github.com/metacubex/mihomo/component/auth"
	C "github.com/metacubex/mihomo/constant"
	authStore "github.com/metacubex/mihomo/listener/auth"
	LC "github.com/metacubex/mihomo/listener/config"
	"github.com/metacubex/mihomo/transport/socks5"

	"github.com/stretchr/testify/require"
)

type udpRecorder struct {
	packets chan *C.Metadata
}

func (t *udpRecorder) HandleTCPConn(conn net.Conn, metadata *C.Metadata) { _ = conn.Close() }

func (t *udpRecorder) HandleUDPPacket(packet C.UDPPacket, metadata *C.Metadata) {
	packet.Drop()
	t.packets <- metadata
}

func (t *udpRecorder) NatTable() C.NatTable { return nil }

func (t *udpRecorder) received(timeout time.Duration) bool {
	select {
	case <-t.packets:
		return true
	case <-time.After(timeout):
		return false
	}
}

func newSocksPair(t *testing.T, store auth.AuthStore) (*Listener, *UDPListener, *udpRecorder) {
	t.Helper()
	tunnel := &udpRecorder{packets: make(chan *C.Metadata, 16)}
	additions := []inbound.Addition{inbound.WithInName("test-socks")}

	tcp, err := NewWithConfig(LC.AuthServer{Enable: true, Listen: "127.0.0.1:0", AuthStore: store}, inbound.NewListenConfig(), tunnel, additions...)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tcp.Close() })

	// The UDP listener shares the TCP listener's port, as the real inbounds do.
	udp, err := NewUDPWithConfig(LC.AuthServer{Enable: true, Listen: tcp.Address(), AuthStore: store}, inbound.NewListenConfig(), tunnel, additions...)
	require.NoError(t, err)
	t.Cleanup(func() { _ = udp.Close() })
	return tcp, udp, tunnel
}

func sendSocksUDP(t *testing.T, client net.PacketConn, to string) {
	t.Helper()
	pkt, err := socks5.EncodeUDPPacket(socks5.ParseAddr("1.1.1.1:53"), []byte("ping"))
	require.NoError(t, err)
	addr, err := net.ResolveUDPAddr("udp", to)
	require.NoError(t, err)
	_, err = client.WriteTo(pkt, addr)
	require.NoError(t, err)
}

func associate(t *testing.T, addr string, user *socks5.User) net.Conn {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	_, err = socks5.ClientHandshake(conn, socks5.ParseAddr("0.0.0.0:0"), socks5.CmdUDPAssociate, user)
	require.NoError(t, err)
	return conn
}

func waitAssociation(t *testing.T, udp *UDPListener, client net.PacketConn, want bool) {
	t.Helper()
	require.Eventually(t, func() bool {
		return hasAssociation(udp.packetConn.LocalAddr(), client.LocalAddr()) == want
	}, 2*time.Second, 10*time.Millisecond)
}

func TestSocksUDPRequiresAssociationWithAuth(t *testing.T) {
	store := authStore.NewAuthStore(auth.NewAuthenticator([]auth.AuthUser{{User: "u", Pass: "p"}}))
	tcp, udp, tunnel := newSocksPair(t, store)

	client, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	defer client.Close()

	// No association: dropped.
	sendSocksUDP(t, client, udp.Address())
	require.False(t, tunnel.received(300*time.Millisecond), "datagram relayed without UDP ASSOCIATE")

	// Wrong credentials: the handshake fails and nothing is associated.
	bad, err := net.Dial("tcp", tcp.Address())
	require.NoError(t, err)
	_, err = socks5.ClientHandshake(bad, socks5.ParseAddr("0.0.0.0:0"), socks5.CmdUDPAssociate, &socks5.User{Username: "u", Password: "wrong"})
	require.Error(t, err)
	_ = bad.Close()
	sendSocksUDP(t, client, udp.Address())
	require.False(t, tunnel.received(300*time.Millisecond), "datagram relayed after a failed handshake")

	// Authenticated association: relayed.
	control := associate(t, tcp.Address(), &socks5.User{Username: "u", Password: "p"})
	waitAssociation(t, udp, client, true)
	sendSocksUDP(t, client, udp.Address())
	require.True(t, tunnel.received(2*time.Second), "datagram not relayed with an open association")

	// Control connection closed: dropped again.
	require.NoError(t, control.Close())
	waitAssociation(t, udp, client, false)
	sendSocksUDP(t, client, udp.Address())
	require.False(t, tunnel.received(300*time.Millisecond), "datagram relayed after the association ended")
}

func TestSocksUDPWithoutAuthNeedsNoAssociation(t *testing.T) {
	_, udp, tunnel := newSocksPair(t, authStore.Nil)

	client, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	defer client.Close()

	sendSocksUDP(t, client, udp.Address())
	require.True(t, tunnel.received(2*time.Second), "unauthenticated listener should keep relaying as before")
}

func TestAssociationRefCount(t *testing.T) {
	local := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1080}
	remote := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 2), Port: 5555}
	udpFrom := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 2), Port: 6000}
	udpLocal := &net.UDPAddr{IP: net.IPv4zero, Port: 1080}

	r1 := addAssociation(local, remote)
	r2 := addAssociation(local, remote)
	require.True(t, hasAssociation(udpLocal, udpFrom))
	r1()
	r1() // idempotent
	require.True(t, hasAssociation(udpLocal, udpFrom))
	r2()
	require.False(t, hasAssociation(udpLocal, udpFrom))

	// IPv4-mapped addresses match their IPv4 form.
	mapped := &net.TCPAddr{IP: net.ParseIP("::ffff:127.0.0.2"), Port: 5555}
	r3 := addAssociation(local, mapped)
	require.True(t, hasAssociation(udpLocal, udpFrom))
	r3()
}
