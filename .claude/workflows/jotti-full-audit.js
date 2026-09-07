export const meta = {
  name: 'jotti-full-audit',
  description: 'Full-repo multi-expert review of jotti (every code, doc and config file) with Fable reviewers, cleanup criteria, adversarial verification and a consolidated findings document',
  phases: [
    { title: 'Review', detail: '22 units × 3 lenses + 8 cross-layer flows', model: 'fable' },
    { title: 'Verify', detail: 'refuters for blocker/major findings', model: 'fable' },
    { title: 'Consolidate', detail: 'per-area consolidation and findings document' },
  ],
}

// Reusable named workflow: Workflow({ name: 'jotti-full-audit', args: { date: 'YYYY-MM-DD' } })
// Optional args: repo (default /home/user/jotti), handbook (cleanup skill dir), outFile.
// Model policy of the orchestrator plan: reviewers and skeptics run on Fable 5.1 (explicit
// owner decision for the audit, overriding the general "no Fable" rule in CLAUDE.md);
// the assembler is mechanical and runs on Opus.
const A = args || {}
const REPO = A.repo || '/home/user/jotti'
const HANDBOOK = A.handbook || '/home/user/handbook/.claude/skills/cleanup'
const OUT = A.outFile || `${REPO}/docs/plans/findings-jotti-audit.md`
const DATE = A.date || 'unbekannt'

const UNITS = [
  { key: 'api-kasse', kind: 'go', paths: ['backend/api/kasse'] },
  { key: 'api-fiskal', kind: 'go', paths: ['backend/api/fiskal'] },
  { key: 'api-druck', kind: 'go', paths: ['backend/api/druck'] },
  { key: 'api-stammdaten', kind: 'go', paths: ['backend/api/stammdaten'] },
  { key: 'api-core', kind: 'go', paths: ['backend/api/*.go', 'backend/api/auth', 'backend/api/middleware', 'backend/api/health', 'backend/api/helper', 'backend/api/test', 'backend/api/reporting'] },
  { key: 'domain-kasse', kind: 'go', paths: ['backend/domain/kasse'] },
  { key: 'domain-rest', kind: 'go', paths: ['backend/domain/tse', 'backend/domain/steuer', 'backend/domain/event', 'backend/domain/jwt', 'backend/domain/user', 'backend/domain/produkt', 'backend/domain/druckstation', 'backend/domain/tisch', 'backend/domain/betreiber', 'backend/domain/reporting'] },
  { key: 'repo-fiscal', kind: 'go', paths: ['backend/repository/tse_repo', 'backend/repository/kassenjournal_repo', 'backend/repository/druckauftrag_repo', 'backend/repository/kassensitzungen_repo'] },
  { key: 'repo-stammdaten', kind: 'go', paths: ['backend/repository/produkt_repo', 'backend/repository/user_repo', 'backend/repository/tisch_repo', 'backend/repository/druckstation_repo', 'backend/repository/betreiber_repo', 'backend/repository/favorit_repo', 'backend/repository/reporting_repo'] },
  { key: 'backend-infra', kind: 'go', paths: ['backend/app', 'backend/bootstrap', 'backend/config', 'backend/db', 'backend/seed', 'backend/dsfinvkpruefung', 'backend/sqlc/queries', 'backend/sqlc.yaml', 'backend/*.go', 'backend/go.mod', 'backend/Dockerfile'] },
  { key: 'fe-service-components', kind: 'ts', paths: ['frontend/src/service/components'] },
  { key: 'fe-service-rest', kind: 'ts', paths: ['frontend/src/service/*.ts', 'frontend/src/service/*.tsx', 'frontend/src/service/table', 'frontend/src/service/direktverkauf', 'frontend/src/service/product'] },
  { key: 'fe-admin-fiscal', kind: 'ts', paths: ['frontend/src/admin/reporting', 'frontend/src/admin/kasse', 'frontend/src/admin/tse', 'frontend/src/admin/finanzamt'] },
  { key: 'fe-admin-stammdaten', kind: 'ts', paths: ['frontend/src/admin/users', 'frontend/src/admin/products', 'frontend/src/admin/tables', 'frontend/src/admin/settings', 'frontend/src/admin/components', 'frontend/src/admin/*.tsx', 'frontend/src/admin/*.ts'] },
  { key: 'fe-shared', kind: 'ts', paths: ['frontend/src/components', 'frontend/src/hooks', 'frontend/src/lib', 'frontend/src/pages', 'frontend/src/*.ts', 'frontend/src/*.tsx', 'frontend/src/test', 'frontend/*.json', 'frontend/*.ts', 'frontend/*.js', 'frontend/Dockerfile', 'frontend/nginx.conf', 'frontend/index.html'] },
  { key: 'website', kind: 'ts', paths: ['website/src', 'website/*.mjs', 'website/*.json', 'website/Dockerfile', 'website/nginx.conf', 'website/public'] },
  { key: 'e2e', kind: 'ts', paths: ['e2e'] },
  { key: 'windows', kind: 'go', paths: ['windows', 'packaging/windows'] },
  { key: 'edge', kind: 'go', paths: ['reverse-proxy', 'resolver'] },
  { key: 'ops', kind: 'ops', paths: ['scripts', 'packaging/cron', 'packaging/systemd', 'Makefile', 'docker-compose*.yml', '.env.example', '.env.fiskaly-test.example', '.github', 'go.work', 'cliff.toml', '.dockerignore', '.editorconfig', '.gitignore', 'database'] },
  { key: 'docs-user', kind: 'md', paths: ['docs/leitfaden', 'README.md', 'SERVICE.md', 'TERMS.md', 'CLA.md', 'LICENSE', 'CHANGELOG.md'] },
  { key: 'docs-internal', kind: 'md', paths: ['docs/*.md', 'docs/adrs', 'docs/prds', 'docs/plans', 'AGENTS.md', 'CLAUDE.md', '.github/instructions', '.github/copilot-instructions.md'] },
]

