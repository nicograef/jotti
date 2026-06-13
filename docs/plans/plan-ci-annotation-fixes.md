# Plan: CI-Annotations sauber beheben

> Source PRD: n/a (Task-Beschreibung)

## Goal

Die beiden wiederkehrenden GitHub-Actions-Annotations dauerhaft an der Wurzel
beheben — nicht nur dort, wo sie beim `v0.9.0`-Release sichtbar wurden:

1. **Node-20-Deprecation** — `docker/setup-buildx-action@v3` und
   `softprops/action-gh-release@v2` laufen noch auf Node 20. Ab 16.06.2026
   erzwingt GitHub Node 24; ab 16.09.2026 ist Node 20 vom Runner entfernt.
2. **`Restore cache failed: … Supported file pattern: go.mod`** — `setup-go`
   sucht den Cache-Key anhand einer `go.sum`/`go.mod` im Repo-Root. Jotti hat
   dort kein Modul (alle Module liegen in Unterordnern), darum schlägt die
   Cache-Wiederherstellung in **jedem** `setup-go`-Job fehl.

## Inventory

**Node-20-Aktionen (nur in `release.yml`):**

- `.github/workflows/release.yml:32` — `docker/setup-buildx-action@v3`
- `.github/workflows/release.yml:120` — `softprops/action-gh-release@v2`

**`setup-go`-Jobs ohne `cache-dependency-path` (Annotation reproduziert auf
`local-proxy-ci`, Run 27464686400):**

- `.github/workflows/ci.yml:52` — `backend-ci` (Modul `backend/`)
- `.github/workflows/ci.yml:95` — `backend-golangci` (Modul `backend/`)
- `.github/workflows/ci.yml:114` — `resolver-ci` (Modul `resolver/`)
- `.github/workflows/ci.yml:151` — `local-proxy-ci` (Modul `reverse-proxy/`)
- `.github/workflows/ci.yml:182` — `cmd-ci` (Matrix `relay`/`starter`)
- `.github/workflows/ci.yml:282` — `backend-integration-tests` (Modul `backend/`)
- `.github/workflows/release.yml:28` — `release` (baut `cmd/starter` + `cmd/relay`)

**Modul-/`go.sum`-Lage (entscheidet Cache-Strategie pro Job):**

- `backend/go.sum` — vorhanden (54 Zeilen, externe Deps)
- `resolver/go.sum` — vorhanden (14 Zeilen)
- `reverse-proxy/go.sum` — vorhanden (2 Zeilen)
- `cmd/starter/` — **kein** `go.sum` (stdlib-only)
- `cmd/relay/` — **kein** `go.sum` (stdlib-only)

**Bestätigte Zielversionen (per GitHub-API geprüft, `using: node24`):**

- `docker/setup-buildx-action` → `v4` (latest v4.1.0, node24)
- `softprops/action-gh-release` → `v3` (latest v3.0.0, node24)

## Resolved decisions

- **Scope: repo-weit.** Beide Workflows werden angefasst, nicht nur
  `release.yml`. Die Cache-Annotation ist Root-Cause-gleich in ~6 CI-Jobs.
- **Pin-Stil: Major-Tag** (`@v4`, `@v3`) — konsistent mit allen übrigen
  Action-Pins im Repo (`@v6`, `@v9`, `@v4`). Kein SHA-Pinning.
- **Cache-Strategie nach Modultyp:**
  - Module **mit** `go.sum` (`backend`, `resolver`, `reverse-proxy`) →
    `cache-dependency-path` auf die jeweilige `go.sum`. Cache bleibt aktiv,
    Warnung verschwindet.
  - Module **ohne** `go.sum` (`cmd/starter`, `cmd/relay` sowie der
    `release`-Job, der genau diese baut) → `cache: false`. Es gibt nichts zu
    cachen; das ist die ehrliche, warnungsfreie Lösung statt eines
    Pseudo-Cache-Keys.

## Open questions / Risks

- **`action-gh-release` v2 → v3** ist ein Major-Sprung. Vor dem Merge die
  Release-Notes auf Breaking Changes prüfen, die unsere genutzten Inputs
  (`files`, `generate_release_notes`) betreffen. Erwartung: keine — beide
  Inputs sind stabil. Verifikation erfolgt ohnehin durch den Release-Dry-Run.
