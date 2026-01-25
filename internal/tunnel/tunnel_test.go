package tunnel

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStartTunnelFindsURL(t *testing.T) {
	binDir := t.TempDir()
	script := filepath.Join(binDir, "cloudflared")
	assert.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho https://abc.trycloudflare.com 1>&2\nsleep 5\n"), 0755))
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	got, err := StartTunnel(ctx, "http://localhost:8080")
	if !assert.NoError(t, err) {
		return
	}
	defer got.Close()

	assert.Equal(t, "https://abc.trycloudflare.com", got.PublicURL)
}
