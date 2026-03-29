# Das Datenmodell von jotti

Dieses Dokument erklärt das vollständige Datenbankschema von jotti — jede Tabelle, jede Spalte und die Entscheidungen dahinter. Es ist als zusammenhängender Lesetext formuliert.

---

## Überblick: Zwei Welten im selben Schema

Das Datenbankschema von jotti teilt sich in zwei grundlegend verschiedene Bereiche: **Stammdaten** und **Kassenbetrieb**.

Die **Stammdaten** — Benutzer, Tische, Produkte und Produktvarianten — werden klassisch nach dem CRUD-Prinzip verwaltet: Anlegen, Lesen, Ändern, Löschen (hier als Soft-Delete). Sie sind die ruhige Infrastruktur des Systems, die sich selten ändert.

Der **Kassenbetrieb** dagegen folgt einem vollständig anderen Muster: **Event-Sourcing**. Jeder finanzielle Vorgang — eine Bestellung, eine Zahlung, eine Stornierung — wird als unveränderlicher Eintrag im Kassenjournal protokolliert, der Haupttabelle des Systems. Darüber hinaus gibt es zwei abgeleitete Strukturen, die das Arbeiten mit diesen Events vereinfachen: eine synchrone Projektion für den laufenden Tischbetrieb und eine CRUD-Entität für die Kassensitzung als Ganzes.

Hinzu kommen zwei kleine Hilfstabellen: eine für Tisch-Favoriten der Servicekräfte und eine für die Bondrucker-Konfiguration.

---

## Enumerationen: Die festen Werteräume

Bevor die eigentlichen Tabellen beschrieben werden, lohnt ein Blick auf die drei Enumerationen, die das Schema durchziehen.

**`UserRole`** kennt drei Werte: `admin`, `serviceleitung` und `service`. Der `admin` hat vollen Zugriff — er verwaltet Stammdaten, eröffnet Kassensitzungen und führt den Tagesabschluss durch. Die `serviceleitung` darf alles, was den laufenden Tischbetrieb betrifft, einschließlich Stornierungen und Auszahlungen. Die einfache Servicekraft (`service`) kann bestellen, die Ausgabe bestätigen und kassieren, aber weder stornieren noch auszahlen.

**`EntityStatus`** wird von allen Stammdatentabellen verwendet: `active`, `inactive` und `deleted`. Datensätze werden in jotti niemals physisch gelöscht — stattdessen wechseln sie in den Status `deleted`. Das hat zwei Gründe: Erstens erfordert die referenzielle Integrität, dass vergangene Bestellungen weiterhin auf Produkte und Varianten zeigen können, die nicht mehr aktiv sind. Zweitens friert jotti Produktdaten zum Zeitpunkt der Bestellung als sogenannte **Fat Events** im Kassenjournal ein — auch das setzt voraus, dass die Originaldatensätze in der Datenbank erhalten bleiben.

**`ProduktKategorie`** unterscheidet zwischen `essen`, `getraenk` und `sonstiges`. Diese Kategorisierung ist zentral für den Bondruck: Jede Kategorie kann einem eigenen Drucker zugeordnet werden, damit Küchenbestellungen zur Küche und Getränkebestellungen zur Theke gehen.

---

## Stammdaten

### Benutzer (`users`)

Die `users`-Tabelle speichert alle Personen, die sich in jotti einloggen können. Sie ist bewusst auf Englisch gehalten — Authentifizierung und Benutzerverwaltung gehören zur technischen Infrastruktur, nicht zur deutschen Fachdomäne.

Jeder Benutzer hat eine automatisch generierte numerische **`id`** als Primärschlüssel. Der **`name`** ist der Anzeigename, der auch im Kassenjournal gespeichert wird — als Fat Event, damit das Protokoll auch nach einer Umbenennung lesbar bleibt. Der **`username`** ist der eindeutige Login-Name und wird in der Datenbank mit einem Index versehen, weil er bei jedem Login-Versuch gesucht wird.

Die Passwort-Verwaltung ist zweistufig: **`password_hash`** ist das reguläre Passwort in Argon2id-Hashform. Es ist `NULL`, wenn ein Benutzer noch kein eigenes Passwort gesetzt hat — also beim ersten Login nach dem Anlegen durch einen Admin. **`onetime_password_hash`** ist ein einmaliges Passwort, das ein Admin für einen neuen Benutzer oder einen Passwort-Reset setzt. Sobald der Benutzer dieses Einmalpasswort nutzt und ein eigenes Passwort definiert, wird dieses Feld wieder auf `NULL` gesetzt.

