# PRD: Einmalpasswort auf 6 Ziffern zurückbauen und Initial-Admin generieren

> Quelle: Bug-Meldung und Analyse des Maintainers vom 2026-07-06, verifiziert
> gegen die Codebase (Backend `domain/user`, Auth-Handler, Frontend
> `identity.ts`/`FormFields.tsx`, `database/migrations/01_initial.up.sql`,
> `reverse-proxy/statuspage.go`, Windows-Starter, `scripts/prod-init.sh`).
> Zwei zusammenhängende Punkte: (1) ein akuter Aussperr-Bug, (2) die saubere
> Erzeugung des Initial-Einmalpassworts. Überarbeitet nach Review vom
> 2026-07-06: Stale-Marker-Behandlung, Wiederverwendung bestehender
> Domain-Pfade, prod-init ohne Log-Parser, Neustart als einheitlicher
> Wiederherstellungsweg.

## Problem Statement

**Punkt 1 — Neuinstallationen sperren sich selbst aus.** Das Einmalpasswort
(OTP) ist kein Passwort, sondern ein Einmal-Code, um beim Erststart ein echtes
Passwort zu setzen. Ursprünglich waren es 6 Ziffern. Ein Coding-Agent hat den
Generator auf 8 alphanumerische Zeichen umgebaut und das Frontend darauf
festgezogen, aber den in der Initialmigration hartkodierten Argon2id-Hash des
Codes `123456` (6 Ziffern) nicht mitgezogen. Ergebnis bei jeder Neuinstallation:
Der einzige Weg zum ersten Login ist „Passwort festlegen“ mit dem Initial-OTP,
aber das Eingabefeld verlangt 8 Zeichen, während der hinterlegte Code `123456`
nur 6 Ziffern hat. Der Admin kann keinen gültigen Code eingeben — Deadlock.
Zusätzlich validiert das Backend das OTP-Format gar nicht (nur „nicht leer“),
sodass Frontend und Backend ohnehin auseinanderlaufen.

**Punkt 2 — Das Initial-OTP steht fest verdrahtet im Schema und in der Doku.**
Der Initial-Admin wird in `01_initial.up.sql` mit dem hartkodierten Hash von
`123456` eingefügt; die Installationsdoku nennt denselben Wert. Für jottis
Einsatz (lokal im Vereinsheim, keine öffentliche Erreichbarkeit) ist die
Angriffsfläche gering, aber ein für alle Installationen identischer, öffentlich
dokumentierter Erst-Code ist unsauber und wirkt unprofessionell. Schwerer wiegt
eine funktionale Sackgasse: Gibt jemand das Initial-OTP fünfmal falsch ein,
wird der Hash verworfen — und weil noch kein Admin mit Passwort existiert, der
ein neues OTP erzeugen könnte, ist die Neuinstallation ohne Datenbankeingriff
endgültig gesperrt.

Verifizierte Fakten, die den Rahmen absichern: Das OTP ist einmalig — bei
erfolgreichem Setzen wird sein Hash gelöscht. Es gibt einen Fehlversuchszähler:
Nach `MaxOnetimePasswordAttempts = 5` Fehlversuchen wird der Hash verworfen und
der Code endgültig ungültig. Bei 6 Ziffern (10^6 Kombinationen) und höchstens 5
Rateversuchen ist die Ratewahrscheinlichkeit vernachlässigbar. 6 Ziffern sind
für diesen Einsatzzweck sicher genug.

## Solution

**Punkt 1.** Der OTP-Generator erzeugt wieder genau 6 Ziffern. Frontend und
Backend validieren einheitlich „genau 6 Ziffern“ aus einer gemeinsamen Regel je
Seite. Damit passt das Eingabefeld wieder zum erzeugten Code, und der Deadlock
verschwindet — auch unabhängig von Punkt 2, weil der bestehende Seed `123456`
dann wieder eingebbar ist.

