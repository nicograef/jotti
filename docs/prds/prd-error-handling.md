# PRD: Konsistentes Error-Handling & verständliche Fehlermeldungen

## Problem Statement

Wenn in jotti eine Aktion fehlschlägt — Kassensitzung öffnen, Tisch anlegen, Bestellung aufnehmen — sieht der Nutzer meistens nur einen generischen Toast: „X fehlgeschlagen. Bitte erneut versuchen." Das gibt weder Auskunft darüber, was falsch war, noch was der Nutzer als Nächstes tun soll. Beim Usability-Test (10.06.2026) führte z. B. das Fehlen von Betreiberinformationen zu einem nichtssagenden Server-Error, obwohl der Backend-Fehlercode `betreiber_nicht_konfiguriert` präzise beschreibt, was zu tun ist. Das zwingt Nutzer entweder zu raten oder den Admin zu kontaktieren — unnötiger Frust bei einer ehrenamtlich betriebenen Vereinsveranstaltung.

Ein zweites Problem ist Inkonsistenz: 14 Frontend-Komponenten verwenden manuelles `try/catch + toast.error()` statt des vorhandenen `useActionSubmit`-Hooks. Neue Error-Codes müssen deshalb an vielen verstreuten Stellen ergänzt werden, und einige Fehler werden ohne jede sichtbare Meldung geschluckt.

Schließlich werden Validierungsfehler vom Backend (`validation_error` mit Feldpfaden im `details`-Array) nur als generischer Toast angezeigt, obwohl die Information vorhanden wäre, das betroffene Formularfeld direkt zu beschriften.

## Solution

1. **Vollständige Fehlermeldungs-Dictionary**: Jeder der ~35 Backend-Error-Codes erhält eine deutsche, kontextspezifische User-Message. Auth-Fehler und Server-Fehler werden separat behandelt.

2. **Unterscheidung recoverable vs. nicht-recoverable**: Domänen- und Validierungsfehler (400/409) zeigen eine handlungsorientierte Nachricht. Server-Fehler (500/unbekannt) zeigen eine eigene Meldung, die den Nutzer auf den Admin verweist.

3. **Inline-Validierungsfehler für Formulare**: Ein neuer `useFormActionSubmit`-Hook mappt `validation_error`-Details aus dem Backend direkt auf die betroffenen Formularfelder via `form.setError()`. Andere Fehler erscheinen weiterhin als Toast.

4. **Konsistentes Frontend-Pattern**: Alle 14 Komponenten, die noch manuelles `try/catch + toast.error()` verwenden, werden auf `useActionSubmit` bzw. `useFormActionSubmit` migriert.

## User Stories

1. Als Admin möchte ich beim Versuch, eine Kassensitzung ohne konfigurierte Betreiberinformationen zu öffnen, einen Hinweis sehen, der mich direkt zu den Einstellungen führt, damit ich das Problem selbst lösen kann.
2. Als Admin möchte ich beim Anlegen eines Tisches mit einem bereits vorhandenen Namen den Fehler direkt am Namensfeld sehen (und nicht nur als Toast), damit ich das Feld sofort korrigieren kann.
3. Als Admin möchte ich beim Anlegen eines Produkts, einer Variante oder eines Benutzers mit ungültigen Feldern die Fehler direkt an den betroffenen Feldern angezeigt bekommen.
4. Als Admin möchte ich beim Versuch, eine Kassensitzung zu öffnen, obwohl bereits eine offen ist, einen klaren Hinweis sehen, damit ich nicht suchen muss, warum es nicht klappt.
5. Als Admin möchte ich beim Tagesabschluss, wenn noch offene Tische existieren, eine verständliche Meldung sehen, die mir erklärt, was ich zuerst tun muss.
6. Als Admin möchte ich beim Anlegen eines Benutzers mit bereits vergebenem Benutzernamen einen aussagekräftigen Fehler direkt am Nutzername-Feld sehen.
7. Als Servicekraft möchte ich beim Kassieren, wenn eine Position nicht mehr bezahlbar ist, einen klaren Hinweis sehen statt einer generischen Fehlermeldung.
8. Als Servicekraft möchte ich beim Stornieren einer Position, die nicht stornierbar ist, einen erklärenden Hinweis sehen.
9. Als Nutzer möchte ich bei einem unerwarteten Server-Fehler (500) eine spezifische Meldung sehen, die mir sagt, dass ich den Admin kontaktieren soll — und nicht dieselbe Meldung wie bei einem Validierungsfehler.
10. Als Nutzer möchte ich bei einem Verbindungsfehler zum TSE-Dienst eine Meldung sehen, die das Problem klar benennt.
11. Als Nutzer möchte ich bei fehlgeschlagener Anmeldung eine Meldung sehen, die mir sagt, was falsch war (ungültige Zugangsdaten, inaktiver Account), ohne sicherheitsrelevante Details preiszugeben.
12. Als Nutzer möchte ich bei einem Anfragetimeout oder zu großer Anfrage eine verständliche Rückmeldung erhalten.
13. Als Entwickler möchte ich, dass jeder neue Backend-Error-Code automatisch eine Fallback-Message erhält, damit kein Fehler vollständig stumm bleibt, auch wenn die Message noch nicht gepflegt wurde.
14. Als Entwickler möchte ich eine einzige Stelle im Frontend haben, an der ich die User-facing Message für einen Error-Code pflegen kann, damit Änderungen nicht an 14 Stellen vorgenommen werden müssen.
15. Als Entwickler möchte ich bei der Migration einer Komponente auf `useFormActionSubmit` keine manuelle Fehler-Mapping-Logik mehr schreiben müssen.

