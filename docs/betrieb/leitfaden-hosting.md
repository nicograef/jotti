# Hosting-Leitfaden: jotti selbst betreiben

## 1. Das Wichtigste in Kürze

- jotti ist self-hosted: Es läuft auf eurer eigenen Hardware, nicht in einer fremden Cloud.
  Gestartet wird es mit Docker, das jotti samt Datenbank in fertigen „Containern" bereitstellt,
  ohne dass ihr etwas von Hand installieren müsst.
- Es gibt zwei Wege:
  1. **Standard-Weg (Windows-PC im Vereinsheim):** Ein vorhandener Windows-Rechner wird zum
     Kassenrechner, die Helfer bedienen jotti auf ihren Handys im selben WLAN. Start per
     Doppelklick, keine Domain, keine laufenden Kosten. Für die allermeisten Vereinsfeste die
     richtige Wahl, auch mit mehreren Helfern, vielen Tischen und Bondruck.
  2. **Experten-Weg (Linux-Server/VPS):** jotti läuft auf einem gemieteten kleinen Server mit
     eigener Internet-Adresse (Domain) und Verschlüsselung (HTTPS) und ist auch außerhalb des
     lokalen WLANs erreichbar. Setzt Grundkenntnisse mit Linux und Kommandozeile voraus.
- **Was kostet das?** jotti selbst ist für berechtigte Vereine kostenlos. Der Standard-Weg
  verursacht keine laufenden Kosten (nur vorhandene Hardware); beim Experten-Weg kommt ein
  kleiner VPS ab etwa 3 bis 8 € pro Monat hinzu (ggf. plus Cloud-TSE-Gebühr, siehe
  Betreiber-Leitfaden).
- **Worum es hier _nicht_ geht:** die rechtlichen Pflichten rund um die Kasse (TSE, Finanzamt,
  10 Jahre Aufbewahrung). Das steht im [Betreiber-Leitfaden](./leitfaden-betreiber.md).

---

## 2. Welcher Weg passt zu uns?

| Frage                 | Standard-Weg: Windows-PC im Vereinsheim                | Experten-Weg: Linux-Server (VPS)             |
| --------------------- | ------------------------------------------------------ | -------------------------------------------- |
| Geeignet für          | Fast alle Vereinsfeste vor Ort                         | Betrieb über das Internet, mehrere Standorte |
| Bedienung             | Kassenrechner + Handys der Helfer im WLAN              | Handys der Helfer über das Internet          |
| Internet nötig?       | Nur beim ersten Start (Zertifikat)                     | Ja, dauerhaft                                |
| Domain & HTTPS        | Domain nein, HTTPS automatisch (Zertifikat + Fallback) | Domain ja, HTTPS automatisch                 |
| Bondruck              | Ja (mit `jotti-relay.exe`)                             | Ja                                           |
| Laufende Kosten       | Keine                                                  | ~3 bis 8 €/Monat (VPS)                       |
| Technisches Vorwissen | Doppelklick, keine Kommandozeile                       | Linux und Kommandozeile                      |

> **Faustregel:** Alle Helfer sind beim Fest vor Ort im selben WLAN? → Standard-Weg. Ihr wollt
> jotti von überall über das Internet erreichen oder keinen Rechner vor Ort betreiben? →
> Experten-Weg.

---

## 3. Standard-Weg: Windows-PC im Vereinsheim

Ein vorhandener Windows-Rechner im Vereinsheim wird zum Kassenrechner: Er startet jotti, die
Helfer bedienen es auf ihren Handys im selben WLAN. Kein Server, keine Domain, keine laufenden
Kosten. Das genügt für die allermeisten Vereinsfeste, auch mit mehreren Helfern, vielen Tischen
und Bondruck.

> 🔒 **Grünes Schloss als Normalfall:** Für den lokalen Betrieb holt jotti automatisch ein
> echtes, vom Browser anerkanntes Zertifikat über die Adresse `…lokal.jotti.rocks` (grünes
> Schloss, keine Warnung). Es wird beim ersten Start mit Internet ausgestellt und selbst
> erneuert, ohne Einrichtung pro Gerät. Welche Adresse gerade gilt, zeigt samt QR-Code die
> Status-Seite `http://localhost:8484` am Kassenrechner.
>
> Greift die grüne Adresse nicht (vor der ersten Ausstellung, ohne Internet oder bei
> [DNS-Rebind-Schutz](./dns-rebind-schutz.md) im Router), springt ein Fallback `https://<LAN-IP>`
> mit selbstsigniertem Zertifikat ein (einmalige Browserwarnung pro Gerät). Nur über diesen
> Fallback bleibt ein aktiver Angreifer im WLAN (MITM) möglich, über die grüne Adresse nicht.