const FLOWS = [
  'Tischbestellung → Zahlung kassieren → Kassenbeleg drucken (frontend service → api/kasse/tischgeschaeft → domain/kasse events → kassenjournal_repo → projections → api/druck/beleg → escpos → relay)',
  'Direktverkauf → Abholbon/Arbeitsbon-Routing → Druckauftrag → Windows relay (frontend direktverkauf → api/kasse/direktverkauf → arbeitsbon_policy → druckauftrag_repo → windows/relay)',
  'TSE-Signatur-Pipeline: Signaturauftrag-Outbox → fiskaly client → Störungsprotokoll → Nachsignierung → Belegvermerk (domain/tse, tse_repo, api/fiskal/signatur, escpos formatter)',
  'Kassensitzung eröffnen → Kassensturz → Tagesabschluss (Z-Bon) → DSFinV-K-Export → dsfinvkpruefung (api/kasse/kassenfuehrung, reporting, api/fiskal export, frontend admin/kasse + finanzamt)',
  'Stammdaten-CRUD und Soft-Delete: Produkte/Varianten/Steuersätze, Tische, Benutzer, Betreiber (frontend admin → api/stammdaten → repos → sqlc → migrations) inkl. Auswirkung auf laufende Kassensitzung',
  'Auth und Onboarding: Bootstrap-Admin mit Einmalpasswort → Login → JWT → Rollen-Guards (admin/serviceleitung/service) → Frontend Guards → Middleware (bootstrap, api/auth, middleware, domain/jwt, frontend guards)',
  'Windows-Starter: Erststart, Zertifikat/lokal.jotti.rocks, Update mit Backup, Restore/Repair, Statusseite (windows/starter, reverse-proxy, resolver, packaging/windows, docker-compose.release.yml, docs/leitfaden)',
  'Release- und Update-Pfad: Makefile release targets → .github/workflows/release.yml → Images/ZIP → prod-update script → PREVIOUS_VERSION upgrade-path gate → migrations README rules',
]

