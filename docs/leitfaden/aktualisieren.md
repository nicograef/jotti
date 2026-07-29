---
title: Aktualisieren (Standardweg)
description: 'Die jotti-Kasse auf dem Windows-Rechner auf eine neue Version bringen: drei Doppelklicks, das automatische Neuladen der Geräte danach, Rauchtest, automatisches Backup und der Weg zurück.'
---

Meldet der Starter beim Hochfahren „Neue Version verfügbar" mit einem
Download-Link, bringt ihr jotti in drei Schritten auf den neuen Stand. Eure Daten
bleiben dabei erhalten.

> ⚠️ **Zuhause mit Internet aktualisieren, nicht auf dem Fest.** Wie beim
> Erststart lädt jotti dabei neue Programmteile herunter. Erledigt das Update in
> Ruhe vorab, nicht erst am Veranstaltungstag.

## In drei Schritten

1. **`jotti-stop.cmd`** doppelklicken, um das laufende jotti sauber zu beenden.
2. Das **neue Release-ZIP entpacken**. Der Ort ist egal: Es muss nicht derselbe
   Ordner wie vorher sein, ein frischer Ordner ist völlig in Ordnung.
3. **`jotti-start.exe`** im neuen Ordner doppelklicken und die UAC-Abfrage mit
   „Ja" bestätigen.

Den alten Programmordner könnt ihr danach gefahrlos löschen.

> ⛔ **Während eines Updates keine TSE-Einrichtung starten.** Die Einrichtung legt
> bei fiskaly eine TSE an — in LIVE eine kostenpflichtige, in TEST eine kostenlose —
> und zeigt Admin-PUK und Admin-PIN genau einmal an, am Ende des Ablaufs. Wird
> jotti mittendrin beendet — und genau das tut `jotti-stop.cmd` —, bricht sie ab.
> Meistens ist das kein Beinbruch: Startet ihr den Assistenten danach erneut,
> erkennt jotti den tatsächlichen Zustand bei fiskaly und macht dort weiter, wo es
> aufgehört hat; eine zweite TSE entsteht dabei nicht. Nur ein schmales Zeitfenster
> ist eine Sackgasse: Bricht die Einrichtung genau zwischen dem Setzen der
> Admin-PIN und der Anzeige des Ergebnisses ab, fragt die Wiederaufnahme nach einer
> Admin-PIN, die euch nie angezeigt wurde. Dann hilft in TEST „Stattdessen neue TSE
> anlegen", in LIVE der fiskaly-Support (siehe
> [TSE-Sonderfälle](tse-sonderfaelle.md)). Richtet die TSE also vor dem Update ein
> oder danach, nie währenddessen — und wenn der Assistent gerade läuft, wartet mit
> dem Update, bis „Verbindung bestätigt" steht.

## Reihenfolge, wenn ihr mitten im Fest aktualisieren müsst

Aktualisiert nach Möglichkeit zuhause (siehe Kasten oben). Muss es doch während
des laufenden Betriebs sein, haltet euch an diese Reihenfolge.

1. **Ansagen, bevor ihr etwas anfasst.** „Zwei Minuten Pause, bitte keine neue
   Bestellung anfangen." Wer gerade tippt, schickt die Bestellung vorher ab; wer
   gerade kassiert, macht das fertig.
2. **Die Kassensitzung bleibt offen.** Für ein Update braucht es keinen
   Tagesabschluss. Die Sitzung liegt in der Datenbank und übersteht das Update
   unverändert.
3. **Die drei Schritte von oben ausführen:** `jotti-stop.cmd`, neues ZIP
   entpacken, `jotti-start.exe`. In dieser Zeit erreicht kein Handy die Kasse.
4. **Warten, bis der Starter „jotti laeuft" meldet.** Erst dann ist der neue
   Stand wirklich da.
