# Plan: TSE-Setup-Wiederaufnahme aus Sackgassen

> Folge-Plan zu [plan-tse-setup-wizard.md](plan-tse-setup-wizard.md). Adressiert die
> UX-Review-Findings F2 und F8 (heilt dabei F7) sowie einen autoritativen
> Recherche-Befund (PUK-basierter PIN-Reset): Wege aus den Sackgassen, in denen ein
> Vereins-Admin ohne Admin-PIN feststeckt.

## Goal

Ein Admin, der eine vorhandene TSS im fiskaly-Konto vorfindet, kommt auch ohne die
verwahrte Admin-PIN aus der Einrichtung heraus:

- **F8 (Korrektheit, ohne Entscheidung):** Ist die vorgefundene TSS bereits
  `INITIALIZED` und ein passender Client bereits `REGISTERED`, ist die TSE faktisch
  einsatzbereit — jotti muss nur noch die lokale Konfiguration speichern. Dafür ist
  keine Admin-PIN nötig (es folgt keine privilegierte fiskaly-Operation; per Spec
  brauchen nur Client-Anlage/-Änderung und der TSS-Wechsel nach `INITIALIZED`/`DISABLED`
  Admin-Auth). Heute verlangt die Übernahme trotzdem die PIN und läuft in eine Sackgasse.
  Der reale Nutzen liegt bei den Fällen, in denen die Kassen-Seriennummer unverändert ist
  (gleiche Datenbank) und nur die `tse_konfiguration` fehlt: **Schlüsselrotation** und
  **„Alle Felder leeren" mit anschließendem Neu-Verbinden** sowie der seltene Fall, dass
  das Speichern nach vollem Lebenszyklus fehlschlug (`acd86e3c`, „recoverable via
  takeover"). Bei Neuinstallation oder DB-Verlust ohne Backup ist die Seriennummer neu
  (per `gen_random_uuid()` einmalig bei der Migration, `database/migrations/01_initial.up.sql:370`);
  dann matcht kein Client, und F8 greift bewusst nicht, weil eine privilegierte
  Client-Registrierung nötig ist. Dass mitten im Setup abgebrochen wird (und so ein Orphan
  entsteht), ist real: Der Deploy-Schritt `CREATED → UNINITIALIZED` braucht laut Spec
  ≥ 30 s Timeout und ist ggf. zu wiederholen.
- **F2 (Produktentscheidung):** In der TEST-Umgebung soll der Admin eine neue TSE
  anlegen können, auch wenn bereits eine (nicht nutzbare, weil PIN-lose) TSS im Konto
  liegt. Heute verweigert das Backend jede Neuanlage, sobald irgendeine nicht-
  `DISABLED` TSS existiert — was beim Ausprobieren der Normalfall ist und den Admin in
  der Sackgasse hält, bis fiskaly die inaktive TEST-TSS von selbst bereinigt (siehe
  Validierung; das dauert bis zu zwei Wochen).
- **PUK-Reset (Recherche-Befund, TEST und LIVE):** Ist die PIN verloren oder nach fünf
  Fehlversuchen gesperrt, aber der Admin-PUK verwahrt, setzt jotti die PIN über den PUK
  zurück und schließt die Übernahme ab — ohne neue, kostenpflichtige TSS. Das ist auch
  für LIVE ein echter Self-Service-Ausweg statt eines Support-Falls.

## Decision (entschieden 2026-06-16)

F2 ist eine Produktentscheidung. **Entschieden: Option A** (TEST erlauben, LIVE harte
Sperre):

- **TEST:** Neuanlage trotz vorhandener übernehmbarer TSS erlauben, hinter einer
  expliziten Sekundäraktion „Stattdessen neue TSE anlegen". Kein Tipp-Zwang (TEST ist
  kostenlos und steuerlich ungültig).
- **LIVE:** harte Sperre beibehalten (`ErrTSEBereitsEingerichtet`). Eine
  versehentliche zweite LIVE-TSS verursacht laufende Kosten; eine echt verlorene
  LIVE-PIN ist ein Support-Fall, kein Self-Service.

Verworfen (Option B): auch in LIVE erlauben, nur warnen — wegen der Kostenfolge nicht
gewählt. Der LIVE-PIN-Verlust ist ab Phase 3 ohnehin per PUK-Reset self-service heilbar
(ohne neue TSS), sodass die LIVE-Sperre keinen echten Self-Service-Bedarf abschneidet.

