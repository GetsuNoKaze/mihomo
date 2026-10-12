package statistic_test

import (
	"io"
	"net"
	"net/netip"
	"testing"

	"github.com/metacubex/mihomo/adapter/outbound"
	"github.com/metacubex/mihomo/common/buf"
	N "github.com/metacubex/mihomo/common/net"
	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/tunnel/statistic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func userMetadata(inbound, user, src string) *C.Metadata {
	return &C.Metadata{
		NetWork: C.TCP,
		Type:    C.INNER,
		SrcIP:   netip.MustParseAddr(src),
		SrcPort: 12345,
		DstIP:   netip.MustParseAddr("1.2.3.4"),
		DstPort: 443,
		InName:  inbound,
		InUser:  user,
	}
}

func userStat(t *testing.T, inbound, user string) statistic.UserStat {
	t.Helper()
	s, ok := statistic.DefaultManager.Users().Get(statistic.UserKey{Inbound: inbound, User: user})
	require.True(t, ok, "no counter for %s/%s", inbound, user)
	return s
}

// tcpPair returns a tracked conn over a pipe and the remote end of the pipe.
func tcpPair(t *testing.T, metadata *C.Metadata, peek int64) (C.Conn, net.Conn) {
	local, remote := net.Pipe()
	t.Cleanup(func() {
		_ = local.Close()
		_ = remote.Close()
	})
	conn := outbound.NewConn(local, outbound.NewDirect())
	return statistic.NewTCPTracker(conn, statistic.DefaultManager, metadata, nil, peek, 0, true), remote
}

func TestTCPTrackerUserCounting(t *testing.T) {
	const inbound = "tcp-tracker-test"
	tracker, remote := tcpPair(t, userMetadata(inbound, "alice", "::ffff:10.1.1.1"), 7)

	go func() {
		b := make([]byte, 64)
		_, _ = io.ReadFull(remote, b[:10]) // Write
		_, _ = io.ReadFull(remote, b[:20]) // WriteBuffer
		_, _ = remote.Write(make([]byte, 30))
		_, _ = remote.Write(make([]byte, 40))
	}()

	_, err := tracker.Write(make([]byte, 10))
	require.NoError(t, err)
	buffer := buf.NewSize(20)
	buffer.Extend(20)
	require.NoError(t, tracker.WriteBuffer(buffer))
	_, err = io.ReadFull(tracker, make([]byte, 30))
	require.NoError(t, err)
	rb := buf.NewSize(64)
	defer rb.Release()
	require.NoError(t, tracker.ReadBuffer(rb))

	s := userStat(t, inbound, "alice")
	assert.Equal(t, int64(7+10+20), s.Up, "peek + Write + WriteBuffer")
	assert.Equal(t, int64(30+rb.Len()), s.Down)
	assert.Equal(t, int64(1), s.Conns)
	require.Len(t, s.IPs, 1)
	assert.Equal(t, netip.MustParseAddr("10.1.1.1"), s.IPs[0].IP, "a mapped IPv4 address is unmapped")

	// the splice/copy path reports through the count functions
	_, readCounters := tracker.(interface {
		UnwrapReader() (io.Reader, []N.CountFunc)
	}).UnwrapReader()
	_, writeCounters := tracker.(interface {
		UnwrapWriter() (io.Writer, []N.CountFunc)
	}).UnwrapWriter()
	for _, f := range readCounters {
		f(1000)
	}
	for _, f := range writeCounters {
		f(500)
	}
	s2 := userStat(t, inbound, "alice")
	assert.Equal(t, s.Up+500, s2.Up)
	assert.Equal(t, s.Down+1000, s2.Down)

	// Close is idempotent: the connection and the address are released once
	require.NoError(t, tracker.Close())
	_ = tracker.Close()
	s = userStat(t, inbound, "alice")
	assert.Equal(t, int64(0), s.Conns)
	require.Len(t, s.IPs, 1)
	assert.Equal(t, int64(0), s.IPs[0].Conns)
}

func TestTrackerWithoutUser(t *testing.T) {
	const inbound = "tracker-anonymous-test"
	before := len(statistic.DefaultManager.Users().SnapshotFunc(func(k statistic.UserKey) bool { return k.Inbound == inbound }))
	tracker, _ := tcpPair(t, userMetadata(inbound, "", "10.1.1.2"), 0)
	defer tracker.Close()
	after := len(statistic.DefaultManager.Users().SnapshotFunc(func(k statistic.UserKey) bool { return k.Inbound == inbound }))
	assert.Equal(t, before, after)
	assert.Equal(t, 0, after)
}

func TestTrackerUserStatsDisabled(t *testing.T) {
	const inbound = "tracker-disabled-test"
	statistic.SetUserStatsEnabled(false)
	defer statistic.SetUserStatsEnabled(true)
	tracker, _ := tcpPair(t, userMetadata(inbound, "alice", "10.1.1.3"), 5)
	defer tracker.Close()
	_, ok := statistic.DefaultManager.Users().Get(statistic.UserKey{Inbound: inbound, User: "alice"})
	assert.False(t, ok)
}

func TestTCPTrackerRelay(t *testing.T) {
	const inbound = "tcp-relay-test"
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()
	const upSize, downSize = 300_000, 700_000
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_, _ = io.CopyN(io.Discard, c, upSize)
		_, _ = c.Write(make([]byte, downSize))
	}()
	raw, err := net.Dial("tcp", ln.Addr().String())
	require.NoError(t, err)
	tracker := statistic.NewTCPTracker(outbound.NewConn(raw, outbound.NewDirect()), statistic.DefaultManager,
		userMetadata(inbound, "bob", "10.1.1.4"), nil, 0, 0, true)

	client, local := net.Pipe()
	go func() {
		_, _ = client.Write(make([]byte, upSize))
		_, _ = io.CopyN(io.Discard, client, downSize)
		_ = client.Close()
	}()
	N.Relay(local, tracker)

	s := userStat(t, inbound, "bob")
	assert.Equal(t, int64(upSize), s.Up)
	assert.Equal(t, int64(downSize), s.Down)
	assert.Equal(t, int64(0), s.Conns)
}