const COMMON = `
Context: jotti (${REPO}) is a German mobile POS for club festivals: Go backend (stdlib net/http, pgx, sqlc, zog), React/TypeScript frontend, Astro website, Windows starter/relay in Go, Docker Compose, PostgreSQL, fiskaly cloud TSE, DSFinV-K export. Read ${REPO}/AGENTS.md first (rules, quality metrics, ubiquitous language, freeze discipline, rule 18 current-state-only docs). Consult ${REPO}/docs/language.md for naming and ${REPO}/docs/handbuch.md for architecture when judging boundaries.
You are a REVIEWER at the highest rigor. Do NOT modify any file in ${REPO}. Read the actual files completely (use git ls-files with the given path globs to enumerate; skip generated code under backend/sqlc/dbgen and lockfiles). Every finding must be anchored to file path and exact line range and hold on re-read; drop anything you cannot anchor. Prefer fewer, true findings over many weak ones. "No issues" is a valid result for a file. Never propose functionality changes disguised as cleanup; flag large refactors as such (effort L, category large-refactor) and move on.`

const FINDINGS = {
  type: 'object',
  properties: {
    findings: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          severity: { type: 'string', enum: ['blocker', 'major', 'minor'] },
          category: { type: 'string', description: 'boundary-consistency | correctness | security | readability | principle | code-smell | architecture | docs-accuracy | convention | test-quality | ops | large-refactor' },
          file: { type: 'string' },
          lines: { type: 'string', description: 'e.g. 42-58' },
          what: { type: 'string', description: 'rule or defect, reference the cleanup reference file and rule name where applicable' },
          why: { type: 'string', description: 'one sentence impact' },
          suggestion: { type: 'string', description: 'concrete minimal change' },
          effort: { type: 'string', enum: ['S', 'M', 'L'] },
          proof: { type: 'string', description: 'quoted code/doc lines or command output that shows it' },
        },
        required: ['severity', 'category', 'file', 'lines', 'what', 'why', 'suggestion', 'effort', 'proof'],
      },
    },
    filesReviewed: { type: 'integer' },
    summary: { type: 'string' },
  },
  required: ['findings', 'filesReviewed', 'summary'],
}

const LENSES = {
  cleanup: (u) => `Lens: CLEANUP (repo-wide scope mode of the cleanup skill). Read the reference files ${HANDBOOK}/readability.md, ${u.kind === 'md' ? HANDBOOK + '/readability-de.md (for German prose),' : ''} ${HANDBOOK}/principles.md, ${HANDBOOK}/code-smells.md, ${HANDBOOK}/architecture.md and apply their passes to every file in scope: readability & clarity (incl. AI-slop), principles (SOLID, DRY, KISS, YAGNI), code smells (code and config), architecture & boundaries (service/domain/handler/repository layers, dependency direction), test readability for test files. Report each issue once under the most specific pass, with the reference file and rule name in "what". Never change functionality; large refactors are flagged, not designed.`,
  correctness: (u) => `Lens: CORRECTNESS AND SECURITY. Hunt real defects: logic errors, off-by-one, nil/undefined handling, error swallowing, races and transaction boundaries, money not in cents, timezone/date handling, input validation at edges (HTTP handlers, CLI flags, env, files), authz gaps (role guards, IDOR), injection (SQL via sqlc params only?, shell in scripts, path traversal in starter/relay), secrets handling, TLS/cert handling in reverse-proxy/resolver/starter, event-sourcing invariants (append-only journal, projections in same transaction, frozen event contracts), fiscal rules (TSE signing outbox, Belegpflichtangaben, DSFinV-K fields) against docs/compliance.md and docs/rechtsquellen where a claim needs the primary text. For docs units: factual errors versus the code (commands, paths, flags, behaviour), broken or misleading instructions. Give proof by quoting code. Run cheap checks where useful (go vet on a package, grep for patterns) but do not build the whole project.`,
  conventions: (u) => `Lens: CONVENTIONS, DOCS AND CONSISTENCY. Check against AGENTS.md rules 1–18, docs/language.md naming per layer (Go, TS, JSON, DB), docs/handbuch.md, README single-source rule, version consistency (Go, Node, pnpm, TypeScript versions across go.mod, Dockerfiles, CI, AGENTS.md, docs), rule 18 (no dated change entries, no "früher/bisher", no deprecation notes outside CHANGELOG/ADRs), dead links and stale paths in Markdown and comments, German user-visible strings, English infrastructure code, comments that describe a state the code no longer has, duplicated content across files, TODO/FIXME left behind, test naming and coverage gaps for public behaviour. For ops units: Makefile targets vs docs, compose files consistency, CI workflow correctness and pinning, script safety (set -euo pipefail, quoting), packaging text accuracy.`,
}

