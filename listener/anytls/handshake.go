package anytls

import (
	"net"
	"time"
)

// handshakeTimeout bounds how long a client may take, from the moment its
// connection reaches the AnyTLS layer, to send the authentication prelude
// (password hash and padding). Without it a client that connects and sends
// nothing holds a goroutine forever before it is authenticated. The deadline
// is cleared once the prelude is read, so an idle session after that is never
// affected (the session has its own idle handling). The value matches the
// VLESS and Trojan listeners. A variable only so that tests can shorten it.
var handshakeTimeout = 15 * time.Second

func setHandshakeDeadline(conn net.Conn) {
	_ = conn.SetReadDeadline(time.Now().Add(handshakeTimeout))
}

func clearHandshakeDeadline(conn net.Conn) {
	_ = conn.SetReadDeadline(time.Time{})
}
