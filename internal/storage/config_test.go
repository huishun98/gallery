package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSettingsFilepath(t *testing.T) {
	assert.Equal(t, filepath.Join("data", "settings.json"), SettingsFilepath("data"))
}

func TestLoadSettingsMissingFileReturnsNil(t *testing.T) {
	tmp := t.TempDir()

	got, err := LoadSettings(tmp)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestLoadSettingsInvalidJSON(t *testing.T) {
	tmp := t.TempDir()

	path := SettingsFilepath(tmp)
	assert.NoError(t, os.WriteFile(path, []byte("{not-json"), 0644))

	got, err := LoadSettings(tmp)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestSaveSettingsAndLoadSettings(t *testing.T) {
	tmp := t.TempDir()

	in := &Settings{
		DataDir:      "data-dir",
		Admin:        gin.Accounts{"admin": "secret"},
		Port:         "8080",
		DanmuEnabled: true,
	}

	err := SaveSettings(tmp, in)
	assert.NoError(t, err)

	_, statErr := os.Stat(SettingsFilepath(tmp))
	assert.NoError(t, statErr)

	out, err := LoadSettings(tmp)
	assert.NoError(t, err)
	assert.Equal(t, in, out)
}

func TestLoadSettingsDefaultsDanmuEnabled(t *testing.T) {
	tmp := t.TempDir()

	raw := []byte(`{"data_directory":"data-dir","admin":{"admin":"secret"},"port":"8080"}`)
	assert.NoError(t, os.WriteFile(SettingsFilepath(tmp), raw, 0644))

	out, err := LoadSettings(tmp)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.False(t, out.DanmuEnabled)
}