Die **`role`** bestimmt, welche Aktionen ein Benutzer ausführen darf, und wird als `UserRole`-Enum gespeichert. Der **`status`** steuert, ob das Konto aktiv (`active`), deaktiviert (`inactive`) oder gelöscht (`deleted`) ist. Die Zeitstempel **`created_at`** und **`updated_at`** protokollieren Anlage und letzte Änderung.

Beim ersten Start legt die Migration automatisch einen Admin-Benutzer an — mit Benutzername `nico` und einem vorgesetzten Einmalpasswort, das beim ersten Login geändert werden muss.

### Tische (`tische`)

Die `tische`-Tabelle ist denkbar einfach: Ein Tisch hat eine automatisch generierte **`id`**, einen eindeutigen **`name`** (z. B. „Tisch 1" oder „Biergarten Nord"), einen **`status`** für Soft-Delete und die üblichen Zeitstempel.

Warum gibt es eine separate Tisch-Tabelle? Weil Tische die räumliche Struktur einer Veranstaltung abbilden. Servicekräfte navigieren nach Tischen, Bestellungen werden Tischen zugeordnet, und der gesamte Kassenbetrieb ist tischbezogen organisiert. Die Tisch-ID taucht als Bestandteil des Event-Stream-Schlüssels im Kassenjournal wieder auf.

### Produkte (`produkte`) und Produktvarianten (`produkt_varianten`)

Das Produkt-Aggregat besteht aus zwei Tabellen. Ein **Produkt** ist die begriffliche Einheit — zum Beispiel „Bier". Eine **Produktvariante** ist die konkrete bestellbare Ausprägung mit Preis — also etwa „Bier 0,5 l" für 3,50 Euro oder „Bier 0,3 l" für 2,50 Euro.

Die `produkte`-Tabelle speichert neben **`id`**, **`name`** und den Zeitstempeln die **`kategorie`** als `ProduktKategorie`-Enum. Der **`status`** ermöglicht auch hier Soft-Delete.

Die `produkt_varianten`-Tabelle referenziert über **`produkt_id`** das zugehörige Produkt. Jede Variante hat einen eigenen **`name`** und einen **`preis_cents`** — Preise werden immer als ganze Cent-Beträge gespeichert, niemals als Fließkommazahl. Ein Preis von 3,50 Euro wird als `350` gespeichert. Das verhindert Rundungsfehler bei Berechnungen. Auch hier gibt es **`status`** für Soft-Delete und die üblichen Zeitstempel.

Die Trennung in Produkt und Variante ist notwendig, weil Produkte unterschiedliche Größen oder Ausführungen haben können, die alle zum gleichen Eintrag im Menü gehören, aber unterschiedliche Preise haben.

---

## Kassenbetrieb

### Kassensitzungen (`kassensitzungen`)

Die `kassensitzungen`-Tabelle ist eine CRUD-Entität, die den administrativen Rahmen eines Betriebstages repräsentiert. Bevor eine Servicekraft die erste Bestellung aufnehmen kann, muss ein Admin eine Kassensitzung eröffnen.

Der Primärschlüssel **`z_nr`** ist die fortlaufende Kassenabschlussnummer. Das Kürzel stammt aus der deutschen Finanzbuchhaltung und entspricht dem `Z_NR`-Feld im DSFinV-K-Standard — dem digitalen Datenformat für Finanzverwaltungen. Diese Nummer ist immer lückenlos aufsteigend und wird bei jeder neuen Kassensitzung als `max(z_nr) + 1` berechnet. Sie darf niemals zurückgesetzt werden.

**`datum`** speichert den Betriebstag. **`bezeichnung`** ist ein optionaler Freitext wie „Sommerfest Tag 1". Der **`status`** kennt genau zwei Werte: `offen` und `abgeschlossen`. Es darf zu jedem Zeitpunkt maximal eine Kassensitzung mit Status `offen` geben.

Warum ist `kassensitzungen` eine CRUD-Entität und nicht nur ein Event im Kassenjournal? Weil bei jedem einzelnen Tischvorgang — jeder Bestellung, jeder Zahlung — als erstes geprüft wird, ob überhaupt eine offene Kassensitzung existiert. Diese Prüfung würde bei einer reinen Event-Store-Architektur eine aufwändige Aggregation erfordern. Die CRUD-Entität macht sie zu einem einfachen `SELECT WHERE status = 'offen'`.

### Das Kassenjournal (`kassenjournal`)

Das Kassenjournal ist das Herzstück von jotti — und das wichtigste architektonische Designentscheidung des gesamten Systems. Es ist ein **append-only Event Store**: Einträge werden nur eingefügt, niemals verändert oder gelöscht.

