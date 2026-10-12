// Package interop runs Citrus against a real Xray-core binary, both ways round,
// on localhost only.
//
//	cd test/interop
//	XRAY_BIN=/path/to/xray go test -tags interop -v ./...
//
// XRAY_BIN unset: every test is skipped. CITRUS_BIN: a prebuilt Citrus binary
// (default: built from the repository root). INTEROP_LOG_DIR: where the process
// logs go (default: a temporary directory, removed afterwards).
//
// Traffic goes to a local HTTP echo server, and the REALITY target is a local
// TLS 1.3 server whose Certificate message is large enough for ML-DSA-65.
package interop
