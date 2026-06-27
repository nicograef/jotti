# Screenshots von jotti

Übersicht über die in diesem Ordner abgelegten Ob
erflächen-Screenshots von **jotti**,
dem mobile-first Gastronomie-Kassensystem (POS).
Alle Aufnahmen stammen aus der
Demo-Instanz (`demo.jotti.rocks`) mit Beispieldat
en der Veranstaltung „Sommerfest 26".
Die Oberfläche gliedert sich in zwei Rollen: den
**Admin-Bereich** (`/admin`) zur
Verwaltung und Auswertung sowie den **Service-Ber
eich** (`/service`) für die Kellner.

Jeder Eintrag verlinkt das Bild, nennt die zugehö
rige Route und beschreibt den Screenshot in einem
 Satz.

## Admin – Verwaltung & Auswertung

- [Navigation / Sidebar](./Screenshot_20260626-224430.png) (`/admin`) — Das geöffnete Admin-Navi
gationsmenü mit den Bereichen Auswertungen, Verwa
ltung, Service und Einstellungen samt Dark-Mode-U
mschalter und „Abmelden", wobei der Punkt „Benutz
er" als aktive Seite markiert ist.
- **[Auswertung – Historische Auswertung](Screens
hot_20260626-225422.png)** (`/admin/auswertung`)
— Die historische Auswertung einer abgeschlossene
n Kassensitzung mit vier Kennzahl-Karten (Gesamtu
msatz, Direktverkauf, Bestellungen, Stornierungen
), DSFinV-K-Export-Button und der Tabelle „Umsatz
 nach Steuersatz".
- **[Auswertung – Live-Dashboard (Stornierungen)]
(Screenshot_20260626-225359.png)** (`/admin/auswe
rtung`) — Das Live-Dashboard im Reiter „Stornieru
ngen" listet die stornierten Vorgänge mit Grund,
Servicekraft, Erstattungsart (Bar-Rückgabe/Geldne
utral) und Betrag; oben warnt ein Banner vor der
nicht konfigurierten TSE.
- **[Produkte verwalten](Screenshot_20260626-2244
44.png)** (`/admin/produkte`) — Die Produktverwal
tung listet alle Produkte als aufklappbare Karten
, wobei „Bratwurst" geöffnet ist und seine aktive
n Varianten mit Preisen samt Bearbeiten-/Lösch-Ak
tionen und dem Button „+ Neues Produkt" zeigt.
- **[Benutzer verwalten](Screenshot_20260626-2244
21.png)** (`/admin/benutzer`) — Die Benutzerverwa
ltung zeigt alle Benutzer als Liste mit Aktiv-Sch
alter, Rollen-Stern (Admin/Serviceleitung), Erste
llungsdatum sowie Bearbeiten-/Lösch-Icons und dem
 Button „+ Neuer Benutzer".
- **[Kassensitzung – Geldtransit](Screenshot_2026
0626-224527.png)** (`/admin/kasse`) — Die offene
Kassensitzung #3 („Sommerfest 26 Sonntag") mit de
m leeren Formular „Geldtransit buchen" zum Erfass
en von Einlage oder Entnahme (Richtung, Betrag, K
ommentar).
- **[Finanzamt – Stammdaten & Kassenidentität](Sc
reenshot_20260626-224332.png)** (`/admin/finanzam
t`) — Die Finanzamt-Seite mit ausgefülltem Betrei
ber-Stammdaten-Formular (Vereinsname, Adresse, St
euernummer, USt-ID) und der darunter liegenden Ka
ssenidentität mit eindeutiger Seriennummer und An
legedatum.

## Service – Kellner

- **[Meine Tische (Übersicht)](Screenshot_2026062
6-225003.png)** (`/service/tische`) — Die Service
-Startseite „Meine Tische" mit den persönlichen K
ennzahlen (Bestellungen, Kassiert), der Karte „De
ine offenen Tische" und den markierten Tisch-Kart
en samt Saldo und „offen"-Badge.
- **[Tischauswahl – Drawer „Alle Tische"](Screens
hot_20260626-224601.png)** (`/service/tische`) —
Der geöffnete Drawer „Alle Tische" mit fokussiert
em Suchfeld und der vollständigen Tischliste (Tis
ch 1–10), jeweils mit Favoriten-Stern und offenem
 Betrag.