phase('Review')
log(`Review: ${UNITS.length} Einheiten × 3 Linsen + ${FLOWS.length} Cross-Layer-Flüsse (Fable)`)
const reviewItems = []
for (const u of UNITS) for (const lens of Object.keys(LENSES)) reviewItems.push({ u, lens })

const reviews = await parallel(
  reviewItems.map(({ u, lens }) => () =>
    agent(
      `${COMMON}\n\nUNIT "${u.key}" — scope (git ls-files globs, relative to ${REPO}): ${u.paths.join(', ')}. Enumerate the files first, read all of them, then review.\n\n${LENSES[lens](u)}\n\nReturn findings for this unit only. Set filesReviewed to the number of files you actually read.`,
      { label: `review:${u.key}:${lens}`, phase: 'Review', schema: FINDINGS, model: 'fable' },
    ),
  ),
)

const flows = await parallel(
  FLOWS.map((flow, i) => () =>
    agent(
      `${COMMON}\n\nLens: CROSS-LAYER TRACE (cleanup skill repo-wide mode, read ${HANDBOOK}/cross-layer.md first). Trace this flow end to end through every layer and file it touches:\n${flow}\n\nLook for shape and validation mismatches between layers (DTO vs domain vs DB vs frontend schema), inconsistent naming for the same concept across layers, error paths that lose information or leave state inconsistent, missing tests for the flow's invariants, and docs (handbuch, language, leitfaden) that describe the flow differently from the code. Cite every file:line.`,
      { label: `flow:${i + 1}`, phase: 'Review', schema: FINDINGS, model: 'fable' },
    ),
  ),
)

const raw = []
reviews.forEach((r, i) => { if (r) r.findings.forEach((f) => raw.push({ ...f, unit: reviewItems[i].u.key, lens: reviewItems[i].lens })) })
flows.forEach((r, i) => { if (r) r.findings.forEach((f) => raw.push({ ...f, unit: `flow-${i + 1}`, lens: 'cross-layer' })) })
const filesReviewed = reviews.filter(Boolean).reduce((a, r) => a + (r.filesReviewed || 0), 0)
const failedReviewers = reviews.filter((r) => !r).length + flows.filter((r) => !r).length

// dedupe by file + first line + severity-insensitive what prefix
const seen = new Map()
for (const f of raw) {
  const firstLine = String(f.lines).split(/[-–,]/)[0].trim()
  const key = `${f.file}:${firstLine}:${f.what.toLowerCase().slice(0, 40)}`
  if (!seen.has(key)) seen.set(key, f)
  else {
    const prev = seen.get(key)
    const rank = { blocker: 3, major: 2, minor: 1 }
    if (rank[f.severity] > rank[prev.severity]) seen.set(key, { ...f, alsoFrom: prev.lens })
  }
}
const deduped = [...seen.values()]
log(`Review fertig: ${raw.length} Rohbefunde, ${deduped.length} nach Dedupe, ${filesReviewed} Dateien gelesen, ${failedReviewers} Reviewer ohne Ergebnis`)

