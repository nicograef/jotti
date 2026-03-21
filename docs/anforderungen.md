# Anforderungen — jotti

Dieses Dokument beschreibt alle funktionalen und querschnittlichen Anforderungen an jotti, gegliedert nach Bounded Contexts, und dient als zentrale Referenz für Entwicklung und Priorisierung.

---

## Legende

**Priorität:**

| Kürzel           | Bedeutung                                                |
| ---------------- | -------------------------------------------------------- |
| **Must-have**    | Kernfunktion, ohne die das System nicht nutzbar ist      |
| **Should-have**  | Wichtig für einen runden Betrieb, aber nicht blockierend |
| **Nice-to-have** | Komfortfunktion, die den Alltag erleichtert              |

**Status:**

| Symbol | Bedeutung  |
| ------ | ---------- |
| ✅     | Umgesetzt  |
| 🔲     | Offen      |
| 🚫     | Won't-have |

---

## Rollen und Berechtigungen

jotti kennt drei Rollen mit abgestuften Berechtigungen:

| Kurzbezeichnung    | Code-Rolle       | Beschreibung                                                                 |
| ------------------ | ---------------- | ---------------------------------------------------------------------------- |
| **Admin**          | `admin`          | Voller Zugriff auf Stammdaten (Produkte, Tische, Benutzer) und Kassenbetrieb |
| **Serviceleitung** | `serviceleitung` | Kassenbetrieb einschließlich Stornierung und Auszahlung                      |
| **Servicekraft**   | `service`        | Kassenbetrieb ohne Stornierung                                               |

### Berechtigungsmatrix

| Aktion                      | Admin | Serviceleitung | Servicekraft |
| --------------------------- | :---: | :------------: | :----------: |
| Produkte verwalten          |   ✔   |                |              |
| Tische verwalten            |   ✔   |                |              |
| Benutzer verwalten          |   ✔   |                |              |
| Passwort zurücksetzen       |   ✔   |                |              |
| Bestellung aufnehmen        |   ✔   |       ✔        |      ✔       |
| Lieferung bestätigen        |   ✔   |       ✔        |      ✔       |
| Zahlung registrieren        |   ✔   |       ✔        |      ✔       |
| Stornierung erteilen (K-04) |   ✔   |       ✔        |              |
| Auszahlung leisten (K-05)   |   ✔   |       ✔        |              |
| Umbuchung durchführen       |   ✔   |       ✔        |      ✔       |
| Tischübersicht einsehen     |   ✔   |       ✔        |      ✔       |
| Kassenjournal einsehen      |   ✔   |       ✔        |      ✔       |
| Abmelden                    |   ✔   |       ✔        |      ✔       |
| Tagesabrechnung einsehen    |   ✔   |                |              |
| Datenexport                 |   ✔   |                |              |
| Tagesabschluss einleiten    |   ✔   |                |              |

---

## 1 · Kassenbetrieb (Core Domain)

### K-01 · Bestellung aufnehmen

> **ID:** K-01 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Die Servicekraft wählt einen Tisch, stellt aus dem Produktkatalog eine Bestellung zusammen und gibt sie ab. Der Produktkatalog ist nach Kategorien (Essen, Getränke, Sonstiges) gegliedert. Pro Tisch können beliebig viele Bestellungen aufgegeben werden.

**Akzeptanzkriterien:**

- Mindestens eine Position erforderlich
- Position = Produktvariante + Menge
- Preise werden bei der Produktauswahl angezeigt (in Cent, kein Float)
- Menge änderbar per +/−-Steuerung
- Positionen vor Abgabe entfernbar (Menge auf 0)
- Gesamte Bestellung vor Abgabe verwerfbar
- Kommentar optional (max. 100 Zeichen)
- Produktkatalog nach Kategorien gruppiert, leere Kategorien ausgeblendet
- Bestellung wird als unveränderliches Event im Kassenjournal des Tisches gespeichert

### K-02 · Zahlung registrieren

> **ID:** K-02 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Die Servicekraft registriert eine Barzahlung an einem Tisch. Dabei können einzelne unbezahlte Positionen ausgewählt werden (Teilzahlung), um getrenntes Bezahlen zu ermöglichen. Der gezahlte Betrag reduziert den Saldo des Tisches.

**Akzeptanzkriterien:**

- Mindestens eine unbezahlte Position muss ausgewählt werden
- Teilzahlung möglich (Auswahl einzelner Positionen)
- Kommentar optional (max. 100 Zeichen)
- Beträge in Cent (kein Float)
- Zahlung wird als unveränderliches Event im Kassenjournal gespeichert
- Saldo des Tisches wird nach Zahlung korrekt aktualisiert

### K-03 · Lieferung bestätigen

> **ID:** K-03 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Die Servicekraft markiert bestellte Positionen als geliefert, nachdem sie dem Gast übergeben wurden. Dies dient der Nachverfolgung, welche Positionen noch offen sind.

**Akzeptanzkriterien:**

