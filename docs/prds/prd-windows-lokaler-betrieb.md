# PRD: Einfacher lokaler Windows-Betrieb für Vereine

## Problem Statement

Ein technisch wenig erfahrener Vereins-Admin möchte jotti für ein kleines bis
mittleres Fest auf einem vorhandenen Windows-Rechner betreiben, sodass die
Helfer mit ihren eigenen Smartphones im selben WLAN bestellen, kassieren und
Bons drucken können. Der heutige lokale Weg (`docker-compose.local.yml`) ist
dafür zu technisch:

- Die Konfigurationsdatei `.env` muss von Hand angelegt und mit selbst
  erzeugten Geheimnissen (`POSTGRES_PASSWORD`, `JWT_SECRET`) gefüllt werden —
  inklusive `openssl`-Aufruf, der unter Windows nicht ohne Weiteres vorhanden
  ist.
- Der Start erfolgt über einen Kommandozeilenbefehl; die lokale IP-Adresse
  fürs Smartphone muss separat ermittelt werden.
- Der Bondruck (Print-Relay) ist im lokalen Stack gar nicht verdrahtet:
  `RELAY_AUTH_TOKEN` fehlt sowohl in `.env.example` als auch in der Backend-Env
  von `docker-compose.local.yml`, und das Relay erwartet heute den Token als
  Pflicht-Flag.
- Typische Startfehler (Docker läuft nicht, Port 80 belegt, Windows-Firewall)
  äußern sich in kryptischen Docker-Meldungen, die ein Ehrenamtlicher nicht
  einordnen kann.

In Summe entsteht ein hoher Einrichtungs- und Support-Aufwand für genau die
Zielgruppe, die am wenigsten technische Hilfe zur Verfügung hat.

## Solution

jotti erhält für den lokalen Windows-Betrieb zwei kleine, eigenständige
Programme, die ein Verein als fertiges Release-ZIP herunterlädt, entpackt und
per Doppelklick startet — **Docker Desktop bleibt die Basis**, es kommt keine
neue Laufzeitumgebung hinzu.

1. **`jotti-start.exe` (Starter)** übernimmt die gesamte Einrichtung und den
   Start:
   - Erzeugt beim ersten Start automatisch eine `.env` mit kryptografisch
     sicheren Zufallswerten für alle Geheimnisse und überschreibt eine
     vorhandene `.env` **nie**.
   - Führt vor dem Start Preflight-Prüfungen aus (ist Docker installiert und
     gestartet? ist der konfigurierte Port frei?) und erklärt Fehler in
     verständlichem Deutsch.
   - Startet den lokalen Compose-Stack, prüft anschließend per Health-Check,
     dass das Backend antwortet, und zeigt die **lokale Zugriffsadresse fürs
     Smartphone** an (z. B. `http://192.168.1.50`).

2. **`jotti-relay.exe` (Print-Relay)** wird vom Verein **separat** per
   Doppelklick gestartet (nicht vom Starter mitgestartet) und ist dadurch
   unabhängig start- und neustartbar. Es liest seinen Token und den lokalen
   HTTP-Port aus **derselben `.env`-Datei**, die der Starter erzeugt hat — die
   Datei ist der einzige Vertrag zwischen beiden Programmen. Eine manuelle
   Token-Eingabe entfällt.

Der Verein muss damit weder eine `.env` von Hand pflegen noch Geheimnisse
erzeugen noch seine IP-Adresse suchen. Zwei Doppelklicks genügen: Starter für
Kasse und Web-UI, Relay für den Bondruck.

> 🔒 Der lokale Modus läuft bewusst unverschlüsselt über HTTP und ist
> ausschließlich für das eigene, vertrauenswürdige WLAN gedacht. Der Rechner
> darf niemals per Port-Weiterleitung ins Internet geöffnet werden.

## User Stories

### Erstinstallation & Konfiguration

1. Als Vereins-Admin möchte ich ein einziges ZIP herunterladen, das alles
   Nötige enthält (Starter, Relay, Compose-Dateien, Kurzanleitung), damit ich
   nichts einzeln zusammensuchen muss.
2. Als Vereins-Admin möchte ich beim ersten Start keine `.env` von Hand
   anlegen müssen, damit ich nicht wissen muss, welche Variablen es gibt.
3. Als Vereins-Admin möchte ich, dass alle Geheimnisse (Datenbank-Passwort,
   JWT-Secret, Relay-Token) automatisch sicher erzeugt werden, damit ich kein
   `openssl` oder Ähnliches brauche.
