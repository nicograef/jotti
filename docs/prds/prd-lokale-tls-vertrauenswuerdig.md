# PRD: Lokale Transportverschlüsselung — vertrauenswürdiges Zertifikat (Option 3, Ziel)

> Referenz-Anforderung: Q-06 (`docs/anforderungen.md`) — vertrauenswürdiges TLS
> für den lokalen Betrieb.
> Vorgänger: `docs/prds/prd-lokale-tls-selbstsigniert.md` (Option 2, Interim) —
> Option 3 löst dessen Restrisiko (aktiver MITM) auf und nimmt dessen
> Mechanismus als eingebauten Fallback auf.
> Datenschutz-Bezug: `docs/lizenz-und-nutzung.md` §7 (Art. 32 DSGVO).
> **Status: Ziel-Architektur, spätere Umsetzung.** Es gibt noch keinen
> Umsetzungsplan.
> **Voraussetzung:** jotti betreibt eine projekteigene Domain (`jotti.rocks`),
> einen kleinen zustandslosen DNS-Resolver und eine acme-dns-Instanz dauerhaft
> — siehe „Open Questions".

## Problem Statement

Option 2 (selbstsigniertes TLS) verschlüsselt zwar, ihr Zertifikat hat aber
**keinen Vertrauensanker**. Ein aktiver Angreifer im selben WLAN kann deshalb
weiterhin per MITM dazwischengehen, indem er ein eigenes selbstsigniertes
Zertifikat präsentiert — der Helfer klickt dieselbe Warnung weg wie sonst auch.

Für eine **vollständige Art.-32-Absicherung gegen genau diesen aktiven
Angreifer** — das reale Vereins-WLAN-Bedrohungsmodell — braucht es ein
**öffentlich vertrauenswürdiges Zertifikat**. Die Hürden:

