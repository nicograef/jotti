# DNS-Rebind-Schutz: Ausnahme für `lokal.jotti.rocks` einrichten

Diese Anleitung hilft, wenn die **vertrauenswürdige Adresse** der lokalen
jotti-Kasse (`https://<ip-mit-bindestrichen>.<install-id>.lokal.jotti.rocks`,
grünes Schloss) auf den Helfer-Handys **nicht lädt**, obwohl die Kasse läuft.
Die Status-Seite (`http://localhost:8484`) verlinkt direkt hierher, wenn sie das
Problem erkennt.

## 1. Worum geht es?

jotti holt für den lokalen Betrieb ein echtes Let's-Encrypt-Zertifikat auf einen
Namen, der eure **private** LAN-IP enthält (z. B.
`192-168-1-50.<id>.lokal.jotti.rocks` → `192.168.1.50`). Das ist gewollt und
sicher — aber viele Router haben einen **DNS-Rebind-Schutz**, der genau diese
Kombination („öffentlicher Name zeigt auf eine private IP") als möglichen Angriff
einstuft und **blockiert**. Die Antwort des jotti-Resolvers kommt dann im WLAN
leer an, und das Handy kann die grüne Adresse nicht öffnen.

Die Lösung ist eine **einmalige Ausnahme** im Router: `lokal.jotti.rocks` von der
Prüfung ausnehmen. Danach funktioniert die grüne Adresse im gesamten
Vereins-WLAN.

> 🔒 **Sicherheit:** Die Ausnahme erlaubt private IPs **nur** für die jotti-Domain
> `lokal.jotti.rocks` — der Rebind-Schutz für alle anderen Domains bleibt aktiv.

## 2. Symptom erkennen

DNS-Rebind-Schutz ist die wahrscheinliche Ursache, wenn:

- die **Fallback-Adresse** `https://<LAN-IP>` (z. B. `https://192.168.1.50`,
  einmalige Browserwarnung) **funktioniert**, die **grüne** Adresse aber nicht;
- der Verkauf grundsätzlich läuft, nur das grüne Schloss fehlt;
- es **„auf Handy A geht, auf Handy B aber nicht"** — siehe Abschnitt 5 (privates
  DNS).

Bis die Ausnahme eingetragen ist, könnt ihr **jederzeit mit der Fallback-Adresse
weiterarbeiten**. Der Verkauf muss nicht warten.

## 3. Fritz!Box (häufigster Router im Vereinsumfeld)

Die Fritz!Box ist der mit Abstand häufigste Router in Vereinsheimen. Der
Rebind-Schutz lässt sich nicht abschalten, aber gezielt um eine Domain ergänzen.

1. Im Browser **`http://fritz.box`** öffnen und am Router anmelden.
2. **Heimnetz → Netzwerk → Netzwerkeinstellungen** öffnen.
3. Auf **„Weitere Einstellungen"** bzw. **„Erweiterte Netzwerkeinstellungen
   ändern"** klicken (je nach FRITZ!OS-Version).
4. Den Abschnitt **„DNS-Rebind-Schutz"** suchen.
5. Im Feld **„Diese Domain(s) ausnehmen"** (Hostname-Ausnahmen) genau Folgendes
   eintragen:

   ```
   lokal.jotti.rocks
   ```

6. Mit **„Übernehmen"** speichern. Danach die grüne Adresse auf dem Handy erneut
   öffnen.

> 📷 _Screenshot folgt (Heimnetz → Netzwerk → Netzwerkeinstellungen →
> DNS-Rebind-Schutz → Ausnahme) — an echter Fritz!Box-Hardware ergänzen._

