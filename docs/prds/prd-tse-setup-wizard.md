# PRD: TSE-Setup-Wizard

> Herkunft: ausgelagert aus Phase 5 von `docs/plans/plan-tse-fiskaly-fixes.md` (Entscheidung 2026-06-11, dort unter „Resolved decisions" dokumentiert)
> Quellen: Audit `docs/audits/2026-06-11-tse-fiskaly-audit.md` (Findings I-09, I-15.4, D-07)
> Ersetzt die BYOT-Festlegung aus `docs/prds/prd-tse-integration.md` („Betreiber legt TSS und Client im fiskaly-Dashboard an; jotti betreibt keinen TSS-Lebenszyklus") — siehe Further Notes.

## Problem Statement

jotti wird von Vereinen eingesetzt, deren Admins und Helfer keine technischen Vorkenntnisse haben. Um die TSE in Betrieb zu nehmen, müssen sie heute den fiskaly-TSS-Lebenszyklus manuell durchlaufen: TSS anlegen, Admin-PUK abholen, Admin-PIN setzen, TSS initialisieren und den Client registrieren. **Keiner dieser Schritte ist im fiskaly-Dashboard möglich** — das Dashboard listet TSS nur und kann weder Clients anlegen noch TSS initialisieren (verifiziert per fiskaly-Support-Doku). Der einzige Weg führt über direkte API-Aufrufe (Postman/curl) mit zwei verschiedenen Authentifizierungs-Flows. Für die Zielgruppe ist das faktisch eine Blockade: Ohne technische Hilfe kann ein Verein jotti nicht fiskalkonform in Betrieb nehmen.

Dazu kommen drei Audit-Findings:

- **I-09:** Der fiskaly-Client muss mit jottis Kassen-Seriennummer als `serial_number` registriert werden, sonst widersprechen sich gedruckter Beleg und QR-Code und die Belegprüfung (Amtsträger-App) kann den Bon keiner gemeldeten Kasse zuordnen. Heute wird das weder erzwungen noch geprüft.
- **I-15.4:** Der Verbindungstest prüft nur den TSS-Zustand; ein deregistrierter Client fällt erst beim Signieren auf — im schlimmsten Fall mitten im Fest.
- **D-07:** Es gibt keine Betreiber-Dokumentation für die TSE-Einrichtung.

Erschwerend: Eine TSS-Anlage in der LIVE-Umgebung ist kostenpflichtig und nicht rückgängig zu machen (TSS sind nicht löschbar, nur deaktivierbar). Ein Einrichtungsweg für Laien muss daher Doppel-Anlagen verhindern und die Umgebung (TEST/LIVE) unübersehbar machen.

## Solution

jotti erhält eine eigene Admin-Seite **„TSE-Einrichtung"** mit einem geführten Wizard. Der Vereins-Admin erstellt — angeleitet durch den neuen Betreiber-Leitfaden — im fiskaly-Dashboard nur noch das Konto und einen API-Key. Alles Weitere übernimmt jotti:

1. **Zugangsdaten:** Admin gibt API-Key und API-Secret ein.
2. **Prüfung (ohne Seiteneffekte):** jotti meldet sich an, zeigt die erkannte Umgebung (TEST/LIVE) deutlich an und listet bereits vorhandene TSS samt Zustand.
3. **Auswahl & Bestätigung:** Der Admin übernimmt eine vorhandene TSS (Schutz vor Doppel-Anlage, Wiederaufnahme nach Abbruch) oder legt bewusst eine neue an. Vor einer kostenwirksamen TSS-Anlage in LIVE muss er das Wort „LIVE" eintippen; in TEST genügt ein Klick.
4. **Durchführung:** jotti führt automatisch die fehlenden Schritte aus — TSS anlegen, PUK abholen, zufällig erzeugte Admin-PIN setzen, TSS initialisieren, Client mit der Kassen-Seriennummer als `serial_number` registrieren.
5. **PUK/PIN-Übergabe:** Admin-PUK und Admin-PIN werden genau einmal angezeigt; der Admin bestätigt, dass er sie extern sicher verwahrt hat. jotti speichert sie nicht.
6. **Abschluss:** Die vollständige Konfiguration (API-Key, API-Secret, TSS-ID, Client-ID) wird gespeichert und ein erweiterter Verbindungstest bestätigt die Signierfähigkeit.

Der Wizard ist **wiederaufsetzbar**: Nach einem Abbruch erkennt die Prüfung den tatsächlichen Zustand bei fiskaly und holt nur die fehlenden Schritte nach. Die manuelle Eingabe von TSS-ID/Client-ID bleibt als Experten-/Fallback-Weg erhalten und zieht auf die neue Seite um; die TSE-Sektion in den Einstellungen zeigt künftig nur noch den Status und verlinkt auf die Einrichtungsseite.

Der **Verbindungstest** prüft zusätzlich den Client-Zustand (`REGISTERED`) und die Übereinstimmung der fiskaly-`serial_number` mit der Kassen-Seriennummer — eine Abweichung ist ein Fehler (schließt I-09 und I-15.4 unabhängig vom gewählten Einrichtungsweg). Der **Betreiber-Leitfaden** (D-07) beschreibt den Gesamtablauf inklusive der Schritte außerhalb von jotti.

## User Stories

### Vereins-Admin — Einrichtung

1. Als Vereins-Admin möchte ich die TSE vollständig über die jotti-Oberfläche einrichten, ohne API-Werkzeuge wie Postman oder curl, damit ich keine technischen Vorkenntnisse brauche.
2. Als Vereins-Admin möchte ich eine eigene Seite „TSE-Einrichtung" mit einer geführten Schritt-für-Schritt-Strecke, damit ich jederzeit sehe, wo ich im Prozess stehe und was als Nächstes passiert.
3. Als Vereins-Admin möchte ich nur API-Key und API-Secret eingeben müssen, damit alle weiteren fiskaly-Schritte automatisch ablaufen.
4. Als Vereins-Admin möchte ich, dass jotti die TSS anlegt, initialisiert und den Client registriert, damit ich den fiskaly-Lebenszyklus (PUK, PIN, Zustände) nicht verstehen muss.
5. Als Vereins-Admin möchte ich, dass der Client automatisch mit jottis Kassen-Seriennummer als `serial_number` registriert wird, damit Beleg und QR-Code zusammenpassen, ohne dass ich etwas abtippe.
6. Als Vereins-Admin möchte ich nach der Einrichtung Admin-PUK und Admin-PIN genau einmal angezeigt bekommen, inklusive Kopier-Funktion und der Erklärung, wofür ich sie brauche, damit ich sie für spätere Verwaltungsaufgaben sicher verwahren kann.
7. Als Vereins-Admin möchte ich bestätigen müssen, dass ich PUK und PIN verwahrt habe, bevor der Wizard abschließt, damit ich den Hinweis nicht versehentlich wegklicke.
8. Als Vereins-Admin möchte ich, dass am Ende automatisch ein Verbindungstest läuft und mir den Erfolg klar bestätigt, damit ich weiß, dass die Kasse signierfähig ist.
9. Als Vereins-Admin möchte ich bei jedem Fehler eine verständliche deutsche Meldung mit konkretem nächsten Schritt sehen, damit ich nicht in technischen Fehlercodes stecken bleibe.

### Vereins-Admin — Umgebung und Kosten

10. Als Vereins-Admin möchte ich in jedem Wizard-Schritt deutlich sehen, ob meine Zugangsdaten zur TEST- oder LIVE-Umgebung gehören, damit ich keinen Echtbetrieb gegen TEST einrichte (oder umgekehrt).
11. Als Vereins-Admin möchte ich vor der Anlage einer LIVE-TSS das Wort „LIVE" eintippen müssen, damit ich nicht versehentlich eine kostenpflichtige, nicht löschbare TSS anlege.
12. Als Vereins-Admin möchte ich in der TEST-Umgebung ohne zusätzliche Hürde fortfahren können, damit Ausprobieren und Üben reibungslos bleibt.
13. Als Vereins-Admin möchte ich vor einer Neuanlage gewarnt werden, wenn im Konto bereits eine aktive TSS existiert, damit keine zweite TSS aus Versehen entsteht.

### Vereins-Admin — vorhandene TSS und Wiederaufnahme

14. Als Vereins-Admin möchte ich vorhandene TSS aus meinem fiskaly-Konto mit ihrem Zustand angezeigt bekommen und eine davon übernehmen können, damit ein abgebrochenes Setup, eine Neuinstallation oder ein früheres manuelles Setup nicht zur Doppel-Anlage führt.
15. Als Vereins-Admin möchte ich, dass der Wizard auf einer übernommenen TSS einen vorhandenen Client mit passender Kassen-Seriennummer erkennt und übernimmt, statt einen neuen anzulegen.
16. Als Vereins-Admin möchte ich, dass der Wizard nach einem Abbruch (Netzfehler, Browser geschlossen) beim nächsten Aufruf am tatsächlichen fiskaly-Zustand weitermacht statt von vorn, damit kein halbfertiger Zustand zurückbleibt.
17. Als Vereins-Admin möchte ich bei der Übernahme einer bereits initialisierten TSS nach der verwahrten Admin-PIN gefragt werden, damit jotti den Client darauf registrieren kann.
18. Als Vereins-Admin möchte ich eine verständliche Erklärung mit Auswegen (fiskaly-Support, bewusste Neuanlage), wenn die Übernahme an einer unbekannten Admin-PIN scheitert, damit ich nicht in einer Sackgasse lande.

### Admin — Experten-Weg und Status

19. Als technisch versierter Admin möchte ich TSS-ID und Client-ID weiterhin manuell eintragen können (als Experten-Bereich auf der Einrichtungsseite), damit extern verwaltete oder bestehende TSS nutzbar bleiben.
20. Als Admin möchte ich in den Einstellungen den TSE-Status sehen und direkt zur Einrichtungsseite gelangen, damit das gesamte TSE-Setup an einem Ort gebündelt ist.
21. Als Admin möchte ich die Konfiguration weiterhin leeren oder ändern können (Schlüsselrotation, Wechsel TEST→LIVE), damit ich die Kontrolle behalte.

### Admin — Verbindungstest

22. Als Admin möchte ich, dass der Verbindungstest auch den Client-Zustand (`REGISTERED`) prüft, damit ein deregistrierter Client vor dem Fest auffällt und nicht erst beim Signieren.
23. Als Admin möchte ich, dass der Verbindungstest die fiskaly-`serial_number` gegen die Kassen-Seriennummer prüft und eine Abweichung als Fehler meldet, damit Beleg und QR-Code garantiert zusammenpassen — auch nach manueller Einrichtung.
24. Als Admin möchte ich das Testergebnis verständlich aufgeschlüsselt sehen (Umgebung, TSS-Zustand, Client-Zustand, Seriennummern-Abgleich), damit ich Probleme gezielt beheben kann.

### Vereins-Admin — Betreiber-Leitfaden

25. Als Vereins-Admin möchte ich einen Schritt-für-Schritt-Leitfaden für die Schritte außerhalb von jotti (fiskaly-Konto registrieren, Organisation anlegen, API-Key erstellen), damit ich auch diesen Teil ohne Vorkenntnisse schaffe.
26. Als Vereins-Admin möchte ich im Leitfaden erklärt bekommen, wie ich von TEST auf LIVE wechsle und welche Kosten dabei entstehen, damit ich den Echtbetrieb informiert vorbereite.
27. Als Vereins-Admin möchte ich im Leitfaden klare Anweisungen zur Verwahrung von Admin-PUK und Admin-PIN (und die Folgen eines Verlusts), damit ich die Verantwortung verstehe.
28. Als Vereins-Admin möchte ich im Leitfaden beide Einrichtungswege (Wizard und manuell) beschrieben finden, damit auch Sonderfälle abgedeckt sind.

### Betriebsprüfer

29. Als Betriebsprüfer möchte ich, dass die im QR-Code enthaltene Client-Seriennummer der auf dem Beleg gedruckten Kassen-Seriennummer entspricht, damit die Belegprüfung den Bon der gemeldeten Kasse zuordnen kann.

### Entwickler / System

30. Als System möchte ich die fiskaly-Setup-Operationen (TSS listen/anlegen, Zustandsübergänge, Admin-PIN setzen, Admin-Authentifizierung, Client anlegen/listen) in einem Modul kapseln, das die Admin-Token-Verwaltung intern erledigt, damit der Rest des Systems nur semantische Operationen sieht.
31. Als System möchte ich vor jeder Einrichtung einen seiteneffektfreien Prüf-Schritt ausführen (Umgebung, vorhandene TSS und Clients, passender Client zur Kassen-Seriennummer), dessen Befund die UI anzeigt, damit keine kostenwirksame Aktion ohne vorherige Anzeige und Bestätigung passiert.
32. Als System möchte ich die Einrichtung als wiederaufsetzbare Schrittfolge ausführen, die anhand des fiskaly-Ist-Zustands nur fehlende Schritte nachholt, damit Teilfehler idempotent heilbar sind.
33. Als System möchte ich die Admin-PIN zufällig erzeugen und PUK und PIN ausschließlich in der Antwort an die UI zurückgeben — niemals persistieren und niemals loggen —, damit die Entscheidung „einmalige Anzeige statt Speicherung" durchgängig gilt.
34. Als System möchte ich die Admin-PUK einer TSS im Zustand CREATED über den idempotenten Anlage-Request erneut beziehen, damit eine Wiederaufnahme in frühen Phasen ohne Nutzereingabe funktioniert.
35. Als System möchte ich die vom Admin bestätigte Umgebung beim Einrichten mitgeben und abbrechen, wenn die tatsächliche Umgebung abweicht, damit zwischenzeitlich gewechselte Zugangsdaten nicht zu Aktionen in der falschen Umgebung führen.
36. Als System möchte ich die TSE-Konfiguration erst nach erfolgreichem Abschluss vollständig speichern (alle vier Felder zusammen), damit nie eine halbfertige Konfiguration in der Datenbank liegt.
37. Als Entwickler möchte ich den Setup-Orchestrator gegen einen Fake-Setup-Client testen können, damit alle Pfade (leeres Konto, vorhandene TSS, Wiederaufnahme, PIN-Nachfrage, Umgebungs-Abweichung) ohne echten fiskaly-Zugang abgedeckt sind.
38. Als Entwickler möchte ich einen per Env-Variable aktivierbaren Integrationstest, der die Einrichtung gegen das echte fiskaly-TEST-Konto durchläuft, damit der Kontrakt zur echten API verifiziert bleibt.

## Implementation Decisions

### UI

- Eigene Admin-Seite „TSE-Einrichtung" (eigene Route). Sie bündelt den Wizard, den manuellen Experten-Bereich (API-Key/Secret, TSS-ID, Client-ID) und den Verbindungstest. Die manuelle Konfiguration zieht aus der Einstellungen-Sektion dorthin um.
- Die TSE-Sektion der Einstellungen wird zu einer Status-Anzeige (konfiguriert ja/nein, Umgebung, letzter Verbindungstest) mit Link auf die Einrichtungsseite.
- Wizard-Schritte: Zugangsdaten → Prüfung/Befund → Auswahl & Bestätigung → Durchführung → PUK/PIN-Anzeige (mit Verwahr-Bestätigung) → Abschluss mit Verbindungstest.
- Umgebung (TEST/LIVE) wird ab erfolgreicher Anmeldung in jedem Schritt sichtbar angezeigt. Vor TSS-Anlage in LIVE: Tipp-Bestätigung (wörtlich „LIVE" eintippen). In TEST genügt ein Bestätigungs-Klick.
- Fehlermeldungen folgen dem bestehenden Muster verständlicher deutscher Texte je Fehlercode.

### Backend

- Zwei neue Admin-Endpunkte (POST-only, analog bestehender Konvention): einer für die seiteneffektfreie Prüfung (nimmt API-Key/Secret entgegen, liefert Umgebung + TSS-/Client-Befund), einer für die Durchführung (Parameter: bestätigte Umgebung, TSS-Auswahl „neu anlegen" oder vorhandene TSS-ID, optional nachgereichte Admin-PIN).
- Die fiskaly-Setup-Operationen werden als Erweiterung der bestehenden fiskaly-Anbindung gebaut und teilen sich deren HTTP-/Auth-/Retry-Unterbau. Die Admin-Authentifizierung (PIN-basierter Token je TSS) ist ein interner Belang dieses Moduls.
- Der Setup-Orchestrator ist eine Zustandsmaschine über den fiskaly-Ist-Zustand: Er ermittelt den nächsten fehlenden Schritt und führt nur diesen Rest aus (TSS anlegen → PUK beziehen → PIN setzen → initialisieren → Client registrieren). Wiederaufnahme: Bei TSS im Zustand CREATED wird die PUK idempotent erneut bezogen; bei späteren Zuständen ohne bekannte PIN fordert der Orchestrator die PIN-Eingabe an.
- Die Admin-PIN wird zufällig erzeugt. PUK und PIN werden ausschließlich in der Antwort an die UI übergeben — keine Persistenz, kein Logging. **Keine Schema-Änderung** an der TSE-Konfiguration.
- Die Client-`serial_number` ist exakt jottis Kassen-Seriennummer (UUID). Das erfüllt zugleich DSFinV-K ≥ 2.3 (keine `/` und `_` in der serial_number).
- Die TSE-Konfiguration wird erst bei erfolgreichem Abschluss atomar und vollständig gespeichert (bestehende Invariante „alle vier Felder zusammen" bleibt unangetastet).
- Der Verbindungstest wird erweitert: zusätzlich Client-Zustand (`REGISTERED` erwartet) und Abgleich `serial_number` ↔ Kassen-Seriennummer; Abweichung ist ein Fehler. Das Ergebnis enthält Umgebung, TSS-Zustand, Client-Zustand und Abgleich-Resultat.

### Dokumentation

- Betreiber-Leitfaden (D-07) im Betriebs-Doku-Bereich: fiskaly-Konto und API-Key im Dashboard erstellen (mit Hinweis, dass TSS/Client-Anlage dort nicht möglich ist), Wizard-Ablauf, PUK/PIN-Verwahrung und Verlustfolgen, TEST→LIVE-Wechsel inkl. Kosten, manueller Fallback-Weg.

## Testing Decisions

- Gute Tests prüfen externes Verhalten — gesendete Requests, gelieferte Befunde und Ergebnisse — nicht Implementierungsdetails. Das Audit hat gezeigt, was passiert, wenn Tests das falsche Verhalten asserten (I-14): Die Kontrakt-Tests bilden deshalb die echte API-Spezifikation ab.
- **Kontrakt-Tests für die Setup-Operationen:** Fake-fiskaly-Server asserted Pfade, Bodies und Auth-Header (insbesondere Admin-Token vs. API-Token) je Lifecycle-Schritt. Vorbild: bestehende Kontrakt-Tests der fiskaly-Anbindung.
- **Unit-Tests für den Setup-Orchestrator** gegen einen Fake-Setup-Client: leeres Konto (Voll-Durchlauf), vorhandene TSS mit/ohne passenden Client, Wiederaufnahme aus jedem Zwischenzustand, PIN-Nachfrage, Umgebungs-Abweichung, Verweigerung der Neuanlage ohne Bestätigung.
- **Env-gated Integrationstest** gegen das echte fiskaly-TEST-Konto (Voll-Durchlauf bis zur signierfähigen TSS). Vorbild: bestehender env-gated Signier-Integrationstest. Hinweis: Jeder Lauf hinterlässt eine TSS im TEST-Konto (nicht löschbar) — sparsam und bewusst ausführen.
- **Keine Frontend-Komponententests** für den Wizard (Entscheidung; manuelle Verifikation über den Ende-zu-Ende-Durchlauf).

## Out of Scope

- **TSS-Stilllegung/Deaktivierung** (Zustand DISABLED) aus jotti heraus — bei Bedarf später, vorerst Leitfaden-Thema.
- **fiskaly-Konto-Registrierung und API-Key-Erstellung** — bleibt im fiskaly-Dashboard (dort möglich), jotti liefert nur die Anleitung.
- **Mehrere Kassen/TSS pro jotti-Instanz** — jotti hat genau eine Kassenidentität; Mehrkassen-Verwaltung ist kein Ziel.
- **Automatisierte API-Key-Rotation** — Schlüsseltausch bleibt manuell (neue Zugangsdaten eintragen).
- **Änderungen am Signierpfad, Nachsignier-Worker oder an Belegen** — laufen separat im Plan `plan-tse-fiskaly-fixes.md`.
- **DSFinV-K-Export und ELSTER-Meldung** — eigene Anforderungen (siehe `docs/anforderungen.md`).

## Further Notes

- **Ablösung der BYOT-Festlegung:** `prd-tse-integration.md` legte fest, dass der Betreiber TSS und Client „im fiskaly-Dashboard anlegt" und jotti keinen TSS-Lebenszyklus betreibt. Recherche (fiskaly-Support/Doku, Juni 2026) hat ergeben, dass das Dashboard das nicht kann: Clients sind nur per API anlegbar, die TSS-Initialisierung ebenso. Diese PRD übernimmt deshalb den **Setup**-Lebenszyklus in jotti; der laufende Betrieb (Signieren über API-Key) bleibt unverändert.
- **Adressierte Audit-Findings:** I-09 (serial_number-Konsistenz, durch Wizard erzwungen und durch Verbindungstest geprüft), I-15.4 (Client-Prüfung im Verbindungstest), D-07 (Betreiber-Leitfaden).
- **Risiken und Gegenmaßnahmen:** LIVE-Kosten/Doppel-Anlage → seiteneffektfreier Befund, Wiederverwendungs-Angebot, Tipp-Bestätigung; PUK/PIN-Verlust beim Betreiber → erzwungene Verwahr-Bestätigung im Wizard, Leitfaden-Kapitel, verständliche Sackgassen-Hinweise (fiskaly-Support bzw. Neuanlage).
- **Offen bis zur Umsetzung:** Exakte fiskaly-Preisstruktur für LIVE-TSS gehört in den Leitfaden und sollte bei Umsetzung aktuell recherchiert werden.
