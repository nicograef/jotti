@echo off
REM jotti aus dem letzten automatischen Backup wiederherstellen.
REM
REM Vor jedem Update sichert jotti-start.exe die Datenbank automatisch in das
REM jotti-backups-Volume. Dieses Skript spielt das NEUESTE dieser Backups zurueck
REM - z. B. wenn ein Update fehlgeschlagen ist. Daten, die seit dem Backup
REM erfasst wurden, gehen dabei verloren.
REM
REM Das Skript startet jotti nicht selbst: nur jotti-start.exe uebergibt dem
REM Reverse-Proxy die LAN-Adresse des Rechners. Zur zurueckgespielten Datenbank
REM passt das vorherige Release.
setlocal
cd /d "%~dp0"
set ENVFILE=%PROGRAMDATA%\jotti\.env
set COMPOSE=docker compose -f docker-compose.release.yml --env-file "%ENVFILE%"

if not exist "%ENVFILE%" goto :noenv

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
if errorlevel 1 goto :error

echo Spiele das letzte Backup ein ...
docker exec jotti-postgres-local sh -c "set -e; F=$(ls -1 /jotti-backups/jotti-*.sql 2>/dev/null | tail -n 1); if [ -z \"$F\" ]; then echo 'Kein Backup gefunden.'; exit 1; fi; echo \"Verwende $F\"; psql -U admin -d jotti -v ON_ERROR_STOP=1 -f \"$F\""
if errorlevel 1 goto :error

echo.
echo Wiederherstellung abgeschlossen. Die Daten stehen wieder auf dem Stand von
echo vor dem Update. jotti laeuft noch nicht.
echo.
echo Jetzt jotti-start.exe doppelklicken - und zwar aus dem vorherigen
echo Release-ZIP: diese Version passt zur zurueckgespielten Datenbank.
goto :end

:noenv
echo.
echo FEHLER: Es wurden keine Zugangsdaten gefunden ("%ENVFILE%").
echo Bitte zuerst jotti-start.exe ausfuehren, oder die .env aus der vorherigen
echo Installation nach "%PROGRAMDATA%\jotti\.env" kopieren und erneut versuchen.
goto :end

:cancel
echo Abgebrochen. Es wurde nichts veraendert.
goto :end

:error
echo.
echo FEHLER bei der Wiederherstellung. Bitte die Ausgabe oben pruefen.

:end
pause
