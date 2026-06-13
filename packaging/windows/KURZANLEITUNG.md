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
  gar nicht, blockiert vermutlich der Router (DNS-Rebind-Schutz) — die
  Router-Anleitung steht in der Betriebsdoku (`docs/betrieb/dns-rebind-schutz.md`).

## Bondruck (optional)

Für den Bondruck zusätzlich **`jotti-relay.exe`** doppelklicken. Es läuft ohne
Administratorrechte und nimmt seine Zugangsdaten aus der `.env` im selben Ordner.

## Probleme

- **„Port 80/443 ist durch ‚X' (PID …) belegt":** Das genannte Programm beenden
  (häufig Skype, IIS oder eine VM-Software) und `jotti-start.exe` erneut starten.
- **Fenster schließt sich zu schnell:** Es bleibt bis zum Enter-Druck offen;
  steht oben eine Fehlermeldung, diese zuerst lesen.

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
Programmordners. Egal wohin ihr entpackt — jotti findet sie beim Start wieder.
Den alten Ordner könnt ihr danach gefahrlos löschen.

**Automatisches Backup vor dem Update.** Erkennt der Starter eine neue Version,
sichert er die Datenbank **vor** der Aktualisierung automatisch. Geht beim Update
etwas schief, stellt **`jotti-restore.cmd`** (Doppelklick) das letzte dieser
Backups wieder her — seit dem Backup erfasste Daten gehen dabei verloren.

> 🔁 **Nur vorwärts, kein Downgrade.** Spielt **keine ältere Version** über eine
> neuere. Updates verändern die Datenbank und lassen sich nicht zurücknehmen;
> eine alte Version kann mit den neuen Daten nicht mehr starten.

---

> 🔒 **Sicherheit:** jotti läuft nur im lokalen WLAN. Öffnet es **niemals** ins
> Internet — richtet im Router **keine Port-Weiterleitung** auf den Kassenrechner ein.
