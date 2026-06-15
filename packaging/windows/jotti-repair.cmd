@echo off
REM jotti reparieren — das Datenbank-Passwort an den Install-Schluessel angleichen.
REM
REM Fuer den Fall "die Daten sind da, aber jotti kommt nicht mehr hinein": Nach
REM einem Upgrade von einer sehr alten Version kann das in der Datenbank
REM gespeicherte Passwort vom aktuellen Install-Schluessel abweichen (migrate oder
REM backend melden dann Authentifizierungsfehler). Dieses Skript gleicht das
REM Datenbank-Passwort datenerhaltend an den aktuellen Install-Schluessel an — ueber
REM den lokalen Trust-Zugang im postgres-Container, ohne das alte Passwort zu
REM kennen. Es veraendert KEINE Daten (nur das Rollen-Passwort) und fasst keine
REM anderen Volumes an. Mehrfaches Ausfuehren ist gefahrlos (idempotent).
setlocal
cd /d "%~dp0"
set ENVFILE=%PROGRAMDATA%\jotti\.env
set COMPOSE=docker compose -f docker-compose.release.yml --env-file "%ENVFILE%"

if not exist "%ENVFILE%" goto :noenv

REM Das Passwort des Install-Schluessels aus der .env lesen. Die Kommentarzeile
REM beginnt mit '#' und wird per eol uebersprungen; der Wert ist reines Hex (keine
REM Sonderzeichen), daher ist das Durchreichen unproblematisch. Der Rollenname ist
REM fest "admin" (wie core.PostgresUser / .env.example) und wird darum nicht aus
REM der .env gelesen.
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

echo Starte jotti neu ...
%COMPOSE% up -d
if errorlevel 1 goto :error

echo.
echo Reparatur abgeschlossen. jotti laeuft wieder.
echo Hinweis: Bitte einmal neu anmelden — bereits ausgestellte Anmeldungen koennen
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
