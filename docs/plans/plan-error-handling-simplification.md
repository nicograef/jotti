# Plan: Vereinfachung `useFormActionSubmit` — ein Inline-Mechanismus, klare Semantik

> Source PRD: n/a (Quelle: Code-Audit der Error-Handling-Implementierung, Commits `d9c1d69`–`d77b6d4`)

## Goal

`useFormActionSubmit` wird von 169 auf ~35 Zeilen reduziert. Es gibt genau **einen** Mechanismus für Inline-Feldfehler (`fieldErrorsByCode`) und genau **eine** Bedeutung für `byCode` (Toast-Override, identisch zu `useActionSubmit`). Der defekte Backend-`validation_error.details`→Feld-Pfad wird ersatzlos gestrichen, ebenso die Magic-Suffix-Heuristik und der handgerollte Runtime-Type-Guard. Tote Error-Code-Mappings entfallen.

## Architectural decisions

- **Kein Backend-Touch.** Das Error-Format `{ code, details? }` und alle Handler bleiben unverändert. `validation_error.details` wird vom Frontend nicht mehr interpretiert (nur noch geloggt via `BackendError`-Message).
- **`byCode` = ausschließlich Toast-Override** — in beiden Hooks identisch. Es erzeugt nie Feldfehler.
- **`fieldErrorsByCode` = einziger Inline-Mechanismus**: `Record<code, Record<feldname, message>>`, explizit pro Call-Site. Kein implizites Erraten des Feldes aus dem Code-String.
- **Server-Validierung erscheint als Toast.** Client-seitige Zod-Validierung (`zodResolver`, Submit-Sperre bei `!isValid`) bleibt die einzige Quelle für Inline-Validierungsfehler.

## Inventory

- `frontend/src/hooks/use-form-action-submit.ts:18-42` — `toValidationDetails`: handgerollter `unknown`→`Record<string,string[]>`-Guard; entfällt
- `frontend/src/hooks/use-form-action-submit.ts:44-61` — `applyValidationErrors`: mappt Backend-details auf Felder; entfällt (defekt: Backend sendet PascalCase-Keys mit englischen zog-Messages, RHF-Felder sind camelCase — empirisch verifiziert: `{"Name":["is required"],…}`)
- `frontend/src/hooks/use-form-action-submit.ts:63-90` — `applyMappedCodeError`: Magic-Suffix-Heuristik (`endsWith('_already_exists')` → `'name'`, Sonderfall `username_already_exists`); entfällt
- `frontend/src/hooks/use-form-action-submit.ts:92-120` — `applyExplicitCodeErrors`: expliziter `fieldErrorsByCode`-Mechanismus; bleibt als einziger, inline in `run`
- `frontend/src/hooks/use-form-action-submit.test.ts:25-53` — Test konstruiert camelCase-details, die das Backend nie sendet; entfällt
- `frontend/src/hooks/use-form-action-submit.test.ts:128-151` — Test der Magic-Suffix-Heuristik; entfällt
- `frontend/src/hooks/use-action-submit.ts:1-40` — Referenz für die Ziel-Semantik von `byCode` (Toast-Override); bleibt unverändert
- `frontend/src/lib/errorMessages.ts:72` — `validation_error: 'Bitte die markierten Eingaben prüfen.'` — Text referenziert Markierungen, die es nicht mehr gibt
- `frontend/src/components/common/PasswordForm.tsx:36-48` — einziger bestehender `fieldErrorsByCode`-Nutzer; Referenzmuster für die Migration, unverändert
- Call-Sites mit `byCode: { *_already_exists: … }` am Form-Hook (Migration auf `fieldErrorsByCode`):
  - `frontend/src/admin/tables/NewTischDialog.tsx:42-48` (`tisch_already_exists` → Feld `name`)
  - `frontend/src/admin/tables/EditTischDialog.tsx:41-47` (`tisch_already_exists` → `name`)
  - `frontend/src/admin/products/NewProductDialog.tsx:62-68` (`produkt_already_exists` → `name`)
  - `frontend/src/admin/products/EditProductDialog.tsx:61-67` (`produkt_already_exists` → `name`)
  - `frontend/src/admin/users/NewUserDialog.tsx:46-52` (`username_already_exists` → `username`)
  - `frontend/src/admin/users/EditUserDialog.tsx:51-57` (`username_already_exists` → `username`)
- Tote Mappings (Code existiert im Backend nicht — kein `ErrVarianteAlreadyExists`, kein Vorkommen):
  - `frontend/src/admin/products/NewVariantDialog.tsx:46-52` (`variante_already_exists`)
  - `frontend/src/admin/products/EditVariantDialog.tsx:44-50` (`variante_already_exists`)
- `frontend/src/admin/kasse/KassensitzungPage.tsx:108-111,215-218,335-338` — Form-Hook ohne `byCode`/`fieldErrorsByCode`; funktioniert unverändert, keine Migration nötig
- `backend/api/user/http/command_handler.go:33-37` + `backend/api/helper/http.go:76-86` — Beleg für das PascalCase/Englisch-Format der details; bleibt unverändert