5. **Kurz warten, dann den [Rauchtest](#rauchtest-fünf-abläufe-vor-dem-weitermachen)
   machen.** Die Bedienoberfläche am Rechner lädt sich von selbst neu, sobald jotti
   wieder läuft; rechnet mit einer halben Minute (siehe nächster Abschnitt).
6. **Erst danach dem Team Bescheid geben**, dass es weitergeht. Ein Neuladen müsst
   ihr nicht ansagen — die Handys laden sich genauso von selbst neu. Einzige
   Ausnahme ist das Update **auf** Version 0.17.3: Dabei muss jedes Gerät noch ein
   letztes Mal von Hand neu geladen werden (siehe Kasten im nächsten Abschnitt).

## Danach: die Geräte laden sich von selbst neu

Handys und Rechner behalten die alte Bedienoberfläche im Speicher, bis die Seite
einmal neu geladen wird. **Seit Version 0.17.3 erledigt jotti das allein.** Jedes
geöffnete jotti fragt im Hintergrund alle halbe Minute nach, welche Version auf
dem Rechner läuft — und zusätzlich immer dann, wenn ein weggelegtes Handy wieder
hervorgeholt wird. Weicht die Version ab, lädt sich die Seite selbst neu: im
Browser am Handy, am Rechner und ebenso in der als App auf dem Startbildschirm
installierten jotti. **Ihr müsst das Neuladen also nicht mehr ansagen**, und
niemand muss eine App wegwischen.

> ⚠️ **Beim Update auf 0.17.3 noch ein letztes Mal von Hand neu laden.** Das
> automatische Neuladen steckt in der Bedienoberfläche selbst — und die alte
> Bedienoberfläche, die auf den Geräten noch im Speicher liegt, kennt es nicht.
> Bei genau diesem Update erneuert sich deshalb kein Gerät von allein, auch nicht
> der Rechner, an dem ihr gleich den Rauchtest macht. Ladet einmal überall von Hand
> neu:
>
> - **Handy im Browser:** die Seite von ganz oben nach unten ziehen.
> - **Rechner:** `Strg` + `F5`.
> - **Als App auf dem Startbildschirm installiert:** Dort gibt es weder Adresszeile
>   noch Neu-laden-Pfeil. Die App **ganz schließen** und aus der Übersicht der
>   laufenden Apps wegwischen, dann neu öffnen.
>
> Beim nächsten Update, das dann von 0.17.3 aus startet, erledigt jotti es allein.

**Wer gerade mitten in etwas steckt, verliert nichts.** Solange ein angefangener
Vorgang offen ist — ein gefüllter Bestellkorb, eine getroffene Auswahl beim
Kassieren oder Stornieren, ein halb ausgefülltes Formular —, wartet das
Neuladen. Stattdessen erscheint auf dem Gerät ein farbiges Band mit diesem Text:

> Der Server läuft mit einer anderen Version als diese Seite. Bitte den laufenden
> Vorgang abschließen oder verwerfen — danach lädt sich die Seite von selbst neu.

Mit „Server" ist der Windows-Rechner gemeint, auf dem jotti läuft. Der offene
Vorgang muss also weg — und zwar wirklich weg: die angefangene Bestellung
abschicken, das Kassieren zu Ende führen oder die gewählten Mengen mit den
Minus-Knöpfen wieder auf null stellen. **Beim Bestellen und Kassieren genügt
„Abbrechen" nicht** — das blendet nur die Eingabe aus, der Korb bleibt gefüllt und
das Band steht weiter. Bedienen lässt sich in der Zwischenzeit alles wie gewohnt.
Wegklicken lässt sich das Band nicht — es verschwindet von selbst.

**Der Ausnahmefall: „Jetzt neu laden".** Steht auf dem Band stattdessen dieser
Text, und daneben eine Schaltfläche, ist das automatische Neuladen nicht
durchgekommen:

> Der Server läuft mit einer anderen Version als diese Seite. Das automatische
> Neuladen hat nicht geklappt — bitte von Hand neu laden.

Dann genügt ein Antippen von **„Jetzt neu laden"**. Das kommt selten vor; ansagen
müsst ihr dafür nichts. Kommt das Band danach noch einmal, wartet einen Moment und
tippt erneut.

## Das Print-Relay bleibt bei Version 0.17.3, wie es ist

Das Print-Relay (`jotti-relay.exe`, das Fenster, das die Bons an die Drucker
schickt) ist seit Version 0.17.1 unverändert, und auch die Verständigung
zwischen jotti und dem Relay hat sich nicht geändert. **Das laufende Relay darf
einfach weiterlaufen** — ihr müsst es weder beenden noch ersetzen. Im
Release-ZIP liegt trotzdem eine `jotti-relay.exe`; sie ist funktional identisch
mit der laufenden. Ob ihr sie tauscht oder nicht, macht keinen Unterschied.

> ⚠️ **Beim nächsten Update nicht raten.** Ändert sich das Relay, muss das alte
> Fenster **erst geschlossen** und dann die neue `jotti-relay.exe` gestartet
> werden. Laufen zwei Relays gleichzeitig, druckt jedes jeden Bon — also alles
> doppelt.

## Rauchtest: fünf Abläufe vor dem Weitermachen

Bevor der Betrieb wieder anläuft, einmal komplett durchspielen. Dauert zwei
Minuten und deckt alles ab, worauf sich das Fest verlässt:

- [ ] An einem Tisch eine **Bestellung aufnehmen**
- [ ] Denselben Tisch **kassieren**
- [ ] Aus dem Vorgang eine Position **stornieren** (geht nur mit der Rolle
      Serviceleitung oder als Admin)
- [ ] Einen **Direktverkauf** buchen (im Service-Bereich über das Benutzermenü
      oben rechts „Zu Direktverkauf wechseln", siehe
      [Modus wechseln](betriebsarten.md#modus-wechseln))
- [ ] Im Admin-Bereich **„Übersicht"** öffnen und prüfen, dass die Stornierung
      dort einer Servicekraft **zugeordnet** ist: Im Block „Team" trägt die
      Servicekraft, die kassiert hatte, die rote Markierung „1 Storno"; in der
      Storno-Zeile darunter steht „Betroffen: " und dahinter ihr Name mit der
      Anzahl. „Details" klappt den einzelnen Eintrag auf.

Fällt einer der fünf Punkte durch, macht **nicht** weiter, sondern schaut in
[Fehlersuche](fehlersuche.md); der Weg zurück steht weiter unten.

## Warum eure Daten erhalten bleiben

Bestellungen, Benutzer, Produkte, der Installations-Schlüssel und das grüne
Zertifikat liegen nicht im Programmordner, sondern geschützt außerhalb (in
Docker-Volumes und unter `C:\ProgramData\jotti`). Egal wohin ihr das neue ZIP
entpackt: jotti findet die Daten beim Start wieder.

**Automatisches Backup vor dem Update.** Erkennt der Starter eine neue Version,
sichert er die Datenbank automatisch, bevor er die Aktualisierung ausführt. Diese
Sicherung landet im geschützten Datenbereich und zusätzlich im Ordner
`C:\ProgramData\jotti\backups` (die letzten fünf werden vorgehalten). Geht beim
Update etwas schief, ist dieses Backup euer Rückweg — wie ihr es einspielt, steht
unter [Der Weg zurück](#der-weg-zurück-wenn-das-update-schiefgeht).

> 🔁 **Nur vorwärts, kein Downgrade.** Spielt keine ältere Version über eine
> neuere. Updates verändern die Datenbank; eine alte Version kann mit den neuen
> Daten nicht mehr starten. Der Starter verweigert einen solchen Rückschritt
> selbst.

> ⛔ **Niemals `docker compose down -v` ausführen.** Das `-v` löscht alle
> Docker-Volumes und damit Daten, Installations-Schlüssel und Zertifikat
> unwiderruflich — auch ein Update bringt sie dann nicht zurück. Zum Beenden immer
> `jotti-stop.cmd` verwenden.

## Der Weg zurück, wenn das Update schiefgeht

Der Rückweg ist **nicht**, einfach das alte ZIP wieder auszupacken. Das
funktioniert nicht: `jotti-start.exe` verweigert den Start einer älteren Version
mit der Meldung „Start verweigert: Diese Version … ist aelter als die zuletzt
gestartete …". Updates verändern die Datenbank, und diese Änderung wird nicht
zurückgenommen.

Der Rückweg ist das automatische Backup von vor dem Update:

1. **`jotti-restore.cmd`** doppelklicken (liegt im entpackten Release-Ordner,
   neben `jotti-start.exe`).
2. Die Rückfrage **`Fortfahren? (j/N)`** mit **`j`** beantworten.
3. Das Skript startet die Datenbank, hält die Anwendung währenddessen an, spielt
   das neueste automatische Backup ein und startet jotti wieder. Am Ende meldet
   es „Wiederherstellung abgeschlossen. jotti laeuft wieder."

> ⚠️ **Alles seit dem Backup ist danach weg.** Das Backup entsteht unmittelbar
> vor dem Update. Aktualisiert ihr mitten im Fest, verliert ihr also jede
> Bestellung und jede Zahlung aus der Zeit nach dem Update. Nehmt diesen Weg nur,
> wenn jotti wirklich nicht mehr arbeitsfähig ist — und sagt vorher im Team an,
> was verloren geht.

## Wenn nach dem Update niemand mehr hineinkommt

Sehr selten passt nach einem Update das in der Datenbank gespeicherte Passwort
nicht mehr zum Installations-Schlüssel; jotti startet dann, aber das Anmelden
schlägt fehl. Eure Daten sind dabei nicht verloren. **`jotti-repair.cmd`**
doppelklicken gleicht beides datenerhaltend wieder an und startet jotti neu;
danach einmal neu anmelden. Mehrfaches Ausführen schadet nicht. Mehr dazu unter
[Fehlersuche](fehlersuche.md#nach-einem-update-klappt-das-anmelden-nicht).
