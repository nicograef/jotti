# ADR 09: Cloud-TSE bei fiskaly behalten, keine Hardware-TSE

- **Status:** akzeptiert (2026-09-07)
- **Kontext-Dokumente:** `docs/compliance.md` §3.5,
  `docs/leitfaden/tse-einrichten.md`, `docs/leitfaden/haeufige-fragen.md`,
  `docs/rechtsquellen/` (AO, KassenSichV, AEAO),
  `backend/domain/tse/client.go` (`TSEClient`)

## Kontext

Ein Verein meldete für die fiskaly-Cloud-TSE eine Mindestabnahme mehrerer Kassen
und eine Mindestlaufzeit. Für zwei bis drei Feste im Jahr ist das die eigentliche
Hürde, nicht der Monatspreis. Eine Web-Recherche sollte deshalb drei Fragen
klären: Was gilt bei fiskaly für **eine** Kasse? Verkauft ein anderer zertifizierter
Anbieter eine Kurzzeitlizenz? Was würde eine Hardware-TSE am Windows-Starter
bedeuten? Alle Quellen dieses ADR wurden am 07.09.2026 abgerufen.

### Konditionen bei fiskaly

| Punkt                           | Befund                                                                | Quelle                                                                                                                              |
| ------------------------------- | --------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| Listenpreis                     | nicht öffentlich; keine Preisseite                                    | [fiskaly SIGN DE](https://www.fiskaly.com/signde)                                                                                   |
| TEST-Umgebung                   | kostenlos, ohne Kreditkarte                                           | [fiskaly Workspace](https://workspace.fiskaly.com/)                                                                                 |
| LIVE-Zugang                     | nur mit Vertrag über den Vertrieb, kein Selfservice-Kauf              | [Switch to LIVE](https://developer.fiskaly.com/hub/switch_to_live)                                                                  |
| Reseller-Orientierung           | 13,25 €/Monat netto oder 143,00 €/Jahr netto je Kasse, inkl. DSFinV-K | [HKSoftware-Shop](https://www.hksoftware-shop.de/p/tse-fiskaly-cloud-inkl-schnittstelle-dsfinv-k)                                   |
| Mindestlaufzeit, Mindestabnahme | nicht öffentlich, nicht verifizierbar                                 | keine Quelle gefunden                                                                                                               |
| Zertifizierung                  | BSI-zertifiziert nach TR-03153, gültig bis 30.03.2033                 | [BSI-K-TR-0717-2025](https://www.bsi.bund.de/SharedDocs/Zertifikate_TR/Technische_Sicherheitseinrichtungen/BSI-K-TR-0717-2025.html) |

Der Vereinsbericht ließ sich damit weder bestätigen noch widerlegen. Belegt ist
allein der Vertragszwang für LIVE. Ein Direktvertrag eines e. V. ist nicht
ausgeschlossen, aber auch nicht belegt.

### Erwogene Alternativen

1. **Swissbit Cloud-TSE 2** — die Laufzeitangaben widersprechen sich: der
   Hersteller schreibt „cancelable on a monthly basis", der Handel nennt
   36 Monate. Öffentlich bepreist ist nur das Abo: Typ S kostet 178,10 € netto
   für drei Jahre
   ([Swissbit](https://www.swissbit.com/en/products/security-products/swissbit-tse/cloud-tse/),
   [acventis](https://acventis.com/SWISSBIT-Cloud-TSE-2-TSE-Cloud-Services-Typ-S-3-Jahre)).
   Eine öffentliche OpenAPI-Spezifikation der beworbenen REST-API war nicht
   auffindbar — das ist der entscheidende Punkt.
2. **Deutsche Fiskal / D-Trust** — verlangt einen lokalen „Fiskal Cloud
   Connector" statt reiner HTTPS-Aufrufe. Wiederverkäufer nennen 12,00 € bzw.
   14,99 €/Monat netto je Kasse bei zwölf Monaten Laufzeit
   ([JTL](https://www.jtl-software.de/jtl-store/erweiterungen/fiskal-cloud-tse),
   [S&S ITS](https://sundsits.com/en/fiskal-cloud-tse-je-kasse/)). Ein
   Herstellerpreis ist nicht öffentlich; API-Auskunft gibt D-Trust nur auf
   Anfrage.
3. **Diebold Nixdorf** — die DN TSE-Cloud ist ein Webservice mit
   JSON-Schnittstelle. Preise und Konditionen nennt der Anbieter nur auf Anfrage;
   die Schnittstellendokumentation wurde nicht geprüft
   ([DN](https://www.dieboldnixdorf.com/de-de/retail/portfolio/systems/tse/)).

[EFSTA](https://www.efsta.eu/en/solutions/tse-germany) ist kein TSE-Anbieter,
sondern eine Middleware („EFR"). Sie bündelt fiskaly, Deutsche Fiskal und
Hardware-TSE und käme nur als zusätzliche Schicht in Frage.

Kein Anbieter verkauft öffentlich eine Lizenz für ein einzelnes Wochenende.

### Hardware-TSE am Windows-Starter

| Punkt               | Befund                                                                                  | Quelle                                                                         |
| ------------------- | --------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| Preis               | Swissbit USB-TSE ab 184,95 € brutto, TSE v2 (microSD/SD/USB) ab 194,95 € brutto         | [gebongt24](https://www.gebongt24.de/kassenzubehoer/tse/)                      |
| Zertifikatslaufzeit | 5 Jahre oder 20 Mio. Signaturen                                                         | [gebongt24](https://www.gebongt24.de/kassenzubehoer/tse/)                      |
| Laufzeitbeginn      | mit der Fertigung, nicht mit der Inbetriebnahme                                         | [HKS Systeme](https://www.hks-systeme.de/ablauf-der-swissbit-tse-zertifikate/) |
| SDK                 | WORM-Bibliothek (`WormAPI.dll` bzw. `libWormAPI.so`), Weitergabe vertraglich beschränkt | [python-tse](https://github.com/bwurst/python-tse)                             |
| Go-Bindung          | keine offizielle gefunden                                                               | keine Quelle gefunden                                                          |

Bei Jahresabrechnung kostet die Cloud-TSE über fünf Jahre rund 715 € netto
(5 × 143,00 €). Monatlich abgerechnet sind es 795 € netto (60 × 13,25 €). Die
Hardware-TSE kostet einmalig 184,95–194,95 € brutto, also rund 155–164 € netto.
Über fünf Jahre ist die Cloud damit etwa das Vierfache der Hardware. Gleichauf
läge sie nur bei monatlicher Abrechnung allein in den Festmonaten
(2–3 × 13,25 € × 5 ≈ 133–199 € netto). Ob fiskaly monatlich abrechnet, ist nicht
belegt. Die Zertifikatslaufzeit der Hardware verstreicht dagegen auch in den
Monaten ohne Fest.

Der Preis spricht also für die Hardware. Technisch wäre sie kein
Adapter-Austausch, sondern ein neues Subsystem:

- `TSEClient` verlangt zwar nur `StartTransaction` und `FinishTransaction`. Der
  Signatur-Worker braucht zusätzlich `RetrieveTransaction`
  (`backend/api/fiskal/signatur/tse_signatur_worker.go`), die Einrichtung
  `TestConnection` und `Umgebung`. `Credentials` ist fiskaly-geschnitten
  (`ApiKey`, `ApiSecret`, `TssID`, `ClientID`). Ein zweiter Adapter ist mehr als
  zwei Methoden.
- Ein cgo-Wrapper um die WORM-Bibliothek (`WormAPI.dll` bzw. `libWormAPI.so`)
  wäre Eigenbau, gebunden an eine nicht frei verteilbare Bibliothek.
- Das Backend läuft im Linux-Container (Service `backend` in
  `docker-compose.release.yml` und `docker-compose.local.yml`). Ein USB-Gerät am
  Windows-Host bräuchte zusätzlich eine Durchreichung in den Container.
- `windows/starter` kennt keinen Gerätezugriff; Vorarbeit gibt es dort nicht.

### Rechtlicher Rahmen

Die TSE-Pflicht knüpft an die Nutzung eines elektronischen Aufzeichnungssystems,
nicht an Umsatz oder Häufigkeit (§ 146a Abs. 1 AO). Die Ausnahmeliste in
§ 1 Abs. 1 KassenSichV nennt weder Vereine noch Feste. § 148 AO erlaubt
Erleichterungen für Einzelfälle oder Fallgruppen bei Härten. Nach dem AEAO zu
§ 146a Nr. 2.5.9 wird die Befreiung von der Belegausgabepflicht nur im Einzelfall
gewährt, und Kosten allein sind keine sachliche Härte. Das betrifft die
Belegausgabe, nicht die TSE-Pflicht. Wer keine TSE will, muss auf jedes
elektronische System verzichten und eine offene Ladenkasse führen. Der
[Referentenentwurf zu § 146b AO-E](https://www.bundesfinanzministerium.de/Content/DE/Gesetzestexte/Gesetze_Gesetzesvorhaben/Abteilungen/Abteilung_IV/21_Legislaturperiode/2026-08-07-G-Kassenpflicht/1-Referentenentwurf.pdf)
(Bearbeitungsstand 05.08.2026) greift erstmals für Zeiträume nach dem 31.12.2027
und ändert bis dahin nichts.

## Entscheidung

**Die Cloud-TSE bei fiskaly bleibt gesetzt. jotti baut keine Hardware-TSE-Anbindung
und integriert keinen zweiten Cloud-Anbieter.**

- **fiskaly bleibt Zielanbieter.** Öffentlich dokumentierte API, kostenlose
  Sandbox und BSI-Zertifizierung sind die Kriterien; kein geprüfter Anbieter
  erfüllt sie besser.
- **Keine Hardware-TSE — aus technischen Gründen, nicht aus Kostengründen.** Über
  fünf Jahre wäre sie billiger. Dagegen stehen: die Zertifikatslaufzeit beginnt
  mit der Fertigung und verfällt zwischen den Festen; die WORM-Bibliothek ist
  proprietär und hat keine Go-Bindung; ein cgo-Wrapper wäre Eigenbau; das Gerät
  müsste in den Linux-Container durchgereicht werden; und `windows/starter` hat
  dafür keine Vorarbeit. Der Aufwand ist nicht das Argument, das Ergebnis ist es:
  mehr bewegliche Teile für ehrenamtliche Teams, ohne Gegenwert.
- **Keine Zweitintegration ohne einsehbare Spezifikation.** Für Swissbit und
  D-Trust ist keine öffentliche API-Spezifikation auffindbar, bei Diebold Nixdorf
  ist sie ungeprüft. Ein Source-Available-Projekt kann nicht gegen eine
  Spezifikation entwickeln, die es nicht einsehen darf.
- **Kostenaussagen tragen Quelle und Datum.** „Nicht öffentlich" bleibt „nicht
  öffentlich"; jotti erfindet keine Bandbreite und nennt Reseller-Preise als
  Orientierung, nicht als Angebot von fiskaly.
- **Der TSE-Vertrag bleibt Sache des Vereins.** jotti vermittelt keinen Vertrag
  und verhandelt keine Rahmenkonditionen für seine Nutzer.

**Wieder aufgreifen, wenn beides zutrifft:** Ein zertifizierter Anbieter
dokumentiert seine API öffentlich und vertragsfrei, **und** er verkauft eine
öffentlich bepreiste Kurzzeit- oder monatlich kündbare Lizenz. Erst dann lohnt
ein zweiter Adapter hinter `TSEClient`. Veröffentlicht fiskaly allein seine
Konditionen, ist das eine FAQ-Aktualisierung.

## Konsequenzen

- Die Zahl steht nur in der FAQ (`docs/leitfaden/haeufige-fragen.md`):
  Größenordnung, Vertragsbindung und Datum der Recherche.
- `docs/leitfaden/tse-einrichten.md` verweist dafür auf die FAQ;
  `docs/compliance.md` §3.5 nennt keinen Betrag, nur die Entscheidung und
  belegte Anbieter-Beispiele.
- Veraltet die Zahl, ist genau eine Stelle zu ändern.
- `TSEClient` bleibt anbieter-agnostisch (Adapter-Pattern). Diese Entscheidung
  schließt einen Wechsel nicht aus, sie vertagt ihn.
- Ein Verein ohne eigenen TSE-Vertrag kann jotti nicht produktiv einsetzen. Das
  bleibt auf der Website und im README ausgeschrieben.
- Eine Hardware-TSE oder ein zweiter Cloud-Anbieter bräuchte ein neues ADR.
