package dns

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoHAddressRedactsUserinfo(t *testing.T) {
	client := newDoHClient("https://user:s3cret@dns.example.com:443/dns-query", nil, false, map[string]string{}, nil, "")
	doh := client.(*dnsOverHTTPS)
	defer doh.Close()

	assert.NotContains(t, doh.Address(), "s3cret")
	assert.Equal(t, "https://user:xxxxx@dns.example.com:443/dns-query", doh.Address())

	// the request itself still carries the credentials
	password, ok := doh.url.User.Password()
	assert.True(t, ok)
	assert.Equal(t, "s3cret", password)
}

func TestDoHAddressWithoutUserinfo(t *testing.T) {
	client := newDoHClient("https://dns.example.com/dns-query", nil, false, map[string]string{}, nil, "")
	doh := client.(*dnsOverHTTPS)
	defer doh.Close()

	assert.Equal(t, "https://dns.example.com/dns-query", doh.Address())
}
