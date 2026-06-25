# Plan: Geldwirksames Storno und Wegfall der Auszahlung

> Source PRD: ../prds/prd-storno-rework.md

## Goal

Storno wird ein einziger, fiskalisch korrekter Vorgang. Bezahlte Positionen
werden als Warenrücknahme (negativer Umsatz am Ursprungssteuersatz, Bar-Rückgabe
im selben Beleg, Referenz auf die Zahlung) storniert, unbezahlte als
geldneutrale Korrektur. Die separate `auszahlung-geleistet` entfällt, der
Tisch-Saldo bedeutet nur noch den offenen Betrag und ist nie negativ. Die
Umbuchung wird ein eigenständiger geldneutraler Vorgang. Tisch-Storno und
Direktverkauf-Storno folgen danach demselben Modell.

## Architectural decisions

Durable decisions, die über alle Phasen gelten:

- **Event-Modell (Domain `kasse`).**
  - `stornierung-erteilt:v1` wird neu definiert: kassenwirksame Warenrücknahme
    bezahlter Positionen. Trägt Positionen (mit Steuersatz), Gesamtbetrag, genau
    eine Referenz auf die ursprüngliche Zahlung (`ZahlungID`) und einen
    Pflichtkommentar. Signiert als `Kassenbeleg-V1` mit negativem Bruttoumsatz je
    Steuersatz und negativer Bar-Zahlung.
  - `bestellung-korrigiert:v1` (neu): geldneutrale Stornierung unbezahlter
    Positionen. Trägt Positionen und einen Kommentar (optional), signiert als
    `Bestellung-V1` ohne Zahlungszeile.
  - `bestellung-umgebucht:v1` (neu): geldneutrale Umbuchung unbezahlter
    Positionen zwischen zwei Tischen. Quell- und Zielstrom erhalten verknüpfte,
    geldneutrale Einträge; ersetzt die bisherige Modellierung als Stornierung +
    Bestellung.
  - `auszahlung-geleistet:v1` entfällt vollständig.
- **Invariante 1 UI-Storno = 1..n getypte Events.** Eine „Stornieren"-Aktion
  bildet auf ein oder mehrere Events ab, jedes mit genau einer TSE-Transaktion.
  Gemischte Anforderungen werden serverseitig aufgeteilt; die Events werden
  atomar geschrieben (alles-oder-nichts, analog zum bestehenden Umbuchungs-Write).
- **Routing nach Bezahlstatus, FIFO je Zahlung.** Bezahlte Mengen werden ihren
  begleichenden Zahlungen FIFO zugeordnet (älteste Zahlung zuerst); je betroffener
  Zahlung entsteht ein `stornierung-erteilt`-Event mit genau einer `ZahlungID`.
  Der Rückgabebetrag folgt aus den Positionen, ist nicht frei wählbar.
- **Projektion `TischSession`.** `SaldoCents` ist die Summe der unbezahlten
  Positionen und nie negativ. Die Warenrücknahme entfernt die betroffenen
  Positionen aus den aktiven Listen (Ausstehend), verändert den Saldo aber nicht.
  Die geldneutrale Korrektur reduziert offenen Betrag und Ausstehend-Liste.
- **Berechtigung.** Beide Storno-Arten (`stornierung-erteilt` und
  `bestellung-korrigiert`) bleiben auf `/serviceleitung`; der gemeinsame Storno-
  Endpunkt teilt serverseitig auf. Die Umbuchung (`bestellung-umgebucht`) wandert
  auf `/service` (Servicekraft).
- **DSFinV-K-Mapper.** Warenrücknahme = negativer Beleg (`BON_TYP` Beleg, GV_TYP
  `Umsatz`, Zahlart Bar, `REF_BON_ID` auf die Zahlung, `BON_STORNO` bleibt `0`).
  `bestellung-korrigiert` und `bestellung-umgebucht` = geldneutrale `AVBestellung`
  (keine Zahlart, keine Bargeldwirkung). `direktverkauf-storniert` wird auf dasselbe
  Modell umgestellt (`BON_STORNO` 1 → 0). GV_TYP `Auszahlung` entfällt.
