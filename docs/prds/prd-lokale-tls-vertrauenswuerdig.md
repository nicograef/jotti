# PRD: Lokale Transportverschlüsselung — vertrauenswürdiges Zertifikat (Option 3, Ziel)

> Referenz-Anforderung: Q-06 (`docs/anforderungen.md`) — vertrauenswürdiges TLS
> für den lokalen Betrieb.
> Vorgänger: `docs/prds/prd-lokale-tls-selbstsigniert.md` (Option 2, Interim) —
> Option 3 löst dessen Restrisiko (aktiver MITM) auf.
> Datenschutz-Bezug: `docs/lizenz-und-nutzung.md` §7 (Art. 32 DSGVO).
> **Status: Ziel-Architektur, spätere Umsetzung.** Es gibt noch keinen
> Umsetzungsplan.
> **Voraussetzung:** jotti betreibt eine projekteigene Domain + einen kleinen
> DNS-Dienst (Basis: `jotti.rocks`) — siehe „Open Questions".

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
  öffentliche CA aus.
- Eine **eigene CA auf 30 BYOD-Handys** auszurollen (iOS zweistufig) ist für
  Ehrenamtliche unpraktikabel.

Gesucht ist also ein öffentlich vertrauenswürdiges Zertifikat für einen **lokal
laufenden jotti auf einer privaten LAN-IP**, **ohne Einrichtung pro Endgerät**.

## Solution

Ein bewährtes Muster (Plex `*.plex.direct`, sslip.io/nip.io) löst genau das:

1. **IP-codierter Hostname statt nackter IP.** Das Smartphone öffnet
   `https://192-168-1-50.lokal.jotti.rocks` statt `https://192.168.1.50`. Ein
   winziger, **zustandsloser** DNS-Dienst (Open-Source, sslip.io-artig, von jotti
   unter `jotti.rocks` betrieben) löst jeden solchen Namen auf die eingebettete
   IP auf — **auch private** Adressen. Kein DNS-Eintrag pro Verein nötig.
2. **Echtes Let's-Encrypt-Zertifikat via DNS-01.** Der lokale jotti holt für
   seinen Hostnamen ein Zertifikat über die **DNS-01-Challenge**. DNS-01 benötigt
   **keine eingehende** Erreichbarkeit des Rechners, nur ausgehendes Internet +
   Kontrolle über die DNS-Zone (die hat jotti). HTTP-01 funktioniert hier
   **nicht** (bräuchte Port 80 aus dem Internet auf die private IP).
3. **Caddy als lokaler Reverse-Proxy** übernimmt Ausstellung **und** automatische
   Erneuerung mit minimaler Konfiguration (DNS-01-Plugin) und ersetzt im lokalen
   Stack das hand-verdrahtete nginx+certbot.

Ergebnis für den Verein: weiterhin „zwei Doppelklicks". Beim ersten Start zieht
Caddy das Zertifikat; das Smartphone öffnet die angezeigte
`…lokal.jotti.rocks`-Adresse (oder scannt einen **QR-Code**) → **grünes Schloss,
jedes Gerät, kein CA-Rollout, keine Warnung**. Der unhandliche Hostname
verschwindet hinter QR-Code / „Zum Startbildschirm hinzufügen".

**Aktiver MITM ist damit ausgeschlossen:** Der Angreifer bekommt für
`…lokal.jotti.rocks` **kein** gültiges Zertifikat → sein Fake erzeugt einen
**harten** Fehler, und weil der legitime Weg gar keine Warnung hat, entsteht auch
keine „Wegklick"-Gewöhnung.

## User Stories

1. Als Vereins-Admin möchte ich ein echtes, vom Browser akzeptiertes Zertifikat
   (grünes Schloss) im lokalen WLAN, ohne eine Domain zu kaufen oder ein
   Zertifikat zu installieren.
2. Als Service-Helfer möchte ich die Zugriffsadresse als QR-Code scannen und ohne
   jede Sicherheitswarnung sofort arbeiten.
