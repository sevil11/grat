@echo off
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
python "grat_agent_20250823225918.py"
pause
