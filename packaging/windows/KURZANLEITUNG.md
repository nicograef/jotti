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
> Danach läuft jotti auch ohne Internet.

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

## Helfer-Handys verbinden

- Handy ins **Vereins-WLAN** bringen (kein Mobilfunk, kein Gastnetz).
- Den **QR-Code** von der Status-Seite scannen oder die angezeigte **grüne
  Adresse** eintippen → **grünes Schloss, keine Warnung**, anmelden.
- **Falls die grüne Adresse (noch) nicht geht:** Die Status-Seite nennt dann den
  **Fallback** `https://<LAN-IP>` — beim ersten Zugriff pro Gerät einmal die
  Browserwarnung bestätigen, danach anmelden. Öffnet ein Handy die grüne Adresse
  gar nicht, blockiert vermutlich der Router (DNS-Rebind-Schutz). Die
  Router-Anleitung verlinkt die Status-Seite; sie steht auch online unter
  <https://github.com/nicograef/jotti/blob/main/docs/leitfaden.md#fehlersuche>.

## Bondruck (optional)

Für den Bondruck zusätzlich **`jotti-relay.exe`** doppelklicken. Es läuft ohne
Administratorrechte und nimmt seine Zugangsdaten aus der `.env`, die
`jotti-start.exe` angelegt hat (in `%PROGRAMDATA%\jotti`).

## Probleme

- **„Port 80/443 ist durch ‚X' (PID …) belegt":** Das genannte Programm beenden
  (häufig Skype, IIS oder eine VM-Software) und `jotti-start.exe` erneut starten.
- **Fenster schließt sich zu schnell:** Es bleibt bis zum Enter-Druck offen;
  steht oben eine Fehlermeldung, diese zuerst lesen.
- **„volume ‚jotti-local_jotti-config' … not created by Docker Compose":** Eine
  **harmlose** Warnung, die nur bei Installationen erscheint, die vor diesem Update
  angelegt wurden — jotti läuft normal weiter. Sie verschwindet, sobald dieses
  Volume einmal neu angelegt wird; neue Installationen zeigen sie gar nicht erst.

## Beenden

**`jotti-stop.cmd`** doppelklicken (oder in Docker Desktop stoppen). **Daten und
Zertifikate bleiben erhalten** und stehen beim nächsten Start wieder bereit.

## Am nächsten Festtag

Wieder dieselben zwei Doppelklicks (`jotti-start.exe`, bei Bedarf
`jotti-relay.exe`) inklusive UAC-Bestätigung. Hat der Rechner eine neue
Netzwerk-Adresse, **zeigt die Status-Seite sie erneut** — es gilt dasselbe
Zertifikat, also **keine neue Warnung**.

## jotti aktualisieren

> ⚠️ **Updates zuhause mit Internet machen, nicht auf dem Fest.** Wie beim
> Erststart lädt jotti dabei neue Programmteile herunter.

Meldet der Starter beim Hochfahren „Neue Version verfügbar" mit einem
Download-Link, so aktualisiert ihr jotti in drei Schritten:

1. **`jotti-stop.cmd`** doppelklicken, um das laufende jotti sauber zu beenden.
2. Das **neue ZIP entpacken** — der **Ort ist egal**. Es muss **nicht** derselbe
   Ordner sein wie vorher; ein frischer Ordner ist völlig in Ordnung.
3. **`jotti-start.exe`** im neuen Ordner doppelklicken (UAC mit „Ja" bestätigen).

**Eure Daten bleiben erhalten:** Bestellungen, Benutzer, Produkte, der
Installations-Schlüssel und das grüne Zertifikat liegen geschützt außerhalb des
Programmordners (in Docker-Volumes). Egal wohin ihr entpackt — der Schlüssel folgt
den Daten, jotti findet beides beim Start wieder. Den alten Ordner könnt ihr danach
gefahrlos löschen.

> ⛔ **Niemals `docker compose down -v` ausführen.** Das `-v` löscht **alle**
> Docker-Volumes — und damit **Daten, Installations-Schlüssel und das grüne
> Zertifikat** unwiderruflich (auch ein Update bringt sie dann nicht zurück). Zum
> Beenden immer **`jotti-stop.cmd`** verwenden: das stoppt nur die Container und
> lässt alles erhalten.

**Automatisches Backup vor dem Update.** Erkennt der Starter eine neue Version,
sichert er die Datenbank **vor** der Aktualisierung automatisch. Geht beim Update
etwas schief, stellt **`jotti-restore.cmd`** (Doppelklick) das letzte dieser
Backups wieder her — seit dem Backup erfasste Daten gehen dabei verloren.

> 🔁 **Nur vorwärts, kein Downgrade.** Spielt **keine ältere Version** über eine
> neuere. Updates verändern die Datenbank und lassen sich nicht zurücknehmen;
> eine alte Version kann mit den neuen Daten nicht mehr starten.

## Wenn nach einem Update niemand mehr hineinkommt

Sehr selten — meist nach einem Update von einer **sehr alten** Version — passt das
in der Datenbank gespeicherte Passwort nicht mehr zum aktuellen
Installations-Schlüssel. jotti startet dann zwar, aber das Anmelden schlägt fehl.
**Eure Daten sind dabei nicht verloren** — nur das Schloss passt nicht zum
Schlüssel. Zwei datenerhaltende Wege zurück:

1. **`jotti-repair.cmd`** doppelklicken. Es gleicht das Datenbank-Passwort an den
   aktuellen Installations-Schlüssel an und startet jotti neu — ohne eure Daten zu
   verändern. Mehrfaches Ausführen schadet nicht. Danach einmal **neu anmelden**.
2. Habt ihr noch die **`.env` aus der alten Installation** (lag früher im
   Programmordner neben `jotti-start.exe`): kopiert sie nach
   **`%PROGRAMDATA%\jotti\.env`** und startet `jotti-start.exe` erneut — dann
   verwendet jotti wieder den ursprünglichen Schlüssel.

Meldet der Starter beim Hochfahren ausdrücklich, es seien **„bereits jotti-Daten
vorhanden, aber keine Zugangsdaten gefunden"**, dann hilft Weg 2: die alte `.env`
an den genannten Ort legen und erneut starten.

---

> 🔒 **Sicherheit:** jotti läuft nur im lokalen WLAN. Öffnet es **niemals** ins
> Internet — richtet im Router **keine Port-Weiterleitung** auf den Kassenrechner ein.
