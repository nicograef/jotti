---
title: Aktualisieren (Standardweg)
description: 'Die jotti-Kasse auf dem Windows-Rechner auf eine neue Version bringen: drei Doppelklicks, das Neuladen der Geräte danach, Rauchtest, automatisches Backup und der Weg zurück.'
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

## Reihenfolge, wenn ihr mitten im Fest aktualisieren müsst

Aktualisiert nach Möglichkeit zuhause (siehe Kasten oben). Muss es doch während
des laufenden Betriebs sein, haltet euch an diese Reihenfolge. Sie ist nicht
beliebig: Ein Handy, das zu früh neu geladen wird, holt sich nur wieder den alten
Stand.

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
5. **Jetzt den Browser am Rechner neu laden** (siehe nächster Abschnitt) und den
   [Rauchtest](#rauchtest-fünf-abläufe-vor-dem-weitermachen) machen.
6. **Erst danach dem Team Bescheid geben**, dass es weitergeht — und ob die
   Handys ebenfalls neu geladen werden sollen. Vorher nicht: Ein Neuladen,
   während jotti aus ist, zeigt nur eine Fehlermeldung.

## Danach: jedes betroffene Gerät einmal neu laden

Handys und Rechner behalten die alte Bedienoberfläche im Speicher, bis die Seite
einmal neu geladen wird. **Von selbst passiert das nicht** — kein Gerät merkt,
dass auf dem Rechner inzwischen eine neue Version läuft. Genau deshalb muss das
Neuladen angesagt werden.

Einmal neu laden genügt, und es wirkt zuverlässig: jotti liefert seine Startseite
grundsätzlich ohne Zwischenspeicher aus, die Programmteile dahinter tragen ihre
Version im Dateinamen. Beim Neuladen fragt das Gerät die Startseite also wirklich
neu an und bekommt die neuen Programmteile mitgeliefert. Den Browser-Cache zu
leeren ist nicht nötig.

**So ladet ihr neu:**

- **Handy im Browser:** die Seite von ganz oben nach unten ziehen
  (Pull-to-Refresh) oder den Neu-laden-Pfeil neben der Adresszeile antippen.
- **Rechner:** `Strg` + `F5` oder den Neu-laden-Pfeil.
- **Wenn jotti als App auf dem Startbildschirm liegt:** jotti lässt sich als App
  installieren; dann läuft es ohne Adresszeile und ohne Neu-laden-Pfeil. In dem
  Fall die App **ganz schließen** — in der Übersicht der laufenden Apps
  wegwischen, nicht nur den Home-Knopf drücken — und danach neu öffnen. Beim
  Öffnen holt sich jotti die Startseite frisch. Wer unsicher ist, öffnet jotti
  stattdessen im normalen Browser über die Adresse vom QR-Code (steht auf der
  Statusseite <http://localhost:8484>) und lädt dort neu.

### Bei Version 0.17.2: nur der Rechner des Admins

Für dieses Update wurde geprüft, welche Geräte das Neuladen wirklich brauchen.

**Der Browser am Admin-Rechner muss neu geladen werden.** Ohne Neuladen brechen
drei Ansichten — nicht mit einer Fehlermeldung, sondern still und irreführend:

- **„Übersicht" (das Live-Dashboard):** zeigt „Keine Kassensitzung geöffnet",
  obwohl die Kasse offen ist, und wirft alle 30 Sekunden die Meldung „Daten
  konnten nicht geladen werden". Lasst euch davon **nicht** dazu verleiten, eine
  zweite Kassensitzung zu eröffnen — jotti lehnt das ohnehin ab.
- **„Berichte & Export":** die rechte Spalte dreht sich endlos, und der Knopf für
  den DSFinV-K-Export erscheint gar nicht erst.
- **„Kassentag":** der rote Hinweis „N Tische sind noch offen" verschwindet
  stillschweigend. Das ist die unauffälligste und zugleich ärgerlichste Folge:
  Wer den Tagesabschluss aus einem nicht neu geladenen Fenster macht, verliert
  diese Warnung. Der Tagesabschluss selbst funktioniert weiter.

**Die Handys der Servicekräfte müssen nicht neu geladen werden.** Bestellen,
Kassieren, Stornieren, Umbuchen und Direktverkauf laufen auf dem alten Stand
unverändert weiter; keine Service-Ansicht bricht. Es fehlt nur der neue Hinweis
in „Meine Tische", der eine zugeordnete Rücknahme erklärt („… wurde von deinen
Zahlungen zurückgegeben. Du gibst damit … ab."). Weil die Helfer damit
abrechnen, lohnt sich das Neuladen trotzdem — es ist bloß nicht dringend und darf
bis zur nächsten ruhigen Minute warten.

**Ohne jedes Zutun** wirkt dagegen die Reparatur der Tischübersicht: Ein
gelöschter Tisch legte bisher die Tischübersicht einzelner Servicekräfte lahm.
Das ist nach dem Update sofort behoben, auch auf Handys, die nicht neu geladen
wurden.

> ℹ️ **Diese Aufstellung gilt nur für Version 0.17.2.** Bei einem späteren Update
> im Zweifel alle Geräte einmal neu laden.

## Das Print-Relay bleibt bei Version 0.17.2, wie es ist

Das Print-Relay (`jotti-relay.exe`, das Fenster, das die Bons an die Drucker
schickt) ist gegenüber Version 0.17.1 unverändert, und auch die Verständigung
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