**Punkt 2.** Der hartkodierte Admin-Insert entfällt aus der Migration. Der
Initial-Admin wird stattdessen vom Backend beim Start erzeugt: Ist die
`users`-Tabelle leer, legt das Backend den aktiven Benutzer `admin` mit einem
zufälligen 6-stelligen OTP an und schreibt den Klartext-Code prominent ins
Log/Konsole. Weil der Klartext bewusst nie gespeichert wird, rotiert das Backend
den Code bei jedem Start neu, solange die Ersteinrichtung noch offen ist (einziger
`admin` ohne gesetztes Passwort). Sobald der Admin ein Passwort gesetzt hat,
endet die Rotation dauerhaft. So sieht der Anwender bei jedem Neustart einen
frischen, gültigen Code — das ist zugleich der Wiederherstellungsweg, falls der
Code verpasst oder nach fünf Fehlversuchen gesperrt wurde; die Sackgasse aus
Punkt 2 heilt damit von selbst.

Der Code wird dort sichtbar, wo der Anwender ohnehin hinschaut: Der
Windows-Starter liest ihn aus den Backend-Logs und zeigt ihn samt Benutzername
und Eingabeort in der Konsole; `make prod-init` (Linux) gibt den fertigen
Log-Befehl aus, mit dem der technische Betreiber den Code abliest. Die lokale
Status-Seite (`http://localhost:8484`) zeigt zusätzlich einen statischen
Hinweis für die Ersteinrichtung. In allen Fehlerfällen gilt derselbe einfache
Wiederherstellungsweg: jotti neu starten — dann wird ein neuer Code angezeigt.

## User Stories

1. Als Vereinshelfer, der jotti frisch auf einem Windows-Rechner startet, will
   ich den ersten Login abschließen können, damit die Kasse überhaupt nutzbar
   wird.
2. Als Vereinshelfer will ich, dass der im „Passwort festlegen“-Formular
   erwartete Code-Umfang exakt dem erzeugten Code entspricht, damit ich keinen
   ungültigen Code eintippe.
3. Als Vereinshelfer will ich beim Erststart genau 6 Ziffernfelder sehen und
   auf dem Tablet direkt die Zifferntastatur bekommen, damit klar ist, dass
   ein reiner Zahlencode erwartet wird.
4. Als Vereinshelfer will ich in der Starter-Konsole den erzeugten Erst-Code
   samt Benutzername (`admin`) und Eingabeort sehen, damit ich nichts erraten
   und nicht in Logs suchen muss.
5. Als Vereinshelfer, der den Code beim Start übersehen hat, will ich ihn durch
   einen Neustart erneut angezeigt bekommen, damit ich mich nicht selbst
   aussperre.
6. Als Vereinshelfer will ich nach abgeschlossener Einrichtung keinen
   veralteten Code mehr angezeigt bekommen, damit ich ihn nicht mit meinem
   Passwort verwechsle.
7. Als Vereinshelfer, der die Startkonsole schon geschlossen hat, will ich auf
   der Status-Seite den Hinweis finden, dass ein Neustart einen neuen Code
   anzeigt, damit ich einen klaren nächsten Schritt habe.
8. Als Betreiber auf einem Linux-VPS will ich am Ende von `make prod-init` den
   fertigen Befehl sehen, mit dem ich den Admin-Code aus den Backend-Logs
   ablese, damit ich ihn direkt weiterverwenden kann.
9. Als Betreiber auf einem Linux-VPS will ich den aktuell gültigen Code
   nachträglich aus den Backend-Logs ablesen können, damit ich ihn bei Bedarf
   erneut erhalte.
10. Als Admin will ich, dass mein selbst gesetztes Passwort die OTP-Rotation
    dauerhaft beendet, damit sich der Anmelde-Code nach der Einrichtung nie mehr
    ändert.
11. Als Admin, der später einen Service-Benutzer anlegt, will ich sicher sein,
    dass dessen noch nicht eingelöstes Einmalpasswort bei Backend-Neustarts
    unverändert bleibt, damit ich es nicht erneut ausgeben muss.
12. Als Servicekraft will ich mein vom Admin ausgegebenes Einmalpasswort als
    6-stelligen Zahlencode eingeben, damit das Onboarding einfach bleibt.
