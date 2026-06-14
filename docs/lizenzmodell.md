# Lizenzmodell: jotti

Dieses Dokument erklärt das Lizenzmodell von jotti: Eigentum, Source-Available-Prinzip, berechtigte Nutzer, Kommerzialisierung und Community-Beiträge. Verbindlich sind die [Nutzungsbedingungen (TERMS.md)](../TERMS.md) und die [LICENSE](../LICENSE)-Datei. Dieses Dokument fasst zusammen und verweist.

---

## 1. Eigentumsverhältnisse

Die Software jotti (Quellcode, Dokumentation, Architektur, Design und alle zugehörigen Artefakte) ist das alleinige geistige Eigentum von Nico Gräf (Freiburg im Breisgau). Alle Rechte vorbehalten; der Autor entscheidet allein über Lizenzierung, Weiterentwicklung und Verbreitung. Die Veröffentlichung des Quellcodes auf [github.com/nicograef/jotti](https://github.com/nicograef/jotti) stellt keine automatische Rechteeinräumung dar. Beiträge Dritter unterliegen dem CLA (→ [Abschnitt 6](#6-community-beiträge-und-cla)).

## 2. Source-Available, nicht Open Source

jotti steht unter einer proprietären Source-Available-Lizenz: Der Quellcode ist öffentlich einsehbar, aber Nutzungsrechte werden nicht automatisch gewährt. Jede Nutzung erfordert eine vorherige Nutzungsvereinbarung in Textform; Ablauf und E-Mail-Vorlage: [TERMS.md → Prozess](../TERMS.md#prozess-nutzungsvereinbarung-abschließen).

| Eigenschaft                       | Source-Available (jotti) | Open Source (z. B. MIT, Apache) |
| --------------------------------- | ------------------------ | ------------------------------- |
| Quellcode einsehbar               | ✅ Ja                    | ✅ Ja                           |
| Nutzung automatisch erlaubt       | ❌ Nein                  | ✅ Ja                           |
| Modifikation erlaubt              | ❌ Nein                  | ✅ Ja                           |
| Weitergabe erlaubt                | ❌ Nein                  | ✅ Ja                           |
| Kommerzielle Nutzung              | ❌ Verboten              | ✅ / ⚠️ Je nach Lizenz          |
| Nutzungsvereinbarung erforderlich | ✅ Ja                    | ❌ Nein                         |
| OSI-zertifiziert                  | ❌ Nein                  | ✅ Ja                           |

Ohne Nutzungsvereinbarung sind genau zwei Aktivitäten gestattet:

1. Ansehen und Lesen des Quellcodes, zu Bildungszwecken, technischer Evaluation oder Sicherheitsüberprüfung.
2. Einreichen von Beiträgen (Pull Requests) an das offizielle Repository, unter den Bedingungen des CLA.

Alles andere (Installation, Deployment, Hosting, Ausführung, Kopieren, Modifizieren, Weitergabe) erfordert eine Nutzungsvereinbarung. Die Veröffentlichung dient Transparenz, Vertrauen und Sicherheitsüberprüfung, nicht der freien Nutzung; jotti ist kein Open-Source-Projekt im Sinne der OSI-Definition.

## 3. Berechtigte: kostenlose Nutzung

Kostenlose Nutzungsvereinbarungen erhalten ausschließlich eingetragene Organisationen ohne Gewinnerzielungsabsicht (verbindlich: [TERMS.md § 2](../TERMS.md)):

| Organisationsform                         | Voraussetzung                                                        |
| ----------------------------------------- | -------------------------------------------------------------------- |
| Eingetragene Vereine (e.V.)               | Eintragung im Vereinsregister gemäß §§ 21 ff. BGB                    |
| Eingetragene gemeinnützige Stiftungen     | Stiftungsregister + Anerkennung der Gemeinnützigkeit                 |
| Gemeinnützige GmbH / UG (gGmbH, gUG)       | Handelsregister + steuerliche Anerkennung der Gemeinnützigkeit       |
| Sonstige eingetragene NGOs / NPOs         | Nachweisbare Registereintragung und fehlende Gewinnerzielungsabsicht |

Keine kostenlose Vereinbarung erhalten gewerbliche Unternehmen, Organisationen ohne gemeinnützigen Status, gewerblich nutzende Einzelpersonen und Dritte, die jotti als Dienstleistung betreiben wollen (SaaS). Sie benötigen eine kostenpflichtige kommerzielle Lizenz (→ [Abschnitt 5](#5-kommerzialisierung-und-dual-licensing)). Entfallen die Voraussetzungen, endet die kostenlose Nutzungslizenz automatisch ([TERMS.md § 2 Abs. 3](../TERMS.md)).

## 4. Forks, Modifikation und Weitergabe

| Aktivität                                                    | Erlaubt?          |
| ------------------------------------------------------------ | ----------------- |
| Quellcode auf GitHub lesen                                   | ✅ Ja             |
| Repository forken, um einen Pull Request einzureichen        | ✅ Ja (unter CLA) |
| Repository forken für eigene Nutzung                         | ❌ Nein           |
| Code modifizieren (außer für PR an das offizielle Repo)      | ❌ Nein           |
| Code weitergeben, veröffentlichen oder einbinden             | ❌ Nein           |
| Code als eigenständige Software betreiben                    | ❌ Nein           |

Diese Restriktionen schützen das geistige Eigentum, halten das Dual-Licensing-Modell offen und sichern die Kontrolle über Verbreitung und Haftungssituation des Autors.

## 5. Kommerzialisierung und Dual Licensing

Als alleiniger Urheber kann Nico Gräf jotti unter beliebig vielen Lizenzen gleichzeitig anbieten:

| Pfad        | Lizenz                                                         | Zielgruppe                                         | Kosten          |
| ----------- | -------------------------------------------------------------- | -------------------------------------------------- | --------------- |
| Non-Profit  | Proprietäre Source-Available + kostenlose Nutzungsvereinbarung | Eingetragene Vereine, gemeinnützige Organisationen | Kostenlos       |
| Kommerziell | Proprietäre kommerzielle Lizenz                                | Unternehmen, gewerbliche Betriebe, SaaS-Anbieter   | Kostenpflichtig |

Ausdrücklich vorbehalten bleiben: kostenpflichtige Nutzungslizenzen, ein SaaS-Angebot durch den Autor (Hosting, Setup-Pakete, Support-Verträge, White-Label), der Ausschluss einzelner Organisationen von der kostenlosen Nutzung, die Änderung der Lizenz zukünftiger Versionen sowie der Widerruf kostenloser Nutzungsvereinbarungen mit angemessener Frist.

## 6. Community-Beiträge und CLA

Jeder Beitrag (Pull Request, Patch, Code-Einreichung) unterliegt dem [Contributor License Agreement (CLA.md)](../CLA.md): Der Beitragende behält das Urheberrecht an seinem Beitrag, gewährt dem Autor aber eine unwiderrufliche, weltweite, gebührenfreie, nicht-exklusive Lizenz einschließlich Relizenzierung und bestätigt, dass der Beitrag seine eigene Schöpfung ist. Ohne CLA könnten Beiträge nicht in kommerzielle Lizenzen aufgenommen werden; es ist die Voraussetzung des Dual-Licensing-Modells.

## 7. Verbindliche Regelungen in den Nutzungsbedingungen

Hosting und Betrieb, Datenschutz (DSGVO), Gewährleistungsausschluss, Haftungsbegrenzung, Freistellung und Compliance-Verantwortung (KassenSichV / TSE) sind abschließend in [TERMS.md §§ 5–10](../TERMS.md) geregelt. Kurzfassung: Die nutzende Organisation ist alleinige Betreiberin und datenschutzrechtlich Verantwortliche; der Autor stellt ausschließlich Quellcode bereit („as-is", Haftung nach Schenkungsrecht § 521 BGB nur für Vorsatz und grobe Fahrlässigkeit) und wird von Ansprüchen Dritter freigestellt. Die fachlichen Compliance-Pflichten der Betreiber beschreibt [compliance.md](compliance.md).
