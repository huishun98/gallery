package main

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func withStdin(t *testing.T, input string, fn func()) {
	t.Helper()

	r, w, err := os.Pipe()
	assert.NoError(t, err)
	_, _ = w.Write([]byte(input))
	_ = w.Close()

	old := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = old
		_ = r.Close()
	}()

	fn()
}

func TestWaitForPortReady(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	defer ln.Close()

	err = waitForPort(ln.Addr().String(), 2*time.Second)
	assert.NoError(t, err)
}

func TestWaitForPortTimeout(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	addr := ln.Addr().String()
	_ = ln.Close()

	err = waitForPort(addr, 200*time.Millisecond)
	assert.Error(t, err)
}

func TestPromptUserInputsDefaults(t *testing.T) {
	tmp := t.TempDir()
	defaultDir := filepath.Join(tmp, "default")

	withStdin(t, "Y\n", func() {
		got, err := promptUserInputs(defaultDir)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, defaultDir, got.DataDir)
			assert.Equal(t, "8000", got.Port)
			assert.Nil(t, got.Admin)
		}
	})
}

func TestPromptUserInputsCustomNoAdmin(t *testing.T) {
	tmp := t.TempDir()
	defaultDir := filepath.Join(tmp, "default")
	customDir := filepath.Join(tmp, "custom")

	withStdin(t, "n\n9090\n"+customDir+"\nn\n", func() {
		got, err := promptUserInputs(defaultDir)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, customDir, got.DataDir)
			assert.Equal(t, "9090", got.Port)
			assert.Nil(t, got.Admin)
		}

		_, statErr := os.Stat(customDir)
		assert.NoError(t, statErr)
	})
}
