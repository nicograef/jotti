# TSE einrichten: Schritt für Schritt mit fiskaly

> ⚖️ **Kein Rechts- oder Steuerrat.** Dieser Leitfaden erklärt die technische Einrichtung der
> TSE in jotti. Wofür ihr die TSE braucht und welche Pflichten daran hängen, steht im
> [Betreiber-Leitfaden](leitfaden-betreiber.md).

Die TSE (Technische Sicherheitseinrichtung) ist das digitale Siegel-Modul, das jeden
Kassenvorgang fälschungssicher signiert. jotti nutzt die Cloud-TSE von fiskaly. Diese
Anleitung führt euch durch die einmalige Einrichtung, ohne dass ihr technische Werkzeuge
wie Postman oder curl braucht.

---

## 1. Überblick

Die Einrichtung hat zwei Teile:

1. **Im fiskaly-Dashboard** (einmalig, außerhalb von jotti): ein Konto registrieren und
   einen API-Key erstellen. Mehr geht im Dashboard nicht, die eigentliche TSS und den
   Client kann man dort nicht anlegen (siehe Hinweis in Abschnitt 2).
2. **In jotti:** Alles Weitere erledigt der geführte Einrichtungs-Assistent. Ihr gebt nur
   API-Key und API-Secret ein, jotti legt die TSS an, initialisiert sie und registriert
   diese Kasse als Client.

Es gibt zwei Wege, die TSE in jotti zu hinterlegen:

| Weg                              | Für wen                                  | Abschnitt |
| -------------------------------- | ---------------------------------------- | --------- |
| Geführte Einrichtung (Assistent) | Empfohlen, auch ohne Vorkenntnisse       | 3         |
| Manuelle Einrichtung (Experten)  | TSS schon anderswo angelegt, Sonderfälle | 6         |

> 💡 **Übt zuerst in TEST.** fiskaly bietet eine kostenlose Test-Umgebung. Richtet die TSE
> dort einmal komplett ein, bevor ihr auf LIVE umstellt. So lauft ihr den ganzen Ablauf
> einmal durch, ohne Kosten und ohne Risiko.

---

## 2. Voraussetzung: fiskaly-Konto und API-Key

Bevor jotti loslegen kann, braucht ihr Zugangsdaten von fiskaly.

