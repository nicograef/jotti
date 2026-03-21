# Lizenz, Nutzung & Rechtliches — jotti

Dieses Dokument regelt die rechtliche und organisatorische Grundlage für die Entwicklung und Nutzung von jotti. Es ist die verbindliche Referenz für alle Fragen zu Eigentum, Lizenzierung, Nutzungsbedingungen, Haftung und Compliance.

> **Status:** Lebendes Dokument. Letzte Aktualisierung: 21. März 2026.
>
> **Hinweis:** Dieses Dokument ist keine Rechtsberatung. Bei Unsicherheiten empfiehlt sich die Prüfung durch eine Rechtsanwältin oder einen Rechtsanwalt. Die verbindliche Lizenz ist die `LICENSE`-Datei im Repository.

---

## Inhaltsverzeichnis

1. [Eigentumsverhältnisse](#1-eigentumsverhältnisse)
2. [Lizenzmodell](#2-lizenzmodell-proprietäre-source-available-lizenz)
3. [Pflicht zur Nutzungsvereinbarung](#3-pflicht-zur-nutzungsvereinbarung)
4. [Berechtigte Nutzer und kostenlose Nutzung](#4-berechtigte-nutzer-und-kostenlose-nutzung)
5. [Forks, Modifikation und Weitergabe](#5-forks-modifikation-und-weitergabe)
6. [Hosting und Betrieb](#6-hosting-und-betrieb)
7. [Datenschutz (DSGVO)](#7-datenschutz-dsgvo)
8. [Haftung und Gewährleistung](#8-haftung-und-gewährleistung)
9. [Freistellungsklausel](#9-freistellungsklausel)
10. [Compliance-Verantwortung (KassenSichV / TSE)](#10-compliance-verantwortung-kassensichv--tse)
11. [Kommerzialisierung und Dual Licensing](#11-kommerzialisierung-und-dual-licensing)
12. [Community-Beiträge und CLA](#12-community-beiträge-und-cla)
13. [Muster-Nutzungsvereinbarung](#13-muster-nutzungsvereinbarung)

**Anhang:**

- [A. Sonderfall: Entwickler ist Vereinsmitglied](#a-sonderfall-entwickler-ist-vereinsmitglied)
- [B. Support und Erwartungsmanagement](#b-support-und-erwartungsmanagement)
- [C. Wirtschaftlicher Hintergrund](#c-wirtschaftlicher-hintergrund)

---

## 1. Eigentumsverhältnisse

Die Software **jotti** — einschließlich Quellcode, Dokumentation, Architektur, Design und aller zugehörigen Artefakte — ist das **alleinige geistige Eigentum** von **Nico Gräf** (Freiburg im Breisgau, Deutschland).

| Aspekt                     | Regelung                                                                                                                     |
| -------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Urheberrecht               | Nico Gräf, seit Projektbeginn 2025                                                                                           |
| IP (Intellectual Property) | Alle Rechte vorbehalten                                                                                                      |
| Repository                 | [github.com/nicograef/jotti](https://github.com/nicograef/jotti)                                                             |
| Lizenz                     | Proprietäre Source-Available-Lizenz — siehe `LICENSE` und [Abschnitt 2](#2-lizenzmodell-proprietäre-source-available-lizenz) |

- Nico Gräf entscheidet **allein** über Lizenzierung, Weiterentwicklung und Verbreitung.
- **Kein Nutzer** (Verein, Organisation, Person) erwirbt durch die Nutzung Rechte an der Software über die in der Nutzungsvereinbarung gewährte Lizenz hinaus.
- Der Quellcode ist öffentlich einsehbar (**Source-Available**), aber die Urheberschaft und alle Rechte bleiben beim Autor.
- Die Veröffentlichung des Quellcodes stellt **keine** automatische Rechteeinräumung dar.
- Beiträge Dritter (Pull Requests) unterliegen dem Contributor License Agreement (CLA); der Autor behält alle Rechte, einschließlich des Rechts auf Relizenzierung.

---

## 2. Lizenzmodell: Proprietäre Source-Available-Lizenz

### Grundprinzip

jotti steht seit dem 21. März 2026 unter einer **proprietären Source-Available-Lizenz**. Der Quellcode ist öffentlich einsehbar, aber **es werden keine Nutzungsrechte automatisch gewährt**. Jede Nutzung erfordert eine vorherige schriftliche Nutzungsvereinbarung mit dem Autor.

### Was bedeutet „Source-Available"?

| Eigenschaft                       | Source-Available (jotti) | Open Source (z.B. MIT, Apache) |
| --------------------------------- | ------------------------ | ------------------------------ |
| Quellcode einsehbar               | ✅ Ja                    | ✅ Ja                          |
| Nutzung automatisch erlaubt       | ❌ Nein                  | ✅ Ja                          |
| Modifikation erlaubt              | ❌ Nein                  | ✅ Ja                          |
| Weitergabe erlaubt                | ❌ Nein                  | ✅ Ja                          |
| Kommerzielle Nutzung              | ❌ Verboten              | ✅ / ⚠️ Je nach Lizenz         |
| Nutzungsvereinbarung erforderlich | ✅ Ja                    | ❌ Nein                        |
| OSI-zertifiziert                  | ❌ Nein                  | ✅ Ja                          |

> **Klarstellung:** jotti ist **kein Open-Source-Projekt** im Sinne der OSI-Definition. Es ist ein proprietäres Software-Projekt mit öffentlich einsehbarem Quellcode. Die Veröffentlichung dient der Transparenz, dem Vertrauen und der Sicherheitsüberprüfung — nicht der freien Nutzung.

### Historischer Lizenzwechsel

| Zeitraum                  | Lizenz                                                                       |
| ------------------------- | ---------------------------------------------------------------------------- |
| Projektbeginn – März 2026 | MIT (permissive Open Source)                                                 |
| März 2026 – 21. März 2026 | AGPL-3.0-or-later + Additional Conditions (Source-Available, Non-Commercial) |
| Ab 21. März 2026          | **Proprietäre Source-Available-Lizenz** (alle Rechte vorbehalten)            |

Bereits veröffentlichte Versionen behalten ihre ursprünglichen Lizenzen. Alle neuen Versionen ab dem 21. März 2026 unterliegen ausschließlich der proprietären Source-Available-Lizenz.

### Was ist ohne Nutzungsvereinbarung erlaubt?

Ausschließlich **zwei Aktivitäten** sind ohne Nutzungsvereinbarung gestattet:

1. **Ansehen und Lesen** des Quellcodes zu persönlichen Bildungszwecken, technischer Evaluation oder Sicherheitsüberprüfung.
2. **Einreichen von Beiträgen** (Pull Requests) an das offizielle Repository des Autors, unter den Bedingungen des Contributor License Agreement (CLA).

**Alles andere** — insbesondere Installation, Deployment, Hosting, Ausführung, Kopieren, Modifizieren und Weitergabe — erfordert eine vorherige schriftliche Nutzungsvereinbarung.

### Kernaussagen des Lizenzmodells

| Szenario                                      | Erlaubt?                                   |
| --------------------------------------------- | ------------------------------------------ |
| Quellcode auf GitHub lesen                    | ✅ Ja, ohne Vereinbarung                   |
| Pull Request einreichen (mit CLA)             | ✅ Ja, ohne Vereinbarung                   |
| Nutzung durch e.V. / gGmbH (mit Vereinbarung) | ✅ Ja, kostenlos                           |
| Nutzung ohne Vereinbarung                     | ❌ Nein — unter keinen Umständen           |
| Nutzung durch gewerbliches Unternehmen        | ❌ Nein — nur mit kommerzieller Lizenz     |
| Fork für eigene Nutzung                       | ❌ Nein                                    |
| Weiterverbreitung des Codes                   | ❌ Nein                                    |
| Modifikation des Codes                        | ❌ Nein (außer PRs an das offizielle Repo) |
| SaaS-Betrieb durch Dritte                     | ❌ Nein — auch nicht kostenlos             |
| White-Label / Rebranding                      | ❌ Nein — nur mit kommerzieller Lizenz     |

---

## 3. Pflicht zur Nutzungsvereinbarung

### Grundsatz

**Jede Organisation, die jotti nutzen möchte, muss vor der ersten Nutzung eine schriftliche Nutzungsvereinbarung mit dem Autor (Nico Gräf) abschließen.**

Es gibt keine Ausnahmen. Weder die Gemeinnützigkeit noch die Unentgeltlichkeit noch die Tatsache, dass der Quellcode öffentlich ist, befreit von der Pflicht zur Nutzungsvereinbarung.

### Ablauf

1. **Anfrage:** Die Organisation kontaktiert den Autor per E-Mail oder über das GitHub-Repository und benennt sich (Name, Rechtsform, Registernummer, Ansprechperson).
2. **Vereinbarung:** Der Autor stellt die Muster-Nutzungsvereinbarung (siehe [Abschnitt 13](#13-muster-nutzungsvereinbarung)) zur Verfügung.
3. **Abschluss:** Die Organisation füllt die Vereinbarung aus, unterzeichnet (physisch oder digital) und sendet sie per E-Mail an den Autor. Der Autor bestätigt per E-Mail.
4. **Nutzungsbeginn:** Erst nach Bestätigung durch den Autor darf die Organisation die Software installieren und betreiben.

### Warum eine Pflicht zur Nutzungsvereinbarung?

| Grund                        | Erläuterung                                                                                                                                                    |
| ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Haftungsausschluss**       | Die Vereinbarung stellt sicher, dass der Autor nicht in Haftung genommen werden kann — weder für Softwarefehler noch für Compliance-Verstöße der Organisation. |
| **Freistellungsklausel**     | Die Organisation stellt den Autor von allen Ansprüchen Dritter frei, die aus der Nutzung resultieren.                                                          |
| **Kein Eigentumsübergang**   | Die Vereinbarung stellt klar, dass durch die Nutzung keinerlei Rechte an der Software übergehen.                                                               |
| **Kein Anspruch**            | Die Vereinbarung stellt klar, dass kein Anspruch auf Support, Weiterentwicklung oder Verfügbarkeit besteht.                                                    |
| **Compliance-Verantwortung** | Die Organisation übernimmt ausdrücklich die alleinige Verantwortung für die Einhaltung gesetzlicher Vorschriften (KassenSichV, GoBD, DSGVO).                   |
| **Klarheit**                 | Eine schriftliche Vereinbarung beugt Missverständnissen vor und schützt beide Seiten.                                                                          |

---

## 4. Berechtigte Nutzer und kostenlose Nutzung

### Wer erhält eine kostenlose Nutzungsvereinbarung?

Kostenlose Nutzungsvereinbarungen werden ausschließlich an folgende **Organisationen ohne Gewinnerzielungsabsicht** vergeben:

| Gruppe / Organisationsform                | Voraussetzung                                                                               |
| ----------------------------------------- | ------------------------------------------------------------------------------------------- |
| **Eingetragene Vereine (e.V.)**           | Eintragung im Vereinsregister gemäß §§ 21 ff. BGB                                           |
| **Eingetragene gemeinnützige Stiftungen** | Eintragung im Stiftungsregister, Anerkennung der Gemeinnützigkeit durch Finanzbehörden      |
| **Gemeinnützige GmbH (gGmbH)**            | Eintragung im Handelsregister, steuerliche Anerkennung der Gemeinnützigkeit                 |
| **Gemeinnützige UG (gUG)**                | Eintragung im Handelsregister, steuerliche Anerkennung der Gemeinnützigkeit                 |
| **Sonstige eingetragene NGOs / NPOs**     | Nachweisbare Eintragung in einem öffentlichen Register und fehlende Gewinnerzielungsabsicht |

### Wer erhält KEINE kostenlose Nutzungsvereinbarung?

- Gewerbliche Betriebe und Unternehmen (GmbH, AG, KG, OHG, Einzelunternehmen)
- Organisationen ohne nachweisbaren gemeinnützigen Status
- Einzelpersonen, die jotti gewerblich einsetzen möchten
- Dritte, die jotti als Dienstleistung für andere betreiben möchten (SaaS)

Diese Gruppen benötigen eine **kostenpflichtige kommerzielle Lizenz** (siehe [Abschnitt 11](#11-kommerzialisierung-und-dual-licensing)).

### Wegfall der Berechtigung

Entfallen die Voraussetzungen der Gemeinnützigkeit oder Registrierung, muss die Organisation den Autor unverzüglich informieren. Die kostenlose Nutzungsvereinbarung endet automatisch. Eine Weiternutzung erfordert eine kostenpflichtige kommerzielle Lizenz.

---

## 5. Forks, Modifikation und Weitergabe

### Grundsatz: Verboten ohne Genehmigung

Die proprietäre Source-Available-Lizenz **verbietet** die Modifikation, Ableitung und Weitergabe des Quellcodes. Im Einzelnen:

| Aktivität                                                   | Erlaubt?          |
| ----------------------------------------------------------- | ----------------- |
| Quellcode auf GitHub **lesen**                              | ✅ Ja             |
| Repository **forken**, um einen Pull Request einzureichen   | ✅ Ja (unter CLA) |
| Repository forken für **eigene Nutzung**                    | ❌ Nein           |
| Code **modifizieren** (außer für PR an das offizielle Repo) | ❌ Nein           |
| Code **weitergeben** oder **veröffentlichen**               | ❌ Nein           |
| Code in **anderes Projekt einbinden**                       | ❌ Nein           |
| Code als **eigenständige Software** betreiben               | ❌ Nein           |

### Warum so restriktiv?

Der Quellcode ist öffentlich, damit Nutzer **Vertrauen** in die Software aufbauen können — sie können prüfen, was die Software tut, ob Sicherheitslücken vorliegen und wie Daten verarbeitet werden. Das ist der Zweck der Veröffentlichung.

Die Veröffentlichung ist **kein Freibrief** zur Nutzung, Kopie oder Modifikation. Der Schutz des geistigen Eigentums und die Kontrolle über die Verbreitung sind essenziell, um:

- Kommerzielle Ausbeutung zu verhindern
- Die Qualität und Integrität der Software sicherzustellen
- Die Möglichkeit des Dual Licensing zu erhalten
- Die Haftungssituation des Autors zu kontrollieren

---

## 6. Hosting und Betrieb

### Grundsatz: Die Organisation ist Betreiberin

Die nutzende Organisation ist für Betrieb, Wartung und Sicherheit der Infrastruktur **allein verantwortlich**. Der Autor stellt **ausschließlich Quellcode** bereit — keine Infrastruktur, kein Hosting, kein SaaS, keine kompilierte Software.

| Aspekt            | Regelung                                                               |
| ----------------- | ---------------------------------------------------------------------- |
| Server            | Verantwortung der Organisation (VPS, Cloud, Webhoster, lokaler Server) |
| Domain            | Von der Organisation registriert und bezahlt                           |
| SSL-Zertifikat    | Let's Encrypt (automatisch, kostenlos)                                 |
| Datenbank-Backups | Verantwortung der Organisation                                         |
| Software-Updates  | Organisation entscheidet, wann aktualisiert wird                       |
| Kompilierung      | Organisation kompiliert und deployt selbst (Docker Compose)            |

### Was der Autor NICHT ist

| Der Autor ist **nicht** …          | Bedeutung                                                                                        |
| ---------------------------------- | ------------------------------------------------------------------------------------------------ |
| …Betreiber                         | Er betreibt keine Instanz der Software für Dritte.                                               |
| …SaaS-Anbieter                     | Er bietet keine gehostete Lösung an.                                                             |
| …Software-Vertreiber               | Er verteilt keine kompilierte oder installierbare Software.                                      |
| …Hersteller im Sinne des ProdHaftG | Er stellt ein Werk (Quellcode) zur Verfügung, kein Produkt im Sinne des Produkthaftungsgesetzes. |
| …IT-Dienstleister                  | Er schuldet keine Arbeitsleistung und keinen Service.                                            |
| …Auftragsverarbeiter               | Er hat keinen Zugriff auf Server oder Daten der Organisation.                                    |

### Hosting durch Dritte

Die Organisation kann die Infrastruktur durch einen externen Hosting-Anbieter betreiben lassen. Bedingungen:

1. **Die Organisation bleibt Betreiberin und Verantwortliche.** Sie schließt den Vertrag mit dem Hosting-Anbieter.
2. **Der Hosting-Anbieter ist Auftragsverarbeiter der Organisation** (Art. 28 DSGVO). AV-Vertrag erforderlich.
3. **Der Autor hat keinen Zugriff** auf den Server oder die Daten.
4. **Kein Multi-Tenant-Hosting.** Bietet ein Dritter jotti als gemeinsamen Dienst für mehrere Organisationen an, benötigt er eine **kommerzielle Lizenz** des Autors.

### Einrichtungshilfe

Der Autor kann bei der Ersteinrichtung **freiwillig** unterstützen (Docker Compose, Dokumentation, einmalige Begleitung). Daraus entsteht **kein Anspruch** auf laufende Betreuung, Wartung oder Betrieb.

---

## 7. Datenschutz (DSGVO)

### Welche Daten speichert jotti?

| Datenkategorie                   | Personenbezug                   | Beispiel                                      |
| -------------------------------- | ------------------------------- | --------------------------------------------- |
| Benutzerdaten (Servicekräfte)    | Ja                              | Name, Rolle, Passwort-Hash                    |
| Bestellungen & Events            | Indirekt (User-ID referenziert) | Bestellung auf Tisch 5, aufgegeben von User 3 |
| Gästedaten                       | **Nein**                        | Gäste haben keine Accounts                    |
| Zahlungsdaten (Kreditkarte etc.) | **Nein**                        | jotti verarbeitet keine Zahlungsmittel        |

### Rollen im Datenschutz

| Rolle (DSGVO)                       | Wer                                        | Begründung                                                                      |
| ----------------------------------- | ------------------------------------------ | ------------------------------------------------------------------------------- |
| **Verantwortlicher** (Art. 4 Nr. 7) | Die nutzende Organisation                  | Die Organisation entscheidet, welche Servicekräfte als Benutzer angelegt werden |
| **Auftragsverarbeiter** (Art. 28)   | Ggf. der Hosting-Anbieter der Organisation | Sofern ein externer Hoster die Infrastruktur betreibt (AV-Vertrag erforderlich) |
| **Kein Auftragsverarbeiter**        | Der Autor (Nico Gräf)                      | Der Autor hat keinen Zugriff auf den Server oder die gespeicherten Daten        |
| **Betroffene Personen**             | Servicekräfte der Organisation             | Ihre Daten werden im System gespeichert                                         |

### Pflichten der Organisation

| Pflicht                                | Beschreibung                                                                                                                          |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| **Informationspflicht** (Art. 13/14)   | Servicekräfte darüber informieren, dass ihre Daten gespeichert werden (Name, Rolle, Aktivitäten)                                      |
| **Löschpflicht** (Art. 17)             | Benutzer auf Wunsch deaktivieren/löschen (Soft-Delete vorhanden)                                                                      |
| **Verarbeitungsverzeichnis** (Art. 30) | jotti als Verarbeitungstätigkeit dokumentieren (bei Organisationen unter 250 Beschäftigten i.d.R. nicht Pflicht, aber empfehlenswert) |
| **Technische Maßnahmen** (Art. 32)     | Server absichern, HTTPS verwenden, Passwörter nicht teilen                                                                            |
| **AV-Vertrag** (Art. 28)               | Mit dem Hosting-Anbieter schließen, sofern dieser die Infrastruktur betreibt                                                          |

---

## 8. Haftung und Gewährleistung

### Haftungsausschluss

Der Autor stellt den Quellcode **unentgeltlich** und **ohne jede Gewähr** zur Verfügung. Es besteht **kein vertragliches oder gesetzliches Schuldverhältnis**, das über die in der Nutzungsvereinbarung gewährte Lizenz hinausgeht.

| Aspekt                | Regelung                                                                                                                                                               |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Gewährleistung**    | Keine — die Software wird „as-is" bereitgestellt. Es gibt keine Zusicherung von Fehlerfreiheit, Funktionalität, Verfügbarkeit oder Eignung für einen bestimmten Zweck. |
| **Support**           | Kein Anspruch. Unterstützung durch den Autor ist freiwillig und unverbindlich.                                                                                         |
| **Weiterentwicklung** | Kein Anspruch. Der Autor entscheidet allein über Art und Umfang der Weiterentwicklung.                                                                                 |
| **Fehlerbehebung**    | Kein Anspruch. Bugs werden nach Ermessen des Autors behoben — oder nicht.                                                                                              |
| **Verfügbarkeit**     | Kein Anspruch. Der Autor kann das Repository oder die Software jederzeit offline nehmen.                                                                               |

### Haftungsbegrenzung nach deutschem Recht

Da die Software **unentgeltlich** überlassen wird, gelten die Grundsätze des Schenkungsrechts (§ 521 BGB):

| Haftungsmaßstab                              | Regelung                                                                                                                                         |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Vorsatz** (§ 276 Abs. 3 BGB)               | Haftung **kann nicht ausgeschlossen werden** — gesetzlich zwingend.                                                                              |
| **Grobe Fahrlässigkeit**                     | Haftung besteht — kann bei unentgeltlicher Überlassung nicht ausgeschlossen werden.                                                              |
| **Einfache Fahrlässigkeit**                  | Haftung **ausgeschlossen** — bei unentgeltlicher Überlassung (§ 521 BGB).                                                                        |
| **Verletzung von Leben, Körper, Gesundheit** | Haftung **kann nicht ausgeschlossen werden** — gesetzlich zwingend (§ 309 Nr. 7a BGB).                                                           |
| **Produkthaftung** (ProdHaftG)               | Soweit anwendbar, bleibt unberührt. Der Autor ist jedoch kein Hersteller im Sinne des ProdHaftG, da er lediglich Quellcode zur Verfügung stellt. |

### Was der Autor NICHT schuldet

- Keine Garantie, dass die Software gesetzliche Anforderungen erfüllt (KassenSichV, GoBD, DSFinV-K, TSE)
- Keine Garantie der Richtigkeit von Berechnungen (Saldo, Steuersätze, Abrechnungen)
- Keine Garantie der Datensicherheit oder Datenverfügbarkeit
- Keine Garantie der Kompatibilität mit bestimmten Systemen oder Browsern
- Kein Schadensersatz bei Datenverlust, Fehlberechnungen, Ausfällen, Betriebsprüfungen oder behördlichen Beanstandungen

---

## 9. Freistellungsklausel

### Umfassende Freistellung durch die Organisation

Die Organisation stellt den Autor (Nico Gräf) von **sämtlichen Ansprüchen, Schäden, Verlusten, Kosten und Aufwendungen** (einschließlich angemessener Rechtsanwalts- und Gerichtskosten) frei, die aus folgenden Umständen entstehen oder damit zusammenhängen:

1. **Nutzung, Deployment oder Betrieb** der Software durch die Organisation
2. **Nichteinhaltung** geltender Gesetze, Vorschriften oder dieser Lizenz durch die Organisation
3. **Ansprüche Dritter** (einschließlich, aber nicht beschränkt auf Finanzbehörden, Aufsichtsbehörden, Arbeitnehmer, Kunden oder Endnutzer), die aus der Nutzung der Software durch die Organisation resultieren
4. **Datenschutzverstöße**, die aus dem Betrieb der Software durch die Organisation resultieren
5. **Steuerliche Beanstandungen**, Ordnungswidrigkeiten oder Bußgelder im Zusammenhang mit der Nutzung der Software als Kassensystem (KassenSichV, § 146a AO, GoBD)

### Begründung

Der Autor stellt Quellcode zur Verfügung. Er betreibt die Software nicht, er vertreibt kein fertiges Produkt, er kontrolliert nicht, wie die Organisation die Software einsetzt, konfiguriert oder betreibt. Die alleinige Verantwortung für den ordnungsgemäßen Betrieb und die Einhaltung gesetzlicher Vorschriften liegt bei der Organisation als Betreiberin.

---

## 10. Compliance-Verantwortung (KassenSichV / TSE)

### Klarstellung: Compliance ist Sache der Organisation

jotti ist ein elektronisches Aufzeichnungssystem im Sinne von § 1 KassenSichV und unterliegt damit der TSE-Pflicht nach § 146a AO. **Die Einhaltung dieser Pflichten obliegt ausschließlich der Organisation als Betreiberin** — nicht dem Autor der Software.

| Pflicht                               | Verantwortlicher | Erläuterung                                                           |
| ------------------------------------- | ---------------- | --------------------------------------------------------------------- |
| TSE-Anbindung (§ 146a AO)             | Organisation     | Die Organisation muss eine zertifizierte TSE beschaffen und anbinden. |
| Kassenmeldung (§ 146a Abs. 4 AO)      | Organisation     | Meldung des Kassensystems an das zuständige Finanzamt über ELSTER.    |
| GoBD-konforme Aufbewahrung            | Organisation     | Aufbewahrungspflicht für digitale Aufzeichnungen (10 Jahre).          |
| DSFinV-K-Export                       | Organisation     | Bereitstellung des Exports bei Betriebsprüfung.                       |
| Belegausgabepflicht (§ 6 KassenSichV) | Organisation     | Ausgabe von Belegen an Kunden.                                        |
| Ordnungsgemäßer Betrieb               | Organisation     | Korrekte Konfiguration, regelmäßige Backups, Systemverfügbarkeit.     |

### Was der Autor leistet

Der Autor implementiert **technische Schnittstellen und Funktionen** zur Unterstützung der Compliance (TSE-Adapter, DSFinV-K-Export, Belegdruck — siehe `docs/roadmap.md`). Der Autor **garantiert jedoch nicht**, dass:

- Die Implementierung fehlerfrei oder vollständig ist
- Die Software alle gesetzlichen Anforderungen erfüllt
- Die Software einer Betriebsprüfung standhält
- Die technischen Schnittstellen korrekt funktionieren

**Die Organisation ist verpflichtet**, die Software vor dem produktiven Einsatz **eigenständig** auf Eignung und Compliance zu prüfen — ggf. unter Hinzuziehung eines Steuerberaters oder einer fachkundigen Person.

---

## 11. Kommerzialisierung und Dual Licensing

Als alleiniger Urheber besitzt Nico Gräf das Recht, jotti unter beliebig vielen Lizenzen gleichzeitig zu veröffentlichen (**Dual-Licensing-Modell**):

| Pfad            | Lizenz                                                         | Zielgruppe                                                    | Kosten          |
| --------------- | -------------------------------------------------------------- | ------------------------------------------------------------- | --------------- |
| **Non-Profit**  | Proprietäre Source-Available + kostenlose Nutzungsvereinbarung | Eingetragene Vereine, gemeinnützige Organisationen, NGOs/NPOs | Kostenlos       |
| **Kommerziell** | Proprietäre kommerzielle Lizenz                                | Unternehmen, gewerbliche Betriebe, SaaS-Anbieter              | Kostenpflichtig |

### Vorbehaltene Rechte

Nico Gräf behält sich ausdrücklich folgende Rechte vor:

1. **Kostenpflichtige Nutzungslizenzen anbieten** — z.B. für gewerbliche Betriebe.
2. **jotti als kostenpflichtiges SaaS-Produkt betreiben** — als gehostete Lösung mit Support und SLA.
3. **Organisationen von der kostenlosen Nutzung auszunehmen** — z.B. wenn Anspruchsvoraussetzungen nicht erfüllt sind.
4. **Die Lizenz zukünftiger Versionen zu ändern.**
5. **Die kostenlose Nutzungsvereinbarung jederzeit zu widerrufen** — mit angemessener Frist.

### Mögliche Kommerzialisierungsmodelle

| Modell                   | Beschreibung                                                                          |
| ------------------------ | ------------------------------------------------------------------------------------- |
| **Hosting-as-a-Service** | jotti als gehostete Lösung (monatliches Abo). Auftragsverarbeiter → AV-Vertrag nötig. |
| **Setup-Pakete**         | Einmalige Einrichtung + Einweisung gegen Festpreis.                                   |
| **Support-Verträge**     | Garantierte Reaktionszeiten, Hotline am Festtag.                                      |
| **Enterprise-Lizenz**    | Gewerbliche Betriebe zahlen eine Lizenzgebühr.                                        |
| **White-Label-Lizenz**   | Dritte dürfen jotti unter eigenem Namen anbieten — nur mit kommerzieller Lizenz.      |

---

## 12. Community-Beiträge und CLA

### Contributor License Agreement (CLA)

Das Projekt nimmt Pull Requests an. **Jeder Beitrag** (Pull Request, Patch, Code-Einreichung) unterliegt dem **Contributor License Agreement (CLA)**, das im Repository veröffentlicht ist (siehe `CLA.md`).

### Kernpunkte des CLA

Durch die Einreichung eines Beitrags gewährt der Beitragende dem Autor:

1. Eine **unwiderrufliche, weltweite, gebührenfreie, nicht-exklusive Lizenz** zur Nutzung, Vervielfältigung, Modifikation, Unterlizenzierung, Relizenzierung und Verbreitung des Beitrags unter beliebigen Bedingungen — einschließlich proprietärer und kommerzieller Lizenzen.
2. Der Beitragende **behält das Urheberrecht** an seinem Beitrag, kann aber die dem Autor gewährten Rechte **nicht widerrufen**.
3. Der Beitragende bestätigt, dass der Beitrag seine **eigene Schöpfung** ist und er berechtigt ist, diese Rechte zu gewähren.

### Warum ein CLA?

Das CLA ist notwendig, um das Dual-Licensing-Modell aufrechtzuerhalten. Ohne CLA könnten Beiträge Dritter nur unter der proprietären Source-Available-Lizenz verbreitet werden — der Autor könnte sie nicht in eine kommerzielle Lizenz aufnehmen. Das CLA stellt sicher, dass der Autor die volle Kontrolle über die Lizenzierung aller Codebestandteile behält.

---

## 13. Muster-Nutzungsvereinbarung

Die folgende Vereinbarung wird als Vorlage für die schriftliche Nutzungsvereinbarung verwendet. Sie ist **kein Rechtsberatungsprodukt** und ersetzt keine anwaltliche Prüfung.

---

### Nutzungsvereinbarung für die Software „jotti"

**Zwischen:**

- **Autor:** Nico Gräf, Freiburg im Breisgau (nachfolgend „Autor")
- **Nutzer:** [ORGANISATIONSNAME], vertreten durch [NAME, FUNKTION] (nachfolgend „Organisation")

**Stand:** \_\_\_\_\_\_\_\_\_\_\_\_\_

---

#### § 1 Gegenstand

(1) Der Autor stellt der Organisation die Software „jotti" (nachfolgend „Software") zur unentgeltlichen Nutzung zur Verfügung. Die Software umfasst den Quellcode, die Dokumentation und die zugehörigen Konfigurationsdateien, wie im Repository unter https://github.com/nicograef/jotti veröffentlicht.

(2) Die Software ist ein persönliches Projekt des Autors. Die Organisation hat die Entwicklung weder beauftragt noch finanziert. Es besteht kein Auftrags-, Dienst- oder Werkvertragsverhältnis.

(3) Der Autor stellt ausschließlich Quellcode zur Verfügung. Er betreibt kein SaaS-Produkt, vertreibt keine kompilierte Software und bietet keine gehostete Lösung an.

#### § 2 Berechtigung zur unentgeltlichen Nutzung

(1) Die unentgeltliche Nutzung steht ausschließlich offiziell eingetragenen und nicht-gewinnorientierten Organisationen offen, insbesondere:

- eingetragenen Vereinen (e.V.) gemäß §§ 21 ff. BGB,
- eingetragenen gemeinnützigen Stiftungen,
- gemeinnützigen GmbH (gGmbH) und gemeinnützigen UG (gUG),
- sonstigen eingetragenen gemeinnützigen Körperschaften ohne Gewinnerzielungsabsicht.

(2) Die Organisation sichert zu, die Voraussetzungen nach Abs. 1 zu erfüllen und weist dies auf Verlangen des Autors nach (z.B. Vereinsregisterauszug, Freistellungsbescheid).

(3) Entfallen die Voraussetzungen nach Abs. 1, informiert die Organisation den Autor unverzüglich. Die unentgeltliche Nutzungslizenz endet in diesem Fall automatisch.

(4) Gewerbliche oder kommerzielle Nutzung ist ohne gesonderte kostenpflichtige Lizenz nicht gestattet.

#### § 3 Geistiges Eigentum

(1) Die Software — einschließlich Quellcode, Dokumentation, Architektur und Design — ist das **alleinige geistige Eigentum** des Autors. Sämtliche Urheber-, Marken- und sonstigen Schutzrechte verbleiben beim Autor.

(2) Durch die Nutzung der Software erwirbt die Organisation **keinerlei Rechte** an der Software über die in § 4 gewährte Lizenz hinaus. Insbesondere geht kein Eigentum, kein Urheberrecht und kein sonstiges Schutzrecht auf die Organisation über.

(3) Der Autor behält sich das Recht vor, die Software jederzeit unter weiteren Lizenzen zu veröffentlichen, zu vermarkten oder anderweitig zu verwerten.

#### § 4 Nutzungslizenz

(1) Die Organisation erhält eine **unentgeltliche, nicht-exklusive, widerrufliche, nicht übertragbare** Lizenz zur Nutzung der Software für den organisationsinternen Kassenbetrieb bei eigenen Veranstaltungen.

(2) Die Nutzung umfasst: Installation, Konfiguration, Betrieb und Aktualisierung auf von der Organisation verantworteter Infrastruktur.

(3) Die Lizenz umfasst **nicht**: Modifikation des Quellcodes, Weitergabe an Dritte, Unterlizenzierung, öffentliches Zugänglichmachen, Einbindung in andere Software, Nutzung als Grundlage für abgeleitete Werke oder White-Label-Betrieb.

(4) Es gelten zusätzlich die Bestimmungen der proprietären Source-Available-Lizenz aus der `LICENSE`-Datei des Repositories.

#### § 5 Hosting und Betrieb

(1) Die Organisation ist **alleinige Betreiberin** der Software und verantwortet die hierfür eingesetzte Infrastruktur (Server, Datenbank, Netzwerk, Backups, Sicherheit).

(2) Der Autor stellt weder Hosting noch Infrastruktur bereit.

(3) Die Organisation kann die Infrastruktur durch einen externen Hosting-Anbieter betreiben lassen. In diesem Fall ist die Organisation für den Abschluss eines Auftragsverarbeitungsvertrags (Art. 28 DSGVO) mit dem Hoster verantwortlich.

(4) Ein Multi-Tenant-Betrieb, bei dem ein Dritter jotti als Dienst für mehrere Organisationen anbietet, ist ohne separate kommerzielle Lizenz des Autors nicht gestattet.

(5) Der Autor kann freiwillig bei der Ersteinrichtung unterstützen. Daraus entsteht kein Anspruch auf laufende Betreuung oder Betrieb.

#### § 6 Datenschutz

(1) Die Organisation ist **Verantwortliche** im Sinne der DSGVO (Art. 4 Nr. 7) für alle in der Software verarbeiteten personenbezogenen Daten.

(2) Der Autor ist **kein Auftragsverarbeiter**, da er keinen Zugriff auf den Server oder die gespeicherten Daten hat. Zwischen dem Autor und der Organisation besteht kein Auftragsverarbeitungsverhältnis.

(3) Die Organisation ist verpflichtet, die einschlägigen Datenschutzbestimmungen einzuhalten, insbesondere die Betroffenen (Servicekräfte) über die Datenverarbeitung zu informieren und mit einem etwaigen Hosting-Anbieter einen AV-Vertrag abzuschließen.

#### § 7 Gewährleistungsausschluss

(1) Die Software wird **ohne jede Gewähr** bereitgestellt („as-is"). Der Autor übernimmt keinerlei Gewährleistung — weder ausdrücklich noch stillschweigend — für Fehlerfreiheit, Funktionalität, Verfügbarkeit, Richtigkeit, Vollständigkeit oder Eignung für einen bestimmten Zweck.

(2) Insbesondere übernimmt der Autor **keine Gewähr** für:

- die Einhaltung gesetzlicher Vorschriften (KassenSichV, GoBD, DSFinV-K, TSE-Anforderungen),
- die Richtigkeit von Berechnungen (Saldo, Steuersätze, Abrechnungen),
- die Datensicherheit oder Datenverfügbarkeit,
- die Kompatibilität mit bestimmten Systemen, Browsern oder Endgeräten.

(3) Die Organisation ist verpflichtet, die Software vor dem produktiven Einsatz eigenständig auf Eignung und Compliance zu prüfen — ggf. unter Hinzuziehung eines Steuerberaters oder einer fachkundigen Person.

#### § 8 Haftungsbegrenzung

(1) Da die Software unentgeltlich überlassen wird, haftet der Autor nach den Grundsätzen des Schenkungsrechts (§ 521 BGB) **nur für Vorsatz und grobe Fahrlässigkeit**.

(2) Die Haftung für einfache Fahrlässigkeit ist — soweit gesetzlich zulässig — ausgeschlossen.

(3) Die Haftung für Verletzung von Leben, Körper oder Gesundheit bleibt unberührt (§ 309 Nr. 7a BGB).

(4) Die Haftung nach dem Produkthaftungsgesetz bleibt unberührt, soweit dessen Anwendungsbereich eröffnet ist. Der Autor weist darauf hin, dass er lediglich Quellcode zur Verfügung stellt und kein Produkt im Sinne des ProdHaftG herstellt oder in Verkehr bringt.

(5) Der Autor haftet **insbesondere nicht** für Schäden aus: Datenverlust, Fehlberechnungen, Systemausfällen, Betriebsprüfungen, steuerlichen Beanstandungen, Ordnungswidrigkeiten, Bußgeldern oder Nachforderungen, die im Zusammenhang mit der Nutzung der Software stehen.

#### § 9 Freistellung

(1) Die Organisation stellt den Autor von **sämtlichen Ansprüchen, Schäden, Verlusten, Kosten und Aufwendungen** (einschließlich angemessener Rechtsanwalts- und Gerichtskosten) frei, die aus den folgenden Umständen entstehen oder damit zusammenhängen:

a) der Nutzung, dem Deployment oder dem Betrieb der Software durch die Organisation;
b) der Nichteinhaltung geltender Gesetze, Vorschriften oder dieser Vereinbarung durch die Organisation;
c) Ansprüchen Dritter (einschließlich Finanzbehörden, Aufsichtsbehörden, Mitarbeitern, Kunden oder Endnutzern), die aus der Nutzung der Software resultieren;
d) Datenschutzverstößen, die aus dem Betrieb der Software resultieren;
e) steuerlichen Beanstandungen, Ordnungswidrigkeiten oder Bußgeldern im Zusammenhang mit der Nutzung der Software als Kassensystem.

(2) Die Freistellungspflicht gilt auch nach Beendigung dieser Vereinbarung fort, soweit die zugrundeliegenden Ansprüche auf Handlungen oder Unterlassungen während der Vertragslaufzeit zurückzuführen sind.

#### § 10 Compliance-Verantwortung

(1) Die Organisation ist **allein verantwortlich** für die Einhaltung sämtlicher gesetzlicher Vorschriften im Zusammenhang mit dem Betrieb eines elektronischen Aufzeichnungssystems, insbesondere:

a) KassenSichV und § 146a AO (TSE-Pflicht, Belegausgabe, Kassenmeldung),
b) GoBD (ordnungsgemäße Buchführung und Aufbewahrung),
c) DSFinV-K (digitale Schnittstelle der Finanzverwaltung),
d) DSGVO (Datenschutz),
e) sonstige steuer- und handelsrechtliche Vorschriften.

(2) Der Autor implementiert technische Schnittstellen zur Unterstützung der Compliance, übernimmt jedoch **keinerlei Garantie** für deren Vollständigkeit, Richtigkeit oder Eignung.

(3) Die Organisation ist verpflichtet, sich vor dem Einsatz der Software eigenständig über die geltenden gesetzlichen Anforderungen zu informieren und deren Einhaltung sicherzustellen.

#### § 11 Support und Weiterentwicklung

(1) Es besteht **kein Anspruch** auf Support, Fehlerbehebung, Weiterentwicklung oder Aktualisierung.

(2) Unterstützung durch den Autor erfolgt freiwillig und unverbindlich.

(3) Feature-Wünsche der Organisation werden zur Kenntnis genommen, begründen jedoch keinen Umsetzungsanspruch.

#### § 12 Laufzeit und Kündigung

(1) Diese Vereinbarung gilt auf unbestimmte Zeit.

(2) Beide Seiten können die Vereinbarung jederzeit ohne Angabe von Gründen und ohne Einhaltung einer Frist beenden. Die Beendigung hat in Textform (E-Mail genügt) zu erfolgen.

(3) Bei Beendigung erlischt die Nutzungslizenz nach § 4 Abs. 1. Die Organisation hat die Software innerhalb von 30 Tagen nach Beendigung von ihrer Infrastruktur zu entfernen und den Autor über die erfolgte Löschung zu informieren.

(4) Die Freistellungspflicht (§ 9) und der Haftungsausschluss (§§ 7, 8) gelten über die Beendigung hinaus fort.

#### § 13 Schlussbestimmungen

(1) Änderungen und Ergänzungen dieser Vereinbarung bedürfen der Textform (E-Mail genügt).

(2) Sollten einzelne Bestimmungen unwirksam sein, bleibt die Wirksamkeit der übrigen Bestimmungen unberührt. An die Stelle der unwirksamen Bestimmung tritt eine wirksame Regelung, die dem wirtschaftlichen Zweck der unwirksamen Bestimmung am nächsten kommt.

(3) Es gilt das Recht der Bundesrepublik Deutschland unter Ausschluss des UN-Kaufrechts (CISG).

(4) Gerichtsstand für alle Streitigkeiten aus oder im Zusammenhang mit dieser Vereinbarung ist — soweit gesetzlich zulässig — Freiburg im Breisgau.

(5) Die Organisation bestätigt, diese Vereinbarung gelesen und verstanden zu haben und erklärt sich mit allen Bestimmungen einverstanden.

---

**Ort, Datum:** \_\_\_\_\_\_\_\_\_\_\_\_\_

**Autor:** \_\_\_\_\_\_\_\_\_\_\_\_\_ (Nico Gräf)

**Organisation:** \_\_\_\_\_\_\_\_\_\_\_\_\_ (vertretungsberechtigte Person, [ORGANISATIONSNAME])

---

> **Hinweis:** Diese Vereinbarung ist eine Vorlage und keine Rechtsberatung. Bei Unsicherheiten empfiehlt sich die Prüfung durch eine Rechtsanwältin oder einen Rechtsanwalt.

---

## Anhang: Hinweise für den Entwickler

---

### A. Sonderfall: Entwickler ist Vereinsmitglied

Wenn der Autor gleichzeitig Mitglied des nutzenden Vereins ist, entsteht ein potenzieller Graubereich: Könnte der Verein argumentieren, die Software sei im Rahmen der Vereinstätigkeit entstanden?

Im deutschen Recht gilt: Vereinsmitgliedschaft begründet kein Arbeitsverhältnis (§ 7 UrhG). Das Urheberrecht entsteht beim Schöpfer. § 69b UrhG (Arbeitnehmerurheberrecht) gilt nicht für ehrenamtliche Vereinstätigkeit.

**Faktenlage bei jotti:** Entwickelt auf eigenem Rechner, in eigener Zeit, mit eigenen Werkzeugen. Öffentliches Projekt auf GitHub unter Nico Gräfs persönlichem Account. Kein Vereinsbeschluss zur Beauftragung. Die Nutzungsvereinbarung (§ 1 Abs. 2) hält explizit fest, dass kein Auftrags-, Dienst- oder Werkvertragsverhältnis besteht.

| Situation                                  | Empfehlung                                                                                             |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------ |
| Verein wünscht sich Feature X              | Prüfen, ob es ins Backlog passt. Kein Versprechen.                                                     |
| Verein sagt „Wir brauchen das bis Samstag" | Klar kommunizieren: „Ich bin kein Auftragnehmer."                                                      |
| Verein bietet Geld für ein Feature         | **Vorsicht** — dann entsteht ein Auftragsverhältnis. Entweder ablehnen oder als Werkvertrag aufsetzen. |

---

### B. Support und Erwartungsmanagement

> **Du stellst Quellcode eines persönlichen Projekts zur Verfügung — du bist nicht IT-Dienstleister der Organisation.**

| Situation                                | Empfohlene Reaktion                                                                |
| ---------------------------------------- | ---------------------------------------------------------------------------------- |
| „Das muss bis Samstag funktionieren"     | „Ich gebe mein Bestes, aber verspreche nichts. Testet vorher."                     |
| „Kannst du noch schnell X einbauen?"     | „Steht auf der Wunschliste. Kommt, wenn es kommt."                                 |
| „Das ist kaputt, reparier das sofort"    | „Ich schau es mir an, wenn ich Zeit habe. Für heute: Stift & Papier als Fallback." |
| Neuer Vorstand kennt die Absprache nicht | Nutzungsvereinbarung ist schriftlich beim Verein hinterlegt.                       |

Praktische Empfehlungen:

- Nutzungsvereinbarung **vor dem ersten Einsatz** per E-Mail abschließen
- Eine feste Ansprechperson im Verein benennen
- Testlauf vor dem echten Einsatz durchführen
- Fallback-Plan (Stift & Papier) definieren

---

### C. Wirtschaftlicher Hintergrund

Um den Wert der unentgeltlich bereitgestellten Software einzuordnen, wurden drei Budgetszenarien berechnet (42 Anforderungen, Neuentwicklung von Grund auf):

| Szenario                                    | Personentage | Budget (netto) | Laufzeit |
| ------------------------------------------- | ------------ | -------------- | -------- |
| **Software-Agentur** (konservativ, 3,5 FTE) | 274 PT       | ~250.000 €     | 9 Monate |
| **Software-Agentur** (optimiert, 2,6 FTE)   | 160 PT       | ~145.000 €     | 7 Monate |
| **Senior Freelancer** (Solo, 900 €/Tag)     | 133 PT       | ~122.000 €     | 8 Monate |

Grundlage: marktübliche Tagessätze (750–1.050 €/PT), vollständige Umsetzung aller Must-have- und Nice-to-have-Anforderungen. Diese Zahlen dienen der Dokumentation des unentgeltlichen Charakters der Bereitstellung sowie dem Erwartungsmanagement gegenüber nutzenden Organisationen.
