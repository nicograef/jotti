# Plan: Konsistentes Error-Handling & verständliche Fehlermeldungen

> Source PRD: docs/prds/prd-error-handling.md

## Goal

Jeder Backend-Error-Code zeigt dem Nutzer eine deutsche, handlungsorientierte Message. Validierungsfehler erscheinen inline an den betroffenen Formularfeldern. Alle 14+ Komponenten mit manuellem `try/catch + toast.error()` werden auf den zentralen `useActionSubmit`-/`useFormActionSubmit`-Hook migriert.

## Architectural decisions

- **Kein neues API-Protokoll**: Backend-Error-Format bleibt `{ code: string, details?: map[string][]string }` (400/409/500). Keine Backend-Protokolländerungen außer dem Audit leerer `MapError`-Maps.
- **details-Format**: `validation_error.details` ist `map[string][]string` (zog `FlattenAndCollect`). Der Hook nimmt den ersten Eintrag je Feld als Message für `form.setError`.
- **Kein neuer State-Layer**: Migration nutzt ausschließlich die zwei Hooks; kein globaler Error-Store.
- **Auth-Fehler out of scope**: `unauthorized`/`missing_authorization`/`invalid_jwt` werden von `Backend.ts` via Redirect behandelt — keine User-facing Messages nötig.

## Inventory

- `frontend/src/lib/errorMessages.ts:1-45` — zentrale Message-Map + `getActionErrorMessage`; 7 Codes vorhanden, ~29 fehlen
- `frontend/src/lib/Backend.ts:4-8` — `ErrorResponseSchema` mit `details: z.string().optional()` — muss auf `z.unknown().optional()` erweitert werden
- `frontend/src/lib/Backend.ts:10-22` — `BackendError` — `details` aktuell `string`, muss `unknown` werden
- `frontend/src/lib/Backend.ts:112-130` — Fehlerbehandlung in `Backend.post`: wirft `BackendError` mit `details`
- `frontend/src/hooks/use-action-submit.ts:1-40` — bestehender Hook; delegiert an `getActionErrorMessage`
- `backend/api/helper/http.go:50-100` — `SendClientError`, `SendConflict`, `SendServerError`, `MapError`, `ReadAndValidateBody`
- `backend/api/table/http/command_handler.go:147-151` — `FavoritEntfernen` mit leerem `MapError`-Map → jeder Fehler wird als 500 zurückgegeben
- Alle 29 Komponenten mit `try {` in `frontend/src/admin/` und `frontend/src/service/` (via Grep verifiziert)

**Komponenten für Migration (Phase 3):**

| Komponente                                  | Typ                                                                 | Formular? |
| ------------------------------------------- | ------------------------------------------------------------------- | --------- |
| `admin/kasse/KassensitzungPage.tsx`         | Mehrere Subformulare (EroeffnenSection, KassenbewegungSection etc.) | ja        |
| `admin/settings/EinstellungenPage.tsx`      | Mehrere Subformulare (Betreiber, TSE etc.)                          | ja        |
| `admin/settings/DruckstationConfigPage.tsx` | Formular                                                            | ja        |
| `admin/products/NewProductDialog.tsx`       | Formular                                                            | ja        |
| `admin/products/EditProductDialog.tsx`      | Formular                                                            | ja        |
| `admin/products/NewVariantDialog.tsx`       | Formular                                                            | ja        |
| `admin/products/EditVariantDialog.tsx`      | Formular                                                            | ja        |
| `admin/products/ProductItem.tsx`            | Aktionen (aktivieren, deaktivieren, löschen, archivieren)           | nein      |
| `admin/tables/NewTischDialog.tsx`           | Formular                                                            | ja        |
| `admin/tables/EditTischDialog.tsx`          | Formular                                                            | ja        |
| `admin/tables/Tische.tsx`                   | Aktionen (aktivieren, deaktivieren, löschen)                        | nein      |
| `admin/users/NewUserDialog.tsx`             | Formular                                                            | ja        |
| `admin/users/EditUserDialog.tsx`            | Formular + Aktion (Passwort zurücksetzen)                           | ja + nein |
| `admin/users/Users.tsx`                     | Aktionen (aktivieren, deaktivieren, löschen)                        | nein      |
| `components/common/PasswordForm.tsx`        | Formular                                                            | ja        |
| `service/components/TischAuswahlDrawer.tsx` | Aktion                                                              | nein      |

## Resolved decisions

