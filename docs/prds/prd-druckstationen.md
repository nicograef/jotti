# PRD: Druckstationen-Konfiguration vereinheitlichen

> **Kontext:** Diese PRD ordnet ausschließlich die **Konfiguration der Bondrucker** neu. Sie ist unabhängig von der Kassenbeleg-Inhalts-PRD ([prd-kassenbeleg.md](prd-kassenbeleg.md), F-03/F-07/F-02) — beide berühren den Bondruck, sind aber getrennt umsetzbar. Pre-Release: Schema- und API-Änderungen erfolgen direkt, ohne Migration oder Abwärtskompatibilität.

## Problem Statement

Die Drucker-Konfiguration ist heute über **zwei Konzepte** und **zwei Admin-Seiten** verstreut, deren Abgrenzung unklar ist:

- **Druckstationen** (`druckstationen`): pro Produktkategorie (Essen/Getränk/Sonstiges) eine Drucker-IP plus Bonmodus — steuert den **Arbeitsbon**. Eigene Seite `/admin/druckstationen`.
- **Bondruck-Einstellungen** (`bondruck_einstellungen`, Singleton): Kassenbeleg-Drucker-IP, ein Direktverkauf-Modus (`kein_bon` / `abholbon` / `an_stationen`) und eine Abholbon-Drucker-IP. Als Abschnitt „Bondruck" auf der Seite `/admin/einstellungen`, vermischt mit Kassenidentität und Betreiber-Stammdaten.

Dadurch entstehen mehrere Verständnisprobleme: Es ist nicht ersichtlich, warum manche Drucker „Stationen" und andere „Einstellungen" sind. Die fachlich gleichartige Frage „welcher Drucker bekommt welchen Bon?" ist auf zwei Seiten verteilt. Der Direktverkauf-Modus mischt zwei Dinge (ob überhaupt ein Bon gedruckt wird **und** wohin), und seine Variante `an_stationen` dupliziert die Arbeitsbon-Routing-Logik, was die Konfiguration zusätzlich verkompliziert.

## Solution

Alle Druckziele werden als **eine** Liste von **Druckstationen** modelliert, jede mit einem expliziten **Zweck**:

- **Arbeitsbon** — drei Stationen, weiterhin nach Produktkategorie (Essen, Getränk, Sonstiges), je mit Drucker-IP und Bonmodus.
- **Kassenbeleg** — eine Station (fiskalischer Beleg), nur Drucker-IP.
- **Abholbon** — eine Station für den Direktverkauf, nur Drucker-IP.

Damit verschwindet das separate Singleton `bondruck_einstellungen` vollständig; seine Inhalte gehen in die Druckstationen-Liste auf. Der Direktverkauf wird vereinfacht: Es gibt **genau einen** eigenen Abholbon-Drucker. Der bisherige Modus entfällt — ist die Abholbon-Drucker-IP leer, wird kein Direktverkauf-Bon gedruckt; ist sie gesetzt, wird genau ein kombinierter Abholbon (ohne Preise) gedruckt. Die Variante „an Stationen routen" entfällt.

Im Admin gibt es künftig **eine** Seite „Druckstationen", die alle fünf Stationen gruppiert nach Zweck zeigt. Die Seite „Einstellungen" enthält nur noch Kassenidentität und Betreiber-Stammdaten.

Die Druckstationen bleiben ein **fester** Satz von fünf Stationen (Seed): Der Admin bearbeitet IP und (bei Arbeitsbon) Bonmodus, kann aber keine Stationen hinzufügen oder löschen.

## User Stories

**Vereinheitlichte Übersicht**

1. Als Admin möchte ich alle Drucker an **einer** Stelle konfigurieren, damit ich nicht zwischen mehreren Seiten und Konzepten wechseln muss.
2. Als Admin möchte ich pro Druckstation deren **Zweck** (Arbeitsbon, Kassenbeleg, Abholbon) klar erkennen, damit ich verstehe, welcher Bon dort gedruckt wird.
3. Als Admin möchte ich die Druckstationen nach Zweck gruppiert sehen, damit die Liste auch bei mehreren Stationen übersichtlich bleibt.
4. Als Admin möchte ich auf der „Einstellungen"-Seite nur noch Kassenidentität und Betreiber-Stammdaten finden, damit diese Seite fokussiert bleibt.

**Arbeitsbon-Stationen (unverändert in der Funktion)**

5. Als Admin möchte ich je Produktkategorie (Essen, Getränk, Sonstiges) eine Drucker-IP setzen, damit Arbeitsbons an der richtigen Station (Küche/Theke) erscheinen.
6. Als Admin möchte ich je Arbeitsbon-Station den Bonmodus (pro Position / pro Bestellung) wählen, damit ich die Bon-Granularität steuern kann.
7. Als Admin möchte ich eine leere IP als „kein Drucker" verstehen, damit Kategorien ohne Station einfach keinen Bon erzeugen.

**Kassenbeleg-Station**

