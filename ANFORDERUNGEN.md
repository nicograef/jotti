# Anforderungskatalog

Dieses Dokument beschreibt alle funktionalen Anforderungen an jotti, abgeleitet aus den Kundengesprächen und ergänzt um technische Einordnung. Die Anforderungen sind semantisch gruppiert, priorisiert und mit dem aktuellen Implementierungsstand abgeglichen.

## Legende

| Kürzel     | Bedeutung                                                                    |
| ---------- | ---------------------------------------------------------------------------- |
| **Prio**   | `Must-have` · `Nice-to-have` · `Später` · `Won't-have`                       |
| **Status** | ✅ Umgesetzt · 🔧 Teilweise · ❌ Offen                                       |
| **Rolle**  | Bediener · Admin · Rechner · Getränkeausgabe · Essensausgabe · Gast · System |

---

## Statusübersicht

### ✅ Umgesetzt

| #   | Anforderung                                                | Prio         | Rolle    |
| --- | ---------------------------------------------------------- | ------------ | -------- |
| 1   | Bestellaufnahme per Smartphone (plattformunabhängig)       | Must-have    | Bediener |
| 2   | Sichere Anmeldung für Servicekräfte                        | Must-have    | Bediener |
| 3   | Preisanzeige bei Produktauswahl                            | Must-have    | Bediener |
| 4   | Mengenänderung per +/−-Buttons mit Bestätigung             | Must-have    | Bediener |
| 5   | Positionen vor Bestellabgabe entfernen                     | Must-have    | Bediener |
| 6   | Tischübersicht als Liste                                   | Must-have    | Bediener |
| 7   | Mehrere Bestellungen pro Tisch                             | Must-have    | Bediener |
| 8   | Gesamtbetrag pro Bestellung einsehen                       | Must-have    | Bediener |
| 9   | Passwort selbst festlegen                                  | Must-have    | Bediener |
| 10  | Benutzerkonten verwalten (CRUD)                            | Must-have    | Admin    |
| 11  | Passwörter zurücksetzen                                    | Must-have    | Admin    |
| 12  | Artikelpreise zentral pflegen                              | Must-have    | Rechner  |
| 13  | Produktvarianten bestellen (z. B. Pommes mit Ketchup/Soße) | Must-have    | Bediener |
| 14  | Kommentar/Notiz pro Bestellvorgang                         | Nice-to-have | Bediener |
| 15  | Bestellung vor Abgabe komplett verwerfen                   | Nice-to-have | Bediener |
| 16  | Einzelne Positionen bei Abrechnung auswählen (Teilzahlung) | Nice-to-have | Bediener |
| 17  | Nachträgliche Stornierung durch Admin/Master               | Must-have    | Bediener |
| 18  | Mehrere Servicekräfte gleichzeitig                         | Must-have    | System   |
| 19  | Schnelle und korrekte Bestellaufnahme                      | Must-have    | Gast     |
| 20  | Einfache, intuitive Benutzeroberfläche                     | Must-have    | System   |

### 🔧 Teilweise umgesetzt

| #   | Anforderung                                                    | Prio      | Rolle    | Stand                                                      |
| --- | -------------------------------------------------------------- | --------- | -------- | ---------------------------------------------------------- |
| 21  | Produktübersicht nach Kategorien (Essen, Alkoholfrei, Alkohol) | Must-have | Bediener | Kategorie im Datenmodell vorhanden, UI zeigt flache Liste  |
| 22  | Servicekraft darf nach Bestellabgabe nicht stornieren          | Must-have | Bediener | Stornierung steht beiden Rollen offen; Rollenprüfung fehlt |

### ❌ Offen

