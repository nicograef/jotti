# PRD: Klickbare Windows-Verpackung für den lokalen Betrieb

> Voraussetzungen:
> `docs/prds/prd-betrieb-relay-haertung.md` (Relay-Env-Konfiguration, `.env`-Vertrag,
> `/health` per GET) und `docs/prds/prd-lokale-tls-selbstsigniert.md` (lokales
> HTTPS — der Starter zeigt eine `https://`-Adresse an).
> Herkunft: gekürzt aus der ursprünglich gemischten Windows-PRD;
> Betriebs-/Relay-Härtung und Transportverschlüsselung sind in eigene PRDs
> abgetrennt. Diese PRD beschreibt ausschließlich die **klickbare
> Windows-Verpackung**.

## Problem Statement

Ein technisch wenig erfahrener Vereins-Admin möchte jotti für ein kleines bis
mittleres Fest auf einem vorhandenen Windows-Rechner betreiben, sodass die
Helfer mit ihren eigenen Smartphones im selben WLAN bestellen, kassieren und
Bons drucken können. Der lokale Docker-Weg bleibt die Grundlage, ist in seiner
**Bedienung** aber zu technisch:

- Die Konfigurationsdatei `.env` muss von Hand angelegt und mit selbst erzeugten
  Geheimnissen gefüllt werden — inklusive `openssl`-Aufruf, der unter Windows
  nicht ohne Weiteres vorhanden ist.
- Der Start erfolgt über einen Kommandozeilenbefehl; die lokale IP-Adresse fürs
  Smartphone muss separat ermittelt werden.
- Der Bondruck (Print-Relay) muss als zweites Programm gestartet werden, ohne
  dass der Admin einen Token oder eine Backend-Adresse von Hand eingeben will.
- Typische Startfehler (Docker läuft nicht, Port belegt, Windows-Firewall)
  äußern sich in kryptischen Docker-Meldungen, die ein Ehrenamtlicher nicht
  einordnen kann.

> Hinweis zur Abgrenzung: Dass `RELAY_AUTH_TOKEN` überhaupt sicher existiert, das
> Relay env-basiert konfiguriert wird und `/health` per GET prüfbar ist, löst die
> Betriebs-/Relay-Härtung (`docs/prds/prd-betrieb-relay-haertung.md`). Diese PRD
> baut darauf auf und macht den Ablauf **per Doppelklick bedienbar**.

## Solution

jotti erhält für den lokalen Windows-Betrieb zwei kleine, eigenständige
Programme, die ein Verein als fertiges Release-ZIP herunterlädt, entpackt und
per Doppelklick startet — **Docker Desktop bleibt die Basis**, es kommt keine
neue Laufzeitumgebung hinzu.

1. **`jotti-start.exe` (Starter)** übernimmt Einrichtung und Start:
   - Erzeugt beim ersten Start automatisch eine `.env` mit kryptografisch
     sicheren Zufallswerten für alle Geheimnisse und überschreibt eine
     vorhandene `.env` **nie**. Schema und Schlüssel folgen dem `.env`-Vertrag
     aus der Betriebs-/Relay-Härtung (inkl. `RELAY_AUTH_TOKEN`).
   - Führt vor dem Start Preflight-Prüfungen aus (ist Docker installiert und
     gestartet? ist der konfigurierte Port frei?) und erklärt Fehler in
     verständlichem Deutsch.
   - Startet den lokalen Compose-Stack, prüft anschließend per **GET**-Aufruf
     gegen `…/health`, dass das Backend antwortet, und zeigt die **lokale
     Zugriffsadresse fürs Smartphone** an (z. B. `https://192.168.1.50`).

2. **`jotti-relay.exe` (Print-Relay)** wird vom Verein **separat** per
   Doppelklick gestartet (nicht vom Starter mitgestartet) und ist dadurch
   unabhängig start- und neustartbar. Es liest seinen Token und die lokale
   Backend-Adresse aus **derselben `.env`-Datei**, die der Starter erzeugt hat —
   die Datei ist der einzige Vertrag zwischen beiden Programmen. Eine manuelle
   Token-Eingabe entfällt.