- **Belegausgabe.** Der Stornobeleg (negativer Betrag, Referenz auf den Ursprung)
  entsteht auf Anforderung über den bestehenden `beleg-drucken`-Endpunkt, analog
  zum Direktverkauf-Storno-Beleg.
- **Aktive Entwicklungsphase (AGENTS.md).** Events werden direkt geändert, alte
  Events nicht migriert, kein Dual-Read. DB-Schema-Änderungen (SQL-Funktionen,
  Queries) erfolgen direkt in `database/migrations/01_initial.up.sql`. Breaking
  Changes erwünscht.

## Inventory

- `backend/domain/kasse/tisch_session_events.go:14-20`: Event-Typen-Konstanten
  (`StornierungErteiltV1`, `AuszahlungGeleistetV1`).
- `backend/domain/kasse/tisch_session_events.go:57-97`: `StornierungErteiltV1Data`,
  `AuszahlungGeleistetV1Data` und ihre Schemata.
- `backend/domain/kasse/tisch_session_events.go:144-188`: Event-Konstruktoren
  (`NewStornierungErteiltEvent`, `NewAuszahlungGeleistetEvent`).
- `backend/domain/kasse/stornierung.go:9-31`, `auszahlung.go:9-29`: Historien-Structs.
- `backend/domain/kasse/tisch_session.go:28-94`: `ApplyEvent` (Storno- und
  Auszahlung-Cases bei :63-85); `:98-126`: `ComputeNichtStorniertePositionen`.
- `backend/domain/kasse/direktverkauf_events.go:37-103`: `DirektverkaufStorniertV1Data`
  und Konstruktor (Referenzmodell für den selbstständigen Storno).
- `backend/domain/kasse/historie.go:12-18,68-73`: Historien-Arten inkl. `auszahlung`.
- `backend/domain/kasse/tse_embedding.go:58-66`: Embed-Funktionen für Storno und
  Auszahlung.
- `backend/domain/kasse/kassensitzung_events.go:80-102,169-187`:
  `TagesabschlussErstelltV1Data` (Feld `AuszahlungenCents`) und Konstruktor.
- `backend/api/table/application/command.go:541-636`: `BestellungUmbuchen`
  (Storno+Bestellung über `WriteUmbuchung`); `:669-705`: `StornierungErteilen`;
  `:742-763`: `AuszahlungLeisten`; `:490-514`: `resolvePositions`.
- `backend/api/table/application/tse_signing.go:33-49`: Storno- und
  Auszahlung-Signierung (`BuildKassenbelegProcessDataWithFaktor`, `BuildEigenbelegProcessData`).
- `backend/api/table/application/kassenbeleg_command.go:176-368`:
  `KassenbelegDrucken` (Zahlung, Direktverkauf, Direktverkauf-Storno-Beleg;
  `:152-174` Negierung für Stornobeleg).
- `backend/api/direktverkauf/application/command.go:132-193`: `DirektverkaufStornieren`;
  `tse_signing.go:23-31`: Storno-Signierung.
- `backend/api/tse/application/processdata.go:16-67`: `BuildKassenbelegProcessData(WithFaktor)`;
  `:72-97`: `BuildBestellungProcessData`; `:99-121`: `BuildGeldtransitProcessData`,
  `BuildEigenbelegProcessData` (Feld-5-Bug).
- `backend/domain/dsfinvk/mapper.go:250-274`: Storno-Beleg (heute geldneutral);
  `:300-323`: Direktverkauf-Storno (`BON_STORNO=1`); `:346-352`: Auszahlung;
  `:37`: `gvTypAuszahlung`; `:423-439`: `ursprungsbons`; `:766-786`: `buildReferences`;
  `:932-938`: `gvTypReihenfolge`; `:1079-1093`: `barbestand`.
- `backend/domain/reporting/reporting.go:9-24,53-63,84-98`: Auszahlungs-Kennzahlen.
- `backend/sqlc/queries/reporting.sql:17-19,27,60,98,126-128`: Auszahlungs- und
  Storno-Aggregation; `GetAusstehendAuszahlungen`.
