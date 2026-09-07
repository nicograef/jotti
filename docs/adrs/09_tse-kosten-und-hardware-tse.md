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
   36 Monate ([Swissbit](https://www.swissbit.com/en/products/security-products/swissbit-tse/cloud-tse/),
   [acventis](https://acventis.com/SWISSBIT-Cloud-TSE-2-TSE-Cloud-Services-Typ-S-3-Jahre)).
   Eine öffentliche OpenAPI-Spezifikation der beworbenen REST-API war nicht
   auffindbar.
2. **Deutsche Fiskal / D-Trust** — verlangt einen lokalen „Fiskal Cloud
   Connector" statt reiner HTTPS-Aufrufe. Wiederverkäufer nennen 12,00 € bzw.
   14,99 €/Monat netto je Kasse bei zwölf Monaten Laufzeit
   ([JTL](https://www.jtl-software.de/jtl-store/erweiterungen/fiskal-cloud-tse),
   [S&S ITS](https://sundsits.com/en/fiskal-cloud-tse-je-kasse/)). Ein
   Herstellerpreis ist nicht öffentlich.
3. **Diebold Nixdorf und EFSTA** — Preise und Konditionen nur auf Anfrage, keine
   öffentliche Schnittstellendokumentation
   ([DN](https://www.dieboldnixdorf.com/de-de/retail/portfolio/systems/tse/),
   [EFSTA](https://www.efsta.eu/en/solutions/tse-germany)).

Kein Anbieter verkauft öffentlich eine Lizenz für ein einzelnes Wochenende.

### Hardware-TSE am Windows-Starter

| Punkt    | Befund                                                                           | Quelle                                                                                                                               |
| -------- | -------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| Preis    | Swissbit USB-TSE ab 184,95 € brutto, TSE v2 (microSD/SD/USB) ab 194,95 € brutto  | [gebongt24](https://www.gebongt24.de/kassenzubehoer/tse/)                                                                            |
| Laufzeit | 5 Jahre oder 20 Mio. Signaturen, beginnend mit der Fertigung, nicht mit dem Kauf | [HKS Systeme](https://www.hks-systeme.de/ablauf-der-swissbit-tse-zertifikate/)                                                       |
| SDK      | `WormAPI.dll`, Weitergabe vertraglich beschränkt; keine offizielle Go-Bindung    | [python-tse](https://github.com/bwurst/python-tse), [cryptovision](https://www.cryptovision.com/en/products/secure-elements-sw/tse/) |

Über fünf Jahre liegt die Hardware damit in derselben Größenordnung wie das
Cloud-Abo. Ihre Laufzeit verstreicht auch in den Monaten ohne Fest.

Technisch wäre es kein Adapter-Austausch, sondern ein neues Subsystem:

- `TSEClient` verlangt nur `StartTransaction` und `FinishTransaction`
  (`backend/domain/tse/client.go`) — die Schnittstelle selbst ist kein Hindernis.
- Ein cgo-Wrapper um `WormAPI.dll` wäre Eigenbau, gebunden an Windows und an eine
  nicht frei verteilbare Bibliothek.
- Das Backend läuft im Linux-Container (`docker-compose.yml`); ein USB-Gerät am
  Windows-Host bräuchte zusätzlich eine Durchreichung in den Container.
- `windows/starter` fährt heute nur den Compose-Stack hoch und kennt keinen
  Gerätezugriff. Vorarbeit gibt es dort nicht.

### Rechtlicher Rahmen

Die TSE-Pflicht knüpft an die Nutzung eines elektronischen Aufzeichnungssystems,
nicht an Umsatz oder Häufigkeit (§ 146a Abs. 1 AO). Die Ausnahmeliste in
§ 1 Abs. 1 KassenSichV nennt weder Vereine noch Feste. § 148 AO erlaubt
Erleichterungen nur im Einzelfall; nach dem AEAO zu § 146a Nr. 2.5.9 sind die
Kosten für sich allein keine sachliche Härte, und die Vorschrift betrifft dort die
Belegausgabe, nicht die TSE. Wer keine TSE will, muss auf jedes elektronische
System verzichten und eine offene Ladenkasse führen. Der
[Referentenentwurf zu § 146b AO-E](https://www.bundesfinanzministerium.de/Content/DE/Gesetzestexte/Gesetze_Gesetzesvorhaben/Abteilungen/Abteilung_IV/21_Legislaturperiode/2026-08-07-G-Kassenpflicht/1-Referentenentwurf.pdf)
(Bearbeitungsstand 05.08.2026) greift erstmals für Zeiträume nach dem 31.12.2027
und ändert bis dahin nichts.

## Entscheidung

**Die Cloud-TSE bei fiskaly bleibt gesetzt. jotti baut keine Hardware-TSE-Anbindung
und integriert keinen zweiten Cloud-Anbieter.**

- **fiskaly bleibt Zielanbieter.** Öffentlich dokumentierte API, kostenlose
  Sandbox und BSI-Zertifizierung sind die Kriterien; kein geprüfter Anbieter
  erfüllt sie besser.
- **Keine Hardware-TSE.** Proprietäre Bibliothek ohne Go-Bindung, Eigenbau per
  cgo, Gerätedurchreichung in den Container, Laufzeit ab Fertigung — und kein
  Kostenvorteil über fünf Jahre. Der Aufwand ist nicht das Argument, das Ergebnis
  ist es: mehr bewegliche Teile für ehrenamtliche Teams, ohne Gegenwert.
- **Keine Zweitintegration ohne öffentliche API.** Swissbit, D-Trust, Diebold
  Nixdorf und EFSTA gelten als „ohne Vertrag nicht integrierbar". Ein
  Source-Available-Projekt kann nicht gegen eine Spezifikation entwickeln, die
  es nicht einsehen darf.
- **Kostenaussagen tragen Quelle und Datum.** „Nicht öffentlich" bleibt „nicht
  öffentlich"; jotti erfindet keine Bandbreite und nennt Reseller-Preise als
  Orientierung, nicht als Angebot von fiskaly.
- **Der TSE-Vertrag bleibt Sache des Vereins.** jotti vermittelt keinen Vertrag
  und verhandelt keine Rahmenkonditionen für seine Nutzer.

**Wieder aufgreifen, wenn** ein zertifizierter Anbieter eine öffentlich bepreiste
Kurzzeit- oder monatlich kündbare Lizenz anbietet, oder wenn er seine API ohne
Vertrag öffentlich dokumentiert. Dann lohnt ein zweiter Adapter hinter
`TSEClient`.

## Konsequenzen

- Die FAQ nennt die Größenordnung der TSE-Kosten, die Vertragsbindung, soweit
  belegbar, und das Datum der Recherche.
- Der Kostenabsatz in `docs/leitfaden/tse-einrichten.md` und `docs/compliance.md`
  §3.5 tragen dieselbe Aussage; §3.5 führt nur belegte Anbieter-Beispiele.
- Die Zahlen altern. Wer eine ändert, zieht alle drei Stellen im selben Schritt
  nach.
- `TSEClient` bleibt anbieter-agnostisch (Adapter-Pattern). Diese Entscheidung
  schließt einen Wechsel nicht aus, sie vertagt ihn.
- Ein Verein ohne eigenen TSE-Vertrag kann jotti nicht produktiv einsetzen. Das
  bleibt auf der Website und im README ausgeschrieben.
- Eine Hardware-TSE oder ein zweiter Cloud-Anbieter bräuchte ein neues ADR.
