package constant

import (
	"context"
	"net"
)

type Listener interface {
	RawAddress() string
	Address() string
	Close() error
}

type MultiAddrListener interface {
	Close() error
	Config() string
	AddrList() (addrList []net.Addr)
}

type InboundListener interface {
	Name() string
	Listen(tunnel Tunnel) error
	Close() error
	Address() string
	RawAddress() string
	Config() InboundConfig
}

// InboundUserLister is implemented by the inbound listeners that know the
// complete list of their users, each as it appears in Metadata.InUser. The
// per-user traffic counters of the users that are no longer listed are
// dropped after a configuration reload.
type InboundUserLister interface {
	InboundUsers() []string
}

type InboundConfig interface {
	Name() string
	Equal(config InboundConfig) bool
}

type InboundListenConfig interface {
	Listen(ctx context.Context, network, address string) (net.Listener, error)
	ListenPacket(ctx context.Context, network, address string) (net.PacketConn, error)
}