func TestUDPTrackerUserCounting(t *testing.T) {
	const inbound = "udp-tracker-test"
	server, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	defer server.Close()
	go func() {
		b := make([]byte, 2048)
		for {
			n, addr, err := server.ReadFrom(b)
			if err != nil {
				return
			}
			_, _ = server.WriteTo(b[:n*2], addr) // answers twice as much
		}
	}()

	raw, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	metadata := userMetadata(inbound, "carol", "10.1.1.5")
	metadata.NetWork = C.UDP
	tracker := statistic.NewUDPTracker(outbound.NewPacketConn(raw, outbound.NewDirect()), statistic.DefaultManager, metadata, nil, 0, 0, true)

	b := make([]byte, 2048)
	_, err = tracker.WriteTo(make([]byte, 100), server.LocalAddr())
	require.NoError(t, err)
	n, _, err := tracker.ReadFrom(b)
	require.NoError(t, err)
	assert.Equal(t, 200, n)
	_, err = tracker.WriteTo(make([]byte, 50), server.LocalAddr())
	require.NoError(t, err)
	data, put, _, err := tracker.WaitReadFrom()
	require.NoError(t, err)
	assert.Len(t, data, 100)
	if put != nil {
		put()
	}

	s := userStat(t, inbound, "carol")
	assert.Equal(t, int64(150), s.Up)
	assert.Equal(t, int64(300), s.Down)
	assert.Equal(t, int64(1), s.Conns)

	require.NoError(t, tracker.Close())
	_ = tracker.Close()
	s = userStat(t, inbound, "carol")
	assert.Equal(t, int64(0), s.Conns)
}

func TestInternalTrackersAreNotAttributed(t *testing.T) {
	// the trackers of the DNS dialer and of dialer-proxy chains (pushToManager
	// false) sit under a tracker that already counts the user
	const inbound = "internal-tracker-test"
	local, remote := net.Pipe()
	defer remote.Close()
	tracker := statistic.NewTCPTracker(outbound.NewConn(local, outbound.NewDirect()), statistic.DefaultManager,
		userMetadata(inbound, "alice", "10.1.1.6"), nil, 10, 0, false)
	defer tracker.Close()
	_, ok := statistic.DefaultManager.Users().Get(statistic.UserKey{Inbound: inbound, User: "alice"})
	assert.False(t, ok)
}