4. Als Vereins-Admin möchte ich, dass eine bereits vorhandene `.env` beim
   erneuten Start niemals überschrieben wird, damit meine Daten und Zugänge
   über mehrere Festtage stabil bleiben.
5. Als Vereins-Admin möchte ich Basiswerte (z. B. den HTTP-Port) bei Bedarf in
   einer einfachen Datei ändern können, ohne dass dabei die Geheimnisse neu
   erzeugt werden.
6. Als Vereins-Admin möchte ich ohne interaktive Rückfragen starten können
   (sinnvolle Voreinstellungen: Port 80, IP automatisch), damit der Start so
   einfach wie möglich bleibt.

### Start & Zugriff

7. Als Vereins-Admin möchte ich jotti per Doppelklick auf `jotti-start.exe`
   starten, damit ich keine Kommandozeilenbefehle eintippen muss.
8. Als Vereins-Admin möchte ich nach dem Start klar angezeigt bekommen, unter
   welcher Adresse die Helfer-Smartphones jotti erreichen (z. B.
   `http://192.168.1.50`), damit ich diese Adresse weitergeben kann.
9. Als Vereins-Admin möchte ich eine Bestätigung sehen, dass das Backend
   wirklich erreichbar ist (Health-Check), bevor ich Helfer dazuhole.
10. Als Service-Helfer möchte ich im selben WLAN mit dem Smartphone-Browser
    über die angezeigte Adresse auf jotti zugreifen, damit ich ohne App-
    Installation arbeiten kann.
11. Als Service-Helfer möchte ich jotti über „Zum Startbildschirm hinzufügen"
    wie eine App ablegen können, damit der Zugriff während des Fests schnell
    geht.

### Fehlerdiagnose

12. Als Vereins-Admin möchte ich eine verständliche Meldung erhalten, wenn
    Docker Desktop nicht installiert oder nicht gestartet ist, damit ich weiß,
    dass ich Docker zuerst starten muss.
13. Als Vereins-Admin möchte ich eine verständliche Meldung erhalten, wenn der
    benötigte Port bereits belegt ist, samt Hinweis, wie ich den Port in der
    Datei ändere.
14. Als Vereins-Admin möchte ich einen Hinweis zur Windows-Firewall erhalten,
    falls das Smartphone den Rechner nicht erreicht, damit ich den Zugriff für
    private Netzwerke freigeben kann.
15. Als Vereins-Admin möchte ich, dass der Starter mit einem klaren Exit-Status
    endet (Erfolg/Fehler), damit ich erkenne, ob der Start geklappt hat.

### Bondruck (Relay)

16. Als Vereins-Admin möchte ich `jotti-relay.exe` per Doppelklick starten, ohne
    einen Token oder eine Backend-Adresse eingeben zu müssen, damit der Bondruck
    ohne technisches Wissen funktioniert.
17. Als Vereins-Admin möchte ich, dass das Relay seinen Token aus derselben
    `.env` bezieht, die der Starter erzeugt hat, damit Backend und Relay
    garantiert denselben Token verwenden.
18. Als Vereins-Admin möchte ich das Relay unabhängig vom Backend neu starten
    können (z. B. nach einem Druckerproblem), ohne den ganzen Stack neu zu
    starten.
19. Als Vereins-Admin möchte ich, dass offene Druckaufträge weiterhin korrekt
    gedruckt werden, sobald das Relay läuft, auch wenn es zwischenzeitlich aus
    war.
20. Als Vereins-Admin möchte ich eine verständliche Relay-Ausgabe sehen
    (verbunden / kein Auftrag / Druckerfehler), damit ich den Druckbetrieb im
    Blick habe.

### Betrieb & Beenden

21. Als Vereins-Admin möchte ich jotti am Ende des Tages sauber beenden können,
    ohne dass meine Daten verloren gehen.
22. Als Vereins-Admin möchte ich am nächsten Festtag mit denselben zwei
    Doppelklicks weitermachen, ohne erneute Einrichtung.
23. Als Vereins-Admin möchte ich, dass meine bestehenden Daten beim erneuten
    Start unverändert bereitstehen (gleiche `.env`, gleiches Datenbank-Volume).

### Dokumentation

24. Als Vereins-Admin möchte ich eine kurze, schrittweise Anleitung im ZIP
    finden, die genau diesen Zwei-Doppelklick-Ablauf beschreibt.
25. Als Vereins-Admin möchte ich im Hosting-Leitfaden wiederfinden, dass der
    lokale Windows-Weg jetzt über den Starter läuft, damit Anleitung und
    Software übereinstimmen.
