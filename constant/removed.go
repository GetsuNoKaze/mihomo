package constant

import "errors"

// ErrRemovedProtocol wraps the errors for protocols and options Citrus dropped
// (ShadowsocksR, Hysteria v1, TUIC v4, Shadowsocks stream ciphers,
// v2ray-plugin). Providers skip such proxies with a warning instead of failing
// the whole provider, so one legacy node does not break a subscription.
var ErrRemovedProtocol = errors.New("removed from Citrus")
