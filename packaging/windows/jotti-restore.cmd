@echo off
REM jotti aus dem letzten automatischen Backup wiederherstellen.
REM
REM Vor jedem Update sichert jotti-start.exe die Datenbank automatisch in das
REM jotti-backups-Volume. Dieses Skript spielt das NEUESTE dieser Backups zurueck
REM — z. B. wenn ein Update fehlgeschlagen ist. Daten, die seit dem Backup
REM erfasst wurden, gehen dabei verloren.
setlocal
cd /d "%~dp0"
set ENVFILE=%PROGRAMDATA%\jotti\.env
set COMPOSE=docker compose -f docker-compose.release.yml --env-file "%ENVFILE%"

echo Diese Wiederherstellung ersetzt die aktuellen Daten durch das letzte
echo automatische Backup (vor dem letzten Update erstellt).
echo Seither erfasste Daten gehen dabei verloren.
echo.
set /p ANSWER=Fortfahren? (j/N):
if /i not "%ANSWER%"=="j" goto :cancel

echo.
echo Starte die Datenbank ...
%COMPOSE% up -d --wait postgres
if errorlevel 1 goto :error

echo Stoppe die Anwendung waehrend der Wiederherstellung ...
%COMPOSE% stop backend frontend reverse-proxy

echo Spiele das letzte Backup ein ...
docker exec jotti-postgres-local sh -c "set -e; F=$(ls -1 /jotti-backups/jotti-*.sql 2>/dev/null | tail -n 1); if [ -z \"$F\" ]; then echo 'Kein Backup gefunden.'; exit 1; fi; echo \"Verwende $F\"; psql -U admin -d jotti -f \"$F\""
if errorlevel 1 goto :error

echo Starte jotti neu ...
%COMPOSE% up -d
if errorlevel 1 goto :error

echo.
echo Wiederherstellung abgeschlossen. jotti laeuft wieder.
goto :end

:cancel
echo Abgebrochen. Es wurde nichts veraendert.
goto :end

:error
echo.
echo FEHLER bei der Wiederherstellung. Bitte die Ausgabe oben pruefen.

:end
pause