Autoritativ gestützt: LIVE-TSS sind nicht löschbar (kein `DELETE`-Endpunkt, nur
`DISABLED`) und persistieren samt Kosten, während TEST-TSS automatisch bereinigt
werden (siehe Validierung). Genau eine LIVE-TSS pro Kasse ist auch fachlich sauber
(keine gesplitteten Fiskaldaten, deckt sich mit „eine TSS je jotti-Instanz" im
Leitfaden). Restrisiko der LIVE-Sperre: Geht bei einem LIVE-DB-Verlust auch die PIN
verloren, bleibt nur der fiskaly-Support, bevor der Verein wieder kassieren kann. Mit
verwahrter PIN ist die LIVE-Wiederaufnahme dagegen Self-Service (Phase-5-Übernahme
registriert einen neuen Client). Diesen Restschmerz im Leitfaden benennen.

## Validierung gegen autoritative Quellen

Geprüft gegen die fiskaly SIGN DE API (OpenAPI v2.2.2, `temp/fiskaly_sign_de_api_spec.json`,
online `developer.fiskaly.com/api/kassensichv/v2`) und die rechtlichen Grundlagen
(KassenSichV/AO §146a, DSFinV-K 2.3, BSI TR-03153-1; Quellen in `temp/tse_ressources.md`).

- **F8-Pfad ist nachweislich mutationsfrei.** Laut Spec verlangen nur Client-Anlage/-Änderung
  und der TSS-Statuswechsel nach `INITIALIZED`/`DISABLED` Admin-Authentifizierung. Eine
  `INITIALIZED` TSS mit `REGISTERED` Client braucht keinen weiteren privilegierten Schritt;
  das Überspringen von `AuthentifiziereAdmin` lässt nichts Offenes zurück.
- **TSS sind nicht löschbar (KassenSichV).** Kein `DELETE`-Endpunkt, nur Statuswechsel nach
  `DISABLED`. In LIVE bleibt eine TSS samt Kosten dauerhaft — stützt die LIVE-Sperre.
- **TEST-TSS werden bereinigt — die bisherige Annahme „nicht löschbar" gilt für TEST nicht.**
  TEST-TSS, die `DISABLED` oder über 14 Tage inaktiv sind, werden mindestens jeden Sonntag
  gelöscht. TEST erlaubt höchstens fünf aktive TSS (`CREATED`/`UNINITIALIZED`/`INITIALIZED`);
  darüber verweigert fiskaly die Neuanlage (`E_TSS_LIMIT_REACHED`), bis eine TSS auf
  `DISABLED` gesetzt wird. Folgen für F2 unter Open questions.
- **Admin-PIN blockiert nach fünf Fehlversuchen** (`E_ADMIN_PIN_BLOCKED`) und ist dann nur
  mit dem Admin-PUK zurücksetzbar (`PATCH /tss/{id}/admin`, derselbe Endpunkt wie
  `SetzeAdminPIN`). Online bestätigt.
- **`serial_number` ist je TSS eindeutig; Deregistrierung ist umkehrbar.** Ein `DEREGISTERED`
  Client wird über `PATCH /client/{id}` mit `state=REGISTERED` reaktiviert, nicht durch einen
  neuen Client mit derselben Seriennummer (verstößt gegen die Eindeutigkeit). Folgen für die
  F7-Heilung unter Architectural decisions.
- **Seriennummer-Konsistenz (I-09) bleibt erfüllt.** jottis Kassen-Seriennummer ist eine UUID
  und erfüllt DSFinV-K 2.3 (kein `/`, kein `_`) und die Eindeutigkeit nach BSI TR-03153-1
  9.3.1. F8/F2 ändern daran nichts.

## Architectural decisions

- **Keine Schema-Änderung.** Beide Findings betreffen nur den Setup-Orchestrator
  (`backend/api/settings/application/setup.go`), das Domain-Setup-Interface
  (`backend/domain/tse`) und den Wizard (`frontend/src/admin/tse/`).
- **„Einsatzbereit ohne Arbeit" ist der einzige PIN-freie Übernahmepfad.** F8 lockert
  die PIN-Pflicht ausschließlich, wenn `state == INITIALIZED` **und** ein passender
  Client mit `state == REGISTERED` existiert. Jeder Pfad, der eine privilegierte
  fiskaly-Operation auslöst (Initialisieren, Client registrieren/neu registrieren),
  braucht weiterhin die Admin-PIN.
