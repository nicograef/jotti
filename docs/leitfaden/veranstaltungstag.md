---
title: Der Veranstaltungstag
description: 'Der Ablauf am Festtag: Kassensitzung mit Anfangsbestand eröffnen, Wechselgeld als Geldtransit buchen und am Abend Kassensturz und Tagesabschluss (Z-Bon) in einem Schritt erstellen.'
---

Am Veranstaltungstag arbeitet ihr mit einer Kassensitzung. Sie beginnt mit dem
gezählten Anfangsbestand, hält alle Verkäufe und Bargeldbewegungen zusammen und
endet mit dem Tagesabschluss.

Die Servicekräfte nehmen währenddessen ganz normal auf ihren Handys Bestellungen
auf, geben aus und kassieren. Die folgenden Schritte macht ein Admin im
Admin-Bereich unter „Kassentag".

An der Theke läuft der Verkauf statt über Tische im Direktverkauf; wann welcher
Arbeitsmodus passt, steht unter [Welcher Modus passt?](betriebsarten.md).

## Vor dem ersten Verkauf: Kassensitzung eröffnen

Zählt das Wechselgeld in der Kasse und öffnet im Admin-Bereich unter „Kassentag"
die Sitzung über „Kassensitzung eröffnen". Tragt den gezählten Betrag als
Anfangsbestand ein.

Ohne offene Kassensitzung nimmt jotti keine Zahlungen an, öffnet sie also vor dem
ersten Verkauf. Ab jetzt rechnet jotti alle Ein- und Auszahlungen gegen diesen
Startwert.

## Während des Fests: Wechselgeld und Entnahmen buchen

Braucht die Kasse zwischendurch mehr Wechselgeld oder nehmt ihr Bargeld heraus?
Dann bucht ihr das über die Buttons „Geld einlegen" und „Geld entnehmen":

- zusätzliches Wechselgeld als Einlage
- herausgenommenes Bargeld als Entnahme

So bleibt der erwartete Kassenbestand (Soll-Bestand) korrekt, und der
Kassensturz am Abend geht auf.

> 💡 **Trinkgeld gehört nicht in die Kasse.** jotti zeigt Trinkgeld beim Kassieren
> nur zur Info an und bucht es nicht. Landet es trotzdem in der Kassenlade, geht
> der Kassensturz am Abend nicht auf.

## Nach dem Fest: Kasse abschließen

Zählt am Ende das gesamte Bargeld in der Kasse und öffnet unter „Kassentag" den
Bereich „Kasse abschließen". Tragt den gezählten Ist-Bestand ein.

jotti zeigt vorab, ob Soll- und Ist-Bestand übereinstimmen. Dann erstellt jotti
Kassensturz und Tagesabschluss (Z-Bon) in einem Schritt. Eine Differenz ist kein
Fehler, sondern wird als Kassendifferenz festgehalten.

Danach sichert ihr die Daten des Tages. Nach jedem Veranstaltungstag erzeugt ihr
den DSFinV-K-Export der abgeschlossenen Kassensitzung. Den legt ihr an zwei
getrennten Orten ab. Die Details stehen unter [Datenaufbewahrung](datenaufbewahrung.md).

## Mehrtägiges Fest

Jeder Festtag bekommt eine eigene Kassensitzung. Am Abend schließt ihr die Kasse
wie oben ab, am nächsten Morgen eröffnet ihr eine neue.

Vor dem Abschluss müssen alle Tische kassiert oder storniert sein. Offene Beträge
lassen sich nicht in den nächsten Tag mitnehmen.

Habt ihr jotti über Nacht mit `jotti-stop.cmd` beendet, startet ihr am Morgen
`jotti-start.exe` und für Bons `jotti-relay.exe`. Zählt das Wechselgeld neu und
tragt es beim Eröffnen als Anfangsbestand ein.

Eine Anmeldung gilt 12 Stunden. Die Helfer melden sich deshalb jeden Morgen neu an.
Produkte, Tische und Benutzer bleiben über alle Tage erhalten.