- `backend/sqlc/queries/kassensitzungen.sql`,
  `database/migrations/01_initial.up.sql:308-311`: `kj_extract_auszahlung_cents`.
- `backend/api/serviceleitung.go:45-47,56`: Routen Storno/Umbuchung/Auszahlung/DV-Storno;
  `backend/api/service.go:65-70`: Service-Routen.
- `backend/api/table/http/command_handler.go:345-470,528,561,653`: Request-Schemata
  (inkl. `beleg-drucken`-Stornobranch) und Handler.
- `frontend/src/service/components/table/AuszahlungDrawer.tsx`,
  `frontend/src/service/TablePage.tsx:96,160,171`,
  `frontend/src/service/table/{Auszahlung.ts,TischBackend.ts}`,
  `frontend/src/service/components/{MeinTischCard.tsx,table/TischHistorie.tsx}`,
  `frontend/src/admin/reporting/{ReportingResults.tsx,types.ts,LiveReportingSection.tsx}`,
  `frontend/src/admin/kasse/KassensitzungPage.tsx`: UI-Berührungspunkte.
- `docs/language.md:53,97,168-182,206-209,274,330-336,456`: Ubiquitous Language.

## Resolved decisions

- Beide Storno-Arten (geldneutrale Korrektur und kassenwirksame Warenrücknahme)
  bleiben auf `/serviceleitung`; der gemeinsame Storno-Endpunkt teilt serverseitig
  nach Bezahlstatus auf. Servicekräfte lösen keinen Storno aus.
- Die Umbuchung wandert auf `/service` (Servicekraft), passend zu User Stories 4–5
  und ihrem geldneutralen Charakter.
- Das Reporting zeigt beide Storno-Arten in Liste und Stornoquote; die
  kassenwirksame Warenrücknahme ist zusätzlich als Bar-Rückgabe erkennbar. Es gibt
  keine Auszahlungs-Kennzahl mehr.

## Open questions / Risks

- **TSE-Signierung mehrerer Events pro Aktion.** Ein gemischter Storno erzeugt ein
  `bestellung-korrigiert` plus n `stornierung-erteilt`. Jedes braucht eine eigene
  TSE-Transaktion (sequenzielle Signierung), alle werden atomar geschrieben.
  Schlägt die TSE während der Erfassung aus, greift der bestehende Nachsignier-
  Mechanismus je Event. Der atomare Mehrfach-Write über mehrere Subjects baut auf
  dem `WriteUmbuchung`-Muster auf (`command.go:33,618`).
- **Abstimmung mit dem Kassenabschluss-PRD.** Der Wegfall von `AuszahlungenCents`
  berührt das separat geplante Kassenabschluss-Vereinfachungs-PRD (Bereinigung der
  0-Beträge im `tagesabschluss-erstellt`-Event). Dieser Plan entfernt nur das Feld
  `AuszahlungenCents`; der weitergehende Umbau bleibt dem anderen PRD vorbehalten.
  Das Event-Schema sollte zwischen beiden Vorhaben konsistent gehalten werden.

---

## Phase 1: Eigenbeleg-`processData` Feld 5

**User stories**: 19, 22, 25 (Enabler: Kassensturzfähigkeit)

### Context

- `backend/api/tse/application/processdata.go:99-121`: `BuildEigenbelegProcessData`
  schreibt heute alle fünf Bruttofelder auf `0.00` und füllt nur die Zahlung, sodass
  Bruttosumme und Zahlung nicht balancieren.
- `backend/api/tse/application/processdata_test.go`: bestehende processData-Tests
  (unit-Build-Tag) als Prior Art.
- `backend/domain/dsfinvk/mapper.go:395-421`: Export bucht den Betrag bereits unter
  `UST_SCHLUESSEL` 5 (`geldbewegung`, `nichtSteuerbar`).

### What to build

Der gemeinsame Eigenbeleg-Builder (genutzt von Geldtransit und Kassendifferenz)
trägt den USt-neutralen Bargeldbetrag im 0-%-Feld (Feld 5), statt alle Bruttofelder
auf `0.00` zu setzen. Danach gleichen sich Bruttosumme und Bar-Zahlung im signierten
`processData` aus, und signierte `processData` stimmen mit dem Export (UST 5) überein.