- Mindestens eine ungelieferte Position muss ausgewählt werden
- Kommentar optional (max. 100 Zeichen)
- Lieferung wird als unveränderliches Event im Kassenjournal gespeichert
- Gelieferte Positionen werden in der Tischübersicht als geliefert angezeigt

### K-04 · Stornierung erteilen

> **ID:** K-04 · **Rolle:** Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Serviceleitung oder Admin können bestellte Positionen nachträglich stornieren — unabhängig davon, ob die Positionen bereits ausgegeben oder bezahlt wurden. Einfache Servicekräfte haben kein Stornierungsrecht. Die Stornierung reduziert den Saldo des Tisches; bei bereits bezahlten Positionen kann der Saldo temporär negativ werden.

**Akzeptanzkriterien:**

- Nur Serviceleitung (`serviceleitung`) und Admin (`admin`) dürfen stornieren
- Servicekraft (`service`) hat keinen Zugriff auf die Stornierungsfunktion
- Alle nicht-stornierten Positionen sind stornierbar — unabhängig vom Ausgabe- und Bezahlstatus
- Mindestens eine Position muss ausgewählt werden
- Kommentar **erforderlich** (mind. 3, max. 100 Zeichen)
- Stornierung wird als unveränderliches `StornierungErteilt`-Event im Kassenjournal gespeichert
- Saldo des Tisches wird nach Stornierung korrekt reduziert
- Negativer Saldo (bei Stornierung bereits bezahlter Positionen) ist ein erwarteter Zustand und wird in der UI prominent hervorgehoben

### K-05 · Auszahlung leisten

> **ID:** K-05 · **Rolle:** Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Should-have

Serviceleitung oder Admin können eine Auszahlung leisten, um einen negativen Tischsaldo auszugleichen. Die Auszahlung ist eine eigenständige Operation ohne Positionsbezug — sie ist unabhängig von der Stornierung (K-04) und kann jederzeit durchgeführt werden.

**Akzeptanzkriterien:**

- Nur Serviceleitung (`serviceleitung`) und Admin (`admin`) dürfen Auszahlungen leisten
- Auszahlung ist positionsunabhängig — kein Bezug zu einzelnen Bestellpositionen
- Auszahlungsbetrag frei wählbar (≥ 1 Cent, kein Float); bei negativem Tischsaldo wird der Absolutbetrag im UI vorausgefüllt
- Kommentar **erforderlich** (mind. 3, max. 100 Zeichen)
- Auszahlung wird als unveränderliches `AuszahlungGeleistet`-Event im Kassenjournal gespeichert
- Saldo des Tisches wird nach Auszahlung korrekt erhöht (kann positiv, null oder weiterhin negativ sein)
- Negativer Saldo wird in der UI prominent hervorgehoben (Tischkarte + Tisch-Detail + Bezahlen-Tab)
- Reporting: `GetReportingStats`, `GetUmsatzProServicekraft`, `GetUmsatzProTisch` berücksichtigen Auszahlungen korrekt

### K-06 · Tischübersicht und Navigation

> **ID:** K-06 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Teilweise umgesetzt · **Prio:** Must-have

Das Service-Dashboard zeigt primär die eigenen Tische der Servicekraft als Rich Cards (→ K-14 Tisch-Favoriten). Über einen Drawer sind alle aktiven Tische erreichbar. Per Tap auf eine Tischkarte navigiert die Servicekraft zum Tisch-Detail mit drei Tabs: Bestellen, Bezahlen, Historie. Liefern ist in den Bestellen-Tab integriert; Stornieren ist für `serviceleitung`/`admin` im Bezahlen-Tab verfügbar.

**Akzeptanzkriterien:**

- Dashboard zeigt "Meine Tische" (Favoriten, K-14) als Rich Cards mit Saldo, ausstehenden Lieferungen, unbezahlten Positionen und Auszahlungsbedarf
- Leerer Zustand (keine Favoriten): Hinweis "Du hast noch keine Tische markiert" mit Button zum Alle-Tische-Drawer
- Alle aktiven Tische sind über einen Drawer ("Alle Tische") erreichbar
- Navigation per Tap auf Tischkarte zum Tisch-Detail
- Tisch-Detail bietet drei Tabs: Bestellen, Bezahlen, Historie
- Liefern ist in den Bestellen-Tab integriert, Stornieren für Serviceleitung/Admin im Bezahlen-Tab
- Operationen öffnen als Drawer (Mobile-optimiertes Overlay)
- Unbezahlte und ungelieferte Positionen sind auf dem Tisch-Detail sichtbar

### K-07 · Kassenjournal (Historie)

> **ID:** K-07 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Jeder Tisch führt ein unveränderliches Kassenjournal (Event Stream), das alle Operationen chronologisch protokolliert. Das Journal ist die einzige Quelle der Wahrheit für den Zustand eines Tisches (Event-Sourcing). Snapshots optimieren die Ladezeit bei langen Journalen.

**Akzeptanzkriterien:**