- **`passenderClient` muss den Client-State berücksichtigen.** Heute matcht
  `passenderClient` (`backend/api/settings/application/query.go:209`) nur die
  `serial_number`. Für F8 muss der „kein Arbeit nötig"-Zweig zusätzlich `REGISTERED`
  prüfen; nur ein `REGISTERED` Client gilt als fertig.
- **`DEREGISTERED` Client wird reaktiviert, nicht neu angelegt.** Autoritativ geklärt
  (siehe Validierung): Die `serial_number` ist je TSS eindeutig, also schlägt ein neuer
  Client mit derselben Seriennummer fehl. Ein `DEREGISTERED` Client mit passender
  Seriennummer wird über `PATCH /tss/{id}/client/{clientID}` mit `state=REGISTERED` auf
  **demselben** `client_id` reaktiviert (Admin-authentifiziert, also weiterhin PIN).
  Das erfordert eine neue Operation `ReaktiviereClient` im `SetupClient`-Interface und
  einen Orchestrator-Zweig; der heutige Code überspringt für jeden gematchten Client die
  Registrierung und speichert die Konfiguration auf einen toten Client (das ist F7). Weil
  jotti nie selbst deregistriert, ist `DEREGISTERED` selten; wer den Aufwand nicht in
  Phase 1 will, descopt F7 und meldet den Zustand wenigstens explizit als nicht-automatisch
  übernehmbar, statt still eine signierunfähige Konfiguration zu speichern.
- **F2 über expliziten Parameter, nicht über stille Lockerung.** `RichteTSEEin` bzw.
  der Endpoint `admin/tse-einrichten` erhält ein Flag `neuAnlegenTrotzVorhandener`
  (Default `false`). Die Sperre `hatAktiveTSS` wird nur übersprungen, wenn das Flag
  gesetzt ist **und** `umgebung == TEST`. So bleibt der LIVE-Schutz unangetastet und
  die Neuanlage ist immer eine bewusste Nutzeraktion.

## Inventory

- `backend/api/settings/application/setup.go:41-135` `RichteTSEEin` (Sperre
  `hatAktiveTSS` bei `:78`); `:153-266` `UebernimmTSE` (PIN-Pflicht bei `:240-242`);
  `:275-314` `vollendeLebenszyklus` (INITIALIZED-Zweig `:309-313`,
  `AuthentifiziereAdmin` vor `if !hatClient`); `:333-342` `hatAktiveTSS`
- `backend/api/settings/application/query.go:207-216` `passenderClient` (matcht nur
  serial_number, liefert aber `State`)
- `backend/api/settings/application/setup_test.go`: Unit-Test-Muster gegen den Fake
- `backend/domain/tse/fake_setup_client.go`: Fake des `SetupClient` (nicht
  `fake_client.go`, das ist der Signier-Fake ohne `AuthentifiziereAdmin`). Zeichnet heute
  nur `AuthentifiziertePIN` auf; für die Assertion „AuthentifiziereAdmin nicht aufgerufen"
  fehlt ein Aufruf-Zähler `AuthAdminCalls int` (analog `CreateTSSCalls`/`HoleAdminPUKCalls`),
  da eine leere PIN sonst nicht von „nicht aufgerufen" unterscheidbar ist. Für die
  F7-Reaktivierung kommt eine `ReaktiviereClient`-Methode samt Aufzeichnung hinzu.
- `backend/api/settings/http/command_handler.go`: zog-Schema des Einrichten-Requests
  (neues Flag ergänzen)
- `frontend/src/admin/tse/TSEEinrichtungWizard.tsx`: `brauchtPin` (PIN-Gating),
  `BefundSchritt` (Verzweigung Übernahme/Neuanlage), `UebernahmeSchritt`,
  `BestaetigungSchritt`, `PinFehltHinweis` (Disclosure aus dem F1-Fix)
- `frontend/src/lib/EinstellungenBackend.ts`: `TSEEinrichten`-Typ, `richteTSEEin`

## Open questions / Risks

