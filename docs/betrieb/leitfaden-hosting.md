# Hosting-Leitfaden: jotti selbst betreiben

## 1. Das Wichtigste in 60 Sekunden

- jotti ist self-hosted: Es läuft nicht in einer fremden Cloud, sondern auf eurer eigenen
  Hardware. Gestartet wird es mit Docker, einem Programm, das jotti samt Datenbank in fertigen
  „Containern" startet, ohne dass ihr irgendetwas von Hand installieren müsst.
- Es gibt zwei Wege, je nach Größe des Fests:
  1. **Weg A (Einzelgerät im WLAN):** jotti läuft auf einem Laptop/PC, bedient von einem
     Tablet oder Handy im selben WLAN. Kein Server, keine Domain, keine laufenden Kosten.
     Ideal für eine einzelne Theke (Direktverkauf, bar, ohne Bondruck).
  2. **Weg B (Eigener Server):** jotti läuft auf einem kleinen gemieteten Server (einem „VPS") mit
     eigener Internet-Adresse (Domain) und Verschlüsselung (HTTPS). Für größere Feste mit
     mehreren Helfern, vielen Tischen und über mehrere Tage.
- **Was kostet das?** jotti selbst ist für berechtigte Vereine kostenlos. Weg A verursacht
  sonst keine laufenden Kosten (nur vorhandene Hardware); bei Weg B kommt ein kleiner VPS ab etwa
  3–8 € pro Monat hinzu (ggf. plus Cloud-TSE-Gebühr, siehe Betreiber-Leitfaden).
- **Technisches Vorwissen?** Grundkenntnisse im Umgang mit der Kommandozeile genügen, alle
  nötigen Befehle stehen in diesem Leitfaden zum Kopieren bereit.
- **Worum es hier _nicht_ geht:** die rechtlichen Pflichten rund um die Kasse (TSE, Finanzamt,
  10 Jahre Aufbewahrung). Das steht im [Betreiber-Leitfaden](./leitfaden-betreiber.md).

---

## 2. Welcher Weg passt zu uns?

| Frage                 | Weg A: Einzelgerät im WLAN               | Weg B: Eigener Server                   |
| --------------------- | ---------------------------------------- | --------------------------------------- |
| Typisches Fest        | Eine Theke, eine Person kassiert         | Mehrere Helfer, viele Tische, mehrtägig |
| Geräte                | 1 Laptop + 1 Tablet/Handy                | Beliebig viele Handys der Helfer        |
| Internet nötig?       | Nein, nur lokales WLAN                   | Ja                                      |
| Domain & HTTPS nötig? | Domain: nein, HTTPS: ja (echtes Zertifikat + Fallback) | Ja                         |
| Bondruck              | Nein                                     | Möglich _(in Entwicklung)_              |
| Laufende Kosten       | Keine                                    | ~3–8 €/Monat (VPS)                      |
| Einrichtungsaufwand   | Sehr gering                              | Etwas höher (einmalig)                  |

> **Faustregel:** Eine kleine Theke an einem Tag? → Weg A. Ein richtiges Fest, bei dem
> Servicekräfte mit ihren eigenen Handys an den Tischen bestellen? → Weg B.

---

## 3. Weg A: Einzelgerät im WLAN (ohne Server)

Für das kleinste Einsatzszenario braucht jotti weder Server noch Domain noch laufende Kosten:
eine einzige Direktverkaufs-Kasse („Theke"), an einem Gerät von einer Person bedient, bar,
ohne Bondruck, nur Eintragen und Kassieren. Der Gast bestellt, zahlt sofort bar, fertig: kein
offener Saldo, keine spätere Ausgabe-Bestätigung.

jotti läuft dabei auf einem vorhandenen Rechner (Windows-Laptop, Mac oder Linux-PC). Ein Tablet
oder Smartphone im selben WLAN bedient die Kasse im Browser über die lokale Adresse des Rechners.

> 🔒 **Grünes Schloss als Normalfall:** Der lokale Betrieb nutzt ein echtes, vom Browser
> anerkanntes Zertifikat (grünes Schloss, keine Warnung) über die Adresse
> `…lokal.jotti.rocks`. jotti holt es beim ersten Start mit Internet automatisch und erneuert es
> selbst, kein CA-Rollout, keine Einrichtung pro Gerät. Welche Adresse gerade gilt und ein
> QR-Code dazu stehen auf der Status-Seite `http://localhost:8484` am Kassenrechner.
>
> Zusätzlich gibt es einen Fallback `https://<LAN-IP>` mit selbstsigniertem Zertifikat (einmalige
> Browserwarnung pro Gerät), er greift vor der ersten Ausstellung, ohne Internet oder wenn der Router
> die grüne Adresse blockiert ([DNS-Rebind-Schutz](./dns-rebind-schutz.md)).
>
> **Restrisiko nur beim Fallback:** Über die grüne Adresse scheitert ein aktiver Angreifer im WLAN
> (MITM) hart. Nur solange ihr den selbstsignierten Fallback nutzt, bleibt (wie bei reinem
> selbstsigniertem Betrieb) ein aktiver MITM möglich.
>
> ℹ️ **Einzeltheke-/localhost-Ausnahme:** Wenn jotti nur direkt am selben Rechner über `localhost`
> bedient wird, entsteht kein WLAN-Transport.

### Voraussetzungen

- Ein Rechner mit Docker Desktop (Windows oder macOS) bzw. Docker Engine + Compose-Plugin
  (Linux), Download: <https://www.docker.com/products/docker-desktop/>
- Rechner und Tablet hängen am selben Router/WLAN.
- Die jotti-Projektdateien liegen auf dem Rechner (ZIP entpackt oder per `git clone`).

### Schnellstart mit dem jotti-Starter (Windows, empfohlen)

Für Windows gibt es einen Doppelklick-Starter, der die `.env` erzeugt, den
Stack hochfährt sowie Docker-Start und Firewall-Freigabe automatisch erledigt,
ganz ohne Kommandozeile.

- **Voraussetzung:** ein Windows-Benutzer mit Administratorrechten, der Starter
  fragt bei jedem Start einmal per UAC nach (er setzt die Firewall-Regel und startet
  Docker Desktop bei Bedarf selbst).
- Das aktuelle Release-ZIP von der
  [GitHub-Releases-Seite](https://github.com/nicograef/jotti/releases) herunterladen,
  entpacken und `jotti-start.exe` doppelklicken. Für den Bondruck zusätzlich
  `jotti-relay.exe`.
- Den vollständigen Ablauf (SmartScreen, UAC, Status-Seite, Beenden) beschreibt die
  `KURZANLEITUNG.md` im ZIP.

Der manuelle Weg unten bleibt die Alternative für macOS, Linux oder fortgeschrittene
Windows-Nutzer.

### Manueller Weg (Kommandozeile)

1. **`.env` anlegen.** Im Projektordner ausführen:

   ```bash
   make init
   ```

Das Kommando erzeugt eine vollständige `.env` mit sicheren Zufallswerten für
`POSTGRES_PASSWORD`, `JWT_SECRET` und `RELAY_AUTH_TOKEN`.

2. **jotti starten.** Im Projektordner:

   ```bash
   docker compose -f docker-compose.local.yml up -d --build
   ```

Der erste Start dauert einige Minuten (die „Container" werden gebaut). Danach laufen Datenbank,
Backend, Frontend und ein Caddy-Reverse-Proxy auf Port 443 (Port 80 leitet auf HTTPS um); Caddy
holt und erneuert das vertrauenswürdige Zertifikat selbst. Mit installiertem `make` alternativ:
`make local-up`, es nennt am Ende die Status-URL.

3. **Status-Seite öffnen.** Am Kassenrechner `http://localhost:8484` im Browser öffnen. Die
   Status-Seite zeigt den aktuellen Zustand, die Zugangsadresse(n) und einen QR-Code. Direkt
   nach dem ersten Start steht dort zunächst die Fallback-Adresse; sobald das Zertifikat ausgestellt
   ist (wenige Sekunden bis Minuten, braucht Internet), wechselt sie automatisch auf die grüne
   Adresse.

4. **Vom Smartphone/Tablet verbinden.** Das Gerät ins Vereins-WLAN bringen (nicht Mobilfunk,
   kein Gastnetz, die [DNS-Rebind-Anleitung](./dns-rebind-schutz.md) erklärt, warum). Dann den
   QR-Code von der Status-Seite scannen oder die grüne Adresse eintippen → grünes Schloss, keine
   Warnung, anmelden. Über „Zum Startbildschirm hinzufügen" lässt sich jotti wie eine App ablegen.

5. **Falls die grüne Adresse (noch) nicht geht.** Die Status-Seite nennt dann die Fallback-Adresse
   `https://<LAN-IP>` (z. B. `https://192.168.1.50`), beim ersten Zugriff pro Gerät die einmalige
   Browserwarnung bestätigen, danach anmelden. Lädt die grüne Adresse auf den Handys gar nicht,
   blockiert vermutlich der Router: dann die [DNS-Rebind-Anleitung](./dns-rebind-schutz.md) befolgen
   (die Status-Seite verlinkt sie ebenfalls).

### Beenden

```bash
docker compose -f docker-compose.local.yml down
```

Die Daten bleiben im Docker-Volume erhalten und stehen beim nächsten Start wieder bereit.
Alternativ: `make local-down`.

### Gut zu wissen

- **Windows-Firewall:** Der jotti-Starter setzt die Freigabe (Port 80/443 im lokalen
  Netz) automatisch. Nur auf dem manuellen Weg fragt Windows beim ersten Start ggf., ob
  der Zugriff erlaubt werden soll, dann für private Netzwerke zulassen, damit das Tablet
  den Rechner über Port 443 erreicht (Port 80 wird nur für Redirect genutzt).
- **Nie ins Internet öffnen.** Der lokale Betrieb ist nur fürs eigene WLAN gedacht. Richtet im
  Router keine Port-Weiterleitung auf den Kassenrechner ein, für den Betrieb übers Internet
  gibt es Weg B mit Domain und HTTPS.
- **Rechner muss laufen.** Während des Betriebs muss der Rechner eingeschaltet und im WLAN sein;
  Energiespar-/Ruhezustand vorher deaktivieren.
- **Erster Start braucht Internet.** Für das vertrauenswürdige Zertifikat braucht der erste Start
  (und der Start nach längerer Pause, wenn das Zertifikat abgelaufen ist) Internet, am besten
  vorab zuhause testen. Ohne Internet läuft der Verkauf über den Fallback weiter.
- **Stabile Adresse (optional).** Bekommt der Rechner per DHCP eine neue LAN-IP, ändert sich der
  Hostname, aber dasselbe Zertifikat gilt weiter (keine neue Warnung) und die Status-Seite zeigt
  die neue Adresse. Wer eine feste Adresse möchte, richtet im Router eine DHCP-Reservierung für
  den Kassenrechner ein.
- **Hardware genügt locker.** Jeder halbwegs aktuelle Laptop bewältigt eine Theke mühelos.
- Für größere Feste (mehrere Helfer, viele Tische, mehrtägig) ist Weg B mit Domain und HTTPS
  die bessere Wahl.

---

## 4. Weg B: Eigener Server (für größere Feste)

Für größere Feste (mehrere Helfer mit eigenen Handys, viele Tische, mehrere Tage) lohnt sich ein
eigener kleiner Server im Internet, ein VPS („Virtual Private Server"). Dann erreichen alle
Helfer jotti über eine Internet-Adresse (Domain) mit Verschlüsselung (HTTPS), nicht nur
im lokalen WLAN vor Ort.

### Welchen Server mieten?

jotti ist sehr genügsam, schon der kleinste VPS reicht für ein durchschnittliches Vereinsfest.
Zwei Richtwerte:

**Minimal:** für ein durchschnittliches Fest (etwa 20 Tische, 10 Helfer gleichzeitig, 2–3 Tage):

| Ressource      | Minimum               |
| -------------- | --------------------- |
| CPU            | 1 vCPU                |
| RAM            | 2 GB                  |
| Speicher       | 20 GB SSD             |
| Netzwerk       | 100 Mbit/s            |
| Betriebssystem | Linux (Debian/Ubuntu) |

Typisches Angebot: netcup VPS 200 oder vergleichbar (~3–4 €/Monat).

**Empfohlen:** für große, mehrtägige Feste (etwa 50 Tische, 30 Helfer gleichzeitig) mit Reserve:

| Ressource      | Empfohlen             |
| -------------- | --------------------- |
| CPU            | 2–4 vCPUs             |
| RAM            | 4 GB                  |
| Speicher       | 40+ GB SSD            |
| Netzwerk       | 200+ Mbit/s           |
| Betriebssystem | Linux (Debian/Ubuntu) |

Typisches Angebot: netcup VPS 500 oder vergleichbar (~5–8 €/Monat).

### Worauf es bei der Hardware ankommt

- **SSD ist Pflicht.** Server mit alter Festplatte (HDD) sind um ein Vielfaches langsamer bei
  Datenbankzugriffen.
- **RAM ist wichtiger als CPU.** Mehr Arbeitsspeicher bedeutet schnellere Auswertungen. Die
  Rechenleistung (CPU) ist selten der Engpass.
- **Speicherplatz ist unkritisch.** Selbst nach Jahren wächst die Datenbank nur auf wenige hundert
  Megabyte.

### Domain und Verschlüsselung (HTTPS)

Für den Server-Betrieb braucht ihr zusätzlich zwei Dinge:

- eine Domain (z. B. `kasse-musterverein.de`), die per DNS auf euren Server zeigt, und
- ein TLS-Zertifikat für HTTPS, kostenlos über Let's Encrypt.

Das Zertifikat holt jotti automatisch. Die Produktions-Konfiguration
(`docker-compose.prod.yml`) bringt einen Caddy-Reverse-Proxy mit, der das Zertifikat beim
ersten Start selbst bei Let's Encrypt anfordert und danach selbst erneuert. Es gibt keine
certbot-Schritte und keine Dateien, die ihr von Hand anpasst: Domain, E-Mail und Version
stehen in der `.env`.

> ⚠️ Ohne HTTPS dürft ihr jotti nicht über das offene Internet betreiben: Anmeldedaten und
> Bestellungen würden sonst unverschlüsselt übertragen.

### Ersteinrichtung

1. **Docker installieren.** Auf dem Server Docker Engine samt Compose-Plugin einrichten
   (Debian/Ubuntu: Anleitung unter <https://docs.docker.com/engine/install/>).

2. **Projektdateien holen.** Das aktuelle Release als ZIP entpacken oder das Repository
   klonen, dann in den Projektordner wechseln.

3. **`.env` anlegen.** Im Projektordner:

   ```bash
   make init
   ```

   Das erzeugt eine `.env` mit sicheren Zufallswerten für die Geheimnisse.

4. **Domain, E-Mail und Version eintragen.** In der `.env` setzen:

   ```bash
   JOTTI_DOMAIN=kasse-musterverein.de
   LETSENCRYPT_EMAIL=vorstand@musterverein.de
   JOTTI_VERSION=v0.2.0
   ```

   `JOTTI_DOMAIN` ist eure Domain, `LETSENCRYPT_EMAIL` eine Kontaktadresse für das
   Let's-Encrypt-Konto, `JOTTI_VERSION` die gewünschte Release-Version (siehe
   [GitHub-Releases](https://github.com/nicograef/jotti/releases); `latest` zieht die neueste).

5. **DNS auf den Server zeigen lassen.** Beim Domain-Anbieter einen A-Record (und bei IPv6
   einen AAAA-Record) auf die öffentliche IP des Servers setzen. Erst wenn die Domain auf
   den Server zeigt, kann Let's Encrypt ein Zertifikat ausstellen.

6. **Stack starten.** Im Projektordner:

   ```bash
   make prod-init
   ```

   Das Skript liest Domain und E-Mail aus der `.env`, prüft Docker und die DNS-Auflösung,
   zieht die gepinnten Images und startet den Stack. Danach wartet es, bis das Backend
   gesund ist und HTTPS antwortet, und gibt eine Zusammenfassung aus. Beim ersten Aufruf
   kann die Zertifikatsausstellung einen Moment dauern.

Danach ist jotti unter `https://<eure-domain>` erreichbar (grünes Schloss); HTTP leitet
automatisch auf HTTPS um.

### jotti aktualisieren

Welche Version läuft, steuert ihr über die `.env`. Tragt unter `JOTTI_VERSION` die
gewünschte Release-Version ein (z. B. `v0.2.0`) und führt dann aus:

```bash
make prod-update
```

Das Skript sichert die Datenbank automatisch, bevor es die neuen Images zieht und die
Migrationen ausführt. Anschließend prüft es, ob der Stack wieder gesund hochkommt
(`/api/health`). Bleibt er ungesund, bricht es ab und zeigt eine kopierfertige Anleitung,
wie ihr mit dem eben erstellten Backup auf die vorherige Version zurückkehrt. Eure vor dem
Update erfassten Daten gehen dabei nicht verloren.

> 🔁 Nur vorwärts, kein Downgrade. Tragt keine ältere als die laufende Version ein. Updates
> verändern die Datenbank und lassen sich nicht zurücknehmen; eine ältere Version kann mit den
> neuen Daten nicht mehr starten. `make prod-update` verweigert ein solches Downgrade. Wollt ihr
> zurück, spielt stattdessen ein Backup ein (`make prod-restore`).

### Backups (sichern und wiederherstellen)

jotti speichert alle Daten in einem Docker-Volume. Macht regelmäßige Backups der Datenbank,
schon wegen der gesetzlichen 10-Jahre-Aufbewahrung (Hintergrund im
[Betreiber-Leitfaden, Schritt 4](./leitfaden-betreiber.md#schritt-4-daten-10-jahre-aufbewahren)).

- **Sichern.** `make prod-backup` zieht einen komprimierten `pg_dump` in den Ordner
  `BACKUP_DIR` (Standard `./backups`) und behält die neuesten `BACKUP_KEEP` Stück
  (Standard 14). Beide Werte lassen sich in der `.env` anpassen.
- **Wiederherstellen.** `make prod-restore` listet die vorhandenen Backups, fragt eine
  Bestätigung ab und spielt das gewählte (standardmäßig das neueste) zurück. Das
  überschreibt die aktuelle Datenbank.

> 💾 Kopiert die Backups regelmäßig vom Server weg (auf einen anderen Rechner oder Speicher).
> Ein Backup, das nur auf demselben Server liegt, hilft bei dessen Ausfall nicht.

**Täglich automatisch sichern.** Für einen täglichen Dump liegen fertige Vorlagen im
Repository: ein systemd-Timer (`packaging/systemd/jotti-backup.service` und
`jotti-backup.timer`) oder alternativ ein cron-Eintrag (`packaging/cron/jotti-backup.cron`).
Die Installationsschritte stehen als Kommentar in den Dateien; passt darin den Pfad zu eurem
jotti-Ordner an.

### Server härten (optional)

Ein öffentlich erreichbarer Server sollte abgesichert werden. jotti bringt dafür ein
optionales Skript mit, das ihr bei Bedarf einmalig ausführt:

```bash
make prod-harden
```

Es richtet eine Firewall (ufw) ein, die nur SSH sowie die Web-Ports 80 und 443 erlaubt und
alles andere von außen sperrt (die Datenbank ist ohnehin nie nach außen geöffnet). Den
SSH-Zugang gibt es frei, bevor die Firewall scharf geschaltet wird, damit ihr euch nicht
aussperrt. Optional aktiviert es fail2ban gegen Anmelde-Angriffe auf SSH und weist auf
automatische Sicherheitsupdates hin. Das Skript ist idempotent (mehrfaches Ausführen schadet
nicht) und kein Teil der Ersteinrichtung.

> ⚠️ Öffnet nach dem Härten eine zweite SSH-Sitzung, bevor ihr die erste schließt, um
> sicherzugehen, dass ihr weiter Zugriff habt.

---

## 5. Glossar

| Begriff                     | Einfache Erklärung                                                                         |
| --------------------------- | ------------------------------------------------------------------------------------------ |
| Self-hosted                 | Ihr betreibt die Software auf eurer eigenen Hardware, nicht in einer fremden Cloud.        |
| Docker                      | Programm, das jotti samt Datenbank in fertigen „Containern" startet, ohne Installieren.    |
| Container                   | Ein abgeschlossenes Paket mit einem laufenden Programmteil (Datenbank, Backend, Frontend). |
| VPS                         | „Virtual Private Server", ein kleiner, gemieteter Server im Internet.                      |
| WLAN / LAN                  | Euer lokales (Funk-)Netzwerk vor Ort.                                                      |
| IP-Adresse                  | Die „Hausnummer" eines Geräts im Netzwerk, z. B. `192.168.1.50`.                           |
| Domain                      | Eine Internet-Adresse wie `kasse-musterverein.de`.                                         |
| HTTPS / TLS                 | Verschlüsselte Verbindung im Internet (Schloss-Symbol im Browser).                         |
| Let's Encrypt               | Kostenlose Stelle, die HTTPS-Zertifikate ausstellt.                                        |
| `.env`                      | Konfigurationsdatei mit Passwörtern und Geheimnissen, niemals öffentlich teilen.           |
| Reverse Proxy (Caddy)       | Vermittler, der Anfragen sicher an jotti weiterleitet und HTTPS bereitstellt.              |
| Docker-Volume               | Der Speicherort, an dem Docker eure Daten dauerhaft aufbewahrt.                            |
| SSD                         | Schnelle Festplatte (Flash-Speicher), Pflicht für flüssigen Betrieb.                       |
| Direktverkauf („Theke")     | Verkauf, bei dem sofort bar kassiert wird, kein offener Tisch-Saldo.                       |
