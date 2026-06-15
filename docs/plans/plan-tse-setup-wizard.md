# Plan: TSE-Setup-Wizard

> Source PRD: [docs/prds/prd-tse-setup-wizard.md](../prds/prd-tse-setup-wizard.md)

## Goal

Vereins-Admins ohne technische Vorkenntnisse können die TSE vollständig über die jotti-Oberfläche einrichten: jotti führt den fiskaly-TSS-Lebenszyklus (TSS anlegen → PUK → PIN → INITIALIZED → Client mit Kassen-Seriennummer registrieren) automatisch durch, schützt vor kostenwirksamen Fehlern (LIVE-Doppel-Anlage) und ist nach Abbrüchen wiederaufsetzbar. Dazu: erweiterter Verbindungstest (Client-State + Seriennummern-Abgleich) und Betreiber-Leitfaden. Schließt die Audit-Findings I-09, I-15.4 und D-07.

## Architectural decisions

Durable Entscheidungen für alle Phasen:

- **Ausgangslage nach dem Finanzamt-Umbau:** Die Route `/admin/tse-einrichtung`, das umgezogene manuelle Formular (`frontend/src/admin/tse/TSEKonfigurationSection.tsx`) und die faktische Status-Zeile (`frontend/src/admin/finanzamt/TSEAnbindungSection.tsx`) stehen bereits aus dem Finanzamt-Plan. Dieser Plan baut darauf auf; die ursprüngliche Phase 2 ist damit erledigt.
- **Backend-Routes (POST-only, admin):** `admin/tse-setup-pruefen` (seiteneffektfreier Befund), `admin/tse-einrichten` (Durchführung). Die Frontend-Route `/admin/tse-einrichtung` existiert bereits und rendert `frontend/src/admin/tse/TSEEinrichtungPage.tsx`; der Wizard wird dort ergänzt.
- **Setup-Authentifizierung nur mit API-Key/-Secret:** Die Setup-Operationen brauchen keine TSS-/Client-ID (`tse.Credentials` verlangt alle vier Felder und bleibt dem Signierbetrieb vorbehalten). Eigener schlanker Credentials-Typ für das Setup.
- **Setup-Operationen im fiskaly-Repository:** Erweiterung in `tse_repo`, teilt sich `doJSONRequest`/Token-Cache/Retry mit dem bestehenden Client. Die Admin-Authentifizierung (PIN-basierter Token je TSS) ist interner Belang dieses Moduls. Domain-Interface + Fake in `backend/domain/tse` analog `FakeClient`.
- **Orchestrator in der Settings-Application-Schicht**, Injection über Factory-Funktion analog `NewTSEConnectionTester`.
- **Keine Schema-Änderung.** PUK/PIN werden ausschließlich in der Response an die UI übergeben, nie persistiert, nie geloggt. Die Admin-PIN wird zufällig erzeugt.
- **Konfiguration wird erst bei Erfolg gespeichert:** atomar und vollständig (alle vier Felder) über das bestehende `UpdateTSEKonfiguration`; die Invariante „alle vier zusammen" bleibt unangetastet.
- **Bestätigte Umgebung als Parameter:** `tse-einrichten` erhält die vom Admin bestätigte Umgebung und bricht bei Abweichung von der tatsächlichen ab. LIVE-TSS-Anlage nur nach Tipp-Bestätigung (wörtlich „LIVE") im Frontend.
- **Client-`serial_number` = jotti-Kassen-Seriennummer** (UUID; erfüllt DSFinV-K ≥ 2.3, keine `/` und `_`).
- **`VerbindungStatus` wird erweitert** um Client-State und Seriennummern-Abgleich (Breaking Change am Response-Schema erlaubt, pre-release).
- **Kein eigener Sidebar-Eintrag:** Die Einrichtungsseite wird aus der TSE-Anbindungs-Sektion der Finanzamt-Seite verlinkt (`frontend/src/admin/finanzamt/TSEAnbindungSection.tsx`); eine Einstellungen-Seite gibt es nicht mehr.

## Inventory

- `backend/repository/tse_repo/fiskaly_client.go:105-144`: Client-Grundgerüst (baseURL, Credentials, Retry); `:305-397` `doJSONRequest`; `:399-434` API-Key-Auth mit Token-Cache, Umgebung aus Token-Claims (`:425`); `:212-244` `TestConnection` (prüft nur TSS-State)
- `backend/repository/tse_repo/fiskaly_client_live_test.go:25-34`: Muster env-gated Integrationstest (`FISKALY_TEST_*`-Variablen)
- `backend/repository/tse_repo/fiskaly_client_test.go`: Kontrakt-Test-Muster (Fake-Server asserted Pfade/Bodies/Header)
- `backend/domain/tse/client.go:20-42`: `Credentials` (alle vier Felder Pflicht); `:50-52` `ConnectionTester`; `:72-85` `VerbindungStatus`
- `backend/domain/tse/fake_client.go`: Fake-Muster für Domain-Interfaces
- `backend/domain/settings/tse_konfiguration.go:17-52`: Validierung „alle vier zusammen", `IstKonfiguriert`
- `backend/domain/settings/kassenidentitaet.go:9-12`: Kassen-Seriennummer (UUID)
- `backend/api/settings/application/query.go:23-28`: Factory-Muster `NewTSEConnectionTester`; `:77-117` `TestTSEVerbindung`; `:119-172` `GetTSEStatus` (liefert bereits Umgebung/`IstKonfiguriert`)
- `backend/api/settings/application/command.go`: `UpdateTSEKonfiguration` (atomares Speichern)
- `backend/api/settings/http/command_handler.go:39-94`: zog-Schema- und Handler-Muster
- `backend/api/admin.go:138-155`: Wiring der Settings-Handler und TSE-Routen; `:142-144` Factory-Injection
- `frontend/src/routes.ts:131-144`: lazy Admin-Routen `finanzamt` und `tse-einrichtung` (beide bereits angelegt)
- `frontend/src/admin/AdminSidebar.tsx:66-69`: Sidebar-Eintrag Finanzamt (Einstellungen entfernt)
- `frontend/src/admin/tse/TSEEinrichtungPage.tsx`: Host der Einrichtungsseite (rendert heute nur `TSEKonfigurationSection`); `frontend/src/admin/tse/TSEKonfigurationSection.tsx:29-256`: umgezogenes manuelles Formular (Speichern/Leeren/Verbindung testen), Verbindungstest-Anzeige bei `:187-196`
- `frontend/src/admin/finanzamt/TSEAnbindungSection.tsx`: faktische Status-Zeile (konfiguriert/Umgebung) mit Link auf `/admin/tse-einrichtung`
- `frontend/src/lib/EinstellungenBackend.ts:31-35`: `TSEVerbindungStatusSchema` (wird erweitert); `:136-142` `testTSEVerbindung`
- `frontend/src/admin/settings/hooks.ts:87-116`: `useTSEKonfiguration`; `:146-157` `useTSEStatus` (TanStack-Query-Muster, bleiben am Ort)
- `temp/fiskaly_sign_de_api_spec.json`, `temp/fiskaly_SIGN_DE_Postman_Environment_collection.json`: API-Spec 2.2.2 und Lifecycle-Referenz (Kontrakt-Tests)
- `docs/betrieb/leitfaden-betreiber.md`: bestehender Betreiber-Leitfaden (Einhängepunkt für Phase 6)
- Live-Test-TSS in fiskaly TEST: TSS `728e3cda-…`, Client `90977ec5-…` (für Verbindungstest-Verifikation; Wizard-Integrationstest legt eigene TSS an)

## Resolved decisions

Aus dem PRD-Prozess (2026-06-11, alle mit User abgestimmt):

- **PUK/PIN einmalig anzeigen, extern verwahren:** keine Speicherung in jotti, keine `admin_puk`/`admin_pin`-Spalten.
- Eigene Admin-Seite „TSE-Einrichtung" mit umgezogener manueller Konfiguration und Status + Link: durch den Finanzamt-Plan bereits umgesetzt (Route, `TSEKonfigurationSection`, `TSEAnbindungSection`). Der Wizard baut darauf auf, statt das erneut anzulegen.
- **Vorhandene TSS:** Übernahme anbieten (Wiederaufnahme nach Teilfehler eingeschlossen), keine stille Doppel-Anlage.
- **LIVE-Schutz:** Tipp-Bestätigung „LIVE" vor kostenwirksamer Anlage; in TEST genügt ein Klick.
- **Tests:** Kontrakt-Tests (Setup-Operationen), Unit-Tests (Orchestrator gegen Fake), env-gated Integrationstest; keine Frontend-Komponententests.
- **Phasenzuschnitt:** sechs Phasen wie unten, vom User bestätigt.

## Open questions / Risks

- **TEST-Konto füllt sich:** Jeder Integrationstest-Lauf von Phase 4 hinterlässt eine nicht löschbare TSS im TEST-Konto. Test bewusst env-gated lassen und sparsam ausführen.
- **LIVE-Pfad nicht real testbar:** Die LIVE-Schutzlogik (Tipp-Bestätigung, Umgebungs-Abgleich) ist nur über Unit-Tests abgesichert, entsprechend sorgfältig testen.
- fiskaly-Preise für den Leitfaden (Phase 6) bei Umsetzung aktuell recherchieren.

---

## Phase 1: Erweiterter Verbindungstest

**User stories**: 22, 23, 24, 29

### Context

- `backend/repository/tse_repo/fiskaly_client.go:212-244`: `TestConnection` prüft nur TSS-State
- `backend/domain/tse/client.go:72-85`: `VerbindungStatus` (Umgebung, TSSState)
- `backend/api/settings/application/query.go:77-117`: `TestTSEVerbindung` (kennt via `SettingsRepo` auch die Kassenidentität)
- `frontend/src/lib/EinstellungenBackend.ts:31-35`: `TSEVerbindungStatusSchema`
- `frontend/src/admin/tse/TSEKonfigurationSection.tsx:187-196`: bisherige Verbindungstest-Anzeige (zeigt heute nur Umgebung + TSS-Status)
- Audit I-15.4 (Client-State ungeprüft), I-09 (Seriennummern-Abgleich)

### What to build

Der Verbindungstest ruft zusätzlich den fiskaly-Client der konfigurierten TSS ab. `VerbindungStatus` transportiert neu Client-State und Client-`serial_number`; die Application-Schicht vergleicht die `serial_number` mit der Kassen-Seriennummer. Die Verbindungstest-Anzeige in `TSEKonfigurationSection.tsx` zeigt das Ergebnis aufgeschlüsselt (Umgebung, TSS-Zustand, Client-Zustand, Seriennummern-Abgleich); ein nicht-`REGISTERED`-Client oder eine Seriennummern-Abweichung wird deutlich als Fehler mit verständlichem deutschen Text gemeldet.

### Acceptance criteria

- [x] Verbindungstest meldet einen nicht-`REGISTERED`-Client als Fehler (Kontrakt-Test mit Fake-Server)
- [x] Abweichung `serial_number` ↔ Kassen-Seriennummer wird als Fehler gemeldet (Unit-Test)
- [x] UI zeigt Umgebung, TSS-Zustand, Client-Zustand und Abgleich-Ergebnis aufgeschlüsselt
- [ ] Manuell gegen die Live-Test-TSS verifiziert; `make check` grün

---

## Phase 2: Seite „TSE-Einrichtung" und Umzug der manuellen Konfiguration — durch den Finanzamt-Plan erledigt

**User stories**: 19, 20, 21

Diese Phase ist vollständig durch [plan-finanzamt.md](plan-finanzamt.md) (Commits `591146c`, `55e4226`) umgesetzt. Sie bleibt zur Nachvollziehbarkeit dokumentiert; hier ist nichts mehr zu bauen.

- Die lazy Route `/admin/tse-einrichtung` existiert (`frontend/src/routes.ts:138-144`) und rendert `frontend/src/admin/tse/TSEEinrichtungPage.tsx`.
- Das manuelle Konfigurationsformular (API-Key/-Secret, TSS-ID, Client-ID, Speichern/Leeren/Verbindung testen) ist unverändert nach `frontend/src/admin/tse/TSEKonfigurationSection.tsx` umgezogen.
- Die frühere Einstellungen-TSE-Sektion ist zur faktischen Status-Zeile geworden und lebt jetzt als `frontend/src/admin/finanzamt/TSEAnbindungSection.tsx` auf der Finanzamt-Seite (konfiguriert ja/nein, Umgebung, Link „Einrichten oder ändern"). Die Einstellungen-Seite gibt es nicht mehr.

Folge für den Wizard: Die nachfolgenden Phasen ergänzen den geführten Ablauf auf der bestehenden `TSEEinrichtungPage`; das manuelle Formular bleibt dort als Experten-/Fallback-Bereich erhalten.

---

## Phase 3: Prüf-Schritt: Befund ohne Seiteneffekte

**User stories**: 3, 10, 13, 31

### Context

- `backend/repository/tse_repo/fiskaly_client.go:399-434`: Auth + Umgebung aus Token-Claims (wiederverwendbar)
- `backend/domain/tse/client.go:20-42`: `Credentials` verlangt alle vier Felder → eigener Setup-Credentials-Typ nötig
- `backend/api/admin.go:138-155`: Wiring-/Routen-Muster; `backend/api/settings/http/command_handler.go:39-94`: zog-Muster
- `temp/fiskaly_sign_de_api_spec.json`: List-TSS-/List-Clients-Endpunkte
- `frontend/src/admin/tse/TSEEinrichtungPage.tsx`: Host der Wizard-Schritte (rendert heute nur `TSEKonfigurationSection`); die Schritte kommen darüber, das manuelle Formular bleibt als Experten-Bereich darunter

### What to build

Erster Teil der Setup-Operationen (Auth nur mit API-Key/-Secret, TSS listen, Clients einer TSS listen) samt Domain-Interface und Fake. Neuer Endpoint `admin/tse-setup-pruefen`: nimmt API-Key/-Secret entgegen, liefert Umgebung und die vorhandenen TSS mit Zustand sowie (je TSS) einen ggf. vorhandenen Client mit passender Kassen-Seriennummer. Im Frontend entstehen auf `TSEEinrichtungPage` die ersten beiden Wizard-Schritte (Zugangsdaten → Befund) mit deutlich sichtbarer Umgebungs-Anzeige (TEST/LIVE). Es passieren ausschließlich Lese-Requests; gespeichert wird nichts.

### Acceptance criteria

- [x] Befund zeigt Umgebung und vorhandene TSS mit Zustand; ein Client mit passender Kassen-Seriennummer wird erkannt und ausgewiesen
- [x] Kontrakt-Tests belegen: Prüfung sendet ausschließlich Auth- und GET-Requests
- [x] Falsche Zugangsdaten führen zu einer verständlichen deutschen Fehlermeldung
- [ ] Manuell verifiziert: Befund gegen das echte TEST-Konto zeigt die Audit-TSS

---

## Phase 4: Einrichtung Ende-zu-Ende (leeres Konto)

**User stories**: 1–12, 30, 32, 33, 35, 36, 38

### Context

- `temp/fiskaly_SIGN_DE_Postman_Environment_collection.json`: Lifecycle-Referenz (TSS anlegen, PUK, PIN, Admin-Auth, INITIALIZED, Client)
- `backend/repository/tse_repo/fiskaly_client_live_test.go:25-34`: Integrationstest-Muster
- `backend/api/settings/application/command.go`: atomares Speichern via `UpdateTSEKonfiguration`
- `backend/domain/settings/kassenidentitaet.go:9-12`: Seriennummer für die Client-Registrierung

### What to build

Restliche Setup-Operationen (TSS idempotent anlegen, PUK beziehen, zufällige Admin-PIN setzen, Admin-Auth, TSS initialisieren, Client registrieren) und der Setup-Orchestrator als Zustandsmaschine über den fiskaly-Ist-Zustand. Endpoint `admin/tse-einrichten` führt den Voll-Durchlauf für ein leeres Konto aus (Parameter: bestätigte Umgebung; Abbruch bei Abweichung). Existiert bereits eine aktive TSS, wird die Neuanlage verweigert (Hinweis auf manuelle Einrichtung; Übernahme folgt in Phase 5). Im Frontend: Bestätigungs-Schritt (in LIVE mit Tipp-Bestätigung „LIVE"), Durchführungs-Anzeige, einmalige PUK/PIN-Anzeige mit erzwungener Verwahr-Bestätigung, atomares Speichern der vollständigen Konfiguration, automatischer Abschluss-Verbindungstest. Env-gated Integrationstest für den Voll-Durchlauf.

### Acceptance criteria

- [ ] Aus einem leeren fiskaly-TEST-Konto entsteht eine signierfähige TSS samt Client mit Kassen-Seriennummer (env-gated Integrationstest; anschließend signiert ein Direktverkauf in der Dev-Umgebung) — Test geschrieben (`fiskaly_setup_live_test.go`), Live-Lauf + Dev-Verifikation noch offen
- [x] PUK und PIN erscheinen genau einmal in der Antwort und werden weder persistiert noch geloggt
- [x] LIVE-Anlage erfordert die Tipp-Bestätigung; Abweichung zwischen bestätigter und tatsächlicher Umgebung bricht ab (Unit-Tests)
- [x] Bei vorhandener aktiver TSS wird keine neue angelegt (Unit-Test)
- [x] Die Konfiguration wird erst nach erfolgreichem Abschluss gespeichert; ein Abbruch hinterlässt keine halbe Konfiguration in der DB

---

## Phase 5: Übernahme vorhandener TSS und Wiederaufnahme

**User stories**: 14–18, 34, 37

### Context

- Orchestrator und Befund aus Phase 3/4
- fiskaly-Zustandsmodell: PUK nur im Zustand CREATED idempotent wiederbeschaffbar; ab UNINITIALIZED ist die PIN nötig

### What to build

Der Wizard kann eine im Befund gewählte vorhandene TSS übernehmen: Ein vorhandener Client mit passender Kassen-Seriennummer wird übernommen statt neu angelegt; fehlt er, wird er registriert. Der Orchestrator setzt aus jedem Zwischenzustand wieder auf: bei CREATED über den idempotenten PUK-Refetch ohne Nutzereingabe, ab UNINITIALIZED über eine PIN-Nachfrage im Wizard (der Admin hat die PIN verwahrt). Ist die PIN unbekannt, endet der Flow in einer verständlichen Sackgassen-Meldung mit Auswegen (fiskaly-Support bzw. bewusste Neuanlage). Die Verweigerung aus Phase 4 („vorhandene aktive TSS") wird durch das Übernahme-Angebot ersetzt.

### Acceptance criteria

- [x] Abbruch nach TSS-Anlage: Ein erneuter Wizard-Lauf vollendet die Einrichtung ohne zweite TSS (Unit-Tests je Zwischenzustand: CREATED, UNINITIALIZED, INITIALIZED ohne Client)
- [x] Übernahme einer TSS mit vorhandenem passendem Client registriert keinen neuen Client
- [x] PIN-Nachfrage: Übernahme einer initialisierten TSS funktioniert mit eingegebener PIN
- [x] Unbekannte PIN führt zu einer verständlichen Meldung mit Auswegen, nicht zu einem technischen Fehler

---

## Phase 6: Betreiber-Leitfaden

**User stories**: 25, 26, 27, 28

### Context

- `docs/betrieb/leitfaden-betreiber.md`: bestehender Betreiber-Leitfaden (Einhängepunkt oder Schwester-Dokument)
- Audit D-07: fehlende Doku zu TSS-Lifecycle und Client-Registrierung
- `docs/compliance.md`: Betreiberpflichten (Querverweis)

### What to build

Betreiber-Dokumentation für die TSE-Einrichtung: fiskaly-Konto registrieren und API-Key im Dashboard erstellen (mit dem Hinweis, dass TSS-/Client-Anlage dort nicht möglich ist), Wizard-Ablauf Schritt für Schritt, Verwahrung von Admin-PUK/-PIN samt Verlustfolgen, Wechsel TEST→LIVE inklusive aktuell recherchierter Kosten, manueller Fallback-Weg für bestehende TSS. Querverweise aus `compliance.md` aktualisieren, soweit sie die Einrichtung betreffen.

### Acceptance criteria

- [ ] Leitfaden beschreibt beide Einrichtungswege vollständig und laienverständlich
- [ ] Kapitel zur PUK/PIN-Verwahrung mit Verlustfolgen vorhanden
- [ ] TEST→LIVE-Wechsel inkl. Kosten beschrieben
- [ ] D-07 damit geschlossen (Abgleich gegen Audit-Empfehlung)