| #   | Anforderung                                                | Prio         | Rolle           |
| --- | ---------------------------------------------------------- | ------------ | --------------- |
| 23  | Tisch-Schnellsuche per Shortcut vom Dashboard              | Must-have    | Bediener        |
| 24  | Übersicht: eigene Bestellungen/Produkte/Tische mit Status  | Must-have    | Bediener        |
| 25  | Bestellungen auf anderen Tisch umbuchen                    | Must-have    | Bediener        |
| 26  | Umsatz pro Bediener für Tagesabrechnung                    | Must-have    | Rechner         |
| 27  | Übersichtlich formatierte Bons drucken                     | Must-have    | System          |
| 28  | Separater Bon pro Position                                 | Must-have    | Bediener        |
| 29  | Automatischer Getränkebon an Getränkeausgabe               | Must-have    | Getränkeausgabe |
| 30  | Automatischer Essensbon an Essensausgabe                   | Must-have    | Essensausgabe   |
| 31  | Freibon mit freier Preiseingabe                            | Must-have    | Bediener        |
| 32  | Druckererkennung und -konfiguration                        | Must-have    | System          |
| 33  | Offline-Fähigkeit (Bestellaufnahme bei Internetausfall)    | Must-have    | System          |
| 34  | Lokale Datenspeicherung bei Absturz                        | Must-have    | System          |
| 35  | Zubereitungsstatus einsehen (in Zubereitung / abholbereit) | Nice-to-have | Bediener        |
| 36  | Bezeichnung/Name pro Bestellung (Gruppen am Tisch)         | Nice-to-have | Bediener        |
| 37  | Rückgeldberechnung (Eingabe des Zahlbetrags)               | Nice-to-have | Bediener        |
| 38  | Tagesumsatz pro Bediener jederzeit einsehen                | Nice-to-have | Rechner         |
| 39  | Gesamtumsatz (alle Bediener) pro Tag                       | Nice-to-have | Rechner         |
| 40  | Datenexport (CSV, Excel)                                   | Nice-to-have | Rechner         |
| 41  | Eng getaktete Datenspeicherung (Stromausfall-Schutz)       | Nice-to-have | System          |
| 42  | Freitext-Notiz pro Position (auf Bon sichtbar)             | Nice-to-have | Bediener        |
| 43  | Benachrichtigung bei abholbereiter Bestellung              | Später       | Bediener        |
| 44  | Getränkeausgabe: offene Getränkebestellungen sehen         | Später       | Getränkeausgabe |
| 45  | Getränkeausgabe: Bestellungen als „fertig" markieren       | Später       | Getränkeausgabe |
| 46  | Essensausgabe: Bestellstatus ändern (Zubereitung/fertig)   | Später       | Essensausgabe   |
| 47  | Essensausgabe: Letzte Bestellungen einsehen (Bon verloren) | Später       | Essensausgabe   |
| 48  | Korrekte Abrechnung für Gäste sichergestellt               | Später       | Gast            |
| 49  | Benachrichtigung, wenn Gast zahlen möchte                  | Won't-have   | Bediener        |
| 50  | Übersicht aller Bon-Stornierungen                          | Won't-have   | Rechner         |

---

## Detaillierte Anforderungen

Die Anforderungen sind nach Funktionsbereichen gruppiert. Jeder Bereich enthält die zugehörigen User Stories und — bei offenen Anforderungen — einen Implementierungsvorschlag.

---

### 1 · Bestellaufnahme

Kernprozess: Servicekraft wählt einen Tisch, stellt eine Bestellung aus dem Produktkatalog zusammen und gibt sie ab.

#### 1.1 Plattformunabhängige Bestellaufnahme per Smartphone ✅

> Als Servicekraft möchte ich Bestellungen auf meinem eigenen Smartphone aufnehmen können, unabhängig vom Betriebssystem.

Umgesetzt als responsive Webapp, aufrufbar über jeden modernen Browser.

#### 1.2 Preisanzeige bei Produktauswahl ✅

> Als Servicekraft möchte ich beim Zusammenstellen einer Bestellung die Preise der Produkte und Varianten sehen.

Preise werden in der Produktliste pro Variante angezeigt (`formatCents()`).

#### 1.3 Mengenänderung per +/−-Buttons ✅

> Als Servicekraft möchte ich die Menge einer Position über +/−-Buttons ändern und die Bestellung per Bestätigung abschließen.

Implemented in `ProductList.tsx` mit Mengen-State und Bestätigung über `OrderDrawer`.

#### 1.4 Positionen vor Abgabe entfernen ✅