13. Als Admin will ich das dem Benutzer angezeigte Einmalpasswort als klaren
    6-Ziffern-Code sehen (Anlegen und Zurücksetzen), damit ich es leicht
    diktieren kann.
14. Als Angreifer im lokalen Netz will ich das Einmalpasswort nicht durch Raten
    treffen, weil es nach 5 Fehlversuchen gesperrt und nach einmaliger Nutzung
    gelöscht wird.
15. Als Betreiber will ich, dass in der Datenbankmigration kein für alle
    Installationen identischer Passwort-Hash steht, damit keine geteilte
    Standard-Kennung existiert.
16. Als Anwender will ich, dass die Installationsdoku keinen festen Erst-Code
    `123456` mehr verspricht, sondern den generierten Code korrekt beschreibt,
    damit die Anleitung zur Realität passt.
17. Als Maintainer will ich, dass Backend und Frontend das OTP-Format
    identisch prüfen, damit die beiden Seiten nicht erneut auseinanderlaufen.
18. Als Maintainer will ich die Erzeugungs-/Rotationslogik als eigenes,
    isoliert testbares Modul, damit die Rotations-Invariante (nie einen
    Live-Service-Code rotieren) durch Tests abgesichert ist.
19. Als Support-Kontakt will ich, dass ein verpasster oder nach Fehlversuchen
    gesperrter Erst-Code ohne Datenbankeingriff wiederherstellbar ist (Neustart
    genügt), damit Ferndiagnose einfach bleibt.
20. Als Betreiber will ich, dass die Erzeugung des Initial-Admins idempotent
    ist, damit ein Neustart bei bereits eingerichtetem System keinen zweiten
    Admin anlegt und kein bestehendes Konto verändert.
21. Als Entwickler will ich, dass ein leeres Feld oder ein Nicht-6-Ziffern-Code
    beim Passwortsetzen mit einer verständlichen Meldung abgelehnt wird, bevor
    ein Hashvergleich stattfindet.

## Implementation Decisions

### OTP-Format (Punkt 1)

- Der Generator im User-Domain-Modul erzeugt genau 6 Ziffern (`0`–`9`,
  gleichverteilt, ohne Modulo-Bias). Der Kommentar zur Entropie wird auf die
  neue Größe (10^6, abgesichert durch Einmaligkeit und 5-Versuche-Sperre)
  aktualisiert.
- Es gibt genau eine kanonische OTP-Format-Regel je Seite:
  - Backend: eine gemeinsame `OnetimePasswordSchema` (genau 6 Ziffern) im
    User-Domain-Modul, verwendet vom Auth-HTTP-Handler beim Passwortsetzen.
    Das ersetzt die bisherige „nur nicht leer“-Prüfung und schließt die
    Validierungslücke. Der Domain-Verify normalisiert höchstens noch durch
    Trimmen; die frühere tolerante Kleinschreibung (für das Alphabet-Alphabet)
    entfällt, weil reine Ziffern sie nicht brauchen.
  - Frontend: `OnetimePasswordSchema` in `identity.ts` prüft `^\d{6}$` mit
    deutscher Fehlermeldung „genau 6 Ziffern“. Der bereits vorhandene, aber
    veraltet-falsche Kommentar (nennt schon 6 Ziffern) wird damit wieder
    wahr.
- Das OTP-Eingabefeld (`FormFields.tsx`, `OTPField`) hat 6 Slots, `maxLength=6`
  und ein reines Ziffern-Muster statt Ziffern-und-Buchstaben. Zusätzlich
  `inputMode="numeric"` und `autocomplete="one-time-code"`, damit Tablets und
  Smartphones direkt die Zifferntastatur anbieten.
- Die Anzeige-Dialoge für erzeugte Codes (Benutzer anlegen, Passwort
  zurücksetzen) zeigen den Code weiterhin generisch groß; keine Längenannahme
  im Markup, daher keine Änderung nötig außer visueller Prüfung.
- Effekt der Vereinheitlichung: Weil beide Seiten denselben Generator/dasselbe
  Format nutzen, gelten 6 Ziffern gleichermaßen für den Initial-Admin, für vom
  Admin angelegte Benutzer und für Passwort-Resets.