Der Verein muss damit weder eine `.env` von Hand pflegen noch Geheimnisse
erzeugen noch seine IP-Adresse suchen. Zwei Doppelklicks genügen: Starter für
Kasse und Web-UI, Relay für den Bondruck.

> 🔒 Die Transportverschlüsselung des lokalen Modus (HTTPS) regelt
> `docs/prds/prd-lokale-tls-selbstsigniert.md`. Der Starter zeigt die
> `https://`-Adresse an und weist auf die einmalige Zertifikatswarnung pro Gerät
> hin. Der Rechner darf niemals per Port-Weiterleitung ins Internet geöffnet
> werden.

## User Stories

### Erstinstallation & Konfiguration

1. Als Vereins-Admin möchte ich ein einziges ZIP herunterladen, das alles Nötige
   enthält (Starter, Relay, Compose-Dateien, Kurzanleitung), damit ich nichts
   einzeln zusammensuchen muss.
2. Als Vereins-Admin möchte ich beim ersten Start keine `.env` von Hand anlegen
   müssen, damit ich nicht wissen muss, welche Variablen es gibt.
3. Als Vereins-Admin möchte ich, dass alle Geheimnisse automatisch sicher erzeugt
   werden, damit ich kein `openssl` oder Ähnliches brauche.
4. Als Vereins-Admin möchte ich, dass eine bereits vorhandene `.env` beim
   erneuten Start niemals überschrieben wird, damit meine Daten und Zugänge über
   mehrere Festtage stabil bleiben.
5. Als Vereins-Admin möchte ich Basiswerte (z. B. den Port) bei Bedarf in einer
   einfachen Datei ändern können, ohne dass dabei die Geheimnisse neu erzeugt
   werden.
6. Als Vereins-Admin möchte ich ohne interaktive Rückfragen starten können
   (sinnvolle Voreinstellungen), damit der Start so einfach wie möglich bleibt.

### Start & Zugriff

7. Als Vereins-Admin möchte ich jotti per Doppelklick auf `jotti-start.exe`
   starten, damit ich keine Kommandozeilenbefehle eintippen muss.
8. Als Vereins-Admin möchte ich nach dem Start klar angezeigt bekommen, unter
   welcher Adresse die Helfer-Smartphones jotti erreichen (z. B.
   `https://192.168.1.50`), damit ich diese Adresse weitergeben kann.
9. Als Vereins-Admin möchte ich eine Bestätigung sehen, dass das Backend wirklich
   erreichbar ist (Health-Check), bevor ich Helfer dazuhole.
10. Als Service-Helfer möchte ich im selben WLAN mit dem Smartphone-Browser über
    die angezeigte Adresse auf jotti zugreifen, damit ich ohne App-Installation
    arbeiten kann.

### Fehlerdiagnose

11. Als Vereins-Admin möchte ich eine verständliche Meldung erhalten, wenn Docker
    Desktop nicht installiert oder nicht gestartet ist, damit ich weiß, dass ich
    Docker zuerst starten muss.
12. Als Vereins-Admin möchte ich eine verständliche Meldung erhalten, wenn der
    benötigte Port bereits belegt ist, samt Hinweis, wie ich den Port in der
    Datei ändere.
13. Als Vereins-Admin möchte ich einen Hinweis zur Windows-Firewall erhalten,
    falls das Smartphone den Rechner nicht erreicht, damit ich den Zugriff für
    private Netzwerke freigeben kann.
14. Als Vereins-Admin möchte ich, dass der Starter mit einem klaren Exit-Status
    endet (Erfolg/Fehler), damit ich erkenne, ob der Start geklappt hat.

### Bondruck per Doppelklick

15. Als Vereins-Admin möchte ich `jotti-relay.exe` per Doppelklick starten, ohne
    einen Token oder eine Backend-Adresse eingeben zu müssen, damit der Bondruck
    ohne technisches Wissen funktioniert.
16. Als Vereins-Admin möchte ich, dass das Relay seinen Token aus derselben
    `.env` bezieht, die der Starter erzeugt hat, damit Backend und Relay
    garantiert denselben Token verwenden.
17. Als Vereins-Admin möchte ich das Relay unabhängig vom Backend neu starten
    können (z. B. nach einem Druckerproblem), ohne den ganzen Stack neu zu
    starten.
