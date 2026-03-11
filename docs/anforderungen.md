# Anforderungen — jotti

Dieses Dokument beschreibt alle funktionalen und querschnittlichen Anforderungen an jotti, gegliedert nach Bounded Contexts, und dient als zentrale Referenz für Entwicklung und Priorisierung.

Für Produktidentität, Zielgruppe und Abgrenzung siehe [Produktbeschreibung](produktbeschreibung.md). Für Architektur, Domänenmodell und technische Designentscheidungen siehe [System Design](design.md). Für die kanonischen Fachbegriffe siehe [Ubiquitous Language](language.md).

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
| **Serviceleitung** | `senior_service` | Kassenbetrieb einschließlich Stornierung                                     |
| **Servicekraft**   | `service`        | Kassenbetrieb ohne Stornierung                                               |

### Berechtigungsmatrix

| Aktion                  | Admin | Serviceleitung | Servicekraft |
| ----------------------- | :---: | :------------: | :----------: |
| Produkte verwalten      |   ✔   |                |              |
| Tische verwalten        |   ✔   |                |              |
| Benutzer verwalten      |   ✔   |                |              |
| Passwort zurücksetzen   |   ✔   |                |              |
| Bestellung aufgeben     |   ✔   |       ✔        |      ✔       |
| Lieferung bestätigen    |   ✔   |       ✔        |      ✔       |
| Zahlung registrieren    |   ✔   |       ✔        |      ✔       |
| Stornierung durchführen |   ✔   |       ✔        |              |
| Tischübersicht einsehen |   ✔   |       ✔        |      ✔       |
| Kassenjournal einsehen  |   ✔   |       ✔        |      ✔       |

---

## 1 · Kassenbetrieb (Core Domain)

### K-01 · Bestellung aufgeben

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

### K-04 · Stornierung

> **ID:** K-04 · **Rolle:** Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Serviceleitung oder Admin können bestellte Positionen nachträglich stornieren. Einfache Servicekräfte haben kein Stornierungsrecht. Die Stornierung reduziert den offenen Saldo des Tisches.

**Akzeptanzkriterien:**

- Nur Serviceleitung (`senior_service`) und Admin (`admin`) dürfen stornieren
- Servicekraft (`service`) hat keinen Zugriff auf die Stornierungsfunktion
- Mindestens eine Position muss ausgewählt werden
- Kommentar optional (max. 100 Zeichen)
- Stornierung wird als unveränderliches Event im Kassenjournal gespeichert
- Saldo des Tisches wird nach Stornierung korrekt reduziert

### K-05 · Tischübersicht und Navigation

> **ID:** K-05 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Die Servicekraft sieht auf der Startseite alle aktiven Tische als Karten und navigiert per Tap zum Tisch-Detail. Dort stehen alle Tischoperationen (Bestellen, Liefern, Bezahlen, Stornieren, Historie) als Tabs zur Verfügung. Tischoperationen öffnen als Drawer (Overlay von unten).

**Akzeptanzkriterien:**

- Alle aktiven Tische werden als Karten angezeigt
- Tischkarte zeigt den aktuellen Saldo
- Navigation per Tap zum Tisch-Detail
- Tisch-Detail bietet Tabs für alle Operationen (Bestellung, Lieferung, Zahlung, Stornierung, Historie)
- Operationen öffnen als Drawer (Mobile-optimiertes Overlay)
- Unbezahlte und ungelieferte Positionen sind auf dem Tisch-Detail sichtbar

### K-06 · Kassenjournal (Historie)

> **ID:** K-06 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Jeder Tisch führt ein unveränderliches Kassenjournal (Event Stream), das alle Operationen chronologisch protokolliert. Das Journal ist die einzige Quelle der Wahrheit für den Zustand eines Tisches (Event-Sourcing). Snapshots optimieren die Ladezeit bei langen Journalen.

**Akzeptanzkriterien:**

- Alle Tischoperationen (Bestellung, Zahlung, Lieferung, Stornierung) werden als Events gespeichert
- Events sind unveränderlich (append-only) — kein Update oder Löschen
- Die Historie ist pro Tisch chronologisch einsehbar
- Der aktuelle Tischzustand (Saldo, unbezahlte/ungelieferte Positionen) wird aus dem Event Stream berechnet
- Snapshots werden zur Performance-Optimierung erstellt, haben aber keine fachliche Bedeutung

### K-07 · Bezeichnung pro Bestellung

> **ID:** K-07 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Einer Bestellung kann ein optionaler Name oder eine Bezeichnung zugewiesen werden (z. B. „Familie Müller", „Gruppe links"), um mehrere Gruppen an einem Tisch unterscheiden zu können. Die Bezeichnung wird in der Tisch-Historie und bei der Abrechnung angezeigt.

**Akzeptanzkriterien:**

- Bestellung kann mit optionalem Namen versehen werden (z. B. „Familie Müller", „Gruppe links")
- Bezeichnung wird in der Tisch-Historie angezeigt
- Bezeichnung wird bei der Abrechnung angezeigt

### K-08 · Bestellungen umbuchen

> **ID:** K-08 · **Rolle:** Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Serviceleitung oder Admin können eine Bestellung nachträglich auf einen anderen Tisch umbuchen, um Eingabefehler zu korrigieren. Das Umbuchen erzeugt eine Stornierung am Quell-Tisch und eine neue Bestellung am Ziel-Tisch in einer atomaren Operation.

**Akzeptanzkriterien:**

