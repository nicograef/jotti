# Handoff: jotti Admin-Redesign — „Ein Festtag, ein roter Faden"

## Overview

Redesign des kompletten Admin-Bereichs von **jotti** (Gastro-Kassensystem für Vereinsfeste, Repo `nicograef/jotti`). Ausgangspunkt war ein Design-Review aller 8 Admin-Seiten. Die Design-Richtung: **Navigation und Seiten folgen dem Ablauf eines Vereinsfests** (Heute / Vorbereitung / Nach dem Fest) statt der Code-Struktur. Kernprinzipien:

1. **Nutzer-Sprache statt System-Sprache** — Zielgruppe sind ehrenamtliche Vereinskassiere, keine Entwickler.
2. **Globaler Status sichtbar** — Kasse offen? TSE signiert? Drucker ok? Immer erreichbar (Sidebar-Chip + Status-Zeile).
3. **Aktionen im Seitenkopf statt FAB** — Desktop-Pattern.
4. **Deaktivieren vor Löschen** — destruktive Aktionen ins „···"-Menü, mit Schutzregeln.
5. **Fehler dort, wo sie behoben werden** — mit Aktions-Button und Klartext.

Das Redesign bleibt **vollständig im bestehenden System**: shadcn/ui-Komponenten, Tailwind CSS 4, Olive/Emerald-Token aus `frontend/src/index.css`, Lucide-Icons, Inter Variable. Es werden keine neuen Farben, Fonts oder Radii eingeführt.

## About the Design Files

Die Dateien in diesem Paket sind **Design-Referenzen in HTML** (Design Components, `.dc.html`) — Prototypen, die Aussehen und Verhalten zeigen. Sie sind **kein Produktionscode**. Aufgabe ist es, diese Designs **im bestehenden Frontend nachzubauen**: React 19 + Vite + Tailwind CSS 4 + shadcn/ui + TypeScript, unter `frontend/src/admin/`, mit den vorhandenen Komponenten (`components/ui/*`), Hooks (`use-action-submit`, `use-form-action-submit`), Backends und Konventionen (deutschsprachige Bezeichner, zod-Schemas, TanStack Query).

`Jotti Admin Review.dc.html` enthält pro Seite ein Paar: links die 1:1-Recreation des **Status quo** (nur Referenz, NICHT bauen), rechts den **Vorschlag** (das ist die Spezifikation). Die Frames sind mit `data-screen-label="… Vorschlag …"` markiert.

## Fidelity

**High-fidelity.** Farben, Typografie, Abstände, Radii und Copy in den „Vorschlag"-Frames sind final gemeint und nutzen exakt die bestehenden Design-Token. Pixel-genau nachbauen — aber immer über die vorhandenen shadcn/ui-Komponenten und Tailwind-Klassen, nie über Inline-Styles. Alle Beispieldaten (Beträge, Namen wie „Sophie Renz", „Sommerfest Tag 2") sind Dummy-Daten und kommen in der Implementierung aus den bestehenden APIs.

---

## Screens / Views

### 0. Neue Sidebar (`AdminSidebarNeu.dc.html` — ersetzt `admin/AdminSidebar.tsx`)

