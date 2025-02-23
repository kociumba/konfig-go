package main

import (
	"log"
	"time"

	"github.com/k0kubun/pp"
	"github.com/kociumba/konfig"
)

// WebServer represents configuration for a web server component
type WebServer struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	RequestTimeout  time.Duration `json:"request_timeout"`
	MaxConnections  int           `json:"max_connections"`
	EnableHTTPS     bool          `json:"enable_https"`
	CertificatePath string        `json:"certificate_path"`
}

// DatabaseCache represents configuration for an independent caching service
type DatabaseCache struct {
	MaxSize       int           `json:"max_size"`
	TTL           time.Duration `json:"ttl"`
	EnabledTables []string      `json:"enabled_tables"`
	Strategy      string        `json:"strategy"`
	Stats         *CacheStats   `json:"stats"`
}

// CacheStats tracks cache usage statistics
type CacheStats struct {
	LastReset   time.Time `json:"last_reset"`
	TotalHits   int64     `json:"total_hits"`
	TotalMisses int64     `json:"total_misses"`
}

var (
	// Independent components with their own configurations
	webServer = WebServer{
		Host:            "localhost",
		Port:            8080,
		RequestTimeout:  time.Second * 30,
		MaxConnections:  1000,
		EnableHTTPS:     false,
		CertificatePath: "/etc/certs/server.pem",
	}

	dbCache = DatabaseCache{
		MaxSize:       1024 * 1024 * 100, // 100MB
		TTL:           time.Hour * 24,
		EnabledTables: []string{"users", "products"},
		Strategy:      "lru",
		Stats: &CacheStats{
			LastReset:   time.Now(),
			TotalHits:   0,
			TotalMisses: 0,
		},
	}
)

func main() {
	// Initialize configuration manager
	mngr, err := konfig.NewKonfigManager(konfig.KonfigOptions{
		Format:       konfig.JSON,
		AutoLoad:     true,
		AutoSave:     true,
		UseCallbacks: true,
		KonfigPath:   "config.json",
	})
	if err != nil {
		log.Fatalf("Failed to create config manager: %v", err)
	}

	// Register web server configuration
	webSection := konfig.NewKonfigSection(&webServer,
		konfig.WithSectionName(func() string { return "web_server" }),
		konfig.WithOnLoad(func() error {
			log.Println("Web server configuration loaded")
			// Simulate reconfiguring the web server
			return nil
		}),
	)

	// Register cache configuration
	cacheSection := konfig.NewKonfigSection(&dbCache,
		konfig.WithSectionName(func() string { return "db_cache" }),
		konfig.WithOnLoad(func() error {
			log.Println("Cache configuration loaded")
			// Simulate reinitializing the cache
			return nil
		}),
	)

	// Register both independent sections
	mngr.RegisterSection(webSection)
	mngr.RegisterSection(cacheSection)

	// Load existing configuration if available
	if err := mngr.Load(); err != nil {
		log.Printf("No existing configuration found, using defaults: %v", err)
	}

	// Simulate some runtime changes
	webServer.MaxConnections = 2000
	dbCache.TTL = time.Hour * 48
	dbCache.Stats.TotalHits++

	// Print current configuration state
	pp.Println("Current Web Server Configuration:", webServer)
	pp.Println("Current Cache Configuration:", dbCache)

	// Ensure configuration is saved when the program exits
	defer func() {
		if err := mngr.Save(); err != nil {
			log.Printf("Error saving configuration: %v", err)
		}
	}()
}