- Der Verein hat **keine eigene Domain** und soll keine kaufen/pflegen müssen.
- Ein Zertifikat auf eine nackte private LAN-IP (`192.168.x.x`) stellt keine
  öffentliche CA aus (auch Let's Encrypts IP-Zertifikate gelten nur für
  öffentliche IPs — siehe „Rejected Alternatives").
- Eine **eigene CA auf 30 BYOD-Handys** auszurollen (iOS zweistufig) ist für
  Ehrenamtliche unpraktikabel.

Gesucht ist also ein öffentlich vertrauenswürdiges Zertifikat für einen **lokal
laufenden jotti auf einer privaten LAN-IP**, **ohne Einrichtung pro Endgerät**.

## Solution

Ein bewährtes Muster (Plex `*.plex.direct`, sslip.io/nip.io), erweitert um eine
**Install-ID**, löst genau das:

1. **IP-codierter Hostname mit Install-ID statt nackter IP.** Das Smartphone
   öffnet `https://192-168-1-50.<install-id>.lokal.jotti.rocks` statt
   `https://192.168.1.50`. Ein winziger, **zustandsloser** DNS-Resolver (von
   jotti unter `jotti.rocks` betrieben) löst jeden solchen Namen rein
   rechnerisch auf die eingebettete IP auf — **auch private** Adressen. Kein
   DNS-Eintrag pro Verein nötig. Die Install-ID (zufällig, einmalig pro
   Installation) macht den Namen weltweit eindeutig — warum das
   sicherheitskritisch ist: siehe „Sicherheitsmodell".
2. **Ein Wildcard-Zertifikat je Installation via DNS-01.** Der lokale jotti
   holt ein echtes Let's-Encrypt-Zertifikat für
   `*.<install-id>.lokal.jotti.rocks` über die **DNS-01-Challenge**. DNS-01
   benötigt **keine eingehende** Erreichbarkeit des Rechners, nur ausgehendes
   Internet; die Challenge wird über **acme-dns** mit eng begrenzten, an die
   Install-ID gebundenen Credentials beantwortet. HTTP-01 funktioniert hier
   **nicht** (bräuchte Port 80 aus dem Internet auf die private IP). Das
   Wildcard deckt jede LAN-IP ab — ein IP-Wechsel (DHCP) braucht **keine**
   Neuausstellung, und Erneuerungen desselben Namens sind von den
   Let's-Encrypt-Rate-Limits **ausgenommen**.
3. **Caddy als lokaler Reverse-Proxy** übernimmt Ausstellung **und**
   automatische Erneuerung (offizielles Modul `dns.providers.acmedns`) und
   serviert zusätzlich `https://<LAN-IP>` mit einem selbstsignierten
   Zertifikat als eingebauten Option-2-Fallback. Er ersetzt im lokalen Stack
   das hand-verdrahtete nginx-Setup samt openssl-Entrypoint-Skript.

Ergebnis für den Verein: weiterhin „zwei Doppelklicks". Beim ersten Start zieht
Caddy das Zertifikat; das Smartphone scannt den angezeigten **QR-Code** (oder
tippt die Adresse) → **grünes Schloss, jedes Gerät, kein CA-Rollout, keine
Warnung**. Der unhandliche Hostname verschwindet hinter QR-Code / „Zum
Startbildschirm hinzufügen".

## Sicherheitsmodell (verbindlich, ehrlich)

- **Warum der aktive MITM scheitert:** Der Angreifer bekommt für den Namen aus
  dem QR-Code **kein** gültiges Zertifikat → sein Fake erzeugt einen **harten**
  Browserfehler (kein Wegklicken möglich), und weil der legitime Weg gar keine
  Warnung hat, entsteht auch keine „Wegklick"-Gewöhnung.
- **Warum die Install-ID zwingend ist:** Ohne sie wären Hostnamen reine
  IP-Codierungen — und private IPs kollidieren zwischen Vereinen massiv (die
  Fritz!Box vergibt standardmäßig `192.168.178.x`; viele Vereine hätten
  **exakt denselben** Hostnamen). Folgen: (a) Let's Encrypts Limit von
  5 Zertifikaten/Woche je identischem Namenssatz wäre nach wenigen Vereinen
  erschöpft; (b) gravierender: In einer offenen, rechnerischen Zone könnte
  **auch ein Angreifer** die DNS-01-Challenge für den Kollisionsnamen bestehen
  und ein **gültiges** Zertifikat für genau den Namen „seines" Opfer-Vereins
  holen — der aktive MITM hätte dann selbst ein grünes Schloss. Die Bindung
  „Challenge nur mit den acme-dns-Credentials dieser Install-ID" schließt
  beides aus.
- **Restrisiken (dokumentiert, nicht Scope dieser PRD):** physisches
  Austauschen des ausgehängten QR-Codes; ein kompromittierter Vereinsrechner;
  eine Kompromittierung der jotti-DNS-Infrastruktur (sie ist der
  Vertrauensanker und entsprechend zu härten und zu betreiben).

## User Stories

1. Als Vereins-Admin möchte ich ein echtes, vom Browser akzeptiertes Zertifikat
   (grünes Schloss) im lokalen WLAN, ohne eine Domain zu kaufen oder ein
   Zertifikat zu installieren.
2. Als Service-Helfer möchte ich die Zugriffsadresse als QR-Code scannen und ohne
   jede Sicherheitswarnung sofort arbeiten.
3. Als Vereins-Admin möchte ich, dass das Zertifikat automatisch ausgestellt und
   erneuert wird, ohne dass ich etwas tun muss — auch wenn die Kasse monatelang
   aus war und das Zertifikat beim Saisonstart abgelaufen ist.
4. Als datenschutzverantwortliche Organisation möchte ich, dass auch ein aktiver
   Angreifer im selben WLAN keine Sitzung übernehmen kann, damit Art. 32 DSGVO
   tatsächlich erfüllt ist.
5. Als jotti-Projektbetreiber möchte ich eng begrenzte, je Installation
   gebundene DNS-Credentials verwenden, damit ein kompromittierter lokaler
   Rechner weder die DNS-Zone noch fremde Vereins-Hostnamen gefährdet.
6. Als Vereins-Admin hinter einem Router mit DNS-Rebind-Schutz (z. B.
   Fritz!Box) möchte ich, dass jotti die Blockade selbst erkennt, mir eine
   Schritt-für-Schritt-Anleitung zeigt und bis dahin eine funktionierende
   Fallback-Adresse anbietet, damit der Verkauf trotzdem starten kann.

## Implementation Decisions

### Hostname-Schema

- `<lan-ip-mit-bindestrichen>.<install-id>.lokal.jotti.rocks`. Die Install-ID
  wird beim ersten Start zufällig erzeugt, lokal persistiert und enthält keine
  personenbezogenen Daten.
- Install-ID = die bei acme-dns registrierte Subdomain (oder daraus
  abgeleitet), damit der Resolver die Challenge-Delegation **rein rechnerisch**
  beantworten kann (s. u.) und selbst vollständig zustandslos bleibt.
- Das Wildcard hält die LAN-IPs zusätzlich aus den öffentlichen
  Certificate-Transparency-Logs heraus — dort erscheint nur
  `*.<install-id>.lokal.jotti.rocks`, nicht die interne Netzstruktur.
- Die Hostname-Ableitung aus LAN-IP + Install-ID ist eine kleine, reine
  Funktion (testbar).

### Von jotti betriebene Infrastruktur

Zwei klar getrennte Komponenten, beide klein:

1. **Zustandsloser Resolver** für `*.lokal.jotti.rocks`: beantwortet A-Records
   aus dem Namen berechnet (auch private IPs) und für
   `_acme-challenge.<install-id>.lokal.jotti.rocks` einen **berechneten**
   CNAME auf `<install-id>.auth.jotti.rocks`. Keine Datenbank. Das ist ein
   **eigenes, kleines Go-Binary** (sslip.io-Muster): der berechnete CNAME ist
   nicht Teil vorhandener sslip.io-artiger Server, und die Kernlogik ist eine
   reine, testbare Funktion.
2. **acme-dns** (Fertigkomponente, aktiv gepflegt) für `auth.jotti.rocks`:
   nimmt TXT-Updates nur mit den bei der Registrierung ausgegebenen
   Credentials an — die von Let's Encrypt empfohlene Best Practice, statt
   volle DNS-API-Schlüssel auf den lokalen Rechner zu legen. Zustand: nur
   Account ↔ Subdomain. Bewusst **kein** eigener Challenge-Dienst: für
   acme-dns existiert das offizielle Caddy-Modul — es ist null eigener
   ACME-Code nötig.

- **Hohe TTL für A-Records** (das Mapping Name → IP ist unveränderlich):
  Geräte, die den Namen einmal aufgelöst haben, überstehen einen
  Internet-Ausfall während des Fests; nur neu dazukommende Geräte brauchen
  dann den Fallback.
- **CAA-Record** auf `jotti.rocks` (nur Let's Encrypt) als billige
  Defense-in-Depth-Maßnahme.
- **Rate Limits sind kein Skalierungsproblem:** Erneuerungen mit identischem
  Namenssatz (bzw. via ARI) sind von Let's Encrypts Limits **ausgenommen**;
  das Limit von 50 neuen Zertifikaten/Woche je registrierter Domain betrifft
  nur **Neu-Installationen** — mehr als 50 neue Vereine pro Woche sind kein
  realistisches Szenario (und falls doch: LE-Erhöhungsformular).

### Lokaler Stack

- **Caddy** (Custom-Build via xcaddy mit dem offiziellen Modul
  `dns.providers.acmedns`) als Reverse-Proxy. Die Site-Adresse im generierten
  Caddyfile ist das Wildcard `*.<install-id>.lokal.jotti.rocks`; Caddy holt
  und erneuert das Zertifikat automatisch.
- Zusätzlich serviert Caddy `https://<LAN-IP>` mit seiner internen CA
  (dokumentiertes Standardverhalten für IP-Adressen) als **eingebauten
  Option-2-Fallback**. Das bisherige openssl-Entrypoint-Skript
  (`reverse-proxy/local-entrypoint.sh`) entfällt.
- **Startablauf behandelt „Zertifikat (noch) nicht da" als Normalfall:** Mit
  den auf 45 Tage sinkenden Laufzeiten (s. u.) ist das Zertifikat beim
  Saisonstart **regelmäßig abgelaufen**. Die Start-Anzeige zeigt deshalb
  sofort die Fallback-Adresse und wechselt auf die grüne
  `…lokal.jotti.rocks`-Adresse (+ QR-Code), sobald Ausstellung/Erneuerung
  abgeschlossen ist (Sekunden bis wenige Minuten, braucht Internet).
- **Rebind-Erkennung beim Start:** Der Stack prüft, ob der eigene Hostname
  über den System-Resolver (= Router) auf die eigene LAN-IP auflöst. Wenn
  nicht: klarer Hinweis in der Start-Anzeige mit Link auf die Router-Anleitung
  und die Fallback-Adresse `https://<LAN-IP>` (einmalige Warnung, wie
  Option 2).

### Anleitungen für Vereinsmitglieder (Teil des Lieferumfangs)

Nicht-technische Helfer sind die Zielgruppe — die Dokumentation ist Teil der
Lösung, nicht Beiwerk:

- **Fritz!Box-Anleitung** (häufigster Router im Vereinsumfeld), mit
  Screenshots: „Heimnetz → Netzwerk → Netzwerkeinstellungen → Erweiterte
  Netzwerkeinstellungen ändern → DNS-Rebind-Schutz → Ausnahme hinzufügen".
  Wildcards (`*.…`) unterstützt die Fritz!Box **nicht**; die Community-Praxis
  zu `plex.direct` legt nahe, dass der Eintrag der Domain
  (`lokal.jotti.rocks`) Subdomains abdeckt — im Umsetzungsplan praktisch zu
  verifizieren (bestimmt, ob die Anleitung einmalig genügt).
- **Generische Router-Hinweise** als kurze Tabelle (Pi-hole/dnsmasq:
  `rebind-domain-ok=/lokal.jotti.rocks/`, OpenWrt, Speedport).
- **WLAN-Hinweise auf QR-Aushang und Anzeige:** Das Handy muss im
  Vereins-WLAN sein (nicht Mobilfunk — der Name löst dort zwar auf, die
  private IP ist aber unerreichbar). **Kein Gastnetz** verwenden: dessen
  Client-Isolation blockiert den Zugriff aufs Kassen-Gerät grundsätzlich
  (betrifft auch Option 2).
- **Diagnose-Hinweis:** Handys mit privatem DNS (DoH/DoT) umgehen den
  Router-Rebind-Schutz — „geht auf Handy A, aber nicht auf Handy B" ist das
  typische Symptom einer Rebind-Blockade.
- **Aktualisierung `docs/betrieb/leitfaden-hosting.md` (Weg A):** grünes
  Schloss als Normalfall, QR-Code-Ablauf, Fallback-Pfad mit
  Restrisiko-Hinweis.
- Die Rebind-Erkennung (s. o.) verlinkt direkt auf diese Anleitung — der
  Admin muss sie nicht suchen.

### Bekannte Einschränkungen

- **DNS-Rebind-Schutz:** Manche Router/ISP-DNS blockieren „öffentlicher Name →
  private IP" (Plex dokumentiert dafür `rebind-domain-ok=/plex.direct/`).
  Antwort dieser PRD: automatische Erkennung + verlinkte Anleitung +
  eingebauter Option-2-Fallback (s. o.).
- **Sinkende Zertifikatslaufzeiten:** Let's Encrypt verkürzt die Laufzeiten in
  den nächsten zwei Jahren schrittweise von 90 über 64 auf **45 Tage**
  (angekündigt 2026-02). Für den Saisonbetrieb heißt das: Beim Start nach
  einer Pause ist das Zertifikat normalerweise abgelaufen → der Startablauf
  (s. o.) braucht Internet und überbrückt mit der Fallback-Adresse. Am Modell
  ändert sich nichts; Erneuerungen bleiben von Rate-Limits ausgenommen.
- Eine offene acme-dns-Registrierung könnten Fremde nutzen, um gültige
  Zertifikate unter `lokal.jotti.rocks` für eigene Zwecke zu holen
  (Reputationsrisiko) → „Open Questions".

## Testing Decisions

- **Hostname-Ableitung** (Unit, rein): LAN-IP + Install-ID → korrekter
  Hostname.
- **Resolver-Logik** (Unit, rein): Name → exakt die eingebettete IP (auch
  privat); `_acme-challenge.<id>…` → korrekt berechneter CNAME; ungültige
  Namen → NXDOMAIN.
- **Start-Zustandslogik** (Unit): kein Zertifikat / gültig / abgelaufen /
  Rebind blockiert → richtige Anzeige-Entscheidung (welche Adresse(n), welcher
  Hinweis).
- **DNS-Auflösung** (Integration): laufender Resolver beantwortet A- und
  CNAME-Anfragen wie spezifiziert.
- **Nicht unit-getestet:** Ausstellung/Erneuerung (Caddy + acme-dns) ist
  Integrations-/Betriebsebene.

## Out of Scope

- **Option 2** (selbstsigniert) als eigenständige PRD
  (`docs/prds/prd-lokale-tls-selbstsigniert.md`) — ihr Mechanismus geht hier
  als eingebauter Fallback auf.
- **prod** — hat bereits echtes TLS via Let's Encrypt.
- Die klickbare Windows-Verpackung selbst (konsumiert dieses TLS nur).
- Änderungen an POST-only, Event-Sourcing oder Datenmodell.

## Rejected Alternatives

- **Lokale CA / mkcert:** echte Authentizität, aber Root-Zertifikat auf jedes
  BYOD-Handy ausrollen (iOS zweistufig) — für Ehrenamtliche unpraktikabel.
- **Tailscale/WireGuard-Overlay:** kryptografisch stark (LE-Certs für `*.ts.net`
  via DNS-01), aber App-Installation + Login je Handy — zu viel Reibung für
  BYOD.
- **Let's-Encrypt-Zertifikat direkt auf die IP:** LE stellt seit 2025
  IP-Zertifikate aus, aber nur für **öffentliche** IPs, nur kurzlebig (~6
  Tage) und nur via HTTP-01/TLS-ALPN — für private LAN-IPs ausgeschlossen
  (verifiziert, Stand 2026-06).
- **mDNS (`jotti.local`):** löst nur die Adresseingabe; keine öffentliche CA
  signiert `.local` → die Browserwarnung bliebe.
- **sslip.io/nip.io direkt nutzen:** geteilte fremde Zone — keine
  Challenge-Bindung je Installation möglich, Let's-Encrypt-Limits treffen alle
  Nutzer der Zone gemeinsam, kein Einfluss auf Betrieb → eigene Zone nötig.
- **Hostname ohne Install-ID (nur IP-Codierung):** Namenskollisionen zwischen
  Vereinen → Duplikat-Rate-Limits **und** gültige Angreifer-Zertifikate
  möglich (siehe „Sicherheitsmodell"). Sicherheitskritisch verworfen.
- **Eigene Challenge-TXT-API statt acme-dns:** spart einen Fertigdienst,
  bräuchte aber ein eigenes Caddy-Plugin und eigenen ACME-Pfad — acme-dns +
  offizielles Caddy-Modul liefern dasselbe ohne eigenen ACME-Code. (Eine
  spätere Zusammenführung von Resolver und acme-dns in ein Binary bleibt als
  Betriebsvereinfachung möglich.)
- **Zentrale Ausstellung nach Plex-Vorbild** (CA-Partnerschaft, zentrale
  CSR-Pipeline): für ein Vereinsprojekt unverhältnismäßig.

## Open Questions

- **Infrastruktur-Zusage:** Option 3 setzt voraus, dass das jotti-Projekt
  Domain, Resolver und acme-dns **dauerhaft** betreibt. Fällt die
  Infrastruktur aus, gibt es kein grünes Schloss mehr (der lokale Fallback
  bleibt funktionsfähig). Diese Betriebsverpflichtung ist vor der Umsetzung zu
  bestätigen.
- **Registrierungs-Gate:** acme-dns-Registrierung offen lassen oder an eine
  jotti-Installation koppeln (z. B. über das Relay-Token aus
  `docs/prds/prd-betrieb-relay-haertung.md`), um Fremdnutzung der Zone zu
  verhindern?
- **Fritz!Box-Ausnahme:** Deckt der Eintrag einer Domain auch deren Subdomains
  ab? Community-Praxis (`plex.direct`) spricht dafür, AVMs Doku formuliert
  „vollständiger Hostname" — im Umsetzungsplan an echter Hardware
  verifizieren; bestimmt die Form der Anleitung.
- **Prior Art und Fakten (verifiziert, Stand 2026-06):** Plex `*.plex.direct`,
  sslip.io/nip.io; Let's Encrypt bestätigt DNS-01 für nicht öffentlich
  erreichbare Webserver; Erneuerungen identischer Namenssätze sind von
  Rate-Limits ausgenommen; Laufzeiten sinken auf 45 Tage; offizielles
  Caddy-Modul `dns.providers.acmedns` vorhanden; acme-dns aktiv gepflegt
  (v1.1, 2024-12); Caddy serviert IP-Adressen automatisch mit interner CA.
