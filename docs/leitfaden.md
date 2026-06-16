# jotti betreiben: der Leitfaden für Vereine

jotti ist eine Kasse im Sinne des Gesetzes (ein „elektronisches
Aufzeichnungssystem"). Damit gelten dieselben Regeln wie für jede Registrierkasse, auch für gemeinnützige Vereine und auch bei kurzen Festen.

Drei Dinge müsst ihr als Verein selbst erledigen:

1. Bei [fiskaly](https://www.fiskaly.com/de/signde) ein Account für die TSE anlegen und die Schlüssel in jotti eintragen (TSE = manipulationssicheres Signaturmodul für Kassensysteme).
2. Eure Kasse beim Finanzamt anmelden (online über ELSTER). Dafür braucht ihr die Seriennummer, die jotti euch im Admin-Bereich anzeigt.
3. Alle Kassendaten 10 Jahre aufbewahren (regelmäßige Backups).

**Was kostet uns das?** jotti selbst ist für euch kostenlos. Kosten
entstehen nur für die TSE von fiskaly (circa 8€ pro Monat), und ggf. für einen Server (circa 5€ pro Monat).

## Standardweg: Computer im Vereinsheim

Für fast alle Vereinsfeste ist das der richtige Weg: Ein vorhandener
Windows-Computer im Vereinsheim wird zum Kassenrechner. Die Servicekräfte bedienen jotti auf ihren eigenen Handys im selben WLAN. Kein Server, keine Domain, keine laufenden Kosten.

### Voraussetzungen

- Ein Windows-Rechner mit Administratorrechten
- Docker Desktop ist installiert: <https://www.docker.com/products/docker-desktop/>
- Internet und WLAN im Vereinsheim

### Start per Doppelklick

Für Windows gibt es einen Doppelklick-Starter, der die `.env` erzeugt, den Stack hochfährt und Docker-Start sowie Firewall-Freigabe selbst erledigt, ganz ohne Kommandozeile.

1. Das aktuelle Release-ZIP von der [GitHub-Releases-Seite](https://github.com/nicograef/jotti/releases) herunterladen und entpacken (alle Dateien bleiben im selben Ordner).
2. `jotti-start.exe` doppelklicken. Beim ersten Mal dauert der Start einige Minuten (Programmteile werden heruntergeladen).
   - SmartScreen mit „Weitere Informationen" → „Trotzdem ausführen" und UAC mit „Ja" bestätigen.
3. Wenn alles läuft, die Status-Seite `http://localhost:8484` am Kassenrechner im Browser öffnen. Dort stehen die Zugangsadresse und ein QR-Code.

Für den Bondruck zusätzlich `jotti-relay.exe` doppelklicken. Den vollständigen Ablauf (SmartScreen, UAC, Bondruck, Beenden, Aktualisieren) beschreibt die `KURZANLEITUNG.md` im ZIP.

> 🔒 **Grünes Schloss als Normalfall.** Für den lokalen Betrieb holt jotti automatisch ein echtes Zertifikat über die Adresse `…lokal.jotti.rocks` (grünes Schloss, keine Warnung). Es wird beim ersten Start ausgestellt und selbst erneuert. Dafür müsst ihr einmalig eine Ausnahme für den DNS-Rebind-Schutz bei eurem Router konfigurieren. Welche Adresse gerade gilt, zeigt samt QR-Code die Status-Seite `http://localhost:8484` am Kassenrechner.
>
> Greift die grüne Adresse nicht, springt ein Fallback `https://<LAN-IP>` mit selbstsigniertem Zertifikat ein (einmalige Browserwarnung pro Gerät, siehe [Fehlersuche](#fehlersuche)).

### Handys der Servicekräfte verbinden

Das Handy ins Vereins-WLAN bringen. Dann den QR-Code von der Status-Seite scannen oder die grüne Adresse eintippen, dann anmelden.

Geht die grüne Adresse nicht, nennt die Status-Seite die Fallback-Adresse (z. B. `https://192.168.1.50`). Beim ersten Zugriff pro Gerät die einmalige Browserwarnung bestätigen, danach anmelden. Lädt die grüne Adresse auf den Handys gar nicht, blockiert vermutlich der Router (siehe [Fehlersuche](#fehlersuche)).

### Beenden

`jotti-stop.cmd` doppelklicken (oder in Docker Desktop stoppen). Die Daten bleiben im Docker-Volume erhalten und stehen beim nächsten Start wieder bereit.

## TSE einrichten (Cloud-TSE von fiskaly)

Die TSE (Technische Sicherheitseinrichtung) signiert jeden Kassenvorgang fälschungssicher. Das Gesetz schreibt sie zwingend vor. jotti nutzt die Cloud-TSE von fiskaly: Ihr bucht sie als Online-Dienst und gebt jotti die Zugangsschlüssel. Den Rest erledigt ein Assistent in jotti Schritt für Schritt.

> 💡 **Übt zuerst in TEST.** fiskaly bietet eine kostenlose Test-Umgebung. Richtet
> die TSE dort einmal komplett ein, bevor ihr auf LIVE umstellt. So lauft ihr den
> ganzen Ablauf einmal durch, ohne Kosten und ohne Risiko.

### Schritt 1: fiskaly-Konto und API-Key

1. Auf [dashboard.fiskaly.com](https://dashboard.fiskaly.com) registrieren und das
   Konto bestätigen.
2. Im Dashboard einen API-Key erstellen. Ihr erhaltet zwei Werte: den **API-Key** (eine Art Benutzername) und das **API-Secret** (das Passwort, wird nur einmal angezeigt).
3. Beide Werte sicher notieren. Das Secret könnt ihr später nicht erneut einsehen, nur neu erzeugen.

Mehr ist im Dashboard nicht nötig. Die TSS anlegen, initialisieren und den Client registrieren übernimmt jottis Assistent.

> 🔒 **API-Key und Secret sind geheim.** Sie gehören nicht in Chats, E-Mails oder
> öffentliche Dokumente. jotti speichert sie verschlüsselt in der Datenbank, ihr
> tragt sie nur einmal im Assistenten ein.

### Schritt 2: Geführter Assistent in jotti

1. Im Admin-Bereich „Finanzamt" öffnen, im Kasten „TSE-Anbindung" auf „Einrichten oder ändern" klicken.
2. API-Key und API-Secret eingeben und auf „fiskaly-Konto prüfen" klicken. Die Prüfung legt nichts an, sie liest nur. jotti zeigt danach die Umgebung an (TEST grau, LIVE rot) und listet die gefundenen TSS auf.
3. Ist das Konto leer, bietet jotti „TSE einrichten" an: Es legt eine neue TSS an, initialisiert sie und registriert diese Kasse als Client. In LIVE müsst ihr erst das Wort „LIVE" eintippen.
4. jotti zeigt danach genau einmal den **Admin-PUK** und die **Admin-PIN** an. Notiert beide sofort und verwahrt sie außerhalb von jotti (siehe unten). Erst nach dem Häkchen „Ich habe Admin-PUK und Admin-PIN sicher verwahrt" geht es weiter.
5. „Verbindung testen & abschließen" klicken. Steht „Verbindung bestätigt", ist die TSE einsatzbereit.

### Admin-PUK und Admin-PIN verwahren

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

> ⚠️ **Verlust hat Folgen.** Ist nur die Admin-PIN verloren oder gesperrt, der Admin-PUK aber verwahrt, setzt jotti die PIN über den PUK zurück (siehe [TSE-Sonderfälle](#tse-sonderfälle)). Gehen PUK und PIN beide verloren, könnt ihr eine bereits personalisierte TSS nicht mehr übernehmen. Der Admin-PUK ist deshalb euer wichtigstes Geheimnis.

### Von TEST zu LIVE wechseln (inkl. Kosten)

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

**Kosten:** Eine LIVE-TSS kostet circa 8€ pro Monat; eine Gebühr pro Vorgang ist unüblich. Eine TSS genügt für eine jotti-Instanz. fiskaly veröffentlicht für SIGN DE keine feste Preisliste, holt für die Budgetplanung also ein aktuelles Angebot direkt bei fiskaly ein.

## Pflichten erfüllen

Die rechtlichen Grundlagen im Detail stehen in [docs/compliance.md](compliance.md)
(KassenSichV, GoBD, DSFinV-K, ELSTER) und die Steuersätze in
[docs/steuerrecht.md](steuerrecht.md). Hier die drei Pflichten in der Praxis.

### Kasse beim Finanzamt anmelden

Seit dem 1. Januar 2025 muss jede elektronische Kasse dem Finanzamt online gemeldet werden. Das ist eine eigene Pflicht, unabhängig von der TSE.

Was ihr braucht:

- Die Seriennummer eurer jotti-Kasse. Da es keine Hardware mit aufgedruckter Nummer
  gibt, erzeugt jotti beim ersten Start automatisch eine eindeutige Nummer (eine
  „UUID") und zeigt sie im Admin-Bereich an.
- Die Zertifizierungs-ID und Seriennummer eurer TSE (bekommt ihr von fiskaly).
- Steuernummer des Vereins, Anschrift der Betriebsstätte,
  Anschaffungs-/Inbetriebnahmedatum, Softwarename „jotti".

So geht ihr vor: Im Mein-ELSTER-Portal ([elster.de](https://www.elster.de))
anmelden, das Formular „Mitteilung über elektronische Aufzeichnungssysteme"
ausfüllen (die Daten findet ihr gebündelt im jotti-Admin-Bereich), absenden und die
Bestätigung aufbewahren.

### Belege und Steuersätze

- **Belegausgabe:** Für jeden Kassiervorgang muss ein Beleg erstellbar sein. Beim
  Vereinsfest greift meist die Befreiung von der Aushändigung („Verkauf an eine
  Vielzahl nicht bekannter Personen", § 146a Abs. 2 Satz 2 AO). Diese Befreiung
  müsst ihr aber beim Finanzamt schriftlich beantragen, sie gilt nicht automatisch.
  Verlangt ein Gast einen Bon, müsst ihr ihn aushändigen.
- **Steuersätze:** Ordnet jedem Produkt im Admin-Bereich den richtigen Steuersatz
  zu (19 %, 7 % oder 0 % / steuerbefreit). Der Steuersatz erscheint auf dem Beleg
  und im Datenexport. Welcher Satz für euch gilt, klärt euer Steuerberater.

### Daten 10 Jahre aufbewahren

- Alle Kassendaten (das Kassenjournal und spätere DSFinV-K-Exporte) müssen 10 Jahre
  vollständig, lesbar und unveränderbar aufbewahrt werden.
- Macht regelmäßige (täglich empfohlene) Backups eurer Datenbank und bewahrt sie
  sicher auf. Das schützt zugleich eure Kassen-Seriennummer.
- Sorgt dafür, dass nur berechtigte Personen Zugriff auf den Server und die Daten
  haben.

> **Gilt das auch für unseren gemeinnützigen Verein?** Ja. Sobald ihr bei einem
> Fest Speisen oder Getränke gegen Geld verkauft, betreibt ihr einen
> „wirtschaftlichen Geschäftsbetrieb". Die Kassenpflichten gelten unabhängig von
> der Gemeinnützigkeit. Ob auf eure Umsätze Steuern anfallen, ist eine andere Frage
> (das klärt euer Steuerberater).

## Checkliste

**Einmalig, vor dem ersten Einsatz:**

- [ ] Nutzungsvereinbarung mit dem Autor abgeschlossen (siehe
      [Lizenzmodell](lizenzmodell.md))
- [ ] fiskaly-Konto registriert, API-Key und API-Secret erstellt und sicher notiert
- [ ] TSE in TEST einmal komplett durchgespielt
- [ ] LIVE-TSS über den Assistenten eingerichtet
- [ ] Admin-PUK und Admin-PIN sicher und dauerhaft verwahrt
- [ ] Abschluss-Verbindungstest steht auf „Verbindung bestätigt"
- [ ] Betreiber-Stammdaten (Vereinsname, Adresse, Steuernummer) im Admin-Bereich
      gepflegt
- [ ] Produkte mit korrekten Steuersätzen angelegt
- [ ] Kasse über ELSTER beim Finanzamt angemeldet (Seriennummer aus dem
      Admin-Bereich)
- [ ] Ggf. Antrag auf Befreiung von der Belegausgabe gestellt
- [ ] Backup-Routine eingerichtet

**Laufend / regelmäßig:**

- [ ] Tägliche Datenbank-Backups laufen
- [ ] Nach jedem Veranstaltungstag: Tagesabschluss (Z-Bon) erstellt
- [ ] Daten werden 10 Jahre archiviert
- [ ] Bei Stilllegung der Kasse: Abmeldung beim Finanzamt innerhalb 1 Monat

## Experten-Weg: eigener Server (optional)

> Dieser Abschnitt ist nur für den Betrieb über das Internet gedacht und setzt
> Grundkenntnisse mit Linux und Kommandozeile voraus. Wer alle Helfer beim Fest vor
> Ort im selben WLAN hat, bleibt beim [Standardweg](#standardweg-computer-im-vereinsheim).

Wer jotti auch außerhalb des lokalen WLANs erreichen will (über das Internet, an
mehreren Standorten oder ohne einen Rechner vor Ort), betreibt es auf einem eigenen
kleinen Server, einem VPS. Dann erreichen alle Helfer jotti über eine
Internet-Adresse (Domain) mit Verschlüsselung (HTTPS).

jotti ist genügsam: Schon der kleinste VPS (1 vCPU, 2 GB RAM, 20 GB SSD, Linux)
reicht für ein durchschnittliches Vereinsfest. Typisches Angebot: netcup VPS 200
oder vergleichbar (circa 5€/Monat). Zusätzlich braucht ihr eine Domain, die per
DNS auf den Server zeigt, sowie ein TLS-Zertifikat. Das Zertifikat holt jotti
automatisch: Die Produktions-Konfiguration (`docker-compose.prod.yml`) bringt einen
Caddy-Reverse-Proxy mit, der es beim ersten Start selbst bei Let's Encrypt
anfordert und danach erneuert.

> ⚠️ Ohne HTTPS dürft ihr jotti nicht über das offene Internet betreiben:
> Anmeldedaten und Bestellungen würden sonst unverschlüsselt übertragen.

### Ersteinrichtung

1. **Docker installieren.** Auf dem Server Docker Engine samt Compose-Plugin
   einrichten (<https://docs.docker.com/engine/install/>).
2. **Projektdateien holen.** Das aktuelle Release als ZIP entpacken oder das
   Repository klonen, dann in den Projektordner wechseln.
3. **`.env` anlegen** mit `make init` (erzeugt sichere Zufallswerte für die
   Geheimnisse).
4. **Domain, E-Mail und Version eintragen** in der `.env`:

   ```bash
   JOTTI_DOMAIN=kasse-musterverein.de
   LETSENCRYPT_EMAIL=vorstand@musterverein.de
   JOTTI_VERSION=v0.2.0
   ```

5. **DNS auf den Server zeigen lassen.** Beim Domain-Anbieter einen A-Record (und
   bei IPv6 einen AAAA-Record) auf die öffentliche IP des Servers setzen. Erst wenn
   die Domain auf den Server zeigt, kann Let's Encrypt ein Zertifikat ausstellen.
6. **Stack starten** mit `make prod-init`. Das Skript prüft Docker und die
   DNS-Auflösung, zieht die gepinnten Images, startet den Stack und wartet, bis
   Backend und HTTPS gesund antworten.

Danach ist jotti unter `https://<eure-domain>` erreichbar; HTTP leitet automatisch
auf HTTPS um.

### Aktualisieren

Tragt unter `JOTTI_VERSION` die gewünschte Release-Version ein und führt
`make prod-update` aus. Das Skript sichert die Datenbank automatisch, bevor es die
neuen Images zieht und die Migrationen ausführt, und prüft danach die Gesundheit.
Bleibt der Stack ungesund, bricht es ab und zeigt, wie ihr mit dem eben erstellten
Backup zurückkehrt.

> 🔁 Nur vorwärts, kein Downgrade. Updates verändern die Datenbank und lassen sich
> nicht zurücknehmen; eine ältere Version kann mit den neuen Daten nicht mehr
> starten. Wollt ihr zurück, spielt ein Backup ein (`make prod-restore`).

### Backups

jotti speichert alle Daten in einem Docker-Volume. Macht regelmäßige Backups, schon
wegen der gesetzlichen 10-Jahre-Aufbewahrung.

- **Sichern:** `make prod-backup` zieht einen komprimierten `pg_dump` in den Ordner
  `BACKUP_DIR` (Standard `./backups`) und behält die neuesten `BACKUP_KEEP` Stück.
- **Wiederherstellen:** `make prod-restore` listet die Backups, fragt eine
  Bestätigung ab und spielt das gewählte zurück. Das überschreibt die aktuelle
  Datenbank.
- **Täglich automatisch:** Für einen täglichen Dump liegen Vorlagen im Repository
  (systemd-Timer unter `packaging/systemd/` oder cron unter `packaging/cron/`); die
  Installationsschritte stehen als Kommentar in den Dateien.

> 💾 Kopiert die Backups regelmäßig vom Server weg. Ein Backup, das nur auf
> demselben Server liegt, hilft bei dessen Ausfall nicht.

## TSE-Sonderfälle

Die folgenden Fälle braucht ihr nur, wenn etwas vom Normalfall abweicht.

**Vorhandene TSS übernehmen.** Findet jotti im Konto bereits eine TSS, bietet es
„TSS übernehmen" an, statt eine zweite anzulegen. Das schützt vor versehentlicher
Doppel-Anlage und nimmt ein abgebrochenes Setup dort wieder auf, wo es stehen
geblieben ist. Ist die TSS bereits personalisiert, fragt jotti nach der verwahrten
Admin-PIN.

**Wiederaufnahme nach Abbruch.** Bricht die Einrichtung ab (Netzfehler, Browser
geschlossen), startet ihr den Assistenten einfach erneut. jotti erkennt den
tatsächlichen Zustand bei fiskaly und holt nur die fehlenden Schritte nach. Es
entsteht keine zweite TSS und kein halbfertiger Zustand.

**PIN per PUK zurücksetzen.** Verlangt jotti die Admin-PIN und habt ihr sie nicht
(oder hat fiskaly sie nach fünf Fehlversuchen gesperrt), bietet der Assistent „Ich
habe den Admin-PUK" an. Gebt dort den verwahrten Admin-PUK ein und klickt „PIN
zurücksetzen und übernehmen": jotti setzt eine neue Admin-PIN und schließt die
Übernahme ab, ohne neue, kostenpflichtige TSS. Das funktioniert in TEST und LIVE.
Sind PUK und PIN beide verloren, hilft nur der fiskaly-Support.

**Test-Limit und Selbstreinigung.** Die Test-Umgebung erlaubt höchstens fünf aktive
TSE. Habt ihr beim Üben fünf erreicht, übernehmt eine vorhandene oder wartet die
automatische Bereinigung ab (fiskaly löscht stillgelegte oder länger als 14 Tage
ungenutzte Test-TSE regelmäßig). Liegt die PIN einer vorhandenen Test-TSE nicht mehr
vor, bietet jotti nur in TEST die Sekundäraktion „Stattdessen neue TSE anlegen" an.
In LIVE gibt es diesen Ausweg nicht; dort helfen der PUK-Reset, die verwahrte PIN
oder der fiskaly-Support.

**Manuelle Konfiguration (Experten).** Habt ihr eine TSS samt Client bereits
außerhalb von jotti angelegt, tragt ihr auf der Seite „TSE-Einrichtung" im Kasten
„Manuelle Konfiguration" API-Key, API-Secret, TSS-ID und Client-ID direkt ein (alle
vier sind Pflicht), speichert und klickt „Verbindung testen". Der Client muss bei
fiskaly mit der Kassen-Seriennummer aus jottis Kassenidentität registriert sein,
sonst meldet der Test einen Fehler. Mit „Alle Felder leeren" entfernt ihr die
Konfiguration wieder, etwa zur Schlüsselrotation.

## Fehlersuche

> Dieser Abschnitt hilft, wenn die grüne Adresse der lokalen jotti-Kasse auf den
> Handys nicht lädt, obwohl die Kasse läuft. Die Status-Seite
> (`http://localhost:8484`) verlinkt direkt hierher, wenn sie das Problem erkennt.

### Grüne Adresse lädt nicht (DNS-Rebind-Schutz)

jotti holt für den lokalen Betrieb ein echtes Let's-Encrypt-Zertifikat auf einen
Namen, der eure private LAN-IP enthält (z. B. `192-168-1-50.<id>.lokal.jotti.rocks`
→ `192.168.1.50`). Das ist gewollt und sicher, aber viele Router haben einen
DNS-Rebind-Schutz, der genau diese Kombination („öffentlicher Name zeigt auf eine
private IP") als möglichen Angriff einstuft und blockiert. Die Antwort kommt dann im
WLAN leer an, und das Handy kann die grüne Adresse nicht öffnen.

DNS-Rebind-Schutz ist die wahrscheinliche Ursache, wenn die Fallback-Adresse
`https://<LAN-IP>` funktioniert, die grüne Adresse aber nicht, oder wenn es „auf
Handy A geht, auf Handy B aber nicht". Bis die Ausnahme eingetragen ist, könnt ihr
jederzeit mit der Fallback-Adresse weiterarbeiten. Der Verkauf muss nicht warten.

Die im Standardweg genannte Router-Ausnahme behebt das: `lokal.jotti.rocks` einmalig
von der Prüfung ausnehmen. Danach funktioniert die grüne Adresse im gesamten
Vereins-WLAN. Die Ausnahme erlaubt private IPs nur für diese eine Domain; der
Rebind-Schutz für alle anderen Domains bleibt aktiv.

### Router-Hinweise

**Fritz!Box** (häufigster Router im Vereinsumfeld): `http://fritz.box` öffnen und
anmelden → Heimnetz → Netzwerk → Netzwerkeinstellungen → „Weitere Einstellungen" →
Abschnitt „DNS-Rebind-Schutz". Im Feld „Diese Domain(s) ausnehmen" genau
`lokal.jotti.rocks` eintragen und mit „Übernehmen" speichern. Falls die grüne
Adresse danach weiterhin blockiert wird, zusätzlich den vollständigen Hostnamen aus
der Status-Seite eintragen.

Andere Router, gleiches Prinzip (`lokal.jotti.rocks` ausnehmen), andere
Bezeichnungen:

- **Pi-hole / dnsmasq:** in der Konfigurationsdatei `rebind-domain-ok=/lokal.jotti.rocks/`
- **OpenWrt (LuCI):** Network → DHCP and DNS → Rebind protection → Domain whitelist
  → `lokal.jotti.rocks`
- **Speedport (Telekom):** Netzwerk → DNS-Rebind-Schutz → Domain zur Liste
  hinzufügen → `lokal.jotti.rocks`

Nach jeder Änderung den DNS-Dienst des Routers neu laden bzw. neu starten. Hat euer
Router keinen Rebind-Schutz, blockiert er auch nichts, dann liegt die Ursache
woanders (siehe unten).

### Weitere Stolpersteine

Die grüne Adresse funktioniert nur, wenn das Handy die private LAN-IP des
Kassenrechners erreichen kann:

- **Vereins-WLAN, nicht Mobilfunk.** Der Name löst zwar auch im Mobilfunknetz auf,
  aber die private IP ist von außerhalb des WLAN nicht erreichbar. Das Handy muss im
  selben WLAN wie der Kassenrechner sein.
- **Kein Gastnetz.** Gastnetze isolieren ihre Geräte und blockieren sowohl die grüne
  als auch die Fallback-Adresse. Alle Handys ins normale Vereins-WLAN.
- **Privates DNS (DoH/DoT).** Handys mit aktiviertem privatem DNS fragen nicht den
  Router, sondern direkt einen Internet-DNS-Dienst. Dann greift die Router-Ausnahme
  nicht, und das einzelne Handy bleibt blockiert, während andere funktionieren.
  Abhilfe: privates DNS auf dem Handy vorübergehend auf „Automatisch"/„Aus" stellen
  oder die Fallback-Adresse verwenden.

### Fallback-Adresse

Die Fallback-Adresse `https://<LAN-IP>` funktioniert unabhängig vom
DNS-Rebind-Schutz und auch ohne Internet. Sie zeigt beim ersten Zugriff pro Gerät
eine einmalige Browserwarnung (selbstsigniertes Zertifikat), die bestätigt werden
muss. Danach ist der Verkauf normal möglich. Der Verkauf muss also nie an der
DNS-Frage scheitern.

## Häufige Fragen

**Brauchen wir beim Fest Internet?** Ja, durch die TSE braucht jotti eine Internetverbindung.

**Was, wenn die grüne Adresse nicht lädt?** Mit der Fallback-Adresse weiterarbeiten
und die [Fehlersuche](#fehlersuche) durchgehen. Der Verkauf muss nie warten.

**Brauchen wir wirklich eine TSE, obwohl wir gemeinnützig sind?** Ja. Die
Kassenpflichten gelten unabhängig von der Gemeinnützigkeit, sobald ihr Speisen oder
Getränke gegen Geld verkauft.

**Können wir die TSE erst testen?** Ja. Richtet sie zuerst in der kostenlosen
TEST-Umgebung von fiskaly ein und wechselt erst für den Echtbetrieb auf LIVE.

**Was kostet der Betrieb?** jotti ist für euch kostenlos. Kosten entstehen nur für
die Cloud-TSE von fiskaly (circa 8€/Monat) und, beim Experten-Weg, für den VPS
(circa 5€/Monat). Der Standardweg kommt ohne Servermiete aus.
