# Lizenz, Nutzung & Rechtliches — jotti

Dieses Dokument regelt die rechtliche und organisatorische Grundlage für die Entwicklung und Nutzung von jotti.

> **Status:** Lebendes Dokument. Letzte Aktualisierung: 11. März 2026.

---

## Inhaltsverzeichnis

1. [Eigentumsverhältnisse](#1-eigentumsverhältnisse)
2. [Lizenzmodell](#2-lizenzmodell)
3. [Berechtigte Nutzer und kostenlose Nutzung](#3-berechtigte-nutzer-und-kostenlose-nutzung)
4. [Hosting und Betrieb](#4-hosting-und-betrieb)
5. [Datenschutz (DSGVO)](#5-datenschutz-dsgvo)
6. [Haftung und Gewährleistung](#6-haftung-und-gewährleistung)
7. [Kommerzialisierung und Dual Licensing](#7-kommerzialisierung-und-dual-licensing)
8. [Muster-Nutzungsvereinbarung](#8-muster-nutzungsvereinbarung)

**Anhang (Hinweise für den Entwickler):**

- [A. Sonderfall: Entwickler ist Vereinsmitglied](#a-sonderfall-entwickler-ist-vereinsmitglied)
- [B. Support und Erwartungsmanagement](#b-support-und-erwartungsmanagement)
- [C. Wirtschaftlicher Hintergrund](#c-wirtschaftlicher-hintergrund)

---

## 1. Eigentumsverhältnisse

Die Software **jotti** — einschließlich Quellcode, Dokumentation, Architektur, Design und aller zugehörigen Artefakte — ist das alleinige geistige Eigentum von **Nico Gräf**.

| Aspekt                     | Regelung                                                         |
| -------------------------- | ---------------------------------------------------------------- |
| Urheberrecht               | Nico Gräf, seit Projektbeginn 2025                               |
| IP (Intellectual Property) | Alle Rechte vorbehalten                                          |
| Repository                 | [github.com/nicograef/jotti](https://github.com/nicograef/jotti) |
| Lizenz                     | AGPL-3.0-or-later + Zusatzbedingungen (Source-Available, Non-Commercial) — siehe [Abschnitt 2](#2-lizenzmodell) |

- Nico Gräf entscheidet allein über Lizenzierung, Weiterentwicklung und Verbreitung.
- Kein Nutzer (Verein, Organisation, Person) erwirbt durch die Nutzung Rechte an der Software.
- Der Quellcode ist öffentlich einsehbar (Open Source), aber die Urheberschaft bleibt unberührt.
- Beiträge Dritter (Pull Requests) unterliegen der Projektlizenz; der Urheber behält das Recht auf Relizenzierung (siehe [Abschnitt 7](#7-kommerzialisierung-und-dual-licensing)).

---

## 2. Lizenzmodell

### Lizenz: AGPL-3.0-or-later mit Zusatzbedingungen (Source-Available, Non-Commercial)

jotti steht seit März 2026 unter der **AGPL-3.0-or-later**, ergänzt durch verbindliche Zusatzbedingungen (Additional Conditions) aus der `LICENSE`-Datei. Zuvor war das Projekt unter MIT lizenziert; der Wechsel wurde vom alleinigen Urheber (Nico Gräf) durchgeführt. Bereits unter MIT veröffentlichte Versionen bleiben unter MIT; alle neuen Versionen stehen unter diesem kombinierten Lizenzmodell.

> **Wichtiger Hinweis:** Aufgrund der Nicht-Kommerziell-Einschränkung (Zusatzbedingung 1) ist jotti **keine** standard-AGPL-3.0-Lizenz im OSI-Sinne und gilt nicht als „Open Source" nach der OSI-Definition. Es handelt sich um eine **source-available, nicht-kommerzielle, Copyleft-Lizenz**.

### Das vollständige Lizenzmodell: AGPL-3.0 + Zusatzbedingungen

jotti wird unter der **AGPL-3.0-or-later** veröffentlicht, ergänzt durch **verbindliche Zusatzbedingungen** (Additional Conditions), die in der `LICENSE`-Datei festgelegt sind. Die Zusatzbedingungen **beschränken** den Umfang der durch AGPL-3.0 gewährten Rechte — der Urheber räumt AGPL-3.0-Rechte nur **vorbehaltlich** dieser Einschränkungen ein. Die Zusatzbedingungen regeln insbesondere die **Nicht-Kommerziell-Einschränkung** und die **Share-Alike-Pflicht für Ableitungen**.

### Kernaussagen des Lizenzmodells

| Kriterium                                                                          | Regelung                                                                                                                                                                                       |
| ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Nutzung durch berechtigte Organisationen (self-hosted)                             | ✅ Kostenlos und uneingeschränkt                                                                                                                                                                |
| Jemand forkt jotti und betreibt es als **kostenloses**, nicht-kommerzielles SaaS   | ⚠️ Erlaubt, aber nur wenn: (1) keine Gewinnerzielungsabsicht, (2) alle Änderungen unter **denselben Lizenzbedingungen** (AGPL-3.0 + Zusatzbedingungen) veröffentlicht und betrieben werden.   |
| Jemand forkt jotti und betreibt es als **kostenpflichtiges oder gewerbliches** SaaS | ❌ Nicht erlaubt — auch nicht bei Offenlegung des Quellcodes unter AGPL. Erfordert zwingend eine separate **kommerzielle Lizenz** vom Urheber (Zusatzbedingung 1).                             |
| Schutz vor proprietärer Abspaltung                                                 | ✅ AGPL-Copyleft + Zusatzbedingungen — Derivate müssen unter denselben Bedingungen (inkl. Nicht-Kommerziell) veröffentlicht werden                                                             |
| Schutz vor kommerziellem Missbrauch trotz Quelloffenlegung                         | ✅ Nicht-Kommerziell-Einschränkung (Zusatzbedingung 1) — der Urheber räumt das AGPL-3.0-Nutzungsrecht nur für nicht-kommerzielle Zwecke ein                                                   |
| Dual Licensing (kommerziell + Source-Available) möglich                            | ✅ Nur der Urheber kann eine kommerzielle Lizenz vergeben                                                                                                                                      |
| SaaS-Hosting ohne Code-Offenlegung                                                 | ❌ Nicht AGPL-konform (§ 13 AGPL — „Remote Network Interaction")                                                                                                                               |
| Beiträge der Community fließen zurück                                              | ✅ Pflicht (Copyleft)                                                                                                                                                                          |

### Bedingungen für Forks und Ableitungen

Wer jotti forkt, modifiziert oder anderweitig ableitet, muss folgende Bedingungen einhalten (Zusatzbedingung 2 der `LICENSE`-Datei):

1. **Gleiche Lizenz (Share-Alike):** Die Ableitung muss vollständig unter denselben Lizenzbedingungen veröffentlicht werden — d.h. AGPL-3.0-or-later **und** den unveränderten Zusatzbedingungen aus der `LICENSE`-Datei.
2. **Keine Aufweichung der Nicht-Kommerziell-Einschränkung:** Die Ableitung darf nicht unter Bedingungen veröffentlicht oder betrieben werden, die gewerbliche Nutzung ohne kommerzielle Lizenz des Urhebers ermöglichen.
3. **Keine restriktivere oder permissivere Lizenz:** Weder eine proprietäre Schließung noch eine permissivere Relizenzierung (z.B. MIT oder Apache) sind zulässig.

**Kurz gesagt:** Wer jotti nutzen und daraus etwas bauen möchte, muss das Ergebnis ebenfalls kostenlos, nicht-kommerziell und quelloffen unter denselben Bedingungen betreiben. Eine kommerzielle Verwertung ist nur mit einer gesonderten Lizenz des Urhebers möglich.

### Auswirkungen auf berechtigte Organisationen

Das Lizenzmodell erlaubt die kostenlose Nutzung, Installation und Modifikation durch berechtigte (nicht-kommerzielle) Organisationen. Solange die Organisation die Software für eigene interne Zwecke betreibt (nicht als Dienst für Dritte oder mit Gewinnerzielungsabsicht anbietet), bestehen keinerlei Pflichten außer der Beibehaltung des Copyright-Hinweises.

---

## 3. Berechtigte Nutzer und kostenlose Nutzung

### Wer darf jotti kostenlos nutzen?

Die kostenlose Nutzungslizenz gilt für folgende **Personengruppen und Organisationen ohne Gewinnerzielungsabsicht**:

| Gruppe / Organisationsform                     | Voraussetzung                                                                           |
| ---------------------------------------------- | --------------------------------------------------------------------------------------- |
| **Eingetragene Vereine (e.V.)**                | Eintragung im Vereinsregister gemäß §§ 21 ff. BGB                                      |
| **Eingetragene gemeinnützige Stiftungen**      | Eintragung im Stiftungsregister, Anerkennung der Gemeinnützigkeit durch Finanzbehörden  |
| **Gemeinnützige GmbH (gGmbH)**                 | Eintragung im Handelsregister, steuerliche Anerkennung der Gemeinnützigkeit             |
| **Gemeinnützige UG (gUG)**                     | Eintragung im Handelsregister, steuerliche Anerkennung der Gemeinnützigkeit             |
| **Sonstige eingetragene NGOs / NPOs**          | Nachweisbare Eintragung in einem öffentlichen Register und fehlende Gewinnerzielungsabsicht |
| **Nicht-kommerzielle Open-Source-Projekte**    | Projekt ist öffentlich zugänglich, verfolgt keine Gewinnerzielungsabsicht, und das Ergebnis wird unter denselben Lizenzbedingungen veröffentlicht und betrieben (AGPL-3.0 + Zusatzbedingungen) |

Kommerzielle Betriebe, Unternehmen ohne gemeinnützigen Status sowie Einzelpersonen, die jotti gewerblich einsetzen möchten, fallen nicht unter die kostenlose Nutzungslizenz und benötigen eine [kommerzielle Lizenz](#7-kommerzialisierung-und-dual-licensing).

### Nutzungsvereinbarung (Kurzfassung)

Die Open-Source-Lizenz (AGPL) regelt die Software-Nutzung. Eine **separate Nutzungsvereinbarung** wird empfohlen, um folgende Punkte unmissverständlich zu klären:

| Punkt          | Regelung                                                   |
| -------------- | ---------------------------------------------------------- |
| IP/Eigentum    | Software ist Eigentum von Nico Gräf                        |
| Lizenztyp      | Unentgeltlich, nicht-exklusiv, widerruflich                |
| Nutzungsumfang | Interne Nutzung für den Kassenbetrieb der Organisation     |
| Gewährleistung | Keine — Software wird „as-is" bereitgestellt               |
| Support        | Freiwillig und unverbindlich                               |
| Datenschutz    | Verantwortlicher ist die nutzende Organisation             |
| Hosting        | Organisation verantwortet die Infrastruktur                |
| Kündigung      | Jederzeit von beiden Seiten, ohne Frist                    |

Die vollständige Muster-Nutzungsvereinbarung findet sich in [Abschnitt 8](#8-muster-nutzungsvereinbarung).

---

## 4. Hosting und Betrieb

### Grundsatz: Die Organisation ist Betreiberin

Die nutzende Organisation ist für Betrieb, Wartung und Sicherheit der Infrastruktur verantwortlich. Der Entwickler stellt keine Infrastruktur bereit.

| Aspekt            | Regelung                                                            |
| ----------------- | ------------------------------------------------------------------- |
| Server            | Verantwortung der Organisation (VPS, Cloud, Webhoster, lokaler Server) |
| Domain            | Von der Organisation registriert und bezahlt                        |
| SSL-Zertifikat    | Let's Encrypt (automatisch, kostenlos)                              |
| Datenbank-Backups | Verantwortung der Organisation                                      |
| Software-Updates  | Organisation entscheidet, wann aktualisiert wird                    |

### Hosting durch Dritte (Webhoster, Cloud, VPS)

In der Praxis wird die Infrastruktur häufig nicht von der Organisation selbst betrieben, sondern durch externe Anbieter (Webhoster, Cloud-Dienste, VPS-Anbieter) im Auftrag der Organisation bereitgestellt. Dies ist zulässig, sofern folgende Bedingungen erfüllt sind:

1. **Die Organisation bleibt Betreiberin und Verantwortliche.** Sie schließt den Vertrag mit dem Hosting-Anbieter und kontrolliert die Infrastruktur.
2. **Der Hosting-Anbieter ist Auftragsverarbeiter der Organisation** (Art. 28 DSGVO). Die Organisation schließt einen AV-Vertrag mit dem Hoster ab. Der Entwickler von jotti ist kein Vertragspartner dieses Verhältnisses.
3. **Der Entwickler hat keinen Zugriff auf den Server oder die Daten.** Damit entsteht kein Auftragsverarbeitungsverhältnis zwischen Entwickler und Organisation.
4. **Kein Multi-Tenant-Hosting.** Bietet ein Dritter jotti als gemeinsamen Dienst für mehrere Organisationen an (SaaS), unterliegt er der AGPL-3.0-Offenlegungspflicht oder benötigt eine kommerzielle Lizenz.

### Einrichtungshilfe

Der Entwickler kann bei der Ersteinrichtung freiwillig unterstützen (Bereitstellen von Docker Compose und Dokumentation, einmalige Begleitung des Setups). Daraus entsteht kein Anspruch auf laufende Betreuung oder Betrieb.

---

## 5. Datenschutz (DSGVO)

### Welche Daten speichert jotti?

| Datenkategorie                   | Personenbezug                   | Beispiel                                      |
| -------------------------------- | ------------------------------- | --------------------------------------------- |
| Benutzerdaten (Servicekräfte)    | Ja                              | Name, Rolle, Passwort-Hash                    |
| Bestellungen & Events            | Indirekt (User-ID referenziert) | Bestellung auf Tisch 5, aufgegeben von User 3 |
| Gästedaten                       | **Nein**                        | Gäste haben keine Accounts                    |
| Zahlungsdaten (Kreditkarte etc.) | **Nein**                        | jotti verarbeitet keine Zahlungsmittel        |

### Rollen im Datenschutz

| Rolle (DSGVO)                       | Wer                                          | Begründung                                                                         |
| ----------------------------------- | -------------------------------------------- | ---------------------------------------------------------------------------------- |
| **Verantwortlicher** (Art. 4 Nr. 7) | Die nutzende Organisation                    | Die Organisation entscheidet, welche Servicekräfte als Benutzer angelegt werden    |
| **Auftragsverarbeiter** (Art. 28)   | Ggf. der Hosting-Anbieter der Organisation   | Sofern ein externer Hoster die Infrastruktur betreibt (AV-Vertrag erforderlich)    |
| **Kein Auftragsverarbeiter**        | Der Entwickler (Nico Gräf)                   | Der Entwickler hat keinen Zugriff auf den Server oder die gespeicherten Daten      |
| **Betroffene Personen**             | Servicekräfte der Organisation               | Ihre Daten werden im System gespeichert                                            |

### Pflichten der Organisation

| Pflicht                                | Beschreibung                                                                                                                        |
| -------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| **Informationspflicht** (Art. 13/14)   | Servicekräfte darüber informieren, dass ihre Daten gespeichert werden (Name, Rolle, Aktivitäten)                                    |
| **Löschpflicht** (Art. 17)             | Benutzer auf Wunsch deaktivieren/löschen (Soft-Delete vorhanden)                                                                    |
| **Verarbeitungsverzeichnis** (Art. 30) | jotti als Verarbeitungstätigkeit dokumentieren (bei Organisationen unter 250 Beschäftigten i.d.R. nicht Pflicht, aber empfehlenswert) |
| **Technische Maßnahmen** (Art. 32)     | Server absichern, HTTPS verwenden, Passwörter nicht teilen                                                                          |
| **AV-Vertrag** (Art. 28)               | Mit dem Hosting-Anbieter schließen, sofern dieser die Infrastruktur betreibt                                                        |

---

## 6. Haftung und Gewährleistung

Die AGPL-3.0-Lizenz enthält einen umfassenden Haftungsausschluss:

> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND.

Die Nutzungsvereinbarung bekräftigt:

- Keine Gewährleistung auf Funktionalität, Richtigkeit oder Verfügbarkeit.
- Kein Anspruch auf Fehlerbehebung.
- Die Organisation nutzt die Software auf eigenes Risiko.
- Der Entwickler haftet nicht für Schäden aus der Nutzung der Software — insbesondere nicht für Datenverlust, Fehlberechnungen oder Ausfälle.

---

## 7. Kommerzialisierung und Dual Licensing

Als alleiniger Urheber besitzt Nico Gräf das Recht, jotti unter beliebig vielen Lizenzen gleichzeitig zu veröffentlichen (**Dual-Licensing-Modell**):

| Pfad                        | Lizenz                              | Zielgruppe                                                                   | Kosten          |
| --------------------------- | ----------------------------------- | ---------------------------------------------------------------------------- | --------------- |
| **Non-Profit / Open Source** | AGPL-3.0 + Zusatzbedingungen        | Eingetragene Vereine (e.V.), gemeinnützige Stiftungen, NGOs/NPOs; nicht-kommerzielle Open-Source-Projekte (unter denselben Bedingungen) | Kostenlos |
| **Kommerziell**             | Proprietäre / kommerzielle Lizenz   | Unternehmen, gewerbliche SaaS-Anbieter, kommerzielle Betriebe                | Kostenpflichtig |

### Vorbehaltene Rechte

Nico Gräf behält sich ausdrücklich folgende Rechte vor:

1. **Kostenpflichtige Nutzungslizenzen anbieten** — z.B. für gewerbliche Betriebe, die jotti ohne AGPL-Pflichten nutzen möchten.
2. **jotti als kostenpflichtiges SaaS-Produkt betreiben** — als gehostete Lösung mit Support, SLA und Wartung.
3. **Organisationen von der kostenlosen Nutzung auszunehmen** — z.B. wenn eine Organisation die Software gewerblich weitervermarktet oder die Anspruchsvoraussetzungen (Eintragung, Gemeinnützigkeit) nicht erfüllt.
4. **Die Lizenz zukünftiger Versionen zu ändern** — solange bestehende AGPL-veröffentlichte Versionen unter AGPL bleiben.

### Warum das Modell funktioniert

- **AGPL-Copyleft** verlangt, dass jeder, der jotti als Netzwerkservice (SaaS) anbietet, den vollständigen Quellcode aller Änderungen unter AGPL veröffentlichen muss.
- **Nicht-Kommerziell-Einschränkung (Zusatzbedingung 1):** Selbst wer den Quellcode unter AGPL offenlegt, darf jotti oder eine Ableitung davon **nicht kommerziell** betreiben oder vermarkten — auch nicht als kostenpflichtiges SaaS mit offenem Quellcode. Die Offenlegung des Quellcodes ist eine notwendige, aber keine hinreichende Bedingung für erlaubte Nutzung.
- **Share-Alike (Zusatzbedingung 2):** Wer jotti forkt und daraus ein Open-Source-Projekt macht, muss das Ergebnis unter denselben Bedingungen (AGPL-3.0 + Nicht-Kommerziell + Share-Alike) veröffentlichen. Damit wird verhindert, dass die Nutzungsbeschränkung durch eine Kette von Ableitungen aufgeweicht wird.
- **Einzige legale Möglichkeit zur kommerziellen Nutzung:** Eine separate kommerzielle Lizenz vom Urheber erwerben.
- **Berechtigte Organisationen, die selbst oder über einen Drittanbieter hosten, sind nicht betroffen.** Interne Nutzung ohne Gewinnerzielungsabsicht löst weder die Copyleft-Pflicht noch die Offenlegungspflicht aus.

### Mögliche Kommerzialisierungsmodelle

| Modell                   | Beschreibung                                                                                            |
| ------------------------ | ------------------------------------------------------------------------------------------------------- |
| **Hosting-as-a-Service** | jotti als gehostete Lösung für Organisationen (monatliches Abo). Dann: Auftragsverarbeiter → AV-Vertrag nötig. |
| **Setup-Pakete**         | Einmalige Einrichtung + Einweisung gegen Festpreis.                                                     |
| **Support-Verträge**     | Garantierte Reaktionszeiten, Hotline am Festtag.                                                        |
| **Enterprise-Lizenz**    | Gewerbliche Betriebe zahlen eine Lizenzgebühr für die Nutzung ohne AGPL- und Nicht-Kommerziell-Beschränkungen. |
| **White-Label-Lizenz**   | Dritte dürfen jotti unter eigenem Namen anbieten — nur mit kommerzieller Lizenz.                        |

### Community-Beiträge

Beiträge Dritter (Pull Requests) stehen unter AGPL. Um Dual Licensing aufrechtzuerhalten, empfiehlt sich ein **Contributor License Agreement (CLA)**, das dem Urheber das Recht zur Relizenzierung einräumt.

---

## 8. Muster-Nutzungsvereinbarung

Die folgende Vereinbarung kann als Vorlage verwendet und angepasst werden. Sie ist **kein Rechtsberatungsprodukt** und ersetzt keine anwaltliche Prüfung.

---

### Nutzungsvereinbarung für die Software „jotti"

**Zwischen:**

- **Entwickler:** Nico Gräf (nachfolgend „Entwickler")
- **Nutzer:** [ORGANISATIONSNAME] (nachfolgend „Organisation")

**Stand:** \_\_\_\_\_\_\_\_\_\_\_\_\_

---

#### § 1 Gegenstand

(1) Der Entwickler stellt der Organisation die Software „jotti" (nachfolgend „Software") zur unentgeltlichen Nutzung zur Verfügung.

(2) Die Software ist ein persönliches Open-Source-Projekt des Entwicklers. Die Organisation hat die Entwicklung weder beauftragt noch finanziert.

(3) Der Quellcode ist öffentlich zugänglich unter: https://github.com/nicograef/jotti

#### § 2 Berechtigung zur unentgeltlichen Nutzung

(1) Die unentgeltliche Nutzung steht ausschließlich offiziell eingetragenen und nicht-gewinnorientierten Organisationen offen, insbesondere:

- eingetragenen Vereinen (e.V.) gemäß §§ 21 ff. BGB,
- eingetragenen gemeinnützigen Stiftungen,
- gemeinnützigen GmbH (gGmbH) und gemeinnützigen UG (gUG),
- sonstigen eingetragenen gemeinnützigen Körperschaften ohne Gewinnerzielungsabsicht.

(2) Die Organisation sichert zu, die Voraussetzungen nach Abs. 1 zu erfüllen, und verpflichtet sich, den Entwickler unverzüglich zu informieren, falls diese Voraussetzungen entfallen.

(3) Gewerbliche oder kommerzielle Nutzung ist ohne gesonderte kostenpflichtige Lizenz nicht gestattet.

#### § 3 Geistiges Eigentum

(1) Die Software — einschließlich Quellcode, Dokumentation, Architektur und Design — ist das alleinige geistige Eigentum des Entwicklers.

(2) Durch die Nutzung der Software erwirbt die Organisation keinerlei Rechte an der Software über die in § 4 gewährte Lizenz hinaus.

(3) Der Entwickler behält sich das Recht vor, die Software jederzeit unter weiteren Lizenzen zu veröffentlichen, zu vermarkten oder anderweitig zu verwerten.

#### § 4 Nutzungslizenz

(1) Die Organisation erhält eine **unentgeltliche, nicht-exklusive, widerrufliche** Lizenz zur Nutzung der Software für den organisationsinternen Kassenbetrieb.

(2) Die Nutzung umfasst: Installation, Konfiguration, Betrieb und Aktualisierung auf von der Organisation verantworteter Infrastruktur.

(3) Es gelten zusätzlich die Bestimmungen der Open-Source-Lizenz (AGPL-3.0-or-later), unter der die Software veröffentlicht ist.

#### § 5 Hosting und Betrieb

(1) Die Organisation ist Betreiberin der Software und verantwortet die hierfür eingesetzte Infrastruktur (Server, Datenbank, Netzwerk).

(2) Der Entwickler stellt weder Hosting noch Infrastruktur bereit.

(3) Die Organisation kann die Infrastruktur durch einen externen Hosting-Anbieter (Webhoster, Cloud-Dienst, VPS-Anbieter) betreiben lassen. In diesem Fall ist die Organisation für den Abschluss eines Auftragsverarbeitungsvertrags (Art. 28 DSGVO) mit dem Hoster verantwortlich.

(4) Ein Multi-Tenant-Betrieb, bei dem ein Dritter jotti als gemeinsamen Dienst für mehrere Organisationen anbietet, ist ohne eine separate kommerzielle Lizenz des Entwicklers nicht gestattet.

(5) Der Entwickler kann freiwillig bei der Ersteinrichtung unterstützen. Daraus entsteht kein Anspruch auf laufende Betreuung.

#### § 6 Datenschutz

(1) Die Organisation ist **Verantwortliche** im Sinne der DSGVO (Art. 4 Nr. 7) für alle in der Software verarbeiteten personenbezogenen Daten.

(2) Der Entwickler ist **kein Auftragsverarbeiter**, da er keinen Zugriff auf den Server oder die gespeicherten Daten hat.

(3) Die Organisation ist verpflichtet, die einschlägigen Datenschutzbestimmungen einzuhalten, insbesondere die Betroffenen (Servicekräfte) über die Datenverarbeitung zu informieren und mit einem etwaigen Hosting-Anbieter einen AV-Vertrag abzuschließen.

#### § 7 Gewährleistung und Haftung

(1) Die Software wird **ohne Gewährleistung** bereitgestellt („as-is"). Es besteht kein Anspruch auf Fehlerfreiheit, Funktionalität, Verfügbarkeit oder Eignung für einen bestimmten Zweck.

(2) Der Entwickler haftet nicht für Schäden, die aus der Nutzung der Software entstehen — insbesondere nicht für Datenverlust, Fehlberechnungen oder Ausfälle.

(3) Die Organisation nutzt die Software auf eigenes Risiko.

#### § 8 Support und Weiterentwicklung

(1) Es besteht **kein Anspruch** auf Support, Fehlerbehebung, Weiterentwicklung oder Aktualisierung.

(2) Unterstützung durch den Entwickler erfolgt freiwillig und unverbindlich.

(3) Feature-Wünsche der Organisation werden zur Kenntnis genommen, begründen jedoch keinen Umsetzungsanspruch.

#### § 9 Laufzeit und Kündigung

(1) Diese Vereinbarung gilt auf unbestimmte Zeit.

(2) Beide Seiten können die Vereinbarung jederzeit ohne Angabe von Gründen und ohne Einhaltung einer Frist beenden.

(3) Bei Beendigung erlischt die Nutzungslizenz nach § 4 Abs. 1. Die Rechte aus der Open-Source-Lizenz (AGPL-3.0) bleiben unberührt.

#### § 10 Schlussbestimmungen

(1) Änderungen und Ergänzungen dieser Vereinbarung bedürfen der Schriftform.

(2) Sollten einzelne Bestimmungen unwirksam sein, bleibt die Wirksamkeit der übrigen Bestimmungen unberührt.

---

**Ort, Datum:** \_\_\_\_\_\_\_\_\_\_\_\_\_

**Entwickler:** \_\_\_\_\_\_\_\_\_\_\_\_\_ (Nico Gräf)

**Organisation:** \_\_\_\_\_\_\_\_\_\_\_\_\_ (vertretungsberechtigte Person, [ORGANISATIONSNAME])

---

> **Hinweis:** Diese Vereinbarung ist ein Muster und keine Rechtsberatung. Bei Unsicherheiten empfiehlt sich die Prüfung durch eine Rechtsanwältin oder einen Rechtsanwalt.

---

## Anhang: Hinweise für den Entwickler

---

### A. Sonderfall: Entwickler ist Vereinsmitglied

Wenn der Entwickler gleichzeitig Mitglied des nutzenden Vereins ist, entsteht ein potenzieller Graubereich: Könnte der Verein argumentieren, die Software sei im Rahmen der Vereinstätigkeit entstanden?

Im deutschen Recht gilt: Vereinsmitgliedschaft begründet kein Arbeitsverhältnis (§ 7 UrhG). Das Urheberrecht entsteht beim Schöpfer. § 69b UrhG (Arbeitnehmerurheberrecht) gilt nicht für ehrenamtliche Vereinstätigkeit.

**Faktenlage bei jotti:** Entwickelt auf eigenem Rechner, in eigener Zeit, mit eigenen Werkzeugen. Öffentliches Open-Source-Projekt auf GitHub unter Nico Gräfs persönlichem Account. Kein Vereinsbeschluss zur Beauftragung.

Die Nutzungsvereinbarung (§ 1 Abs. 2) hält explizit fest, dass die Entwicklung unabhängig von der Vereinsmitgliedschaft erfolgt und der Verein weder beauftragt noch finanziert hat.

| Situation                                  | Empfehlung                                                                                                  |
| ------------------------------------------ | ----------------------------------------------------------------------------------------------------------- |
| Verein wünscht sich Feature X              | Prüfen, ob es ins Backlog passt. Kein Versprechen.                                                          |
| Verein sagt „Wir brauchen das bis Samstag" | Klar kommunizieren: „Ich bin kein Auftragnehmer."                                                           |
| Verein bietet Geld für ein Feature         | **Vorsicht** — dann entsteht ein Auftragsverhältnis. Entweder ablehnen oder als Werkvertrag aufsetzen.      |

---

### B. Support und Erwartungsmanagement

> **Du bist Entwickler eines Open-Source-Projekts, das die Organisation nutzen darf — nicht ihr IT-Dienstleister.**

| Situation                                     | Empfohlene Reaktion                                                                       |
| --------------------------------------------- | ----------------------------------------------------------------------------------------- |
| „Das muss bis Samstag funktionieren"          | „Ich gebe mein Bestes, aber verspreche nichts. Testet vorher."                            |
| „Kannst du noch schnell X einbauen?"          | „Steht auf der Wunschliste. Kommt, wenn es kommt."                                        |
| „Das ist kaputt, reparier das sofort"         | „Ich schau es mir an, wenn ich Zeit habe. Für heute: Stift & Papier als Fallback."        |
| Neuer Vorstand kennt die Absprache nicht      | Nutzungsvereinbarung schriftlich beim Verein hinterlegen.                                 |

Praktische Empfehlungen: Nutzungsvereinbarung vor dem ersten Einsatz unterschreiben lassen; eine feste Ansprechperson im Verein benennen; Testlauf vor dem echten Einsatz durchführen; Fallback-Plan (Stift & Papier) definieren.

Für externe Organisationen können optional Dienstleistungen gegen Bezahlung angeboten werden (Einrichtungshilfe, Einweisung, individuelle Anpassungen, Hosting-as-a-Service). Bei entgeltlichem Hosting gilt: Auftragsverarbeiter → AV-Vertrag erforderlich; ggf. Gewerbe anmelden.

---

### C. Wirtschaftlicher Hintergrund

Um den Wert der unentgeltlich bereitgestellten Software einzuordnen, wurden drei Budgetszenarien berechnet (42 Anforderungen, Neuentwicklung von Grund auf):

| Szenario                                    | Personentage | Budget (netto) | Laufzeit |
| ------------------------------------------- | ------------ | -------------- | -------- |
| **Software-Agentur** (konservativ, 3,5 FTE) | 274 PT       | ~250.000 €     | 9 Monate |
| **Software-Agentur** (optimiert, 2,6 FTE)   | 160 PT       | ~145.000 €     | 7 Monate |
| **Senior Freelancer** (Solo, 900 €/Tag)     | 133 PT       | ~122.000 €     | 8 Monate |

Grundlage: marktübliche Tagessätze (750–1.050 €/PT), vollständige Umsetzung aller Must-have- und Nice-to-have-Anforderungen. Diese Zahlen dienen der Dokumentation des unentgeltlichen Charakters der Bereitstellung sowie dem Erwartungsmanagement gegenüber nutzenden Organisationen.