- Alle Tischoperationen (Bestellung, Zahlung, Lieferung, Stornierung) werden als Events gespeichert
- Events sind unveränderlich (append-only) — kein Update oder Löschen
- Die Historie ist pro Tisch chronologisch einsehbar
- Der aktuelle Tischzustand (Saldo, unbezahlte/ungelieferte Positionen) wird aus dem Event Stream berechnet
- Snapshots werden zur Performance-Optimierung erstellt, haben aber keine fachliche Bedeutung

### K-08 · Bezeichnung pro Bestellung

> **ID:** K-08 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** � Won't-have · **Prio:** ~~Nice-to-have~~

~~Einer Bestellung kann ein optionaler Name oder eine Bezeichnung zugewiesen werden (z. B. „Familie Müller", „Gruppe links"), um mehrere Gruppen an einem Tisch unterscheiden zu können.~~

**Entscheidung:** Wird über das bestehende Kommentarfeld der Bestellung gelöst (K-01, max. 100 Zeichen). Der Kommentar wird bereits in der Tisch-Historie (K-07) angezeigt. Ein eigenes Bezeichnungsfeld ist daher nicht notwendig.

### K-09 · Bestellungen umbuchen

> **ID:** K-09 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Serviceleitung oder Admin können eine Bestellung nachträglich auf einen anderen Tisch umbuchen, um Eingabefehler zu korrigieren. Das Umbuchen erzeugt eine Stornierung am Quell-Tisch und eine neue Bestellung am Ziel-Tisch in einer atomaren Operation.

**Akzeptanzkriterien:**

- Bestellung kann auf einen anderen Tisch umgebucht werden
- Umbuchung erzeugt eine Stornierung am Quell-Tisch und eine neue Bestellung am Ziel-Tisch
- Umbuchung erfolgt atomar (beide Operationen in einer Transaktion)

### K-10 · Rückgeldberechnung

> **ID:** K-10 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Bei der Zahlung kann die Servicekraft den vom Gast erhaltenen Bargeldbetrag eingeben. Das System berechnet und zeigt das Rückgeld an. Die Berechnung erfolgt rein clientseitig.

**Akzeptanzkriterien:**

- Eingabefeld für den erhaltenen Bargeldbetrag bei der Zahlung
- System berechnet und zeigt das Rückgeld an
- Berechnung erfolgt rein clientseitig (kein Backend-Aufruf)

### K-11 · Tisch-Schnellsuche

> **ID:** K-11 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Im Alle-Tische-Drawer (K-06) kann die Servicekraft über ein Suchfeld Tische nach Name oder Nummer filtern, um schnell zum gewünschten Tisch zu navigieren oder Favoriten zu verwalten — ohne durch die gesamte Liste scrollen zu müssen.

**Akzeptanzkriterien:**

- Suchfeld am Anfang des Alle-Tische-Drawers
- Clientseitige Filterung der Tischliste nach Tischname/-nummer (case-insensitive)
- Filtereingabe reduziert die angezeigte Tischliste in Echtzeit
- Bei leerem Suchfeld werden alle aktiven Tische angezeigt

### K-12 · Bondruck

> **ID:** K-12 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Should-have

Bons werden automatisch bei Bestellaufnahme an die zugeordneten Ausgabestationen (Küche, Getränketheke) gesendet.

**Akzeptanzkriterien:**

- Bons werden automatisch bei Bestellaufnahme gedruckt (kein manueller Schritt)
- Bon enthält: Tischname, Servicekraft, Positionen mit Mengen, Uhrzeit, optionaler Kommentar — Preise bewusst nicht (Arbeitsauftrag, keine Rechnung)
- Essenspositionen gehen an den Küchendrucker, Getränkepositionen an den Thekendrucker
- Kategorien ohne konfigurierten Drucker erzeugen keinen Bon (kein Fehler)
- Bonmodus wählbar pro Kategorie: **Pro Position** (Standard, 1 Bon je Position) oder **Pro Bestellung** (1 Sammelbon je Kategorie)
- Drucker sind im Admin-Bereich konfigurierbar (IP-Adresse und Bonmodus pro Kategorie)
- Kein Doppeldruck bei Neustart des Print-Relay
- Bei unerreichbarem Drucker wird automatisch wiederholt; andere Drucker werden nicht blockiert

### K-13 · Küchendisplay (KDS)

> **ID:** K-13 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Should-have

Mitarbeiter an den Ausgabestationen (Getränketheke, Küche) sehen auf einem eigenen Bildschirm in Echtzeit die eingehenden Bestellungen ihrer Kategorie. Das Display dient als passive Anzeige — es zeigt offene Bestellungen gruppiert nach Tisch, sodass Ausgabestationen auch bei Bon-Verlust die Bestellungen nachvollziehen können.

**Akzeptanzkriterien:**

- Echtzeit-Anzeige offener Bestellungen nach Kategorie (Essen, Getränke), gruppiert nach Tisch
- Getränkeausgabe sieht offene Getränkebestellungen, Essensausgabe sieht offene Essensbestellungen
- Letzte Bestellungen sind einsehbar (bei Bon-Verlust)