phase('Verify')
const VERDICT = { type: 'object', properties: { refuted: { type: 'boolean' }, reason: { type: 'string' } }, required: ['refuted', 'reason'] }
const toVerify = deduped.filter((f) => f.severity !== 'minor')
const CAP = 400
const capped = toVerify.length > CAP
const verifyList = capped ? toVerify.sort((a, b) => (a.severity === 'blocker' ? -1 : 1) - (b.severity === 'blocker' ? -1 : 1)).slice(0, CAP) : toVerify
if (capped) log(`Verifikation gekappt: ${verifyList.length} von ${toVerify.length} Blocker/Major-Befunden werden geprüft (Blocker zuerst); der Rest bleibt „unverifiziert"`)
const verifiedSet = new Map()
const verified = await parallel(
  verifyList.map((f, idx) => () => {
    const modes = f.severity === 'blocker' ? ['source', 'consequence', 'rules'] : ['source', 'consequence']
    return parallel(
      modes.map((mode) => () =>
        agent(
          `${COMMON}\n\nYou are a SKEPTIC (${mode}). A reviewer (unit ${f.unit}, lens ${f.lens}) claims:\nFILE: ${f.file}:${f.lines}\nSEVERITY: ${f.severity} (${f.category})\nWHAT: ${f.what}\nWHY: ${f.why}\nSUGGESTION: ${f.suggestion}\nPROOF: ${f.proof}\n\nTry to REFUTE it by ${mode === 'source' ? 're-reading the exact lines and surrounding code/docs: is the claim literally true at that location?' : mode === 'consequence' ? 'asking whether it matters: is there a reachable path, a real reader confusion, or a real maintenance cost? Is the suggestion minimal and behaviour-preserving (for cleanup) or correct (for defects)?' : 'checking the suggestion against AGENTS.md, freeze discipline, rule 18, naming conventions and the product conservatism: would applying it break a rule or widen scope?'}\nDefault to refuted=true when uncertain or when the issue is cosmetic taste rather than a rule or defect.`,
          { label: `verify:${idx}:${mode}`, phase: 'Verify', schema: VERDICT, model: 'fable' },
        ),
      ),
    ).then((votes) => {
      const v = votes.filter(Boolean)
      const notRefuted = v.filter((x) => !x.refuted).length
      const stands = f.severity === 'blocker' ? notRefuted >= 2 : notRefuted === v.length && v.length > 0
      return { ...f, verified: stands ? 'bestätigt' : 'verworfen', votes: v.map((x) => x.reason) }
    })
  }),
)
verified.filter(Boolean).forEach((f) => verifiedSet.set(`${f.file}:${f.lines}:${f.what.slice(0, 40)}`, f))
const finalFindings = deduped.map((f) => {
  const v = verifiedSet.get(`${f.file}:${f.lines}:${f.what.slice(0, 40)}`)
  if (v) return v
  return { ...f, verified: f.severity === 'minor' ? 'unverifiziert (minor)' : 'unverifiziert (Kappung)' }
}).filter((f) => f.verified !== 'verworfen')
const dropped = verified.filter(Boolean).filter((f) => f.verified === 'verworfen')
log(`Verifikation: ${verified.filter(Boolean).length} geprüft, ${dropped.length} verworfen, ${finalFindings.length} Befunde bleiben`)

phase('Consolidate')
const AREAS = {
  Backend: ['api-kasse', 'api-fiskal', 'api-druck', 'api-stammdaten', 'api-core', 'domain-kasse', 'domain-rest', 'repo-fiscal', 'repo-stammdaten', 'backend-infra'],
  Frontend: ['fe-service-components', 'fe-service-rest', 'fe-admin-fiscal', 'fe-admin-stammdaten', 'fe-shared'],
  'Website und E2E': ['website', 'e2e'],
  'Windows, Edge und Ops': ['windows', 'edge', 'ops'],
  Dokumentation: ['docs-user', 'docs-internal'],
  'Cross-Layer-Flüsse': FLOWS.map((_, i) => `flow-${i + 1}`),
}
const SECTION = { type: 'object', properties: { markdown: { type: 'string' }, defectClasses: { type: 'array', items: { type: 'string' } }, top: { type: 'array', items: { type: 'string' } } }, required: ['markdown', 'defectClasses', 'top'] }
const sections = await parallel(
  Object.entries(AREAS).map(([area, units]) => () => {
    const fs = finalFindings.filter((f) => units.includes(f.unit))
    return agent(
      `${COMMON}\n\nYou are the CONSOLIDATOR for the area "${area}". Input: ${fs.length} findings (JSON below). Produce a German Markdown section for a self-contained findings document: heading "## ${area}", one-line scope, a severity count line, then findings grouped by file (sorted by severity: blocker, major, minor), each in this exact shape:\n\n**[What]** (Kategorie, Referenz)\nDatei: pfad:zeilen\nWarum: <ein Satz>\nVorschlag: <konkrete minimale Änderung>\nAufwand: S|M|L · Status: bestätigt | unverifiziert (minor) | unverifiziert (Kappung)\n\nMerge near-duplicates that survived dedupe, keep the stronger proof, never invent findings, never drop a 'bestätigt' finding. Sentences ≤ 20 words. Also return defectClasses (recurring classes that deserve a lint/gate, one line each) and top (up to 5 highest-impact items for this area as one-liners with file refs).\n\nFINDINGS:\n${JSON.stringify(fs.map(({ votes, alsoFrom, ...rest }) => rest), null, 0)}`,
      { label: `consolidate:${area}`, phase: 'Consolidate', schema: SECTION, model: 'fable' },
    )
  }),
)

