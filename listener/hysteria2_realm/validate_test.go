package hysteria2_realm

import (
	"testing"

	"github.com/metacubex/mihomo/common/httputils"

	"github.com/metacubex/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientIP(t *testing.T) {
	request := func(remote, xff string) *http.Request {
		r := &http.Request{RemoteAddr: remote, Header: http.Header{}}
		if xff != "" {
			r.Header.Set("X-Forwarded-For", xff)
		}
		return r
	}

	// no trusted-proxies: the header is believed from any client, as before
	// the option existed
	legacy, err := httputils.NewRealIP("X-Forwarded-For", nil, true)
	require.NoError(t, err)
	assert.Equal(t, "6.6.6.6", clientIP(request("198.51.100.1:1000", "6.6.6.6, 203.0.113.7"), legacy))
	assert.Equal(t, "198.51.100.1", clientIP(request("198.51.100.1:1000", ""), legacy))

	trusted, err := httputils.NewRealIP("X-Forwarded-For", []string{"loopback"}, true)
	require.NoError(t, err)
	assert.Equal(t, "198.51.100.1", clientIP(request("198.51.100.1:1000", "203.0.113.7"), trusted))
	assert.Equal(t, "203.0.113.7", clientIP(request("127.0.0.1:1000", "6.6.6.6, 203.0.113.7"), trusted))
	assert.Equal(t, "127.0.0.1", clientIP(request("[::ffff:127.0.0.1]:1000", ""), trusted))

	off, err := httputils.NewRealIP("", nil, true)
	require.NoError(t, err)
	assert.Equal(t, "198.51.100.1", clientIP(request("198.51.100.1:1000", "203.0.113.7"), off))
}
