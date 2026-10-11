//go:build unix

package resource

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeWriteCreatesPrivateFiles(t *testing.T) {
	oldMask := syscall.Umask(0)
	defer syscall.Umask(oldMask)

	dir := filepath.Join(t.TempDir(), "providers")
	path := filepath.Join(dir, "provider.yaml")
	require.NoError(t, safeWrite(path, []byte("payload: []")))

	dirInfo, err := os.Stat(dir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())

	fileInfo, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())
}