### Acceptance criteria

- [x] `BuildEigenbelegProcessData` setzt das 0-%-Bruttofeld (Feld 5) auf die
      Magnitude des Bargeldbetrags; die Bar-Zahlung gleicht die Bruttosumme aus.
- [x] Geldtransit (Einlage/Entnahme) und Kassendifferenz erzeugen balancierende
      `processData`; Vorzeichen für Entnahme/Fehlbetrag korrekt.
- [x] processData-Tests decken Einlage, Entnahme und Kassendifferenz ab und prüfen
      Feld 5 sowie Zahlungszeile.
- [x] Der DSFinV-K-Export dieser Vorgänge bleibt unverändert korrekt (UST 5,
      Bargeldbestand stimmt).

---

## Phase 2: Umbuchung als eigener geldneutraler Vorgang

**User stories**: 4, 5, 6, 21

### Context

- `backend/api/table/application/command.go:541-636`: `BestellungUmbuchen` erzeugt
  heute ein `stornierung-erteilt` (Quelle) plus `bestellung-aufgenommen` (Ziel) und
  schreibt sie über `WriteUmbuchung` (`:33,618`), beide unsigniert.
- `backend/domain/kasse/tisch_session.go:28-94,98-126`: Projektion und
  `ComputeNichtStorniertePositionen` müssen den neuen Event-Typ kennen.
- `backend/domain/kasse/historie.go:37-86`: Historie braucht eine Umbuchungs-Art.
- `backend/domain/dsfinvk/mapper.go:188-274`: Mapper braucht einen Case für die
  geldneutrale Umbuchung (`AVBestellung`).
- `backend/api/service.go:65-70`, `backend/api/serviceleitung.go:46`: Endpunkt
  wandert von Serviceleitung auf Service.

### What to build

Ein neues, geldneutrales Event `bestellung-umgebucht:v1` ersetzt das Storno-plus-
Bestellung-Paar. Quell- und Zielstrom erhalten verknüpfte Einträge (gemeinsame
Umbuchungs-ID): Quelle entfernt die Positionen aus Saldo/Unbezahlt/Ausstehend, Ziel
nimmt sie geldneutral auf. Beide Einträge werden als `Bestellung-V1` signiert (keine
`:Bar`-Zeile) und atomar geschrieben. In Historie und Export erscheint der Vorgang
als „Umbuchung", nicht als „Storno". Der Endpunkt liegt auf `/service`.

### Acceptance criteria

- [x] `bestellung-umgebucht:v1` ist als geldneutraler Vorgang definiert (Domain,
      Schema, Embed); nur unbezahlte Positionen sind umbuchbar.
- [x] Quell- und Zielstrom erhalten verknüpfte Einträge, atomar geschrieben; beide
      als `Bestellung-V1` signiert, ohne `:Bar`-Zeile.
- [x] Projektion: Umbuchung verschiebt nur unbezahlte Positionen; Quell-Saldo sinkt,
      Ziel-Saldo steigt um denselben Betrag; `ComputeNichtStorniertePositionen` zählt
      korrekt.
- [x] Mapper: Umbuchung ist `AVBestellung` (geldneutral, keine Zahlart, keine
      Bargeldwirkung); Quelle und Ziel sind verknüpft erkennbar.
- [x] Historie/Export weisen den Vorgang als „Umbuchung" aus, nicht als „Stornierung".
- [x] Endpunkt `/bestellung-umbuchen` liegt auf `/service`; Frontend ruft ihn von der
      Servicekraft-Oberfläche.
- [x] Projektions-, Mapper- und Command-Tests decken die Umbuchung ab.

---

## Phase 3: Storno-Rework (Warenrücknahme + Korrektur + Routing)

**User stories**: 1, 3, 7, 8, 10, 11, 12, 14, 15, 16, 17, 18, 20, 26, 27, 29

### Context

- `backend/domain/kasse/tisch_session_events.go:57-71,144-158`: `StornierungErteiltV1Data`
  wird um die Zahlungsreferenz erweitert; neues `bestellung-korrigiert:v1`.
