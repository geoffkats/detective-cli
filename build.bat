@echo off
echo Creating directory structure...

mkdir pkg\models 2>nul
mkdir internal\scanner 2>nul
mkdir internal\git 2>nul
mkdir internal\inference 2>nul
mkdir internal\reporter 2>nul
mkdir cmd\detective 2>nul

echo Downloading dependencies...
go mod download

echo Building detective...
go build -o detective.exe ./cmd/detective

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Build successful! detective.exe created.
    echo.
    echo Usage: detective.exe open ^<path^>
    echo Example: detective.exe open .
) else (
    echo.
    echo Build failed. Check errors above.
)

pause
