export const meta = {
  name: 'review-phase',
  description:
    'Review of one implement-plan phase diff: one Fable sweep hands hotspots to Opus probes (criterion audit + correctness/conventions + cleanup lens), then Opus skeptics refute blocker/major findings',
  phases: [
    {
      title: 'Sweep',
      detail: 'one Fable sweep of the diff → hand-over brief',
      model: 'fable',
    },
    {
      title: 'Probe',
      detail: 'read-only Opus probes over the phase diff',
      model: 'opus',
    },
    {
      title: 'Refute',
      detail: 'one skeptic per blocker/major finding',
      model: 'opus',
    },
  ],
}

// args: { phase, worktree, branch, base, planPath, repo, handbook, gateSummary, extraLens, slug }
// slug = the plan slug used in the commit trailer "Plan: <slug> phase <N> criterion <M>" (required)
const A = args || {}
for (const k of ['phase', 'worktree', 'branch', 'base', 'planPath', 'slug'])
  if (!A[k]) throw new Error(`review-phase: Argument ${k} fehlt`)
const SLUG = A.slug
const REPO = A.repo || '/home/user/jotti'
const HANDBOOK = A.handbook || '/home/user/handbook/.claude/skills'
const DIFF = `git -C ${A.worktree} diff ${A.base}...${A.branch}`
const LOG = `git -C ${A.worktree} log --oneline ${A.base}..${A.branch}`

const COMMON = `
You are a REVIEWER (read-only) at the highest rigor for jotti, a German mPOS for club festivals (Go backend with stdlib net/http, pgx, sqlc, zog; React/TypeScript frontend; Astro website; Windows starter/relay in Go; PostgreSQL; fiskaly cloud TSE).
Repository checkout to review: ${A.worktree} (branch ${A.branch}); base commit ${A.base}. The diff under review is exactly: ${DIFF} (commit list: ${LOG}).
Read ${A.worktree}/AGENTS.md first (rules 1-18, freeze discipline, ubiquitous language, rule 18 current-state-only, quality metrics). Consult ${A.worktree}/docs/language.md for naming per layer and ${A.worktree}/docs/handbuch.md for architecture when judging boundaries.
The plan is ${A.planPath} inside that worktree; review the section "## Phase ${A.phase}" (Context, What to build, Acceptance criteria). Read the file with sed/cat, never paste it back.
Gate already run by the worker (do not repeat the full gate): ${A.gateSummary || 'see the phase report'}. You may run cheap targeted checks (go vet on a package, go test on one package, grep, node -e, pnpm exec tsc --noEmit -p <dir>, git show) but never the full make check/verify, never docker, never pnpm install.
STRICTLY READ-ONLY: do not edit, create, delete, commit, stash, checkout or reset anything. If you must measure by editing, restore with git restore -- <file> and prove "git -C ${A.worktree} status --porcelain" is empty before returning. Never touch ${REPO} (the main checkout) or the plan file.
Every finding must carry proof: quoted lines with file:line or command output. A worry you cannot demonstrate is not a finding. Prefer fewer, true findings; "no findings" is a valid result. Never propose functionality changes disguised as cleanup; never widen scope beyond the phase. Sentences <= 20 words.`

