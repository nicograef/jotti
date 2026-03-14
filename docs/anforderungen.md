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
| **Serviceleitung** | `serviceleitung` | Kassenbetrieb einschließlich Stornierung                                     |
| **Servicekraft**   | `service`        | Kassenbetrieb ohne Stornierung                                               |

### Berechtigungsmatrix

| Aktion                   | Admin | Serviceleitung | Servicekraft |
| ------------------------ | :---: | :------------: | :----------: |
| Produkte verwalten       |   ✔   |                |              |
| Tische verwalten         |   ✔   |                |              |
| Benutzer verwalten       |   ✔   |                |              |
| Passwort zurücksetzen    |   ✔   |                |              |
| Bestellung aufgeben      |   ✔   |       ✔        |      ✔       |
| Lieferung bestätigen     |   ✔   |       ✔        |      ✔       |
| Zahlung registrieren     |   ✔   |       ✔        |      ✔       |
| Stornierung (K-04a/b)    |   ✔   |       ✔        |              |
| Umbuchung durchführen    |   ✔   |       ✔        |      ✔       |
| Tischübersicht einsehen  |   ✔   |       ✔        |      ✔       |
| Kassenjournal einsehen   |   ✔   |       ✔        |      ✔       |
| Abmelden                 |   ✔   |       ✔        |      ✔       |
| Tagesabrechnung einsehen |   ✔   |                |              |
| Datenexport              |   ✔   |                |              |
| Tagesabschluss einleiten |   ✔   |                |              |

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

### K-04a · Stornierung nicht-bezahlter Positionen

> **ID:** K-04a · **Rolle:** Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Serviceleitung oder Admin können unbezahlte bestellte Positionen nachträglich stornieren — unabhängig vom Lieferstatus. Einfache Servicekräfte haben kein Stornierungsrecht. Die Stornierung reduziert den offenen Saldo des Tisches.

**Akzeptanzkriterien:**

- Nur Serviceleitung (`serviceleitung`) und Admin (`admin`) dürfen stornieren
- Servicekraft (`service`) hat keinen Zugriff auf die Stornierungsfunktion
- Unbezahlte, nicht-stornierte Positionen sind stornierbar — unabhängig davon, ob sie bereits geliefert wurden
- Mindestens eine Position muss ausgewählt werden
- Kommentar optional (max. 100 Zeichen)
- Stornierung wird als unveränderliches `ProdukteStorniert`-Event im Kassenjournal gespeichert
- Saldo des Tisches wird nach Stornierung korrekt reduziert

### K-04b · Stornierung bereits bezahlter Positionen (Rückzahlung)

> **ID:** K-04b · **Rolle:** Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Should-have

Serviceleitung oder Admin können auch bereits bezahlte Positionen nachträglich stornieren. Da der Gast bereits gezahlt hat, entsteht ein negativer Saldo am Tisch — das System registriert die Rückzahlung als eigenes Event und korrigiert den Umsatz im Reporting entsprechend.

**Akzeptanzkriterien:**

- Bereits bezahlte, nicht-stornierte Positionen sind stornierbar (wie K-04a)
- Die Stornierung erzeugt zusätzlich ein `RueckzahlungRegistriert`-Event, das den zurückzuzahlenden Betrag festhält
- Saldo des Tisches wird nach Stornierung korrekt angepasst (kann negativ werden, bis die Rückzahlung quittiert wird)
- Rückzahlungsbetrag ist im Kassenjournal des Tisches sichtbar
- Reporting: `GetReportingStats`, `GetUmsatzProServicekraft` und `GetUmsatzProTisch` berücksichtigen Rückzahlungen korrekt (Umsatz = Zahlungen − Rückzahlungen)
- Frontend-Tooltips weisen nicht mehr auf fehlende Rückzahlungsberücksichtigung hin

### K-05 · Tischübersicht und Navigation

> **ID:** K-05 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** ✅ Umgesetzt · **Prio:** Must-have

Die Servicekraft sieht auf der Startseite alle aktiven Tische als Karten und navigiert per Tap zum Tisch-Detail. Dort stehen die Tischoperationen in drei Tabs zur Verfügung: Bestellen, Bezahlen und Historie. Liefern ist in den Bestellen-Tab integriert; Stornieren ist für `serviceleitung`/`admin` im Bezahlen-Tab verfügbar. Tischoperationen öffnen als Drawer (Overlay von unten).

