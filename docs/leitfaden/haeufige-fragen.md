---
title: Häufige Fragen
description: 'Kurze Antworten auf häufige Fragen zum jotti-Betrieb: Internet, WLAN, Stromausfall, TSE-Pflicht, Österreich und Schweiz, Geräteanzahl, Trinkgeld, Helfer-Verzehr, Demo und Testen, Drucker, erster Login, Zugang, Version, mehrtägige Feste und Kosten.'
---

**Brauchen wir beim Fest Internet?** Ja, durch die TSE braucht jotti eine Internetverbindung. Fällt sie kurzzeitig aus, läuft der Verkauf trotzdem weiter (siehe nächste Frage).

Ohne Internet kann auch die grüne Adresse ausfallen. Die Handys arbeiten dann
über die [Fallback-Adresse](fehlersuche.md#fallback-adresse) weiter.

**Was, wenn das Internet oder die TSE ausfällt?** Weiterverkaufen ist erlaubt.
jotti bucht ganz normal weiter. Sobald die Verbindung zurück ist, signiert jotti
die Vorgänge automatisch nach und dokumentiert die Störung dabei automatisch.

Nachsignierte Belege tragen den Vermerk „Nachsigniert am …". Ihr müsst nichts
weiter tun.

**Was, wenn die grüne Adresse nicht lädt?** Mit der Fallback-Adresse weiterarbeiten
und die [Fehlersuche](fehlersuche.md) durchgehen. Der Verkauf muss nie warten.

**Was, wenn das WLAN am Fest schwach ist?** Jedes Handy braucht WLAN-Verbindung
zum Kassenrechner. Ohne sie bucht es nichts; einen Offline-Modus gibt es nicht.

Bricht das Absenden ab, bleibt die Bestellung im Korb. Tippt ohne Änderung erneut
auf „Bestellung aufnehmen"; jotti bucht sie dabei nicht doppelt.

Reicht das WLAN nicht über das ganze Gelände, erreichen die Handys jotti auch
über Mobilfunk. Das geht nur mit eigenem Server (siehe [Self-Hosting](self-hosting.md)).

**Was passiert bei einem Stromausfall?** Der Kassenrechner geht aus, eure Daten
bleiben erhalten. Die Kassensitzung bleibt offen, ihr macht danach einfach weiter.

Kommt der Strom zurück, startet ihr `jotti-start.exe` und für Bons `jotti-relay.exe`
neu. Die Status-Seite zeigt danach die dann gültige Adresse.

Eine noch nicht abgeschickte Bestellung bleibt im Korb, solange die Seite auf dem
Handy offen bleibt. Schickt sie ab, sobald jotti wieder läuft.

Eine USV für den Kassenrechner und den Router überbrückt kurze Ausfälle. Der
Router kann notfalls auch an einer Powerbank hängen. Für einen Stundenbetrieb
reicht das nicht.

**Brauchen wir wirklich eine TSE, obwohl wir gemeinnützig sind?** Ja. Die
Kassenpflichten gelten unabhängig von der Gemeinnützigkeit, sobald ihr Speisen oder
Getränke gegen Geld verkauft.

**Dürfen Vereine aus Österreich oder der Schweiz jotti nutzen?** Ja, die Nutzung
ist erlaubt. jotti setzt aber nur das deutsche Kassenrecht um, mit einer deutschen
Cloud-TSE.

Für die österreichische Registrierkassenpflicht (RKSV) und die Schweiz gibt es
keine Anbindung. jotti rechnet in Euro und nur mit den deutschen Steuersätzen.
Ob jotti das dort geltende Recht erfüllt, prüft ihr selbst.

**Wie viele Handys können gleichzeitig mitmachen?** Jede Servicekraft nutzt ihr
eigenes Handy (BYOD). jotti ist für Teams bis 30 Helfer ausgelegt — die übliche
Größe bei einem Vereinsfest.

Jeder Tisch bucht für sich, viele Handys bremsen sich also nicht gegenseitig.

Die TSE signiert die Vorgänge der Reihe nach, rund 0,3 Sekunden je Vorgang. Bei
Andrang wartet ein Kassenbeleg deshalb einige Sekunden. Bestellen und Kassieren
laufen ohne Pause weiter.

**Was passiert mit Trinkgeld?** jotti zeigt das Trinkgeld beim Kassieren nur zur
Info an, gebucht wird es nicht.

Es gehört deshalb nicht in die Kassenlade, sonst geht der Kassensturz am Abend
nicht auf. Trinkgeld direkt an die Servicekraft ist in der Regel steuerfrei.

**Wie erfassen wir den Verzehr unserer Helfer?** Über einen eigenen Tisch pro
Helfer. Bestellungen laufen normal darüber. Am Abend kassiert oder storniert
ihr diese Tische wie jeden anderen.

**Gibt es eine Demo oder ein Testsystem?** Auf der Startseite von jotti.rocks
könnt ihr eine Demo durchklicken. Zum echten Ausprobieren installiert ihr jotti
zuhause auf einem Windows-Rechner.

Die TSE richtet ihr dafür in der kostenlosen TEST-Umgebung von fiskaly ein (siehe
nächste Frage).

**Können wir die TSE erst testen?** Ja. Richtet sie zuerst in der kostenlosen
TEST-Umgebung von fiskaly ein und wechselt erst für den Echtbetrieb auf LIVE.

**Brauchen wir einen Bondrucker?** Ja. Auch eine bewilligte Befreiung von der
Aushändigung (siehe [Belege und Steuersätze](belege-steuersaetze.md)) erspart nur
das ungefragte Aushändigen. Verlangt ein Gast einen Bon, müsst ihr ihn drucken können.

Einen Bon per E-Mail gibt es nicht; das ist bewusst kein Ziel von jotti
(siehe [Nicht-Ziele](../anforderungen.md#nicht-ziele)). Passende Drucker nennt
[Bondruck einrichten](installation.md#bondruck-einrichten).

**Wo stehen die Anmeldedaten für den ersten Login?** In der Startkonsole, dem
Fenster von `jotti-start.exe`. Dort steht ein 6-stelliger Code für den Benutzer
`admin` (siehe [Erster Login](installation.md#erster-login)).

Mit dem Code wählt ihr „Neues Passwort festlegen" und vergebt ein eigenes
Passwort. Ist das Fenster schon zu, startet jotti neu; dann erscheint ein neuer Code.

**Wo bekommen wir Zugangsdaten, Freigabecode oder Lizenzschlüssel?** Nirgends,
jotti braucht nichts davon. Die Nutzungsvereinbarung schließt ihr mit einer
einzigen E-Mail ab (siehe [Lizenzmodell](../lizenzmodell.md)).

Danach dürft ihr sofort installieren; auf eine Antwort müsst ihr nicht warten.
jotti läuft auf eurem Rechner, ein Konto braucht ihr nur bei fiskaly für die TSE.

Antwortet der Autor doch, landet die E-Mail manchmal im Spam-Ordner. Schaut dort
nach, wenn ihr eine Antwort erwartet.

**Nehmen wir für unser Fest die aktuelle Version oder warten wir auf 1.0?** Nehmt
die aktuelle Version. Auf 1.0 aktualisiert ihr später wie bei jedem Update; eure
Daten bleiben erhalten.

Aktualisiert zuhause vor dem Fest, nicht am Veranstaltungstag (siehe
[Aktualisieren](aktualisieren.md)).

**Welche Rechtsform gibt eine Ortsgruppe eines Verbands an?** Wählt im Formular
„Untergliederung eines eingetragenen Verbands". Eine eigene Eintragung braucht die
Ortsgruppe nicht (siehe [Nutzungsbedingungen § 2](../../TERMS.md)).

Die Ortsgruppe schließt die Vereinbarung im eigenen Namen. Als Namen nennt ihr
Ortsgruppe und Verband, etwa: Ortsgruppe Musterhausen im Landesverband Musterland e.V.

**Darf ein Partnerverein unsere Installation mitnutzen?** Nein. Jeder Verein schließt
seine eigene Nutzungsvereinbarung und betreibt eine eigene Installation.

Jeder Beleg, der Z-Bon und der Export für das Finanzamt tragen Namen und Steuernummer
genau eines Betreibers. Verkäufe eines anderen Vereins stünden sonst unter eurem Namen.

**Wie läuft ein mehrtägiges Fest?** Jeder Festtag bekommt eine eigene
Kassensitzung. Den Ablauf beschreibt [Mehrtägiges Fest](veranstaltungstag.md#mehrtägiges-fest).

**Was kostet der Betrieb?** jotti ist für euch kostenlos. Laufende Kosten entstehen
nur für die Cloud-TSE von fiskaly und, beim Experten-Weg, für den VPS
(ca. 6 €/Monat). Der Standardweg kommt ohne Servermiete aus.

fiskaly veröffentlicht für die TSE keinen Preis. Ein Wiederverkäufer (HKSoftware)
nennt rund 13 €/Monat oder 143 €/Jahr netto je Kasse. Er bündelt die TSE mit
eigener Kassensoftware; der Preis ist nur eine Orientierung.

Den Vertrag schließt ihr direkt mit fiskaly; holt dort ein Angebot ein.
Mindestlaufzeit und Mindestabnahme nennt fiskaly nicht öffentlich; ein Verein
berichtet von beidem — fragt vor Vertragsschluss danach.

Stand der Recherche: 07.09.2026.
