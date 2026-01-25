package tunnel

import (
	"bufio"
	"context"
	"errors"
	"gallery/internal/binary"
	"os/exec"
	"regexp"
	"time"
)

var urlRegex = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)

type Tunnel struct {
	Cmd       *exec.Cmd
	PublicURL string
}

// StartTunnel starts cloudflared and blocks until a public URL is available.
func StartTunnel(ctx context.Context, localURL string) (*Tunnel, error) {

	cloudflared, appDir, err := binary.WhereIs("cloudflared")
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(
		ctx,
		cloudflared,
		"tunnel",
		"--url",
		localURL,
		"--no-autoupdate",
	)
	cmd.Dir = appDir

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stderr)
	timeout := time.After(10 * time.Second)

	for {
		select {
		case <-timeout:
			cmd.Process.Kill()
			return nil, errors.New("timeout waiting for Cloudflare tunnel URL")

		default:
			if scanner.Scan() {
				line := scanner.Text()
				if match := urlRegex.FindString(line); match != "" {
					return &Tunnel{
						Cmd:       cmd,
						PublicURL: match,
					}, nil
				}
			}
		}
	}
}

func (t *Tunnel) Close() {
	if t.Cmd != nil && t.Cmd.Process != nil {
		_ = t.Cmd.Process.Kill()
	}
}