Diese Immutabilität ist nicht nur eine technische Wahl, sondern eine gesetzliche Anforderung. Nach § 146 der Abgabenordnung müssen Aufzeichnungen einer Kasse vollständig, unveränderlich und nachvollziehbar sein. Das Kassenjournal erfüllt diese Anforderung auf Datenbankebene: Drei Datenbank-Trigger (`kassenjournal_no_update`, `kassenjournal_no_delete`, `kassenjournal_no_truncate`) werfen bei jedem Versuch, einen Eintrag zu ändern oder zu löschen, eine Exception. Zusätzlich werden dem `PUBLIC`-Datenbankrollen nur `SELECT` und `INSERT`-Berechtigungen erteilt — `UPDATE` und `DELETE` sind selbst für Tabellenbesitzer über die Trigger gesperrt.

Jeder Eintrag im Kassenjournal repräsentiert einen **Domain Event** — einen fachlichen Geschäftsvorfall, der tatsächlich stattgefunden hat.

Die **`id`** ist ein automatisch generierter Integer-Primärschlüssel. Die **`user_id`** und der **`user_name`** identifizieren, wer die Aktion ausgeführt hat. Der Name wird zum Zeitpunkt des Events eingefroren (Fat Event), damit das Protokoll auch bei späteren Umbenennungen lesbar bleibt.

Der **`type`** beschreibt die Art des Events, zum Beispiel `bestellung-aufgenommen:v1` oder `zahlung-kassiert:v1`. Die Versionsnummer im Typ erlaubt es, in Zukunft das Datenformat zu ändern, ohne historische Events zu verlieren.

Das **`subject`** ist der Stream-Schlüssel und folgt einem hierarchischen Schema: Tisch-bezogene Events haben ein Subject wie `kassensitzung-1/tisch-42`, Kassensitzung-Events einfach `kassensitzung-1`. Dieses Design erlaubt es, alle Events eines Tisches unabhängig von den Events anderer Tische zu lesen und zu verarbeiten.

Die **`version`** ist eine aufsteigende Ganzzahl pro Subject. Zusammen mit `subject` bildet sie einen `UNIQUE`-Constraint — das ist der Mechanismus für **Optimistic Concurrency Control**: Wenn zwei Servicekräfte gleichzeitig denselben Tisch bearbeiten, kann nur eine der beiden Transaktionen erfolgreich einfügen; die andere erhält einen Constraint-Fehler und muss neu laden und erneut versuchen.

Der **`timestamp`** ist der Zeitpunkt, zu dem das Event erzeugt wurde. **`data`** enthält die event-spezifischen Nutzdaten als JSONB — zum Beispiel bei einer Bestellung die Positionen mit Produktnamen, Variantennamen und Einzelpreisen. Diese Daten sind eingefroren: Selbst wenn ein Produkt später umbenannt oder sein Preis geändert wird, bleiben die historischen Bestellungen korrekt dokumentiert.

Die **`kassensitzung_nr`** ist eine denormalisierte Spalte, die angibt, zu welcher Kassensitzung dieses Event gehört. Sie ermöglicht schnelle Queries über alle Events einer Kassensitzung — zum Beispiel für den Kassenbestand oder den Tagesabschluss — ohne fragile String-Pattern-Matching auf dem Subject.

Für die häufigsten Zugriffsmuster sind sechs Indizes angelegt: auf `user_id` (wer hat etwas getan), auf `subject` (Single-Stream-Replay), auf `type` (alle Events eines bestimmten Typs), auf die Kombination `(subject, type)` (z. B. alle Zahlungen eines Tisches), auf `(type, timestamp)` (zeitliche Auswertungen) und auf `kassensitzung_nr` (Cross-Stream-Aggregationen).

### Die Tisch-Session-Projektion (`tisch_sessions`)

Das Kassenjournal ist die unveränderliche Quelle der Wahrheit — aber es wäre ineffizient, bei jedem Seitenaufruf alle Events eines Tisches neu zu lesen und zu aggregieren. Deshalb gibt es die `tisch_sessions`-Tabelle: eine **synchrone CQRS-Projektion**, die den aktuellen Zustand jedes Tisches materialisiert.

Der entscheidende Mechanismus: Diese Projektion wird **in derselben Datenbanktransaktion** wie der Event-INSERT aktualisiert. Es gibt also keinen asynchronen Prozess, keine eventuelle Konsistenz, keine Chance auf einen inkonsistenten Lesezustand. Wenn ein Event erfolgreich ins Kassenjournal geschrieben wurde, ist die Projektion bereits aktuell.