## Implementation Decisions

### Backend: keine strukturellen Änderungen notwendig

Das Backend hat bereits eine saubere Error-Handling-Architektur: `helper.MapError`, `SendClientError`, `SendConflict`, `SendServerError`. Alle relevanten Domain-Fehler sind als benannte `error`-Variablen definiert und in den Handlern explizit auf String-Codes gemappt. Keine strukturellen Änderungen nötig.

Eine Ausnahme: Audit aller `MapError`-Aufrufe auf Vollständigkeit. Wenn ein Handler `map[error]string{}` (leere Map) verwendet, fällt jeder Fehler auf `internal_server_error` zurück — das ist in mindestens einem Tisch-Handler der Fall und ist zu prüfen.

### Frontend Modul 1: `lib/errorMessages.ts` — Vollständige Message-Dictionary

- Die `commonErrorMessages`-Map wird um alle bislang fehlenden ~29 Error-Codes erweitert.
- Gruppierung nach Fehlerart (Kommentare im Quellcode):
  - **Domänen-/Geschäftslogik-Fehler** (400/409): konkreter Hinweis was zu tun ist.
  - **Validierungsfehler** (`validation_error`): generischer Text „Bitte Eingaben prüfen." — Felddetails werden inline angezeigt (siehe Modul 3).
  - **Server-Fehler** (`internal_server_error`, `unknown`): eigene Nachricht „Unerwarteter Fehler. Bitte Seite neu laden oder den Administrator kontaktieren."
  - **Auth-Fehler** (`invalid_credentials`, `user_inactive` etc.): bleiben in der Login-Komponente isoliert behandelt (byCode).
- `getActionErrorMessage` bekommt eine dritte Unterscheidungsstufe: wenn `error.status >= 500` oder `error.code === 'internal_server_error'`, wird die Server-Error-Message zurückgegeben, unabhängig vom ActionLabel.
- Wichtig: die Auth-Fehler (`unauthorized`, `missing_authorization`, `invalid_jwt`) werden von `Backend.ts` schon via Redirect behandelt; sie benötigen keine User-facing Message, aber einen Eintrag als leerer String oder eine Notiz wäre sinnvoll.

Vollständige Code-Liste der zu deckenden Codes (dedupliziert, alle Backend-Error-Codes):
`already_has_password`, `betreiber_nicht_konfiguriert`, `cannot_delete_self`, `conflict`, `invalid_json`, `invalid_kassensitzung_nr`, `invalid_produkt_data`, `invalid_tisch_data`, `invalid_variante_data`, `insufficient_permissions`, `internal_server_error`, `kasse_bereits_geoeffnet`, `kasse_nicht_geoeffnet`, `kassensturz_erforderlich`, `no_password_set`, `password_too_weak`, `position_nicht_stornierbar`, `produkt_already_exists`, `produkt_not_found`, `request_too_large`, `tisch_already_exists`, `tisch_not_found`, `tische_saldo_offen`, `tse_nicht_konfiguriert`, `tse_verbindung_fehlgeschlagen`, `user_inactive`, `user_not_found`, `username_already_exists`, `validation_error`, `variante_not_found`, `verkauf_not_found`, `unknown`

### Frontend Modul 2: `hooks/use-form-action-submit.ts` — Inline-Validierungsfehler

Neuer Hook, der `useActionSubmit` für Formulare mit react-hook-form erweitert:

**Interface:**

```
useFormActionSubmit({
  form: UseFormReturn,
  actionLabel: string,
  byCode?: Record<string, string>,
  onSuccess?: () => void,
}) → { loading: boolean, run: (fn: () => Promise<void>) => Promise<void> }
```

**Verhalten bei `validation_error`:**

- `error.details` ist ein Array von `{ path: string, code: string, message: string }`.
- Der Hook iteriert über das Array und ruft für jeden Eintrag `form.setError(path, { message })` auf.
- Kein Toast für `validation_error` (die Inline-Fehler sind sichtbarer).
- Wenn `details` leer oder fehlt: Fallback-Toast mit der Message für `validation_error`.

**Verhalten bei anderen Fehlern:**

- Wie `useActionSubmit`: Toast mit `getActionErrorMessage(...)`.

**Verhalten bei `produkt_already_exists`, `tisch_already_exists`, `username_already_exists`:**

- Diese sind keine `validation_error`, sondern Domain-Fehler. Sie können via `byCode` auf `form.setError('name', ...)` gemappt werden — die Komponente entscheidet, ob sie `byCode` nutzt oder den Toast akzeptiert. Der Hook muss das nicht selbst wissen.

