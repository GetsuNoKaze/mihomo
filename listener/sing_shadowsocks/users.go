package sing_shadowsocks

import (
	"encoding/base64"
	"fmt"

	LC "github.com/metacubex/mihomo/listener/config"
)

// multiUserKeyLength lists the Shadowsocks 2022 methods that support
// multiple users through the extensible identity header (EIH) and their key
// length. The chacha20/chacha8 2022 methods have no EIH.
var multiUserKeyLength = map[string]int{
	"2022-blake3-aes-128-gcm": 16,
	"2022-blake3-aes-256-gcm": 32,
	"2022-blake3-aes-128-ccm": 16,
	"2022-blake3-aes-256-ccm": 32,
}

// ValidateUsers checks the multi-user part of a Shadowsocks server config:
// users are only allowed with the AES Shadowsocks 2022 methods, the server
// key (iPSK) and every user key (uPSK) must be base64 of exactly the method's
// key length, and user names and keys must be unique. No users is valid.
func ValidateUsers(cipher, password string, users []LC.ShadowsocksUser) error {
	if len(users) == 0 {
		return nil
	}
	keyLength, ok := multiUserKeyLength[cipher]
	if !ok {
		return fmt.Errorf("shadowsocks: users are only supported with 2022-blake3-aes-* methods, not %q", cipher)
	}
	if err := checkKey(password, keyLength); err != nil {
		return fmt.Errorf("shadowsocks: server password (iPSK): %w", err)
	}
	names := make(map[string]struct{}, len(users))
	keys := make(map[string]string, len(users)+1)
	keys[password] = ""
	for i, user := range users {
		if user.Name == "" {
			return fmt.Errorf("shadowsocks: users[%d]: empty name", i)
		}
		if _, dup := names[user.Name]; dup {
			return fmt.Errorf("shadowsocks: duplicate user name %q", user.Name)
		}
		names[user.Name] = struct{}{}
		if err := checkKey(user.Password, keyLength); err != nil {
			return fmt.Errorf("shadowsocks: user %q password (uPSK): %w", user.Name, err)
		}
		if other, dup := keys[user.Password]; dup {
			if other == "" {
				return fmt.Errorf("shadowsocks: user %q password equals the server password", user.Name)
			}
			return fmt.Errorf("shadowsocks: users %q and %q share a password", other, user.Name)
		}
		keys[user.Password] = user.Name
	}
	return nil
}

func checkKey(key string, length int) error {
	if key == "" {
		return fmt.Errorf("missing")
	}
	b, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("decode base64: %w", err)
	}
	if len(b) != length {
		return fmt.Errorf("bad key length, required %d bytes, got %d", length, len(b))
	}
	return nil
}
