package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
)

type Config struct {
	WatchDirs      []string `json:"watch_dirs" yaml:"watch_dirs"`
	IgnoreDirs     []string `json:"ignore_dirs" yaml:"ignore_dirs"`
	Rules          []Rule   `json:"rules" yaml:"rules"`
	DebounceTime   string   `json:"debounce_time" yaml:"debounce_time"`
	LiveReload     bool     `json:"live_reload" yaml:"live_reload"`
	LiveReloadPort int      `json:"live_reload_port" yaml:"live_reload_port"`
}

type Rule struct {
	Extensions []string `json:"extensions" yaml:"extensions"`
	Command    string   `json:"command" yaml:"command"`
}

var (
	configFile     = flag.String("config", "", "Path to config file (JSON or YAML)")
	watchDir       = flag.String("dir", ".", "Directory to watch (default: current directory)")
	ignoreDirs     = flag.String("ignore", "node_modules,.git", "Comma-separated list of directories to ignore")
	debounce       = flag.String("debounce", "500ms", "Debounce time for file changes")
	liveReload     = flag.Bool("live-reload", false, "Enable live reload for frontend workflows")
	liveReloadPort = flag.Int("live-reload-port", 35729, "Port for live reload server")
	clients        = make(map[*websocket.Conn]bool)
	clientsMutex   = &sync.Mutex{}
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	flag.Parse()

	// Initialize logger
	logger := log.New(os.Stdout, "[go-watch] ", log.LstdFlags|log.Lshortfile)

	// Default configuration
	config := Config{
		WatchDirs:      []string{*watchDir},
		IgnoreDirs:     strings.Split(*ignoreDirs, ","),
		Rules:          []Rule{},
		DebounceTime:   *debounce,
		LiveReload:     *liveReload,
		LiveReloadPort: *liveReloadPort,
	}

	// Load config from file if provided
	if *configFile != "" {
		var err error
		config, err = loadConfig(*configFile, logger)
		if err != nil {
			logger.Fatalf("Failed to load config: %v", err)
		}
	}

	// Validate debounce time
	debounceDuration, err := time.ParseDuration(config.DebounceTime)
	if err != nil {
		logger.Fatalf("Invalid debounce time: %v", err)
	}

	// Start live reload server if enabled
	if config.LiveReload {
		go startLiveReloadServer(config.LiveReloadPort, logger)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	// Start watching directories
	for _, dir := range config.WatchDirs {
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error walking directory %s: %v", path, err)
			}
			if info.IsDir() && !isIgnored(path, config.IgnoreDirs) {
				if err := watcher.Add(path); err != nil {
					return fmt.Errorf("failed to add %s to watcher: %v", path, err)
				}
			}
			return nil
		})
		if err != nil {
			logger.Fatalf("Failed to watch directory: %v", err)
		}
	}

	logger.Println("Watching for changes...")

	// Run commands on start
	for _, rule := range config.Rules {
		runCommand(rule.Command, logger)
	}

	var (
		lastRestart time.Time
	)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if isWatchedFile(event.Name, config.Rules) && time.Since(lastRestart) > debounceDuration {
				logger.Printf("Change detected: %s", event.Name)
				lastRestart = time.Now()

				// Run commands for matching rules
				for _, rule := range config.Rules {
					if isWatchedFile(event.Name, []Rule{rule}) {
						runCommand(rule.Command, logger)
					}
				}

				// Trigger live reload for frontend files
				if config.LiveReload && isFrontendFile(event.Name) {
					notifyLiveReload(logger)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Printf("Watcher error: %v", err)
		}
	}
}

// runCommand executes a command and logs its output
func runCommand(command string, logger *log.Logger) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logger.Printf("Command failed: %v", err)
	}
}

// isIgnored checks if a directory should be ignored
func isIgnored(path string, ignoreDirs []string) bool {
	for _, dir := range ignoreDirs {
		if strings.Contains(path, dir) {
			return true
		}
	}
	return false
}

// isWatchedFile checks if a file matches any rule
func isWatchedFile(path string, rules []Rule) bool {
	ext := filepath.Ext(path)
	for _, rule := range rules {
		for _, e := range rule.Extensions {
			if e == "*" || "."+e == ext {
				return true
			}
		}
	}
	return false
}

// isFrontendFile checks if a file is a frontend file (HTML, CSS, JS)
func isFrontendFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".html" || ext == ".css" || ext == ".js"
}

// loadConfig loads configuration from a JSON or YAML file
func loadConfig(path string, logger *log.Logger) (Config, error) {
	var config Config
	file, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %v", err)
	}

	switch {
	case strings.HasSuffix(path, ".json"):
		err = json.Unmarshal(file, &config)
	case strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml"):
		err = yaml.Unmarshal(file, &config)
	default:
		return config, fmt.Errorf("unsupported config file format: %s", path)
	}

	if err != nil {
		return config, fmt.Errorf("failed to parse config file: %v", err)
	}

	logger.Printf("Config file location: %s", path)
	return config, nil
}

// startLiveReloadServer starts a WebSocket server for live reload
func startLiveReloadServer(port int, logger *log.Logger) {
	http.HandleFunc("/livereload", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Printf("Live reload WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		clientsMutex.Lock()
		clients[conn] = true
		clientsMutex.Unlock()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				clientsMutex.Lock()
				delete(clients, conn)
				clientsMutex.Unlock()
				break
			}
		}
	})

	logger.Printf("Live reload server started on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", port), nil); err != nil {
		logger.Fatalf("Failed to start live reload server: %v", err)
	}
}

// notifyLiveReload sends a reload message to all connected clients
func notifyLiveReload(logger *log.Logger) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, []byte("reload")); err != nil {
			logger.Printf("Failed to send live reload message: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}
