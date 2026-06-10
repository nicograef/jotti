# PRD: Lokale Transportverschlüsselung — selbstsigniertes TLS (Option 2, Interim)

> Referenz-Anforderung: Q-06 (`docs/anforderungen.md`) — erweitert TLS auf den
> lokalen Betrieb.
> Verwandt: `docs/prds/prd-lokale-tls-vertrauenswuerdig.md` (Option 3,
> Ziel-Architektur), `docs/prds/prd-windows-verpackung.md`,
> `docs/betrieb/leitfaden-hosting.md` (Weg A).
> Datenschutz-Bezug: `docs/lizenz-und-nutzung.md` §7 (Art. 32 DSGVO,
> technische Maßnahmen).
> **Status: Interim-Lösung.** Schließt nur die passive Mitlese-Lücke. Den aktiven
> Angreifer im selben WLAN schließt erst Option 3 (siehe „Sicherheitsmodell").

## Problem Statement

Der lokale LAN-Betrieb (Weg A im Hosting-Leitfaden, auch der Windows-Starter-Weg)
serviert heute **unverschlüsseltes HTTP**. In einem Vereins-WLAN kennen viele
Personen das WLAN-Passwort — ein Angreifer befindet sich also mit geringer Hürde
im selben Netz.

- Beim Login wird das **Klartext-Passwort** übertragen (das Backend hasht erst
  dort mit Argon2id, `docs/handbuch.md`). Danach reist in **jedem** Request das
  **JWT** mit.
- Über HTTP sind beide für jeden im WLAN **passiv mitlesbar** und per
  ARP-Spoofing aktiv manipulierbar.
- Das ist primär ein **DSGVO-Art.-32-Problem** (Sicherheit der Verarbeitung,
  technische Maßnahmen), kein KassenSichV/TSE-Problem: Die TSE sichert die
  Integrität des Kassenjournals, nicht die Transportverschlüsselung.
- **Q-06** („HTTPS/TLS") ist heute nur für den prod-Weg umgesetzt (nginx +
  Let's Encrypt). Der lokale Weg hat **keine** Transportverschlüsselung.

Das Risiko entsteht nur beim **Mehrgeräte-Betrieb** (Helfer-Smartphones im WLAN).
Eine einzelne Theke direkt am Rechner (`localhost`) überträgt nichts über das
Netz und ist unkritisch.

## Solution

Der lokale Reverse-Proxy serviert **HTTPS mit einem automatisch erzeugten
selbstsignierten Zertifikat**:

- Das Zertifikat wird beim ersten Start erzeugt und danach wiederverwendet
  (idempotent). HTTP wird auf HTTPS umgeleitet.
- Der gesamte Verkehr zwischen Helfer-Smartphones und dem Host ist damit
  **verschlüsselt** — passives Mitlesen von Passwörtern und JWTs ist
  ausgeschlossen.
- Eine **einmalige Browserwarnung** („Verbindung nicht sicher") pro Gerät wird
  als bewusster Kompromiss akzeptiert. **PWA / „Zum Startbildschirm hinzufügen"
  ist für diese Option ausdrücklich nicht erforderlich.**

Für den Verein ändert sich der Ablauf kaum: Der Starter zeigt die
`https://`-Adresse an; beim ersten Zugriff bestätigt jeder Helfer einmal die
Ausnahme.

## Sicherheitsmodell (verbindlich, ehrlich)

TLS leistet zwei Dinge: **Verschlüsselung** (niemand liest mit) und
**Authentizität** (man redet wirklich mit jottis Server). Die Authentizität kommt
allein aus der Zertifikatskette zu einer vertrauenswürdigen CA.

- Ein selbstsigniertes Zertifikat hat **keinen Vertrauensanker** — niemand
  Vertrauenswürdiges bürgt dafür. Daher die Browserwarnung.
- **Aktiver MITM bleibt möglich:** Ein Angreifer im selben WLAN (kennt das
  WLAN-Passwort, also genau das hier relevante Bedrohungsmodell) kann per
  ARP-Spoofing / Rogue-DHCP / Evil-Twin dazwischengehen und **sein eigenes**
  selbstsigniertes Zertifikat präsentieren. Das Smartphone zeigt **dieselbe**
  Warnung, die der Helfer ohnehin wegklickt → die TLS-Sitzung besteht mit dem
  Angreifer, Passwort und JWT sind offengelegt.
- Auch der „beim ersten Mal merken"-Effekt trägt nicht: War schon die **erste**
  Verbindung über den Angreifer, merkt sich das Gerät dessen Zertifikat.

**Fazit:** Selbstsigniertes TLS schützt **nur gegen passives Mitlesen**, nicht
gegen den aktiven Angreifer im selben Netz. Eine vollständige Art.-32-Absicherung
gegen diesen aktiven Angreifer erfordert ein **öffentlich vertrauenswürdiges
Zertifikat** → `docs/prds/prd-lokale-tls-vertrauenswuerdig.md` (Option 3). Diese
PRD ist ein bewusster **Zwischenschritt** (Klartext → verschlüsselt) auf dem Weg
dorthin.

## User Stories

1. Als Vereins-Admin möchte ich, dass die Verbindung im WLAN verschlüsselt ist,
   damit Passwörter und Sitzungs-Token nicht im Klartext mitlesbar sind — ohne
   eine Domain kaufen oder ein Zertifikat installieren zu müssen.
2. Als Service-Helfer möchte ich beim ersten Zugriff die einmalige
   Sicherheitswarnung bestätigen und danach normal arbeiten können.
3. Als Vereins-Admin möchte ich, dass das selbstsignierte Zertifikat automatisch
   erzeugt und über mehrere Festtage wiederverwendet wird, ohne dass ich etwas
   tun muss.
4. Als datenschutzverantwortliche Organisation möchte ich das verbleibende
   Restrisiko (aktiver MITM) klar dokumentiert sehen, damit ich eine informierte
   Entscheidung über den Interim-Einsatz treffen und Option 3 einplanen kann.
5. Als Betreiber einer Einzeltheke direkt am Rechner möchte ich wissen, dass für
   den reinen `localhost`-Betrieb keine Transportverschlüsselung nötig ist.

## Implementation Decisions

### Zertifikatserzeugung im lokalen Stack (plattformunabhängig)

- Die Erzeugung des selbstsignierten Zertifikats liegt **im lokalen Stack**
  (Init-Schritt/Entrypoint des Reverse-Proxy), nicht im Windows-Starter. So
  funktioniert sie identisch auf Windows, macOS und Linux; der Windows-Starter
  zeigt nur die `https://`-Adresse an.
- Erzeugung ist **idempotent**: vorhandenes Zertifikat wird wiederverwendet,
  fehlendes neu erzeugt und in einem Volume abgelegt.
- Die **LAN-IP des Hosts** gehört als SAN ins Zertifikat (plus `localhost`),
  damit es zur vom Smartphone genutzten Adresse passt.

### Reverse-Proxy

- Der lokale Reverse-Proxy serviert HTTPS (Port 443) mit dem selbstsignierten
  Zertifikat und leitet HTTP auf HTTPS um.
- Die angezeigte Zugriffsadresse wird `https://<LAN-IP>`.

### Ripple-Updates (im Umsetzungsplan dieser PRD, nicht in dieser PRD)

- **Q-06** in `docs/anforderungen.md`: TLS-Beschreibung auf den lokalen Betrieb
  ausweiten (selbstsigniert, Interim).
- **Weg A** in `docs/betrieb/leitfaden-hosting.md`: nicht mehr „unverschlüsseltes
  HTTP", sondern selbstsigniertes HTTPS + einmalige Warnung + Restrisiko-Hinweis
  - Verweis auf Option 3.
- **`docs/prds/prd-windows-verpackung.md`**: zeigt bereits `https://` und den
  Warnungs-Hinweis (konsistent gehalten).

## Testing Decisions

- **Zertifikatserzeugung** (idempotent): vorhandenes Zertifikat wird
  wiederverwendet; fehlendes wird mit korrektem SAN (LAN-IP + `localhost`)
  erzeugt.
- **Reverse-Proxy**: serviert HTTPS auf 443; HTTP → HTTPS-Redirect greift.
- **Nicht getestet (architektonisch dokumentiert):** Die Restrisiko-Eigenschaft
  (aktiver MITM trotz Verschlüsselung) ist eine Eigenschaft des
  Vertrauensmodells, kein testbares Verhalten.

## Out of Scope

- **Vertrauenswürdiges/öffentliches Zertifikat und IP-Hostname-DNS** → Option 3
  (`docs/prds/prd-lokale-tls-vertrauenswuerdig.md`).
- **prod** — hat bereits echtes TLS via Let's Encrypt (Q-06).
- Ausrollen einer eigenen CA auf die Endgeräte.
- PWA-/Service-Worker-/„Secure-Context"-Funktionen.
- Änderungen an POST-only, Event-Sourcing oder Datenmodell.

## Further Notes

- Dies ist eine **Interim-Minderung**; die compliance-vollständige Antwort ist
  Option 3. Die Reihenfolge (zuerst Option 2, danach Option 3) ist bewusst.
- Der Einzeltheken-Fall direkt am Rechner (`localhost`) braucht kein TLS, weil
  kein Verkehr über das Netz geht.
