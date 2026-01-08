package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"mimirprompt/internal/routes"

	_ "github.com/go-sql-driver/mysql"
)

// Config represents the application configuration
type Config struct {
	Database struct {
		DSN string `json:"dsn"`
	} `json:"database"`
	Server struct {
		Port string `json:"port"`
	} `json:"server"`
	Data struct {
		PromptsJSON     string `json:"prompts_json"`
		ImagesDir       string `json:"images_dir"`
		ImagesURLPrefix string `json:"images_url_prefix"`
	} `json:"data"`
}

func main() {
	fmt.Println("🚀 MimirPrompt API Server")
	fmt.Println("=========================")

	// Load config
	configPath := "config/config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfgRaw, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("❌ Cannot read config: %v\n", err)
		os.Exit(1)
	}

	var cfg Config
	if err := json.Unmarshal(cfgRaw, &cfg); err != nil {
		fmt.Printf("❌ Cannot parse config: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	fmt.Println("🔌 Connecting to database...")
	db, err := sql.Open("mysql", cfg.Database.DSN)
	if err != nil {
		fmt.Printf("❌ Cannot connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("❌ Cannot ping database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Connected to database")

	// Setup router
	router := routes.SetupRouter(db, cfg.Data.ImagesDir)

	// Start server
	fmt.Printf("🌐 Starting server on port %s...\n", cfg.Server.Port)
	fmt.Printf("📍 API available at http://localhost:%s/api\n", cfg.Server.Port)

	if err := router.Run(":" + cfg.Server.Port); err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
		os.Exit(1)
	}
}
