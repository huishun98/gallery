package main

import (
	"bufio"
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"gallery/internal/handlers"
	"gallery/internal/logger"
	"gallery/internal/router"
	"gallery/internal/storage"
	"gallery/internal/tunnel"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"net"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

//go:embed templates/*
var templatesFS embed.FS

func waitForPort(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return errors.New("timeout waiting for port")
}

func promptUserInputs(defaultDataDir string) (*storage.Settings, error) {
	reader := bufio.NewReader(os.Stdin)

	username := "admin"
	password := "admin"

	config := &storage.Settings{
		DataDir: defaultDataDir,
		Port:    "8000",
		Admin: gin.Accounts{
			username: password,
		},
		ApprovalsEnabled: false,
		DanmuEnabled:     true,
		UploadsEnabled:   true,
	}

	// 0. Ask if should use defaults
	fmt.Print("Use default settings? [Y/n]: ")
	useDefaults, _ := reader.ReadString('\n')
	useDefaults = strings.TrimSpace(strings.ToLower(useDefaults))

	if useDefaults != "n" {
		return config, nil
	}

	// 1. Ask for port number
	fmt.Printf("Port [%s]: ", config.Port)
	portInput, _ := reader.ReadString('\n')
	portInput = strings.TrimSpace(portInput)
	if portInput != "" {
		config.Port = portInput
	}

	// 2. Ask for data directory
	fmt.Printf("Storage location [%s]: ", defaultDataDir)
	dataDirInput, _ := reader.ReadString('\n')
	dataDirInput = strings.TrimSpace(dataDirInput)
	if dataDirInput != "" {
		config.DataDir = dataDirInput
	}
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return config, err
	}

	// 3. Ask if guest uploads should be enabled
	fmt.Print("Allow guest uploads? (Y/n): ")
	enableUploadsRaw, _ := reader.ReadString('\n')
	enableUploads := strings.TrimSpace(strings.ToLower(enableUploadsRaw))
	if enableUploads == "n" {
		config.UploadsEnabled = false
		config.DanmuEnabled = true
		config.ApprovalsEnabled = false
	} else {
		// 4. Ask if danmu comments should be enabled
		fmt.Print("Enable danmu comments? (Y/n): ")
		enableDanmuRaw, _ := reader.ReadString('\n')
		enableDanmu := strings.TrimSpace(strings.ToLower(enableDanmuRaw))
		if enableDanmu == "n" {
			config.DanmuEnabled = false
		}

		// 5. Ask if admin site is needed
		fmt.Print("Require admin approval for uploads? (y/N): ")
		enableApprovalsRaw, _ := reader.ReadString('\n')
		enableApprovals := strings.TrimSpace(strings.ToLower(enableApprovalsRaw))
		if enableApprovals == "y" {
			config.ApprovalsEnabled = true
		}
	}

	// 6. Username
	fmt.Print("Set admin username [admin]: ")
	usernameInput, _ := reader.ReadString('\n')
	usernameInput = strings.TrimSpace(usernameInput)
	if usernameInput != "" {
		username = usernameInput
	}

	// 7. Password (hidden input)
	fmt.Print("Set admin password [admin]: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // newline after hidden input
	if err != nil {
		return config, err
	}
	passwordInput := strings.TrimSpace(string(passwordBytes))
	if passwordInput != "" {
		password = passwordInput
	}

	config.Admin = gin.Accounts{
		username: password,
	}

	return config, nil
}

func createMediaDirs(dataDir string, uploadsEnabled, approvalsEnabled bool) error {
	mediaDir := filepath.Join(dataDir, "media")
	approvedDir := filepath.Join(mediaDir, "media")
	if err := os.MkdirAll(approvedDir, 0755); err != nil {
		return fmt.Errorf("failed to create media directory: %w", err)
	}
	if uploadsEnabled && approvalsEnabled {
		if err := os.MkdirAll(filepath.Join(mediaDir, "pending"), 0755); err != nil {
			return fmt.Errorf("failed to create pending directory: %w", err)
		}
		if err := os.MkdirAll(filepath.Join(mediaDir, "rejected"), 0755); err != nil {
			return fmt.Errorf("failed to create rejected directory: %w", err)
		}
	}
	return nil
}

func run(log *logrus.Logger, debug bool) error {
	// Force Gin into release mode before engine init
	gin.SetMode(gin.ReleaseMode)

	myFigure := figure.NewFigure("GALLERY", "", true)
	myFigure.Print()
	fmt.Println()

	fmt.Println("Gallery — your local event photo sharing server")
	fmt.Println("Press Enter to continue or Ctrl+C to abort.")
	fmt.Println()

	defaultDataDir, err := storage.DataDir("Gallery")
	if err != nil {
		return fmt.Errorf("cannot find data directory: %w", err)
	}

	// Load settings
	config, err := storage.LoadSettings(defaultDataDir)
	if err != nil {
		return fmt.Errorf("failed to get settings file: %w", err)
	}

	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Prompt for inputs
	if storage.ValidateSettings(config) != nil {
		config, err = promptUserInputs(defaultDataDir)
		if err != nil {
			return err
		}
		fmt.Println()
		if err := storage.SaveSettings(defaultDataDir, config); err != nil {
			return fmt.Errorf("failed to save setting: %w", err)
		}
	}
	config.TunnelDisabled = debug

	// Initialize database
	dbPath := filepath.Join(config.DataDir, "app.db")
	db, err := storage.InitDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open db: %v, please check data directory path in configuration file: %s", err, storage.SettingsFilepath(defaultDataDir))
	}
	defer db.Close()

	// Create necessary directories
	if err := createMediaDirs(config.DataDir, config.UploadsEnabled, config.ApprovalsEnabled); err != nil {
		return err
	}

	r := gin.Default()
	r.SetHTMLTemplate(template.Must(template.ParseFS(templatesFS, "templates/*")))

	router.SetupRoutes(r, db, config.DataDir, config.Admin, config.ApprovalsEnabled, config.DanmuEnabled, config.UploadsEnabled)

	// Start Gin server
	go func() {
		if err := r.Run(":" + config.Port); err != nil {
			log.Errorf("failed to run server: %v", err)
		}
	}()

	fmt.Println("Please wait while we prepare your server...")

	// Wait for port 8000 to be ready
	if err := waitForPort("127.0.0.1:"+config.Port, 5*time.Second); err != nil {
		return fmt.Errorf("server not ready: %w", err)
	}

	// Start Cloudflare tunnel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	publicURL := "http://localhost:" + config.Port
	if !config.TunnelDisabled {
		tunnelClient, err := tunnel.StartTunnel(ctx, publicURL)
		if err != nil {
			return err
		}
		defer tunnelClient.Close()
		publicURL = tunnelClient.PublicURL
	}

	fmt.Println()
	fmt.Printf("● Your Gallery is ready at: %s/slideshow", publicURL)
	fmt.Println()
	if config.ApprovalsEnabled {
		fmt.Printf("● Approve or reject uploaded photos at: %s/admin/review", publicURL)
		fmt.Println()
	}
	if config.UploadsEnabled {
		fmt.Printf("● Your photos and videos are stored at: %s", config.DataDir)
	} else {
		fmt.Printf("● Uploads are disabled. Add photos/videos to: %s", filepath.Join(config.DataDir, "media", "media"))
	}
	fmt.Println()
	fmt.Printf("● Update your application settings at: %s", storage.SettingsFilepath(defaultDataDir))
	fmt.Println()
	fmt.Println("● Need help? Check out our FAQ: https://github.com/huishun98/gallery/wiki/FAQ")
	fmt.Println()
	fmt.Println()

	handlers.NgrokURL.Store(publicURL)

	// Block forever
	select {}
}

func main() {
	debug := flag.Bool("debug", false, "Disable the Cloudflare tunnel and run in debug mode")
	flag.Parse()

	log := logger.New()
	if err := run(log, *debug); err != nil {
		log.Error(err)
		if term.IsTerminal(int(os.Stdin.Fd())) {
			fmt.Println("Press Enter to exit...")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		}
	}
}
