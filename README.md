# jotti

jotti ist ein einfaches Bestellsystem für Vereine und Nonprofit-Organisationen.

Mit jotti können Produkte (Getränke und Speisen) verwaltet werden, sowie Bestellungen und Bezahlungen auf Tische gebucht werden.

Als Webapp kann jotti auf jedem Smartphone genutzt werden. jotti bietet Administratoren einen gesonderten Web-Zugang, um Produkte und Tische zu verwalten, sowie Buchhaltungsaufgaben zu bearbeiten.

---

Im Folgenden wird das Konzept für das MVP von jotti beschrieben.

## Entities

**Benutzer**

- Benutzer entsprechen in jotti Mitarbeiter:innen der Gastroeinrichtung bzw. Veranstaltung.
- Benutzer können Servicekräfte, PoS-Personal oder auch Verwaltungspersonal sein.
- Benutzer melden sich via Benutzername und Passwort an
- Manche Benutzer sind als Administrator (Admin) gekennzeichnet.
  - Admin-Benutzer haben alle Berechtigungen: z.B. Produkte verwalten, Tische verwalten etc.
  - Normale Benutzer können nur Bestellungen und Bezahlungen tätigen.

**Produkt**

- Ein Produkt entspricht in jotti einem Getränk oder einem Gericht.
- Ein Produkt hat einen Namen und einen Preis, und ist einer Kategorie zugeordnet ("Getränk", "Essen").
- Produkte werden über Bestellungen verkauft.
- Ein Produkt kann ausverkauft sein und kann in diesem Fall nicht mehr bestellt werden.
- Produkte werden von Administratoren verwaltet.
- Produkte können mehrfach bestellt werden.
- Beispiel: Kaffee, Bier, Pommes, Pizza

**Tisch**

- Tische bestellen Produkte und müssen diese anschließend bezahlen.
- Tische werden von Administratoren verwaltet.
- Tische haben einen Namen.
- Tische können Bestellungen und Bezahlungen haben.
- Beispiel: "Tisch 1", "Tisch 2", "Selbstbedienungskasse"

## Aggregates

**Bestellung**

- Jede Bestellung wird auf einen Tisch gebucht.
- Eine Bestellung beinhaltet eine Liste von Produkten mit Mengenangabe.

**Bezahlung**

- Jede Bezahlung wird auf einen Tisch gebucht und beinhaltet eine Liste von Produkten (und Mengenangaben)
- Bezahlungen können nur getätigt werden, wenn die ausgewählten Produkte (inkl. der angegebenen Menge) bei diesem Tisch noch unbezahlt sind.
- Bezahlungen sind unabhängig von Bestellungen. D.h. es wird nicht eine Bestellung bezahlt, sondern eine Menge von Produkten.
- jotti kennt die Art der Bezahlung (Bar, Karte, Gutschein etc.) und auch den tatsächlichen Kassenstand (Wechselgeld, Trinkgeld etc.) nicht. Diese werden extern verwaltet.

## Funktionale Anforderungen

- Benutzer (und Administratoren) können sich anmelden und abmelden.
- Administratoren können Produkte anlegen, bearbeiten und löschen.
- Administratoren können Tische anlegen, bearbeiten und löschen.
- Benutzer können für einen Tisch eine Bestellung aufgeben, indem sie Produkte und Mengen angeben.
- Benutzer können für einen Tisch eine Bezahlung tätigen, indem sie Produkte und Mengen angeben, die bezahlt werden sollen.
- Das System verhindert, dass mehr Produkte bezahlt werden, als bestellt wurden.
- Das System zeigt den aktuellen Status eines Tisches an: bestellte Produkte, bezahlte Produkte, offene (unbezahlte) Produkte.
- Administratoren können Berichte über Umsätze je Produkt und Zeitraum generieren.
- Administratoren können einen Tagesabschlussbericht generieren oder alle Bestellungen und Bezahlungen eines Zeitraums exportieren (z.B. als CSV).
- Das System protokolliert alle Bestellungen und Bezahlungen für Auditzwecke.
- Das System ist nur auf deutscher Sprache verfügbar.
- Das System unterstützt Mehrwertsteuer (z.B. 7% und 19%) und zeigt diese in Berichten an.
- Das System ist DSGVO-konform und speichert keine personenbezogenen Daten außer Benutzername und Passwort-Hash.
-