3. Als Vereins-Admin möchte ich, dass das Zertifikat automatisch ausgestellt und
   erneuert wird, ohne dass ich etwas tun muss.
4. Als datenschutzverantwortliche Organisation möchte ich, dass auch ein aktiver
   Angreifer im selben WLAN keine Sitzung übernehmen kann, damit Art. 32 DSGVO
   tatsächlich erfüllt ist.
5. Als jotti-Projektbetreiber möchte ich eng begrenzte DNS-Credentials
   verwenden, damit ein kompromittierter lokaler Rechner nicht die ganze
   DNS-Zone gefährdet.

## Implementation Decisions

### Von jotti betriebene Infrastruktur

- Eine Domain (z. B. `lokal.jotti.rocks`) plus ein winziger, **zustandsloser**
  sslip.io-artiger DNS-Server (Binary auf einem kleinen VPS), der
  IP-codierte Hostnamen — inklusive privater IPs — auf die eingebettete Adresse
  auflöst.
- Der DNS-01-Pfad nutzt **acme-dns** (oder gleichwertig) für **eng begrenzte
  Credentials** — die von Let's Encrypt empfohlene Best Practice, statt volle
  DNS-API-Schlüssel auf den lokalen Rechner zu legen.

### Lokaler Stack

- **Caddy** als Reverse-Proxy mit DNS-01-Plugin gegen den jotti-DNS/acme-dns;
  automatische Ausstellung + Erneuerung.
- **Hostname-Ableitung:** Aus der aktuellen LAN-IP wird beim Start der Hostname
  `192-168-1-50.lokal.jotti.rocks` gebildet und angezeigt (+ QR-Code). Diese
  Ableitung ist eine kleine, reine Funktion (testbar).

### Bekannte Einschränkung

- **DNS-Rebinding-Schutz:** Manche Router/ISP-DNS blockieren „öffentlicher Name →
  private IP" (Plex dokumentiert dafür `rebind-domain-ok=/plex.direct/`). Es
  braucht einen dokumentierten Workaround/Fallback (z. B. Rückfall auf das
  Option-2-Zertifikat oder eine Router-Konfigurationsanleitung).
- Die Erneuerung benötigt periodisch Internet (ohnehin vorhanden; die TSE braucht
  es ebenfalls).

## Testing Decisions

- **Hostname-Ableitung** (Unit-Test, rein): LAN-IP → korrekter
  `…lokal.jotti.rocks`-Hostname.
- **DNS-Auflösung** (Integration): eingebettete IP → genau diese IP, auch privat.
- **Nicht unit-getestet:** Ausstellung/Erneuerung (Caddy + acme-dns) ist
  Integrations-/Betriebsebene.

## Out of Scope

- **Option 2** (selbstsigniert) — Vorgänger/Geschwister
  (`docs/prds/prd-lokale-tls-selbstsigniert.md`).
- **prod** — hat bereits echtes TLS via Let's Encrypt.
- Die klickbare Windows-Verpackung selbst (konsumiert dieses TLS nur).
- Änderungen an POST-only, Event-Sourcing oder Datenmodell.

## Rejected Alternatives

- **Lokale CA / mkcert:** echte Authentizität, aber Root-Zertifikat auf jedes
  BYOD-Handy ausrollen (iOS zweistufig) — für Ehrenamtliche unpraktikabel.
- **Tailscale/WireGuard-Overlay:** kryptografisch stark (LE-Certs für `*.ts.net`
  via DNS-01), aber App-Installation + Login je Handy — zu viel Reibung für BYOD.

## Open Questions

- **Infrastruktur-Zusage:** Option 3 setzt voraus, dass das jotti-Projekt eine
  Domain (Basis `jotti.rocks`) **und** den zustandslosen DNS-Dienst dauerhaft
  betreibt. Diese Betriebsverpflichtung ist vor der Umsetzung zu bestätigen.
- **Prior Art (verifiziert):** Plex `*.plex.direct`, sslip.io/nip.io;
  Let's Encrypt bestätigt, dass DNS-01 auch für nicht öffentlich erreichbare
  Webserver funktioniert.