Der Primärschlüssel **`subject`** entspricht exakt dem Subject aus dem Kassenjournal: `kassensitzung-1/tisch-42`. Die `tisch_sessions`-Tabelle ist session-scoped — jede neue Kassensitzung startet mit einem leeren Projektionszustand.

**`tisch_id`** und **`kassensitzung_nr`** sind Fremdschlüssel auf die Stammdatentabellen. **`saldo_cents`** ist der aktuelle offene Betrag am Tisch — die Differenz aus Bestellungen, Zahlungen, Stornierungen und Auszahlungen, immer in Cent.

**`unbezahlte_positionen`** ist ein JSONB-Array mit allen Positionen, die bestellt, aber noch nicht bezahlt wurden. **`ausstehende_positionen`** enthält alle Positionen, die bestellt, aber noch nicht ausgegeben wurden. Diese zwei Listen ermöglichen es der Servicekraft, auf einen Blick zu sehen, was noch aussteht. **`gesamt_zahlungen_cents`** ist die Summe aller kassierten Zahlungen — nützlich für den Kassenbestandsüberblick.

**`last_event_id`** und **`last_event_version`** verweisen auf das zuletzt verarbeitete Event. Sie dienen als Konsistenzprüfung: Falls ein Neuaufbau der Projektion notwendig ist, kann man genau bis zu diesem Punkt vorwärts replayed haben.

### Tisch-Favoriten (`tisch_favoriten`)

Die `tisch_favoriten`-Tabelle ist eine einfache Many-to-Many-Beziehung zwischen Benutzern und Tischen. Eine Servicekraft kann Tische, für die sie zuständig ist, als Favoriten markieren — diese erscheinen dann prominent im Dashboard.

Der zusammengesetzte Primärschlüssel aus **`user_id`** und **`tisch_id`** stellt sicher, dass ein Benutzer denselben Tisch nicht doppelt favorisieren kann. **`created_at`** protokolliert, wann ein Favorit gesetzt wurde.

### Bondrucker-Konfiguration (`kategorie_drucker`)

Die `kategorie_drucker`-Tabelle konfiguriert, an welchen Drucker Bestellungen welcher Kategorie weitergeleitet werden. Der Primärschlüssel ist die **`kategorie`** (`essen`, `getraenk`, `sonstiges`) — es gibt also genau eine Zeile pro Kategorie, und die Migration fügt diese drei Zeilen direkt beim Setup ein.

**`drucker_ip`** ist die IPv4-Adresse des Bondruckers. Ein leerer String bedeutet, dass für diese Kategorie kein Drucker konfiguriert ist. **`bonmodus`** steuert, ob pro bestellter Position ein einzelner Bon gedruckt wird (`pro_position`) oder ob alle Positionen einer Bestellung auf einem Sammelbon landen (`pro_bestellung`). **`updated_at`** protokolliert die letzte Änderung der Konfiguration.

Diese Konfiguration wird zur Laufzeit gelesen — Änderungen wirken sofort für alle künftigen Bestellungen, ohne dass ein Neustart nötig ist.

---

## Das Zusammenspiel

Das gesamte Datenmodell folgt einer klaren Hierarchie:

Die **Stammdatentabellen** (`users`, `tische`, `produkte`, `produkt_varianten`) sind die Voraussetzung für den Betrieb. Sie definieren, wer handeln darf, wo Bestellungen aufgenommen werden und was bestellt werden kann.

Die **`kassensitzungen`**-Tabelle öffnet und schließt den zeitlichen Rahmen eines Betriebstages. Ohne offene Kassensitzung ist kein Tischbetrieb möglich.

Das **`kassenjournal`** protokolliert jeden Geschäftsvorfall unveränderlich. Es ist die Single Source of Truth — aus ihm könnte jeder andere Zustand im System rekonstruiert werden.

Die **`tisch_sessions`**-Projektion macht den aktuellen Tischzustand ohne teure Aggregationen lesbar. Sie ist ein Spiegel des Kassenjournals, der bei jedem Schreibvorgang synchron gehalten wird.

**`tisch_favoriten`** und **`kategorie_drucker`** sind unterstützende Strukturen: die eine für die Benutzerfreundlichkeit der Servicekräfte, die andere für die operative Verknüpfung mit physischer Druckerhardware.

Diese Architektur — CRUD für Stammdaten, Event-Sourcing für Kassenoperationen mit synchroner Projektion — gibt jotti lückenlose Nachvollziehbarkeit bei gleichzeitig einfacher operativer Lesbarkeit.