- `backend/domain/kasse/tisch_session.go:28-94`: Projektions-Cases: Warenrücknahme
  (Saldo unverändert, Positionen aus Ausstehend entfernen) und Korrektur (Saldo +
  Ausstehend reduzieren); `:98-126`: `ComputeNichtStorniertePositionen` für beide.
- `backend/api/table/application/command.go:669-705`: `StornierungErteilen` wird zum
  aufteilenden Routing (Bezahlstatus, FIFO je Zahlung, atomarer Mehrfach-Write);
  `:490-514`: `resolvePositions` als Basis.
- `backend/api/table/application/tse_signing.go:33-41`: Signierung: Warenrücknahme
  bleibt `Kassenbeleg-V1` mit Faktor -1; Korrektur wird `Bestellung-V1`.
- `backend/domain/dsfinvk/mapper.go:250-274,300-323,423-439,766-786`: Warenrücknahme
  als negativer Beleg mit `REF_BON_ID` auf die Zahlung (`BON_STORNO=0`); Korrektur als
  geldneutrale `AVBestellung`; Direktverkauf-Storno auf `BON_STORNO=0` angleichen.
- `backend/api/table/application/kassenbeleg_command.go:176-368`: Stornobeleg auf
  Anforderung, neuer Branch `tischId` + `stornierungId`;
  `command_handler.go:345-470`: Request-Schema erweitern.
- `backend/api/direktverkauf/application/command.go:132-193`: Direktverkauf-Storno
  bleibt funktional gleich, nur die fiskalische Darstellung (Mapper) wird angeglichen.

### What to build

Der Kern. `stornierung-erteilt:v1` wird zur kassenwirksamen Warenrücknahme bezahlter
Positionen mit Referenz auf genau eine Zahlung und Pflichtkommentar; `bestellung-
korrigiert:v1` ist die geldneutrale Stornierung unbezahlter Positionen. Eine
„Stornieren"-Anforderung wird serverseitig nach Bezahlstatus aufgeteilt: unbezahlte
Mengen in ein `bestellung-korrigiert`, bezahlte Mengen FIFO ihren Zahlungen zugeordnet
in je ein `stornierung-erteilt` pro Zahlung. Alle entstehenden Events werden atomar
geschrieben, jedes mit eigener TSE-Transaktion. Die Projektion hält den Saldo als
Summe unbezahlter Positionen (nie negativ): die Warenrücknahme entfernt Positionen aus
der Ausstehend-Liste ohne Saldoänderung, die Korrektur reduziert Saldo und Ausstehend.
Der Mapper bucht die Warenrücknahme als negativen Beleg (Bar, Referenz auf die Zahlung,
ohne Storno-Kennzeichen), die Korrektur geldneutral; der Direktverkauf-Storno wird auf
dasselbe Modell umgestellt. Auf Anforderung wird ein Stornobeleg gedruckt.

### Acceptance criteria

- [x] `stornierung-erteilt:v1` trägt Positionen, Gesamtbetrag, genau eine
      Zahlungsreferenz und einen Pflichtkommentar; signiert als `Kassenbeleg-V1` mit
      negativem Bruttoumsatz je Steuersatz und negativer Bar-Zahlung (Summe stimmt,
      inkl. Kombi-Positionen).
- [x] `bestellung-korrigiert:v1` ist geldneutral, signiert als `Bestellung-V1` ohne
      `:Bar`-Zeile.
- [x] Command-Routing: reine unbezahlte Anforderung → ein `bestellung-korrigiert`;
      reine bezahlte → ein `stornierung-erteilt` je betroffener Zahlung (FIFO, genau
      eine `ZahlungID`); gemischte → ein geldneutraler plus die kassenwirksamen Teile,
      atomar geschrieben.
- [x] Der Rückgabebetrag folgt aus den bezahlten Positionen und ist nicht frei wählbar;
      der kassenwirksame Storno ist auf `/serviceleitung` beschränkt.