### K-14 · Tisch-Favoriten

> **ID:** K-14 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Must-have

Jede Servicekraft kann Tische als Favoriten markieren ("Meine Tische"). Favoriten sind serverseitig pro Benutzer gespeichert und bleiben über Browser-Sessions hinweg erhalten. Die markierten Tische erscheinen als Rich Cards auf dem Service-Dashboard (K-06).

**Akzeptanzkriterien:**

- Servicekraft kann einen aktiven Tisch als Favorit hinzufügen (Stern-Toggle im Alle-Tische-Drawer)
- Servicekraft kann einen Favoriten entfernen (Stern-Toggle im Alle-Tische-Drawer)
- Favoriten sind benutzerspezifisch und unabhängig von anderen Servicekräften
- Favoriten werden serverseitig in der DB gespeichert (`tisch_favoriten`-Tabelle, `user_id` + `tisch_id` als Composite PK)
- Das Hinzufügen eines bereits vorhandenen Favoriten ist idempotent (kein Fehler)
- Das Entfernen eines nicht vorhandenen Favoriten ist idempotent (kein Fehler)
- Nur aktive Tische können als Favorit markiert werden

### K-15 · Ausgabestationen mit Zubereitungsstatus

> **ID:** K-15 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Aufbauend auf dem Küchendisplay (K-13) können Mitarbeiter an Ausgabestationen den Zubereitungsstatus einzelner Positionen verwalten. Servicekräfte sehen den Zubereitungsstatus und wissen, wann Positionen abholbereit sind.

**Akzeptanzkriterien:**

- Positionen können als „in Zubereitung" und „fertig" markiert werden
- Servicekraft kann den Zubereitungsstatus ihrer Bestellungen einsehen

---

## 2 · Stammdaten (Supporting Domain)

### S-01 · Produktverwaltung

> **ID:** S-01 · **Rolle:** Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Der Admin verwaltet den Produktkatalog: Produkte anlegen, bearbeiten und entfernen. Jedes Produkt gehört zu einer Kategorie (Essen, Getränke, Sonstiges) und kann beliebig viele Varianten besitzen. Varianten haben jeweils einen eigenen Namen und Preis. Produkte und Varianten werden per Soft-Delete entfernt (Status `deleted`), sodass historische Bestellungen valide bleiben.

**Akzeptanzkriterien:**

- Produkt anlegen mit Name und Kategorie (`food`, `beverage`, `other`)
- Produkt bearbeiten (Name, Kategorie ändern)
- Produkt entfernen (Soft-Delete — Status wird auf `deleted` gesetzt)
- Variante pro Produkt anlegen mit Name und Preis in Cent
- Variante bearbeiten (Name, Preis ändern)
- Variante aktivieren / deaktivieren (Status `active` / `inactive`)
- Preise sind ganzzahlig in Cent — kein Float
- Entfernte Produkte/Varianten erscheinen nicht im Service-Produktkatalog
- Bereits in Bestellungen verwendete Varianten bleiben im Kassenjournal erhalten (Fat Events)

### S-02 · Tischverwaltung

> **ID:** S-02 · **Rolle:** Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Der Admin verwaltet die Tische der Veranstaltung: Tische anlegen, umbenennen und entfernen. Jeder Tisch hat einen Namen und einen Status (aktiv/inaktiv). Nur aktive Tische erscheinen in der Tischübersicht der Servicekräfte. Tische werden per Soft-Delete entfernt (Status `deleted`).

**Akzeptanzkriterien:**

- Tisch anlegen mit Name
- Tisch umbenennen (Name ändern)
- Tisch aktivieren / deaktivieren (Status `active` / `inactive`)
- Tisch entfernen (Soft-Delete — Status wird auf `deleted` gesetzt)
- Nur aktive Tische erscheinen in der Service-Tischübersicht
- Entfernte Tische bleiben in der Datenbank erhalten (historische Events bleiben valide)

### S-03 · Benutzerverwaltung

> **ID:** S-03 · **Rolle:** Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Der Admin verwaltet die Benutzerkonten: Benutzer anlegen, bearbeiten, aktivieren/deaktivieren und entfernen. Jedem Benutzer wird eine Rolle zugewiesen (Admin, Serviceleitung oder Servicekraft). Der Admin kann Passwörter zurücksetzen, woraufhin der Benutzer ein neues Einmalpasswort erhält. Benutzer werden per Soft-Delete entfernt (Status `deleted`).

**Akzeptanzkriterien:**

- Benutzer anlegen mit Name, Benutzername und Rolle (`admin`, `serviceleitung`, `service`)
- Benutzer bearbeiten (Name, Benutzername, Rolle ändern)
- Benutzer aktivieren / deaktivieren
- Benutzer entfernen (Soft-Delete — Status wird auf `deleted` gesetzt)
- Passwort zurücksetzen (generiert ein neues 6-stelliges Einmalpasswort)
- Deaktivierte Benutzer können sich nicht anmelden
- Entfernte Benutzer können sich nicht anmelden

---

## 3 · Auth (Infrastruktur)

