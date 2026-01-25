package main

import (
	"bufio"
	"context"
	"embed"
	"errors"
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
	config := &storage.Settings{
		DataDir:      defaultDataDir,
		Port:         "8000",
		Admin:        nil,
		DanmuEnabled: true,
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

	// 3. Ask if danmu comments should be enabled
	fmt.Print("Enable danmu comments? (Y/n): ")
	enableDanmuRaw, _ := reader.ReadString('\n')
	enableDanmu := strings.TrimSpace(strings.ToLower(enableDanmuRaw))
	if enableDanmu == "n" {
		config.DanmuEnabled = false
	}

	// 4. Ask if admin site is needed
	fmt.Print("Require admin approval for uploads? (y/N): ")
	enableAdminRaw, _ := reader.ReadString('\n')
	enableAdmin := strings.TrimSpace(strings.ToLower(enableAdminRaw))

	if enableAdmin != "y" {
		return config, nil
	}

	// 4. Username
	fmt.Print("Set admin username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	// 5. Password (hidden input)
	fmt.Print("Set admin password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // newline after hidden input
	if err != nil {
		return config, err
	}
	password := strings.TrimSpace(string(passwordBytes))
	config.Admin = gin.Accounts{
		username: password,
	}

	return config, nil
}

func main() {
	// Force Gin into release mode before engine init
	gin.SetMode(gin.ReleaseMode)

	myFigure := figure.NewFigure("GALLERY", "", true)
	myFigure.Print()
	fmt.Println()

	fmt.Println("Gallery — your local event photo sharing server")
	fmt.Println("Press Enter to continue or Ctrl+C to abort.")
	fmt.Println()

	log := logger.New()

	defaultDataDir, err := storage.DataDir("Gallery")
	if err != nil {
		log.Fatal("cannot find data directory:", err)
	}

	config, err := storage.LoadSettings(defaultDataDir)
	if err != nil {
		log.Error("failed to get settings file:", err)
	}

	// Prompt for inputs
	if config == nil {
		config, err = promptUserInputs(defaultDataDir)
		if err != nil {
			log.Fatal(err)
		}
		if err := storage.SaveSettings(defaultDataDir, config); err != nil {
			log.Fatal("failed to save setting: ", err)
		}
	}

	dbPath := filepath.Join(config.DataDir, "app.db")
	db, err := storage.InitDB(dbPath)
	if err != nil {
		log.Fatalf("failed to open db: %v, please check data directory path in configuration file: %s", err, storage.SettingsFilepath(defaultDataDir))
	}
	defer db.Close()

	r := gin.Default()
	r.SetHTMLTemplate(template.Must(template.ParseFS(templatesFS, "templates/*")))

	router.SetupRoutes(r, db, config.DataDir, config.Admin, config.DanmuEnabled)

	// Start Gin server
	go func() {
		if err := r.Run(":" + config.Port); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println("Please wait while we prepare your server...")

	// Wait for port 8000 to be ready
	if err := waitForPort("127.0.0.1:"+config.Port, 5*time.Second); err != nil {
		log.Fatal("server not ready:", err)
	}

	// Start Cloudflare tunnel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tunnel, err := tunnel.StartTunnel(ctx, "http://localhost:"+config.Port)
	if err != nil {
		log.Fatal(err)
	}
	defer tunnel.Close()

	fmt.Println()
	fmt.Printf("● Your Gallery is ready at: %s/slideshow", tunnel.PublicURL)
	fmt.Println()
	if config.Admin != nil {
		fmt.Printf("● Approve or reject uploaded photos at: %s/admin", tunnel.PublicURL)
		fmt.Println()
	}
	fmt.Printf("● Your photos and videos are stored at: %s", config.DataDir)
	fmt.Println()
	fmt.Printf("● Your application settings are stored at: %s", storage.SettingsFilepath(defaultDataDir))
	fmt.Println()
	fmt.Println("● Need help? Check out our FAQ: https://github.com/huishun98/gallery/wiki")
	fmt.Println()
	fmt.Println()

	handlers.NgrokURL.Store(tunnel.PublicURL)

	// Block forever
	select {}
}