- **Breite** 256 px (16 rem, wie bisher `--sidebar-width`), Hintergrund `--sidebar` (olive-50), rechter Rand `--sidebar-border`.
- **Header**: Wortmarke „jotti" 28 px / 800, linksbündig (statt 36 px zentriert). Darunter ein **Event-Status-Chip**: weiße Karte (`bg-background`, `border`, `rounded-lg`, `shadow-xs`, Padding 10×12 px) mit Event-Name (13 px / 600) und Statuszeile (12 px, `text-muted-foreground`) „Kasse offen · seit 10:02" mit 7-px-Punkt in `--primary`. Zustände: Kasse offen (grüner Punkt) / keine Kassensitzung (grauer Punkt, „Kasse geschlossen"). Chip verlinkt auf `/admin/kasse`.
- **Gruppen** (Label wie bisher: 12 px / 500, 70 % Deckkraft, Höhe 32 px):
  - **Heute**: Übersicht (`LayoutDashboard`, Route `/admin/auswertung`) · Kassentag (`Wallet`, `/admin/kasse`) — Kassentag trägt einen 7-px-Statuspunkt (grün = offen).
  - **Vorbereitung**: Produkte & Preise (`Utensils`, `/admin/produkte`) · Tische (`Lamp`, `/admin/tische`) · Helfer & Zugänge (`Users`, `/admin/benutzer`) · Bondrucker (`Printer`, `/admin/druckstationen`) — Bondrucker zeigt einen roten Punkt (`--destructive`), wenn fehlgeschlagene Druckaufträge existieren (`useFehlgeschlageneDruckauftraege`).
  - **Nach dem Fest**: Berichte & Export (`FileText`, `/admin/kassenberichte`) · Finanzamt & TSE (`Landmark`, `/admin/finanzamt`) — analog Warnpunkt, wenn TSE nicht konfiguriert / Signaturen fehlgeschlagen (`useTSEStatus`, `useTSESignaturQueue`).
  - **Service**: „Zum Service-Bereich" (`LogOut`-Icon, `/service/tische`).
- **Footer** (durch `border-t` abgetrennt): „Dunkles Design" (Theme-Toggle wie bisher), „Abmelden", darunter einzeilig zentriert 12 px muted: „jotti {version} · Nico Gräf".
- Menü-Items unverändert zum shadcn-Sidebar-Pattern: Höhe 32 px, `rounded-md`, Icon 16 px, 14 px Text; aktiv = `bg-sidebar-accent` + `font-medium`.

### 1a. Übersicht (ersetzt `reporting/AdminDashboardPage.tsx` + `LiveReportingSection.tsx`)

- **Kopf**: H1 „Übersicht" (24 px / 700), Unterzeile „{Bezeichnung} · {Wochentag, dd.MM.yyyy}" (14 px muted). Rechts: „● Live · aktualisiert HH:MM" (12 px, grüner 7-px-Punkt) + Outline-Button „Jetzt" (32 px hoch, `RefreshCw` 14 px). **Verhalten: Auto-Refresh alle 60 s** (TanStack Query `refetchInterval`), Button = sofortiger `refetch`.
- **Status-Zeile** (ersetzt die Alert-Banner): 3-spaltiges Grid, Gap 12 px. Jede Zelle `border rounded-lg` (8 px), Padding 10×14 px, Icon 16 px + Titel 13 px/600 + Unterzeile 12 px muted:
  1. Kasse: `CircleCheck` in `--primary`, „Kasse offen", „seit {Zeit} · Soll-Bestand {Betrag} €".
  2. TSE: `ShieldCheck`, „TSE signiert", „{n} Vorgänge in Warteschlange (normal)". Fehlerzustand (Rückstand ≥ 60 s oder fehlgeschlagen): rote Variante wie (3).
  3. Drucker-Fehlerzustand: Rahmen `--destructive` @ 40 %, Hintergrund `--destructive` @ 4 % auf Weiß, `TriangleAlert`, Titel in `--destructive` „1 Bon nicht gedruckt", Unterzeile „Drucker ‚Essen' prüfen", rechts **Button „Beheben"** (28 px hoch, destructive-Button-Variante) → navigiert zu `/admin/druckstationen`. Ohne Fehler: grüne Normal-Variante „Bondrucker ok".
- **Kennzahlen**: Grid `1.3fr 1fr 1fr 1fr 1fr`, Gap 12 px.
  - Hero-Karte „Kassierter Umsatz": `bg-muted` (olive-100), `rounded-xl`, Padding 18×20 px, Label 13 px muted, **Wert 34 px / 800 / letter-spacing −0.02em**, Unterzeile 12 px „bereits bezahlt, Stornos abgezogen".
  - 4 Nebenkarten (weiß, `ring-1 ring-foreground/10`, `rounded-xl`): „Noch offen" (412,00 € · „auf 7 Tischen bestellt, noch nicht bezahlt"), „Bestellt gesamt" („bezahlt + offen zusammen"), „Direktverkauf" („63 Verkäufe ohne Tisch"), „Storniert" (Label in `--destructive`). Wert 20 px / 700.
- **Zwei-Spalten-Bereich** (Grid 1fr 1fr, Gap 16 px), Karten `rounded-xl ring-1 ring-foreground/10` Padding 20 px:
  - **Offene Tische**: Kopfzeile „Offene Tische" 15 px/600 + rechts „7 Tische · 412,00 €" 13 px muted. Zeilen (14 px, `border-b` außer letzte, Padding 8 px vertikal): Tischname + Servicekraft-Kürzel muted 12 px, rechts Betrag 600. Nach 5 Zeilen Link „Alle 7 anzeigen" (Linkfarbe `--primary`). Sortierung: Saldo absteigend (kommt so vom Backend).
  - **Team** (Servicekräfte): Kopf „Team" + „3 Servicekräfte aktiv". Zeile: Name 14 px/500 + Login-Name 12 px muted; darunter 12 px Status — fertig: `CircleCheck` + „Alles abgerechnet" in `--primary`; sonst „Offen: {Betrag} auf {n} Tischen"; Stornos als Zusatz „· 2 Stornos" in `--destructive`/500. Rechts kassierter Betrag 14 px/600.
- **Storno-Zeile** (kompakt, unten): `border rounded-lg` Padding 12×16 px, `Ban`-Icon in `--destructive`, Text „**4 Stornierungen** · 36,50 € — felix 2, sophie 1, markus 1", rechts Outline-Button „Details" mit `ChevronDown`. **Verhalten**: Klick expandiert die vorhandene Storno-Detail-Liste (bestehende `StornoItem`-Komponente) inline; eingeklappt per Default.
- Leerzustand (keine Kassensitzung): bestehende `Empty`-Komponente beibehalten, Link zu `/admin/kasse`.

### 1b. Berichte & Export (ersetzt `reporting/KassenberichtePage.tsx`)

