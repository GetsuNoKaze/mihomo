package outbound

import (
	"testing"
	"time"

	"github.com/metacubex/mihomo/transport/xhttp"

	"github.com/stretchr/testify/assert"
)

func TestXHTTPReuseSettingsDefault(t *testing.T) {
	for _, settings := range []*XHTTPReuseSettings{nil, {}} {
		cfg, keepAlive := settings.build()
		assert.Equal(t, xhttp.DefaultReuseConfig(), cfg)
		assert.Zero(t, keepAlive)
	}

	// explicit settings stay authoritative, even when they disable the reuse
	cfg, keepAlive := (&XHTTPReuseSettings{MaxConnections: "0", HKeepAlivePeriod: 30}).build()
	assert.Equal(t, &xhttp.ReuseConfig{MaxConnections: "0"}, cfg)
	assert.Equal(t, 30*time.Second, keepAlive)
}