- `Backend.ts` wird angepasst: `details: z.unknown().optional()` im Schema, `BackendError.details: unknown` im Konstruktor — Voraussetzung für inline Feldfehler.
- `validation_error.details` bleibt `map[string][]string`; `useFormActionSubmit` nimmt `details[field][0]` als Message für `form.setError`.
- Backend-Audit ist im Scope: leere `MapError`-Maps in `backend/api/` werden geprüft und ergänzt (mindestens `FavoritEntfernen`).
- 3 Phasen: Foundation → Hook → Migration.

## Open questions / Risks

- `tisch_not_active` ist bereits in `commonErrorMessages`, taucht aber nicht als explizit gesendeter Backend-Code auf. Vor Phase 3 prüfen, ob der Eintrag genutzt wird oder entfernt werden kann.
- `EinstellungenPage.tsx` hat 5 separate `try {}`-Blöcke (verschiedene Subformulare). Migration auf Hook erfordert, dass jeder Block einzeln betrachtet wird — ggf. die komplizierteste Datei in Phase 3.

---

## Phase 1: Foundation — Error Dictionary + Backend.ts fix + Backend-Audit

**User stories**: 1, 4, 5, 7, 8, 9, 10, 11, 12, 13, 14

### Context

- `frontend/src/lib/errorMessages.ts:1-45` — aktuelle Map mit 7 Codes; `getActionErrorMessage` unterscheidet noch nicht 5xx von 4xx
- `frontend/src/lib/Backend.ts:4-22` — `ErrorResponseSchema.details: z.string().optional()` und `BackendError(status, code, details?: string)` — beide müssen `unknown` akzeptieren, damit `map[string][]string` vom Backend nicht als Parse-Fehler verworfen wird
- `frontend/src/lib/Backend.test.ts` — bestehende Tests als Vorlage für neue Unit-Tests
- `backend/api/helper/http.go:85-100` — `MapError`-Implementierung
- `backend/api/table/http/command_handler.go:147-151` — leeres `MapError`-Map in `FavoritEntfernen`

### What to build

**Backend.ts**: `ErrorResponseSchema.details` auf `z.unknown().optional()` ändern. `BackendError` nimmt `details?: unknown` statt `string`. `BackendError.details` ist `unknown`.

**errorMessages.ts**: `commonErrorMessages` um alle fehlenden Codes erweitern (vollständige Liste aus PRD). `getActionErrorMessage` bekommt eine dritte Unterscheidungsstufe: bei `error.status >= 500` oder `error.code === 'internal_server_error'` wird eine feste Server-Error-Message zurückgegeben, unabhängig vom `actionLabel`. `position_nicht_bezahlbar` wird als Fallback-Eintrag ergänzt.

**Backend-Audit**: Alle `MapError`-Aufrufe in `backend/api/` auf leere oder unvollständige Maps prüfen. Mindestens `FavoritEntfernen` (`table/http/command_handler.go:149`) enthält `map[error]string{}` — alle zutreffenden Domain-Fehler werden ermittelt und eingetragen.

**Tests** (`frontend/src/lib/errorMessages.test.ts`): Unit-Tests für `getActionErrorMessage` — je ein Test pro neuem Code, Server-Error-Zweig (5xx), `byCode`-Override-Priorität, Nicht-`BackendError`-Fallback.

### Acceptance criteria

- [ ] `BackendError.details` ist `unknown`; `ErrorResponseSchema.details` akzeptiert beliebige JSON-Werte
- [ ] Alle ~29 neuen Error-Codes haben eine deutsche, nicht-generische Message in `commonErrorMessages`
- [ ] `getActionErrorMessage` gibt bei `status >= 500` eine eigene Server-Error-Message zurück (nicht den `actionLabel`-Fallback)
- [ ] `position_nicht_bezahlbar` ist in `commonErrorMessages` eingetragen
- [ ] Alle leeren `MapError`-Maps im Backend sind behoben; `FavoritEntfernen` hat eine vollständige Map
- [ ] Unit-Tests für `getActionErrorMessage` laufen grün (`make check`)

---

## Phase 2: `useFormActionSubmit` Hook

**User stories**: 2, 3, 6, 15

### Context

- `frontend/src/hooks/use-action-submit.ts:1-40` — Vorlage für den neuen Hook (Interface, `useState`, `toast`, `getActionErrorMessage`)
- `frontend/src/lib/Backend.ts` — `BackendError.details` ist nach Phase 1 `unknown`
- `frontend/src/lib/errorMessages.ts` — `getActionErrorMessage` ist nach Phase 1 vollständig

### What to build

Neuer Hook `frontend/src/hooks/use-form-action-submit.ts` mit folgendem Verhalten:

