@echo off
setlocal

if "%~1"=="" (
    echo Error: No path specified
    echo Usage: setup.bat
    exit /b 1
)

echo Creating directory structure...

mkdir pkg\models 2>nul
mkdir internal\scanner 2>nul
mkdir internal\git 2>nul
mkdir internal\inference 2>nul
mkdir internal\reporter 2>nul
mkdir cmd\detective 2>nul

echo Directory structure created successfully!
echo.
echo Next steps:
echo 1. Run: copy-files.bat (to be created with all source files)
echo 2. Run: build.bat to compile
echo.

pause
