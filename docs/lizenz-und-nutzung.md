# Lizenz, Nutzung & Rechtliches — jotti

Dieses Dokument regelt die rechtliche und organisatorische Grundlage für die Entwicklung und Nutzung von jotti. Es ist die verbindliche Referenz für alle Fragen zu Eigentum, Lizenzierung, Nutzungsbedingungen, Haftung und Compliance.

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
13. [Nutzungsbedingungen](#13-nutzungsbedingungen)

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

jotti steht seit dem 21. März 2026 unter einer **proprietären Source-Available-Lizenz**. Der Quellcode ist öffentlich einsehbar, aber **es werden keine Nutzungsrechte automatisch gewährt**. Jede Nutzung erfordert eine vorherige Nutzungsvereinbarung in Textform mit dem Autor.

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

**Alles andere** — insbesondere Installation, Deployment, Hosting, Ausführung, Kopieren, Modifizieren und Weitergabe — erfordert eine vorherige Nutzungsvereinbarung in Textform.

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

**Jede Organisation, die jotti nutzen möchte, muss vor der ersten Nutzung eine Nutzungsvereinbarung in Textform mit dem Autor (Nico Gräf) abschließen.**

Es gibt keine Ausnahmen. Weder die Gemeinnützigkeit noch die Unentgeltlichkeit noch die Tatsache, dass der Quellcode öffentlich ist, befreit von der Pflicht zur Nutzungsvereinbarung.

### Ablauf

1. **Anfrage:** Die Organisation sendet eine E-Mail an graef.nico@gmail.com mit Name, Rechtsform, Registernummer, Ansprechperson und der ausdrücklichen Erklärung, die Nutzungsbedingungen (Stand: 3. Juni 2026) zu akzeptieren — vollständiger Text und E-Mail-Vorlage: [docs/nutzungsbedingungen.md](nutzungsbedingungen.md).
2. **Bestätigung:** Der Autor antwortet per E-Mail. Erst nach dieser Bestätigung darf die Organisation die Software installieren und betreiben.

### Warum eine Pflicht zur Nutzungsvereinbarung?

| Grund                        | Erläuterung                                                                                                                                                    |
| ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Haftungsausschluss**       | Die Vereinbarung stellt sicher, dass der Autor nicht in Haftung genommen werden kann — weder für Softwarefehler noch für Compliance-Verstöße der Organisation. |
| **Freistellungsklausel**     | Die Organisation stellt den Autor von allen Ansprüchen Dritter frei, die aus der Nutzung resultieren.                                                          |
| **Kein Eigentumsübergang**   | Die Vereinbarung stellt klar, dass durch die Nutzung keinerlei Rechte an der Software übergehen.                                                               |
| **Kein Anspruch**            | Die Vereinbarung stellt klar, dass kein Anspruch auf Support, Weiterentwicklung oder Verfügbarkeit besteht.                                                    |
| **Compliance-Verantwortung** | Die Organisation übernimmt ausdrücklich die alleinige Verantwortung für die Einhaltung gesetzlicher Vorschriften (KassenSichV, GoBD, DSGVO).                   |
| **Klarheit**                 | Eine Vereinbarung in Textform beugt Missverständnissen vor und schützt beide Seiten.                                                                           |

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

Die nutzende Organisation ist **alleinige Betreiberin**: Sie verantwortet Server, Domain, Datenbank-Backups, SSL (Let's Encrypt) und Updates und kompiliert die Software selbst (Docker Compose). Der Autor stellt **ausschließlich Quellcode** bereit — kein Hosting, kein SaaS, keine kompilierte Software — und ist damit weder Betreiber noch SaaS-Anbieter, Software-Vertreiber, Auftragsverarbeiter oder Hersteller im Sinne des ProdHaftG. Ein externer Hoster ist Auftragsverarbeiter der Organisation (AV-Vertrag nach Art. 28 DSGVO); Multi-Tenant-Betrieb für mehrere Organisationen erfordert eine kommerzielle Lizenz. Eine freiwillige Einrichtungshilfe begründet keinen Anspruch auf laufende Betreuung.

Verbindlich geregelt in [Nutzungsbedingungen § 5](nutzungsbedingungen.md).

---

## 7. Datenschutz (DSGVO)

jotti speichert personenbezogene Daten nur über die Servicekräfte (Name, Rolle, Passwort-Hash) und indirekt über deren Aktivitäten (Bestellungen, Events); Gäste haben keine Accounts, Zahlungsmittel werden nicht verarbeitet. Datenschutzrechtlich ist die Organisation **Verantwortliche** (Art. 4 Nr. 7 DSGVO) und damit zuständig für Informations- und Löschpflichten, technische Maßnahmen (Art. 32) und — bei externem Hosting — den AV-Vertrag (Art. 28). Der Autor ist **kein** Auftragsverarbeiter, da er keinen Zugriff auf Server oder Daten hat.

Verbindlich geregelt in [Nutzungsbedingungen § 6](nutzungsbedingungen.md).

---

## 8. Haftung und Gewährleistung

Der Autor stellt den Quellcode **unentgeltlich** und **ohne jede Gewähr** („as-is") bereit: keine Zusicherung von Fehlerfreiheit, Funktionalität, Verfügbarkeit oder Eignung und kein Anspruch auf Support, Weiterentwicklung oder Fehlerbehebung. Da die Überlassung unentgeltlich ist, haftet der Autor nach Schenkungsrecht (§ 521 BGB) nur für Vorsatz und grobe Fahrlässigkeit; die Haftung für die Verletzung von Leben, Körper und Gesundheit (§ 309 Nr. 7a BGB) sowie nach dem Produkthaftungsgesetz bleibt unberührt. Insbesondere haftet der Autor nicht für Datenverlust, Fehlberechnungen, Ausfälle oder Beanstandungen aus Betriebsprüfungen.

Verbindlich geregelt in [Nutzungsbedingungen §§ 7–8](nutzungsbedingungen.md).

---

## 9. Freistellungsklausel

Die Organisation stellt den Autor von **sämtlichen Ansprüchen, Schäden, Verlusten und Kosten** frei, die aus ihrer Nutzung, ihrem Deployment oder Betrieb der Software, aus Gesetzes- oder Datenschutzverstößen, aus Ansprüchen Dritter (z. B. Finanz- und Aufsichtsbehörden, Arbeitnehmer, Kunden) oder aus steuerlichen Beanstandungen als Kassensystem (KassenSichV, § 146a AO, GoBD) entstehen. Hintergrund: Der Autor stellt nur Quellcode bereit und kontrolliert weder Konfiguration noch Betrieb — die alleinige Verantwortung trägt die Organisation als Betreiberin.

Verbindlich geregelt in [Nutzungsbedingungen § 9](nutzungsbedingungen.md).

---

## 10. Compliance-Verantwortung (KassenSichV / TSE)

jotti ist ein elektronisches Aufzeichnungssystem im Sinne von § 1 KassenSichV (TSE-Pflicht nach § 146a AO). Die Einhaltung aller daraus folgenden Pflichten — TSE-Anbindung, Kassenmeldung über ELSTER, GoBD-konforme Aufbewahrung, DSFinV-K-Export, Belegausgabe und ordnungsgemäßer Betrieb — obliegt **allein der Organisation** als Betreiberin. Der Autor implementiert unterstützende technische Schnittstellen (TSE-Adapter, DSFinV-K-Export, Belegdruck — siehe [compliance.md](compliance.md)), garantiert aber weder deren Vollständigkeit noch Richtigkeit; die Organisation muss die Software vor dem produktiven Einsatz eigenständig auf Eignung und Compliance prüfen, ggf. mit Steuerberater.

Verbindlich geregelt in [Nutzungsbedingungen § 10](nutzungsbedingungen.md).

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

## 13. Nutzungsbedingungen

Die vollständigen Nutzungsbedingungen für jotti — einschließlich aller §§ 1–13, des Ablaufs zur Vereinbarung in Textform und der E-Mail-Vorlage für die Anfrage — sind in einem separaten Dokument zusammengefasst:

**[docs/nutzungsbedingungen.md](nutzungsbedingungen.md)**

Die Nutzungsbedingungen (Stand: 3. Juni 2026) ersetzen die bisherige Muster-Nutzungsvereinbarung und bilden die rechtliche Grundlage für jede Nutzung von jotti.

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
| Neuer Vorstand kennt die Absprache nicht | Nutzungsvereinbarung ist in Textform (E-Mail) beim Verein hinterlegt.              |

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