### Frontend Modul 3: `hooks/use-action-submit.ts` — Server-Error-Unterscheidung

Der bestehende Hook delegiert an `getActionErrorMessage`. Durch die Änderung in Modul 1 (`getActionErrorMessage` unterscheidet 500 vs. 400/409) ist keine Änderung am Hook selbst notwendig.

### Frontend Modul 4: Migration — 14 Komponenten

Alle 14 Komponenten, die manuelles `try/catch + toast.error()` verwenden, werden migriert:

- Formulare mit `UseFormReturn`: Migration zu `useFormActionSubmit`.
- Aktionen ohne Formular (Löschen, Aktivieren/Deaktivieren etc.): Migration zu `useActionSubmit`.
- Sonderfälle (z. B. Tisch- oder Produktanlage, wo ein `already_exists`-Fehler als Inline-Fehler am Namensfeld erscheinen soll): `useFormActionSubmit` + `byCode` für den spezifischen Error-Code.

Betroffene Komponenten:
`KassensitzungPage`, `EditProductDialog`, `EditVariantDialog`, `NewProductDialog`, `NewVariantDialog`, `ProductItem`, `DruckstationConfigPage`, `EinstellungenPage`, `EditTischDialog`, `NewTischDialog`, `Tische`, `EditUserDialog`, `PasswordForm`, `TischAuswahlDrawer`

### Schema-Änderungen

Keine.

### API-Änderungen

Keine.

## Testing Decisions

**Was macht einen guten Test hier aus:** Tests prüfen das externe Verhalten — was zeigt die Funktion zurück, welche Fehler werden aufgerufen —, nicht wie intern iteriert oder konditionell verzweigt wird. Kein Mocking von Implementierungsdetails.

### Modul 1: `lib/errorMessages.ts`

- Unit-Tests für `getActionErrorMessage`:
  - Für jeden neuen Error-Code in `commonErrorMessages`: `BackendError(400, code)` → erwartete Message.
  - `BackendError(500, 'internal_server_error')` → Server-Error-Message (nicht ActionLabel-Fallback).
  - `byCode`-Override: `byCode`-Message hat Vorrang vor `commonErrorMessages`.
  - Nicht-`BackendError`-Fehler: generischer Fallback.
- Prior art: vorhandene Tests in `frontend/src/lib/` (falls vorhanden), sonst analoge Muster aus dem Projekt.

### Modul 2: `hooks/use-form-action-submit.ts`

- Unit-Tests mit `renderHook` (Vitest + Testing Library):
  - `validation_error` mit `details` → `form.setError` wird für jeden Pfad aufgerufen; kein Toast.
  - `validation_error` ohne `details` → Toast wird aufgerufen.
  - Anderer Fehlercode → Toast mit korrekter Message, kein `form.setError`.
  - Erfolg → `onSuccess` wird aufgerufen; kein Toast, kein `setError`.
- Prior art: vorhandene Hook-Tests im Projekt (z. B. in `frontend/src/hooks/`).

### Kein separates Testing für Modul 4

Die Migration selbst wird nicht gesondert getestet — die bestehende manuelle Fehlerbehandlung wurde ohnehin nicht getestet. Korrektheit wird durch Modul 1 + 2 abgedeckt.

## Out of Scope

- Inline-Fehlerbanner auf Seiten-Ebene (oberhalb von Formularen) — Toasts sind für jotti ausreichend.
- React Error Boundaries / Full-Page-Fehlerseiten.
- Retry-Mechanismen im Frontend (Nutzer klickt im Toast auf „Erneut versuchen").
- Backend-Fehlerlogging verbessern (zerolog ist bereits vorhanden und ausreichend).
- TSE-spezifisches Error-Reporting (eigener Bereich, komplexere UX).
- Relay-Fehler (separater Prozess, eigenes Error-Handling).
- Internationalisierung / Mehrsprachigkeit.
- Fehlerseiten für HTTP 404 / 401 auf Routing-Ebene.

## Further Notes

- `tisch_not_active` ist in `commonErrorMessages` bereits vorhanden, taucht aber im Backend-Error-Code-Audit nicht als explizit gesendeter Code auf. Herkunft vor der Migration prüfen — ist es ein `byCode`-Override aus einer Komponente oder ein bisher ungenutzter Eintrag?
- `position_nicht_bezahlbar` taucht in `ZahlungDrawer.tsx` via `byCode` auf, fehlt aber in `commonErrorMessages`. Als Fallback sollte er ergänzt werden, damit er auch ohne expliziten `byCode`-Override eine sinnvolle Message hat.
- Auth-Fehler (`unauthorized`, `missing_authorization`, `invalid_jwt`) werden aktuell in `Backend.ts` via Redirect behandelt, bevor ein Toast angezeigt wird. Dieses Verhalten bleibt unverändert; die Codes brauchen keine User-facing Message.
- Die Fehler `invalid_credentials` und `no_password_set` / `already_has_password` sind spezifisch für Login/Passwort-Flows. Sie sollten als `byCode`-Override in den jeweiligen Komponenten bleiben (kontextabhängig, nicht generisch).