- **F2-Entscheidung getroffen:** Option A (TEST erlauben, LIVE harte Sperre), siehe oben.
- **F2 ist in TEST nach oben begrenzt, nicht „füllt sich dauerhaft".** Die Annahme aus
  dem Wizard-Plan (nicht löschbare Test-TSS) ist autoritativ widerlegt: TEST-TSS werden
  bei `DISABLED` oder >14 Tagen Inaktivität wöchentlich gelöscht. Stattdessen gilt eine
  harte Obergrenze von fünf aktiven TSS in TEST. Nach ca. fünf abgebrochenen, PIN-losen
  TSS verweigert fiskaly die Neuanlage (`E_TSS_LIMIT_REACHED`), und jotti kann die alten
  nicht auf `DISABLED` setzen (keine PIN, keine Disable-Operation — laut PRD Out of Scope).
  Der Verein ist dann bis zur nächsten Inaktivitäts-Bereinigung blockiert. Akzeptabel, weil
  selbstheilend, aber im Leitfaden korrekt beschreiben (nicht „nicht löschbar").
- **Admin-PIN-Sperre verschärft die Sackgasse.** Die PIN blockiert nach fünf Fehlversuchen
  und ist dann nur per PUK rücksetzbar. Die heutige Übernahme-Meldung („verwahrte PIN
  erneut prüfen") lädt zu wiederholten Versuchen und damit zur Selbst-Aussperrung ein.
  Adressiert durch Phase 3 (PUK-Reset behandelt verlorene und gesperrte PIN) plus
  Warnhinweise in Wizard und Leitfaden vor der 5-Versuche-Sperre.
- **F7-Reaktivierung gegen die echte API absichern.** Der `DEREGISTERED`-Pfad muss per
  `PATCH state=REGISTERED` reaktivieren (nicht neu anlegen); im env-gated Integrationstest
  schwer reproduzierbar, da jotti nie deregistriert. Mindestens Kontrakt-Test gegen den
  Fake-Server (richtiger Pfad/Body/Admin-Token).
- **F2 vermehrt übernehmbare TSS in der Befund-Liste.** Da alle TSS derselben Instanz mit
  derselben Seriennummer registriert werden, hat nach mehreren F2-Läufen jede einen
  `REGISTERED` Client und wird mit F8 als PIN-frei übernehmbar angeboten — der Wizard
  rendert je TSS einen `UebernahmeSchritt`. In TEST harmlos, aber unübersichtlich; in
  `BefundSchritt` die jüngste/eigene TSS hervorheben oder die Liste eindampfen.
- **F8-Reichweite ist durch die Seriennummer begrenzt** (siehe Goal/F8): greift nur bei
  unveränderter Kassen-Seriennummer (gleiche DB). Bei Neuinstallation/DB-Verlust ohne
  Backup hilft nur die PIN-gegatete Übernahme (Phase 5) oder ein DB-Restore aus `pg_dump`.
- **F8-Sicherheit:** Autoritativ bestätigt mutationsfrei (siehe Validierung). Durch den
  Fake-Test mit Aufruf-Assertion (`AuthAdminCalls == 0`) absichern; das abschließende
  Verbindungstest fängt Fehleinschätzungen zusätzlich ab.

---

## Phase 1: F8 — PIN-freie Übernahme einer einsatzbereiten TSS

**Findings:** F8, heilt F7

### Context

- `backend/api/settings/application/setup.go:240-242`: PIN-Pflicht für
  `UNINITIALIZED`/`INITIALIZED` greift heute auch, wenn nichts mehr zu tun ist
- `:309-313`: `AuthentifiziereAdmin(pin)` läuft vor dem `if !hatClient`-Guard und
  gatet ausschließlich die danach übersprungene `RegistriereClient`
- `backend/api/settings/application/query.go:209-216`: `passenderClient` liefert
  `State`, wird aber für die hatClient-Entscheidung in `setup.go:216` ohne State-Check
  verwendet

### What to build

- Backend: In `UebernimmTSE` einen Zweig „bereits einsatzbereit" erkennen
  (`state == INITIALIZED` **und** passender Client **`REGISTERED`**). In diesem Fall
  die PIN-Pflicht (`:240-242`) nicht anwenden und in `vollendeLebenszyklus` den
  `AuthentifiziereAdmin`-Aufruf überspringen (keine fiskaly-Mutation). Anschließend
  wie gehabt atomar die Konfiguration speichern. Den State-Check in die
  hatClient-Bestimmung ziehen, sodass nur ein `REGISTERED` Client als „fertig" gilt.
  Fake um `AuthAdminCalls int` erweitern, damit der „nicht aufgerufen"-Nachweis
  unabhängig von der (leeren) PIN trägt.
- Backend (F7): Einen passenden, aber `DEREGISTERED` Client nicht als fertig werten und
  auch keinen neuen anlegen (Seriennummer je TSS eindeutig). Stattdessen über eine neue
  `SetupClient`-Operation `ReaktiviereClient` (`PATCH /tss/{id}/client/{clientID}`,
  `state=REGISTERED`, Admin-authentifiziert) denselben Client reaktivieren; die PIN
  bleibt erforderlich. Kontrakt-Test gegen den Fake-Server für Pfad/Body/Admin-Token.
- Frontend: `brauchtPin` so erweitern, dass es für eine `INITIALIZED` TSS mit
  passendem `REGISTERED` Client `false` liefert. Der `UebernahmeSchritt` zeigt dann
  kein PIN-Feld und keinen `PinFehltHinweis`, sondern bietet direkt „TSE übernehmen"
  an, das nur die Konfiguration speichert.

### Acceptance criteria

- [x] Übernahme einer `INITIALIZED` TSS mit passendem `REGISTERED` Client gelingt mit
      leerer PIN; der Fake belegt über `AuthAdminCalls == 0`, dass `AuthentifiziereAdmin`
      **nicht** aufgerufen wird (Unit-Test)
- [x] Übernahme einer `INITIALIZED` TSS mit passendem **`DEREGISTERED`** Client verlangt
      weiterhin die PIN und reaktiviert per `PATCH state=REGISTERED` **denselben**
      `client_id`; es wird **kein** neuer Client angelegt (Unit-Test, Aufruf-Assertion)
- [x] Übernahme einer `INITIALIZED` TSS **ohne** passenden Client verlangt weiterhin
      die PIN (Registrierung nötig) — bestehendes Verhalten bleibt
- [x] Wizard zeigt für die einsatzbereite TSS kein PIN-Feld und schließt die Übernahme
      mit einem Klick ab; `make check` grün

---

## Phase 2: F2 — Neue TSE in TEST trotz vorhandener TSS

**Findings:** F2 · **setzt die Decision oben voraus**

### Context

- `backend/api/settings/application/setup.go:78` + `:333-342`: `hatAktiveTSS`-Sperre,
  umgebungsunabhängig
- `frontend/src/admin/tse/TSEEinrichtungWizard.tsx` `BefundSchritt`: rendert
  `BestaetigungSchritt` (Neuanlage) heute nur bei `nurDisabledOderLeer`

### What to build

- Backend: Parameter `neuAnlegenTrotzVorhandener bool` in `RichteTSEEin` und im
  zog-Schema von `admin/tse-einrichten`. `hatAktiveTSS`-Sperre nur überspringen,
  wenn Flag gesetzt **und** `umgebung == UmgebungTest`. LIVE bleibt hart gesperrt.
- Frontend: In `BefundSchritt` zusätzlich zum Übernahme-Angebot in **TEST** eine
  klar untergeordnete Aktion „Stattdessen neue TSE anlegen" anbieten (ruft
  `richteTSEEin` mit gesetztem Flag). Den `PinFehltHinweis` (F1) auf diesen Weg
  verweisen lassen, sodass der PIN-lose Admin in TEST einen echten Self-Service-
  Ausweg hat. In LIVE bleibt es bei Übernahme bzw. Support-Verweis. Bei mehreren
  übernehmbaren TSS die jüngste/eigene hervorheben, damit die Liste nach wiederholter
  Neuanlage übersichtlich bleibt.
- Backend/Frontend: Das fiskaly-Limit von fünf aktiven TEST-TSS (`E_TSS_LIMIT_REACHED`)
  als eigene, verständliche Meldung abfangen („In TEST sind fünf TSS erreicht; alte
  werden bei Inaktivität automatisch bereinigt"), statt als technischer Fehler. jotti
  kann die alten nicht stilllegen (keine PIN, keine Disable-Operation).

### Acceptance criteria

- [ ] In TEST legt „Stattdessen neue TSE anlegen" trotz vorhandener `INITIALIZED` TSS
      eine zweite, frische TSE an (Unit-Test mit gesetztem Flag)
- [ ] In LIVE wird die Neuanlage trotz Flag verweigert (`ErrTSEBereitsEingerichtet`,
      Unit-Test)
- [ ] Ohne Flag bleibt das bisherige Verhalten unverändert (Sperre greift in beiden
      Umgebungen)
- [ ] `E_TSS_LIMIT_REACHED` (fünf aktive TEST-TSS) wird als verständliche Meldung
      angezeigt, nicht als technischer Fehler
- [ ] Wizard: der PIN-lose Admin findet in TEST den Neuanlage-Ausweg ohne Umweg über
      einen Fehlversuch; Leitfaden ergänzt um den TEST-Hinweis und die Korrektur, dass
      TEST-TSS automatisch bereinigt werden (nicht „nicht löschbar"), inkl. 5-TSS-Limit
      und 5-Versuche-PIN-Sperre

---

## Phase 3: PUK-basierter Admin-PIN-Reset

**Findings:** Recherche-Befund (autoritativ) · TEST und LIVE

### Context

- fiskaly SIGN DE Spec: `PATCH /tss/{id}/admin` (Body `admin_puk` + `new_admin_pin`) setzt
  bzw. entsperrt die Admin-PIN — laut Spec ausdrücklich der Weg, eine nach fünf
  Fehlversuchen gesperrte PIN (`E_ADMIN_PIN_BLOCKED`) zurückzusetzen. Funktioniert auch
  auf einer `UNINITIALIZED`/`INITIALIZED` TSS. Es entsteht keine neue TSS, also keine
  Kosten — deshalb in TEST **und** LIVE vertretbar.
- jotti ruft diesen Endpunkt heute nur im CREATED-Lebenszyklus auf
  (`backend/repository/tse_repo/fiskaly_setup.go:189` `SetzeAdminPIN`); die Übernahme ab
  UNINITIALIZED bietet keinen Reset an.
- `frontend/src/admin/tse/TSEEinrichtungWizard.tsx`: Einhängepunkt ist die PIN-Sackgasse
  (`pinUnbekannt` `:342-352`, `PinFehltHinweis` `:368-392`). `ErgebnisSchritt` macht neue
  Geheimnisse heute an `ergebnis.puk !== ''` fest (`:496`) — für einen PIN-only-Reset
  anzupassen.
- Voraussetzung ist der verwahrte Admin-PUK (Leitfaden Abschnitt 4 verlangt PUK und PIN
  zu verwahren). Sind beide verloren, bleibt nur der fiskaly-Support.

### What to build

- Backend: Orchestrator-Pfad „PIN per PUK zurücksetzen". Auf einer vorhandenen TSS ab
  UNINITIALIZED mit übergebenem PUK eine frische Zufalls-PIN setzen (`SetzeAdminPIN`),
  danach den bestehenden Übernahme-Lebenszyklus mit dieser PIN fortsetzen (initialisieren
  falls nötig, Client registrieren bzw. reaktivieren) und atomar speichern. Die neue PIN
  geht einmalig in die Antwort wie bei der Neuanlage, wird nie persistiert/geloggt; der
  PUK ändert sich nicht und wird nicht erneut angezeigt. Gilt in TEST und LIVE. Falscher
  PUK ergibt eine verständliche Meldung (z. B. `ErrTSESetupPUKUnbekannt`), keinen
  technischen Fehler.
- Frontend: In der PIN-Sackgasse (abgelehnte/unbekannte PIN und gesperrte PIN
  `E_ADMIN_PIN_BLOCKED`) zusätzlich „Ich habe den Admin-PUK" anbieten → PUK-Eingabefeld →
  „PIN zurücksetzen und übernehmen". Erfolg endet im `ErgebnisSchritt` mit einmaliger
  Anzeige der neuen PIN (Geheimnis-Anzeige an nicht-leerer PIN festmachen, nicht nur am
  PUK). Hinweis ergänzen, dass die PIN nach fünf Fehlversuchen sperrt.

### Acceptance criteria

- [ ] Mit korrektem PUK setzt der Wizard eine neue Zufalls-PIN, schließt die Übernahme ab
      und zeigt die neue PIN genau einmal; PUK/PIN werden nicht geloggt/persistiert
      (Unit-Test)
- [ ] Der Reset funktioniert in TEST und LIVE (Unit-Test je Umgebung)
- [ ] Eine gesperrte PIN (`E_ADMIN_PIN_BLOCKED`) führt in den PUK-Reset statt in einen
      technischen Fehler
- [ ] Falscher PUK ergibt eine verständliche Meldung mit Ausweg (fiskaly-Support), keinen
      technischen Fehler
- [ ] Wizard und Leitfaden warnen vor der 5-Versuche-Sperre; `make check` grün
