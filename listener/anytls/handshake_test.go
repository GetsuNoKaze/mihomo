package anytls

import (
	"crypto/sha256"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/listener/inbounduser"
	"github.com/metacubex/mihomo/transport/anytls/padding"
)

// serveAnyTLSOnce accepts one TCP connection and serves it with l.HandleConn.
func serveAnyTLSOnce(t *testing.T) net.Conn {
	t.Helper()
	users, _ := inbounduser.New(userKey, []C.InboundUser{{Name: "test", Secret: "password"}})
	l := &Listener{users: users}
	padding.UpdatePaddingScheme(padding.DefaultPaddingScheme, &l.padding)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		l.HandleConn(c, nil) // no stream is opened, so no handler is needed
	}()
	client, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func shortHandshakeTimeout(t *testing.T, d time.Duration) {
	old := handshakeTimeout
	handshakeTimeout = d
	t.Cleanup(func() { handshakeTimeout = old })
}

func TestAnyTLSHandshakeTimeoutSilentClient(t *testing.T) {
	shortHandshakeTimeout(t, 200*time.Millisecond)
	client := serveAnyTLSOnce(t)
	_ = client.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err := client.Read(make([]byte, 1))
	if errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatal("server kept the unauthenticated conn open")
	}
	if err == nil {
		t.Fatal("server sent data to an unauthenticated client")
	}
}

func TestAnyTLSHandshakeDeadlineClearedAfterAuth(t *testing.T) {
	timeout := 200 * time.Millisecond
	shortHandshakeTimeout(t, timeout)
	client := serveAnyTLSOnce(t)
	hash := sha256.Sum256([]byte("password"))
	if _, err := client.Write(append(hash[:], 0, 0)); err != nil { // no padding
		t.Fatal(err)
	}
	// the session waits for frames; an idle authenticated conn must stay open
	_ = client.SetReadDeadline(time.Now().Add(4 * timeout))
	_, err := client.Read(make([]byte, 1))
	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("authenticated idle session was cut: %v", err)
	}
}
