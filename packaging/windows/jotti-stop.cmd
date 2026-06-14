@echo off
REM jotti sauber beenden — Doppelklick stoppt alle Container.
REM Daten und Caddy-Zertifikate bleiben in den Docker-Volumes erhalten.
setlocal
cd /d "%~dp0"
set ENVFILE=%PROGRAMDATA%\jotti\.env
set COMPOSE=docker compose -f docker-compose.release.yml
if exist "%ENVFILE%" set COMPOSE=%COMPOSE% --env-file "%ENVFILE%"
%COMPOSE% down
echo.
echo jotti wurde gestoppt. Daten und Zertifikate bleiben erhalten.
pause
