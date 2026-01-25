package binary

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func WhereIs(name string) (p string, binDir string, err error) {
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(name), ".exe") {
		name = name + ".exe"
	}

	// 1) Look for bundled cloudflared inside app bundle
	exePath, err := os.Executable()
	if err == nil {
		binDir = filepath.Dir(exePath)
		bundled := filepath.Join(binDir, name)
		if _, err := os.Stat(bundled); err == nil {
			return bundled, binDir, nil
		}
	}

	// 2) Try PATH
	if p, err := exec.LookPath(name); err == nil {
		return p, "", nil
	}

	return "", "", fmt.Errorf("%[1]s not found; install via: brew install %[1]s", name)
}
