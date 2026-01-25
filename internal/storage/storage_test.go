package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataDirCreatesLocalDataDir(t *testing.T) {
	tmp := t.TempDir()

	wd, err := os.Getwd()
	assert.NoError(t, err)
	defer func() {
		_ = os.Chdir(wd)
	}()
	assert.NoError(t, os.Chdir(tmp))

	got, err := DataDir("Gallery")
	assert.NoError(t, err)

	cwd, err := os.Getwd()
	assert.NoError(t, err)
	want := filepath.Join(cwd, ".data")
	assert.Equal(t, want, got)

	_, err = os.Stat(got)
	assert.NoError(t, err)
}
