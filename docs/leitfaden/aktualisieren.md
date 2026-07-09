---
title: Aktualisieren (Standardweg)
description: 'Die jotti-Kasse auf dem Windows-Rechner auf eine neue Version bringen: drei Doppelklicks, automatisches Backup vor dem Update, und was ihr niemals tun dürft.'
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

## Warum eure Daten erhalten bleiben

Bestellungen, Benutzer, Produkte, der Installations-Schlüssel und das grüne
Zertifikat liegen nicht im Programmordner, sondern geschützt außerhalb (in
Docker-Volumes und unter `C:\ProgramData\jotti`). Egal wohin ihr das neue ZIP
entpackt: jotti findet die Daten beim Start wieder.

**Automatisches Backup vor dem Update.** Erkennt der Starter eine neue Version,
sichert er die Datenbank automatisch, bevor er die Aktualisierung ausführt. Diese
Sicherung landet im geschützten Datenbereich und zusätzlich im Ordner
`C:\ProgramData\jotti\backups` (die letzten fünf werden vorgehalten). Geht beim
Update etwas schief, stellt **`jotti-restore.cmd`** (Doppelklick) das letzte
dieser Backups wieder her.

> 🔁 **Nur vorwärts, kein Downgrade.** Spielt keine ältere Version über eine
> neuere. Updates verändern die Datenbank; eine alte Version kann mit den neuen
> Daten nicht mehr starten. Der Starter verweigert einen solchen Rückschritt
> selbst.

> ⛔ **Niemals `docker compose down -v` ausführen.** Das `-v` löscht alle
> Docker-Volumes und damit Daten, Installations-Schlüssel und Zertifikat
> unwiderruflich — auch ein Update bringt sie dann nicht zurück. Zum Beenden immer
> `jotti-stop.cmd` verwenden.

## Wenn nach dem Update niemand mehr hineinkommt

Sehr selten passt nach einem Update das in der Datenbank gespeicherte Passwort
nicht mehr zum Installations-Schlüssel; jotti startet dann, aber das Anmelden
schlägt fehl. Eure Daten sind dabei nicht verloren. **`jotti-repair.cmd`**
doppelklicken gleicht beides datenerhaltend wieder an und startet jotti neu;
danach einmal neu anmelden. Mehrfaches Ausführen schadet nicht. Mehr dazu unter
[Fehlersuche](fehlersuche.md#nach-einem-update-klappt-das-anmelden-nicht).
