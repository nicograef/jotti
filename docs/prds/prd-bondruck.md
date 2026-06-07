# PRD: Bondruck — Neuordnung (Arbeitsbon, Kassenbeleg, Outbox)

> **Scope:** Dieses PRD ordnet die Bon-Domäne für den **bedienten Tisch** neu und baut
> die gemeinsame Druck-Infrastruktur (Outbox, Relay als Transport). Es ist
> **self-contained** und sofort umsetzbar. Die **Direktverkauf-Bons** (Abholbon,
> Stations-Routing, Direktverkauf-Kassenbeleg) sind **nicht** Teil dieses PRDs — sie
> konsumieren diese Infrastruktur und sind in `prd-direktverkauf.md` spezifiziert.

## Problem Statement

Heute vermischt jotti zwei fachlich entgegengesetzte Dinge unter einem einzigen
Begriff „Bon":

1. **Der operative Arbeitsbon.** Nach einer Tischbestellung soll in der Küche (Essen)
   bzw. an der Getränketheke automatisch ein Bon mit den bestellten Produkten
   gedruckt werden — je nach Konfiguration pro Position oder pro Bestellung. Dieser
   Bon ist eine reine **interne Arbeitsanweisung**: Was muss zubereitet/ausgegeben
   werden? Er trägt **keine Preise** und ist rechtlich bedeutungslos.

2. **Der gesetzliche Kassenbeleg.** Nach dem **Kassieren** muss dem Gast ein Beleg
   **angeboten** werden (§ 146a Abs. 2 AO, § 6 KassenSichV). Dieser Beleg ist eine
   Quittung/Rechnung: alle Positionen mit Preisen, Vereinsdaten, Kassen-Seriennummer
   und — nach TSE-Anbindung — kryptografische Prüfwerte. Auf einem Vereinsfest wird er
   **selten** gebraucht und soll daher nur **auf ausdrückliche Anforderung** durch den
   Service gedruckt werden.

Der aktuelle Code kennt **nur Familie 1**. Der Print-Relay reagiert ausschließlich
auf `bestellung-aufgenommen:v1`; einen Druck nach dem Kassieren gibt es nicht.
Trotzdem behauptet `docs/anforderungen.md` (F-03): „Beleg wird nach jeder Zahlung
automatisch an den Bondrucker gesendet" — das ist **faktisch falsch** und erzeugt
eine gefährliche **Schein-Compliance**: Der vorhandene Arbeitsbon sieht aus wie ein
Beleg, erfüllt aber keine einzige Anforderung aus § 6 KassenSichV. Die
Konfigurationstabelle heißt generisch `kategorie_drucker`/„Bondruck", konfiguriert
aber in Wahrheit nur Arbeitsbon-Stationen.

Zusätzlich bringt das geplante Feature **Direktverkauf** (Theke, bezahlen + ausgeben
in einem Schritt) eine weitere Quelle ins Spiel, die ebenfalls beide Bon-Familien
speist. Solange „Bon" ein undifferenzierter Sammelbegriff bleibt, lässt sich auch
dieser Fall nicht sauber einordnen. Dieses PRD schafft die saubere Grundlage; die
konkreten Direktverkauf-Bons (Abholbon, Stations-Routing) bauen darauf auf und werden
in `prd-direktverkauf.md` spezifiziert.

### Kernerkenntnis

> Es gibt nicht „drei Bon-Typen", sondern **zwei fachlich getrennte Bon-Familien** —
> den **operativen Arbeitsbon** (nicht-fiskalisch, automatisch, ausgelöst durch das
> Entstehen von Ware) und den **gesetzlichen Kassenbeleg** (fiskalisch, auf
> Anforderung, ausgelöst durch eine Zahlung). Beide teilen nur die Outbox als
> Transport. Der spätere Direktverkauf ist **kein** dritter Typ, sondern nur eine
> zusätzliche **Quelle**, die beide Familien speist (→ `prd-direktverkauf.md`).