const counts = { blocker: 0, major: 0, minor: 0 }
finalFindings.forEach((f) => { counts[f.severity]++ })
const header = `# Findings: Vollreview jotti (${DATE})

> Quelle: Multi-Experten-Review aller Dateien des Repos (Stand \`main\` @ 2ee9cbaa plus Branch-Pläne),
> ${UNITS.length} Einheiten × 3 Linsen (Cleanup-Skill, Korrektheit/Security, Konventionen/Doku) plus
> ${FLOWS.length} Cross-Layer-Flüsse; Reviewer und Prüfer: Fable 5.1. Jeder Blocker/Major-Befund wurde von
> ${'2–3'} unabhängigen Skeptikern gegengeprüft${capped ? ` (gekappt auf ${CAP} Befunde, Blocker zuerst)` : ''}; Minor-Befunde sind ungeprüft.
> Ausgeschlossen: \`backend/sqlc/dbgen/\`, Lockfiles, Binärdateien, \`docs/rechtsquellen/\`.
> Dieses Dokument ist die Eingabe für \`plan-jotti-audit-fixes.md\` (create-plan). Es enthält keine Personendaten.

## Zahlen

| Kennzahl | Wert |
| --- | --- |
| Gelesene Dateien (laut Reviewern) | ${filesReviewed} |
| Rohbefunde | ${raw.length} |
| Nach Dedupe | ${deduped.length} |
| Verifiziert | ${verified.filter(Boolean).length} |
| Verworfen | ${dropped.length} |
| Verbleibend | ${finalFindings.length} (Blocker ${counts.blocker}, Major ${counts.major}, Minor ${counts.minor}) |
| Reviewer ohne Ergebnis | ${failedReviewers} |
`

const assembled = await agent(
  `You are the ASSEMBLER (mechanical). Write the file ${OUT} with exactly this content, in this order, nothing else:\n1. The header block below verbatim.\n2. A section "## Top 10 repo-weit" — pick the 10 highest-impact items from the per-area top lists below (prefer confirmed blockers, then majors, then defect classes), one bullet each with file reference.\n3. A section "## Defektklassen und vorgeschlagene Gates" — merge the per-area defect classes, one bullet each, with a concrete gate proposal (lint rule, grep in CI, test).\n4. The per-area Markdown sections verbatim, in the given order.\n5. A section "## Verworfene Befunde" listing the dropped claims as one-liners (file: claim), so they are not re-raised.\nThen run: cd ${REPO}/frontend && npx --no-install prettier --write ${OUT} ; and return the final line count of the file.\n\nHEADER:\n${header}\n\nPER-AREA TOP LISTS:\n${JSON.stringify(sections.filter(Boolean).map((s, i) => ({ area: Object.keys(AREAS)[i], top: s.top })), null, 0)}\n\nPER-AREA DEFECT CLASSES:\n${JSON.stringify(sections.filter(Boolean).map((s, i) => ({ area: Object.keys(AREAS)[i], classes: s.defectClasses })), null, 0)}\n\nPER-AREA SECTIONS (verbatim, in order):\n${sections.map((s, i) => (s ? s.markdown : `## ${Object.keys(AREAS)[i]}\n\n(Konsolidierung fehlgeschlagen)`)).join('\n\n')}\n\nDROPPED:\n${JSON.stringify(dropped.map((d) => `${d.file}:${d.lines}: ${d.what}`), null, 0)}`,
  { label: 'assemble', phase: 'Consolidate', model: 'opus' },
)

return { filesReviewed, raw: raw.length, deduped: deduped.length, verified: verified.filter(Boolean).length, dropped: dropped.length, remaining: finalFindings.length, counts, failedReviewers, capped, outFile: OUT, assembled }