const FINDINGS = {
  type: 'object',
  properties: {
    findings: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          severity: { type: 'string', enum: ['blocker', 'major', 'minor'] },
          category: {
            type: 'string',
            description:
              'correctness | criterion | convention | rule18 | naming | security | readability | principle | code-smell | architecture | docs-accuracy | test-quality',
          },
          file: { type: 'string' },
          lines: { type: 'string' },
          what: { type: 'string' },
          why: { type: 'string' },
          suggestion: {
            type: 'string',
            description: 'concrete minimal change',
          },
          proof: { type: 'string' },
        },
        required: [
          'severity',
          'category',
          'file',
          'lines',
          'what',
          'why',
          'suggestion',
          'proof',
        ],
      },
    },
    criteria: {
      type: 'array',
      description:
        'one entry per acceptance criterion of the phase, in plan order',
      items: {
        type: 'object',
        properties: {
          index: { type: 'integer' },
          text: {
            type: 'string',
            description: 'first 80 chars of the criterion',
          },
          status: {
            type: 'string',
            enum: ['verified', 'failed', 'not-verifiable-here'],
          },
          evidence: {
            type: 'string',
            description:
              'command + output excerpt, or why it cannot be verified in this session',
          },
        },
        required: ['index', 'text', 'status', 'evidence'],
      },
    },
    defectClasses: { type: 'array', items: { type: 'string' } },
    summary: { type: 'string' },
  },
  required: ['findings', 'criteria', 'defectClasses', 'summary'],
}

const LENSES = [
  {
    key: 'criteria-correctness',
    prompt: `Lens: CRITERION AUDIT AND CORRECTNESS. For every acceptance criterion of Phase ${A.phase}, verify it against the diff and with cheap commands; mark it verified / failed / not-verifiable-here with evidence. Then hunt real defects in the diff: logic errors, nil/undefined handling, error swallowing, transaction boundaries, money not in cents, validation gaps at HTTP edges (zog/Zod on both sides), authz, SQL via sqlc params only, event-sourcing invariants (append-only journal, projections in same transaction, frozen event contracts), migration rules (database/migrations/README.md: additive, next free number, BEGIN/COMMIT, 01_initial untouched), generated code (sqlc/dbgen never hand-edited, regenerated when queries changed). For doc/website phases: factual errors versus the code, dead links, broken rendering (published-docs.ts pipeline), contradictions between files.`,
  },
  {
    key: 'conventions-cleanup',
    prompt: `Lens: CONVENTIONS AND CLEANUP. Read ${HANDBOOK}/cleanup/readability.md, ${HANDBOOK}/cleanup/readability-de.md (German prose), ${HANDBOOK}/cleanup/principles.md, ${HANDBOOK}/cleanup/code-smells.md, ${HANDBOOK}/cleanup/architecture.md and apply them to every file the diff touches (read the full files for context, not only hunks). Check AGENTS.md rules: German ubiquitous language in domain code and user-visible strings, English infrastructure code, no json tags in domain structs, POST-only endpoints, no direct fetch() in the frontend, backend as single source of truth for filtering, rule 18 (no "früher/bisher/vormals", no dated change entries, no deprecation notes outside CHANGELOG/ADRs; every statement the change makes false must be rewritten in the same change — search the repo for such statements), version consistency across go.work/go.mod/Dockerfiles/CI/AGENTS.md/docs, comments that describe a state the code no longer has, duplicated content, TODO/FIXME left behind, test naming and coverage for new public behaviour, commit messages (Conventional Commits, English, trailer "Plan: ${SLUG} phase ${A.phase} criterion <M>", no AI attribution trailers).`,
  },
]
if (A.extraLens) LENSES.push({ key: 'extra', prompt: A.extraLens })

phase('Sweep')
const SWEEP = {
  type: 'object',
  properties: {
    hotspots: { type: 'array', items: { type: 'string' } },
    criteriaAtRisk: { type: 'array', items: { type: 'string' } },
    questionsForProbes: { type: 'array', items: { type: 'string' } },
    summary: { type: 'string' },
  },
  required: ['hotspots', 'criteriaAtRisk', 'questionsForProbes', 'summary'],
}
const sweep = await agent(
  `${COMMON}\n\nYou are the SWEEPER (fast, shallow; cap yourself at about 25 tool calls). Read the diff once (${DIFF}) and the phase's acceptance criteria. Do not produce findings; produce a hand-over brief for the Opus probes: hotspots (file:line ranges that look wrong, non-minimal, behaviour-changing, rule-18-relevant or untested), criteriaAtRisk (criteria whose evidence in the diff looks thin), questionsForProbes (what to verify with which command). One line each.`,
  {
    label: 'sweep',
    phase: 'Sweep',
    schema: SWEEP,
    model: 'fable',
    effort: 'medium',
  },
)
const sweepText = sweep
  ? `HAND-OVER BRIEF from the Fable sweep (prioritise with it, then still do the full lens):\n${JSON.stringify(sweep, null, 0)}`
  : 'No sweep brief (sweep failed); probe from scratch.'