**Akzeptanzkriterien:**

- Alle aktiven Tische werden als Karten angezeigt
- Tischkarte zeigt den aktuellen Saldo
- Navigation per Tap zum Tisch-Detail
- Tisch-Detail bietet drei Tabs: Bestellen, Bezahlen, Historie
- Liefern ist in den Bestellen-Tab integriert, Stornieren für Serviceleitung/Admin im Bezahlen-Tab
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
> **Status:** � Won't-have · **Prio:** ~~Nice-to-have~~

~~Einer Bestellung kann ein optionaler Name oder eine Bezeichnung zugewiesen werden (z. B. „Familie Müller", „Gruppe links"), um mehrere Gruppen an einem Tisch unterscheiden zu können.~~

**Entscheidung:** Wird über das bestehende Kommentarfeld der Bestellung gelöst (K-01, max. 100 Zeichen). Der Kommentar wird bereits in der Tisch-Historie (K-06) angezeigt. Ein eigenes Bezeichnungsfeld ist daher nicht notwendig.

### K-08 · Bestellungen umbuchen

> **ID:** K-08 · **Rolle:** Servicekraft · Serviceleitung · Admin
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

Bons werden automatisch oder manuell gedruckt, damit Ausgabestationen (Küche, Getränketheke) die bestellten Positionen erhalten. Bons enthalten alle relevanten Informationen: Tisch, Servicekraft, Positionen mit Mengen, Zeitstempel und optionalen Kommentar.

**Akzeptanzkriterien:**

- Bons sind übersichtlich formatiert (Tisch, Servicekraft, Positionen, Mengen, Gesamtpreis, Uhrzeit, Kommentar)
- Separater Bon pro Position druckbar (z. B. bei Fehlbon Nachdruck einer einzelnen Position)
- Getränkepositionen werden automatisch an den Getränkedrucker gesendet
- Essenspositionen werden automatisch an den Küchendrucker gesendet
- Drucker sind vom Admin konfigurierbar (Zuordnung Drucker zu Kategorie)

### K-12 · Küchendisplay (KDS)

> **ID:** K-12 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Should-have

Mitarbeiter an den Ausgabestationen (Getränketheke, Küche) sehen auf einem eigenen Bildschirm in Echtzeit die eingehenden Bestellungen ihrer Kategorie. Das Display dient als passive Anzeige — es zeigt offene Bestellungen gruppiert nach Tisch, sodass Ausgabestationen auch bei Bon-Verlust die Bestellungen nachvollziehen können.

**Akzeptanzkriterien:**

- Echtzeit-Anzeige offener Bestellungen nach Kategorie (Essen, Getränke), gruppiert nach Tisch
- Getränkeausgabe sieht offene Getränkebestellungen, Essensausgabe sieht offene Essensbestellungen
- Letzte Bestellungen sind einsehbar (bei Bon-Verlust)

### K-13 · Ausgabestationen mit Zubereitungsstatus

> **ID:** K-13 · **Rolle:** Servicekraft · Serviceleitung · Admin
> **Status:** 🔲 Offen · **Prio:** Nice-to-have

Aufbauend auf dem Küchendisplay (K-12) können Mitarbeiter an Ausgabestationen den Zubereitungsstatus einzelner Positionen verwalten. Servicekräfte sehen den Zubereitungsstatus und wissen, wann Positionen abholbereit sind.

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

Jede Servicekraft kann eine Übersicht über die eigenen Aktivitäten einsehen: aufgegebene Bestellungen, registrierte Zahlungen und deren Status. Dies gibt dem Helfer einen persönlichen Überblick über den eigenen Beitrag, ohne dass ein Admin benötigt wird.

**Akzeptanzkriterien:**

- Servicekraft kann eigene Bestellungen und deren Status einsehen (bestellt, geliefert, bezahlt, storniert)
- Anzeige des eigenen kassieren Umsatzes (Summe der selbst registrierten Zahlungen)
- Nur eigene Daten sichtbar — kein Einblick in Daten anderer Servicekräfte

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