## Commands

- `v1/bestellung-aufgeben`: Ein Benutzer gibt eine Bestellung für einen Tisch auf.
- `v1/bezahlung-registrieren`: Ein Benutzer registriert eine Bezahlung für einen Tisch.
- `v1/produkt-anlegen`: Ein Administrator legt ein neues Produkt an.
- `v1/produkt-bearbeiten`: Ein Administrator bearbeitet ein bestehendes Produkt.
- `v1/produkt-loeschen`: Ein Administrator löscht ein bestehendes Produkt.
- `v1/tisch-anlegen`: Ein Administrator legt einen neuen Tisch an.
- `v1/tisch-bearbeiten`: Ein Administrator bearbeitet einen bestehenden Tisch.
- `v1/tisch-loeschen`: Ein Administrator löscht einen bestehenden Tisch.

## Events

- `bestellung-aufgegeben:v1`
  - Beschreibt, dass ein Tisch eine Bestellung aufgegeben hat.
  - Subject gibt den Tisch (z.B. `tisch:42`) an.
  - Die bestellten Produkte und ihre Menge sind in der Payload angegeben.
- `bezahlung-registriert:v1`
  - Beschreibt, dass ein Benutzer eine Bezahlung eines Tisches registriert hat.
  - Subject gibt den Tisch an (z.B. `tisch:42`).
  - Die bezahlten Produkte und ihre Menge sind in der Payload angegeben.

## Technische Spezifikation

- Single-Tenant: Es gibt ein System pro Verein/Veranstaltung.
- Als Datenbank wird Postgres eingesetzt.
  - Benutzer und Produkte werden in relationalen Tabellen gespeichert.
  - Bestellungen und Bezahlungen werden als Event-Sourcing Events in einer Event-Log Tabelle gespeichert.
- Der Server wird in Go geschrieben.
  - Der Server stellt eine HTTP API zur Verfügung.
  - Die API ist im Command- und Query-Pattern aufgebaut und verwendet JSON für die Datenübertragung.
  - Der Server implementiert Event Sourcing für Bestellungen und Bezahlungen.
- Das Frontend ist eine React SPA Webapp.
  - Die Webapp kommuniziert mit dem Server via HTTP API.
  - Die Webapp ist responsive und funktioniert auf Smartphones und Tablets.
  - Die Webapp hat einen Admin-Bereich für Administratoren.
- Als Webserver wird nginx eingesetzt, inklusive SSL Verschlüsselung via Let's Encrypt.
- Alle Services (Datenbank, Server und Webserver) laufen in Docker Containern und werden via Docker-Compose orchestriert.
- On-Premise Installation: Das System kann lokal auf einem Rechner oder Server installiert werden, ohne Cloud-Anbindung.
- Anmeldung via Benutzername und Passwort
  - Passwort-Hashing mit Argon2id [owasp-cheatsheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
  - Sessions werden via JSON Web Tokens (JWT) realisiert (12 Stunden Gültigkeit).

## Offene Fragen

- Steuern/Mehrwertsteuer: mehrere Sätze pro Produkt? Brutto/Netto-Preise, Rundung?
- Dynamische Preise (Happy Hour, Event-spezifisch), Rabatte/Aktionen, Gratisartikel?
- Ausverkauft: manuell gesetzt vs. Bestandsführung?

## Ideen für spätere Weiterentwicklungen

- Belegdruck: Küchen-/Bontickets, Re-Druck, Drucker pro Bereich?
- Stornierung: Benutzer können Bestellungen innerhalb von einer Minute stornieren. Danach sind Bestellungen final.
- Workflows: Tische zusammenlegen/teilen, Positionen zwischen Tischen verschieben, umbenennen, schließen/wieder öffnen?
- Offline‑First: Bestellungen und Bezahlungen müssen bei Verbindungsproblemen (z. B. Zeltfest mit schwachem Internet) lokal gepuffert und asynchron an den Server gesendet werden.
- Optionen für Produkte wie Beispielsweise "mit Soße".
