# Lizenz, Nutzung & Rechtliches — jotti

Dieses Dokument regelt die rechtliche und organisatorische Grundlage für die Entwicklung und Nutzung von jotti. Es dient als Referenz für den Entwickler (Nico Gräf), nutzende Vereine und potenzielle Mitwirkende.

> **Status:** Lebendes Dokument. Letzte Aktualisierung: 10. März 2026.
>
> **Verwandte Dokumente:**
>
> - [Anforderungskatalog](requirements.md) — 50 Anforderungen mit Status
> - [Entwicklung & Deployment](development.md) — Setup, Tests, CI/CD
> - [AGENTS.md](../AGENTS.md) — Technische Konventionen für KI-Agenten

---

## Inhaltsverzeichnis

1. [Eigentumsverhältnisse](#1-eigentumsverhältnisse)
2. [Open-Source-Lizenz](#2-open-source-lizenz)
3. [Nutzungsvereinbarung für Vereine](#3-nutzungsvereinbarung-für-vereine)
4. [Sonderfall: Entwickler ist Vereinsmitglied](#4-sonderfall-entwickler-ist-vereinsmitglied)
5. [Externe Vereine und Organisationen](#5-externe-vereine-und-organisationen)
6. [Hosting und Betrieb](#6-hosting-und-betrieb)
7. [Datenschutz (DSGVO)](#7-datenschutz-dsgvo)
8. [Haftung und Gewährleistung](#8-haftung-und-gewährleistung)
9. [Support und Erwartungsmanagement](#9-support-und-erwartungsmanagement)
10. [Kommerzialisierung und Dual Licensing](#10-kommerzialisierung-und-dual-licensing)
11. [Wirtschaftlicher Hintergrund](#11-wirtschaftlicher-hintergrund)
12. [Muster-Nutzungsvereinbarung](#12-muster-nutzungsvereinbarung)

---

## 1. Eigentumsverhältnisse

### Grundsatz

Die Software **jotti** — einschließlich Quellcode, Dokumentation, Architektur, Design und aller zugehörigen Artefakte — ist das alleinige geistige Eigentum von **Nico Gräf**.

| Aspekt                     | Regelung                                                         |
| -------------------------- | ---------------------------------------------------------------- |
| Urheberrecht               | Nico Gräf, seit Projektbeginn 2025                               |
| IP (Intellectual Property) | Alle Rechte vorbehalten                                          |
| Repository                 | [github.com/nicograef/jotti](https://github.com/nicograef/jotti) |
| Lizenz                     | AGPL-3.0-or-later (siehe [Abschnitt 2](#2-open-source-lizenz))   |

### Was bedeutet das konkret?

- Nico Gräf entscheidet allein über Lizenzierung, Weiterentwicklung und Verbreitung.
- Kein Nutzer (Verein, Organisation, Person) erwirbt durch die Nutzung Rechte an der Software.
- Der Quellcode ist öffentlich einsehbar (Open Source), aber die Urheberschaft bleibt unberührt.
- Beiträge Dritter (Pull Requests) unterliegen der Projektlizenz; der Urheber behält das Recht auf Relizenzierung (siehe [Abschnitt 10](#10-kommerzialisierung-und-dual-licensing)).

---

## 2. Open-Source-Lizenz

### Lizenz: AGPL-3.0-or-later

jotti steht seit März 2026 unter der **AGPL-3.0-or-later** (GNU Affero General Public License v3). Zuvor war das Projekt unter MIT lizenziert; der Wechsel wurde vom alleinigen Urheber (Nico Gräf) durchgeführt.

### Warum AGPL?

| Kriterium                                          | MIT                                             | AGPL-3.0                                                                                                                 |
| -------------------------------------------------- | ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| Nutzung durch Vereine (intern, self-hosted)        | ✅ Frei                                         | ✅ Frei                                                                                                                  |
| Jemand forkt jotti und verkauft es als SaaS        | ✅ Erlaubt, keine Pflichten                     | ⚠️ Erlaubt, aber alle Änderungen müssen als Open Source veröffentlicht werden. Proprietäre Abspaltung ist nicht möglich. |
| Schutz vor proprietärer Abspaltung                 | ❌ Kein Schutz                                  | ✅ Copyleft schützt — wer jotti nutzt oder modifiziert, muss den Quellcode offenlegen                                    |
| Dual Licensing (kommerziell + Open Source) möglich | ✅ Ja, aber andere können das gleiche kostenlos | ✅ Ja, und nur der Urheber kann eine kommerzielle Alternative anbieten                                                   |
| Verbreitung als Self-Hosted-Software               | Kein Unterschied                                | Kein Unterschied                                                                                                         |
| Beiträge der Community fließen zurück              | ❌ Keine Pflicht                                | ✅ Pflicht (Copyleft)                                                                                                    |

### Was ändert sich für Vereine?

**Nichts.** AGPL-3.0 erlaubt die kostenlose Nutzung, Installation und Modifikation. Solange der Verein die Software intern betreibt (Self-Hosted), hat er keinerlei Pflichten außer der Beibehaltung des Copyright-Hinweises.

### Lizenzwechsel

Da Nico Gräf der alleinige Urheber ist, kann er die Lizenz jederzeit ändern. Der Wechsel von MIT zu AGPL-3.0 wurde im März 2026 vollzogen. Bereits unter MIT veröffentlichte Versionen bleiben unter MIT; alle neuen Versionen stehen unter AGPL-3.0.

### Schutz vor kommerzieller Ausbeutung

Die AGPL-3.0 stellt sicher:

1. **Niemand kann jotti forken und als proprietäres Produkt verkaufen.** Wer jotti modifiziert und als Netzwerkservice (SaaS) anbietet, muss den vollständigen Quellcode aller Änderungen unter AGPL-3.0 veröffentlichen.
2. **Nur der Urheber kann eine kommerzielle Lizenz anbieten.** Nico Gräf behält sich als alleiniger Urheber das Recht vor, jotti zusätzlich unter einer proprietären/kommerziellen Lizenz anzubieten (Dual Licensing). Dritte können das nicht.
3. **SaaS-Anbieter haben keine Schlupflöcher.** Anders als bei GPL/LGPL schließt die AGPL auch die Nutzung über ein Netzwerk ein (§ 13 AGPL — „Remote Network Interaction"). Ein reines Hosting von jotti ohne Code-Offenlegung ist nicht AGPL-konform.

---

## 3. Nutzungsvereinbarung für Vereine

### Warum eine Nutzungsvereinbarung?

Die Open-Source-Lizenz (AGPL) regelt die Software-Nutzung. Eine **separate Nutzungsvereinbarung** ist nicht zwingend nötig, aber empfehlenswert, um folgende Punkte unmissverständlich zu klären:

- Der Verein ist kein Auftraggeber.
- Es besteht kein Anspruch auf Support, Weiterentwicklung oder Verfügbarkeit.
- IP gehört dem Entwickler.
- Der Verein ist für Hosting, Betrieb und Datenschutz selbst verantwortlich.

Die vollständige Muster-Nutzungsvereinbarung findet sich in [Abschnitt 12](#12-muster-nutzungsvereinbarung).

### Kernpunkte (Kurzfassung)

| Punkt          | Regelung                                     |
| -------------- | -------------------------------------------- |
| IP/Eigentum    | Software ist Eigentum von Nico Gräf          |
| Lizenztyp      | Unentgeltlich, nicht-exklusiv, widerruflich  |
| Nutzungsumfang | Vereinsinterne Nutzung für den Kassenbetrieb |
| Gewährleistung | Keine — Software wird „as-is" bereitgestellt |
| Support        | Freiwillig und unverbindlich                 |
| Datenschutz    | Verantwortlicher ist der Verein              |
| Hosting        | Verein betreibt die Infrastruktur selbst     |
| Kündigung      | Jederzeit von beiden Seiten, ohne Frist      |

---

## 4. Sonderfall: Entwickler ist Vereinsmitglied

### Das Problem

Wenn der Entwickler gleichzeitig Mitglied des nutzenden Vereins ist, entsteht ein potenzieller Graubereich:

> Könnte der Verein argumentieren, die Software sei im Rahmen der Vereinstätigkeit entstanden und gehöre damit dem Verein?

### Die Antwort: Nein — aber Klarstellung ist wichtig

Im deutschen Recht gilt:

1. **Vereinsmitgliedschaft begründet kein Arbeitsverhältnis.** Ehrenamtliche Vereinsarbeit ist kein Dienst- oder Werkvertrag. Das Urheberrecht entsteht beim Schöpfer (§ 7 UrhG).

2. **Ausnahme wäre:** Wenn der Entwickler im Rahmen eines expliziten Vereinsamts (z.B. „IT-Beauftragter") und mit konkretem Vereinsauftrag entwickelt hätte. Aber selbst dann ist die Rechtslage bei Vereinen nicht so klar wie im Arbeitsrecht (§ 69b UrhG gilt nur für Arbeitnehmer).

3. **Faktenlage bei jotti:**
   - Entwickelt auf eigenem Rechner, in eigener Zeit, mit eigenen Werkzeugen.
   - Öffentliches Open-Source-Projekt auf GitHub unter Nico Gräfs persönlichem Account.
   - Kein Vereinsbeschluss zur Beauftragung.
   - Kein Austausch mit dem Verein während der Hauptentwicklung.
   - Der Verein ist einer von potenziell vielen Nutzern.

### Empfehlung: Schriftlich festhalten

Die Nutzungsvereinbarung (siehe [Abschnitt 12](#12-muster-nutzungsvereinbarung)) hält explizit fest:

> „Die Software ist ein persönliches Projekt des Entwicklers. Die Entwicklung erfolgt unabhängig von der Vereinsmitgliedschaft. Der Verein hat die Entwicklung weder beauftragt noch finanziert."

Damit ist der Sachverhalt dokumentiert — auch für zukünftige Vorstandswechsel.

### Was, wenn der Verein „Aufträge" erteilt?

Sobald der Verein spezifische Feature-Wünsche äußert, liegt die Entscheidung bei dir:

| Situation                                  | Empfehlung                                                                                                  |
| ------------------------------------------ | ----------------------------------------------------------------------------------------------------------- |
| Verein wünscht sich Feature X              | Du prüfst, ob es aufs Backlog passt. Kein Versprechen.                                                      |
| Verein sagt „Wir brauchen das bis Samstag" | Klar kommunizieren: „Ich bin kein Auftragnehmer. Ich kann es versuchen."                                    |
| Verein bietet Geld für ein Feature         | **Vorsicht.** Dann wird es ein Auftragsverhältnis. Entweder ablehnen oder sauber als Werkvertrag aufsetzen. |

---

## 5. Externe Vereine und Organisationen

### Szenario

Ein anderer Verein oder eine Organisation möchte jotti einsetzen. Du bist dort **kein Mitglied**.

### Rechtliche Lage: Deutlich einfacher

- Du bist ein Dritter, der Open-Source-Software veröffentlicht hat.
- Der Verein entscheidet sich, diese Software zu nutzen.
- Kein Graubereich bezüglich Vereinstätigkeit.
- Die AGPL-Lizenz regelt alles.

### Empfehlung: Gleiche Nutzungsvereinbarung

Auch für externe Vereine empfiehlt sich die Muster-Nutzungsvereinbarung aus [Abschnitt 12](#12-muster-nutzungsvereinbarung) — angepasst mit dem jeweiligen Vereinsnamen. Der Abschnitt zur Vereinsmitgliedschaft entfällt dann.

### Optionale Dienstleistungen

Für externe Vereine könntest du zusätzlich anbieten (gegen Bezahlung):

| Leistung                                  | Modell                                                                |
| ----------------------------------------- | --------------------------------------------------------------------- |
| Einrichtungshilfe (Setup, Docker, Server) | Einmaliges Festpreispaket                                             |
| Einweisung für Admins                     | Stundensatz oder Pauschale                                            |
| Hosting als Service                       | Monatliches Abo (dann bist du Auftragsverarbeiter → AV-Vertrag nötig) |
| Individuelle Anpassungen                  | Werkvertrag pro Feature                                               |

Das ist dann kein Hobby mehr, sondern Gewerbe — steuerlich beachten (Kleinunternehmerregelung oder Gewerbe).

---

## 6. Hosting und Betrieb

### Grundsatz: Der Verein hostet selbst

| Aspekt            | Regelung                                                  |
| ----------------- | --------------------------------------------------------- |
| Server            | Vom Verein betrieben (VPS, Raspberry Pi, lokaler Rechner) |
| Domain            | Vom Verein registriert und bezahlt                        |
| SSL-Zertifikat    | Let's Encrypt (automatisch, kostenlos)                    |
| Datenbank-Backups | Verantwortung des Vereins                                 |
| Software-Updates  | Verein entscheidet, wann aktualisiert wird                |

### Warum der Verein selbst hosten soll

1. **Kein Auftragsverarbeitungsvertrag (AV-Vertrag) nötig.** Wenn du die Daten nicht verarbeitest, bist du kein Auftragsverarbeiter im Sinne der DSGVO.
2. **Keine laufenden Kosten für dich.** Server, Domain, Traffic bezahlt der Verein.
3. **Keine Verfügbarkeitspflicht.** Wenn der Server ausfällt, ist das Problem des Vereins, nicht deins.
4. **Saubere Trennung.** Du stellst Software bereit, der Verein betreibt sie.

### Einrichtungshilfe

Du kannst dem Verein **beim Setup helfen** — das ändert nichts an der Eigentümerstruktur. Empfehlung:

- Docker Compose + Dokumentation bereitstellen (vorhanden).
- Einmalige Einrichtung zusammen mit dem Vereinsadmin durchführen.
- Danach übernimmt der Verein den Betrieb.

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

| Rolle (DSGVO)                       | Wer                        | Begründung                                                                |
| ----------------------------------- | -------------------------- | ------------------------------------------------------------------------- |
| **Verantwortlicher** (Art. 4 Nr. 7) | Der Verein                 | Der Verein entscheidet, welche Servicekräfte als Benutzer angelegt werden |
| **Auftragsverarbeiter** (Art. 28)   | Niemand (bei Self-Hosting) | Der Entwickler hat keinen Zugriff auf die Daten                           |
| **Betroffene Personen**             | Servicekräfte des Vereins  | Ihre Daten werden im System gespeichert                                   |

### Pflichten des Vereins

| Pflicht                                | Beschreibung                                                                                                                   |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| **Informationspflicht** (Art. 13/14)   | Servicekräfte darüber informieren, dass ihre Daten gespeichert werden (Name, Rolle, Aktivitäten)                               |
| **Löschpflicht** (Art. 17)             | Benutzer auf Wunsch deaktivieren/löschen (Soft-Delete vorhanden)                                                               |
| **Verarbeitungsverzeichnis** (Art. 30) | jotti als Verarbeitungstätigkeit dokumentieren (bei Vereinen unter 250 Mitarbeitern i.d.R. nicht Pflicht, aber empfehlenswert) |
| **Technische Maßnahmen** (Art. 32)     | Server absichern, HTTPS verwenden, Passwörter nicht teilen                                                                     |

### Pflichten des Entwicklers

Keine — solange du nicht hostest und keinen Zugriff auf die Produktivdaten hast.

**Wichtig:** Wenn der Verein dich bittet, „kurz auf den Server zu schauen" und du dafür SSH-Zugriff bekommst, bist du potenziell Auftragsverarbeiter. Empfehlung: Nur per Screensharing helfen, nicht selbst auf den Server zugreifen.

---

## 8. Haftung und Gewährleistung

### Grundsatz: Keine Haftung

Die AGPL-3.0-Lizenz enthält bereits einen Haftungsausschluss:

> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND.

Die Nutzungsvereinbarung bekräftigt:

- Keine Gewährleistung auf Funktionalität, Richtigkeit oder Verfügbarkeit.
- Kein Anspruch auf Fehlerbehebung.
- Der Verein nutzt die Software auf eigenes Risiko.

### Reale Risikoszenarien

| Szenario                                              | Risiko       | Einschätzung                                                                                                          |
| ----------------------------------------------------- | ------------ | --------------------------------------------------------------------------------------------------------------------- |
| Software berechnet Saldo falsch, Verein verliert Geld | Niedrig      | Event-Sourcing ist mathematisch konsistent (Cent-Beträge, keine Floats). Immutable Events verhindern Datenkorruption. |
| Datenbank geht verloren                               | Mittel       | Verantwortung des Vereins (Backups). Empfehlung: Regelmäßige PostgreSQL-Dumps dokumentieren.                          |
| Sicherheitslücke, Daten werden geleakt                | Niedrig      | Lokales Netz, JWT + Argon2id, keine Gästedaten. Angriffsfläche gering.                                                |
| Software fällt mitten im Fest aus                     | Mittel       | Kein SLA. Empfehlung: Testlauf vor dem Fest, Fallback-Plan (Stift & Papier).                                          |
| Verein verklagt Entwickler                            | Sehr niedrig | Kostenlose Nutzung, Haftungsausschluss, kein Vertragsverhältnis. Kaum Anspruchsgrundlage.                             |

---

## 9. Support und Erwartungsmanagement

### Das größte reale Risiko ist kein rechtliches — es ist ein menschliches.

Sobald der Verein jotti einsetzt, können folgende Situationen entstehen:

| Situation                                     | Was passiert                      | Empfohlene Reaktion                                                                       |
| --------------------------------------------- | --------------------------------- | ----------------------------------------------------------------------------------------- |
| „Das muss bis Samstag funktionieren"          | Zeitdruck vor dem Vereinsfest     | „Ich gebe mein Bestes, aber verspreche nichts. Testet vorher."                            |
| „Kannst du noch schnell X einbauen?"          | Feature Creep                     | „Steht auf der Wunschliste. Kommt, wenn es kommt."                                        |
| „Das ist kaputt, reparier das sofort"         | Bug im Live-Betrieb               | „Ich schau es mir an, wenn ich Zeit habe. Für heute: Stift & Papier als Fallback."        |
| „Der Nico macht das doch"                     | Du wirst als IT-Abteilung gesehen | Grenzen setzen. Du bist Mitglied, nicht Angestellter.                                     |
| Neuer Vorstand kennt die Absprache nicht      | Wissenstransfer-Problem           | Nutzungsvereinbarung schriftlich beim Verein hinterlegen.                                 |
| Verein möchte Änderungen, die du nicht willst | Interessenkonflikt                | „Das Projekt hat ein öffentliches Backlog. Vorschläge gerne, Entscheidung liegt bei mir." |

### Grundregel

> **Du bist Entwickler eines Open-Source-Projekts, das der Verein nutzen darf — nicht der IT-Dienstleister des Vereins.**

### Praktische Empfehlungen

1. **Vor dem ersten Einsatz** ein kurzes Gespräch mit dem Vorstand führen und die Nutzungsvereinbarung unterschreiben lassen.
2. **Ansprechperson im Verein** benennen (Admin/Thomas im Event-Storming-Szenario), die das System betreut — nicht du.
3. **Testlauf** vor dem echten Fest. Nicht am Tag des Maihocks das erste Mal live gehen.
4. **Fallback definieren.** Was passiert, wenn jotti ausfällt? Zettel + Stift müssen bereitliegen.

---

## 10. Kommerzialisierung und Dual Licensing

### Ausgangslage

Als alleiniger Urheber besitzt Nico Gräf das Recht, jotti unter beliebig vielen Lizenzen gleichzeitig zu veröffentlichen. Dies ermöglicht ein **Dual-Licensing-Modell:**

| Pfad            | Lizenz             | Zielgruppe                                  | Kosten          |
| --------------- | ------------------ | ------------------------------------------- | --------------- |
| **Open Source** | AGPL-3.0           | Vereine, Non-Profit, Self-Hosted            | Kostenlos       |
| **Kommerziell** | Proprietäre Lizenz | Unternehmen, SaaS-Anbieter, Gastro-Betriebe | Kostenpflichtig |

### Vorbehaltene Rechte

Nico Gräf behält sich ausdrücklich folgende Rechte vor:

1. **Kostenpflichtige Nutzungslizenzen anbieten** — z.B. für kommerzielle Gastro-Betriebe, die jotti ohne AGPL-Pflichten nutzen möchten.
2. **jotti als kostenpflichtiges SaaS-Produkt betreiben** — als gehostete Lösung mit Support, SLA und Wartung.
3. **Einzelne Vereine/Organisationen von der kostenlosen Nutzung auszunehmen** — z.B. wenn ein Verein die Software gewerblich weitervermarktet.
4. **Die Lizenz zukünftiger Versionen zu ändern** — solange bestehende AGPL-veröffentlichte Versionen unter AGPL bleiben.

Der Heimat-/Erstverein [VEREINSNAME] e.V. erhält eine **dauerhafte, unentgeltliche Nutzungszusage** gemäß der Nutzungsvereinbarung in [Abschnitt 12](#12-muster-nutzungsvereinbarung).

### Warum das funktioniert

- **AGPL-Copyleft** verlangt, dass jeder, der jotti als Netzwerkservice (SaaS) anbietet, den vollständigen Quellcode aller Änderungen unter AGPL veröffentlichen muss. Die meisten kommerziellen Anbieter wollen das nicht.
- **Einziger Ausweg für kommerzielle Nutzung ohne Offenlegungspflicht:** Eine kommerzielle Lizenz vom Urheber kaufen — also von Nico Gräf.
- **Vereine, die self-hosten, sind nicht betroffen.** Interne Nutzung löst die Copyleft-Pflicht nicht aus.

### Mögliche Kommerzialisierungsmodelle (Zukunft)

| Modell                   | Beschreibung                                                                                            |
| ------------------------ | ------------------------------------------------------------------------------------------------------- |
| **Hosting-as-a-Service** | jotti als gehostete Lösung für Vereine (monatliches Abo). Dann: Auftragsverarbeiter → AV-Vertrag nötig. |
| **Setup-Pakete**         | Einmalige Einrichtung + Einweisung gegen Festpreis.                                                     |
| **Support-Verträge**     | Garantierte Reaktionszeiten, Hotline am Festtag.                                                        |
| **Enterprise-Lizenz**    | Kommerzielle Gastro-Betriebe zahlen eine Lizenzgebühr für die Nutzung ohne AGPL-Pflichten.              |
| **White-Label-Lizenz**   | Dritte dürfen jotti unter eigenem Namen anbieten — nur mit kommerzieller Lizenz.                        |

### Wichtig: Community-Beiträge

Wenn andere Entwickler Pull Requests einreichen, stehen deren Beiträge unter AGPL. Um weiterhin Dual Licensing anbieten zu können, solltest du entweder:

- Ein **Contributor License Agreement (CLA)** verwenden, das dir das Recht zur Relizenzierung einräumt, oder
- Beiträge nur von dir selbst committen (bei einem Hobby-Projekt realistisch).

---

## 11. Wirtschaftlicher Hintergrund

### Was wäre jotti als kommerzielle Entwicklung wert?

Um den tatsächlichen Wert der unentgeltlich bereitgestellten Software einzuordnen, wurden drei Budgetszenarien durchgerechnet. Grundlage: 42 Anforderungen (31 Must-have + 11 Nice-to-have), Neuentwicklung von Grund auf.

| Szenario                                    | Personentage | Budget (netto) | Laufzeit |
| ------------------------------------------- | ------------ | -------------- | -------- |
| **Software-Agentur** (konservativ, 3,5 FTE) | 274 PT       | ~250.000 €     | 9 Monate |
| **Software-Agentur** (optimiert, 2,6 FTE)   | 160 PT       | ~145.000 €     | 7 Monate |
| **Senior Freelancer** (Solo, 900 €/Tag)     | 133 PT       | ~122.000 €     | 8 Monate |

### Einordnung

Der Verein erhält ein System, dessen kommerzielle Neuentwicklung **zwischen 120.000 € und 250.000 €** kosten würde.

Diese Zahlen beruhen auf marktüblichen Tagessätzen (750–1.050 €/PT) und einer vollständigen Umsetzung aller Must-have- und Nice-to-have-Anforderungen. Die reinen Entwicklungskosten (ohne PM, QA, DevOps) liegen bei ~100.000 € im günstigsten Szenario.

### Warum diese Zahl relevant ist

1. **Wertschätzung:** Der Verein sollte verstehen, dass er ein Geschenk in erheblichem Umfang erhält.
2. **Erwartungsmanagement:** Bei einem kostenfreien Produkt mit sechsstelligem Entwicklungswert darf der Verein keine Agentur-Level-Erwartungen an Support und Deadlines stellen.
3. **Dokumentation:** Falls jemals die Frage aufkommt, ob der Entwickler „angemessen entlohnt" wurde — die Antwort ist: Es gab keine Entlohnung, und es bestand kein Anspruch darauf.

---

## 12. Muster-Nutzungsvereinbarung

Die folgende Vereinbarung kann als vorlage verwendet und angepasst werden. Sie ist **kein Rechtsberatungsprodukt** und ersetzt keine anwaltliche Prüfung.

---

### Nutzungsvereinbarung für die Software „jotti"

**Zwischen:**

- **Entwickler:** Nico Gräf (nachfolgend „Entwickler")
- **Nutzer:** [VEREINSNAME] e.V. (nachfolgend „Verein")

**Stand:** \_\_\_\_\_\_\_\_\_\_\_\_\_

---

#### § 1 Gegenstand

(1) Der Entwickler stellt dem Verein die Software „jotti" (nachfolgend „Software") zur unentgeltlichen Nutzung zur Verfügung.

(2) Die Software ist ein persönliches Open-Source-Projekt des Entwicklers. Die Entwicklung erfolgt unabhängig von der Vereinsmitgliedschaft des Entwicklers. Der Verein hat die Entwicklung weder beauftragt noch finanziert.

(3) Der Quellcode ist öffentlich zugänglich unter: https://github.com/nicograef/jotti

#### § 2 Geistiges Eigentum

(1) Die Software — einschließlich Quellcode, Dokumentation, Architektur und Design — ist das alleinige geistige Eigentum des Entwicklers.

(2) Durch die Nutzung der Software erwirbt der Verein keinerlei Rechte an der Software über die in § 3 gewährte Lizenz hinaus.

(3) Der Entwickler behält sich das Recht vor, die Software jederzeit unter beliebigen weiteren Lizenzen zu veröffentlichen, zu vermarkten oder anderweitig zu verwerten.

#### § 3 Nutzungslizenz

(1) Der Verein erhält eine **unentgeltliche, nicht-exklusive, widerrufliche** Lizenz zur Nutzung der Software für den vereinsinternen Kassenbetrieb.

(2) Die Nutzung umfasst: Installation, Konfiguration, Betrieb und Aktualisierung auf vom Verein betriebener Infrastruktur.

(3) Es gelten zusätzlich die Bestimmungen der Open-Source-Lizenz (AGPL-3.0-or-later), unter der die Software veröffentlicht ist.

#### § 4 Hosting und Betrieb

(1) Der Verein betreibt die Software auf eigener Infrastruktur (Server, Datenbank, Netzwerk).

(2) Der Entwickler stellt weder Hosting noch Infrastruktur bereit.

(3) Der Entwickler kann freiwillig bei der Ersteinrichtung unterstützen. Daraus entsteht kein Anspruch auf laufende Betreuung.

#### § 5 Datenschutz

(1) Der Verein ist **Verantwortlicher** im Sinne der DSGVO (Art. 4 Nr. 7) für alle in der Software verarbeiteten personenbezogenen Daten.

(2) Der Entwickler ist **kein Auftragsverarbeiter**, da er weder Zugriff auf den Server noch auf die gespeicherten Daten hat.

(3) Der Verein ist verpflichtet, die einschlägigen Datenschutzbestimmungen einzuhalten, insbesondere die Information der Betroffenen (Servicekräfte) über die Datenverarbeitung.

#### § 6 Gewährleistung und Haftung

(1) Die Software wird **ohne Gewährleistung** bereitgestellt („as-is"). Es besteht kein Anspruch auf Fehlerfreiheit, Funktionalität, Verfügbarkeit oder Eignung für einen bestimmten Zweck.

(2) Der Entwickler haftet nicht für Schäden, die aus der Nutzung der Software entstehen — insbesondere nicht für Datenverlust, Fehlberechnungen oder Ausfälle.

(3) Der Verein nutzt die Software auf eigenes Risiko.

#### § 7 Support und Weiterentwicklung

(1) Es besteht **kein Anspruch** auf Support, Fehlerbehebung, Weiterentwicklung oder Aktualisierung.

(2) Unterstützung durch den Entwickler erfolgt freiwillig und unverbindlich.

(3) Feature-Wünsche des Vereins werden zur Kenntnis genommen, begründen jedoch keinen Umsetzungsanspruch.

#### § 8 Laufzeit und Kündigung

(1) Diese Vereinbarung gilt auf unbestimmte Zeit.

(2) Beide Seiten können die Vereinbarung jederzeit ohne Angabe von Gründen und ohne Einhaltung einer Frist beenden.

(3) Bei Beendigung erlischt die Nutzungslizenz nach § 3 Abs. 1. Die Rechte aus der Open-Source-Lizenz (AGPL-3.0) bleiben unberührt.

#### § 9 Schlussbestimmungen

(1) Änderungen und Ergänzungen dieser Vereinbarung bedürfen der Schriftform.

(2) Sollten einzelne Bestimmungen unwirksam sein, bleibt die Wirksamkeit der übrigen Bestimmungen unberührt.

---

**Ort, Datum:** \_\_\_\_\_\_\_\_\_\_\_\_\_

**Entwickler:** \_\_\_\_\_\_\_\_\_\_\_\_\_ (Nico Gräf)

**Verein:** \_\_\_\_\_\_\_\_\_\_\_\_\_ (vertretungsberechtigte Person, [VEREINSNAME] e.V.)

---

> **Hinweis:** Diese Vereinbarung ist ein Muster und keine Rechtsberatung. Bei Unsicherheiten empfiehlt sich die Prüfung durch eine Rechtsanwältin oder einen Rechtsanwalt.