phase('Probe')
log(
  `Phase ${A.phase}: Fable-Sweep, dann ${LENSES.length} Opus-Proben über ${A.branch}`,
)
const probes = await parallel(
  LENSES.map(
    (l) => () =>
      agent(
        `${COMMON}\n\n${l.prompt}\n\n${sweepText}\n\nReturn findings for this lens only; the criteria list is required from every lens (re-verify independently).`,
        {
          label: `probe:${l.key}`,
          phase: 'Probe',
          schema: FINDINGS,
          model: 'opus',
          effort: 'xhigh',
        },
      ),
  ),
)
const ok = probes.filter(Boolean)
const raw = []
probes.forEach((p, i) => {
  if (p) p.findings.forEach((f) => raw.push({ ...f, lens: LENSES[i].key }))
})
const seen = new Map()
for (const f of raw) {
  const key = `${f.file}:${String(f.lines).split(/[-–,]/)[0].trim()}:${f.what.toLowerCase().slice(0, 40)}`
  if (!seen.has(key)) seen.set(key, f)
}
const deduped = [...seen.values()]
log(
  `Proben: ${ok.length}/${LENSES.length} geantwortet, ${raw.length} Rohbefunde, ${deduped.length} nach Dedupe`,
)

phase('Refute')
const VERDICT = {
  type: 'object',
  properties: { refuted: { type: 'boolean' }, reason: { type: 'string' } },
  required: ['refuted', 'reason'],
}
const toRefute = deduped.filter((f) => f.severity !== 'minor')
const verdicts = await parallel(
  toRefute.map(
    (f, i) => () =>
      agent(
        `${COMMON}\n\nYou are a SKEPTIC. A reviewer (lens ${f.lens}) claims:\nFILE: ${f.file}:${f.lines}\nSEVERITY: ${f.severity} (${f.category})\nWHAT: ${f.what}\nWHY: ${f.why}\nSUGGESTION: ${f.suggestion}\nPROOF: ${f.proof}\n\nTry to REFUTE it: re-read the exact lines and surrounding code/docs; is the claim literally true at that location, is it reachable or a real reader confusion, is the suggestion minimal, behaviour-preserving (for cleanup) or correct (for defects), and within the phase's scope and AGENTS.md rules? Default to refuted=true when uncertain or when it is cosmetic taste rather than a rule or defect.`,
        {
          label: `refute:${i}`,
          phase: 'Refute',
          schema: VERDICT,
          model: 'opus',
          effort: 'high',
        },
      ),
  ),
)
const confirmed = []
const dropped = []
toRefute.forEach((f, i) => {
  const v = verdicts[i]
  if (v && v.refuted) dropped.push({ ...f, reason: v.reason })
  else
    confirmed.push({
      ...f,
      skeptic: v
        ? v.reason
        : 'kein Urteil (Agent ohne Ergebnis) — bleibt bestehen',
    })
})
const minors = deduped.filter((f) => f.severity === 'minor')
log(
  `Refutation: ${confirmed.length} bestätigt, ${dropped.length} verworfen, ${minors.length} Minor ungeprüft`,
)

return {
  phase: A.phase,
  sweep,
  probesAnswered: ok.length,
  probesTotal: LENSES.length,
  criteria: ok.map((p) => p.criteria),
  confirmed,
  minors,
  dropped,
  defectClasses: [...new Set(ok.flatMap((p) => p.defectClasses))],
  summaries: ok.map((p) => p.summary),
}
