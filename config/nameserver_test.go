package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNameServerKeepsDoHUserinfo(t *testing.T) {
	servers, err := parseNameServer([]string{"https://user:s3cret@dns.example.com/dns-query"}, false, false)
	require.NoError(t, err)
	require.Len(t, servers, 1)
	assert.Equal(t, "https", servers[0].Net)
	assert.Equal(t, "https://user:s3cret@dns.example.com:443/dns-query", servers[0].Addr)
}

func TestParseNameServerErrorHidesUserinfo(t *testing.T) {
	_, err := parseNameServer([]string{"https://user:s3cret@[::1/dns-query"}, false, false)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "s3cret")
}