### A-01 · Login

> **ID:** A-01 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Alle Benutzer melden sich mit Benutzername und Passwort an. Nach erfolgreicher Authentifizierung wird ein JWT ausgestellt, das die Rolle des Benutzers enthält. Das JWT wird bei jedem API-Aufruf mitgesendet und serverseitig geprüft. Passwörter werden mit Argon2id gehasht.

**Akzeptanzkriterien:**

- Login mit Benutzername und Passwort
- Passwörter werden mit Argon2id gehasht und verglichen
- Bei Erfolg wird ein JWT mit Benutzer-ID und Rolle ausgestellt
- JWT-Gültigkeit: 12 Stunden
- Deaktivierte oder entfernte Benutzer können sich nicht anmelden
- Bei ungültigen Zugangsdaten wird eine generische Fehlermeldung angezeigt (kein Hinweis, ob Benutzername oder Passwort falsch ist)
- Abgelaufene oder ungültige Tokens führen zur automatischen Weiterleitung auf die Login-Seite

### A-02 · Passwort setzen

> **ID:** A-02 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Neue Benutzer erhalten vom Admin ein 6-stelliges Einmalpasswort. Bei der Erstanmeldung mit dem Einmalpasswort werden sie automatisch zur Seite „Passwort setzen" weitergeleitet, wo sie ein eigenes Passwort vergeben. Dasselbe gilt nach einem Passwort-Reset durch den Admin.

**Akzeptanzkriterien:**

- Admin erstellt Benutzer → System generiert ein 6-stelliges Einmalpasswort
- Erstanmeldung mit Einmalpasswort leitet automatisch zu „Passwort setzen" weiter
- Benutzer vergibt eigenes Passwort (min. 6 Zeichen)
- Neues Passwort wird mit Argon2id gehasht gespeichert
- Nach Passwort-Reset durch den Admin wird erneut ein Einmalpasswort generiert
- Nach erfolgreichem Setzen des Passworts kann sich der Benutzer normal anmelden

### A-03 · Logout

> **ID:** A-03 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Alle Benutzer können sich über einen „Abmelden"-Button aktiv ausloggen. Der Logout entfernt das JWT aus dem Speicher und leitet auf die Login-Seite weiter. Die Funktion ist sowohl im Service- als auch im Admin-Bereich verfügbar.

**Akzeptanzkriterien:**

- Benutzer kann sich über einen „Abmelden"-Button aktiv ausloggen
- Logout entfernt den JWT aus dem Speicher
- Nach dem Logout wird der Benutzer auf die Login-Seite weitergeleitet
- Logout ist in Service- und Admin-Bereich verfügbar

---

## 4 · Querschnittsanforderungen

### Q-01 · Usability und Mobile-first

> **ID:** Q-01 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

jotti wird auf den eigenen Smartphones der ehrenamtlichen Helfer bedient (BYOD — Bring Your Own Device). Die gesamte Oberfläche ist für Smartphone-Browser und Touch-Bedienung optimiert. Tischoperationen (Bestellen, Liefern, Bezahlen, Stornieren) öffnen als Drawer (Overlay von unten), um auf kleinen Bildschirmen eine fokussierte Eingabe zu ermöglichen.

**Akzeptanzkriterien:**

- Alle Seiten sind auf Smartphone-Bildschirmen (ab 360 px Breite) vollständig bedienbar
- Schaltflächen und Eingabefelder sind touch-optimiert (ausreichend große Tippflächen)
- Tischoperationen öffnen als Drawer (Overlay von unten, mobil-optimiert)
- Keine Abhängigkeit von Maus, Hover oder Desktop-spezifischen Interaktionen
- Kein App-Download erforderlich — Zugriff ausschließlich über den Browser

### Q-02 · Mehrbenutzerfähigkeit

> **ID:** Q-02 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Mehrere Servicekräfte arbeiten gleichzeitig mit dem System — auf verschiedenen Smartphones und an verschiedenen Tischen. Das System muss parallele Zugriffe korrekt verarbeiten, ohne dass Daten verloren gehen oder inkonsistent werden.

**Akzeptanzkriterien:**

- Mehrere Benutzer können gleichzeitig angemeldet sein und arbeiten
- Parallele Bestellungen an unterschiedlichen Tischen werden korrekt und unabhängig verarbeitet
- Parallele Operationen am selben Tisch führen nicht zu Datenverlust oder inkonsistentem Zustand
- Jeder Benutzer sieht nach Neuladen den aktuellen Tischzustand

### Q-03 · Validierung

> **ID:** Q-03 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Alle Eingaben werden sowohl im Frontend als auch im Backend validiert. Das Frontend validiert mit Zod-Schemas und zeigt Fehlermeldungen direkt am Eingabefeld an. Das Backend validiert unabhängig davon mit zog-Schemas und lehnt ungültige Anfragen ab. Die doppelte Validierung stellt sicher, dass keine ungültigen Daten in die Datenbank gelangen — auch bei manipulierten Anfragen.

**Akzeptanzkriterien:**

