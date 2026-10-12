package socks

import (
	"net"
	"net/netip"
	"sync"
)

// UDP ASSOCIATE bookkeeping (RFC 1928 §7): a SOCKS5 UDP relay should only carry
// datagrams for clients that negotiated an association over TCP, and only while
// that TCP control connection is open. Associations are keyed by the listener's
// local port (the TCP and UDP listeners of one inbound share it) and the client's
// IP address.

type associationKey struct {
	port uint16
	ip   netip.Addr
}

var associations = struct {
	sync.RWMutex
	count map[associationKey]int
}{count: make(map[associationKey]int)}

func newAssociationKey(local, remote net.Addr) (associationKey, bool) {
	localAddr, ok := addrPort(local)
	if !ok {
		return associationKey{}, false
	}
	remoteAddr, ok := addrPort(remote)
	if !ok {
		return associationKey{}, false
	}
	return associationKey{port: localAddr.Port(), ip: remoteAddr.Addr()}, true
}

func addrPort(addr net.Addr) (netip.AddrPort, bool) {
	var ap netip.AddrPort
	switch a := addr.(type) {
	case *net.TCPAddr:
		ap = a.AddrPort()
	case *net.UDPAddr:
		ap = a.AddrPort()
	default:
		if addr == nil {
			return netip.AddrPort{}, false
		}
		parsed, err := netip.ParseAddrPort(addr.String())
		if err != nil {
			return netip.AddrPort{}, false
		}
		ap = parsed
	}
	if !ap.IsValid() {
		return netip.AddrPort{}, false
	}
	return netip.AddrPortFrom(ap.Addr().Unmap().WithZone(""), ap.Port()), true
}

// addAssociation records an association for the control connection with the
// given local and remote addresses. The returned func removes it again.
func addAssociation(local, remote net.Addr) (release func()) {
	key, ok := newAssociationKey(local, remote)
	if !ok {
		return func() {}
	}
	associations.Lock()
	associations.count[key]++
	associations.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			associations.Lock()
			if associations.count[key] <= 1 {
				delete(associations.count, key)
			} else {
				associations.count[key]--
			}
			associations.Unlock()
		})
	}
}

// hasAssociation reports whether a datagram that arrived on local from remote
// belongs to an open UDP ASSOCIATE.
func hasAssociation(local, remote net.Addr) bool {
	key, ok := newAssociationKey(local, remote)
	if !ok {
		return false
	}
	associations.RLock()
	defer associations.RUnlock()
	return associations.count[key] > 0
}
