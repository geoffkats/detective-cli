@echo off
setlocal enabledelayedexpansion

echo.
echo ========================================
echo  Detective - Forensic CLI Installer
echo ========================================
echo.

REM Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
  echo Error: Go is not installed or not on PATH
  echo Please install Go 1.21+ from https://golang.org/dl/
  pause
  exit /b 1
)

echo [1/3] Installing detective...
go install ./cmd/detective
if errorlevel 1 (
  echo Error: Failed to build detective
  pause
  exit /b 1
)

echo [2/3] Configuring system...
set GOBIN=%USERPROFILE%\go\bin

REM Check if already on PATH
echo %PATH% | find /I "%GOBIN%" >nul
if errorlevel 1 (
  echo [3/3] Adding to system PATH...
  
  REM Use PowerShell to add to user PATH
  powershell -NoProfile -Command ^
    "[System.Environment]::SetEnvironmentVariable('PATH', $([System.Environment]::GetEnvironmentVariable('PATH', 'User')) + ';' + '%GOBIN%', 'User'); Write-Host 'Added to PATH'"
  
  if errorlevel 1 (
    echo Error: Failed to update PATH
    pause
    exit /b 1
  )
  
  echo.
  echo ========================================
  echo  Installation Complete!
  echo ========================================
  echo.
  echo Detective has been installed to: %GOBIN%
  echo.
  echo IMPORTANT: Open a NEW terminal window and run:
  echo   detective -verbose
  echo.
  echo Or use it anywhere:
  echo   detective -path C:\path\to\project
  echo.
) else (
  echo.
  echo ========================================
  echo  Installation Complete!
  echo ========================================
  echo.
  echo Detective is already on your PATH!
  echo.
  echo Try it now in a new terminal:
  echo   detective -verbose
  echo.
)

pause
endlocal