| Merkmal      | **Arbeitsbon** (operativ)                                             | **Kassenbeleg** (fiskalisch)                                              |
| ------------ | --------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| Zweck        | Was zubereiten/ausgeben?                                              | Zahlungsnachweis für den Gast (§ 146a AO)                                 |
| Auslöser     | Ware entsteht (Tischbestellung)                                       | Geld fließt (Zahlung)                                                     |
| Modus        | **automatisch** nach Konfiguration                                    | **auf Anforderung** (am Fest selten)                                      |
| Inhalt       | Quelle (Tisch), Artikel, Menge, Kommentar, Uhrzeit — **keine Preise** | Alle Positionen + Preise + Vereinsdaten + Kassen-ID (+ später Steuer/TSE) |
| Empfänger    | Küchen-/Thekenpersonal                                                | Gast                                                                      |
| Rechtsstatus | irrelevant (KassenSichV regelt ihn nicht)                             | **der** gesetzliche Beleg                                                 |
| Heute        | implementiert (heißt nur „Bon")                                       | dokumentiert, **nicht** implementiert                                     |

## Solution

Die Bon-Domäne wird entlang der zwei Familien neu geordnet, mit einer einheitlichen
Druck-Infrastruktur (Outbox) darunter:

- **Arbeitsbon (operativ, nicht-fiskalisch).** Automatischer Druck an konfigurierte
  **Druckstationen** (Küche/Theke) beim Entstehen von Ware. Inhalt unverändert zum
  heutigen Bon (Quelle, Artikel, Menge, Kommentar, Uhrzeit, Bedienung — keine Preise).
  Ausgelöst durch `bestellung-aufgenommen:v1`.

- **Kassenbeleg (fiskalisch, § 146a).** Wird **nur auf Anforderung** durch den Service
  gedruckt — pro Kassiervorgang. Enthält Vereinsdaten (Betreiber-Stammdaten, K-20),
  Kassen-Seriennummer (F-01), Datum/Uhrzeit, alle Positionen mit Einzelpreisen,
  Gesamtbetrag und Zahlungsart (bar). **Steuer-Aufschlüsselung** (F-07) und
  **TSE-Pflichtfelder** (F-02) werden als ausgewiesene Folgeschritte ergänzt, sobald
  diese Grundlagen existieren — der Basis-Beleg ist davon unabhängig druckbar.

- **Druckauftrags-Outbox (`druckauftraege`).** Eine append-orientierte Tabelle ist die
  **Single Source of Truth** für alle Druckaufträge — operativ **und** auf Anforderung.
  Das Backend reiht fertige Aufträge ein (Ziel-IP + ESC/POS-Payload + Status); der
  Relay **leert** die Queue (poll → drucken → quittieren). Damit verschwindet das
  fragile „Cursor + lokale JSON-State"-Modell des Relays; der Relay wird zum reinen
  **Transport**. Die ESC/POS-Formatierung liegt im Backend — sie wandert aus dem
  Relay-API-Paket in den neuen Bondruck-Bereich; der Relay-Dienst (`cmd/relay`)
  formatiert ohnehin nichts und erhält weiterhin fertige Payloads.

- **Druckstation statt „Bondruck-Konfiguration".** `kategorie_drucker` wird zu
  **`druckstationen`** umbenannt und ehrlich benannt: ein Drucker je Produktkategorie
  für **Arbeitsbons**. Der Kassenbeleg-Drucker liegt in einer eigenen Einstellung
  **`bondruck_einstellungen`** (Singleton).

## Ubiquitous Language (neue/geänderte Begriffe)

| Begriff          | Bedeutung                                                                             | bisher                     |
| ---------------- | ------------------------------------------------------------------------------------- | -------------------------- |
| **Arbeitsbon**   | Operativer, nicht-fiskalischer Bon an eine Ausgabestation. Keine Preise. Automatisch. | „Bon" / „Bondruck"         |
| **Druckstation** | Konfigurierter Drucker an einem Ausgabeort (Küche, Theke), je Produktkategorie.       | `kategorie_drucker`        |
| **Kassenbeleg**  | Fiskalischer Zahlungsbeleg (§ 146a AO) für den Gast, **auf Anforderung** gedruckt.    | dokumentiert, fehlte       |
| **Druckauftrag** | Konkreter Druckjob (Ziel-IP + ESC/POS-Payload + Status) in der Outbox.                | transientes `DruckAuftrag` |
| **Bondruck**     | Oberbegriff/Policy-Bereich, der beide Familien + Outbox umfasst.                      | (uneinheitlich verwendet)  |

**Abgrenzung:** Der **Arbeitsbon** ist **niemals** ein Beleg im Sinne von § 146a AO.
Die Belegausgabepflicht hängt am **Kassiervorgang**, nicht an der Bestellung/Ausgabe.

## User Stories

### Arbeitsbon — Tischbestellung (Familie 1, operativ)

1. Als Küchenkraft möchte ich nach jeder Essensbestellung automatisch einen Arbeitsbon
   mit Tisch, Artikel, Menge und Kommentar erhalten, damit ich weiß, was zuzubereiten ist.
2. Als Thekenkraft möchte ich nach jeder Getränkebestellung automatisch einen Arbeitsbon
   erhalten, damit ich die Getränke ausgeben kann.
3. Als Betreiber möchte ich pro Produktkategorie (Essen/Getränk/Sonstiges) eine
   Druckstation (IP + Modus) konfigurieren, damit jeder Bon am richtigen Ort landet.
4. Als Betreiber möchte ich je Druckstation zwischen „ein Bon pro Position" und „ein
   Sammelbon pro Bestellung" wählen, damit der Druck zu meinem Arbeitsablauf passt.
5. Als Betreiber möchte ich eine Kategorie ohne Drucker (leere IP) lassen, damit nur
   dort gedruckt wird, wo ich es will.
6. Als Küchenkraft möchte ich, dass der Arbeitsbon **keine Preise** zeigt, weil er eine
   Arbeitsanweisung und kein Beleg ist.
7. Als Betreiber möchte ich, dass eine Konfigurationsänderung für künftige Bestellungen
   wirkt, ohne dass bereits erfasste Aufträge verloren gehen.

### Kassenbeleg — auf Anforderung (Familie 2, fiskalisch)

8. Als Servicekraft möchte ich nach dem Kassieren einer Tischzahlung **auf Knopfdruck**
   einen Kassenbeleg drucken, **nur wenn** der Gast ihn verlangt, damit ich nicht bei
   jedem Vorgang unnötig Papier produziere.
9. Als Gast möchte ich einen Beleg erhalten, der Vereinsname und -adresse, Datum/Uhrzeit,
   alle Positionen mit Preisen, den Gesamtbetrag und die Zahlungsart (bar) ausweist,
   damit ich eine gültige Quittung habe.
10. Als Gast möchte ich, dass der Beleg die Kassen-Seriennummer trägt (§ 6 KassenSichV),
    damit er den Pflichtangaben entspricht.
11. Als Servicekraft möchte ich den Kassenbeleg für einen bestimmten Kassiervorgang
    (eine konkrete Zahlung) drucken, nicht für „den ganzen Tisch", damit der Beleg
    genau einen Geschäftsvorfall abbildet.
12. Als Servicekraft möchte ich den Kassenbeleg bei Bedarf erneut drucken können (z. B.
    Papierstau), ohne den Kassiervorgang zu wiederholen.
13. Als Betreiber möchte ich den Kassenbeleg-Drucker getrennt von den Arbeitsbon-
    Stationen konfigurieren, weil er an der Kasse/Theke steht und an den Gast geht.
14. Als Compliance-Verantwortlicher möchte ich, dass der Kassenbeleg später nahtlos um
    Steuer-Aufschlüsselung (F-07) und TSE-Pflichtfelder (F-02) erweitert werden kann,
    ohne das Druckmodell neu zu bauen.

### Druck-Infrastruktur (Outbox & Relay)

15. Als Betreiber möchte ich, dass ein einmal erfasster Druckauftrag auch dann gedruckt
    wird, wenn der Relay kurz nicht erreichbar war, damit keine Bons verloren gehen.
16. Als Entwickler möchte ich, dass der Relay **keine** fachliche Logik (ESC/POS,
    Kategorien, Bonmodus) mehr kennt, sondern nur Aufträge abholt, druckt und quittiert,
    damit Transport und Fachlichkeit getrennt sind.
17. Als Betreiber möchte ich, dass der Druckstatus in der Datenbank steht (offen/gedruckt),
    nicht in einer lokalen Relay-Datei, damit ein Relay-Neustart oder Dateiverlust
    keinen Doppel- oder Nichtdruck verursacht.
18. Als Entwickler möchte ich, dass sowohl automatische Arbeitsbons als auch
    angeforderte Kassenbelege über **denselben** Outbox-Mechanismus laufen, damit es
    nur einen Druckpfad gibt.
19. Als Betreiber möchte ich, dass ein nicht erreichbarer Drucker den Auftrag nicht
    still verschluckt, sondern (wie bisher) mit Wiederholungen versucht und protokolliert.

### Konfiguration & Verwaltung

20. Als Admin möchte ich die Druckstationen (Kategorie → IP + Modus) in einer klar
    benannten Oberfläche pflegen, die ehrlich „Druckstationen" und nicht „Bondruck"
    heißt.
21. Als Admin möchte ich den Kassenbeleg-Drucker an einer eigenen, getrennten Stelle
    (getrennt von den Arbeitsbon-Stationen) einstellen, damit operative und
    belegbezogene Konfiguration nicht vermischt werden.
22. Als Admin möchte ich, dass IP-Adressen beidseitig validiert werden (zog + Zod),
    damit Fehleingaben früh auffallen.

### Compliance & Korrektheit

23. Als Compliance-Verantwortlicher möchte ich, dass die Dokumentation klar zwischen
    nicht-fiskalischem Arbeitsbon und fiskalischem Kassenbeleg unterscheidet, damit keine
    Schein-Compliance entsteht.
24. Als Compliance-Verantwortlicher möchte ich, dass die falsche F-03-Behauptung
    („Beleg wird nach jeder Zahlung automatisch gedruckt") korrigiert wird, damit der
    Status den tatsächlichen Code widerspiegelt.
25. Als Betreiber möchte ich in der Betreiber-Doku den Hinweis finden, dass die
    Belegausgabe am Fest unter die Befreiung „Verkauf an eine Vielzahl unbekannter
    Personen" (§ 146a Abs. 2 Satz 2 AO) fallen kann — der Beleg aber jederzeit
    **erstellbar** sein muss und die Aufzeichnung lückenlos bleibt.

### Robustheit / Randfälle

26. Als Servicekraft möchte ich, dass ein Kassenbeleg nur für einen tatsächlich
    kassierten Vorgang erzeugt werden kann, damit kein Beleg ohne Zahlung entsteht.
27. Als Betreiber möchte ich, dass ein Kassenbeleg auch ohne konfigurierten Drucker
    angefordert werden kann und dann eine klare Rückmeldung „kein Kassenbeleg-Drucker
    konfiguriert" erscheint, statt still zu scheitern.
28. Als Entwickler möchte ich, dass das Umbenennen (`kategorie_drucker` → `druckstationen`)
    das Verhalten der Tisch-Arbeitsbons nicht verändert, damit der Rename risikolos ist.

## Implementation Decisions

### Domäne / Begriffe

- **Zwei Bon-Familien als Leitprinzip.** Arbeitsbon (operativ, nicht-fiskalisch) und
  Kassenbeleg (fiskalisch). Sie teilen **keinen** Trigger, **keinen** Inhalt, **keinen**
  Formatter und **keinen** Rechtsstatus — nur die Outbox als Transport.
- **Bondruck wird ein eigener Anwendungsbereich** im Backend (Policy + On-Demand-Command
  - ESC/POS-Formatter + Outbox-Repository). Die ESC/POS-Formatierung wandert aus dem
    Relay-API-Paket (`backend/api/relay/application/escpos`) in den neuen Bondruck-Bereich;
    der Relay-Dienst (`cmd/relay`) formatiert ohnehin nichts und bleibt reiner Transport.
- **Voller Rename.** `kategorie_drucker` → `druckstationen`; das Drucker-Konfig-Paket,
  die Admin-Routen, die Response-DTOs und die Frontend-Klasse/Seite werden auf
  „Druckstation" umgestellt. Die `ProduktKategorie`-Enum bleibt unverändert.

### Druckauftrags-Outbox (`druckauftraege`)

- **Neue Tabelle** als Single Source of Truth für Druckjobs. Felder (Vorschlag):
  `id`, `ziel_ip`, `payload` (Base64-ESC/POS), `status` (`offen` | `gedruckt`),
  `bon_art` (`arbeitsbon` | `kassenbeleg` — `abholbon` kommt mit `prd-direktverkauf.md`),
  `referenz` (Event-Bezug, optional), `erstellt_am`, `gedruckt_am`.
- **Backend reiht ein, Relay leert.** Kein „compute at poll" mehr. Der Payload wird
  beim Einreihen **eingefroren** (unkritisch für den Arbeitsbon; beim Kassenbeleg als
  Basis-Beleg deterministisch rekonstruierbar, ab TSE-Phase fachlich zwingend).
- **Status in der DB ist autoritativ.** Der Relay-State (lokale JSON-Datei mit Cursor +
  Idempotenzliste) entfällt. Idempotenz ergibt sich aus `status = 'offen'`.
- **Append-only-Geist, aber kein Kassenjournal.** Die Outbox ist **kein** fiskalisches
  Journal — sie ist eine technische Druck-Warteschlange. Der einzige erlaubte
  Status-Übergang ist `offen → gedruckt` (per Quittierung).

### Relay (Transport)

- **Protokoll neu:** `POST /relay/poll` liefert **offene Druckaufträge** (`id`,
  `zielIp`, `payload`). Ein neuer `POST /relay/quittieren` meldet erfolgreich gedruckte
  IDs zurück; das Backend setzt `status = 'gedruckt'` + `gedruckt_am`.
- **Relay-Logik schrumpft** auf: offene Jobs holen → drucken (bestehende
  Drucker-Erreichbarkeits-/Retry-Logik bleibt) → quittieren. Kein ESC/POS, keine
  Kategorien, kein Cursor, keine State-Datei.

### Arbeitsbon-Policy (automatisch)

- **Eine Policy reiht Arbeitsbons ein**, ausgelöst durch `bestellung-aufgenommen:v1`.
  Sie lädt die Druckstationen, gruppiert Positionen nach Kategorie, erzeugt ESC/POS und
  schreibt `druckauftraege`-Zeilen. Die bestehende Gruppierungs-/Formatierlogik wird
  nahezu unverändert übernommen.
- **Enqueue-Zeitpunkt:** unmittelbar nach erfolgreichem Event-Write durch denselben
  Application-Command. (Transaktionale Kopplung vs. Post-Commit-Enqueue ist die zentrale
  Implementierungsentscheidung — siehe Plan.)

### Kassenbeleg (auf Anforderung)

- **Neuer Service-Command** `KassenbelegDrucken` (Rollen: `admin`, `serviceleitung`,
  `service`): nimmt eine Referenz auf **einen Kassiervorgang** (eine konkrete
  `zahlung-kassiert:v1`-Zahlung), lädt die
  Positionen (Fat Event), die Betreiber-Stammdaten und die Kassen-Seriennummer, erzeugt
  den Kassenbeleg-ESC/POS und reiht **einen** `druckauftraege`-Eintrag für den
  Kassenbeleg-Drucker ein. Endpunkt `POST /service/beleg-drucken`.
- **Beleginhalt jetzt:** Vereinsname + Adresse (Betreiber, K-20), Kassen-Seriennummer
  (Kassenidentitaet, F-01), Datum/Uhrzeit, Positionen (Artikel, Menge, Einzelpreis,
  Zeilensumme), Gesamtbetrag, Zahlungsart „bar", Bon-Nummer.
- **Beleginhalt später (ausgewiesen, nicht jetzt):** Steuer-Aufschlüsselung pro
  Steuersatz (hängt an **F-07**, da `Position` heute **kein** `Steuersatz`-Feld hat) und
  TSE-Pflichtfelder (hängt an **F-02**, Druck **nach** `FinishTransaction`).
- **Fehlende Drucker-Konfiguration** → klare Fehlermeldung (kein stilles Verschlucken,
  anders als beim optionalen Arbeitsbon).

### Konfiguration

- **`druckstationen`** (3 Kategorie-Zeilen, wie bisher `kategorie_drucker`): `kategorie`
  (PK), `drucker_ip`, `bonmodus` (`pro_position` | `pro_bestellung`). Genutzt von den
  Tisch-Arbeitsbons.
- **`bondruck_einstellungen`** (Singleton): `kassenbeleg_drucker_ip`. Folgt dem
  bestehenden Singleton-Muster (`betreiber`, `kassenidentitaet`). Der Direktverkauf
  erweitert dieses Singleton später um `direktverkauf_modus` und `abholbon_drucker_ip`
  (siehe `prd-direktverkauf.md`).
- **Validierung beidseitig** (zog + Zod): IPv4-Regex wie heute.

### HTTP / API (POST-only)

- Admin: `…/get-druckstationen`, `…/update-druckstation` (Rename der bisherigen
  drucker-Routen); `…/get-bondruck-einstellungen`, `…/update-bondruck-einstellungen`.
- Service: `…/beleg-drucken`.
- Relay: `…/poll` (Semantik neu: offene Jobs), `…/quittieren` (neu).
- **Response-DTOs in der HTTP-Schicht**; Domain-Modelle nie direkt serialisiert.

### Frontend

- **`DruckstationBackend`** (umbenannt aus `DruckerBackend`) auf Basis des
  `BackendClient`-Interfaces; neue `BondruckEinstellungenBackend`-Methoden.
- **Druckstationen-Konfigseite** (umbenannt aus `DruckerConfigPage`); neue
  Bondruck-Einstellungen-Sektion (Kassenbeleg-Drucker).
- **„Beleg drucken"-Aktion** im Service nach dem Kassieren bzw. in der Tisch-Historie
  pro Zahlung. Kein direktes `fetch()`.

### Dokumentation (zu aktualisieren)

- `docs/language.md` — neue Begriffe (Arbeitsbon, Druckstation, Kassenbeleg,
  Druckauftrag); „DruckerKonfiguration" → „Druckstation". **(In Phase 0 erledigt.)**
- `docs/anforderungen.md` — K-12 → „Arbeitsbon" geschärft; **F-03 korrigiert** (kein
  Auto-Druck nach Zahlung; stattdessen On-Demand-Kassenbeleg mit Pflichtfeldern).
  **(In Phase 0 erledigt.)**
- `docs/handbuch.md` — §4.6 neu (zwei Familien, Outbox, Relay als Transport); §3.12
  Policy „Arbeitsbon-Druck"; Outbox-Tabelle ergänzen. **Mit dem jeweiligen Slice**, nicht
  vorab (sonst beschreibt die Doku Ungebautes).
- `docs/compliance.md` — Arbeitsbon ausdrücklich als nicht-fiskalisch kennzeichnen;
  Kassenbeleg = § 146a-Beleg; Belegausgabe-Befreiung am Fest (Betreiberpflicht).
  **Mit dem Kassenbeleg-Slice.**

## Testing Decisions

Tests prüfen **externes Verhalten** an der öffentlichen Modul-Schnittstelle, nicht
Implementierungsdetails. Priorisiert:

1. **ESC/POS-Formatter (Backend, isoliert) — Hauptkandidat.** Arbeitsbon (pro_position
   / pro_bestellung, keine Preise, Kommentar, Quelle) und Kassenbeleg (Vereinsdaten,
   Kassen-ID, Positionen mit Preisen, Gesamtsumme, „bar"). Reine Funktionen, hervorragend
   testbar. Prior Art: bestehender `escpos/formatter_test.go`.
2. **Arbeitsbon-Policy.** Aus einem `bestellung-aufgenommen`-Event entstehen die
   richtigen Druckaufträge je Kategorie/Modus; Kategorien ohne Druckstation werden
   übersprungen. Prior Art: bestehender `relay/application/query_test.go`.
3. **Kassenbeleg-Command.** Beleg nur für tatsächlich kassierten Vorgang; fehlender
   Kassenbeleg-Drucker → klarer Fehler; genau **ein** Druckauftrag pro Anforderung.
   Prior Art: `api/table/application/command_test.go`.
4. **Outbox-Repository + Relay (Integrationstest).** Einreihen → `poll` liefert offene
   Jobs → `quittieren` setzt `gedruckt` → erneuter `poll` liefert sie nicht mehr;
   Reprint nur über erneutes Einreihen. Prior Art: `relay/relay_integration_test.go`,
   `repository/kassenjournal_repo/repo_test.go`.
5. **Trennungs-Guard.** Ein Arbeitsbon enthält nie Preise; ein Kassenbeleg wird nie
   automatisch durch eine Bestellung ausgelöst. Beide Pfade teilen nur die Outbox.
6. **Frontend.** „Beleg drucken" löst genau einen Backend-Aufruf aus; Druckstationen-
   Seite rendert/speichert wie zuvor (Rename ohne Verhaltensänderung). Prior Art:
   bestehende Service-/Admin-Komponententests, `routes.test.ts`.

Mit dem Nutzer abzustimmen: ob alle Bereiche getestet werden oder der Fokus zunächst
auf dem Backend-Kern (1–4) liegt.

## Out of Scope

- **TSE-Pflichtfelder auf dem Kassenbeleg** (Signatur, Transaktionsnummer,
  Signaturzähler, TSE-Seriennummer) — hängt an **F-02**; Druck erst **nach**
  `FinishTransaction`.
- **Steuer-Aufschlüsselung auf dem Kassenbeleg** (Netto/Steuer pro Satz) — hängt an
  **F-07** (`Position` hat heute kein `Steuersatz`-Feld).
- **eBeleg / QR-Code** (digitaler Beleg, § 146a Abs. 2) — F-09, spätere Phase.
- **Küchendisplay (KDS, K-13)** und Zubereitungsstatus (K-15) — eigene Anforderung.
- **Belegausgabe-Befreiungsantrag** beim Finanzamt — Betreiberpflicht, gehört in die
  Betreiber-Doku, nicht in Code.
- **DSFinV-K-Export** (F-04) — eigene Phase; die Outbox ist kein Export.
- **Direktverkauf-Bons** (Abholbon, Stations-Routing, `direktverkauf_modus`,
  Direktverkauf-Kassenbeleg) — konsumieren diese Infrastruktur, sind aber in
  `prd-direktverkauf.md` spezifiziert.
- **Mehrere benannte Verkaufsstellen / Theken** — bleibt ausgeschlossen (siehe
  `prd-direktverkauf.md`).

## Further Notes

- **Verhältnis zur Direktverkauf-PRD.** Dieses PRD ist **self-contained** und sofort
  umsetzbar (Tisch-Arbeitsbon via Outbox, Kassenbeleg auf Anforderung, Druckstation-
  Rename, Relay-Transport, Doku-Korrektur). Es liefert die **Infrastruktur**, die der
  Direktverkauf konsumiert: Outbox, Arbeitsbon-Policy, Kassenbeleg-Command und das
  `bondruck_einstellungen`-Singleton. Die Direktverkauf-Bons (Abholbon, Stations-Routing,
  `direktverkauf_modus`, Direktverkauf-Kassenbeleg) sind in `prd-direktverkauf.md`
  spezifiziert und setzen dessen Events voraus.
- **Warum Outbox statt „compute at poll".** Der heutige Pull rechnet beim Poll aus den
  Events die Payloads — ideal für „automatisch, alle neuen Bestellungen", aber
  ungeeignet für „drucke genau diesen Beleg jetzt, weil der Gast ihn will". Die Outbox
  vereinheitlicht automatisch **und** auf Anforderung in **einem** Mechanismus und legt
  den Status revisionsfest in die DB. Der **Arbeitsbon**-Payload wird beim Einreihen
  eingefroren (unkritisch). Der **Kassenbeleg** wird als Basis-Beleg deterministisch aus dem
  immutablen `zahlung-kassiert`-Event rekonstruiert; das Einfrieren des Payloads **zum
  Zahlzeitpunkt** wird erst mit der TSE-Signatur (F-02) fachlich zwingend — die Outbox
  ist darauf bereits vorbereitet.