18. Als Vereins-Admin möchte ich eine verständliche Relay-Ausgabe sehen
    (verbunden / kein Auftrag / Druckerfehler), damit ich den Druckbetrieb im
    Blick habe.

### Betrieb, Beenden & Dokumentation

19. Als Vereins-Admin möchte ich jotti am Ende des Tages sauber beenden können,
    ohne dass meine Daten verloren gehen.
20. Als Vereins-Admin möchte ich am nächsten Festtag mit denselben zwei
    Doppelklicks weitermachen, ohne erneute Einrichtung und mit unveränderten
    Daten (gleiche `.env`, gleiches Datenbank-Volume).
21. Als Vereins-Admin möchte ich eine kurze, schrittweise Anleitung im ZIP
    finden, die genau diesen Zwei-Doppelklick-Ablauf beschreibt.
22. Als Vereins-Admin möchte ich im Hosting-Leitfaden wiederfinden, dass der
    lokale Windows-Weg über den Starter läuft, damit Anleitung und Software
    übereinstimmen.

## Implementation Decisions

### Scope & Laufzeit

- Der bestehende lokale Docker-Stack bleibt die Grundlage; nur der Betrieb
  drumherum wird per Doppelklick bedienbar. Kein Einbetten des Frontends ins
  Go-Backend, kein Entfernen von nginx, kein nativer Installer.
- **Docker Desktop bleibt Pflicht-Basis.**
- **Distribution** als GitHub-Release-ZIP mit vorgebauten Windows-Binaries
  (`jotti-start.exe`, `jotti-relay.exe`), den lokalen Compose-Dateien und einer
  Kurzanleitung. **Es werden nur Windows-Binaries ausgeliefert** (der Go-Code
  bleibt portabel, aber andere Plattformen sind nicht Teil dieses Releases).

### Modul: Starter-Core (deep, rein/testbar)

Kapselt die Entscheidungs- und Aufbereitungslogik als seiteneffektfreie
Funktionen, getrennt von Docker- und Prozess-Aufrufen:

- **Secret-Erzeugung:** kryptografisch sichere Zufallswerte für die Secrets des
  `.env`-Vertrags (`POSTGRES_PASSWORD`, `JWT_SECRET`, `RELAY_AUTH_TOKEN`).
- **`.env`-Materialisierung:** erzeugt die `.env` nur, wenn sie fehlt; eine
  vorhandene Datei wird nie überschrieben (idempotent). `POSTGRES_USER` und der
  Port erhalten sinnvolle Defaults.
- **Preflight-Auswertung:** bildet die geprüften Bedingungen (Docker
  vorhanden/gestartet, Port frei) auf verständliche, deutsche Diagnose-Texte mit
  Handlungshinweis ab.
- **LAN-IP-Auswahl:** wählt aus den Netzwerkschnittstellen die passende private
  IPv4-Adresse für den WLAN-Zugriff aus (Loopback/Link-Local ignorieren).
- **Zugriffs-URL-Bau:** setzt aus IP und Port die anzuzeigende `https://`-Adresse
  zusammen.

### Modul: Starter-Shell (dünn, nicht unit-getestet)

- Ruft `docker compose … up -d --build` auf.
- Führt nach dem Start den **Health-Check als GET** gegen `…/health` aus
  (`/health` ist der bewusst von POST-only ausgenommene Ops-Endpunkt, siehe
  `docs/prds/prd-betrieb-relay-haertung.md`).
- Gibt die vom Core gelieferten Diagnosen/URL auf der Konsole aus und setzt
  passende Exit-Codes. Es gibt **keine** laufende Statussicht und **keine**
  Status-Webseite — nur diese einmalige Ausgabe beim Start.

### Modul: Relay-Start auf Windows

- `jotti-relay.exe` wird ohne Kommandozeilen-Argumente gestartet. Die
  env-basierte Konfiguration des Relays (`RELAY_AUTH_TOKEN`, `RELAY_BACKEND_URL`,
  `RELAY_POLL_SECONDS`) liefert die Betriebs-/Relay-Härtung. **Windows-spezifisch
  ergänzt diese PRD nur**, dass die `.exe` diese Werte aus derselben `.env`-Datei
  liest, die der Starter erzeugt hat (minimaler Key=Value-Parser), weil ein
  nackter Prozess auf Windows keine Compose-Env-Injektion hat.