1. Auf [dashboard.fiskaly.com](https://dashboard.fiskaly.com) registrieren und das Konto
   bestätigen.
2. Im Dashboard einen API-Key erstellen. Ihr erhaltet zwei Werte:
   - **API-Key:** eine Art Benutzername.
   - **API-Secret:** das zugehörige Passwort, wird nur einmal angezeigt.
3. Beide Werte sicher notieren. Das Secret könnt ihr später nicht erneut einsehen, nur neu
   erzeugen.

> ℹ️ **Was im Dashboard NICHT geht.** Das fiskaly-Dashboard listet TSS nur auf. Eine neue
> TSS anlegen, sie initialisieren oder einen Client registrieren ist dort nicht möglich,
> das geht ausschließlich über die API. Genau diese Schritte nimmt euch jottis Assistent
> ab. Ihr müsst im Dashboard also wirklich nur Konto und API-Key erstellen.

> 🔒 **API-Key und Secret sind geheim.** Sie gehören nicht in Chats, E-Mails oder
> öffentliche Dokumente. In jotti werden sie verschlüsselt in der Datenbank abgelegt, ihr
> tragt sie nur einmal im Assistenten ein.

---

## 3. Weg 1: Geführte Einrichtung (empfohlen)

### 3.1 Einrichtungsseite öffnen

Im Admin-Bereich: „Finanzamt" öffnen, im Kasten „TSE-Anbindung" auf „Einrichten oder ändern"
klicken. Ihr landet auf der Seite „TSE-Einrichtung" mit dem Kasten „Geführte Einrichtung"
oben und der manuellen Konfiguration darunter.

### 3.2 Zugangsdaten eingeben

API-Key und API-Secret aus Abschnitt 2 eintragen und auf „fiskaly-Konto prüfen" klicken.
jotti meldet sich bei fiskaly an. Bei dieser Prüfung wird nichts angelegt und nichts
gespeichert, es werden nur Daten gelesen.

> Stimmen Key oder Secret nicht, erscheint ein verständlicher Hinweis. Tippt die Werte dann
> erneut ein oder erstellt im Dashboard einen neuen Key.

### 3.3 Befund prüfen: TEST oder LIVE

Nach erfolgreicher Anmeldung zeigt jotti die Umgebung deutlich an:

- **TEST** (graue Markierung): Spielwiese zum Üben. Hier signierte Belege sind steuerlich
  nicht gültig.
- **LIVE** (rote Markierung): die echte Produktivumgebung. Hier angelegte TSS verursachen
  Kosten und lassen sich nicht löschen.

Darunter listet jotti die in eurem Konto gefundenen TSS mit ihrem Zustand auf und weist je
TSS aus, ob bereits ein Client mit dieser Kassen-Seriennummer existiert.

### 3.4 Neu anlegen oder vorhandene TSS übernehmen

Je nach Befund bietet jotti den passenden nächsten Schritt an:

- **Konto leer (oder nur stillgelegte TSS):** jotti bietet „TSE einrichten" an und legt eine
  neue TSS an, initialisiert sie und registriert diese Kasse als Client. In der
  LIVE-Umgebung müsst ihr zur Sicherheit erst das Wort „LIVE" in ein Feld tippen, in TEST
  genügt der Klick.
- **Vorhandene TSS gefunden:** jotti bietet „TSS übernehmen" an, statt eine zweite anzulegen.
  Das schützt vor versehentlicher Doppel-Anlage und nimmt ein abgebrochenes Setup genau
  dort wieder auf, wo es stehen geblieben ist. Ist die TSS bereits personalisiert, fragt
  jotti nach der bei der ersten Einrichtung verwahrten Admin-PIN (siehe Abschnitt 4). Ein
  vorhandener Client mit passender Seriennummer wird übernommen, fehlt er, wird er
  registriert.

> ♻️ **Wiederaufnahme nach Abbruch.** Bricht die Einrichtung ab (Netzfehler, Browser
> geschlossen), startet ihr den Assistenten einfach erneut. jotti erkennt den tatsächlichen
> Zustand bei fiskaly und holt nur die fehlenden Schritte nach. Es entsteht keine zweite
> TSS und kein halbfertiger Zustand.

> 🧪 **Nur in TEST: neue TSE statt Übernahme.** Liegt die Admin-PIN einer vorhandenen
> Test-TSE nicht mehr vor, lässt sie sich nicht übernehmen. In der Test-Umgebung bietet
> jotti dann unter den Übernahme-Optionen die Sekundäraktion „Stattdessen neue TSE anlegen"
> an, mit der ihr ohne PIN eine frische Test-TSE einrichtet. In LIVE gibt es diesen Ausweg
> nicht: eine zweite LIVE-TSS verursacht Kosten, hier helfen nur die verwahrte PIN oder der
> fiskaly-Support.

> ♻️ **Test-TSE räumen sich selbst auf.** fiskaly löscht Test-TSE, die stillgelegt oder
> länger als 14 Tage ungenutzt sind, regelmäßig (mindestens sonntags). Gleichzeitig erlaubt
> die Test-Umgebung höchstens fünf aktive TSE. Habt ihr beim Üben fünf erreicht, meldet
> jotti das verständlich; legt dann keine weitere an, sondern übernehmt eine vorhandene oder
> wartet die automatische Bereinigung ab. In LIVE gilt das nicht: dort bleibt jede TSS
> dauerhaft bestehen und verursacht Kosten.

### 3.5 Admin-PUK und Admin-PIN verwahren

Bei einer Neu-Anlage zeigt jotti danach genau einmal den Admin-PUK und die Admin-PIN an.
Notiert beide sofort und verwahrt sie sicher außerhalb von jotti. jotti speichert sie nicht
und kann sie nicht erneut anzeigen.

Erst wenn ihr das Häkchen „Ich habe Admin-PUK und Admin-PIN sicher verwahrt" setzt, geht es
weiter. Wofür ihr PUK und PIN braucht und was ihr Verlust bedeutet, steht in Abschnitt 4.

> Bei der Übernahme einer bereits personalisierten TSS erscheinen keine neuen Geheimnisse.
> Es gelten eure bereits verwahrten PUK und PIN unverändert weiter.

### 3.6 Abschluss-Verbindungstest

Zum Schluss klickt ihr auf „Verbindung testen & abschließen". jotti speichert jetzt die
vollständige Konfiguration und prüft, ob die Kasse wirklich signieren kann. Der Test zeigt
aufgeschlüsselt:

- Umgebung (TEST/LIVE)
- TSS-Zustand (soll `INITIALIZED` sein)
- Client-Zustand (soll `REGISTERED` sein)
- Seriennummern-Abgleich (Client-Seriennummer muss zur Kassen-Seriennummer passen)

Steht „Verbindung bestätigt", ist die TSE einsatzbereit. Bei Auffälligkeiten nennt jotti das
betroffene Feld, dann prüft ihr die Einrichtung (oder nutzt den Verbindungstest in der
manuellen Konfiguration darunter).

---

## 4. Admin-PUK und Admin-PIN: sicher verwahren

Beim Anlegen einer TSS vergibt fiskaly einen Admin-PUK, mit dem jotti eine zufällige
Admin-PIN setzt. Beide gehören zur TSS, nicht zu jotti.

**Wofür ihr sie braucht:** für spätere Verwaltungsaufgaben direkt an der TSS, zum Beispiel
wenn ihr die TSS später auf einer neuen Installation übernehmt (Abschnitt 3.4) oder mit dem
fiskaly-Support arbeitet. Im normalen Kassenbetrieb braucht ihr sie nicht, jotti signiert
über API-Key und Secret.

> ⚠️ **Verlust hat Folgen.** jotti speichert PUK und PIN bewusst nicht. Gehen sie verloren,
> könnt ihr eine bereits personalisierte TSS nicht mehr übernehmen (etwa nach einem
> Datenbankverlust oder Serverwechsel). Dann bleiben nur der fiskaly-Support oder, in der
> LIVE-Umgebung kostenpflichtig, eine bewusste Neu-Anlage. Verwahrt PUK und PIN deshalb so
> sorgfältig wie die Zugangsdaten zum fiskaly-Konto.

> ⚠️ **Fünf Fehlversuche sperren die PIN.** Gebt ihr die Admin-PIN fünfmal falsch ein,
> sperrt fiskaly sie. Sie lässt sich dann nur noch mit dem Admin-PUK zurücksetzen. Ratet
> deshalb nicht wiederholt, sondern schaut die verwahrte PIN in euren Unterlagen nach.

**So verwahrt ihr richtig:**

- An einem sicheren, dauerhaften Ort, getrennt vom Server (Passwort-Manager des Vorstands,
  versiegelter Ausdruck im Vereinssafe).
- Nicht nur auf dem Gerät, das ihr für die Einrichtung benutzt habt.
- So, dass die Nachfolge im Vorstand sie wiederfindet.

---

## 5. Von TEST zu LIVE wechseln (inkl. Kosten)

Habt ihr in TEST geübt, richtet ihr für den Echtbetrieb eine LIVE-TSS ein:

1. Im fiskaly-Dashboard einen API-Key für die LIVE-Umgebung erstellen (oder einen
   bestehenden LIVE-Key verwenden). TEST- und LIVE-Schlüssel sind getrennt.
2. In jottis Assistent diese LIVE-Zugangsdaten eingeben und prüfen. jotti zeigt jetzt die
   rote LIVE-Markierung.
3. Vor der Anlage das Wort „LIVE" eintippen (Schutz vor versehentlicher Anlage) und die
   Einrichtung durchlaufen wie in Abschnitt 3.

> ⚠️ **LIVE kostet und ist endgültig.** Eine LIVE-TSS verursacht laufende Kosten und kann
> nicht gelöscht, nur stillgelegt werden. Legt sie erst an, wenn ihr in den Echtbetrieb
> geht.

**Kosten (Stand Mitte 2026):** fiskaly veröffentlicht für SIGN DE keine frei buchbare
Preisliste und vertreibt die Cloud-TSE überwiegend gebündelt über Kassenanbieter (dort
üblich rund 8 bis 10 € pro Monat). Als selbst-gehosteter Betreiber braucht ihr einen
direkten fiskaly-Zugang mit eigenem API-Key. Eigenständige Cloud-TSE-Lizenzen lagen Mitte
2026 grob bei 120 bis 190 € pro Jahr je TSS, teils mit einmaliger Einrichtungsgebühr; eine
Gebühr pro Vorgang ist unüblich. Eine TSS genügt für eine jotti-Instanz (eine Kasse).

> 💶 **Holt ein aktuelles Angebot direkt bei fiskaly ein.** Die Preise ändern sich und
> hängen vom Vertrag ab. Verlasst euch für die Budgetplanung nicht auf die Richtwerte oben,
> sondern auf das schriftliche Angebot.

Quellen: [kassensystemevergleich.de/tse-kosten](https://www.kassensystemevergleich.de/tse-kosten/), [fiskaly.com/signde](https://www.fiskaly.com/signde)

---

## 6. Weg 2: Manuelle Einrichtung (Experten)

Habt ihr eine TSS samt Client bereits außerhalb von jotti angelegt (oder verwaltet sie ein
Dienstleister), tragt ihr die Zugangsdaten direkt ein. Auf der Seite „TSE-Einrichtung" im
Kasten „Manuelle Konfiguration (Experten)":

1. API-Key, API-Secret, TSS-ID und Client-ID eintragen (alle vier sind Pflicht).
2. Auf „Speichern" klicken.
3. Auf „Verbindung testen" klicken und das aufgeschlüsselte Ergebnis prüfen (wie in
   Abschnitt 3.6).

> ⚠️ **Seriennummer muss passen.** Der Client muss bei fiskaly mit der Kassen-Seriennummer
> aus jottis Kassenidentität als `serial_number` registriert sein. Stimmen Client- und
> Kassen-Seriennummer nicht überein, meldet der Verbindungstest einen Fehler, und Beleg und
> QR-Code würden nicht zusammenpassen. Den geführten Weg (Abschnitt 3) zu nutzen ist
> einfacher, weil jotti die Seriennummer automatisch korrekt setzt.

Mit „Alle Felder leeren" entfernt ihr die Konfiguration wieder, etwa zur Schlüsselrotation
oder beim Wechsel von TEST auf LIVE.

---

## 7. Wenn die Einrichtung hängt

| Meldung                                        | Bedeutung und Ausweg                                                                                                                                                                                                         |
| ---------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Zugangsdaten ungültig                          | API-Key oder Secret stimmt nicht. Werte erneut eingeben oder im Dashboard neu erstellen.                                                                                                                                     |
| Umgebung hat sich geändert                     | Zwischen Prüfung und Einrichtung wurden die Zugangsdaten gewechselt. Konto erneut prüfen.                                                                                                                                    |
| Admin-PIN nicht akzeptiert                     | fiskaly lehnt die eingegebene PIN ab. Verwahrte PIN prüfen, dabei nicht wiederholt raten (fünf Fehlversuche sperren die PIN). Sonst fiskaly-Support oder, in TEST, eine neue TSE anlegen.                                     |
| TSS in nicht übernehmbarem Zustand             | Die vorhandene TSS lässt sich nicht automatisch übernehmen. fiskaly-Support kontaktieren oder manuelle Einrichtung.                                                                                                          |
| Test-Limit erreicht (fünf TSE)                 | Nur in TEST: fiskaly erlaubt höchstens fünf aktive Test-TSE. Eine vorhandene übernehmen oder die automatische Bereinigung (bei Inaktivität) abwarten.                                                                        |
| Serverfehler direkt nach dem Anlegen           | Selten: die TSS wurde angelegt, das Speichern in jotti schlug fehl. Assistent erneut starten, jotti bietet die Übernahme an (Abschnitt 3.4). Voraussetzung ist die verwahrte Admin-PIN, sonst hilft nur der fiskaly-Support. |
| Verbindung mit Auffälligkeiten (Abschlusstest) | Das genannte Feld (TSS, Client oder Seriennummer) prüfen, ggf. Einrichtung wiederholen.                                                                                                                                      |

Bei einer unbekannten Admin-PIN sitzt ihr nicht in einer Sackgasse: Entweder die verwahrte
PIN noch einmal genau prüfen, den fiskaly-Support einschalten, oder mit anderen
Zugangsdaten bewusst eine neue TSS anlegen (in LIVE kostenpflichtig).

---

## 8. Checkliste

- [ ] fiskaly-Konto registriert
- [ ] API-Key und API-Secret erstellt und sicher notiert
- [ ] In TEST einmal komplett durchgespielt
- [ ] LIVE-Zugangsdaten erstellt
- [ ] LIVE-TSS über den Assistenten eingerichtet
- [ ] Admin-PUK und Admin-PIN sicher und dauerhaft verwahrt
- [ ] Abschluss-Verbindungstest steht auf „Verbindung bestätigt"
- [ ] Danach: Kasse beim Finanzamt anmelden (→ [Betreiber-Leitfaden](leitfaden-betreiber.md), Schritt 2)
