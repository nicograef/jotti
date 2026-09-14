@echo off
REM jotti reparieren - das Datenbank-Passwort an den Install-Schluessel angleichen.
REM
REM Fuer den Fall "die Daten sind da, aber jotti kommt nicht mehr hinein": Weicht das
REM in der Datenbank gespeicherte Passwort vom aktuellen Install-Schluessel ab
REM (migrate oder backend melden Authentifizierungsfehler), gleicht dieses Skript es
REM ueber den lokalen Trust-Zugang im postgres-Container an - ohne das alte Passwort
REM zu kennen. Es veraendert KEINE Daten (nur das Rollen-Passwort) und ist idempotent.
REM
REM Das Skript startet jotti nicht selbst: nur jotti-start.exe uebergibt dem
REM Reverse-Proxy die LAN-Adresse des Rechners.
setlocal
cd /d "%~dp0"
set ENVFILE=%PROGRAMDATA%\jotti\.env
set COMPOSE=docker compose -f docker-compose.release.yml --env-file "%ENVFILE%"

if not exist "%ENVFILE%" goto :noenv

REM Passwort aus der .env lesen: eol=# ueberspringt die Kommentarzeile, der Wert ist
REM reines Hex (keine Sonderzeichen). Der Rollenname ist fest "admin" (wie
REM core.PostgresUser / .env.example) und wird nicht aus der .env gelesen.
set DBPASS=
for /f "usebackq eol=# tokens=1,* delims==" %%a in ("%ENVFILE%") do (
  if /i "%%a"=="POSTGRES_PASSWORD" set DBPASS=%%b
)
if not defined DBPASS goto :noenv

echo Diese Reparatur gleicht das Datenbank-Passwort an den aktuellen
echo Installations-Schluessel an. Eure Daten bleiben unveraendert erhalten.
echo.
set /p ANSWER=Fortfahren? (j/N):
if /i not "%ANSWER%"=="j" goto :cancel

echo.
echo Starte die Datenbank ...
%COMPOSE% up -d --wait postgres
if errorlevel 1 goto :error

echo Gleiche das Datenbank-Passwort an den Installations-Schluessel an ...
docker exec jotti-postgres-local psql -U admin -d jotti -v ON_ERROR_STOP=1 -c "ALTER USER admin PASSWORD '%DBPASS%'"
if errorlevel 1 goto :error

echo.
echo Reparatur abgeschlossen. jotti wurde nicht neu gestartet.
echo Jetzt jotti-start.exe doppelklicken.
echo Hinweis: Bitte einmal neu anmelden - bereits ausgestellte Anmeldungen koennen
echo durch einen zwischenzeitlich erneuerten Schluessel ungueltig geworden sein.
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
echo FEHLER bei der Reparatur. Bitte die Ausgabe oben pruefen.

:end
pause