26. Als Vereins-Admin möchte ich den ausdrücklichen Sicherheitshinweis lesen,
    dass dieser Modus nur fürs lokale WLAN gedacht ist und nie ins Internet
    geöffnet werden darf.

## Implementation Decisions

### Scope & Laufzeit

- **Nur Phase A** aus der ursprünglichen Idee: Der bestehende lokale
  Docker-Stack bleibt die Grundlage; der Betrieb drumherum wird vereinfacht.
  Kein Einbetten des Frontends ins Go-Backend, kein Entfernen von nginx, kein
  nativer Installer.
- **Docker Desktop bleibt Pflicht-Basis** für den lokalen Windows-Betrieb.
- **Distribution** als GitHub-Release-ZIP mit vorgebauten Windows-Binaries
  (`jotti-start.exe`, `jotti-relay.exe`), den lokalen Compose-Dateien und einer
  Kurzanleitung. **Es werden nur Windows-Binaries ausgeliefert** (der Go-Code
  bleibt portabel, aber andere Plattformen sind nicht Teil dieses Releases).

### Modul: Starter-Core (deep, rein/testbar)

Kapselt die gesamte Entscheidungs- und Aufbereitungslogik als seiteneffektfreie
Funktionen, getrennt von Docker- und Prozess-Aufrufen:

- **Secret-Erzeugung:** kryptografisch sichere Zufallswerte für
  `POSTGRES_PASSWORD`, `JWT_SECRET` und `RELAY_AUTH_TOKEN`.
- **`.env`-Materialisierung:** erzeugt die `.env` nur, wenn sie fehlt; eine
  vorhandene Datei wird nie überschrieben (idempotent). `POSTGRES_USER` und
  `HTTP_PORT` erhalten sinnvolle Defaults (`admin`, `80`).
- **Preflight-Auswertung:** bildet die geprüften Bedingungen (Docker
  vorhanden/gestartet, Port frei) auf verständliche, deutsche Diagnose-Texte
  mit Handlungshinweis ab.
- **LAN-IP-Auswahl:** wählt aus den Netzwerkschnittstellen die passende
  private IPv4-Adresse für den WLAN-Zugriff aus.
- **Zugriffs-URL-Bau:** setzt aus IP und Port die anzuzeigende Adresse
  zusammen (Port 80 ohne expliziten Port-Suffix).

### Modul: Starter-Shell (dünn, nicht unit-getestet)

- Ruft `docker compose -f docker-compose.local.yml up -d --build` auf.
- Führt nach dem Start den **Health-Check als POST** gegen `…/health` aus
  (alle jotti-Routen sind POST-only, inklusive `/health`).
- Gibt die vom Core gelieferten Diagnosen/URL auf der Konsole aus und setzt
  passende Exit-Codes. Es gibt **keine** laufende Statussicht und **keine**
  Status-Webseite — nur diese einmalige Ausgabe beim Start.

### Modul: Relay-Konfiguration

- `jotti-relay.exe` liest beim Start die `.env` aus seinem Arbeitsverzeichnis
  und entnimmt ihr `RELAY_AUTH_TOKEN` sowie `HTTP_PORT`. Daraus bildet es die
  lokale Backend-URL (Default-Port 80) und spricht das Backend über den
  Reverse-Proxy unter dem `/api`-Pfad an.
- Der Token ist damit **kein Pflicht-Flag mehr**; vorhandene Flags
  (`--backend`, `--token`, `--poll`) bleiben als Override für Sonderfälle
  erhalten (z. B. Relay auf einem zweiten Gerät).
- Erfordert einen minimalen `.env`-Parser im Relay (Key=Value-Zeilen).

### Compose- & `.env`-Verdrahtung

- `docker-compose.local.yml`: Der Backend-Service erhält
  `RELAY_AUTH_TOKEN: ${RELAY_AUTH_TOKEN}` (Compose lädt die `.env` automatisch),
  damit der Bondruck im lokalen Modus überhaupt funktioniert.
- `docker-compose.local.yml`: Die Port-Veröffentlichung des Reverse-Proxy wird
  konfigurierbar (`${HTTP_PORT:-80}:80`), damit der „Port belegt"-Fall durch
  Editieren der `.env` lösbar ist.
- `.env.example` dokumentiert zusätzlich `RELAY_AUTH_TOKEN` und `HTTP_PORT`.

### Build & Make

- Neues Make-Target zum Bauen des Starters analog zu `build-relay`; beide
  Targets erzeugen die Windows-Binaries für das Release.

## Testing Decisions

