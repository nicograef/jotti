---
title: TSE-Sonderfälle
description: 'Seltene TSE-Fälle: vorhandene TSS übernehmen, Setup nach Abbruch fortsetzen, PIN per PUK zurücksetzen, Test-Limit und manuelle Konfiguration.'
---

Die folgenden Fälle braucht ihr nur, wenn etwas vom Normalfall abweicht.

**Vorhandene TSS übernehmen.** Findet jotti im Konto bereits eine TSS, bietet es
„TSE übernehmen" an, statt eine zweite anzulegen.

Das schützt vor versehentlicher Doppel-Anlage. Außerdem nimmt es ein abgebrochenes
Setup dort wieder auf, wo es stehen geblieben ist.

Ist die TSS bereits personalisiert und noch nicht einsatzbereit, fragt jotti nach
der verwahrten Admin-PIN. Nicht einsatzbereit heißt: nicht initialisiert oder
diese Kasse dort nicht angemeldet.

**Wiederaufnahme nach Abbruch.** Bricht die Einrichtung ab (Netzfehler, Browser
geschlossen), startet ihr den Assistenten einfach erneut. jotti erkennt den
tatsächlichen Zustand bei fiskaly und holt nur die fehlenden Schritte nach.

Es entsteht keine zweite TSS. Nur ein seltener Fall ist anders: ein Abbruch genau
zwischen dem Setzen der Admin-PIN und der Ergebnis-Anzeige. Dann fragt die
Wiederaufnahme nach einer Admin-PIN, die euch nie angezeigt wurde.

In TEST hilft dann „Stattdessen neue TSE anlegen", in LIVE der fiskaly-Support.

**PIN per PUK zurücksetzen.** Verlangt jotti die Admin-PIN und habt ihr sie
nicht, bietet der Assistent „Ich habe den Admin-PUK" an. Dasselbe gilt, wenn
fiskaly die PIN nach fünf Fehlversuchen gesperrt hat.

Gebt dort den verwahrten Admin-PUK ein und klickt „PIN zurücksetzen und
übernehmen". jotti setzt eine neue Admin-PIN und schließt die Übernahme ab. Eine
neue, kostenpflichtige TSS entsteht dabei nicht.

Danach zeigt jotti die neue Admin-PIN einmalig an. Sofort notieren, die alte
Notiz ersetzen und das Häkchen „Ich habe die neue Admin-PIN sicher verwahrt."
setzen. Der Admin-PUK bleibt unverändert gültig.

Das funktioniert in TEST und LIVE. Sind PUK und PIN beide verloren, hilft nur
der fiskaly-Support.

**Test-Limit und Selbstreinigung.** Die Test-Umgebung erlaubt höchstens fünf aktive
TSE. Habt ihr beim Üben fünf erreicht, übernehmt eine vorhandene oder wartet die
automatische Bereinigung ab.

fiskaly löscht stillgelegte oder länger als 14 Tage ungenutzte Test-TSE
regelmäßig.

Liegt die PIN einer vorhandenen Test-TSE nicht mehr vor, gibt es nur in TEST einen
Ausweg. jotti bietet dort die Sekundäraktion „Stattdessen neue TSE anlegen" an.

In LIVE gibt es diesen Ausweg nicht; dort helfen der PUK-Reset, die verwahrte PIN
oder der fiskaly-Support.

**TSE unter v0.14.0 eingerichtet?** jotti v0.14.0 hat beim Einrichten eine leere
TSE-Seriennummer gespeichert. Im DSFinV-K-Export bliebe dadurch das Pflichtfeld
`TSE_SERIAL` leer (die Exportprüfung meldet das als Verstoß).

Nach dem Update auf v0.15.0 lauft ihr vor dem ersten DSFinV-K-Export einmal den
TSE-Assistenten erneut durch („TSE übernehmen", siehe oben). Dabei zieht jotti
die Seriennummer nach; es entsteht keine zweite TSS.

**Manuelle Konfiguration (Experten).** Habt ihr eine TSS samt Client bereits
außerhalb von jotti angelegt, geht ihr so vor:

1. Öffnet im Admin-Bereich „Finanzamt & TSE" den Assistenten über „TSE
   einrichten".
2. Tragt im Kasten „Manuelle Konfiguration (Experten)" API-Key, API-Secret,
   TSS-ID und Client-ID direkt ein. Alle vier sind Pflicht.
3. Speichert und klickt „Verbindung testen".

Der Client muss bei fiskaly mit der Kassen-Seriennummer aus jottis
Kassenidentität registriert sein, sonst meldet der Test einen Fehler. Mit „Alle
Felder leeren" entfernt ihr die Konfiguration wieder, etwa zur Schlüsselrotation.