### Initial-Admin-Bootstrap (Punkt 2)

- Der Admin-`INSERT` samt hartkodiertem Hash und erklärendem Kommentar entfällt
  aus `01_initial.up.sql`. Die Migration erzeugt nur noch das Schema; sie bleibt
  reines, deterministisches SQL (der Migrate-Container kann kein Argon2id
  rechnen).
- Neues, isoliert testbares Bootstrap-Modul mit einfacher Schnittstelle über ein
  User-Repository (Zählen/Lesen, Anlegen, Aktualisieren). Eine Funktion
  entscheidet anhand des DB-Zustands genau eine von drei Aktionen und liefert
  den Klartext-Code plus die gewählte Aktion zurück:
  - **create**: `users` ist leer. Lege `admin` an: Rolle `admin`, Status
    `active` (nicht `inactive` — der erste Admin kann sich sonst nach dem
    Passwortsetzen nicht anmelden, da Login `active` verlangt), kein Passwort,
    frisches 6-Ziffern-OTP.
  - **rotate**: Genau ein Benutzer, dieser ist `admin` und hat noch kein
    Passwort gesetzt (Onboarding offen). Erzeuge ein neues OTP, setze den
    Fehlversuchszähler zurück, persistiere. Dies ist der Rotations-/
    Wiederherstellungsfall.
  - **skip**: Jeder andere Zustand (Admin hat ein Passwort, oder es existieren
    weitere Benutzer). Keine Änderung. Damit wird nach der Ersteinrichtung nie
    wieder rotiert und insbesondere nie das noch offene Einmalpasswort eines
    Service-Benutzers angetastet.
- Der Domain-Teil kommt ohne neue Pfade aus: `create` komponiert das
  bestehende `NewUser` (erzeugt `inactive` plus frisches OTP) mit dem
  bestehenden `Activate()`; `rotate` ist exakt das bestehende `ResetPassword()`
  (neues OTP, Zähler zurück, Passwort bleibt leer, Status bleibt unberührt).
  Das Bootstrap-Modul enthält nur die Entscheidungslogik.
- Das Bootstrap läuft beim Backend-Start vor der Auslieferung von Anfragen,
  nach dem erfolgreichen DB-Ping und nach den Migrationen (der Backend-Service
  hängt bereits per `depends_on` am Migrate-Service). Es ist idempotent über die
  drei Aktionen; die eindeutige Username-Bedingung schützt zusätzlich gegen
  Doppelanlage.
- Repository-Erweiterung: eine Zähl-/Existenzabfrage für Benutzer sowie die
  bereits vorhandenen Anlege-/Aktualisieren-Pfade genügen.

### Sichtbarkeit des Codes (Konsole/Logs + Status-Hinweis)

- Das Backend loggt bei `create` und `rotate` eine stabile, maschinen-greifbare
  Markerzeile mit dem Klartext-Code (fester Präfix zum Grep), zusätzlich zu
  einer menschenlesbaren Bannerzeile. Der Marker ist die einzige Stelle, an der
  der Klartext das Backend verlässt; er geht nur in den Log-Strom, nicht über
  das Netz.
- Markerzeilen veralten: Docker-Logs überleben Container-Neustarts, und nach
  abgeschlossener Einrichtung bleibt die letzte Markerzeile aus der Setup-Phase
  im Log-Strom stehen. Log-Konsumenten dürfen deshalb nur Zeilen seit dem
  aktuellen Container-Start berücksichtigen (`docker inspect`
  `.State.StartedAt` plus `docker logs --since`). Bei offener Einrichtung loggt
  jeder Start einen frischen Marker; fehlt er im aktuellen Boot, ist die
  Einrichtung abgeschlossen und es wird kein Code angezeigt.
- Linux: `scripts/prod-init.sh` parst keine Logs. Das Abschluss-Summary nennt
  stattdessen den fertigen Befehl, mit dem der Betreiber den Code abliest
  (`docker compose … logs backend | grep <Präfix>`). VPS-Betreiber sind
  technisch; das spart einen spröden Shell-Parser samt der Frage, woher das
  Skript den Einrichtungszustand kennen soll.