- Frontend validiert alle Benutzereingaben vor dem Absenden (Zod-Schemas)
- Backend validiert alle eingehenden Anfragen unabhängig vom Frontend (zog-Schemas)
- Ungültige Eingaben werden im Frontend mit verständlichen Fehlermeldungen auf Deutsch angezeigt
- Ungültige Anfragen im Backend werden mit einem passenden Fehlercode abgelehnt
- Das Backend vertraut keinen Daten vom Client — es ist die letzte Validierungsinstanz

### Q-04 · Datenintegrität

> **ID:** Q-04 · **Rolle:** —
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Alle Datenänderungen sind transaktionssicher. Das Kassenjournal ist unveränderlich (append-only) — Events werden nie aktualisiert oder gelöscht. Geldbeträge werden durchgehend als ganzzahlige Cent-Werte gespeichert und verarbeitet, um Rundungsfehler durch Fließkommazahlen auszuschließen.

**Akzeptanzkriterien:**

- Alle schreibenden Operationen laufen innerhalb einer PostgreSQL-Transaktion
- Events im Kassenjournal sind unveränderlich — kein Update, kein Delete (append-only)
- Geldbeträge werden in der gesamten Anwendung als ganzzahlige Cent-Werte geführt (kein Float)
- Soft-Deletes für Stammdaten (Produkte, Varianten, Tische, Benutzer) — kein physisches Löschen
- Der Tischzustand wird ausschließlich aus dem Event Stream berechnet (Event-Sourcing)

### Q-05 · Offline-Fähigkeit

> **ID:** Q-05 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Bei einem Internetausfall während der Veranstaltung soll die Bestellaufnahme weiterhin möglich sein. Bestellungen werden lokal zwischengespeichert und bei Wiederherstellung der Verbindung automatisch synchronisiert. Auch laufende, noch nicht abgesendete Bestellungen sollen bei einem App-Absturz oder Stromausfall erhalten bleiben.

**Akzeptanzkriterien:**

- Die App bleibt bei Verbindungsverlust bedienbar (App-Shell wird aus Cache geladen)
- Bestellungen können offline aufgenommen und lokal zwischengespeichert werden
- Bei Wiederherstellung der Verbindung werden lokale Bestellungen automatisch synchronisiert
- Der Produktkatalog ist lokal gecacht und offline verfügbar
- Noch nicht abgesendete Bestellungen überleben einen App-Neustart (lokale Persistierung)
- Der Benutzer wird sichtbar über den Offline-Zustand informiert

### Q-06 · HTTPS / TLS

> **ID:** Q-06 · **Rolle:** —
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Alle Kommunikation zwischen Client und Server ist TLS-verschlüsselt. Der Reverse Proxy (nginx) terminiert TLS, und Zertifikate werden automatisch über Let's Encrypt bezogen und erneuert. HTTP-Anfragen werden auf HTTPS umgeleitet.

**Akzeptanzkriterien:**

- Alle Kommunikation zwischen Client und Server ist TLS-verschlüsselt (HTTPS)
- TLS-Zertifikate werden automatisch über Let's Encrypt bezogen und erneuert
- HTTP-Anfragen werden auf HTTPS umgeleitet
- Der Reverse Proxy (nginx) terminiert TLS

### Q-07 · Rate Limiting

> **ID:** Q-07 · **Rolle:** —
> **Status:** ✅ Umgesetzt · **Prio:** Should-have

Der Login-Endpunkt ist durch Rate Limiting geschützt, um Brute-Force-Angriffe zu erschweren. Bei Überschreitung des Limits wird eine entsprechende HTTP-Antwort zurückgegeben.

**Akzeptanzkriterien:**

- Login-Endpunkt ist durch Rate Limiting geschützt (z. B. max. 10 Versuche pro Minute pro IP)
- Bei Überschreitung wird ein HTTP 429 (Too Many Requests) zurückgegeben
- Rate Limiting ist serverseitig implementiert (nicht clientseitig)

### Q-08 · Security Headers

> **ID:** Q-08 · **Rolle:** —
> **Status:** ✅ Umgesetzt · **Prio:** Should-have

Sicherheitsrelevante HTTP-Header werden gesetzt, um gängige Angriffsvektoren (XSS, Clickjacking, MIME-Sniffing) zu mitigieren. Die Header werden vom Reverse Proxy oder Backend gesetzt.

**Akzeptanzkriterien:**

- Content Security Policy (CSP) Header wird gesetzt
- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- Strict-Transport-Security (HSTS) Header wird gesetzt
- Headers werden vom Reverse Proxy oder Backend gesetzt

---

## 5 · Reporting und Auswertung

> **Abrechnungszeitraum:** Vereinsfeste enden häufig nach Mitternacht (z. B. 17:00–03:00). Ein „Tag" im Sinne der Abrechnung entspricht daher **nicht** zwingend dem Kalendertag 0:00–24:00. Alle zeitraumbezogenen Auswertungen (R-01, R-03–R-05) beziehen sich auf einen vom Admin wählbaren **Abrechnungszeitraum** (Von–Bis). Standardmäßig wird der Zeitraum seit dem letzten Tagesabschluss (R-07) vorgeschlagen; der Admin kann ihn jederzeit manuell anpassen.

