---
title: Fehlersuche
description: 'Wenn die grüne Adresse der lokalen jotti-Kasse auf den Handys nicht lädt: DNS-Rebind-Schutz erkennen, Router-Ausnahme eintragen, Fallback nutzen. Und was bei Internet- oder TSE-Ausfall gilt.'
---

> Dieser Abschnitt hilft, wenn die grüne Adresse der lokalen jotti-Kasse auf den
> Handys nicht lädt, obwohl die Kasse läuft. Die Status-Seite
> (`http://localhost:8484`) verlinkt direkt hierher, wenn sie das Problem erkennt.

## Grüne Adresse lädt nicht (DNS-Rebind-Schutz)

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

## Router-Hinweise

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

## Weitere Stolpersteine

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

## Fallback-Adresse

Die Fallback-Adresse `https://<LAN-IP>` funktioniert unabhängig vom
DNS-Rebind-Schutz und auch ohne Internet. Sie zeigt beim ersten Zugriff pro Gerät
eine einmalige Browserwarnung (selbstsigniertes Zertifikat), die bestätigt werden
muss. Danach ist der Verkauf normal möglich.

## Internet oder TSE fällt aus

Weiterverkaufen ist erlaubt, ihr müsst den Verkauf nicht stoppen. jotti bucht
ganz normal weiter und signiert alle in der Ausfallzeit gebuchten Vorgänge
automatisch nach, sobald die Verbindung zur TSE zurück ist. Die Störung wird
dabei automatisch dokumentiert; nachsignierte Belege tragen den Vermerk
„Nachsigniert am …". Ihr müsst nichts weiter tun, nur die Internetverbindung
wiederherstellen (Router prüfen, ggf. neu starten).
