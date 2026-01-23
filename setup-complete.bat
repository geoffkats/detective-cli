@echo off
echo ========================================
echo Detective CLI - Complete Setup
echo ========================================
echo.

echo [1/7] Creating directories...
mkdir pkg\models 2>nul
mkdir internal\scanner 2>nul  
mkdir internal\git 2>nul
mkdir internal\inference 2>nul
mkdir internal\reporter 2>nul
mkdir cmd\detective 2>nul
echo Done.

echo.
echo [2/7] All directories created.
echo.
echo Next: Create source files manually or use the provided templates.
echo.
echo Directory structure:
tree /F /A
echo.
echo Run build.bat when source files are ready.
pause