Ein guter Test prüft **externes Verhalten, keine Implementierungsdetails**: Er
beschreibt, was ein Modul für gegebene Eingaben liefert, und bleibt stabil,
wenn sich die innere Umsetzung ändert.

- **Starter-Core** (Unit-Tests, `-tags=unit`):
  - `.env`-Idempotenz: eine vorhandene Datei wird nie überschrieben; fehlt sie,
    wird sie mit allen erwarteten Schlüsseln erzeugt.
  - Secret-Erzeugung: erzeugte Werte haben ausreichende Länge/Entropie und das
    erwartete Format; zwei Aufrufe liefern unterschiedliche Werte.
  - LAN-IP-Auswahl: aus einer Menge von Schnittstellen-Adressen wird die
    passende private IPv4 gewählt (und Loopback/Link-Local ignoriert).
  - Preflight-Auswertung: jede geprüfte Bedingung bildet auf die korrekte,
    verständliche Diagnose mit Handlungshinweis ab.
  - Zugriffs-URL-Bau: korrekte Adresse für Default-Port und abweichenden Port.
- **Relay-Konfiguration** (Unit-Tests, `-tags=unit`):
  - Token wird aus der `.env` gelesen; die lokale Backend-URL wird mit
    Default-Port korrekt gebildet; ein explizites Flag hat Vorrang vor dem
    `.env`-Wert.

- **Prior Art:** bestehende Go-Unit-Tests im Repo, z. B. `config/config_test.go`
  (Env-Parsing) und die HTTP-Handler-Tests unter `backend/api/**/http/*_test.go`,
  die mit `-tags=unit` laufen und ausschließlich beobachtbares Verhalten prüfen.
- **Nicht unit-getestet:** die dünne Starter-Shell (Docker-/Prozess-Aufrufe,
  Konsolenausgabe) — ihr Verhalten ergibt sich aus dem getesteten Core plus
  Standard-Tooling.

## Out of Scope

- **Phase B:** Frontend ins Go-Backend einbetten, nginx im lokalen Modus
  entfernen, UI-GET/API-POST-Routing zusammenführen.
- **Phase C:** nativer Windows-Installer (MSI/Setup), Einrichtungs-Wizard,
  Windows-Dienst-Steuerung, gebündelte PostgreSQL, Ablösung von Docker.
- Speicherung von Geheimnissen im Windows Secret Store / via DPAPI (es bleibt
  bei der `.env`-Datei).
- Interaktiver Konfigurations-Wizard (Start erfolgt ohne Rückfragen).
- Laufende Statusanzeige oder Status-Webseite während des Betriebs.
- Internet-/Cloud-Betrieb, TLS/HTTPS, Domain — dafür bleibt der bestehende
  Server-Weg (`docker-compose.prod.yml`) zuständig.
- Builds/Releases für macOS und Linux.
- Änderungen an der POST-only-Regel, am Event-Sourcing oder am Datenmodell.
- Drucker-/Druckstations-Einrichtung (Ziel-IP der Drucker) — das bleibt eine
  bestehende In-App-Funktion und steckt bereits im Druckauftrag.

## Further Notes

- **Entkopplung über die Datei:** Starter und Relay kommunizieren nicht direkt
  miteinander; die vom Starter erzeugte `.env` ist der alleinige Vertrag.
  Dadurch bleiben beide Programme unabhängig start- und neustartbar — ein
  ausdrückliches Akzeptanzkriterium.
- **Relay auf zweitem Gerät (optional):** Der Standardfall ist „beides auf
  demselben Rechner". Wer das Relay auf einer zweiten Station betreiben will,
  kann Backend-URL und Token weiterhin per Flag übergeben. Dieser Fall wird nur
  dokumentiert, nicht eigens gebaut.
- **Sicherheitsmodell:** Der lokale Modus ist unverschlüsseltes HTTP für ein
  vertrauenswürdiges WLAN. Der Starter weist im Erfolgsfall sichtbar darauf hin,
  den Rechner nie ins Internet zu öffnen.
- **Ausblick native App (Phase C, nicht Teil dieser PRD):** Go könnte das
  Backend samt eingebettetem Frontend (`go:embed`) zu einer einzigen `jotti.exe`
  kompilieren und eine offizielle eigenständige PostgreSQL-Windows-Binary als
  Kindprozess mit Datenverzeichnis unter `%APPDATA%` starten; verpackt in einen
  Inno-Setup-/MSI-Installer mit Windows-Dienst ergäbe das einen echten
  Doppelklick-Betrieb ganz ohne Docker. Das ist als spätere Ausbaustufe
  vorgemerkt.
