package xhttp

import (
	"testing"

	"github.com/metacubex/http"
	"github.com/metacubex/http/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXPaddingBytesZeroMeansDefault(t *testing.T) {
	for _, value := range []string{"", "0", "0-0"} {
		r, err := (&Config{XPaddingBytes: value}).GetNormalizedXPaddingBytes()
		require.NoError(t, err, value)
		assert.Equal(t, Range{100, 1000}, r, value)
	}

	r, err := (&Config{XPaddingBytes: "16-32"}).GetNormalizedXPaddingBytes()
	require.NoError(t, err)
	assert.Equal(t, Range{16, 32}, r)

	// a client with "0-0" pads its requests, so a server with "0-0" accepts them
	cfg := &Config{Path: "/xhttp", XPaddingBytes: "0-0"}
	req := httptest.NewRequest(http.MethodGet, "https://example.com/xhttp/session", nil)
	require.NoError(t, cfg.FillStreamRequest(req, ""))
	assert.Contains(t, req.Header.Get("Referer"), "x_padding=")
}

func TestScRangesZeroMeansDefault(t *testing.T) {
	cfg := &Config{
		ScMaxEachPostBytes:   "0",
		ScMinPostsIntervalMs: "0-0",
		ScMaxBufferedPosts:   "0",
		ScStreamUpServerSecs: "0",
	}
	r, err := cfg.GetNormalizedScMaxEachPostBytes()
	require.NoError(t, err)
	assert.Equal(t, Range{1000000, 1000000}, r)
	r, err = cfg.GetNormalizedScMinPostsIntervalMs()
	require.NoError(t, err)
	assert.Equal(t, Range{30, 30}, r)
	r, err = cfg.GetNormalizedScMaxBufferedPosts()
	require.NoError(t, err)
	assert.Equal(t, Range{30, 30}, r)
	r, err = cfg.GetNormalizedScStreamUpServerSecs()
	require.NoError(t, err)
	assert.Equal(t, Range{20, 80}, r)

	_, err = (&Config{ScMaxEachPostBytes: "0-100"}).GetNormalizedScMaxEachPostBytes()
	assert.Error(t, err)
}