> Als Servicekraft möchte ich einzelne Positionen aus der Bestellung entfernen können, bevor ich sie abschicke.

Über die −-Buttons kann die Menge auf 0 gesetzt werden. Positionen mit Menge 0 werden nicht übermittelt.

#### 1.5 Bestellung vor Abgabe komplett verwerfen ✅

> Als Servicekraft möchte ich eine laufende Bestellung komplett verwerfen können, nachdem ich den Abbruch bestätigt habe.

Der OrderDrawer hat einen „Abbrechen"-Button; alternativ kann die Seite verlassen werden — Mengen liegen nur im lokalen React-State.

#### 1.6 Produktvarianten bestellen ✅

> Als Servicekraft möchte ich Produktvarianten bestellen können (z. B. Pommes mit Ketchup, Pommes mit Soße).

Vollständig umgesetzt: Produkt-Varianten-Modell in DB, Admin-CRUD, Variantenauswahl in der Service-Bestellansicht.

#### 1.7 Mehrere Bestellungen pro Tisch ✅

> Als Servicekraft möchte ich für einen Tisch beliebig viele Bestellungen aufgeben können.

Jede Bestellung wird als eigenes Event (`table.order-placed:v1`) gespeichert; pro Tisch unbegrenzt möglich.

#### 1.8 Kommentar/Notiz pro Bestellvorgang ✅

> Als Servicekraft möchte ich zu einem Bestellvorgang einen Kommentar hinterlegen können.

`CommentField`-Komponente ist in allen vier Drawern (Bestellen, Bezahlen, Liefern, Stornieren) eingebunden. Maximal 100 Zeichen, wird im Event gespeichert.

#### 1.9 Produktübersicht nach Kategorien 🔧

> Als Servicekraft möchte ich einen klaren Überblick über das Sortiment haben — gegliedert nach Kategorien (z. B. Essen, Alkoholfreie Getränke, Alkoholische Getränke).

**Stand:** Das Datenmodell kennt `ProductCategory` (food, beverage, other). Die Service-UI zeigt Produkte jedoch als flache, ungruppierte Liste.

**Implementierungsvorschlag:**

