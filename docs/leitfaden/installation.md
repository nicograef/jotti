---
title: Installation und Start
description: 'Standardweg: einen Windows-Computer im Vereinsheim per Doppelklick zur jotti-Kasse machen und die Handys der Servicekräfte verbinden.'
---

Für fast alle Vereinsfeste ist das der richtige Weg: Ein vorhandener
Windows-Computer im Vereinsheim wird zum Kassenrechner. Die Servicekräfte bedienen jotti auf ihren eigenen Handys im selben WLAN. Kein Server, keine Domain, keine Servermiete; nur die gesetzlich vorgeschriebene TSE kostet laufend (siehe [Was ist jotti?](was-ist-jotti.md)).

## Voraussetzungen

- Ein Windows-Rechner mit Administratorrechten
- Docker Desktop ist installiert: <https://www.docker.com/products/docker-desktop/>
- Internet und WLAN im Vereinsheim

## Start per Doppelklick

Für Windows gibt es einen Doppelklick-Starter, der die `.env` erzeugt, den Stack hochfährt und Docker-Start sowie Firewall-Freigabe selbst erledigt, ganz ohne Kommandozeile.

> ⚠️ **Erststart zuhause mit Internet, nicht auf dem Fest.** Beim ersten Start lädt jotti seine Programmteile herunter und holt das grüne Zertifikat — beides braucht Internet. Macht den Erststart (und spätere [Updates](aktualisieren.md)) in Ruhe vorab, nicht erst am Veranstaltungstag.

1. Das aktuelle Release-ZIP von der [GitHub-Releases-Seite](https://github.com/nicograef/jotti/releases) herunterladen und entpacken (alle Dateien bleiben im selben Ordner).
2. `jotti-start.exe` doppelklicken. Beim ersten Mal dauert der Start einige Minuten (Programmteile werden heruntergeladen).
   - SmartScreen mit „Weitere Informationen" → „Trotzdem ausführen" und UAC mit „Ja" bestätigen.
3. Wenn alles läuft, die Status-Seite `http://localhost:8484` am Kassenrechner im Browser öffnen. Dort stehen die Zugangsadresse und ein QR-Code.

Den vollständigen Windows-Ablauf (SmartScreen, UAC, Beenden) beschreibt auch die `KURZANLEITUNG.md` im ZIP. Für gedruckte Bons folgt weiter unten der Abschnitt „Bondruck einrichten".

> 🔒 **Grünes Schloss als Normalfall.** Für den lokalen Betrieb holt jotti automatisch ein echtes Zertifikat über die Adresse `…lokal.jotti.rocks` (grünes Schloss, keine Warnung). Es wird beim ersten Start ausgestellt und selbst erneuert. Dafür müsst ihr einmalig eine Ausnahme für den DNS-Rebind-Schutz an eurem Router eintragen ([Anleitung je Router](fehlersuche.md#router-hinweise)); bis dahin arbeitet ihr über die Fallback-Adresse ganz normal weiter. Welche Adresse gerade gilt, zeigt samt QR-Code die Status-Seite `http://localhost:8484` am Kassenrechner.
>
> Greift die grüne Adresse nicht, springt ein Fallback `https://<LAN-IP>` mit selbstsigniertem Zertifikat ein (einmalige Browserwarnung pro Gerät, siehe [Fehlersuche](fehlersuche.md)).

## Erster Login

Beim ersten Start legt das Backend automatisch den Admin-Benutzer an und erzeugt einen einmaligen Anmelde-Code aus 6 Ziffern. Der Code steht in der Startkonsole (dem Fenster von `jotti-start.exe`); beim Serverbetrieb per `make prod-init` erscheint er in der Ausgabe des Befehls. Ist die Konsole schon geschlossen, jotti einfach neu starten, dann wird ein neuer Code erzeugt und angezeigt.

Öffnet die jotti-Oberfläche (die Zugangsadresse steht auf der Status-Seite `http://localhost:8484`) und meldet euch **nicht** normal an, sondern wählt „Neues Passwort festlegen":

- **Benutzername:** `admin`
- **Einmalpasswort:** der 6-stellige Code aus der Startkonsole
- **Neues Passwort:** ein eigenes, sicheres Passwort wählen

Nach dem Speichern ist das Einmalpasswort ungültig und der Login mit dem neuen Passwort möglich. Dieser Schritt ist einmalig; alle weiteren Admin-Konten legt ihr danach selbst im Admin-Bereich an.

## Handys der Servicekräfte verbinden

Das Handy ins Vereins-WLAN bringen. Dann den QR-Code von der Status-Seite scannen oder die grüne Adresse eintippen, dann anmelden.

Für die Theke verbindet ihr statt eines Handys ein Tablet oder einen Laptop und stellt es in den Direktverkauf ([Welcher Modus passt?](betriebsarten.md)).

Geht die grüne Adresse nicht, nennt die Status-Seite die Fallback-Adresse (z. B. `https://192.168.1.50`). Beim ersten Zugriff pro Gerät die einmalige Browserwarnung bestätigen, danach anmelden. Lädt die grüne Adresse auf den Handys gar nicht, blockiert vermutlich der Router (siehe [Fehlersuche](fehlersuche.md)).

## Bondruck einrichten (optional)

Für gedruckte Bons braucht ihr einen netzwerkfähigen Bondrucker (ESC/POS, 80 mm,
Ethernet, TCP-Port 9100; eine feste IP-Adresse ist empfohlen). Die Einrichtung hat
zwei Teile:

1. **Druckstationen im Admin-Bereich anlegen.** Unter „Druckstationen" je
   Produktkategorie die „Drucker-IP" und den „Bonmodus" eintragen. Ohne
   konfigurierte Station wird nichts gedruckt.
2. **Drucker-Programm starten.** Auf dem Kassenrechner zusätzlich `jotti-relay.exe`
   doppelklicken. Es läuft ohne Administratorrechte und nimmt seine Zugangsdaten aus
   der `.env`, die `jotti-start.exe` angelegt hat.

## Aktualisieren

Eine neue Version spielt ihr in drei Doppelklicks ein, eure Daten bleiben erhalten.
Den genauen Ablauf beschreibt [Aktualisieren](aktualisieren.md).

## Beenden

`jotti-stop.cmd` doppelklicken (oder in Docker Desktop stoppen). Die Daten bleiben im Docker-Volume erhalten und stehen beim nächsten Start wieder bereit.