## Resolved decisions

- Backend-`details`→Feld-Pfad wird **gestrichen**, nicht repariert (Nutzer-Entscheidung). Begründung: Pfad ist defekt, alle RHF-Formulare validieren bereits client-seitig und sperren den Submit bei `!isValid`.
- Hook-Komposition (`useFormActionSubmit` auf `useActionSubmit` aufsetzen) wird **nicht** umgesetzt — beide Hooks bleiben eigenständig und trivial lesbar.
- **Eine Phase**: Alles hängt an derselben Signatur-Änderung; ein Schnitt vermeidet einen Zwischenzustand mit degradiertem Verhalten (Inline-Fehler → Toast).
- `variante_already_exists`-Mappings werden ersatzlos entfernt (totes Mapping, Code wird nie gesendet).
- `validation_error`-Message wird sprachlich angepasst (keine markierten Felder mehr).

## Open questions / Risks

- **Varianten-Namens-Eindeutigkeit**: Das Backend kennt keinen `variante_already_exists`-Fehler — Duplikate werden entweder erlaubt oder landen als generischer Fehler. Ob eine Eindeutigkeitsprüfung gewünscht ist, ist ein separates Feature (out of scope, ggf. eigenes PRD).
- Falls backend-getriebene Inline-Validierung später doch gewünscht wird: erfordert `zog:"…"`-Tags (camelCase) an allen Request-Structs plus deutsche zog-Messages — bewusst verschoben.

---

## Phase 1: Hook vereinfachen, Call-Sites migrieren, Tests angleichen

### Context

- `frontend/src/hooks/use-form-action-submit.ts:1-169` — wird auf ~35 Zeilen reduziert
- `frontend/src/hooks/use-form-action-submit.test.ts:1-173` — Tests werden an die neue Semantik angeglichen
- `frontend/src/hooks/use-action-submit.ts:19-37` — `run`-Gerüst als Stil-Referenz (try/catch/finally, `console.error`, Toast)
- `frontend/src/lib/errorMessages.ts:72` — Message-Text `validation_error`
- 8 Dialog-Call-Sites aus dem Inventory (6 Migrationen + 2 tote Mappings)

### What to build

Der Form-Hook behält Signatur-Felder `form`, `actionLabel`, `byCode`, `fieldErrorsByCode`, `onSuccess` und das `{ loading, run }`-Interface — Verhalten im Fehlerfall:

1. Ist der Fehler ein `BackendError` und `fieldErrorsByCode[error.code]` definiert → `form.setError(feld, { message })` pro Eintrag, kein Toast.
2. Sonst → Toast via `getActionErrorMessage({ actionLabel, error, byCode })` (identisch zu `useActionSubmit`).

Entfernt werden: `toValidationDetails`, `applyValidationErrors`, `applyMappedCodeError` sowie das Auslesen von `error.details`. Die vier Helper entfallen; die verbleibende Logik steht inline in `run`.

Die sechs `*_already_exists`-Dialoge übergeben ihr Mapping explizit als `fieldErrorsByCode: { <code>: { name|username: '…' } }` (Messages unverändert). Die beiden Varianten-Dialoge verlieren ihr totes `variante_already_exists`-Mapping ersatzlos. `PasswordForm` und die drei `KassensitzungPage`-Subformulare bleiben unverändert.

`errorMessages.ts`: `validation_error`-Text zu „Bitte die Eingaben prüfen und erneut versuchen." ändern (keine Felder werden mehr markiert).

Tests: Der camelCase-details-Test und der Magic-Suffix-Test entfallen. Verbleibende/neue Fälle: `fieldErrorsByCode`-Treffer → `setError` pro Feld, kein Toast; `BackendError` ohne `fieldErrorsByCode`-Treffer → Toast, kein `setError`; `validation_error` → Toast mit neuer Message; Erfolg → `onSuccess`, weder Toast noch `setError`. Kein Test konstruiert ein Backend-Format, das real nicht existiert.

### Acceptance criteria

- [x] `use-form-action-submit.ts` enthält kein `error.details`-Parsing, keine Suffix-Heuristik und keine Helper-Funktionen mehr; nur `fieldErrorsByCode` (inline) und `byCode` (Toast)
- [x] `byCode` hat in beiden Hooks identische Bedeutung (Toast-Override); kein `byCode`-Eintrag erzeugt einen Feldfehler
- [x] `tisch_already_exists`, `produkt_already_exists`, `username_already_exists` erscheinen weiterhin inline am `name`- bzw. `username`-Feld der sechs Dialoge
- [x] `variante_already_exists` kommt im Frontend nicht mehr vor
- [x] `validation_error`-Toast-Text referenziert keine „markierten" Eingaben mehr
- [x] Kein Test konstruiert `validation_error.details` in einem Format, das das Backend nicht sendet
- [x] `make check-frontend` läuft ohne Fehler (Format, Lint, Tests, Build)