- Windows: Der Starter liest nach erfolgreichem Health-Check die jüngste
  Markerzeile seit dem aktuellen Container-Start und zeigt sie als vollständige
  Handlungsanweisung: Benutzername `admin`, der 6-stellige Code und der
  Eingabeort („Passwort festlegen“ in der App). Schlägt das Lesen/Parsen fehl,
  ist das nicht fatal; die Meldung lautet dann „jotti neu starten — dann wird
  ein neuer Code angezeigt“. Ein Verweis auf Logs hilft einem Vereinshelfer
  nicht; der Neustart ist der Wiederherstellungsweg.
- Status-Seite (`reverse-proxy/statuspage.go`): eine zusätzliche, statische,
  stets unbedenkliche Hinweiskarte, konditional formuliert: „Falls Sie jotti
  gerade zum ersten Mal einrichten: Der Anmelde-Code steht in der Startkonsole.
  Konsole schon geschlossen? jotti neu starten — dann wird ein neuer Code
  angezeigt.“ Kein Datenbankzugriff, kein Backend-Aufruf, kein Klartext-Code
  auf der Seite — die Status-Seite ist ein eigenes Binary ohne DB-Anbindung.
- Wiederherstellung: Da bei offener Einrichtung jeder Start ein frisches OTP
  loggt und den Fehlversuchszähler zurücksetzt, genügt ein Backend-Neustart —
  auch nachdem fünf Fehlversuche das OTP gesperrt haben. Kein SQL-Runbook,
  kein eigenes Reset-Kommando nötig.

### Dokumentation

- `docs/leitfaden/installation.md`: Der feste Wert `123456` entfällt. Der
  Abschnitt „Erster Login“ beschreibt, dass der Erst-Code beim Start erzeugt und
  in der Startkonsole (Windows) bzw. in der `make prod-init`-Ausgabe/den Logs
  (VPS) angezeigt wird, und dass ein Neustart einen neuen Code zeigt, solange
  noch kein Passwort gesetzt ist.
- `docs/language.md` (Glossar „Einmalpasswort“) und `docs/handbuch.md`
  (Onboarding-Ablauf, Benutzer-Invarianten): „8 Zeichen (Kleinbuchstaben und
  Ziffern)“ wird zu „6 Ziffern“. Der Onboarding-Abschnitt spiegelt zusätzlich
  den generierten, rotierenden Initial-Admin-Code statt des festen `123456`.

## Testing Decisions

Gute Tests prüfen beobachtbares Verhalten an der Modulschnittstelle, nicht
Implementierungsdetails. Sie beschreiben, was hineingeht und was
herauskommt/persistiert wird, und bleiben stabil, wenn sich das Innenleben
ändert. Getestet werden (alle vier vom Maintainer bestätigt):

1. **Initial-Admin-Bootstrap-Entscheidung** (höchster Wert). Tabellengetrieben
   gegen ein Fake-Repository, ohne DB:
   - leeres Repo → `create`: aktiver `admin`, kein Passwort, OTP-Hash gesetzt,
     Rückgabe-OTP sind 6 Ziffern.
   - einziger `admin` ohne Passwort → `rotate`: neuer OTP-Hash, Zähler auf 0,
     Status/Passwort unverändert leer.
   - einziger `admin` ohne Passwort, OTP nach Fehlversuchen gesperrt (Hash
     leer) → `rotate`: frischer OTP-Hash, Zähler auf 0 (Selbstheilung der
     Sackgasse).
   - `admin` mit Passwort → `skip`: keine Änderung.
   - mehrere Benutzer (inkl. Service-Benutzer mit offenem OTP) → `skip`: der
     Service-OTP bleibt unangetastet (sichert die Rotations-Invariante).
   Vorlage: `backend/domain/user/user_test.go`, `password_test.go` sowie die
   Fake-/Repo-Muster in `backend/repository/user_repo/repo_test.go`.
2. **OTP-Generator**: Rückgabe hat genau 6 Zeichen, ausschließlich `0`–`9`.
   Vorlage: bestehende Fälle in `password_test.go`.
