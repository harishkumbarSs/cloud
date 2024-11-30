@echo off
echo Creating required directories...
mkdir uploads 2>nul
mkdir web\static 2>nul
mkdir web\templates 2>nul
mkdir notes 2>nul

echo Starting server...
go run cmd/server/main.go