### Port-Konfiguration & Build

- Die Port-Veröffentlichung des Reverse-Proxy ist konfigurierbar
  (`${HTTP_PORT:-…}` in der lokalen Compose-Datei), damit der „Port belegt"-Fall
  durch Editieren der `.env` lösbar ist; der Starter weist im Preflight darauf
  hin.
- Neues Make-Target zum Bauen des Starters analog `build-relay`; beide Targets
  erzeugen die Windows-Binaries für das Release.

## Testing Decisions

- **Starter-Core** (Unit-Tests, `-tags=unit`):
  - `.env`-Idempotenz: vorhandene Datei wird nie überschrieben; fehlt sie, wird
    sie mit allen erwarteten Schlüsseln erzeugt.
  - Secret-Erzeugung: ausreichende Länge/Entropie, erwartetes Format; zwei
    Aufrufe liefern unterschiedliche Werte.
  - LAN-IP-Auswahl: passende private IPv4 gewählt, Loopback/Link-Local ignoriert.
  - Preflight-Auswertung: jede Bedingung bildet auf die korrekte Diagnose mit
    Handlungshinweis ab.
  - Zugriffs-URL-Bau: korrekte `https://`-Adresse für Default-Port und
    abweichenden Port.
- **Relay-`.env`-Parser** (Unit-Tests, `-tags=unit`): Token und Backend-URL
  werden aus der `.env` gelesen; fehlt `RELAY_BACKEND_URL`, gilt der lokale
  Default.
- **Nicht unit-getestet:** die dünne Starter-Shell (Docker-/Prozess-Aufrufe,
  Konsolenausgabe).

## Out of Scope

- **Betriebs-/Relay-Härtung** (Relay-Token-Pflicht, Relay-Env-Semantik, `/health`
  per GET, Compose-Entdoppelung) → `docs/prds/prd-betrieb-relay-haertung.md`.
- **Transportverschlüsselung (TLS)** → `docs/prds/prd-lokale-tls-selbstsigniert.md`
  (Option 2) und `docs/prds/prd-lokale-tls-vertrauenswuerdig.md` (Option 3). Diese
  PRD konsumiert das lokale HTTPS nur (zeigt die `https://`-Adresse), definiert es
  aber nicht.
- **Phase B:** Frontend ins Go-Backend einbetten, nginx im lokalen Modus
  entfernen.
- **Phase C:** nativer Windows-Installer (MSI/Setup), Einrichtungs-Wizard,
  Windows-Dienst-Steuerung, gebündelte PostgreSQL, Ablösung von Docker.
- Speicherung von Geheimnissen im Windows Secret Store / via DPAPI (es bleibt bei
  der `.env`-Datei).
- Interaktiver Konfigurations-Wizard; laufende Statusanzeige oder Status-Webseite.
- Builds/Releases für macOS und Linux.
- Änderungen an POST-only, Event-Sourcing oder Datenmodell.
- Drucker-/Druckstations-Einrichtung (Ziel-IP der Drucker) — bestehende
  In-App-Funktion.

## Further Notes

- **Entkopplung über die Datei:** Starter und Relay kommunizieren nicht direkt;
  die vom Starter erzeugte `.env` ist der alleinige Vertrag. Beide Programme
  bleiben unabhängig start- und neustartbar.
- **Sicherheitsmodell:** Der lokale Modus ist für ein vertrauenswürdiges WLAN
  gedacht; die Transportverschlüsselung regelt die Option-2-PRD. Der Starter
  weist im Erfolgsfall sichtbar darauf hin, den Rechner nie ins Internet zu
  öffnen.
- **Ausblick native App (Phase C, nicht Teil dieser PRD):** Go könnte das Backend
  samt eingebettetem Frontend (`go:embed`) zu einer einzigen `jotti.exe`
  kompilieren und eine eigenständige PostgreSQL-Windows-Binary als Kindprozess
  starten; verpackt in einen Inno-Setup-/MSI-Installer ergäbe das echten
  Doppelklick-Betrieb ganz ohne Docker. Spätere Ausbaustufe.