3. **Backend-OTP-Formatvalidierung**: Der Passwort-Setzen-Handler lehnt ein
   Nicht-6-Ziffern-OTP (leer, zu kurz, zu lang, nicht-numerisch) mit dem
   Client-Fehler ab, bevor ein Hashvergleich passiert. Vorlage:
   `backend/api/auth/http/command_handler_test.go`.
4. **Frontend-OTP-Zod-Schema**: `OnetimePasswordSchema` akzeptiert `123456` und
   lehnt `12345`, `1234567`, `abcdef` ab. Vorlage: bestehende Vitest-Tests wie
   `frontend/src/lib/errorMessages.test.ts`.

Bestehende Domain-Tests zum Fehlversuchszähler und zur Einmaligkeit bleiben
gültig; ihre Fixtures werden von 8-Zeichen- auf 6-Ziffern-OTPs umgestellt.

Nicht als Unit-Test abgesichert: das Log-Parsen im Windows-Starter (Log-Parsing
ist spröde; `prod-init.sh` parst nicht mehr, sondern gibt nur den Log-Befehl
aus). Dieser Pfad wird manuell bzw. über den bestehenden Release-Smoke-Test
abgedeckt — inklusive des Falls „Einrichtung abgeschlossen, kein Marker im
aktuellen Boot, kein Code in der Ausgabe“. Ein fehlgeschlagenes Parsen ist
bewusst nicht fatal und endet in der Neustart-Meldung.

## Out of Scope

- Anzeige des Klartext-OTP direkt auf der Status-Seite (würde eine geteilte
  Datei im Volume mit Lebenszyklus oder einen Backend-Aufruf erfordern; bewusst
  verworfen zugunsten des statischen Hinweises).
- Ein separates `reset-admin-otp`-Kommando oder ein SQL-Runbook zur
  Wiederherstellung — durch die Neustart-Rotation überflüssig.
- Ein „Ersteinrichtung offen“-Flag im Health-Endpoint als Signal für die
  Log-Konsumenten — der `--since`-Filter auf den aktuellen Container-Start
  genügt, und das Flag würde im LAN verraten, dass ein unbeanspruchter Admin
  existiert.
- Änderungen an der Passwort-Policy selbst (Mindestlänge 6, max. 72 bleiben).
- Änderungen an der Anmelde-Drosselung (`loginThrottle`) und am
  Fehlversuchslimit (5 bleibt).
- OTP-Zustellung per E-Mail/SMS oder QR — der Code wird wie bisher mündlich/
  über die Konsole weitergegeben.
- Mehrinstanz-Betrieb des Backends (jotti läuft als eine Instanz); ein
  Advisory-Lock für das Bootstrap ist deshalb nicht nötig, die
  Username-Eindeutigkeit genügt als Schutz.

## Further Notes

- **Reihenfolge/Risiko**: Punkt 1 (6-Ziffern-Rückbau) behebt den Deadlock
  bereits allein, weil der bestehende Seed `123456` dann wieder eingebbar ist.
  Punkt 2 (Migration entkoppeln, Bootstrap) kann darauf aufbauen. Die
  Umsetzung kann in zwei Schritten erfolgen, gehört aber fachlich zusammen und
  bleibt ein PRD.
- **Migrationskompatibilität**: Bestehende Installationen haben den Admin schon
  (aus der alten Migration) und sind nicht leer, treffen also `skip` — kein
  Zweitanlegen, keine Störung. Nur echte Neuinstallationen durchlaufen `create`.
- **Bewusst akzeptierte Kante**: Ein Ein-Admin-System, das nach der Einrichtung
  das eigene Admin-Passwort zurücksetzt (wieder genau ein `admin` ohne Passwort)
  und dann neu startet, trifft erneut `rotate`. Das ist selten und unschädlich —
  der neue Code steht wie gewohnt in der Konsole.
- **Sicherheitsrahmen**: Der Klartext-Code erscheint nur lokal (Konsole/Logs am
  Kassenrechner bzw. VPS-Shell); die Status-Seite bindet nur an `127.0.0.1`. Für
  jottis Einsatz ohne öffentliche Erreichbarkeit ist das angemessen.