- In `ProductList.tsx` die Produkte nach `category` gruppieren und mit Überschriften (z. B. „Essen", „Getränke", „Sonstiges") rendern.
- Optional: Sticky-Headers oder Tab-Navigation pro Kategorie für schnelles Scrollen auf dem Smartphone.
- Ggf. die Kategorie-Werte um „Alkoholfrei" und „Alkoholisch" erweitern (Enum-Migration + Backend-Anpassung).

#### 1.10 Tisch-Schnellsuche per Shortcut ❌

> Als Servicekraft möchte ich über einen Shortcut (z. B. Button unten rechts) die Tischnummer direkt eingeben können, um schnell eine neue Bestellung zu starten.

**Implementierungsvorschlag:**

- Floating Action Button (FAB) auf der Tischübersicht (`TableSelectionPage.tsx`), z. B. mit Lucide-Icon `Search` oder `Hash`.
- Tap öffnet ein Suchfeld/Nummernpad; Eingabe einer Tischnummer navigiert direkt zu `/service/table/:id`.
- Alternativ: Suchfeld oben auf der Seite, das die Tischliste filtert.

#### 1.11 Bestellungen auf anderen Tisch umbuchen ❌

> Als Servicekraft möchte ich eine Bestellung nachträglich auf einen anderen Tisch umbuchen können, um Eingabefehler zu korrigieren.

**Implementierungsvorschlag:**

- Neuer Event-Typ `table.order-transferred:v1` mit Quell- und Ziel-Tisch-ID.
- Backend: Neuer Endpunkt `POST /service/transfer-table-order` — erzeugt ein Storno-Event am Quell-Tisch und ein Order-Event am Ziel-Tisch (atomar in einer Transaktion).
- Frontend: Aktion „Umbuchen" im Tisch-Menü, öffnet einen Drawer zur Tischauswahl.

#### 1.12 Freibon mit freier Preiseingabe ❌

> Als Servicekraft möchte ich einen Freibon mit freier Preiseingabe zur Bestellung hinzufügen können (für Sonderpositionen, die nicht im Katalog stehen).

**Implementierungsvorschlag:**

- Neue Produkt-Variante „Freibon" im System (oder spezielle Position ohne Produkt-Zuordnung).
- In der Bestellansicht ein Button „Freibon hinzufügen", der ein Formular mit Bezeichnung + Preis (Euro-Eingabe → Cent) öffnet.
- Backend: Freibon-Positionen als `LineItem` mit `variant_id = null` und eigenem Preis/Bezeichnung im Event-Data speichern. Alternativ: Spezielle „Freibon"-Variante, deren Preis pro Bestellung überschrieben wird.

#### 1.13 Freitext-Notiz pro Position ❌

> Als Servicekraft möchte ich zu einzelnen Positionen (nicht nur zur Gesamtbestellung) eine Freitext-Notiz hinzufügen können, die auf dem Bon sichtbar ist.

**Implementierungsvorschlag:**

- `LineItem`-Struct um optionales Feld `note: string` (max. 80 Zeichen) erweitern.
- In `ProductList.tsx` pro Variante ein kleines Notiz-Icon, das ein Textfeld öffnet.
- Backend: `note` in `OrderPlacedEvent.data.items[].note` mitführen, bei Bon-Druck ausgeben.

#### 1.14 Bezeichnung/Name pro Bestellung ❌

> Als Servicekraft möchte ich einer Bestellung einen Namen oder eine Bezeichnung geben können, um mehrere Gruppen an einem Tisch unterscheiden zu können.

**Implementierungsvorschlag:**

- Optionales Feld `label: string` in `OrderPlacedEvent.data` (z. B. „Familie Müller", „Gruppe links").
- Im OrderDrawer ein optionales Textfeld oberhalb der Positionen.
- In der Tisch-Historie und auf Bons das Label anzeigen, um Zuordnung zu erleichtern.

---

### 2 · Abrechnung und Bezahlung

Kassiervorgang: Bezahlung registrieren, Teilzahlung, Rückgeld, Gesamtbeträge einsehen.

#### 2.1 Gesamtbetrag pro Bestellung einsehen ✅

> Als Servicekraft möchte ich den Gesamtbetrag pro Tisch / pro Bestellung einsehen, um korrekt abrechnen zu können.

Tisch-Balance wird über `get-table-balance` berechnet und prominent auf der Tisch-Seite angezeigt. Einzelbeträge in der `Receipt`-Komponente.

#### 2.2 Einzelne Positionen bei Abrechnung auswählen (Teilzahlung) ✅

> Als Servicekraft möchte ich bei der Abrechnung einzelne Positionen und Mengen auswählen können, die bezahlt werden sollen.

`Payment.tsx` zeigt unbezahlte Varianten mit +/−-Mengenwahl. Nur ausgewählte Positionen werden als `register-table-payment` Event gespeichert.

#### 2.3 Rückgeldberechnung ❌

> Als Servicekraft möchte ich den vom Gast erhaltenen Betrag eingeben können, damit das System das Rückgeld berechnet.

**Implementierungsvorschlag:**

- Im `PaymentDrawer` ein zusätzliches Eingabefeld „Erhalten" (Euro-Betrag).
- Darunter die Anzeige: „Rückgeld: X,XX €" (Differenz zum Gesamtbetrag der ausgewählten Positionen).
- Rein clientseitige Berechnung, kein Backend-Endpunkt nötig.

---

### 3 · Stornierung

Wer darf stornieren, und unter welchen Bedingungen.

#### 3.1 Servicekraft darf nach Abgabe nicht stornieren 🔧

> Als Servicekraft darf ich nach Aufgabe der Bestellung diese nicht mehr stornieren können. Nur ein Master/Admin soll dies nachträglich tun können.

**Stand:** Stornierung ist über den `cancel-table-variants`-Endpunkt für beide Rollen (`admin` und `service`) erlaubt. Die geforderte Einschränkung auf Admins fehlt.

**Implementierungsvorschlag:**

- Im Service-Router eine zusätzliche Middleware oder Rollenprüfung einbauen, die `cancel-table-variants` nur für Role `admin` freigibt.
- Alternativ: Separater Admin-Stornierungsendpunkt, Service-Variante entfernen.
- Frontend: Stornierungstab im `TablePage` nur für Admins rendern (Role aus JWT/Auth-Context prüfen).

#### 3.2 Nachträgliche Stornierung durch Admin ✅

> Ein Admin/Master soll Bestellungen nach Aufgabe stornieren können.

Funktioniert — Admin-Rolle hat Zugriff auf `cancel-table-variants` im Service-Bereich.

---

### 4 · Tischverwaltung

Tischübersicht und -auswahl im Servicebetrieb.

#### 4.1 Tischübersicht als Liste ✅

> Als Servicekraft möchte ich alle aktiven Tische als Übersicht sehen, um schnell den richtigen Tisch auszuwählen.

`TableSelectionPage.tsx` zeigt alle aktiven Tische als Karten-Grid.

#### 4.2 Übersicht: Bestellungen, Produkte und Status ❌

> Als Servicekraft möchte ich eine Übersicht über meine Bestellungen, Produkte und Tische mit deren aktuellem Status haben.

**Implementierungsvorschlag:**

- Neue Service-Seite „Meine Bestellungen" unter `/service/orders`.
- Backend: Neuer Endpunkt `POST /service/get-user-orders`, der alle Events für den aktuellen `user_id` aggregiert (Bestellungen, Status, Tisch).
- Frontend: Chronologische Liste der eigenen Bestellungen mit Tisch, Zeitstempel, Positionen und Status (bestellt/geliefert/bezahlt/storniert).

---

### 5 · Benutzerverwaltung

Admin-Funktionen für Konten, Passwörter und Rollen.

#### 5.1 Sichere Anmeldung ✅

> Als Servicekraft möchte ich mich sicher auf meinem Smartphone anmelden können.

JWT-Auth mit HS256, 12h Gültigkeit, Argon2id-Passwort-Hashing.

#### 5.2 Passwort selbst festlegen ✅

> Als Servicekraft möchte ich mein Passwort selbst festlegen können.

Endpunkt `POST /auth/set-password` — ermöglicht das Setzen eines neuen Passworts nach Erstanmeldung mit Einmalpasswort.

#### 5.3 Benutzerkonten verwalten ✅

> Als Admin möchte ich Accounts der Servicekräfte anlegen, bearbeiten und deaktivieren können.

Vollständiger CRUD über Admin-Endpunkte: `create-user`, `update-user`, `activate-user`, `deactivate-user`. Soft-Deletes über Status.

#### 5.4 Passwörter zurücksetzen ✅

> Als Admin möchte ich Passwörter von Servicekräften zurücksetzen können.

Endpunkt `POST /admin/reset-password` — generiert ein neues Einmalpasswort.

---

### 6 · Produktverwaltung

Admin-Funktionen für Produkte, Varianten und Preise.

#### 6.1 Artikelpreise zentral pflegen ✅

> Als zuständige Person möchte ich Artikelpreise zentral pflegen können, um kurzfristige Änderungen vorzunehmen.

Admin-Endpunkte für Produkt- und Varianten-CRUD. Preise in Cent, änderbar über `update-variant`.

---

### 7 · Bondruck und Ausgabe

Drucken, Formatieren und Verteilen von Bons an Ausgabestationen.

#### 7.1 Übersichtlich formatierte Bons drucken ❌

> Das System soll übersichtlich formatierte Bons drucken können: Bediener, Tisch, Bestellung, Menge, Gesamtpreis, Uhrzeit, Notiz.

**Implementierungsvorschlag:**

- ESC/POS-kompatible Bon-Formatierung im Backend (Go-Library, z. B. `escpos`).
- Neuer Endpunkt `POST /service/print-receipt`, der Event-Daten zu einem druckbaren Bon aufbereitet.
- Alternativ: Web-basierter Druck über `window.print()` mit speziell gestyltem Print-CSS (58mm/80mm Bonformat).
- Die bestehende `Receipt.tsx`-Komponente als visuelle Vorlage nutzen.

#### 7.2 Separater Bon pro Position ❌

> Als Servicekraft möchte ich pro Position einen separaten Bon drucken lassen (z. B. bei Fehlbon nur eine Position nachdrucken).

**Implementierungsvorschlag:**

- Im Bon-Druck-Dialog Auswahl: „Alle Positionen" oder „Einzelne Position".
- Bei Einzelbon: Position auswählen → eigener Bon mit nur dieser Position.
- Setzt die Bon-Druckinfrastruktur aus 7.1 voraus.

#### 7.3 Automatischer Getränkebon an Getränkeausgabe ❌

> Bei Bestelleingang soll automatisch ein Getränkebon an die Getränkeausgabe gedruckt werden.

**Implementierungsvorschlag:**

- Bons nach Produktkategorie aufteilen: `beverage`-Positionen → Getränkedrucker, `food`-Positionen → Küchendrucker.
- Konfiguration der Drucker-Zuordnung pro Kategorie in den Admin-Einstellungen.
- Backend: Nach `order-placed`-Event automatisch Druckaufträge per Kategorie generieren.
- Setzt Drucker-Konfiguration (7.6) voraus.

#### 7.4 Automatischer Essensbon an Essensausgabe ❌

> Bei Bestelleingang soll automatisch ein Essensbon an die Essensausgabe gedruckt werden.

Analog zu 7.3, aber für Kategorie `food`. Gleiche Infrastruktur.

#### 7.5 Freibon-Druck ❌

Freibons (vgl. 1.12) müssen ebenfalls auf Bons gedruckt werden. Hängt von Bon-Infrastruktur ab.

#### 7.6 Druckererkennung und -konfiguration ❌

> Das System soll Drucker automatisch erkennen oder einfach konfigurierbar machen.

**Implementierungsvorschlag:**

- Admin-Einstellungsseite für Drucker: IP/Hostname, Port, Typ (ESC/POS Netzwerkdrucker).
- Pro Drucker: Zuordnung zu Kategorie (Getränke, Essen, Kasse).
- Backend: Drucker-Tabelle in DB, Health-Check-Endpunkt zum Testen der Verbindung.
- Alternativ für einfachere Setups: Browser-Print-Dialog (keine Netzwerkdrucker-Konfiguration nötig).

---

### 8 · Reporting und Auswertung

Umsatzübersichten, Tagesabrechnungen und Datenexport.

#### 8.1 Umsatz pro Bediener für Tagesabrechnung ❌

> Als verantwortliche Person möchte ich alle Umsätze pro Bediener einsehen können, um die Tagesabrechnung zu erstellen.

**Implementierungsvorschlag:**

- Neuer Admin-Endpunkt `POST /admin/get-revenue-by-user`, der `payment-registered`-Events nach `user_id` aggregiert.
- Zeitraum-Parameter (z. B. `date` oder `from`/`to`).
- Frontend: Neue Admin-Seite „Tagesabrechnung" mit Tabelle: Bediener | Umsatz | Anzahl Zahlungen.

#### 8.2 Tagesumsatz pro Bediener jederzeit einsehen ❌

> Als verantwortliche Person möchte ich jederzeit den aktuellen Tagesumsatz einzelner Servicekräfte einsehen können.

Gleiche Infrastruktur wie 8.1, aber als Live-Ansicht zugänglich (nicht nur am Tagesende).

#### 8.3 Gesamtumsatz aller Bediener pro Tag ❌

> Als verantwortliche Person möchte ich den Gesamtumsatz aller Servicekräfte pro Tag einsehen können.

**Implementierungsvorschlag:**

- Erweitert 8.1 um eine Summenzeile über alle Bediener.
- Alternativ: Eigener Endpunkt `POST /admin/get-daily-revenue`.
- Aggregation aller `payment-registered`-Events des Tages.

#### 8.4 Datenexport (CSV, Excel) ❌

> Als verantwortliche Person möchte ich Umsätze, Bedienungen und Artikel exportieren können (CSV, Excel).

**Implementierungsvorschlag:**

- Backend: Endpunkt `POST /admin/export-data` mit Parameter `format` (csv/xlsx) und `type` (revenue/orders/products).
- CSV-Export mit Go's `encoding/csv`, Excel mit einer Library wie `excelize`.
- Frontend: Download-Button auf der Reporting-Seite, Response als Datei-Download (`Content-Disposition: attachment`).

#### 8.5 Übersicht aller Bon-Stornierungen ❌ _(Won't-have)_

> Als verantwortliche Person möchte ich eine Übersicht aller Bon-Stornierungen haben.

**Implementierungsvorschlag (falls doch gewünscht):**

- Alle `variants-canceled`-Events abfragen, nach Zeitraum und Bediener filterbar.
- Admin-Seite mit Tabelle: Zeitpunkt, Bediener, Tisch, stornierte Positionen, Betrag.

---

### 9 · Ausgabestationen

Getränke- und Essensausgabe: Bestellstatus verwalten, Bons einsehen.

#### 9.1 Getränkeausgabe: Offene Bestellungen sehen ❌ _(Später)_

> Als Mitarbeiter an der Getränkeausgabe möchte ich sehen, welche Getränkebestellungen noch offen sind.

**Implementierungsvorschlag:**

- Neue Rolle `station` oder eigener Bereich `/station/drinks`.
- Endpunkt: Alle ungelieferten Positionen der Kategorie `beverage`, gruppiert nach Tisch.
- Echtzeit-Updates über Polling oder WebSocket.

#### 9.2 Getränkeausgabe: Bestellungen als „fertig" markieren ❌ _(Später)_

> Als Mitarbeiter an der Getränkeausgabe möchte ich Bestellungen als „fertig" markieren können.

**Implementierungsvorschlag:**

- Neuer Event-Typ `table.variants-prepared:v1` (Zubereitung abgeschlossen, bereit zur Abholung).
- Oder Nutzung des bestehenden `deliver`-Events, falls „fertig" gleichbedeutend mit „geliefert an Bediener" ist.

#### 9.3 Essensausgabe: Status ändern ❌ _(Später)_

> Als Mitarbeiter an der Essensausgabe möchte ich den Bestellstatus ändern können (in Zubereitung → fertig).

**Implementierungsvorschlag:**

- Neuer Event-Typ `table.variants-status-changed:v1` mit Status-Feld (preparing/ready).
- Analog zur Getränkeausgabe, aber für `food`-Kategorie.

#### 9.4 Essensausgabe: Letzte Bestellungen einsehen ❌ _(Später)_

> Als Mitarbeiter an der Essensausgabe möchte ich letzte Bestellungen einsehen können, falls ein Bon verloren geht.

**Implementierungsvorschlag:**

- Ansicht der letzten N `order-placed`-Events mit `food`-Positionen.
- Filterbar nach Tisch oder Zeitraum.

#### 9.5 Zubereitungsstatus für Servicekraft einsehen ❌

> Als Servicekraft möchte ich sehen, welche Produkte noch in Zubereitung sind und welche für mich zur Abholung bereitstehen.

Hängt von 9.2/9.3 ab — nur möglich, wenn Ausgabestationen den Status pflegen.

---

### 10 · Benachrichtigungen

Push- oder In-App-Benachrichtigungen bei Statusänderungen.

#### 10.1 Benachrichtigung bei abholbereiter Bestellung ❌ _(Später)_

> Als Servicekraft möchte ich benachrichtigt werden, sobald meine Bestellung abholbereit ist.

**Implementierungsvorschlag:**

- Web Push Notifications (Service Worker + Push API) oder In-App-Banner.
- Trigger: `variants-prepared`-Event (setzt 9.2/9.3 voraus).
- PWA-Manifest + Service Worker für Push-Fähigkeit hinzufügen.

#### 10.2 Benachrichtigung, wenn Gast zahlen möchte ❌ _(Won't-have)_

> Als Servicekraft möchte ich benachrichtigt werden, wenn ein Gast zahlen möchte.

Würde eine Gast-Interaktion voraussetzen (z. B. QR-Code am Tisch, Klingel-Button). Aktuell nicht vorgesehen.

---

### 11 · System und Infrastruktur

Nicht-funktionale Anforderungen: Offline, Ausfallsicherheit, Usability, Mehrbenutzerfähigkeit.

#### 11.1 Mehrere Servicekräfte gleichzeitig ✅

> Das System soll mehrere Servicekräfte gleichzeitig unterstützen.

Standard-Webarchitektur: Stateless REST-API, JWT pro Nutzer, PostgreSQL mit Connection-Pooling. Keine Einschränkung der gleichzeitigen Nutzer.

#### 11.2 Einfache, intuitive Benutzeroberfläche ✅

> Das System soll eine einfache Benutzeroberfläche bieten, damit sich alle Servicekräfte schnell zurechtfinden.

Umgesetzt mit shadcn/ui (konsistente Komponenten), Drawer-Pattern für alle Aktionen, mobile-first Design, Dark Mode.

#### 11.3 Offline-Fähigkeit ❌

> Das System soll auch bei Internetausfall funktionieren, damit Bestellungen weiterhin aufgenommen werden können.

**Stand:** Keine Offline-Unterstützung vorhanden. Kein Service Worker, kein PWA-Manifest, kein IndexedDB-Buffering. Im CSV als „Done" markiert, tatsächlich nicht umgesetzt.

**Implementierungsvorschlag:**

- `vite-plugin-pwa` integrieren → generiert Service Worker + `manifest.json`.
- Offline-Queue: Bestellungen in IndexedDB zwischenspeichern, bei Wiederverbindung synchronisieren.
- Statische Assets über Service Worker cachen (App-Shell bleibt offline nutzbar).
- Produktkatalog lokal cachen (selten geändert, per Cache-Invalidation aktualisierbar).
- Komplexität: Hoch — Konfliktauflösung bei gleichzeitiger Offline-Nutzung mehrerer Geräte.

#### 11.4 Lokale Datenspeicherung bei Absturz ❌

> Das System soll Daten lokal speichern, damit bei Absturz oder Stromausfall keine Bestellungen verloren gehen.

**Implementierungsvorschlag:**

- In-Progress-Bestellungen (Mengenauswahl) in `localStorage` oder `IndexedDB` persistieren.
- Bei App-Start prüfen, ob ungesendete Bestellungen vorhanden sind → Recovery-Dialog.
- Serverseitig sind abgeschlossene Bestellungen in PostgreSQL persistiert und gehen nicht verloren.

#### 11.5 Eng getaktete Datenspeicherung ❌

> Das System soll Daten möglichst häufig speichern, damit auch bei Stromausfall keine Daten verloren gehen.

**Implementierungsvorschlag:**

- PostgreSQL: `synchronous_commit = on` (Standard, bereits aktiv).
- Frontend: Offene Warenkörbe regelmäßig per `localStorage.setItem()` sichern (z. B. bei jedem +/−-Klick).
- WAL-basiertes Backup kann für zusätzliche Sicherheit konfiguriert werden.

#### 11.6 Schnelle und korrekte Bestellaufnahme ✅

> Als Gast möchte ich, dass meine Bestellung schnell und korrekt aufgenommen wird.

Gegeben durch die einfache UI, +/−-Mengenauswahl, Bestätigungs-Drawer mit Zusammenfassung vor Abgabe.

#### 11.7 Korrekte Abrechnung ✅ _(implizit)_

> Als Gast möchte ich korrekt abgerechnet werden.

Event-Sourcing garantiert konsistente Balance: Bestellungen − Bezahlungen − Stornierungen. Keine Rundungsfehler (Cent-Beträge).

---

## Zusammenfassung nach Priorität

| Priorität    | Gesamt | ✅ Umgesetzt | 🔧 Teilweise | ❌ Offen |
| ------------ | ------ | ------------ | ------------ | -------- |
| Must-have    | 31     | 19           | 2            | 10       |
| Nice-to-have | 11     | 2            | 0            | 9        |
| Später       | 6      | 0            | 0            | 6        |
| Won't-have   | 2      | 0            | 0            | 2        |
| **Gesamt**   | **50** | **21**       | **2**        | **27**   |
