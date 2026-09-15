# jotti — Kurzanleitung (Windows)

jotti per Doppelklick auf einem Windows-Rechner starten — ohne Kommandozeile.
Ein **Kassenrechner** im WLAN, die Helfer bedienen jotti auf ihren **Handys**.

## Voraussetzungen

- Ein Windows-Benutzer mit **Administratorrechten** (der Starter fragt bei jedem
  Start einmal per UAC nach).
- **Docker Desktop** ist installiert: <https://www.docker.com/products/docker-desktop/>
  — einmalig installieren, **nicht** vorab starten. Das erledigt jotti selbst.

## Einmalig vorbereiten — zuhause mit Internet

> ⚠️ **Den ersten Start unbedingt vorab zuhause mit Internet machen, nicht erst
> auf dem Fest.** Beim Erststart lädt jotti seine Programmteile herunter **und**
> holt das vertrauenswürdige Zertifikat (grünes Schloss). Beides braucht Internet.
> Zertifikat und Fallback-Adresse laufen danach ohne Internet; die grüne
> Adresse und die TSE brauchen beim Fest Internet.

1. Das ZIP **entpacken** (Rechtsklick → „Alle extrahieren"). Alle Dateien müssen
   im selben Ordner bleiben.
2. **`jotti-start.exe`** doppelklicken.
   - **SmartScreen** („Der Computer wurde durch Windows geschützt"): auf
     **„Weitere Informationen" → „Trotzdem ausführen"** klicken.
   - **UAC** („Möchten Sie zulassen, dass …"): mit **„Ja"** bestätigen.
     „Unbekannter Herausgeber" ist normal — die Programme sind nicht signiert.
3. **Warten.** Der Starter erledigt alles allein: Docker Desktop starten,
   **Firewall-Freigabe** setzen, Container herunterladen und jotti hochfahren.
   Beim ersten Mal dauert das einige Minuten. Das Fenster bleibt offen, bis
   ihr **Enter** drückt.
4. Wenn alles läuft, zeigt der Starter den Hinweis auf die **Status-Seite**
   `http://localhost:8484`. Diese im Browser am Kassenrechner öffnen — dort
   stehen die **Zugangsadresse** und ein **QR-Code** für die Helfer-Handys.

## Erster Login

Beim ersten Start legt jotti automatisch den Admin-Benutzer an und erzeugt
einen einmaligen Anmelde-Code aus 6 Ziffern. Der Code steht in der
Startkonsole (dem Fenster von `jotti-start.exe`). Ist die Konsole schon
geschlossen, jotti einfach neu starten, dann wird ein neuer Code erzeugt und
angezeigt.

Die jotti-Oberfläche öffnen (die Zugangsadresse steht auf der Status-Seite
`http://localhost:8484`) und **nicht** normal anmelden, sondern „Neues
Passwort festlegen" wählen:

- **Benutzername:** `admin`
- **Einmalpasswort:** der 6-stellige Code aus der Startkonsole
- **Neues Passwort:** ein eigenes, sicheres Passwort wählen

Nach dem Speichern ist das Einmalpasswort ungültig und der Login mit dem neuen
Passwort möglich. Dieser Schritt ist einmalig; alle weiteren Admin-Konten legt
ihr danach selbst im Admin-Bereich an.

## Helfer-Handys verbinden

- Handy ins **Vereins-WLAN** bringen (kein Mobilfunk, kein Gastnetz).
- Den **QR-Code** von der Status-Seite scannen oder die angezeigte **grüne
  Adresse** eintippen → **grünes Schloss, keine Warnung**, anmelden.
- **Falls die grüne Adresse (noch) nicht geht:** Die Status-Seite nennt dann den
  **Fallback** `https://<LAN-IP>` — beim ersten Zugriff pro Gerät einmal die
  Browserwarnung bestätigen, danach anmelden. Öffnet ein Handy die grüne Adresse
  gar nicht, blockiert vermutlich der Router (DNS-Rebind-Schutz). Die
  Router-Anleitung verlinkt die Status-Seite; sie steht auch online unter
  <https://jotti.rocks/docs/leitfaden/fehlersuche/>.

## Bondruck

Der gedruckte Kassenbeleg braucht einen Drucker (siehe
<https://jotti.rocks/docs/leitfaden/haeufige-fragen/>). Für den Bondruck
zusätzlich **`jotti-relay.exe`** doppelklicken. Es läuft ohne
Administratorrechte und nimmt seine Zugangsdaten aus der `.env`, die
`jotti-start.exe` angelegt hat (in `C:\ProgramData\jotti`).

Außerdem im Admin-Bereich unter „Bondrucker" je Station die „Drucker-IP"
eintragen — auch für den Kassenbeleg. Ohne konfigurierte Station wird nichts
gedruckt.

## Probleme

- **„Port 80 ist durch ‚X' (PID …) belegt"** (oder Port 443): Das genannte
  Programm beenden (häufig Skype, IIS oder eine VM-Software) und
  `jotti-start.exe` erneut starten.
- **Fenster schließt sich zu schnell:** Es bleibt bis zum Enter-Druck offen;
  steht oben eine Fehlermeldung, diese zuerst lesen.
- **„volume ‚jotti-local_jotti-config' … not created by Docker Compose":** Eine
  **harmlose** Warnung — jotti läuft normal weiter. Sie erscheint, wenn dieses
  Volume ohne die Compose-Kennzeichnung angelegt wurde; nachträglich lässt sie
  sich nicht setzen. Sie verschwindet, sobald das Volume einmal neu angelegt wird.

## Beenden

**`jotti-stop.cmd`** doppelklicken (oder in Docker Desktop stoppen). **Daten und
Zertifikate bleiben erhalten** und stehen beim nächsten Start wieder bereit.

## Am nächsten Festtag

Wieder dieselben zwei Doppelklicks (`jotti-start.exe`, bei Bedarf
`jotti-relay.exe`) inklusive UAC-Bestätigung. Hat der Rechner eine neue
Netzwerk-Adresse, **zeigt die Status-Seite sie erneut** — für die grüne Adresse
gilt weiterhin dasselbe Zertifikat, also **keine neue Warnung**. Auf der
Fallback-Adresse bestätigt jedes Gerät die Browserwarnung für die neue Adresse
einmal erneut.

## Daten nach dem Fest sichern (optional)

Wollt ihr die Kassendaten zusätzlich extern sichern (z. B. auf einen
**USB-Stick**), erstellt eine Sicherungsdatei. jotti muss dazu laufen. Die
**Eingabeaufforderung** (cmd) öffnen und diese drei Zeilen hineinkopieren:

```
md "%PROGRAMDATA%\jotti\backups" 2>nul
docker exec jotti-postgres-local pg_dump --clean --if-exists -U admin -d jotti > "%PROGRAMDATA%\jotti\backups\manuell-%DATE:~-4%%DATE:~-7,2%%DATE:~-10,2%.sql"
for %f in ("%PROGRAMDATA%\jotti\backups\manuell-*.sql") do @echo %~zf Bytes  %~nxf
```

- Zeile 1 legt den Ordner an, falls er fehlt; `2>nul` schluckt die Meldung, wenn
  er schon da ist.
- Zeile 2 schreibt die Sicherung. Das Datum steckt im Dateinamen, damit die
  Sicherung eines anderen Tages die erste nicht überschreibt.
- Zeile 3 listet jede vorhandene Sicherung mit ihrer Größe. Erscheint keine
  Zeile oder **0 Bytes**, ist die Sicherung fehlgeschlagen — dann lief jotti
  nicht. Löscht die leere Datei und versucht es erneut.

Die Dateien liegen im Ordner `C:\ProgramData\jotti\backups`. In denselben Ordner
spiegelt jotti auch die **automatischen Backups vor jedem Update**. Diesen Ordner
könnt ihr komplett auf einen USB-Stick oder in eine Cloud kopieren.

## jotti aktualisieren

Meldet der Starter beim Hochfahren „Neue Version verfügbar" mit einem
Download-Link, so aktualisiert ihr jotti in drei Schritten:

1. **`jotti-stop.cmd`** doppelklicken, um das laufende jotti sauber zu beenden.
2. Das **neue ZIP entpacken** — der **Ort ist egal**. Es muss **nicht** derselbe
   Ordner sein wie vorher; ein frischer Ordner ist völlig in Ordnung.
3. **`jotti-start.exe`** im neuen Ordner doppelklicken (UAC mit „Ja" bestätigen).

**Eure Daten bleiben erhalten:** Bestellungen, Benutzer, Produkte, der
Installations-Schlüssel und das grüne Zertifikat liegen geschützt außerhalb des
Programmordners (in Docker-Volumes). Egal wohin ihr entpackt — der Schlüssel folgt
den Daten, jotti findet beides beim Start wieder. Den alten Ordner erst löschen,
wenn das nächste Fest gelaufen ist: bis dahin liegt darin die `jotti-start.exe` des
vorherigen Release — der Rückweg, falls das Update Ärger macht.

> ⛔ **Niemals `docker compose down -v` ausführen.** Das `-v` löscht **alle**
> Docker-Volumes — und damit **Daten, Installations-Schlüssel und das grüne
> Zertifikat** unwiderruflich (auch ein Update bringt sie dann nicht zurück). Zum
> Beenden immer **`jotti-stop.cmd`** verwenden: das stoppt nur die Container und
> lässt alles erhalten.

**Geht beim Update etwas schief:** **`jotti-restore.cmd`** doppelklicken — es
spielt das automatische Backup von vor dem Update zurück. Den vollständigen
Ablauf beschreibt <https://jotti.rocks/docs/leitfaden/aktualisieren/>.

## Wenn nach einem Update niemand mehr hineinkommt

Sehr selten — meist nach einem Update von einer **sehr alten** Version — passt das
in der Datenbank gespeicherte Passwort nicht mehr zum aktuellen
Installations-Schlüssel. jotti startet dann gar nicht mehr; der Starter bricht mit einer Fehlermeldung ab.
**Eure Daten sind dabei nicht verloren** — nur das Schloss passt nicht zum
Schlüssel. Zwei datenerhaltende Wege zurück:

1. **`jotti-repair.cmd`** doppelklicken. Es gleicht das Datenbank-Passwort an den
   aktuellen Installations-Schlüssel an, ohne eure Daten zu verändern, und endet
   mit dem Hinweis, `jotti-start.exe` zu doppelklicken. Mehrfaches Ausführen
   schadet nicht. Danach einmal **neu anmelden**.
2. Habt ihr noch die **`.env` aus der alten Installation** (liegt ggf. im
   Programmordner neben `jotti-start.exe`): kopiert sie nach
   **`C:\ProgramData\jotti\.env`** und startet `jotti-start.exe` erneut — dann
   verwendet jotti wieder den ursprünglichen Schlüssel.

> 🔒 **Sicherheit:** Dieser Windows-Aufbau läuft nur im lokalen WLAN. Öffnet ihn
> **niemals** ins Internet — richtet im Router **keine Port-Weiterleitung** auf
> den Kassenrechner ein.