- **`setup-buildx-action` v3 → v4** gilt als Drop-in für `--load`-Builds; kein
  Konfig-Change an unserer Nutzung nötig.

---

## Phase 1: Node-24-Aktionen in `release.yml`

### Context

- `.github/workflows/release.yml:32` — `docker/setup-buildx-action@v3` → `@v4`
- `.github/workflows/release.yml:120` — `softprops/action-gh-release@v2` → `@v3`

### What to build

Beide in `release.yml` als Node-20 markierten Actions auf ihre Node-24-Majors
heben (`setup-buildx-action@v4`, `action-gh-release@v3`). Keine weiteren
Konfig-Änderungen an den Steps, sofern die v3-Release-Notes von
`action-gh-release` keine Breaking Changes für `files`/`generate_release_notes`
ausweisen.

### Acceptance criteria

- [x] `release.yml` referenziert `docker/setup-buildx-action@v4` und
      `softprops/action-gh-release@v3`.
- [x] Release-Notes von `action-gh-release@v3` auf Breaking Changes geprüft;
      unsere Inputs bleiben gültig. (v3.0.0: einziger Breaking Change ist der
      Node-20→24-Runtime-Wechsel; `files` und `generate_release_notes` sind in
      `action.yml@v3.0.0` unverändert definiert.)
- [ ] Release-Dry-Run (`gh workflow run release.yml --ref main`) läuft grün
      durch (Build + Smoke-Test). — offen: läuft erst nach Merge auf `main`
      (workflow_dispatch nutzt den Ref); Trigger durch den Maintainer.
- [ ] Die Node-20-Deprecation-Annotation erscheint im Dry-Run **nicht** mehr.
      — offen: wird mit dem Dry-Run oben verifiziert.

---

## Phase 2: `go.mod`-Cache repo-weit auflösen

### Context

- `.github/workflows/ci.yml:52,95,282` — `backend`-Jobs → `cache-dependency-path: backend/go.sum`
- `.github/workflows/ci.yml:114` — `resolver-ci` → `cache-dependency-path: resolver/go.sum`
- `.github/workflows/ci.yml:151` — `local-proxy-ci` → `cache-dependency-path: reverse-proxy/go.sum`
- `.github/workflows/ci.yml:182` — `cmd-ci` (stdlib-only Matrix) → `cache: false`
- `.github/workflows/release.yml:28` — `release` (baut stdlib-only cmd-Module) → `cache: false`

### What to build

Jeden `setup-go`-Step so konfigurieren, dass der Cache-Key auflösbar ist:
Module mit `go.sum` bekommen ein passendes `cache-dependency-path`, stdlib-only
Module bekommen `cache: false`. Damit verschwindet die Annotation in CI **und**
Release, und der Go-Modul-Cache funktioniert dort, wo er etwas bringt.

### Acceptance criteria

- [x] Alle `setup-go`-Steps in `ci.yml` und `release.yml` setzen entweder
      `cache-dependency-path` (Module mit `go.sum`) oder `cache: false`
      (stdlib-only). (6 Steps in `ci.yml`: `backend`/`backend-golangci`/
      `backend-integration-tests` → `backend/go.sum`, `resolver` →
      `resolver/go.sum`, `local-proxy` → `reverse-proxy/go.sum`, `cmd` →
      `cache: false`; 1 Step in `release.yml` → `cache: false`.)
- [ ] Ein CI-Lauf, der `backend`, `resolver`, `reverse-proxy` und `cmd`
      berührt, zeigt **keine** `Restore cache failed`-Annotation mehr.
      — offen: verifiziert sich beim ersten CI-Lauf nach Merge.
- [ ] Für Module mit `go.sum` greift der Cache (Job-Log: „Cache restored“ /
      „Cache saved“ statt Fehlschlag). — offen: siehe CI-Lauf oben.
- [ ] Der Release-Dry-Run zeigt die Cache-Annotation **nicht** mehr.
      — offen: gemeinsam mit dem Phase-1-Dry-Run nach Merge.
