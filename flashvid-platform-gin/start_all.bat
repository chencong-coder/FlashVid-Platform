@echo off
echo Starting FlashVid Platform...

echo.
echo [1/2] Starting API Server...
start "FlashVid API" cmd /k "go run cmd/server/api/main.go"

timeout /t 2 /nobreak >nul

echo [2/2] Starting Worker (Consumers)...
start "FlashVid Worker" cmd /k "go run cmd/server/worker/main.go"

echo.
echo All services started!
echo - API Server: http://localhost:8080
echo - Worker: Running in background
echo.
pause