- [x] Projektion über beliebige Event-Folgen: offener Betrag bleibt ≥ 0; Warenrücknahme
      lässt den offenen Betrag unverändert und entfernt die Positionen aus Ausstehend;
      Korrektur reduziert offenen Betrag und Ausstehend.
- [x] Mapper: Warenrücknahme = negativer Beleg (`BON_TYP` Beleg, GV_TYP `Umsatz`,
      Zahlart Bar, `REF_BON_ID` = Zahlung, `BON_STORNO=0`); Korrektur = geldneutrale
      `AVBestellung`; `direktverkauf-storniert` setzt `BON_STORNO=0`; summierter
      Bargeldbestand stimmt mit den tatsächlichen Bar-Bewegungen (keine Doppelbuchung).
- [x] Stornobeleg auf Anforderung über `beleg-drucken` (`tischId` + `stornierungId`),
      negativer Betrag, Referenz auf den Ursprungsbeleg, analog zum Direktverkauf-Storno.
- [x] Die signierten `processData` und der DSFinV-K-Export zeigen dieselbe Bar-Bewegung.
- [x] Bestehende Storno-UI funktioniert unverändert (eine „Stornieren"-Aktion); kein
      negativer Saldo mehr nach einem Storno bezahlter Positionen.

---

## Phase 4: Wegfall `auszahlung-geleistet`

**User stories**: 9, 13, 28

### Context

- `backend/domain/kasse/tisch_session_events.go:19,85-97,175-188`: Event-Typ, Struct,
  Konstruktor; `tisch_session.go:79-85,117`: Projektions- und Replay-Cases.
- `backend/domain/kasse/auszahlung.go`, `historie.go:17,68-73`: Struct und Historien-Art.
- `backend/domain/kasse/tse_embedding.go:63-66`: Embed-Funktion.
- `backend/domain/dsfinvk/mapper.go:37,346-352,933-938,1079-1093`: Mapper-Case,
  GV_TYP `Auszahlung`, Reihenfolge, Bargeldbestand.
- `backend/api/table/application/command.go:742-763`, `tse_signing.go:43-49`,
  `backend/api/serviceleitung.go:47`, `command_handler.go:653`: Command, Signierung,
  Route, Handler.
- `backend/domain/kasse/kassensitzung_events.go:86,99,169-187`: Feld
  `AuszahlungenCents` im Tagesabschluss-Event.
- `frontend/src/service/components/table/AuszahlungDrawer.tsx`,
  `frontend/src/service/table/{Auszahlung.ts,TischBackend.ts}`,
  `frontend/src/service/{TablePage.tsx,components/MeinTischCard.tsx}`: UI und Aufrufe.

### What to build

Der gesamte Auszahlungs-Pfad wird entfernt: Event-Typ, Struct, Konstruktor, Command,
Handler, Route, Embed, Projektions- und Replay-Cases, Mapper-Case samt GV_TYP
`Auszahlung`, Historien-Art und das Feld `AuszahlungenCents` im
`tagesabschluss-erstellt`-Event. Im Frontend entfällt der `AuszahlungDrawer` samt
Aufrufen; die Storno-Oberfläche zeigt durchgehend „Storno" ohne „Auszahlung
ausstehend"-Hinweis. Sicher, weil Phase 3 den Saldo nie negativ werden lässt.

### Acceptance criteria

- [x] Kein Vorkommen von `auszahlung-geleistet`, `AuszahlungGeleistet`,
      `AuszahlungLeisten` oder `AuszahlungDrawer` mehr im Backend/Frontend (außer in
      Reporting/SQL, das Phase 5 räumt).
- [x] `tagesabschluss-erstellt:v1` trägt kein `AuszahlungenCents` mehr; Konstruktor und
      Schema angepasst, Tagesabschluss-Erstellung baut darauf.
- [x] Mapper kennt keinen GV_TYP `Auszahlung`; Bargeldbestand und businesscases stimmen
      ohne Auszahlungs-Case.
- [x] Frontend: kein Auszahlungs-Button/Drawer, kein „Auszahlung ausstehend"-Hinweis;
      Storno-UX zeigt durchgehend „Storno"/„Stornieren".
- [x] Build, Tests und Linter (Backend + Frontend) grün.

---

## Phase 5: Reporting-Rework

**User stories**: 17, 23, 24, 25

### Context

- `backend/domain/reporting/reporting.go:9-24,53-63,84-98`: Auszahlungs-Kennzahlen
  (`AuszahlungenCents`, `GesamtAuszahlungenCents`, `AusstehendAuszahlungenCents`).
- `backend/sqlc/queries/reporting.sql:17-19,27,60,98,126-128`: Auszahlungs- und
  Storno-Aggregation, `GetAusstehendAuszahlungen`.
- `backend/sqlc/queries/kassensitzungen.sql`,
  `database/migrations/01_initial.up.sql:308-311`: `kj_extract_auszahlung_cents`.
- `backend/repository/reporting_repo/repo.go:185-269`: Storno-Aufbereitung; nimmt
  beide Storno-Arten auf.
- `frontend/src/admin/reporting/{ReportingResults.tsx,types.ts,LiveReportingSection.tsx}`,
  `frontend/src/admin/kasse/KassensitzungPage.tsx`: Reporting-UI.

### What to build

Das Reporting verliert alle Auszahlungs-Kennzahlen (`kj_extract_auszahlung_cents`,
`GetAusstehendAuszahlungen`, `AuszahlungenCents`, `AusstehendAuszahlungenCents`). Die
Stornierungsliste und die Stornoquote umfassen beide Storno-Arten
(`stornierung-erteilt` und `bestellung-korrigiert`); die kassenwirksame Warenrücknahme
ist zusätzlich als Bar-Rückgabe erkennbar. Der Z-Bon weist Warenrücknahmen als negative
Umsätze aus, nicht als Auszahlungen. Die Reporting-UI spiegelt das wider.

### Acceptance criteria

- [ ] SQL-Funktion `kj_extract_auszahlung_cents` und Query `GetAusstehendAuszahlungen`
      entfernt (Migration + Queries); keine Query referenziert `auszahlung-geleistet:v1`.
- [ ] Reporting-Domain und -Repository ohne Auszahlungs-Felder; Code kompiliert und
      sqlc-Generierung ist konsistent.
- [ ] Stornierungsliste und Stornoquote zählen beide Storno-Arten; die kassenwirksame
      Warenrücknahme ist als Bar-Rückgabe markiert.
- [ ] Z-Bon/Reporting weisen Warenrücknahmen als negative Umsätze aus; Kassenbestand
      stimmt (kassenwirksame Stornos mindern, geldneutrale Vorgänge nicht).
- [ ] Reporting-UI zeigt keine Auszahlungs-Kennzahl mehr; beide Storno-Arten sichtbar.

---

## Phase 6: Ubiquitous Language / Docs

**User stories**: 2, 28

### Context

- `docs/language.md:53,97,168-182,206-209,274,330-336,456`: Rollen, Aggregat,
  Stornierung/Auszahlung, Reporting-Begriffe, Tagesabschluss-Schema, Stornoquote.

### What to build

`docs/language.md` wird an das neue Modell angepasst: Auszahlung gestrichen,
Stornierung neu gefasst (kassenwirksame Warenrücknahme vs. geldneutrale Korrektur),
Umbuchung und `bestellung-korrigiert` ergänzt, Reporting-Begriffe und das
Tagesabschluss-Schema (ohne `auszahlungenCents`) aktualisiert. Begleitende Docs
(handbuch/ADR) nur, wo sie das alte Modell beschreiben.

### Acceptance criteria

- [x] `docs/language.md` nennt keine Auszahlung mehr; Stornierung ist als
      Warenrücknahme (bezahlt) plus geldneutrale Korrektur (unbezahlt) beschrieben.
- [x] Umbuchung (`bestellung-umgebucht`) und `bestellung-korrigiert` sind als eigene
      Vorgänge dokumentiert; Rollen-/Endpunktzuordnung stimmt.
- [x] Tagesabschluss-Schema und Reporting-Begriffe ohne Auszahlung.
- [x] Keine weitere Doc beschreibt das alte Storno-plus-Auszahlung-Modell.