- **[Tisch – Bestellen (Produktauswahl)](Screensh
ot_20260626-224641.png)** (`/service/tische/:id`)
 — Die Einzeltisch-Ansicht von „Tisch 4" im Reite
r „Bestellen" mit der nach Kategorien gruppierten
, eingeklappten Produktauswahl und der noch leere
n Aktionsleiste „Bestellung überprüfen" (0,00 €).
- **[Tisch – Bestellen (Artikel ausgewählt)](Scre
enshot_20260626-224654.png)** (`/service/tische/:
id`) — Im Reiter „Bestellen" ist die Kategorie „S
alat" aufgeklappt und es werden Artikel über Meng
en-Stepper hinzugefügt, sodass die grüne Aktionsl
eiste 3 Positionen für 14,50 € zeigt.
- **[Tisch – Bestellung überprüfen](Screenshot_20
260626-224724.png)** (`/service/tische/4`) — Der
geöffnete Bestellung-Drawer für „Tisch 4" fasst d
ie zusammengestellte Bestellung (vier Positionen,
 Gesamt 21,90 €) zusammen und bietet ein optional
es Kommentarfeld sowie „Bestellung aufnehmen"/„Ab
brechen".
- **[Tisch – Kassieren (Positionen wählen)](Scree
nshot_20260626-224756.png)** (`/service/tische/4`
) — Der Reiter „Kassieren" von „Tisch 4" zeigt di
e offenen Positionen mit Mengen-Steppern, wobei 2
 Artikel (6,30 €) zum Kassieren ausgewählt sind u
nd die grüne „Kassieren"-Leiste aktiv ist.
- **[Tisch – Zahlung](Screenshot_20260626-224828.
png)** (`/service/tische/4`) — Der Zahlung-Drawer
 für „Tisch 4" mit den Euro-Eingabefeldern „inklu
sive Trinkgeld" (7,00) und „Erhalten" (10,00) sow
ie automatisch berechnetem Rückgeld (3,00 €) und
Trinkgeld (0,70 €) vor dem Abschluss über „Kassie
ren".
- **[Tisch – Stornierung](Screenshot_20260626-224
926.png)** (`/service/tische/:id`) — Der Storno-D
rawer „Stornierung aus Vorgang …" zum positionswe
isen Stornieren, in dem 1× Kaffee Espresso (1,80
€) gewählt ist und der Button „Stornierung erteil
en" bis zum Pflicht-Kommentar deaktiviert bleibt.
- **[Tisch – Historie](Screenshot_20260626-224950
.png)** (`/service/tische/:id`) — Der Reiter „His
torie" von „Tisch 4" listet alle Vorgänge (Storni
erung, Zahlung, Bestellung, Ausgabe, Umbuchung) c
hronologisch mit Betrag, Bearbeiter und Zeitstemp
el sowie Aktions-Icons zum Ansehen, Stornieren un
d Umbuchen.
- **[Direktverkauf – Verkaufen](Screenshot_202606
26-225248.png)** (`/service/direktverkauf`) — Der
 Direktverkauf im Reiter „Verkaufen" mit Verkaufs
zusammenfassung (Gesamt 26,30 €, Erhalten 30,00 €
, Rückgeld 3,70 €), Kommentarfeld und Button „Ver
kauf abschließen" über der aufklappbaren Produktl
iste.
- **[Direktverkauf – Historie](Screenshot_2026062
6-225326.png)** (`/service/direktverkauf`) — Der
Direktverkauf im Reiter „Historie" listet abgesch
lossene Verkäufe mit Betrag, Zeitstempel, Benutze
r und Druck-/Storno-Buttons, wobei der oberste Ve
rkauf eine durchgeführte Teil-Stornierung von -2,
00 € zeigt.