8. Als Admin möchte ich die Kassenbeleg-Drucker-IP als eigene Druckstation mit Zweck „Kassenbeleg" konfigurieren, damit der fiskalische Beleg klar einem Drucker zugeordnet ist.
9. Als Servicekraft möchte ich beim Belegdruck weiterhin eine klare Fehlermeldung erhalten, wenn keine Kassenbeleg-Station konfiguriert ist, damit ich weiß, dass der Admin sie hinterlegen muss.

**Abholbon-Station / Direktverkauf-Vereinfachung**

10. Als Admin möchte ich genau eine Abholbon-Drucker-IP für den Direktverkauf konfigurieren, damit Abholbons an der Theke gedruckt werden.
11. Als Admin möchte ich, dass eine leere Abholbon-IP „kein Direktverkauf-Bon" bedeutet, damit ich den Bondruck ohne separaten Modus-Schalter ein- und ausschalten kann.
12. Als Servicekraft möchte ich, dass ein Direktverkauf mit gesetzter Abholbon-IP genau einen kombinierten Abholbon (ohne Preise, Label „Direktverkauf") erzeugt, damit die Theke die Abholung vorbereiten kann.
13. Als Admin möchte ich, dass die frühere Modus-Variante „an Stationen" entfällt, damit die Direktverkauf-Konfiguration nur noch aus einer IP besteht und nicht mit der Arbeitsbon-Logik vermischt ist.

**Feste Stationen & Validierung**

14. Als Admin möchte ich einen festen, vorgegebenen Satz von fünf Stationen vorfinden, damit die Konfiguration einfach und vorhersehbar bleibt.
15. Als Admin möchte ich bei jeder IP eine strikte IPv4-Validierung (im Frontend und Backend) erhalten, damit Tippfehler sofort auffallen.
16. Als Admin möchte ich, dass der Bonmodus nur bei Arbeitsbon-Stationen angeboten wird, damit bei Kassenbeleg/Abholbon keine bedeutungslose Option erscheint.

**Konsistenz / Dokumentation**

17. Als Mitwirkender möchte ich, dass Handbuch und Ubiquitous-Language-Dokument das vereinheitlichte Modell widerspiegeln, damit Code und Doku übereinstimmen.

## Implementation Decisions

### Datenmodell (Schema-Änderung)

- Es bleibt **eine** Tabelle `druckstationen`. Sie wird um die Rollen-Stationen erweitert und enthält einen festen Seed von **fünf** Zeilen: `essen`, `getraenk`, `sonstiges`, `kassenbeleg`, `abholbon`.
- Jede Station trägt: einen **stabilen Schlüssel** (Stationskennung), einen **Zweck** (`arbeitsbon` | `kassenbeleg` | `abholbon`), eine **Drucker-IP** (leer = kein Drucker) und einen **Bonmodus**, der **nur** bei Arbeitsbon-Stationen gesetzt ist (sonst leer/NULL). Eine Konsistenzbedingung stellt sicher, dass Bonmodus genau dann vorhanden ist, wenn der Zweck `arbeitsbon` ist.
- Ob der Zweck als eigene Spalte gespeichert oder aus dem Stationsschlüssel abgeleitet wird, ist ein Implementierungsdetail.
- Die Tabelle `bondruck_einstellungen` entfällt **vollständig** (Schema, Queries, Repository-Methoden). Kassenbeleg- und Abholbon-IP leben nun als Stationen in `druckstationen`.
- Pre-Release: Änderung direkt in der bestehenden Initial-Migration, kein neuer Migrationspfad; Dev-DB wird neu aufgesetzt.

### Domänenmodell

- Die `Druckstation` der Domäne wird um den **Zweck** erweitert und repräsentiert künftig auch die Kassenbeleg- und Abholbon-Station.
- Das Settings-Domänenmodell `BondruckEinstellungen` und der Typ `DirektverkaufModus` (inklusive aller Modus-Konstanten) entfallen.

### Druck-Policies / Anwendungslogik

- **Arbeitsbon-Routing** (Tisch-Bestellungen) bleibt unverändert: pro Kategorie eine Station, Bonmodus wie gehabt.
- **Direktverkauf-Bondruck** wird vereinfacht: Der bisherige Modus-Switch entfällt. Ist die Abholbon-Station-IP gesetzt, entsteht **genau ein** kombinierter Abholbon (ohne Preise, festes Label „Direktverkauf"); ist sie leer, entsteht kein Auftrag. Der frühere Pfad „an Stationen routen" wird entfernt.
- **Kassenbeleg-Druck** liest die Ziel-IP künftig aus der Kassenbeleg-Station statt aus den Bondruck-Einstellungen. Der bestehende Fehler bei fehlender IP (klare Service-Fehlermeldung) bleibt erhalten.
- Die Druckauftrags-Outbox und der Relay-Transport bleiben **unverändert** (technische Warteschlange, `bon_art` weiterhin `arbeitsbon` | `kassenbeleg`).

### API-Verträge

- Ein Endpunktpaar verwaltet alle Stationen: das bestehende `get-druckstationen` / `update-druckstationen` liefert bzw. akzeptiert künftig **alle fünf** Stationen (je Station: Schlüssel/Kategorie, Zweck, Drucker-IP, Bonmodus nur bei Arbeitsbon).
- Die Endpunkte `get-bondruck-einstellungen` / `update-bondruck-einstellungen` entfallen.
- Validierung: alle IPs optional, aber strikte IPv4-Prüfung wenn gesetzt; Bonmodus nur bei Arbeitsbon-Stationen zulässig — identisch in Backend (zog) und Frontend (Zod).

### Frontend

- Die Druckstationen-Seite zeigt künftig alle fünf Stationen, nach Zweck gruppiert (Arbeitsbon: Essen/Getränk/Sonstiges mit Bonmodus; Kassenbeleg; Abholbon). Bonmodus wird nur bei Arbeitsbon angezeigt.
- Der Bondruck-Abschnitt auf der Einstellungen-Seite entfällt; die Einstellungen-Seite behält Kassenidentität und Betreiber-Stammdaten.
- Der Backend-Client für Bondruck-Einstellungen entfällt; der Druckstationen-Client wird um die zusätzlichen Stationen erweitert.
- Routen und Sidebar bleiben unverändert (Menüpunkte „Druckstationen" und „Einstellungen" bestehen weiter).

### Dokumentation / Ubiquitous Language

- Handbuch (Bondruck-Abschnitt) und Sprachdokument werden angepasst: „Druckstation" deckt nun alle Zwecke ab; der Eintrag „Direktverkauf-Modus (Bondruck)" entfällt. Die Beschreibung des Direktverkauf-Bondrucks wird auf „eigene Abholbon-Station, leere IP = kein Bon" reduziert.
- Verweise im lokalen Bondruck-Testplan auf den alten Modus werden angeglichen.

## Testing Decisions

**Was einen guten Test ausmacht:** Geprüft wird **externes Verhalten** (welche Druckaufträge entstehen, welche Stationen die API liefert/annimmt, welche Validierung greift), nicht interne Struktur.

**Zu testende Module:**

- **Direktverkauf-Bondruck-Policy (Pflicht):** Abholbon-IP gesetzt → **genau ein** Abholbon-Auftrag (ohne Preise); Abholbon-IP leer → **kein** Auftrag. Vorbild: die bestehenden Arbeitsbon-Policy-Tests.
- **Druckstationen-Repository / API (Pflicht):** Round-Trip der fünf Stationen; Bonmodus nur bei Arbeitsbon; IP-Validierung (gültig/ungültig/leer). Vorbild: die bestehenden Druckstationen- und Bondruck-Einstellungen-Tests (Letztere werden zusammengeführt).
- **Kassenbeleg-Druck (Pflicht):** Ziel-IP wird aus der Kassenbeleg-Station gelesen; fehlende IP führt weiterhin zur definierten Fehlermeldung. Vorbild: die bestehenden Kassenbeleg-Command-Tests.
- **Frontend Druckstationen-Seite (Soll):** rendert die nach Zweck gruppierten Stationen und speichert je Station; Bonmodus erscheint nur bei Arbeitsbon. Frontend Einstellungen-Seite enthält keinen Bondruck-Abschnitt mehr. Vorbild: bestehende Settings-/Druckstationen-Frontendtests.

**Regressionsschutz:** Arbeitsbon-Routing und der Relay-Transport bleiben durch ihre bestehenden Tests abgesichert; deren Verhalten ändert sich nicht.

## Out of Scope

- **Kassenbeleg-Inhalt** (Steueraufschlüsselung F-07, TSE-Felder F-02): eigene PRD ([prd-kassenbeleg.md](prd-kassenbeleg.md)).
- **Mehrere Drucker pro Kategorie / frei anlegbare Stationen:** bewusst nicht — fester Satz von fünf Stationen.
- **Drucker-Entität mit Routing-Regeln (Voll-Normalisierung):** verworfen zugunsten der Einfachheit.
- **Relay/Transport-Änderungen, ESC/POS-Formatierung der Bons:** unverändert.
- **Arbeitsbon-Routing-Logik und Bonmodus-Semantik:** unverändert (nur in die vereinheitlichte Oberfläche integriert).
- **KDS (K-13), Zubereitungsstatus (K-15):** unberührt.

## Further Notes

- Der Direktverkauf bleibt ein eigener Event-Stream ohne Stammdaten-Entität; geändert wird ausschließlich das **Bondruck-Verhalten** (eigene Abholbon-Station statt Modus). Die fachliche Direktverkauf-Logik bleibt unberührt.
- Die fiskalische Natur des Kassenbelegs liegt im **Druckauftrag** (`bon_art = 'kassenbeleg'`) und im Belegdruck-Flow, nicht in der Drucker-Konfiguration. Die Kassenbeleg-Station in derselben Liste zu führen, vermischt keine fiskalischen Eigenschaften — sie wird in der UI lediglich als eigener Zweck gruppiert.
- Abholbon-Inhalt ist unverändert: kombinierter Bon ohne Preise mit festem Label „Direktverkauf".
- Eine leere IP bedeutet einheitlich „kein Drucker" — konsistent über alle Zwecke (Arbeitsbon, Kassenbeleg, Abholbon).
