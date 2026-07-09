---
title: TSE-Sonderfälle
description: 'Seltene TSE-Fälle: vorhandene TSS übernehmen, Setup nach Abbruch fortsetzen, PIN per PUK zurücksetzen, Test-Limit und manuelle Konfiguration.'
---

Die folgenden Fälle braucht ihr nur, wenn etwas vom Normalfall abweicht.

**Vorhandene TSS übernehmen.** Findet jotti im Konto bereits eine TSS, bietet es
„TSE übernehmen" an, statt eine zweite anzulegen. Das schützt vor versehentlicher
Doppel-Anlage und nimmt ein abgebrochenes Setup dort wieder auf, wo es stehen
geblieben ist. Ist die TSS bereits personalisiert, fragt jotti nach der verwahrten
Admin-PIN.

**Wiederaufnahme nach Abbruch.** Bricht die Einrichtung ab (Netzfehler, Browser
geschlossen), startet ihr den Assistenten einfach erneut. jotti erkennt den
tatsächlichen Zustand bei fiskaly und holt nur die fehlenden Schritte nach. Es
entsteht keine zweite TSS. Nur in einem seltenen Fall — bricht die Einrichtung
genau zwischen dem Setzen der Admin-PIN und der Ergebnis-Anzeige ab — fragt die
Wiederaufnahme nach einer Admin-PIN, die euch nie angezeigt wurde; dann hilft in
TEST „Stattdessen neue TSE anlegen", in LIVE der fiskaly-Support.

**PIN per PUK zurücksetzen.** Verlangt jotti die Admin-PIN und habt ihr sie nicht
(oder hat fiskaly sie nach fünf Fehlversuchen gesperrt), bietet der Assistent „Ich
habe den Admin-PUK" an. Gebt dort den verwahrten Admin-PUK ein und klickt „PIN
zurücksetzen und übernehmen": jotti setzt eine neue Admin-PIN und schließt die
Übernahme ab, ohne neue, kostenpflichtige TSS. Danach zeigt jotti die neue
Admin-PIN einmalig an — sofort notieren, die alte Notiz ersetzen und das Häkchen
„Ich habe die neue Admin-PIN sicher verwahrt." setzen; der Admin-PUK bleibt
unverändert gültig. Das funktioniert in TEST und LIVE. Sind PUK und PIN beide
verloren, hilft nur der fiskaly-Support.

**Test-Limit und Selbstreinigung.** Die Test-Umgebung erlaubt höchstens fünf aktive
TSE. Habt ihr beim Üben fünf erreicht, übernehmt eine vorhandene oder wartet die
automatische Bereinigung ab (fiskaly löscht stillgelegte oder länger als 14 Tage
ungenutzte Test-TSE regelmäßig). Liegt die PIN einer vorhandenen Test-TSE nicht mehr
vor, bietet jotti nur in TEST die Sekundäraktion „Stattdessen neue TSE anlegen" an.
In LIVE gibt es diesen Ausweg nicht; dort helfen der PUK-Reset, die verwahrte PIN
oder der fiskaly-Support.

**TSE unter v0.14.0 eingerichtet?** jotti v0.14.0 hat beim Einrichten eine leere
TSE-Seriennummer gespeichert; im DSFinV-K-Export bliebe dadurch das Pflichtfeld
`TSE_SERIAL` leer (die Exportprüfung meldet das als Verstoß). Nach dem Update
auf v0.15.0 lauft ihr vor dem ersten DSFinV-K-Export einmal den
TSE-Assistenten erneut durch („TSE übernehmen", siehe oben) — dabei zieht jotti
die Seriennummer nach; es entsteht keine zweite TSS.

**Manuelle Konfiguration (Experten).** Habt ihr eine TSS samt Client bereits
außerhalb von jotti angelegt, öffnet ihr im Admin-Bereich „Finanzamt", klickt im
Kasten „TSE-Anbindung" auf „Einrichten oder ändern" und tragt im Kasten „Manuelle
Konfiguration (Experten)" API-Key, API-Secret, TSS-ID und Client-ID direkt ein
(alle vier sind Pflicht), speichert und klickt „Verbindung testen". Der Client muss bei
fiskaly mit der Kassen-Seriennummer aus jottis Kassenidentität registriert sein,
sonst meldet der Test einen Fehler. Mit „Alle Felder leeren" entfernt ihr die
Konfiguration wieder, etwa zur Schlüsselrotation.