> ℹ️ **Ein Eintrag oder mehrere?** Die Fritz!Box unterstützt **keine** Wildcards
> (`*.…`). Die Vermutung (Community-Praxis zu `plex.direct`) ist, dass der Eintrag
> der **Domain** `lokal.jotti.rocks` auch alle Subdomains
> (`<ip>.<id>.lokal.jotti.rocks`) abdeckt — dann genügt der eine Eintrag oben.
> Dies ist an echter Hardware noch zu bestätigen. **Falls die grüne Adresse danach
> weiterhin blockiert wird**, zusätzlich den **vollständigen Hostnamen** aus der
> Status-Seite eintragen (z. B. `192-168-1-50.<install-id>.lokal.jotti.rocks`).
> Ändert sich die LAN-IP (DHCP), ändert sich dann auch dieser Eintrag.

## 4. Andere Router

Das Prinzip ist überall gleich: `lokal.jotti.rocks` von der Rebind-Prüfung
ausnehmen. Die Bezeichnungen unterscheiden sich.

| Router / DNS                   | Wo                                                              | Eintrag                                    |
| ------------------------------ | -------------------------------------------------------------- | ------------------------------------------ |
| **Pi-hole / dnsmasq**          | Konfigurationsdatei (`/etc/dnsmasq.d/…`)                        | `rebind-domain-ok=/lokal.jotti.rocks/`     |
| **OpenWrt** (LuCI)             | Network → DHCP and DNS → Rebind protection → **Domain whitelist** | `lokal.jotti.rocks`                        |
| **OpenWrt** (Datei)            | `/etc/config/dhcp` (Abschnitt `dnsmasq`)                        | `list rebind_domain 'lokal.jotti.rocks'`   |
| **Speedport** (Telekom, z. B. Smart 4) | Netzwerk → **DNS-Rebind-Schutz** → Domain zur Liste hinzufügen | `lokal.jotti.rocks`                        |

> Nach jeder Änderung den DNS-Dienst des Routers neu laden bzw. neu starten
> (bei Pi-hole/dnsmasq z. B. `pihole restartdns` bzw. `service dnsmasq restart`).

Hat euer Router keinen Rebind-Schutz, blockiert er auch nichts — dann liegt die
Ursache woanders (siehe Abschnitt 5).

## 5. WLAN- und Diagnose-Hinweise

Auch ohne Rebind-Schutz gibt es ein paar Stolpersteine. Die grüne Adresse
funktioniert nur, wenn das Handy die **private LAN-IP** des Kassenrechners
tatsächlich erreichen kann:

- **Vereins-WLAN, nicht Mobilfunk.** Der Name löst zwar auch im Mobilfunknetz auf
  (auf dieselbe private IP), aber diese IP ist von außerhalb des WLAN **nicht
  erreichbar**. Das Handy muss im **selben WLAN** wie der Kassenrechner sein.
- **Kein Gastnetz.** Gastnetze haben eine **Client-Isolation**: Geräte im Gastnetz
  dürfen das Kassengerät grundsätzlich nicht erreichen — das blockiert sowohl die
  grüne als auch die Fallback-Adresse. Alle Handys ins **normale Vereins-WLAN**.
- **Privates DNS (DoH/DoT) — „geht auf Handy A, nicht auf Handy B".** Handys mit
  aktiviertem **privatem DNS** (DNS over HTTPS/TLS, z. B. Android-Einstellung
  „Privates DNS" oder iOS-DNS-Profile) fragen **nicht** den Router, sondern direkt
  einen Internet-DNS-Dienst. Dann greift die Router-Ausnahme aus Abschnitt 3/4
  nicht — und das einzelne Handy bleibt blockiert, während andere funktionieren.
  Abhilfe auf dem betroffenen Handy: privates DNS vorübergehend auf
  „Automatisch"/„Aus" stellen, **oder** einfach die **Fallback-Adresse**
  `https://<LAN-IP>` (einmalige Warnung) verwenden.

## 6. Wenn nichts hilft

Die **Fallback-Adresse** `https://<LAN-IP>` funktioniert unabhängig vom
DNS-Rebind-Schutz und auch ohne Internet. Sie zeigt beim ersten Zugriff pro Gerät
eine einmalige Browserwarnung (selbstsigniertes Zertifikat), die bestätigt werden
muss — danach ist der Verkauf normal möglich. Der Verkauf muss also nie an der
DNS-Frage scheitern.