- Bestellung kann auf einen anderen Tisch umgebucht werden
- Umbuchung erzeugt eine Stornierung am Quell-Tisch und eine neue Bestellung am Ziel-Tisch
- Umbuchung erfolgt atomar (beide Operationen in einer Transaktion)

### K-09 · Rückgeldberechnung

> **ID:** K-09 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Bei der Zahlung kann die Servicekraft den vom Gast erhaltenen Bargeldbetrag eingeben. Das System berechnet und zeigt das Rückgeld an. Die Berechnung erfolgt rein clientseitig.

**Akzeptanzkriterien:**

- Eingabefeld für den erhaltenen Bargeldbetrag bei der Zahlung
- System berechnet und zeigt das Rückgeld an
- Berechnung erfolgt rein clientseitig (kein Backend-Aufruf)

### K-10 · Tisch-Schnellsuche

> **ID:** K-10 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Auf der Tischübersicht kann die Servicekraft über ein Suchfeld oder Nummernpad direkt eine Tischnummer eingeben, um schnell zum gewünschten Tisch zu navigieren — ohne durch die Karten scrollen zu müssen.

**Akzeptanzkriterien:**

- Suchfeld oder Nummernpad auf der Tischübersicht
- Direkte Navigation zum gesuchten Tisch per Eingabe der Tischnummer

### K-11 · Bondruck

> **ID:** K-11 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Should-have

Bons werden automatisch oder manuell gedruckt, damit Ausgabestationen (Küche, Getränketheke) die bestellten Positionen erhalten. Bons enthalten alle relevanten Informationen: Tisch, Servicekraft, Positionen mit Mengen, Zeitstempel und optionalen Kommentar. Freibons mit freier Preiseingabe ermöglichen Sonderpositionen außerhalb des Produktkatalogs.

**Akzeptanzkriterien:**

- Bons sind übersichtlich formatiert (Tisch, Servicekraft, Positionen, Mengen, Gesamtpreis, Uhrzeit, Kommentar)
- Separater Bon pro Position druckbar (z. B. bei Fehlbon Nachdruck einer einzelnen Position)
- Getränkepositionen werden automatisch an den Getränkedrucker gesendet
- Essenspositionen werden automatisch an den Küchendrucker gesendet
- Freibon mit freier Bezeichnung und Preiseingabe möglich (Sonderpositionen)
- Drucker sind vom Admin konfigurierbar (Zuordnung Drucker zu Kategorie)

### K-12 · Ausgabestationen

> **ID:** K-12 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Mitarbeiter an den Ausgabestationen (Getränketheke, Küche) sehen auf einem eigenen Bildschirm die offenen Bestellungen ihrer Kategorie. Sie können Bestellungen als „in Zubereitung" und „fertig" markieren. Servicekräfte sehen den Zubereitungsstatus und wissen, wann Positionen abholbereit sind.

**Akzeptanzkriterien:**

- Getränkeausgabe sieht offene Getränkebestellungen, gruppiert nach Tisch
- Essensausgabe sieht offene Essensbestellungen, gruppiert nach Tisch
- Ausgabestationen können Positionen als „in Zubereitung" und „fertig" markieren
- Servicekraft kann den Zubereitungsstatus ihrer Bestellungen einsehen
- Letzte Bestellungen sind an der Ausgabestation einsehbar (bei Bon-Verlust)

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

- Benutzer anlegen mit Name, Benutzername und Rolle (`admin`, `senior_service`, `service`)
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
- Benutzer vergibt eigenes Passwort (min. 8 Zeichen)
- Neues Passwort wird mit Argon2id gehasht gespeichert
- Nach Passwort-Reset durch den Admin wird erneut ein Einmalpasswort generiert
- Nach erfolgreichem Setzen des Passworts kann sich der Benutzer normal anmelden

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

---

## 5 · Reporting und Auswertung

### R-01 · Tagesabrechnung

> **ID:** R-01 · **Rolle:** Admin
> **Status:** 🔲 Offen · **Prio:** Should-have

Der Admin kann am Ende einer Veranstaltung oder jederzeit zwischendurch eine Tagesabrechnung einsehen. Diese zeigt den Gesamtumsatz, den Umsatz pro Servicekraft sowie eine Übersicht aller Stornierungen. Damit erhalten Verantwortliche ein vollständiges Bild über Einnahmen und Korrekturen des Tages.

**Akzeptanzkriterien:**

- Gesamtumsatz des Tages einsehbar (Summe aller registrierten Zahlungen)
- Umsatz pro Servicekraft einsehbar (aufgeschlüsselt nach Benutzer)
- Übersicht aller Stornierungen mit Zeitpunkt, Tisch, stornierten Positionen und Betrag
- Abruf jederzeit möglich (nicht nur bei Tagesabschluss)
- Servicekraft kann eigene Bestellungen und deren Status einsehen (bestellt, geliefert, bezahlt, storniert)

### R-02 · Datenexport

> **ID:** R-02 · **Rolle:** Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Der Admin kann Umsätze, Bestellungen und Artikeldaten als CSV exportieren, um sie extern weiterverarbeiten zu können (z. B. für die Vereinsbuchhaltung).

**Akzeptanzkriterien:**

- Export von Umsätzen, Bestellungen und Artikeldaten als CSV
- Export jederzeit durch den Admin auslösbar

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

Für eine detaillierte Gegenüberstellung mit kommerziellen POS-Systemen siehe [Produktbeschreibung — Abgrenzung](produktbeschreibung.md#7-abgrenzung).