### R-01 · Tagesabrechnung

> **ID:** R-01 · **Rolle:** Admin
> **Status:** ✅ Umgesetzt · **Prio:** Should-have

Der Admin kann jederzeit ein einheitliches Reporting für einen wählbaren Abrechnungszeitraum einsehen. Die Ansicht kombiniert KPIs, Aufschlüsselungen und Stornierungsdetails in einem Datenmodell und ist die zentrale Auswertung für die Vereinsbuchhaltung.

**Akzeptanzkriterien:**

- Ein Endpoint liefert das Reporting als einheitliches Datenmodell (KPIs, Breakdown-Sektionen, Stornierungen)
- Abrechnungszeitraum ist wählbar (Von–Bis als Datum + Uhrzeit); Default im UI: heute 00:00 bis jetzt
- Zeitraumsemantik ist eindeutig: `von` inklusiv, `bis` exklusiv, UTC-only
- Anzeige der wichtigsten Kennzahlen auf einen Blick: Gesamtumsatz, Anzahl offener Tische, Anzahl Bestellungen und Stornierungen
- Gesamtbetrag der Bestellungen, der Stornierungen und offener Saldi im Zeitraum
- Umsatz pro Servicekraft und Umsatz pro Tisch als Übersichtslisten
- Übersicht aller Stornierungen mit Zeitpunkt, Tisch, stornierten Positionen und Betrag
- Kein Live-Dashboard und kein Auto-Refresh erforderlich; Auswertung erfolgt bewusst on-demand
- Nur durch Admin einsehbar

### R-02 · Datenexport

> **ID:** R-02 · **Rolle:** Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Der Admin kann Umsätze, Bestellungen und Artikeldaten als CSV exportieren, um sie extern weiterverarbeiten zu können (z. B. für die Vereinsbuchhaltung). Der Export bezieht sich auf den gewählten Abrechnungszeitraum.

**Akzeptanzkriterien:**

- Export von Umsätzen, Bestellungen und Artikeldaten als CSV
- Export bezieht sich auf den gewählten Abrechnungszeitraum
- Export jederzeit durch den Admin auslösbar

### R-03 · Abrechnung pro Tisch

> **ID:** R-03 · **Rolle:** Admin
> **Status:** 🔲 Offen · **Prio:** Should-have

Der Admin kann eine detaillierte Abrechnung pro Tisch einsehen. Diese enthält alle Bestellungen, Zahlungen, Lieferungen und Stornierungen des jeweiligen Tisches in chronologischer Reihenfolge sowie einen Gesamt-Saldo — gefiltert auf den gewählten Abrechnungszeitraum.

**Akzeptanzkriterien:**

- Detaillierte Aufstellung aller Bestellungen, Zahlungen, Lieferungen und Stornierungen eines Tisches im Abrechnungszeitraum
- Anzeige des Gesamt-Saldos (bestellt, bezahlt, offen, storniert)
- Abrufbar für jeden einzelnen Tisch durch den Admin
- Chronologische Reihenfolge der Ereignisse

### R-04 · Abrechnung pro Servicekraft

> **ID:** R-04 · **Rolle:** Admin
> **Status:** 🔲 Offen · **Prio:** Should-have

Der Admin kann eine personenbezogene Abrechnung pro Servicekraft für den gewählten Abrechnungszeitraum einsehen. Diese zeigt den kassieren Umsatz, die Anzahl und Summe der aufgegebenen Bestellungen, Stornierungen sowie den Vergleich Bestellt vs. Kassiert — für die Endabrechnung und Nachvollziehbarkeit.

**Akzeptanzkriterien:**

- Umsatz pro Servicekraft im Abrechnungszeitraum (Summe aller von dieser Person registrierten Zahlungen)
- Bestellvolumen pro Servicekraft (Anzahl und Summe der aufgegebenen Bestellungen)
- Anzahl und Betrag der Stornierungen pro Servicekraft
- Gegenüberstellung: Bestellsumme vs. kassierte Summe pro Person (Differenz = offene Beträge)
- Nur durch Admin einsehbar

### R-05 · Produktumsatz-Reporting

> **ID:** R-05 · **Rolle:** Admin
> **Status:** 🔲 Offen · **Prio:** Should-have

Der Admin kann Auswertungen über Produktumsätze im gewählten Abrechnungszeitraum einsehen: verkaufte Mengen pro Produkt und Variante, ein Ranking der meistverkauften Varianten sowie Gesamteinnahmen pro Produkt.

**Akzeptanzkriterien:**

- Übersicht über verkaufte Mengen pro Produkt und Variante im Abrechnungszeitraum
- Ranking der meistverkauften Varianten
- Gesamteinnahmen pro Produkt/Variante
- Nur durch Admin einsehbar

### R-06 · Eigene Übersicht (Servicekraft)

