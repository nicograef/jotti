---
title: TSE einrichten (fiskaly)
description: 'Die Cloud-TSE von fiskaly Schritt für Schritt anbinden: Konto und API-Key, geführter Assistent, Admin-PUK und Admin-PIN verwahren, von TEST zu LIVE wechseln.'
---

Die TSE (Technische Sicherheitseinrichtung) signiert jeden Kassenvorgang fälschungssicher. Das Gesetz schreibt sie zwingend vor. jotti nutzt die Cloud-TSE von fiskaly: Ihr bucht sie als Online-Dienst und gebt jotti die Zugangsschlüssel. Den Rest erledigt ein Assistent in jotti Schritt für Schritt.

> 💡 **Übt zuerst in TEST.** fiskaly bietet eine kostenlose Test-Umgebung. Richtet
> die TSE dort einmal komplett ein, bevor ihr auf LIVE umstellt. So lauft ihr den
> ganzen Ablauf einmal durch, ohne Kosten und ohne Risiko.

## Schritt 1: fiskaly-Konto und API-Key

1. Auf [dashboard.fiskaly.com](https://dashboard.fiskaly.com) registrieren und das
   Konto bestätigen.
2. Im Dashboard einen API-Key erstellen. Ihr erhaltet zwei Werte: den **API-Key** (eine Art Benutzername) und das **API-Secret** (das Passwort, wird nur einmal angezeigt).
3. Beide Werte sicher notieren. Das Secret könnt ihr später nicht erneut einsehen, nur neu erzeugen.

Mehr ist im Dashboard nicht nötig. Die TSS anlegen, initialisieren und den Client registrieren übernimmt jottis Assistent.

> 🔒 **API-Key und Secret sind geheim.** Sie gehören nicht in Chats, E-Mails oder
> öffentliche Dokumente. jotti speichert sie verschlüsselt in der Datenbank, ihr
> tragt sie nur einmal im Assistenten ein.

## Schritt 2: Geführter Assistent in jotti

1. Im Admin-Bereich „Finanzamt" öffnen, im Kasten „TSE-Anbindung" auf „Einrichten oder ändern" klicken.
2. API-Key und API-Secret eingeben und auf „fiskaly-Konto prüfen" klicken. Die Prüfung legt nichts an, sie liest nur. jotti zeigt danach die Umgebung an (TEST grau, LIVE rot) und listet die gefundenen TSS auf.
3. Ist das Konto leer, bietet jotti „TSE einrichten" an: Es legt eine neue TSS an, initialisiert sie und registriert diese Kasse als Client. In LIVE müsst ihr erst das Wort „LIVE" eintippen.
4. jotti zeigt danach genau einmal den **Admin-PUK** und die **Admin-PIN** an. Notiert beide sofort und verwahrt sie außerhalb von jotti (siehe unten). Erst nach dem Häkchen „Ich habe Admin-PUK und Admin-PIN sicher verwahrt" geht es weiter.
5. „Verbindung testen & abschließen" klicken. Steht „Verbindung bestätigt", ist die TSE einsatzbereit.

## Admin-PUK und Admin-PIN verwahren

fiskaly vergibt beim Anlegen einer TSS einen Admin-PUK, mit dem jotti eine
zufällige Admin-PIN setzt. Beide gehören zur TSS, nicht zu jotti. Im normalen
Kassenbetrieb braucht ihr sie nicht (jotti signiert über API-Key und Secret), wohl
aber für spätere Verwaltungsaufgaben, etwa wenn ihr die TSS auf einer neuen
Installation übernehmt.

So verwahrt ihr richtig:

- An einem sicheren, dauerhaften Ort, getrennt vom Server (Passwort-Manager des
  Vorstands, versiegelter Ausdruck im Vereinssafe).
- Nicht nur auf dem Gerät, das ihr für die Einrichtung benutzt habt.
- So, dass die Nachfolge im Vorstand sie wiederfindet.

> ⚠️ **Verlust hat Folgen.** Ist nur die Admin-PIN verloren oder gesperrt, der Admin-PUK aber verwahrt, setzt jotti die PIN über den PUK zurück (siehe [TSE-Sonderfälle](tse-sonderfaelle.md)). Gehen PUK und PIN beide verloren, könnt ihr eine bereits personalisierte TSS nicht mehr übernehmen. Der Admin-PUK ist deshalb euer wichtigstes Geheimnis.

## Von TEST zu LIVE wechseln (inkl. Kosten)

Habt ihr in TEST geübt, richtet ihr für den Echtbetrieb eine LIVE-TSS ein:

1. Im fiskaly-Dashboard einen API-Key für die LIVE-Umgebung erstellen. TEST- und
   LIVE-Schlüssel sind getrennt.
2. Im Assistenten diese LIVE-Zugangsdaten eingeben und prüfen. jotti zeigt jetzt
   die rote LIVE-Markierung.
3. Vor der Anlage das Wort „LIVE" eintippen und die Einrichtung wie in TEST
   durchlaufen.

> ⚠️ **LIVE kostet und ist endgültig.** Eine LIVE-TSS verursacht laufende Kosten
> und kann nicht gelöscht, nur stillgelegt werden. Legt sie erst an, wenn ihr in
> den Echtbetrieb geht.

**Kosten:** Eine LIVE-TSS kostet ca. 8 € pro Monat; eine Gebühr pro Vorgang ist unüblich. Eine TSS genügt für eine jotti-Instanz. fiskaly veröffentlicht für SIGN DE keine feste Preisliste, holt für die Budgetplanung also ein aktuelles Angebot direkt bei fiskaly ein.
