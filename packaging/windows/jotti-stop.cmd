@echo off
REM jotti sauber beenden — Doppelklick stoppt alle Container.
REM Daten und Caddy-Zertifikate bleiben in den Docker-Volumes erhalten.
cd /d "%~dp0"
docker compose -f docker-compose.release.yml down
echo.
echo jotti wurde gestoppt. Daten und Zertifikate bleiben erhalten.
pause