> **ID:** R-06 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Jede Servicekraft sieht auf dem Service-Dashboard eine kompakte KPI-Sektion mit ihren eigenen Aktivitäten: Anzahl und Summe der eigenen Bestellungen sowie Anzahl und Summe der eigenen kassierten Zahlungen. Die Karten bieten einen schnellen persönlichen Überblick über den eigenen Beitrag, ohne dass ein Admin benötigt wird.

**Akzeptanzkriterien:**

- KPI-Sektion direkt auf dem Service-Dashboard sichtbar (unterhalb des Headers, oberhalb der Tischkarten)
- Karte "Bestellungen": Anzahl eigener Bestellungen + Gesamtsumme in Euro
- Karte "Kassiert": Anzahl eigener Zahlungen + Gesamtsumme in Euro
- Nur eigene Daten sichtbar — kein Einblick in Daten anderer Servicekräfte
- Werte werden aus den Events der Servicekraft berechnet (Event-Store-Query gefiltert auf `user_id`)

### R-07 · Tagesabschluss

> **ID:** R-07 · **Rolle:** Admin
> **Status:** 🔲 Offen · **Prio:** Should-have

Der Admin leitet am Ende einer Veranstaltung einen Tagesabschluss ein. Dabei wird der aktuelle Abrechnungszeitraum abgeschlossen, ein Abschlussbericht generiert und das System für den nächsten Veranstaltungstag vorbereitet. Der Tagesabschluss dient als Schnittstelle zur Vereinsbuchhaltung und als Grundlage für die Auszahlung an Servicekräfte.

**Hinweis:** Das Zurücksetzen wirft eine offene Frage zur Event-Sourcing-Kompatibilität auf — Events werden nicht gelöscht (Append-only-Prinzip), sondern der Tagesabschluss markiert eine logische Zäsur im Event Stream.

**Akzeptanzkriterien:**

- Admin kann einen Tagesabschluss einleiten
- Offene Tische (Saldo ≠ 0) werden vor Abschluss angezeigt und müssen bestätigt werden
- Abschlussbericht wird generiert (entspricht Tagesabrechnung R-01 für den abgeschlossenen Zeitraum)
- Tagesabschluss setzt den Default-Abrechnungszeitraum für nachfolgende Auswertungen auf „ab jetzt"
- Abgeschlossene Zeiträume bleiben über R-01 weiterhin abrufbar (Archiv)
- Optional: Tisch-Saldi können auf 0 zurückgesetzt werden (erzeugt ein Abschluss-Event pro Tisch, kein Löschen)

---

## 6 · Bewusste Abgrenzung

jotti ist kein Allzweck-Kassensystem. Folgende Features sind bewusst **nicht** enthalten — jedes zusätzliche Feature erhöht Komplexität, Wartungsaufwand und Einarbeitungszeit für ehrenamtliche Teams.

| Feature                            | Kurzbegründung                                                                                                                  |
| ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| 🚫 Kartenzahlung / Zahlungsgateway | Vereinsfeste arbeiten mit Bargeld. Kartenterminals verursachen Kosten, Verträge und technische Abhängigkeiten.                  |
| 🚫 TSE / KassenSichV               | Gemeinnützige Vereine unterliegen in der Regel keiner TSE-Pflicht. Event-Sourcing erfüllt die GoBD-Grundsätze bereits.          |
| 🚫 Reservierungssystem             | Vereinsfeste haben keine Tischreservierungen — Gäste setzen sich, wo Platz ist.                                                 |
| 🚫 Warenwirtschaft / Inventory     | Bestandsverwaltung ist für temporäre Veranstaltungen mit überschaubarem Sortiment unverhältnismäßig aufwändig.                  |
| 🚫 Lieferservice-Integration       | jotti deckt ausschließlich Vor-Ort-Gastronomie ab, nicht Liefer- oder Abholservice.                                             |
| 🚫 Multi-Standort-Verwaltung       | Jede Veranstaltung betreibt eine eigene jotti-Instanz. Eine standortübergreifende Verwaltung wird nicht benötigt.               |
| 🚫 CRM / Kundendatenbank           | Vereinsfeste haben keine wiederkehrenden Kundenbeziehungen, die ein CRM rechtfertigen.                                          |
| 🚫 Kiosk-Modus / Self-Order        | Self-Order erhöht die Systemkomplexität erheblich und widerspricht dem persönlichen Service durch ehrenamtliche Helfer.         |
| 🚫 Gast-Benachrichtigung           | Gäste sitzen am Tisch und werden persönlich bedient — Push-Benachrichtigungen an Gäste sind im Vereinsfest-Kontext überflüssig. |
| 🚫 Trinkgeld-Tracking              | Bei ehrenamtlichen Veranstaltungen ist Trinkgeld-Verwaltung unüblich und unnötig komplex.                                       |

Für eine detaillierte Gegenüberstellung mit kommerziellen POS-Systemen siehe [Produktbeschreibung — Abgrenzung](produktbeschreibung.md#7-abgrenzung).
