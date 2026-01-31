package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Settings represents user preferences
type Settings struct {
	DataDir          string       `json:"data_directory" validate:"required,notblank"`
	Admin            gin.Accounts `json:"admin" validate:"required,admin_accounts"`
	Port             string       `json:"port" validate:"required,notblank"`
	DanmuEnabled     bool         `json:"danmu_enabled"`
	ApprovalsEnabled bool         `json:"approvals_enabled"`
}

// App constants
const settingsFile = "settings.json"

func SettingsFilepath(dataDir string) string {
	return filepath.Join(dataDir, settingsFile)
}

// LoadSettings loads settings from JSON file
func LoadSettings(dataDir string) (*Settings, error) {

	path := SettingsFilepath(dataDir)

	// If file doesn't exist, return default settings
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s Settings
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

// SaveSettings saves settings to JSON file safely (atomic write)
func SaveSettings(dataDir string, s *Settings) error {

	path := SettingsFilepath(dataDir)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: write to temp file then rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func ValidateSettings(s *Settings) error {
	if s == nil {
		return errors.New("settings is required")
	}

	v := validator.New()
	_ = v.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
		value, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}
		return strings.TrimSpace(value) != ""
	})
	_ = v.RegisterValidation("admin_accounts", func(fl validator.FieldLevel) bool {
		accounts, ok := fl.Field().Interface().(gin.Accounts)
		if !ok || len(accounts) == 0 {
			return false
		}
		for user, password := range accounts {
			if strings.TrimSpace(user) == "" || strings.TrimSpace(password) == "" {
				return false
			}
		}
		return true
	})

	return v.Struct(s)
}
