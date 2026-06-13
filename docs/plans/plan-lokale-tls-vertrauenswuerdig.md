# Plan: Lokale Transportverschlüsselung — vertrauenswürdiges Zertifikat (Option 3)

> Source PRD: `docs/prds/prd-lokale-tls-vertrauenswuerdig.md`
>
> Erstellt 2026-06-12. Die Klärungsrunde hat die offenen Fragen der PRD
> entschieden (siehe „Resolved decisions"); damit ist die in der PRD geforderte
> Betriebszusage bestätigt und der PRD-Status „kein Umsetzungsplan" überholt
> (Ripple in Phase 6).

## Goal

Ein Helfer-Smartphone öffnet die lokale jotti-Kasse über
`https://<lan-ip-mit-bindestrichen>.<install-id>.lokal.jotti.rocks` mit einem
echten Let's-Encrypt-Zertifikat — grünes Schloss, keine Warnung, kein
CA-Rollout, keine Einrichtung pro Gerät. Ein aktiver MITM im Vereins-WLAN
scheitert hart (Art. 32 DSGVO). Der bisherige Option-2-Mechanismus
(selbstsigniert auf `https://<LAN-IP>`) bleibt als eingebauter Fallback
erhalten; das hand-verdrahtete nginx-Setup des lokalen Stacks wird durch Caddy
ersetzt. jotti betreibt dafür dauerhaft einen zustandslosen DNS-Resolver und
eine acme-dns-Instanz auf dem bestehenden rocks-VPS.

## Architectural decisions

### Hostname & Install-ID

- **Hostname-Schema:** `<lan-ip-mit-bindestrichen>.<install-id>.lokal.jotti.rocks`
  (z. B. `192-168-1-50.8e5700b1-….lokal.jotti.rocks`). Die Ableitung
  LAN-IP + Install-ID → Hostname ist eine reine, unit-getestete Funktion.
- **Install-ID = acme-dns-Subdomain, unverändert übernommen.** acme-dns vergibt
  bei der Registrierung eine UUID-Subdomain (36 Zeichen < 63-Zeichen-Label-Limit).
  Dadurch kann der Resolver den Challenge-CNAME
  `_acme-challenge.<id>.lokal.jotti.rocks → <id>.auth.jotti.rocks` rein
  rechnerisch beantworten und bleibt vollständig zustandslos.
- **Zertifikat:** ein Wildcard `*.<install-id>.lokal.jotti.rocks` je
  Installation via DNS-01 (acme-dns-Credentials der Installation). IP-Wechsel
  (DHCP) ändert nur den Hostnamen, nie das Zertifikat.

### DNS-Topologie auf dem rocks-VPS (ein Port 53 je IP)

- **Resolver** (neues, kleines Go-Binary, sslip.io-Muster) bindet als einziger
  Prozess Port 53 (UDP+TCP) auf dem VPS und ist autoritativ für
  `lokal.jotti.rocks`: berechnete A-Records (hohe TTL, 86400 s — das Mapping
  Name → IP ist unveränderlich), berechnete `_acme-challenge`-CNAMEs, SOA/NS
  für die Zone, sonst NXDOMAIN.
- **acme-dns** (offizielles Image, Fertigkomponente) ist autoritativ für
  `auth.jotti.rocks`, lauscht aber **nur Docker-intern**. Der Resolver leitet
  DNS-Anfragen für `auth.jotti.rocks` und Subdomains intern an acme-dns weiter
  (Conditional Forwarding) — so brauchen beide Zonen nur eine öffentliche
  IP:53.
- **acme-dns-HTTP-API** (`/register`, `/update`) läuft hinter dem bestehenden
  rocks-nginx unter `https://auth.jotti.rocks` (der A-Record dafür kommt aus
  acme-dns selbst und zeigt auf den VPS; Zertifikat via certbot wie die
  übrigen rocks-Domains).
- **Delegation beim DNS-Hoster:** `dns.jotti.rocks A <VPS-IP>` plus
  `lokal.jotti.rocks NS dns.jotti.rocks` und
  `auth.jotti.rocks NS dns.jotti.rocks`. Dazu CAA auf `jotti.rocks`
  (nur Let's Encrypt) als Defense-in-Depth.

### Lokaler Stack

- **Caddy ersetzt nginx** in `docker-compose.local.yml`: Custom-Image via
  xcaddy mit dem offiziellen Modul `dns.providers.acmedns` (gepinnte
  Versionen). prod/rocks bleiben unverändert auf nginx.
- **Ein Hilfsprogramm, ein Container:** ein kleines Go-Programm (eigenes
  Modul, Muster `cmd/relay`) ist Entrypoint des Caddy-Containers. Ablauf:
  State sicherstellen (Install-ID + acme-dns-Credentials, einmalige
  Registrierung) → LAN-IP erkennen → Caddyfile rendern → Caddy als
  Kindprozess starten → Status-Seite servieren. Es ersetzt
  `reverse-proxy/local-entrypoint.sh` und `reverse-proxy/nginx.local.conf`.
- **Zwei Sites im generierten Caddyfile:** das Wildcard
  `*.<install-id>.lokal.jotti.rocks` (Let's Encrypt via DNS-01, automatische
  Erneuerung durch Caddy) und der Fallback `https://<LAN-IP>` mit Caddys
  interner CA (eingebauter Option-2-Ersatz, einmalige Warnung). Beide proxyen
  identisch: `/api/*` → backend:3000 (Prefix-Strip), `/` → frontend:80,
  Security-Header wie bisher.
- **Kein Rate-Limit-Modul im lokalen Caddy:** Das Backend hat eigenes
  Per-IP-Rate-Limiting (`backend/api/middleware/middleware.go:66-98`, Q-07);
  das nginx-Limit war Defense-in-Depth und entfällt lokal ersatzlos (KISS).
- **Status-Seite** unter `http://localhost:8484`, nur an `127.0.0.1`
  publiziert (nur am Kassen-Rechner sichtbar): aktueller Zustand, beide
  Adressen, QR-Code, Rebind-Hinweis mit Anleitung-Link. `make local-up` (und
  später der Windows-Starter) verweisen nur auf diese URL — die Anzeige selbst
  ist plattformneutral im Stack.
- **State & Volumes:** ein Volume für den Installations-State
  (Install-ID + Credentials als JSON) und Caddys Datenverzeichnis
  (Zertifikate, ACME-Account); das bisherige `tls-certs`-Volume entfällt.
- **„Zertifikat (noch) nicht da" ist Normalfall** (45-Tage-Laufzeiten,
  Saisonbetrieb): Caddy startet sofort und stellt im Hintergrund aus/erneuert;
  die Status-Seite zeigt bis dahin die Fallback-Adresse. Ohne Internet bleibt
  der Fallback dauerhaft funktionsfähig.

### Repo-Layout & CI

- **Resolver:** neues Top-Level-Verzeichnis `resolver/` — eigenes Go-Modul
  (einzige Dependency: `miekg/dns`), eigener multi-stage Dockerfile (Muster
  `backend/`). CI: paths-filter + Job analog `backend-ci`.
- **Lokales Hilfsprogramm:** Go-Modul unter `reverse-proxy/` (Quellort des
  lokalen Proxys), wird im selben Dockerfile wie der xcaddy-Build kompiliert.
  Dependencies: stdlib + eine QR-Code-Bibliothek.

## Inventory

- `reverse-proxy/local-entrypoint.sh:13-27` — LAN-IP-Erkennung über
  Default-Route (Logik übernimmt das Hilfsprogramm); `:29-84` —
  openssl-Zertifikatserzeugung und Idempotenz (entfällt komplett).
- `reverse-proxy/nginx.local.conf:13-14` — nginx-Rate-Limit (entfällt, s. o.);
  `:35-41` — Security-Header (Parität in Caddy); `:44-54` — `/api/`-Proxy mit
  Prefix-Strip; `:56-64` — Frontend-Proxy.
- `docker-compose.local.yml:101-119` — reverse-proxy-Service (nginx-Image,
  Entrypoint-Mounts → zu ersetzen); `:128-130` — Volumes (`tls-certs`
  entfällt).
- `Makefile:152-154` — `local-up` inkl. `--force-recreate reverse-proxy`
  (bleibt: das Caddyfile-Rendering bei IP-Wechsel läuft im Entrypoint).
- `docker-compose.rocks.yml:9-16` — rocks-Overlay (Ankerpunkt für die neuen
  Services resolver + acme-dns).
- `reverse-proxy/nginx.rocks.conf:20-32` — Port-80-Block (certbot-Challenges);
  `:50-97` — Haupt-Server-Block als Muster für den neuen
  `auth.jotti.rocks`-Block.
- `scripts/rocks-init.sh:20-22, 80-103` — Domain-Liste + DNS-Checks +
  certbot-Domains (um `auth.jotti.rocks` ergänzen).
- `backend/api/middleware/middleware.go:66-98` — Backend-Rate-Limiting
  (Begründung für den Wegfall des Proxy-Limits).
- `cmd/relay/go.mod` + `cmd/relay/main_test.go` — Muster für eigenständige
  Go-Module und ungetaggte Tests.
- `.github/workflows/ci.yml:20-32` — paths-filter ohne `resolver/**` und ohne
  `reverse-proxy/**`.
- `docs/betrieb/leitfaden-hosting.md:42-128` — Weg A (Sicherheitshinweis und
  Ablauf werden auf das grüne Schloss umgestellt).
- `docs/anforderungen.md:143` — Q-06 (lokal: „selbstsigniert, Browserwarnung"
  → aktualisieren).
- `docs/plans/plan-windows-verpackung.md:203` — Release-ZIP enthält
  `nginx.local.conf` + `local-entrypoint.sh` (Touchpoint: wird durch
  Caddy-Artefakte ersetzt, je nachdem welcher Plan später umgesetzt wird).

## Resolved decisions

- **Betriebszusage bestätigt; Infra auf dem bestehenden rocks-VPS** —
  Resolver + acme-dns als Services im rocks-Deployment
  (`docker-compose.rocks.yml`), kein separater Server (Klärung 2026-06-12).
- **acme-dns-Registrierung offen + Schutz** statt Registrierungs-Gate:
  nginx-Rate-Limit auf `/register`, CAA-Record, CT-Log-Monitoring als
  Betriebsroutine. Sicherheitsrelevant ist Offenheit nicht (Credentials sind
  je Install-ID gebunden; ein Fremder kann nur Zertifikate für seine *eigene*
  zufällige ID holen) — es bleibt ein Reputationsrisiko; ein Gate ist der
  dokumentierte Eskalationspfad (Klärung 2026-06-12).
- **Start-Anzeige als lokale Status-Seite** (localhost-only) im Stack selbst —
  nicht Konsole, nicht Frontend-App. Konsole (`make local-up`) und der spätere
  Windows-Starter verweisen nur auf die URL (Klärung 2026-06-12).
- **Unabhängig von der Windows-Verpackung geplant:** Ziel ist der lokale Stack
  (`make local-up`); die Berührungspunkte zum Windows-Plan werden in Phase 6
  als Ripples dokumentiert, wer später kommt, passt an (Klärung 2026-06-12).
- **Install-ID = acme-dns-Subdomain direkt** (keine eigene Ableitung) — hält
  den Resolver zustandslos und spart einen eigenen ID-Mechanismus.
- **Phasenzuschnitt (6 Phasen) bestätigt** (Klärung 2026-06-12).

## Open questions / Risks

- **Port 53 auf dem VPS:** Vorab prüfen, dass nichts öffentlich auf :53
  lauscht (Ubuntu/systemd-resolved bindet nur `127.0.0.53` — i. d. R.
  unkritisch) und die Provider-Firewall eingehend 53/UDP+TCP erlaubt.
- **DNS-Hoster:** muss NS-Records für Subdomain-Delegation unterstützen
  (Standard, aber vor Phase 2 verifizieren).
- **Fritz!Box-Subdomain-Frage** (deckt der Rebind-Ausnahme-Eintrag
  `lokal.jotti.rocks` Subdomains ab?): an echter Hardware in Phase 6
  verifizieren — bestimmt, ob die Anleitung einmalig genügt.
- **Rebind-Check-Fidelity:** Der Check läuft im Container über Dockers
  eingebetteten DNS → Host-Resolver → Router; das ist repräsentativ für die
  Handys im WLAN, aber Handys mit privatem DNS (DoH/DoT) können abweichen —
  bleibt Diagnose-Hinweis in der Doku, keine technische Lösung.
- **Let's-Encrypt-Staging in der Entwicklung:** Ausstellungs-Tests laufen
  gegen die LE-Staging-CA (env-Schalter im Hilfsprogramm), sonst drohen beim
  Testen Rate-Limits auf die echte Zone.
- **Missbrauch trotz offener Registrierung:** Rate-Limit + CT-Monitoring
  mindern, verhindern nicht. Eskalationspfad: Registrierungs-Gate (eigene
  Entscheidung, nicht Teil dieses Plans).
- **Windows-Plan-Drift:** Wird die Windows-Verpackung vor Option 3 umgesetzt,
  ändert sich danach der ZIP-Inhalt (Caddy-Image von GHCR statt
  nginx-Conf + Entrypoint-Skript) und die Starter-Erfolgsausgabe (Verweis auf
  die Status-Seite). In Phase 6 als Touchpoint dokumentiert.

---

## Phase 1: DNS-Resolver für `lokal.jotti.rocks`

**User stories**: 4, 5 (Fundament)

### Context

- `cmd/relay/go.mod`, `cmd/relay/main_test.go` — Muster für ein kleines
  eigenständiges Go-Modul mit ungetaggten Tests.
- `backend/Dockerfile` — Muster für den multi-stage Build.
- `.github/workflows/ci.yml:20-32` — paths-filter, dem `resolver/**` fehlt.
- PRD „Von jotti betriebene Infrastruktur" + „Testing Decisions" —
  Resolver-Verhalten und Testumfang sind dort verbindlich spezifiziert.

### What to build

Das neue Top-Level-Modul `resolver/`: ein zustandsloses, autoritatives
DNS-Binary für `lokal.jotti.rocks` (sslip.io-Muster, `miekg/dns`).

- **Reine Kernfunktion** (Frage → Antwort-Entscheidung, ohne I/O):
  - A-Frage `<ip-mit-bindestrichen>.<install-id>.lokal.jotti.rocks` → die
    eingebettete IPv4 (auch private Adressen), TTL 86400.
  - `_acme-challenge.<install-id>.lokal.jotti.rocks` → berechneter CNAME
    `<install-id>.auth.jotti.rocks`.
  - SOA/NS-Antworten für den Zone-Apex (Delegation funktioniert).
  - Alles andere (ungültige IP-Oktette, falsche Label-Anzahl, unbekannte
    Namen) → NXDOMAIN mit SOA im Authority-Abschnitt.
- **Server-Schale:** UDP+TCP auf :53; Konfiguration per Env (Zone,
  Listen-Adresse, VPS-IP für NS/A des eigenen Nameservers,
  Forward-Adresse der auth-Zone).
- **Conditional Forwarding:** Anfragen für `auth.jotti.rocks` und Subdomains
  werden unverändert an die konfigurierte acme-dns-Adresse weitergereicht,
  die Antwort durchgereicht.
- **Dockerfile** (multi-stage, kleines Laufzeit-Image) und **CI**:
  paths-filter um `resolver/**` ergänzen, Job analog `backend-ci`
  (vet, build, test).
- **Tests** gemäß PRD: Unit (Kernfunktion, tabellengetrieben) und Integration
  (laufender Server auf zufälligem Port, echte A-/CNAME-/NXDOMAIN-Anfragen
  über den `miekg/dns`-Client).

### Acceptance criteria

- [x] `dig @localhost -p <port> 192-168-1-50.<id>.lokal.jotti.rocks A` liefert
      `192.168.1.50` (auch für private IPs), TTL 86400.
- [x] `dig _acme-challenge.<id>.lokal.jotti.rocks` liefert den CNAME
      `<id>.auth.jotti.rocks`.
- [x] Ungültige Namen liefern NXDOMAIN; der Zone-Apex beantwortet SOA und NS.
- [x] Anfragen für `*.auth.jotti.rocks` werden an die konfigurierte Adresse
      weitergeleitet und die Antwort durchgereicht.
- [x] Unit- und Integrationstests grün; Kernlogik ist eine reine Funktion.
- [x] Docker-Image baut; CI-Job läuft bei Änderungen unter `resolver/**`.

---

## Phase 2: Infra-Deployment auf dem rocks-VPS

**User stories**: 4, 5

### Context

- `docker-compose.rocks.yml:9-16` — rocks-Overlay, hier kommen die Services
  dazu.
- `reverse-proxy/nginx.rocks.conf:20-32, 50-97` — Port-80-Block
  (certbot-Challenges) und Server-Block-Muster für `auth.jotti.rocks`.
- `scripts/rocks-init.sh:20-22, 80-103` — Domain-Liste, DNS-Checks,
  certbot-Domains.
- PRD „Von jotti betriebene Infrastruktur" — CAA, TTL, Rate-Limit-Rahmen.

### What to build

Die öffentliche DNS-Infrastruktur, end-to-end verifiziert:

- **acme-dns-Service** im rocks-Overlay: offizielles Image,
  `general.domain = auth.jotti.rocks`, DNS-Port nur Docker-intern, API nur
  hinter nginx erreichbar, SQLite-State in einem Volume.
- **Resolver-Service** (Phase-1-Image) im rocks-Overlay: Port 53 UDP+TCP
  publiziert, Forwarding der auth-Zone auf den acme-dns-Service.
- **nginx-Erweiterung:** Server-Block `auth.jotti.rocks` → Proxy auf die
  acme-dns-API; eigene `limit_req`-Zone mit strengem Limit auf `/register`;
  `rocks-init.sh` nimmt `auth.jotti.rocks` in DNS-Check und
  certbot-Zertifikat auf.
- **DNS-Hoster (manuelle, dokumentierte Schritte):** `dns.jotti.rocks A
  <VPS-IP>`, NS-Delegationen für `lokal.jotti.rocks` und `auth.jotti.rocks`
  auf `dns.jotti.rocks`, CAA-Record `jotti.rocks → letsencrypt.org`. Die
  Schritte landen im Betreiber-Leitfaden (Betriebsdoku), inkl.
  CT-Monitoring-Routine (z. B. crt.sh auf `lokal.jotti.rocks`).
- **End-to-End-Verifikation** (manuell, dokumentiert): von außerhalb
  `dig 10-0-0-1.test.lokal.jotti.rocks`; Registrierung per
  `curl https://auth.jotti.rocks/register`; TXT-Update; `dig TXT
  <id>.auth.jotti.rocks`; abschließend ein Test-Wildcard-Zertifikat für
  `*.<test-id>.lokal.jotti.rocks` über die LE-Staging-CA (z. B. acme.sh mit
  DNS-01) — der Beweis, dass der komplette DNS-01-Pfad steht.

### Acceptance criteria

- [x] Öffentliche Auflösung: A-Record und `_acme-challenge`-CNAME für
      beliebige Install-IDs funktionieren von außerhalb des VPS.
- [x] `POST https://auth.jotti.rocks/register` liefert Credentials;
      TXT-Updates nur mit diesen Credentials; `/register` ist rate-limitiert.
- [x] CAA-Record aktiv; nur Let's Encrypt darf für `jotti.rocks` ausstellen.
- [x] Ein Staging-Wildcard-Zertifikat für `*.<test-id>.lokal.jotti.rocks`
      wurde erfolgreich über DNS-01 ausgestellt (dokumentierter Testlauf).
- [x] Deploy-Schritte und DNS-Hoster-Einträge sind im Betreiber-Leitfaden
      dokumentiert (inkl. CT-Monitoring-Routine).

> Phase abgeschlossen 2026-06-13: VPS deployed, netcup-DNS-Einträge gesetzt,
> komplette End-to-End-Verifikation nach Leitfaden Abschnitt 5 durchgeführt.
> Staging-Testlauf: acme.sh (`--server letsencrypt_test`, dns_acmedns) stellte
> `*.70fea5d5-3136-43e5-9024-9ef086bd971f.lokal.jotti.rocks` erfolgreich über
> DNS-01 aus.

---

## Phase 3: Caddy ersetzt nginx im lokalen Stack (Fallback-Pfad)

**User stories**: keine direkt (technisches Fundament; erhält das
Option-2-Verhalten)

### Context

- `docker-compose.local.yml:101-119, 128-130` — zu ersetzender
  reverse-proxy-Service und Volumes.
- `reverse-proxy/nginx.local.conf` — komplette Proxy-Semantik (Header,
  Prefix-Strip, Redirect), die Caddy übernehmen muss.
- `reverse-proxy/local-entrypoint.sh` — entfällt; LAN-IP-Logik (`:13-27`)
  dient als Referenz für später (Phase 4).
- `Makefile:152-162` — local-Targets (Kommentartexte erwähnen das
  selbstsignierte Zertifikat).

### What to build

Der lokale Stack läuft auf Caddy, verhält sich aber noch wie Option 2:

- **Custom-Caddy-Image:** Dockerfile unter `reverse-proxy/` mit xcaddy-Build
  (gepinnte Caddy- und Modul-Version inkl. `dns.providers.acmedns` — schon
  jetzt eingebaut, damit Phase 4 nur Konfiguration ändert).
- **Caddyfile (in dieser Phase noch statisch):** HTTP→HTTPS-Redirect,
  Catch-all-HTTPS-Site mit Caddys interner CA (Fallback `https://<LAN-IP>`,
  einmalige Warnung wie bisher), `/api/*` → backend:3000 mit Prefix-Strip,
  alles andere → frontend:80, Security-Header aus `nginx.local.conf`
  übernommen.
- **Compose-Umbau:** reverse-proxy-Service baut das neue Image;
  `tls-certs`-Volume und Entrypoint-Mounts entfallen, dafür ein Volume für
  Caddys Datenverzeichnis. `nginx.local.conf` und `local-entrypoint.sh`
  werden gelöscht; Makefile-Kommentare angepasst. prod/rocks bleiben
  unberührt.

### Acceptance criteria

- [x] `make local-up` startet den Stack mit Caddy; `https://<LAN-IP>` zeigt
      die App nach einmaliger Zertifikatswarnung (interner CA), Login und
      API-Aufrufe funktionieren.
- [x] HTTP auf Port 80 leitet auf HTTPS um.
- [x] Security-Header (CSP, HSTS, X-Content-Type-Options, …) sind äquivalent
      zu `nginx.local.conf` gesetzt.
- [x] `reverse-proxy/local-entrypoint.sh` und
      `reverse-proxy/nginx.local.conf` existieren nicht mehr; kein
      openssl-Schritt im Stack.
- [x] prod-/rocks-Deployments sind unverändert (nginx).

> Phase abgeschlossen 2026-06-13: Caddy 2.11.4 + acmedns v0.7.0 (gepinnt, xcaddy).
> Fallback-Site als Catch-all mit interner CA: `on_demand` stellt beim ersten
> Handshake aus; `sign_with_root` + `lifetime 365d` halten die Warnung einmalig
> (Parität zum alten 365-Tage-openssl-Zertifikat). Verifiziert per curl/openssl
> gegen die LAN-IP: Redirect 301, Frontend, `/api/health`, Login-Validierung,
> alle sechs Security-Header identisch, identisches Zertifikat nach
> Proxy-Restart (caddy-data-Volume). `make check` grün.

---

## Phase 4: Install-ID + grünes Zertifikat end-to-end

**User stories**: 1, 3, 4, 5

### Context

- Phase 2 — laufende Infra (`auth.jotti.rocks`-API, Resolver).
- Phase 3 — Custom-Caddy-Image mit acmedns-Modul.
- `reverse-proxy/local-entrypoint.sh:13-27` (gelöscht in Phase 3, Git-History)
  — Referenz für die LAN-IP-Erkennung über die Default-Route im Container.
- PRD „Hostname-Schema" + „Lokaler Stack" — verbindliche Form von Hostname,
  Registrierung und Startablauf.

### What to build

Das Hilfsprogramm (Go-Modul unter `reverse-proxy/`, im Caddy-Image
mitkompiliert) wird Entrypoint des Caddy-Containers:

1. **State sicherstellen:** Installations-State (Install-ID +
   acme-dns-Credentials als JSON) aus dem State-Volume lesen; fehlt er,
   einmalig `POST https://auth.jotti.rocks/register` und Antwort persistieren.
   Install-ID = registrierte Subdomain. Idempotent über Neustarts; keine
   personenbezogenen Daten.
2. **LAN-IP erkennen** (Default-Route-Interface, wie bisheriges
   Entrypoint-Skript).
3. **Hostname ableiten** — reine Funktion LAN-IP + Install-ID → Hostname.
4. **Caddyfile rendern:** Wildcard-Site `*.<install-id>.lokal.jotti.rocks`
   mit `tls { dns acmedns … }` (Credentials aus dem State) + die
   Fallback-Site aus Phase 3. Env-Schalter für die LE-Staging-CA
   (Entwicklung/Tests).
5. **Caddy als Kindprozess starten**, Signale durchreichen, Exit-Status
   spiegeln.

Ausstellung und Erneuerung übernimmt Caddy asynchron — der Stack ist sofort
über den Fallback nutzbar, auch offline oder vor der ersten Ausstellung
(Startablauf-Normalfall laut PRD). Ohne State und ohne Internet startet nur
die Fallback-Site; die Registrierung wird beim nächsten Start nachgeholt.

Tests gemäß PRD: Hostname-Ableitung (Unit, rein), State-Handling
(Unit, injizierte Dateizugriffe: vorhandener State wird nie überschrieben),
Registrierungs-Client gegen einen Test-HTTP-Server. Ausstellung/Erneuerung
selbst bleibt bewusst ungetestet (Integrations-/Betriebsebene).

### Acceptance criteria

- [ ] Erster Start registriert genau einmal bei acme-dns und persistiert den
      State; jeder weitere Start verwendet ihn unverändert.
- [ ] Ein Smartphone im WLAN öffnet
      `https://<ip-mit-bindestrichen>.<id>.lokal.jotti.rocks` mit grünem
      Schloss — ohne Warnung, ohne Geräte-Einrichtung (verifiziert mit echtem
      LE-Zertifikat).
- [ ] Fallback `https://<LAN-IP>` funktioniert parallel weiter (einmalige
      Warnung).
- [ ] DHCP-IP-Wechsel: neuer Hostname nach Neustart des Proxys, dasselbe
      Wildcard-Zertifikat, keine Neuausstellung.
- [ ] Start ohne Internet: Stack läuft über den Fallback; keine Crash-Loops
      durch fehlgeschlagene ACME-Versuche.
- [ ] Unit-Tests für Hostname-Ableitung, State-Idempotenz und
      Registrierungs-Client grün.

---

## Phase 5: Status-Seite — Start-Zustandslogik, QR-Code, Rebind-Erkennung

**User stories**: 2, 3, 6

### Context

- Phase 4 — Hilfsprogramm mit State, Hostname und laufendem Caddy.
- PRD „Lokaler Stack" (Startablauf, Rebind-Erkennung) + „Testing Decisions"
  (Start-Zustandslogik als Unit-Test).
- `Makefile:152-154` — `local-up` gibt künftig die Status-URL aus.

### What to build

Das Hilfsprogramm serviert zusätzlich die Status-Seite auf
`http://localhost:8484` (im Compose nur an `127.0.0.1` publiziert):

- **Reine Start-Zustandslogik:** Eingaben (Zertifikat vorhanden/gültig?,
  Rebind-Check-Ergebnis, Registrierung vorhanden?) → Anzeige-Entscheidung
  (welche Adresse(n) primär, welcher Hinweis). Unit-getestet über alle
  Zustände: kein Zertifikat / gültig / abgelaufen / Rebind blockiert.
- **Zertifikats-Probe:** TLS-Handshake gegen den eigenen Caddy mit dem
  abgeleiteten Hostnamen als SNI; gültige Let's-Encrypt-Kette ⇒ „grün".
- **Rebind-Check beim Start:** abgeleiteten Hostnamen über den
  System-Resolver auflösen; Ergebnis ≠ eigene LAN-IP ⇒ „blockiert" mit
  Hinweis und Link auf die Router-Anleitung (Phase 6).
- **Status-Seite (HTML, selbst aktualisierend bis „grün"):** Zustand,
  Fallback-Adresse sofort, nach Ausstellung die grüne Adresse prominent mit
  QR-Code (Go-QR-Bibliothek, als Bild eingebettet), WLAN-Hinweis
  (Vereins-WLAN, kein Gastnetz), bei Blockade die Rebind-Anleitung.
- **Konsolen-Verweis:** `make local-up` gibt nach dem Start die Status-URL
  aus („Status & Zugangsadresse: http://localhost:8484").

### Acceptance criteria

- [ ] Direkt nach dem ersten Start zeigt die Status-Seite die
      Fallback-Adresse; nach erfolgreicher Ausstellung wechselt sie ohne
      manuelles Neuladen auf die grüne Adresse mit QR-Code.
- [ ] Der QR-Code öffnet auf einem Smartphone die grüne Adresse.
- [ ] Simulierte Rebind-Blockade (Hostname löst nicht zur LAN-IP auf) zeigt
      den Hinweis samt Anleitung-Link und die Fallback-Adresse.
- [ ] Status-Seite ist nur vom Kassen-Rechner erreichbar
      (`127.0.0.1`-Binding), nicht aus dem WLAN.
- [ ] Start-Zustandslogik ist über alle vier Zustände unit-getestet.
- [ ] `make local-up` nennt die Status-URL.

---

## Phase 6: Anleitungen + Doku-Ripples

**User stories**: 2, 6

### Context

- PRD „Anleitungen für Vereinsmitglieder" — verbindlicher Lieferumfang.
- `docs/betrieb/leitfaden-hosting.md:42-128` — Weg A.
- `docs/anforderungen.md:143` — Q-06.
- `docs/prds/prd-lokale-tls-vertrauenswuerdig.md:9-10` — Status-Hinweis „kein
  Umsetzungsplan" (überholt).
- `docs/plans/plan-windows-verpackung.md:203` — ZIP-Inhalt mit nginx-Dateien.

### What to build

- **Fritz!Box-Verifikation an echter Hardware** (zuerst, bestimmt die
  Anleitung): Deckt der Rebind-Ausnahme-Eintrag `lokal.jotti.rocks` auch
  Subdomains ab? Ergebnis im Plan/der Anleitung festhalten.
- **Fritz!Box-Anleitung** mit Screenshots (Heimnetz → Netzwerk →
  Netzwerkeinstellungen → DNS-Rebind-Schutz → Ausnahme), als eigenes Dokument
  unter `docs/betrieb/`; die Status-Seite (Phase 5) verlinkt genau hierauf.
- **Generische Router-Tabelle:** Pi-hole/dnsmasq
  (`rebind-domain-ok=/lokal.jotti.rocks/`), OpenWrt, Speedport.
- **WLAN-/Diagnose-Hinweise:** Handy muss ins Vereins-WLAN (nicht Mobilfunk),
  kein Gastnetz (Client-Isolation), DoH/DoT-Symptom („geht auf Handy A, nicht
  auf Handy B") — auf Status-Seite knapp, in der Anleitung ausführlich.
- **`docs/betrieb/leitfaden-hosting.md` (Weg A):** grünes Schloss als
  Normalfall, QR-Code-Ablauf über die Status-Seite, Fallback-Pfad mit
  Restrisiko-Hinweis; Sicherheitshinweis-Kasten aktualisieren.
- **`docs/anforderungen.md` Q-06:** lokalen Betrieb auf „vertrauenswürdiges
  Zertifikat via lokal.jotti.rocks, selbstsignierter Fallback" umstellen.
- **PRD-Status aktualisieren** (Plan existiert, Betriebszusage bestätigt,
  offene Fragen entschieden → Verweis auf diesen Plan).
- **Touchpoint im Windows-Plan dokumentieren:** Hinweis-Absatz in
  `docs/plans/plan-windows-verpackung.md`, dass Option 3 den ZIP-Inhalt
  (Caddy statt nginx-Conf + Entrypoint-Skript, Caddy-Image via GHCR) und die
  Starter-Erfolgsausgabe (Verweis auf `http://localhost:8484`) ändert.

### Acceptance criteria

- [ ] Fritz!Box-Frage an echter Hardware verifiziert und das Ergebnis in der
      Anleitung umgesetzt (einmaliger Eintrag vs. Eintrag je Hostname).
- [ ] Fritz!Box-Anleitung mit Screenshots existiert; Status-Seite verlinkt
      sie.
- [ ] Router-Tabelle und WLAN-/Gastnetz-/DoH-Hinweise sind Teil der
      Betriebsdoku.
- [ ] Weg A im Hosting-Leitfaden beschreibt grünes Schloss, QR-Ablauf und
      Fallback inkl. Restrisiko.
- [ ] Q-06, PRD-Status und Windows-Plan-Touchpoint sind aktualisiert.
