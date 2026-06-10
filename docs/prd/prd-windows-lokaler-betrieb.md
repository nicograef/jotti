# PRD: Einfacher lokaler Windows-Betrieb für Vereine

## 1. Kontext

jotti richtet sich an Vereine mit nicht-technischen ehrenamtlichen Helfern. Für kleine bis mittlere Veranstaltungen soll der Betrieb auf einem vorhandenen Windows-Rechner mit minimalem Einrichtungsaufwand möglich sein, inklusive:

- Backend-Server
- Web-UI für Helfer-Smartphones im gleichen WLAN
- Bondruck über Print-Relay

Ziel ist ein deutlich vereinfachter Betriebsmodus ohne tiefes Docker-, Netzwerk- oder `.env`-Wissen.

## 2. Problem

Der aktuelle Self-Hosting-Weg ist für viele Vereine zu technisch:

- manuelle `.env`-Pflege mit Secrets
- mehrere bewegliche Teile (Container, Reverse-Proxy, Relay)
- unklare Entscheidung zwischen lokalen und serverbasierten Betriebswegen
- hoher Support-Aufwand bei Setup-Fehlern

## 3. Ziele

1. **One-device Betrieb auf Windows vereinfachen** (Host-Rechner + Smartphones im WLAN).
2. **Bondruck verlässlich integrieren**, ohne Fachlogik im Relay.
3. **Konfiguration stark reduzieren** (Assistenz statt manueller `.env`-Bearbeitung).
4. **Betriebssicherheit erhöhen** (Statussicht, klare Fehlerhinweise, einfache Neustarts).

## 4. Nicht-Ziele

- Einführung neuer Domänenfeatures außerhalb des Betriebs-/Deploymentscopes
- Ablösung von PostgreSQL in dieser Initiative
- Internet-/Cloud-Betrieb als Primärziel
- Änderung der API-Regel „POST-only“

## 5. Zielgruppen

- **Primär:** Admin-Verantwortliche im Verein (technisch wenig erfahren)
- **Sekundär:** Helfer im Service auf BYOD-Smartphones

## 6. Anforderungen

### 6.1 Betrieb & Packaging

- Windows-tauglicher Standardweg mit **Docker Desktop** bleibt Basis.
- Bereitstellung eines lokalen Betriebsmodus für WLAN-Nutzung über lokale IP.
- Print-Relay als **separate Windows-Executable** (bevorzugt), unabhängig start-/neustartbar.

### 6.2 Architekturvarianten

- **Variante A (kurzfristig):** bestehender lokaler Docker-Stack als Standard.
- **Variante B (mittelfristig):** Frontend direkt vom Go-Backend ausliefern (nginx optional entfernen), API bleibt POST-only.
- **Variante C (langfristig):** native Install-/Start-Erfahrung auf Windows (Wizard + Service-Steuerung), weiterhin mit PostgreSQL.

### 6.3 Konfiguration

- Nutzer sollen keine Secrets manuell erzeugen müssen.
- Setup-Assistent erzeugt JWT-Secret und Relay-Token automatisch.
- Konfiguration in nutzerfreundlicher Form (z. B. `config.json`), keine Pflicht zur manuellen `.env`-Pflege.
- Klare Trennung zwischen editierbaren Basiswerten (Port, Hostname, Vereinsname) und sensitiven Werten.

### 6.4 Bedienbarkeit

- Start mit klarer Statusanzeige: Backend erreichbar, UI erreichbar, Relay verbunden.
- Anzeige der lokalen Zugriffsadresse für Smartphones (z. B. `http://192.168.x.x`).
- Verständliche Fehlermeldungen für häufige Ursachen (Firewall, Port belegt, Docker nicht aktiv).

## 7. Lösungsskizze (phasenweise)

### Phase A — Betriebsvereinfachung auf bestehender Basis

- Einstieg über einen lokalen „Startpunkt“ (Script/Starter), der:
  - lokale Anforderungen prüft (Docker verfügbar, Ports frei)
  - Stack startet
  - lokale URL/IP ausgibt
  - optional Relay-Start erklärt/auslöst
- Dokumentation für Vereine auf dieses Einstiegsszenario fokussieren.

### Phase B — Reduktion der Infrastrukturkomplexität

- Frontend-Build in das Go-Backend integrieren und statisch ausliefern.
- Reverse-Proxy für lokalen Modus optional machen.
- Routing so trennen, dass UI-GET und API-POST sauber koexistieren.

### Phase C — Native Windows-Bedienung

- Bereitstellung eines Windows-Setup-Wizards.
- Automatische Secret-Erzeugung und sichere Speicherung.
- Service-Management für Backend und Relay (Start/Stop/Status).

## 8. Akzeptanzkriterien

1. Ein Verein kann mit einer kompakten Anleitung jotti lokal auf Windows starten, ohne `.env` manuell zu bearbeiten.
2. Helfer können im selben WLAN mit Smartphone-Browser auf die UI zugreifen.
3. Relay kann unabhängig vom Backend gestartet/neu gestartet werden und druckt offene Aufträge weiterhin korrekt.
4. Setup- und Laufzeitfehler sind für Nicht-Techniker verständlich beschrieben.
5. Die bestehende fachliche Trennung bleibt erhalten: Druck-Fachlogik im Backend, Relay als Transport.

## 9. Risiken & Gegenmaßnahmen

- **Docker Desktop-Abhängigkeit bleibt bestehen:** klare Voraussetzungen + Diagnosechecks im Starter.
- **Komplexität bei Frontend-Embedding:** explizite Trennung von UI-Routing und API-Middleware.
- **Supportlast bei Windows-Netzwerkproblemen:** gezielte FAQ/Fehlerbilder (Firewall, WLAN-Isolation, Portkonflikte).
- **Token-/Secret-Handling:** automatische Generierung und sichere Ablage, kein Hardcoding.

## 10. Offene Entscheidungen

1. Soll Relay standardmäßig auf demselben Rechner laufen oder als optionale zweite Station dokumentiert werden?
2. Welche Konfigurationsspeicherung ist final (Datei-only vs. Datei + Windows Secret Store)?
3. Wird nginx im lokalen Standardmodus vollständig entfernt oder nur optional?
4. Welches Installationsformat ist Ziel (z. B. MSI) und in welchem Release-Fenster?

## 11. Erfolgsmessung

- Zeit bis zum ersten erfolgreichen lokalen Start
- Anteil erfolgreicher Erstinstallationen ohne manuelle Nacharbeit
- Anzahl typischer Supportfälle (Firewall/Config/Relay) pro Veranstaltung
- Stabilität des Betriebs während eines Veranstaltungstags (keine ungeplanten Neustarts)