### Voraussetzungen

- Ein Windows-Rechner mit Administratorrechten (der Starter fragt bei jedem Start einmal per UAC
  nach).
- Docker Desktop ist installiert, nur installieren, nicht vorab starten:
  <https://www.docker.com/products/docker-desktop/>
- Rechner und Handys hängen am selben Router/WLAN.

### Schnellstart per Doppelklick (empfohlen)

Für Windows gibt es einen Doppelklick-Starter, der die `.env` erzeugt, den Stack hochfährt und
Docker-Start sowie Firewall-Freigabe selbst erledigt, ganz ohne Kommandozeile.

1. Das aktuelle Release-ZIP von der
   [GitHub-Releases-Seite](https://github.com/nicograef/jotti/releases) herunterladen und
   entpacken (alle Dateien bleiben im selben Ordner).
2. `jotti-start.exe` doppelklicken. Beim ersten Mal dauert der Start einige Minuten (Programmteile
   werden heruntergeladen). SmartScreen mit „Weitere Informationen" → „Trotzdem ausführen" und
   UAC mit „Ja" bestätigen.
3. Wenn alles läuft, zeigt der Starter den Hinweis auf die Status-Seite `http://localhost:8484`.
   Diese am Kassenrechner im Browser öffnen, dort stehen die Zugangsadresse und ein QR-Code.

> ⚠️ Den ersten Start unbedingt vorab zuhause mit Internet machen, nicht erst auf dem Fest. Beim
> Erststart lädt jotti seine Programmteile herunter und holt das vertrauenswürdige Zertifikat.
> Beides braucht Internet. Danach läuft jotti auch ohne Internet.

Den vollständigen Ablauf (SmartScreen, UAC, Bondruck, Beenden, Aktualisieren) beschreibt die
`KURZANLEITUNG.md` im ZIP. Für den Bondruck zusätzlich `jotti-relay.exe` doppelklicken.

### Helfer-Handys verbinden

Das Handy ins Vereins-WLAN bringen (nicht Mobilfunk, kein Gastnetz, die
[DNS-Rebind-Anleitung](./dns-rebind-schutz.md) erklärt, warum). Dann den QR-Code von der
Status-Seite scannen oder die grüne Adresse eintippen → grünes Schloss, keine Warnung, anmelden.
Über „Zum Startbildschirm hinzufügen" lässt sich jotti wie eine App ablegen.

Falls die grüne Adresse (noch) nicht geht, nennt die Status-Seite die Fallback-Adresse
`https://<LAN-IP>` (z. B. `https://192.168.1.50`); beim ersten Zugriff pro Gerät die einmalige
Browserwarnung bestätigen, danach anmelden. Lädt die grüne Adresse auf den Handys gar nicht,
blockiert vermutlich der Router: dann die [DNS-Rebind-Anleitung](./dns-rebind-schutz.md) befolgen
(die Status-Seite verlinkt sie ebenfalls).

### Beenden

`jotti-stop.cmd` doppelklicken (oder in Docker Desktop stoppen). Die Daten bleiben im
Docker-Volume erhalten und stehen beim nächsten Start wieder bereit.

### Manueller Weg (macOS, Linux, fortgeschrittene Windows-Nutzer)

Ohne den Doppelklick-Starter läuft der lokale Betrieb über die Kommandozeile.

1. **`.env` anlegen.** Im Projektordner (ZIP entpackt oder per `git clone`):

   ```bash
   make init
   ```

   Das erzeugt eine vollständige `.env` mit sicheren Zufallswerten für `POSTGRES_PASSWORD`,
   `JWT_SECRET` und `RELAY_AUTH_TOKEN`.

2. **jotti starten.** Im Projektordner:

   ```bash
   docker compose -f docker-compose.local.yml up -d --build
   ```

   Der erste Start dauert einige Minuten (die Container werden gebaut). Danach laufen Datenbank,
   Backend, Frontend und ein Caddy-Reverse-Proxy auf Port 443 (Port 80 leitet auf HTTPS um);
   Caddy holt und erneuert das vertrauenswürdige Zertifikat selbst. Mit installiertem `make`
   alternativ `make local-up`, es nennt am Ende die Status-URL. Beenden mit
   `docker compose -f docker-compose.local.yml down` bzw. `make local-down`.

3. **Status-Seite und Handys.** Weiter wie oben: `http://localhost:8484` am Kassenrechner öffnen,
   dann die Handys über QR-Code bzw. grüne Adresse verbinden.

> Auf dem manuellen Weg fragt Windows beim ersten Start ggf., ob der Zugriff erlaubt werden soll;
> dann für private Netzwerke zulassen, damit das Tablet den Rechner über Port 443 erreicht (Port
> 80 wird nur für den Redirect genutzt). Der Doppelklick-Starter setzt diese Freigabe selbst.

### Gut zu wissen

- **Nie ins Internet öffnen.** Der lokale Betrieb ist nur fürs eigene WLAN gedacht. Richtet im
  Router keine Port-Weiterleitung auf den Kassenrechner ein; für den Betrieb übers Internet gibt
  es den Experten-Weg mit Domain und HTTPS.
- **Rechner muss laufen.** Während des Betriebs muss der Rechner eingeschaltet und im WLAN sein;
  Energiespar-/Ruhezustand vorher deaktivieren.
- **Erster Start braucht Internet.** Für das vertrauenswürdige Zertifikat braucht der erste Start
  (und der Start nach längerer Pause, wenn das Zertifikat abgelaufen ist) Internet, am besten
  vorab zuhause testen. Ohne Internet läuft der Verkauf über den Fallback weiter.
- **Stabile Adresse (optional).** Bekommt der Rechner per DHCP eine neue LAN-IP, ändert sich der
  Hostname, aber dasselbe Zertifikat gilt weiter (keine neue Warnung) und die Status-Seite zeigt
  die neue Adresse. Wer eine feste Adresse möchte, richtet im Router eine DHCP-Reservierung für
  den Kassenrechner ein.
- **Noch kleiner geht auch.** Wird jotti nur direkt am Kassenrechner über `localhost` bedient
  (eine Theke, eine Person), entsteht kein WLAN-Transport und es braucht weder Zertifikat noch
  Handy-Verbindung.
- **Hardware genügt locker.** Jeder halbwegs aktuelle Rechner bewältigt auch ein größeres Fest.

---

## 4. Experten-Weg: Linux-Server (VPS)

Wer jotti auch außerhalb des lokalen WLANs erreichen will (über das Internet, an mehreren
Standorten oder ohne einen Rechner vor Ort), betreibt es auf einem eigenen kleinen Server, einem
VPS („Virtual Private Server"). Dann erreichen alle Helfer jotti über eine Internet-Adresse
(Domain) mit Verschlüsselung (HTTPS).

### Welchen Server mieten?

jotti ist sehr genügsam, schon der kleinste VPS reicht für ein durchschnittliches Vereinsfest
(etwa 20 Tische, 10 Helfer gleichzeitig, 2 bis 3 Tage):

| Ressource      | Minimum               |
| -------------- | --------------------- |
| CPU            | 1 vCPU                |
| RAM            | 2 GB                  |
| Speicher       | 20 GB SSD             |
| Netzwerk       | 100 Mbit/s            |
| Betriebssystem | Linux (Debian/Ubuntu) |

Typisches Angebot: netcup VPS 200 oder vergleichbar (~3 bis 4 €/Monat).

### Domain und Verschlüsselung (HTTPS)

Für den Server-Betrieb braucht ihr zusätzlich zwei Dinge:

- eine Domain (z. B. `kasse-musterverein.de`), die per DNS auf euren Server zeigt, und
- ein TLS-Zertifikat für HTTPS, kostenlos über Let's Encrypt.

Das Zertifikat holt jotti automatisch. Die Produktions-Konfiguration (`docker-compose.prod.yml`)
bringt einen Caddy-Reverse-Proxy mit, der das Zertifikat beim ersten Start selbst bei Let's
Encrypt anfordert und danach erneuert. Es gibt keine certbot-Schritte und keine Dateien, die ihr
von Hand anpasst: Domain, E-Mail und Version stehen in der `.env`.

> ⚠️ Ohne HTTPS dürft ihr jotti nicht über das offene Internet betreiben: Anmeldedaten und
> Bestellungen würden sonst unverschlüsselt übertragen.

### Ersteinrichtung

1. **Docker installieren.** Auf dem Server Docker Engine samt Compose-Plugin einrichten
   (Debian/Ubuntu: Anleitung unter <https://docs.docker.com/engine/install/>).

2. **Projektdateien holen.** Das aktuelle Release als ZIP entpacken oder das Repository klonen,
   dann in den Projektordner wechseln.

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

5. **DNS auf den Server zeigen lassen.** Beim Domain-Anbieter einen A-Record (und bei IPv6 einen
   AAAA-Record) auf die öffentliche IP des Servers setzen. Erst wenn die Domain auf den Server
   zeigt, kann Let's Encrypt ein Zertifikat ausstellen.

6. **Stack starten.** Im Projektordner:

   ```bash
   make prod-init
   ```

   Das Skript liest Domain und E-Mail aus der `.env`, prüft Docker und die DNS-Auflösung, zieht
   die gepinnten Images und startet den Stack. Danach wartet es, bis das Backend gesund ist und
   HTTPS antwortet, und gibt eine Zusammenfassung aus. Beim ersten Aufruf kann die
   Zertifikatsausstellung einen Moment dauern.

Danach ist jotti unter `https://<eure-domain>` erreichbar (grünes Schloss); HTTP leitet
automatisch auf HTTPS um.

### jotti aktualisieren

Welche Version läuft, steuert ihr über die `.env`. Tragt unter `JOTTI_VERSION` die gewünschte
Release-Version ein (z. B. `v0.2.0`) und führt dann aus:

```bash
make prod-update
```

Das Skript sichert die Datenbank automatisch, bevor es die neuen Images zieht und die Migrationen
ausführt. Anschließend prüft es, ob der Stack wieder gesund hochkommt (`/api/health`). Bleibt er
ungesund, bricht es ab und zeigt eine kopierfertige Anleitung, wie ihr mit dem eben erstellten
Backup auf die vorherige Version zurückkehrt. Eure vor dem Update erfassten Daten gehen dabei
nicht verloren.

> 🔁 Nur vorwärts, kein Downgrade. Tragt keine ältere als die laufende Version ein. Updates
> verändern die Datenbank und lassen sich nicht zurücknehmen; eine ältere Version kann mit den
> neuen Daten nicht mehr starten. `make prod-update` verweigert ein solches Downgrade. Wollt ihr
> zurück, spielt stattdessen ein Backup ein (`make prod-restore`).

### Backups (sichern und wiederherstellen)

jotti speichert alle Daten in einem Docker-Volume. Macht regelmäßige Backups der Datenbank, schon
wegen der gesetzlichen 10-Jahre-Aufbewahrung (Hintergrund im
[Betreiber-Leitfaden, Schritt 4](./leitfaden-betreiber.md#schritt-4-daten-10-jahre-aufbewahren)).

- **Sichern.** `make prod-backup` zieht einen komprimierten `pg_dump` in den Ordner `BACKUP_DIR`
  (Standard `./backups`) und behält die neuesten `BACKUP_KEEP` Stück (Standard 14). Beide Werte
  lassen sich in der `.env` anpassen.
- **Wiederherstellen.** `make prod-restore` listet die vorhandenen Backups, fragt eine
  Bestätigung ab und spielt das gewählte (standardmäßig das neueste) zurück. Das überschreibt die
  aktuelle Datenbank.

> 💾 Kopiert die Backups regelmäßig vom Server weg (auf einen anderen Rechner oder Speicher). Ein
> Backup, das nur auf demselben Server liegt, hilft bei dessen Ausfall nicht.

**Täglich automatisch sichern.** Für einen täglichen Dump liegen fertige Vorlagen im Repository:
ein systemd-Timer (`packaging/systemd/jotti-backup.service` und `jotti-backup.timer`) oder
alternativ ein cron-Eintrag (`packaging/cron/jotti-backup.cron`). Die Installationsschritte stehen
als Kommentar in den Dateien; passt darin den Pfad zu eurem jotti-Ordner an.

### Server härten (optional)

Ein öffentlich erreichbarer Server sollte abgesichert werden. jotti bringt dafür ein optionales
Skript mit, das ihr bei Bedarf einmalig ausführt:

```bash
make prod-harden
```

Es richtet eine Firewall (ufw) ein, die nur SSH sowie die Web-Ports 80 und 443 erlaubt und alles
andere von außen sperrt (die Datenbank ist ohnehin nie nach außen geöffnet). Den SSH-Zugang gibt
es frei, bevor die Firewall scharf geschaltet wird, damit ihr euch nicht aussperrt. Optional
aktiviert es fail2ban gegen Anmelde-Angriffe auf SSH und weist auf automatische
Sicherheitsupdates hin. Das Skript ist idempotent (mehrfaches Ausführen schadet nicht) und kein
Teil der Ersteinrichtung.

> ⚠️ Öffnet nach dem Härten eine zweite SSH-Sitzung, bevor ihr die erste schließt, um
> sicherzugehen, dass ihr weiter Zugriff habt.