- **Kopf**: H1 „Berichte & Export", Unterzeile „Jede abgeschlossene Kassensitzung ergibt einen Tagesbericht (Z-Bon)."
- **Layout**: Grid `280px 1fr`, Gap 20 px.
- **Linke Spalte — Sitzungsliste** (ersetzt das Select): Karten je Sitzung (`border rounded-lg`, Padding 12×14 px, vertikaler Gap 8 px). Inhalt: Zeile 1 „{Wochentag, dd.MM.} · Nr. {zNr}" 13 px/600 + rechts Gesamtumsatz 13 px/600; Zeile 2 „{Bezeichnung} · abgeschlossen" 12 px muted. **Ausgewählt**: Rahmen `--primary`, Hintergrund `--primary` @ 5 %. **Offene Sitzung** (heute): 65 % Deckkraft, statt Betrag ein Status „● offen" (grün), Unterzeile „läuft — siehe Übersicht", nicht wählbar (Klick → `/admin/auswertung`). Status-Emojis (🟢🔴) entfallen ersatzlos.
- **Rechte Spalte — Bericht** (Karte `rounded-xl ring-1 ring-foreground/10`, Padding 20 px), alles untereinander sichtbar, **keine Tabs**:
  1. **Berichtskopf**: „Tagesbericht Nr. {zNr} — {Bezeichnung}" 17 px/700 + Badge `secondary` „abgeschlossen"; Meta-Zeile 13 px muted: „{Datum} · eröffnet {HH:MM} · abgeschlossen {HH:MM} von {user} · Kassensturz-Differenz {±Betrag} €". Rechts Outline-Button „Drucken" (`window.print`, Print-Stylesheet für die Berichtsspalte).
  2. **4 Kennzahl-Kacheln**: `bg-muted rounded-lg`, Padding 12×14 px, Label 12 px muted, Wert 18 px/700 („Kassierter Umsatz", „Bestellungen", „Direktverkauf", „Storniert" in rot).
  3. **Umsatz nach Steuersatz als Tabelle**: `border rounded-lg`, Kopfzeile `bg-muted` 12 px/600 muted (Steuersatz | Brutto | Netto | Steuer, Beträge rechtsbündig), Zeilen 14 px mit `border-t`, Brutto-Spalte 600. Steuersatz-Label wie `STEUERSATZ_LABEL`, optional Zusatz 12 px muted (z. B. „Getränke").
  4. **Zwei Mini-Listen** nebeneinander: „Umsatz pro Servicekraft" (Name + Login, rechts Betrag 600; Storno-Anzahl rot annotiert) und „Stornierungen (n · Betrag)" (Zeile: „HH:MM · Quelle · user · geldneutral/bar", rechts Betrag; 13 px muted).
- **Export-Block** (separate Karte unter dem Bericht, `bg-sidebar`/olive-50, `border rounded-xl`, Padding 16×20 px): `Landmark`-Icon 20 px in `--primary`, Titel „Für Steuerberater & Finanzamt" 14 px/600, Erklärtext 13 px muted: „Das DSFinV-K-Archiv ist das maschinenlesbare Kassenprotokoll dieser Sitzung. Bei einer Prüfung wird genau diese Datei verlangt — einfach herunterladen und weitergeben." Rechts **Primary-Button „Archiv herunterladen (ZIP)"** mit `Download`-Icon → bestehender `useDsfinvkExport`.

### 1c. Produkte & Preise (ersetzt `products/AdminProductsPage.tsx` + `Products.tsx` + `ProductItem.tsx`)

- **Kopf**: H1 „Produkte & Preise", Unterzeile „{n} Produkte · {m} Varianten · Änderungen wirken sofort auf allen Service-Handys". Rechts **Primary-Button „+ Neues Produkt"** (öffnet bestehenden `NewProductDialog`; FAB entfällt, ebenso `adminListBottomClearance`).
- **Gruppierung nach Kategorie** (Reihenfolge Essen, Getränke, Sonstiges): Abschnitts-Label 13 px/600/uppercase/letter-spacing 0.05em muted + Zusatz normal („· 7 % ermäßigt · Bons an Station ‚Essen'" — Steuersatz aus Produktdaten, Station aus Druckstation-Konfig).
- **Produktzeilen** in einem `border rounded-lg`-Container, Zeilen durch `border-t` getrennt, Padding 12×16 px, Flex mit Gap 14 px:
  - Produktname: feste Spalte 180 px, 14 px/600.
  - **Varianten-Chips** (flex-wrap, Gap 8 px): Pill `border rounded-full`, Padding 4 px / 12 px links / 6 px rechts, 13 px: „{Variantenname} **{Preis} €**" + Mini-Switch (24×14 px, checked `--primary`). **Inaktive Variante**: gestrichelter Rahmen, `text-muted-foreground`, Zusatzwort „aus", Switch unchecked. Switch-Klick = `aktiviereVariante`/`deaktiviereVariante` (optimistisch, wie bisher). Chip-Klick (außer Switch) öffnet `EditVariantDialog`.
  - „+ Variante"-Ghost-Pill: 24 px hoch, gestrichelter Rahmen `--input`, 12 px muted → `NewVariantDialog`.
  - Rechts: Ghost-Icon-Button `Pencil` (32 px, öffnet `EditProductDialog`) + „···"-Menü (`DropdownMenu`) mit: „Umbenennen", „Alle Varianten deaktivieren", Separator, „Löschen…" (destructive, öffnet bestehenden `AlertDialog`; **nur aktiv, wenn Produkt keine Verkäufe in einer Kassensitzung hat** — sonst disabled mit Tooltip „Produkte mit Verkäufen können nur deaktiviert werden").
- **Hinweis-Karte** unten (`bg-sidebar`, `border rounded-lg`, Padding 12×16 px, `Info`-Icon, 13 px muted): „**Ausverkauft?** Schalter aus statt löschen — die Variante verschwindet sofort von den Service-Handys, bleibt aber in allen Berichten. Löschen ist im ‚···'-Menü und nur für Produkte ohne Verkäufe möglich."
- Kategorie-Icons (Hamburger/Wine/Shell) entfallen; „Erstellt am" entfällt.

### 1d. Tische (ersetzt `tables/AdminTablesPage.tsx` + `Tische.tsx` + `TischItem.tsx`)

- **Kopf**: H1 „Tische", Unterzeile „{n} Tische · {m} aktiv · Tische mit offenem Saldo lassen sich nicht deaktivieren". Rechts Primary „+ Neuer Tisch" (bestehender Dialog).
- **Serienanlage** (neue Funktion, Karte `bg-sidebar border rounded-lg`, Padding 14×16 px): Label „Mehrere auf einmal anlegen" (12 px muted), dann Inputs: Präfix (Text, Default „Tisch", 120 px) · „Nr." · Start (56 px, zentriert) · „bis" · Ende (56 px) + Outline-Button „{n} Tische anlegen" + Live-Hinweis 12 px muted „legt ‚Tisch 15' bis ‚Tisch 24' an". **Backend**: n Einzel-Creates über bestehenden `TischBackend.createTisch` sequenziell, Fehler je Name sammeln (bereits vergebene Namen überspringen und im Toast nennen).
- **Kachel-Grid** statt Karten-Liste: nach Namens-Präfix gruppiert (alles vor der letzten Zahl; Gruppen: „Tische", „Biergarten", „Halle", sonst „Weitere"). Grid 6 Spalten, Gap 10 px. Kachel: `border rounded-lg`, Padding 12 px, Name 15 px/600, darunter Mini-Switch (24×14) + Status-Text 12 px:
  - aktiv → „aktiv" muted; deaktiviert → Kachel `bg-muted` + 75 % Deckkraft, „aus";
  - **offener Saldo** → statt „aktiv": „{Betrag} € offen" in `--primary`/500; Switch disabled (Schutzregel; Saldo aus Live-Reporting `offeneTische`).
  - Kachel-Klick öffnet Edit-Dialog (Umbenennen + „Löschen…" darin, mit gleicher Schutzregel).
- **Hinweis-Karte** unten wie 1c: „Tische mit **offenem Saldo** zeigen den Betrag an und sind gegen Deaktivieren und Löschen geschützt, bis abgerechnet wurde."

### 1e. Helfer & Zugänge (ersetzt `users/AdminUsersPage.tsx` + `Users.tsx` + `UserItem.tsx`)

- **Kopf**: H1 „Helfer & Zugänge", Unterzeile „{n} Zugänge · {m} aktiv". Rechts Primary „+ Neuer Helfer" (bestehender `NewUserDialog` + `UserCreatedDialog`).
- **Layout**: Grid `1fr 320px`, Gap 20 px.
- **Tabelle** (`border rounded-lg`): Kopfzeile `bg-muted` 12 px/600 muted: „Name · Login | Rolle | Status | (Aktionen)". Spalten `1.4fr 1fr 0.8fr auto`. Zeilen 14 px, `border-t`, Padding 11×16 px:
  - Name 500 + Login-Name 12 px `font-mono` muted. Eigener Account: Badge `bg-muted` 11 px „das bist du" (und kein Löschen im Menü).
  - **Rolle als Badge** (ersetzt Stern-Icons): Admin = Badge default (`bg-primary`), Serviceleitung = Badge outline, Service = Badge secondary.
  - Status: Mini-Switch + „aktiv"/„deaktiviert" 13 px muted (Toggle = activate/deactivateUser). Deaktivierte Zeile: `bg-sidebar`, Name muted.
  - Aktionen: Ghost `Pencil` (Edit-Dialog) + „···"-Menü: „**Passwort zurücksetzen**" (→ bestehender `PasswordResetDialog`-Flow, jetzt direkt an der Zeile), Separator, „Löschen…" (destructive, nicht bei sich selbst).
- **Rechte Spalte, 2 Panels**:
  1. „So kommt ein Helfer rein" (`bg-sidebar border rounded-lg`, Padding 16 px): 3 nummerierte Schritte (20-px-Kreise, weiß mit Border), 13 px muted: 1. „‚Neuer Helfer' anlegen — jotti erzeugt ein **Einmalpasswort**." 2. „Passwort dem Helfer zeigen (oder QR-Code scannen lassen)." 3. „Helfer meldet sich am eigenen Handy an und wählt ein eigenes Passwort."
  2. „Was Rollen dürfen" (`border rounded-lg`): „**Service** bestellt & kassiert · **Serviceleitung** darf zusätzlich stornieren · **Admin** verwaltet alles hier." + „Passwort vergessen? ‚···' → **Passwort zurücksetzen** erzeugt ein neues Einmalpasswort."

### 1f. Kassentag (ersetzt `kasse/KassensitzungPage.tsx` + Sections)

- **Kopf**: H1 „Kassentag Nr. {zNr} — {Bezeichnung}", Unterzeile „{Wochentag, dd.MM.yyyy} · Ein Kassentag läuft von der Eröffnung bis zum Tagesabschluss (Z-Bon)." Datum immer formatiert (nie ISO-String).
- **Vertikaler 3-Schritte-Stepper**: linke Spur 28 px breit; Schritt-Kreise 28 px; Verbindungslinie 2 px `--border`. Zustände: erledigt = gefüllter Kreis `--primary` mit `Check`; aktiv = weißer Kreis, 2 px Rahmen `--primary`, Ziffer in `--primary`; anstehend = weißer Kreis, Rahmen `--border`, Ziffer muted.
  1. **„1 · Kasse eröffnet"** (erledigt): flache Karte `bg-sidebar border rounded-xl`, „Heute {HH:MM} von {user} · Wechselgeld (Anfangsbestand): {Betrag} €".
  2. **„2 · Laufender Betrieb"** (aktive Karte, `ring-1 ring-foreground/10 rounded-xl`, Padding 20 px):
     - Rechts oben **Soll-Bestand groß**: 28 px/800, darunter „Soll-Bestand · Stand HH:MM" 12 px muted. (Quelle: bestehender `useKassenbestand`.)
     - **Aufschlüsselung** als 4 Kacheln `bg-muted rounded-lg` (12 px Label / 15 px 600 Wert): Anfangsbestand · + Bareinnahmen · + Einlagen · − Entnahmen. *(Benötigt neue/erweiterte Backend-Antwort mit Komponenten des Soll-Bestands — laut README existiert „aufgeschlüsselt nach Komponenten" bereits.)*
     - **Heutige Kassenbewegungen**: Kopfzeile 13 px/600 + rechts zwei Outline-Buttons „+ Geld einlegen" / „− Geld entnehmen" (32 px; öffnen einen Dialog mit dem bestehenden Geldtransit-Formular: Betrag, Kommentar; Richtung vorbelegt). Darunter Liste `border rounded-lg`, Zeilen 13 px: „{HH:MM} · {Einlage|Entnahme} · „{Kommentar}" · {user}" links muted, rechts Betrag 600 (Entnahmen in `--destructive` mit −). *(Benötigt Endpoint „Geldtransit-Liste der offenen Sitzung"; Events existieren im Journal.)*
  3. **„3 · Am Ende des Tages: Kasse abschließen"** (Karte wie 2):
     - Erklärtext 13 px muted: „Bargeld zählen, Betrag eintragen — Kassensturz und Z-Bon werden in einem Schritt gebucht. Das lässt sich nicht rückgängig machen."
     - **Warnung bei offenen Tischen** (nur wenn `offeneTische.length > 0`): rote Hinweis-Box (Rahmen `--destructive` @ 40 %, bg @ 4 %): „**7 Tische sind noch offen (412,00 €).** Erst abrechnen lassen, dann abschließen — sonst landen die Beträge als offene Posten im Bericht."
     - Formularzeile: EuroField „Gezählter Ist-Bestand" (176 px) · Outline-Button „Zählhilfe öffnen" (Dialog: Stückzahl je Münze/Schein eingeben, Summe wird übernommen — neue kleine Client-Komponente, kein Backend) · rechts **Live-Rechnung**: Soll / Gezählt / Differenz (Differenz 700, negativ in `--destructive`), aktualisiert bei jeder Eingabe (Differenz = Soll − Ist).
     - Fußzeile (über `border-t`): links 12 px muted „Kleine Differenzen sind normal und werden automatisch dokumentiert.", rechts Button **„Kasse endgültig abschließen…"** (destructive-Variante: bg `--destructive` @ 10 %, Text+Rahmen `--destructive`) → bestehender Bestätigungs-AlertDialog (Soll/Ist/Differenz + Z-Bon-Vorschau) bleibt als zweite Stufe erhalten, inkl. `signaturen_ausstehend`-Retry-Logik.
- **Leerzustand** (keine Sitzung offen): Schritt 1 wird zur aktiven Karte mit dem bestehenden Eröffnen-Formular (Bezeichnung + Anfangsbestand + TSE-Warn-Dialog), Schritte 2–3 ausgegraut.

### 1g. Bondrucker (ersetzt `settings/DruckstationConfigPage.tsx`)

- **Kopf**: H1 „Bondrucker" (endlich H1 statt h2), Unterzeile „Jede Station bekommt einen Drucker im WLAN/LAN zugewiesen. Ohne Drucker wird für die Station nichts gedruckt."
- **Fehlgeschlagene Bons ZUERST** (nur wenn vorhanden; rote Alarm-Karte, Rahmen `--destructive` @ 40 %, bg @ 4 %, `rounded-xl`, Padding 16×20 px):
  - Titel mit `TriangleAlert`: „{n} Bon(s) konnte(n) nicht gedruckt werden — die Küche hat ihn nicht!" 15 px/600 in `--destructive`.
  - Je Auftrag eine weiße Unterkarte: **Inhalt in Klartext** „Arbeitsbon Essen · Tisch 4 · 2× Pommes klein, 1× Rote Wurst" (14 px/600) + Meta „21:31 Uhr · Drucker ‚Essen' (192.168.1.50) nicht erreichbar · 6 Versuche" (13 px muted). Rechts Primary „Nochmal drucken" (`erneutVersuchen`) + Ghost/Outline „Verwerfen". *(Klartext-Positionen erfordern, dass der Endpoint Bon-Inhalt/Positionen mitliefert; Fallback: bestehende `formatDruckauftragReferenz`.)*
  - Fußnote 12 px muted: „Tipp: Erst Drucker prüfen (Strom, Netzwerk, Papier), dann ‚Nochmal drucken'." Bei > 1 Auftrag zusätzlich „Alle verwerfen" mit bestehendem Bestätigungs-Dialog.
- **Stationskarten** (Grid 2 Spalten, Gap 14 px; Karte `border rounded-xl`, Padding 18×20 px):
  - Kopf: Stationsname 15 px/600 + Kurzbeschreibung 12 px muted inline („— Bons für die Essensausgabe" / „— Bons für den Ausschank" / „— Beleg für Gäste"). Rechts **Status**: „● verbunden · letzter Druck HH:MM" (grün) oder „● nicht erreichbar" (rot). *(Status = letzter Erfolg/Fehlversuch aus Relay-Meldungen; falls kein Signal: „noch kein Druck".)*
  - Zeile: Input „Drucker-IP" (flex 1, Validierung via `validateDruckerIp`, speichert on-blur oder per Enter — der einzelne „Speichern"-Button pro Feld entfällt, Erfolg als Toast) + Outline-Button **„Testbon"** mit `Printer`-Icon → *neuer Endpoint: Testdruck an Station*.
  - **Bonmodus als erklärende Options-Karten** (nur Stationen mit `hatBonmodus`): Label „Wie sollen Bons gedruckt werden?" 12 px muted, dann 2 Radio-Karten (`rounded-lg` Padding 10×12 px; ausgewählt: Rahmen `--primary` + bg @ 5 %): „**Pro Position** — je Gericht ein Abreiß-Bon" / „**Pro Bestellung** — ein Sammelbon pro Tisch".
  - **Nicht konfigurierte Stationen** (Sonstiges, Abholbon) zusammengefasst als eine gestrichelte Karte (`border-dashed`, `bg-sidebar`): „Sonstiges & Abholbon — kein Drucker" + Kurzerklärung Abholbon + Outline-Button „Drucker zuweisen" (expandiert zur normalen Karte).

### 1h. Finanzamt & TSE (ersetzt `finanzamt/FinanzamtPage.tsx` + Sections)

- **Kopf**: H1 „Finanzamt & TSE", Unterzeile „Einmal einrichten, dann läuft es im Hintergrund. jotti erinnert dich, wenn etwas fehlt."
- **Karte 1 — „Einrichtung — {x} von 3 Schritten erledigt"** (Checkliste, 3-spaltiges Grid):
  1. **Vereinsdaten**: erledigt-Zustand = `CircleCheck` grün, Zusammenfassung (Name, Adresse einzeilig) + Link „Bearbeiten" → expandiert das bestehende Betreiber-Formular inline (oder Dialog). Unerledigt = rote Warn-Karte + Formular direkt offen.
  2. **TSE aktiv**: erledigt = grün, „Cloud-TSE (fiskaly) verbunden, Umgebung {LIVE|TEST}." + Link „Verbindung testen" (bestehender Test-Flow). Unerledigt = rote Karte „TSE einrichten" mit Primary-Button → `/admin/tse-einrichtung` (Wizard existiert).
  3. **Kasse beim Finanzamt melden**: solange offen = rote Karte (`TriangleAlert`): „Noch offen — Frist: innerhalb 1 Monat nach Inbetriebnahme, über ELSTER. (§ 146a Abs. 4 AO)" — Paragraf klein (11 px). Darin: Seriennummer als `font-mono`-Code-Pill mit Copy-Button (bestehende Kassenidentität), Links „Anleitung öffnen" (Leitfaden-URL) und „Als erledigt markieren" (**neues persistiertes Flag** `elsterGemeldetAm`; danach grüner Zustand „Gemeldet am {Datum}").
- **Karte 2 — „Läuft alles?"** (ersetzt Signaturaufträge-Metriken + Ausfalldoku als Ampel): Kopf 15 px/600 + rechts Gesamtstatus „● Ja — TSE signiert normal" (grün) / rote Variante bei Rückstand ≥ 60 s oder fehlgeschlagenen Aufträgen (Text aus bestehender Banner-Logik). Darunter 2 Panels:
  - „Signatur-Warteschlange": Klartext „{n} Vorgänge warten (ältester {t}) — normal bei vollem Betrieb. Kein Vorgang fehlgeschlagen." + Link „Technische Details" (Collapsible mit den 4 Roh-Metriken inkl. p95 — für Power-User erhalten).
  - „Störungsprotokoll": „{n} dokumentierte Störung(en) ({Datum}, {Dauer}, {Grund-Label})… Wird automatisch für die gesetzliche Ausfalldokumentation geführt." + Link „Protokoll ansehen" (Collapsible mit bestehender `StoerungRow`-Liste).
- **Karte 3 — „Gut zu wissen"** (`bg-sidebar`): 2 Link-/Info-Kacheln: „Leitfaden für Vereine" (extern) und „10 Jahre aufbewahren — Kassendaten regelmäßig sichern (§ 147 AO)".
- Manuelle TSE-Konfiguration („Experten") bleibt auf der TSE-Einrichtungsseite.

---

## Interactions & Behavior

- **Navigation**: Routen bleiben unverändert (`/admin/auswertung`, `/admin/kassenberichte`, …) — nur Labels/Gruppen der Sidebar ändern sich. Aktiv-Zustand wie bisher über `location.pathname`.
- **Auto-Refresh Übersicht**: `refetchInterval: 60_000` + manueller Refetch; „aktualisiert HH:MM" aus `dataUpdatedAt` (bestehendes `formatStand`).
- **Alle destruktiven Aktionen** (Produkt/Tisch/Helfer löschen, Kasse abschließen, Bons verwerfen): weiterhin `AlertDialog`-Bestätigung; Lösch-Einstiege nur in „···"-Dropdowns; AlertDialog-Action-Button rot wie bisher (`bg-destructive text-white`).
- **Schutzregeln**: Tisch mit offenem Saldo → Switch/Löschen disabled + Tooltip; Produkt mit Verkäufen → Löschen disabled + Tooltip; eigener Account → kein Löschen.
- **Kassentag**: Differenz live berechnen (`parseCents`, Soll − Ist); Abschluss-Dialog und 409-Retry (`signaturen_ausstehend`) unverändert übernehmen; Zählhilfe-Dialog summiert Stückzahlen × Nennwert (1 ct … 200 €) und schreibt das Ergebnis ins Ist-Bestand-Feld.
- **Toasts** (sonner) für alle Erfolge wie bisher; Fehler über `use-action-submit`/`errorMessages`.
- **Hover/Focus**: shadcn-Defaults (Buttons `hover:bg-primary/80` etc., Focus-Ring `--ring`). Keine eigenen Animationen; Collapsibles mit vorhandener `collapsible`-Komponente.
- **Responsive**: Zielgerät Desktop/Laptop. Grids brechen unter `lg` auf 1 Spalte; Sidebar-Verhalten (offcanvas + Mobile-Header) unverändert. Touch-Ziele ≥ 44 px für primäre Aktionen beibehalten.
- **Dark Mode**: keine Sonderbehandlung nötig — ausschließlich Token verwenden (`--card`, `--muted`, `--destructive` @ Alpha via Tailwind `/10`-Syntax), dann funktioniert `.dark` automatisch. Die im Design verwendeten `color-mix`-Abschwächungen entsprechen Tailwind `destructive/40`, `destructive/4`, `primary/5`.

## State Management

- Bestehende TanStack-Query-Hooks weiterverwenden: `useLiveReporting`, `useTSEStatus`, `useTSESignaturQueue`, `useFehlgeschlageneDruckauftraege`, `useOffeneKassensitzung`, `useKassenbestand`, `useAbgeschlosseneKassensitzungen`, `useReport`, `useAllProdukte/Tische/Users`, `useDruckstationen`.
- **Neu/erweitert (Backend, alle POST wie üblich)**:
  1. Kassenbestand-Komponenten (Anfangsbestand, Bareinnahmen, Einlagen, Entnahmen) — evtl. vorhanden, sonst ergänzen.
  2. Geldtransit-Liste der offenen Sitzung (Zeit, Richtung, Betrag, Kommentar, User) — aus dem Event-Journal projizierbar.
  3. Kassensitzungs-Metadaten für den Berichtskopf: eröffnet/abgeschlossen-Zeitpunkt, abschließender User, Kassensturz-Differenz, Gesamtumsatz je Sitzung (für die Sitzungsliste).
  4. Druckstation-Status (letzter erfolgreicher Druck / letzter Fehlversuch je Station) + Testdruck-Endpoint.
  5. Bon-Klartext (Positionen) am fehlgeschlagenen Druckauftrag.
  6. Flag `elsterGemeldetAm` (Betreiber-Stammdaten).
  7. Optional: offener Saldo je Tisch für die Tische-Seite (aus `tisch_sessions` vorhanden).
- Lokaler UI-State: ausgewählte Sitzung (1b), expandierte Storno-Details (1a), Collapsibles (1h), Zählhilfe-Eingaben (1f), Serienanlage-Felder (1d).

## Design Tokens

Unverändert aus `frontend/src/index.css` (Light / Dark vorhanden). Wichtigste im Design verwendete Werte:

- `--primary` emerald-700 `oklch(0.508 0.118 165.612)` · `--primary-foreground` `oklch(0.979 0.021 166.113)`
- `--destructive` red-600 `oklch(0.577 0.245 27.325)`
- `--foreground` olive-950 `oklch(0.153 0.006 107.1)` · `--muted-foreground` olive-600 `oklch(0.53 0.031 107.3)`
- `--muted`/`--accent` olive-100 `oklch(0.966 0.005 106.5)` · `--sidebar` olive-50 `oklch(0.988 0.003 106.5)`
- `--border` olive-200 `oklch(0.93 0.007 106.5)` · `--input` olive-500 `oklch(0.65 0.021 106.9)`
- `--secondary` zinc-100 `oklch(0.967 0.001 286.375)`
- Radius: `--radius: 0.45rem` → rounded-md ≈ 5.2 px (Buttons/Inputs), rounded-lg 7.2/8 px (Boxen), rounded-xl ≈ 11.2 px (Karten), rounded-full (Pills/Badges/Switches)
- Schatten: `shadow-xs` + Karten `ring-1 ring-foreground/10`
- Typo (Inter Variable): H1 24/700 · Kartentitel 15–17/600–700 · Hero-Zahl 34/800 (ls −0.02em) · Kennzahl 18–20/700 · Body 14 · Meta 12–13 muted · Abschnittslabel 13/600 uppercase ls 0.05em · Badges 12/500, Höhe 20 px
- Statuspunkte: 7–8 px Kreis in `--primary` (ok) bzw. `--destructive` (Fehler); Warnflächen: Rahmen `destructive/40`, Fläche `destructive/4`; Auswahl: Rahmen `--primary` + Fläche `primary/5`

## Assets

Ausschließlich **Lucide-Icons** (bereits als `lucide-react` im Projekt): LayoutDashboard, FileText, Utensils, Lamp, Users, Wallet, Printer, Landmark, Moon, Sun, LogOut, RefreshCw, Download, CircleCheck, ShieldCheck, TriangleAlert, Ban, Pencil, Plus, ChevronDown, Copy, Check, Info, ExternalLink. Keine Bilder, keine neuen Fonts (Inter Variable via `@fontsource-variable/inter` vorhanden).

## Files

- `Jotti Admin Review.dc.html` — Haupt-Board: pro Seite Status-quo-Recreation (Referenz), Befund-Notizen und **Vorschlag-Frame (= Spezifikation)**; Frames per `data-screen-label` benannt (z. B. „1f Vorschlag Kassentag").
- `AdminSidebarNeu.dc.html` — neue Sidebar isoliert (Prop `active`).
- `AdminSidebar.dc.html` — Status-quo-Sidebar (nur Referenz).

## Implementierungsplan (empfohlene Reihenfolge)

1. **Sidebar** (`AdminSidebar.tsx`): Gruppen/Labels/Status-Chip + Statuspunkte. Rein Frontend, sofort sichtbar. Tests in `AdminSidebar.test.tsx` anpassen.
2. **Kopfzeilen-Konvention**: einheitlicher Seitenkopf (H1 + Unterzeile + Aktions-Slot) als kleine Layout-Komponente; FABs entfernen, `adminListBottomClearance` löschen.
3. **Übersicht (1a)**: Status-Zeile, Hero-Kennzahl, Zwei-Spalten, Storno-Collapse, Auto-Refresh. Nur vorhandene Daten.
4. **Produkte (1c), Tische (1d), Helfer (1e)**: Listen-Redesigns + Schutzregeln + „···"-Menüs. Serienanlage (1d) clientseitig.
5. **Berichte (1b)**: Sitzungsliste + Berichtskopf (braucht Backend-Punkt 3) + Export-Block + Print-Stylesheet.
6. **Kassentag (1f)**: Stepper + Bestandsaufschlüsselung + Bewegungsliste (Backend 1–2) + Zählhilfe + Abschluss-Gate.
7. **Bondrucker (1g)**: Alarm-Karte oben, Stationskarten, Bonmodus-Radio-Karten; Testdruck + Status (Backend 4–5) ggf. als Folge-PR.
8. **Finanzamt (1h)**: Checkliste + Ampel + Collapsibles; `elsterGemeldetAm` (Backend 6).

Jeder Schritt ist unabhängig shipbar; bestehende Tests (`*.test.tsx`) pro Seite mitziehen.
