package binary

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhereIsUsesPATH(t *testing.T) {
	tmp := t.TempDir()
	fake := filepath.Join(tmp, "fakebin")
	assert.NoError(t, os.WriteFile(fake, []byte("#!/bin/sh\necho ok\n"), 0755))

	pathEnv := tmp + string(os.PathListSeparator) + os.Getenv("PATH")
	t.Setenv("PATH", pathEnv)

	p, dir, err := WhereIs("fakebin")
	assert.NoError(t, err)
	assert.Equal(t, fake, p)
	assert.Equal(t, "", dir)
}

func TestWhereIsNotFound(t *testing.T) {
	_, _, err := WhereIs("codex-not-real-binary")
	assert.Error(t, err)
}
