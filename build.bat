@echo off
REM xoji build script for Windows
REM Builds cross-platform binaries to dist/

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set SRC_DIR=%SCRIPT_DIR%src
set DIST_DIR=%SCRIPT_DIR%dist
set VERSION=1.0.0

echo Building xoji v%VERSION%...
if not exist "%DIST_DIR%" mkdir "%DIST_DIR%"

cd /d "%SRC_DIR%"

REM Windows x86-64
echo   - Windows x86-64...
set GOOS=windows
set GOARCH=amd64
go build -o "%DIST_DIR%\xoji-windows.exe" .
if errorlevel 1 goto error

REM macOS ARM64
echo   - macOS ARM64...
set GOOS=darwin
set GOARCH=arm64
go build -o "%DIST_DIR%\xoji-mac-arm64" .
if errorlevel 1 goto error

REM macOS x86-64
echo   - macOS x86-64...
set GOOS=darwin
set GOARCH=amd64
go build -o "%DIST_DIR%\xoji-mac-amd64" .
if errorlevel 1 goto error

REM Linux x86-64
echo   - Linux x86-64...
set GOOS=linux
set GOARCH=amd64
go build -o "%DIST_DIR%\xoji-linux" .
if errorlevel 1 goto error

REM Native build for Windows
echo   - Native build (Windows)...
set GOOS=
set GOARCH=
go build -o "%SCRIPT_DIR%xoji.exe" .
if errorlevel 1 goto error

echo.
echo Build complete!
echo.
echo Binaries in dist\:
dir "%DIST_DIR%"
echo.
echo Quick start:
echo   xoji.exe setup ..\my_xojo_project
echo   xoji.exe index ..\my_xojo_project
goto end

:error
echo Build failed!
exit /b 1

:end
endlocal
