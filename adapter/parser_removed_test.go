package adapter

import (
	"errors"
	"testing"

	C "github.com/metacubex/mihomo/constant"
)

func TestParseProxyRemovedProtocols(t *testing.T) {
	cases := map[string]map[string]any{
		"ssr":      {"name": "a", "type": "ssr", "server": "127.0.0.1", "port": 443},
		"hysteria": {"name": "b", "type": "hysteria", "server": "127.0.0.1", "port": 443},
		"tuic v4":  {"name": "c", "type": "tuic", "server": "127.0.0.1", "port": 443, "token": "t"},
		"stream cipher": {"name": "d", "type": "ss", "server": "127.0.0.1", "port": 443,
			"cipher": "rc4-md5", "password": "p"},
		"v2ray-plugin": {"name": "e", "type": "ss", "server": "127.0.0.1", "port": 443,
			"cipher": "aes-128-gcm", "password": "p", "plugin": "v2ray-plugin"},
	}
	for name, mapping := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ParseProxy(mapping)
			if !errors.Is(err, C.ErrRemovedProtocol) {
				t.Fatalf("want ErrRemovedProtocol, got %v", err)
			}
		})
	}

	_, err := ParseProxy(map[string]any{"name": "f", "type": "nope", "server": "127.0.0.1", "port": 443})
	if err == nil || errors.Is(err, C.ErrRemovedProtocol) {
		t.Fatalf("unknown type must not be treated as removed: %v", err)
	}
}