**Signatur:**

```ts
useFormActionSubmit({
  form: UseFormReturn,
  actionLabel: string,
  byCode?: Record<string, string>,
  onSuccess?: () => void,
}) → { loading: boolean, run: (fn: () => Promise<void>) => Promise<void> }
```

**Bei `validation_error`**: `error.details` (jetzt `unknown`) wird als `Record<string, string[]>` interpretiert. Für jeden Eintrag wird `form.setError(field, { message: messages[0] })` aufgerufen. Kein Toast. Wenn `details` fehlt oder leer ist: Fallback-Toast mit der Message für `validation_error`.

**Bei allen anderen Fehlern**: wie `useActionSubmit` — Toast via `getActionErrorMessage({ actionLabel, error, byCode })`.

**Bei Erfolg**: `onSuccess?.()` aufgerufen; kein Toast, kein `setError`.

**Tests** (`frontend/src/hooks/use-form-action-submit.test.ts`): `renderHook` mit Vitest + Testing Library. Prüft: `validation_error` mit `details` → `form.setError` pro Feld, kein Toast; `validation_error` ohne `details` → Toast; anderer Fehlercode → Toast, kein `setError`; Erfolg → `onSuccess` aufgerufen.

### Acceptance criteria

- [ ] `useFormActionSubmit` existiert in `frontend/src/hooks/use-form-action-submit.ts`
- [ ] `validation_error` mit `details` ruft `form.setError` für jeden Feldpfad auf — kein Toast
- [ ] `validation_error` ohne `details` zeigt einen Fallback-Toast
- [ ] Andere Fehler zeigen einen Toast via `getActionErrorMessage`; kein `form.setError`
- [ ] Erfolg löst `onSuccess` aus; kein Toast, kein `setError`
- [ ] Unit-Tests laufen grün (`make check`)

---

## Phase 3: Migration — alle 16 Komponenten

**User stories**: alle (End-to-End-Verhalten der Nutzer)

### Context

- Alle 16 Komponenten aus der Inventar-Tabelle oben mit `try {…} catch (error) { toast.error(…) }`-Blöcken
- `frontend/src/hooks/use-action-submit.ts` — für Aktionen ohne Formular
- `frontend/src/hooks/use-form-action-submit.ts` — für Formulare mit `UseFormReturn` (nach Phase 2 verfügbar)
- `frontend/src/admin/tables/NewTischDialog.tsx` — Beispiel für `already_exists`-Fehler am Namensfeld via `byCode`

### What to build

Jede Komponente wird migriert:

- **Formulare** (`NewProductDialog`, `EditProductDialog`, `NewVariantDialog`, `EditVariantDialog`, `NewTischDialog`, `EditTischDialog`, `NewUserDialog`, `EditUserDialog`, `PasswordForm`, `DruckstationConfigPage`, Subformulare in `KassensitzungPage` und `EinstellungenPage`): manuelles `try/catch + setState(loading)` durch `useFormActionSubmit` ersetzen.

- **Reine Aktionen** (`ProductItem`, `Tische`, `Users`, `TischAuswahlDrawer`, Passwort-Reset in `EditUserDialog`): manuelles `try/catch + setState(loading)` durch `useActionSubmit` ersetzen.

- **`already_exists`-Fehler am Namensfeld** (`NewTischDialog`, `EditTischDialog`, `NewProductDialog`, `NewVariantDialog`, `NewUserDialog`): `byCode: { tisch_already_exists: '…', … }` wird an `useFormActionSubmit` übergeben, sodass der Fehler inline am `name`-Feld erscheint statt als generischer Toast.

- Alle manuellen `loading`-States, die durch den Hook ersetzt werden, werden entfernt. Kein toter Code.

Keine neuen Tests: die Logik ist durch Phase 1 (errorMessages) und Phase 2 (Hook) bereits getestet.

### Acceptance criteria

- [ ] Keine Komponente in `frontend/src/admin/` oder `frontend/src/service/` enthält noch manuelles `try/catch + toast.error()` (außer `Auth.ts`)
- [ ] Formulare mit `validation_error`-Fehlern zeigen inline Feldfehler statt Toast
- [ ] `tisch_already_exists`, `produkt_already_exists`, `username_already_exists` erscheinen am `name`-Feld der jeweiligen Dialoge
- [ ] `betreiber_nicht_konfiguriert` (bei Kassensitzung eröffnen) zeigt die handlungsorientierte Message aus `commonErrorMessages`
- [ ] `make check` läuft ohne Fehler (lint + unit tests)
