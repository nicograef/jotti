export const meta = {
  name: 'plan-audit-fixes',
  description: 'Phase C: Opus planner writes plan-jotti-audit-fixes.md from the findings document per create-plan; one Fable sweep hands hotspots to three Opus critics; the planner incorporates the critique; an Opus recheck confirms; open questions are returned to the lead',
  phases: [
    { title: 'Plan', detail: 'Opus planner, create-plan skill', model: 'opus' },
    { title: 'Sweep', detail: 'one Fable sweep of plan vs findings → hand-over', model: 'fable' },
    { title: 'Critique', detail: 'three Opus lenses', model: 'opus' },
    { title: 'Revise', detail: 'Opus planner incorporates the critique', model: 'opus' },
    { title: 'Recheck', detail: 'Opus confirms the incorporation', model: 'opus' },
  ],
}

// args: { repo, findings, planFile, handbook (skills dir), praxisPlan, decisions (lead decisions file), date }
const A = args || {}
const REPO = A.repo || '/home/user/jotti'
const FINDINGS = A.findings || `${REPO}/docs/plans/findings-jotti-audit.md`
const PLAN = A.planFile || `${REPO}/docs/plans/plan-jotti-audit-fixes.md`
const SKILLS = A.handbook || '/home/user/handbook/.claude/skills'
const PRAXIS = A.praxisPlan || `${REPO}/docs/plans/plan-praxis-feedback.md`
const DECISIONS = A.decisions || ''
const DATE = A.date || 'unbekannt'

const COMMON = `Context: jotti (${REPO}) is a German mobile POS for club festivals (Go backend, React/TypeScript frontend, Astro website, Windows starter/relay, PostgreSQL, fiskaly cloud TSE). Read ${REPO}/AGENTS.md first (rules 1-18, freeze discipline, quality metrics, product conservatism, rule 18 current-state-only docs, no AI attribution). Plan prose is German; sentences ≤ 20 words; tables → lists → paragraphs. Never modify any file except the plan file ${PLAN}. Do not run make, docker or tests.`

const PLANNER_BRIEF = `You are the PLANNER (Opus). Produce ${PLAN} strictly per the create-plan skill: read ${SKILLS}/create-plan/SKILL.md (workflow, slice rules, template, self-review), ${SKILLS}/clarify/question-rules.md (ask gate), ${SKILLS}/quality.md (self-review checklist), ${SKILLS}/output-style.md (plan text follows it) and ${SKILLS}/implement-plan/SKILL.md (how the plan will be executed: Depends on lines, criteria that name their own change, one commit per criterion).
Input: the findings document ${FINDINGS} (read it completely: Zahlen, Top 10, Defektklassen, per-area findings with status, verworfene Befunde). Also read ${PRAXIS} (already implemented; avoid overlap and do not re-plan its open owner items)${DECISIONS ? ` and the lead decisions in ${DECISIONS} (binding; its Phase-B list names known drift to include)` : ''}.
Orchestration requirements (from ${REPO}/docs/plans/plan-orchestrierung.md Phase C, binding):
- Phases are vertical slices per defect class or area. Order: gates first (lint rules, CI greps, tests that would have caught the class), then blockers, then majors, then minors. Large refactors (effort L / category large-refactor) go into their own decision phase that produces an ADR or a documented no-go, not code.
- Every phase names: **Depends on**, **Modell** (Opus or Sonnet per the CLAUDE.md routing: Sonnet for mechanical, clearly specified work; Opus for implementation, subtle correctness, architecture), **Review-Tier** (Fable review per phase is the standard for this run), **Gate** (the exact command(s) that prove the phase: make check, make verify, make website-check, make test-e2e, a specific go test/vitest invocation, or a CI grep), and testable acceptance criteria that each name their own change (implement-plan commits one criterion at a time).
- Every confirmed blocker and major finding is covered by exactly one phase criterion, referenced as "Befund: <Datei>:<Zeilen>" so coverage is checkable. Minor findings are grouped per file/area into criteria; a minor may be dropped only with a one-line reason under "Nicht übernommene Befunde" (also list every dropped/unverified finding you deliberately exclude, with reason).
- Cleanup never changes behaviour; freeze discipline holds (schema only via new additive migration; event JSON frozen; sqlc/dbgen never hand-edited); rule 18 for all docs; no new dependencies without an explicit Open question; product conservatism (no feature creep disguised as a fix).
- Known pre-existing drift the lead already collected (include as criteria in the matching phase if the findings document lists them; otherwise add them to the matching docs phase): README "sechs" vs "drei" Fehlversuche; handbuch.md names DruckerConfigPage instead of DruckstationConfigPage; produktbeschreibung.md "null Kosten" vs README/ADR 09 (website-published: gate make website-check); docs/adrs/README.md and docs/adrs/04_warn-bestaetigung.md not prettier-clean; e2e files not prettier-clean; scripts/setup-dev-tools.sh does not rebuild golangci-lint on a toolchain-only change; TERMS.md "setzt um" wording (owner decision; Open question, not a criterion); PREVIOUS_VERSION v0.17.1 vs v0.17.3 belongs to the release phase, not this plan.
- Ask gate: run it for every unknown. Unknowns the repo, AGENTS.md or the findings settle become Resolved decisions with reasoning. Only genuine ≥2-survivor judgment calls go under "## Open questions / Risks", each with enumerated options, consequences and a recommended answer; the lead decides them. Do NOT call AskUserQuestion; write the questions into the plan.
- Precise references: file path plus symbol name (no line numbers in the plan; the findings' line ranges may be quoted inside "Befund:" references since the fix plan is consumed right away).
- Header: "# Plan: Audit-Fixes jotti" with "> Source PRD: docs/plans/findings-jotti-audit.md (Stand ${DATE})"; sections per template: Goal, Architectural decisions, Inventory, Resolved decisions, Open questions / Risks, then phases. Add a section "## Abdeckung" with a table mapping every confirmed blocker/major (Datei:Zeilen → Phase/Kriterium) and a section "## Nicht übernommene Befunde".
- Self-review per create-plan step 7 (placeholder scan, cross-phase consistency) and ${SKILLS}/quality.md before returning. Run: cd ${REPO}/frontend && npx --no-install prettier --write ${PLAN}.
Verify every claim about the code by reading the file (rule 11) before turning a finding into a criterion; if a finding no longer holds on the current tree, list it under "Nicht übernommene Befunde" with the proof.`

const CRITIC_LENSES = {
  completeness: `Lens: COMPLETENESS AGAINST THE FINDINGS. Build the full list of confirmed blocker/major findings and all defect classes from ${FINDINGS}; check each has exactly one covering criterion in ${PLAN} (the Abdeckung table must agree with the phases); check every minor finding is either covered or listed under "Nicht übernommene Befunde" with a reason; check the gates phase actually contains a gate for every defect class in the findings' "Defektklassen und vorgeschlagene Gates"; check that dropped findings ("Verworfene Befunde") are not re-planned. Report each gap with the finding reference.`,
  safety: `Lens: BEHAVIOUR PRESERVATION, FREEZE DISCIPLINE, RULE 18, PRODUCT CONSERVATISM. For every criterion: does it change user-visible behaviour while being labelled cleanup? Does anything touch persisted data (schema without a new additive migration, event JSON in place, kassenjournal), sqlc/dbgen by hand, or add a dependency without an Open question? Does any docs criterion introduce dated change entries, "früher/bisher", deprecation notes, or leave a superseded statement beside its replacement? Does any phase widen scope into features (AGENTS.md product conservatism)? Read the referenced files in ${REPO} to check that the proposed minimal change is correct and actually minimal; cite file:symbol.`,
  plan_quality: `Lens: PLAN QUALITY PER create-plan AND implement-plan. Read ${SKILLS}/create-plan/SKILL.md, ${SKILLS}/implement-plan/SKILL.md and ${SKILLS}/output-style.md. Check: template sections present and in order; every phase has Depends on, Modell, Review-Tier, Gate, and criteria that each name one concrete change (implement-plan commits per criterion) and are testable; no placeholders (TBD/TODO/"add validation" without how/"similar to Phase N"); cross-phase name consistency; Depends on lines allow the intended concurrency (gates phase first; independent areas parallel; shared files such as README.md, handbuch.md, language.md, docs/adrs/README.md, Makefile, ci.yml named as choke points); Modell routing follows CLAUDE.md (Sonnet only for mechanical, clearly specified work); Open questions pass the ask gate (each has options, consequences, recommendation) and nothing already settled by the repo is asked; German prose ≤ 20 words per sentence; precise references (path + symbol, no bare line numbers outside "Befund:" references). Report concrete defects with the plan line.`,
}

const CRITIQUE = {
  type: 'object',
  properties: {
    findings: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          severity: { type: 'string', enum: ['blocker', 'major', 'minor'] },
          where: { type: 'string', description: 'plan section / phase / criterion, or "Abdeckung"' },
          what: { type: 'string' },
          why: { type: 'string' },
          fix: { type: 'string', description: 'concrete change to the plan text' },
          proof: { type: 'string', description: 'quoted plan/findings/code lines' },
        },
        required: ['severity', 'where', 'what', 'why', 'fix', 'proof'],
      },
    },
    summary: { type: 'string' },
  },
  required: ['findings', 'summary'],
}

phase('Plan')
log('Opus-Planer schreibt plan-jotti-audit-fixes.md')
const draft = await agent(`${COMMON}\n\n${PLANNER_BRIEF}\n\nReturn: the plan's phase list (number, title, Modell, Gate, criteria count), the Abdeckung counts (blocker/major covered vs. total in the findings), the list of Open questions verbatim, and the final line count of ${PLAN}.`, {
  label: 'plan:draft', phase: 'Plan', model: 'opus', effort: 'xhigh',
})

phase('Sweep')
const SWEEP = { type: 'object', properties: { hotspots: { type: 'array', items: { type: 'string' } }, suspectedGaps: { type: 'array', items: { type: 'string' } }, questionsForCritics: { type: 'array', items: { type: 'string' } }, summary: { type: 'string' } }, required: ['hotspots', 'suspectedGaps', 'questionsForCritics', 'summary'] }
const sweep = await agent(`${COMMON}\n\nYou are the SWEEPER (fast, shallow; cap yourself at about 30 tool calls). Read ${PLAN} once and skim ${FINDINGS} (Zahlen, Top 10, Defektklassen, per-area headings, Abdeckung of the plan). Do not produce findings; produce a hand-over brief for three Opus critics: hotspots (plan phases/criteria that look risky, non-minimal, behaviour-changing or vague), suspectedGaps (findings or defect classes that appear uncovered or dropped without reason), questionsForCritics (what each critic must verify in the repo). One line each.`, {
  label: 'sweep', phase: 'Sweep', schema: SWEEP, model: 'fable', effort: 'medium',
})
const sweepText = sweep ? `HAND-OVER BRIEF from the Fable sweep (prioritise with it, then still do the full check):\n${JSON.stringify(sweep, null, 0)}` : 'No sweep brief (sweep failed); critique from scratch.'

phase('Critique')
const critiques = await parallel(
  Object.entries(CRITIC_LENSES).map(([lens, text]) => () =>
    agent(`${COMMON}\n\nYou are a CRITIC (read-only, highest rigor). Plan under review: ${PLAN}. Findings document: ${FINDINGS}. Praxis plan already implemented: ${PRAXIS}.\n${text}\n\n${sweepText}\n\nSeverity: blocker = a confirmed blocker/major finding or defect class is uncovered, or a criterion would break freeze discipline/behaviour; major = wrong or non-minimal fix, missing gate, unexecutable criterion, missing Depends on/Modell/Gate; minor = wording, ordering, style. "No issues" is a valid result. Never manufacture findings.`, {
      label: `critique:${lens}`, phase: 'Critique', schema: CRITIQUE, model: 'opus', effort: 'xhigh',
    }),
  ),
)
const all = critiques.filter(Boolean).flatMap((c, i) => c.findings.map((f) => ({ ...f, lens: Object.keys(CRITIC_LENSES)[i] })))
const counts = { blocker: 0, major: 0, minor: 0 }
all.forEach((f) => { counts[f.severity]++ })
log(`Kritik: ${all.length} Befunde (Blocker ${counts.blocker}, Major ${counts.major}, Minor ${counts.minor}), ${critiques.filter((c) => !c).length} Kritiker ohne Ergebnis`)

phase('Revise')
let revision = null
if (all.length > 0) {
  revision = await agent(`${COMMON}\n\nYou are the PLANNER (Opus) revising ${PLAN}. Three Fable critics reviewed your draft; incorporate every blocker and major finding and every minor you agree with (state which minors you reject and why — only for a verifiable reason). Keep the create-plan template and the orchestration requirements (gates first, Depends on/Modell/Review-Tier/Gate per phase, Abdeckung table, Nicht übernommene Befunde, Open questions per ask gate). Verify each critic claim against the files before acting (rule 11). Append under "## Resolved decisions" a bullet list "Fable-Kritik (${DATE})" naming what was incorporated in one line each (this is the required documentation of the critique). Run prettier as before. Return: per critique item, "eingearbeitet" or "abgelehnt: <reason>", plus the updated Open questions verbatim and the final line count.\n\nCRITIQUE:\n${JSON.stringify(all, null, 0)}`, {
    label: 'plan:revise', phase: 'Revise', model: 'opus', effort: 'xhigh',
  })
} else {
  log('Keine Kritikpunkte — keine Überarbeitung nötig')
}

phase('Recheck')
const RECHECK = { type: 'object', properties: { ok: { type: 'boolean' }, remaining: { type: 'array', items: { type: 'string' } }, openQuestions: { type: 'array', items: { type: 'string' } }, coverage: { type: 'string' }, summary: { type: 'string' } }, required: ['ok', 'remaining', 'openQuestions', 'coverage', 'summary'] }
const recheck = await agent(`${COMMON}\n\nYou are the RECHECKER (read-only). Plan: ${PLAN}; findings: ${FINDINGS}. Confirm: (1) every blocker/major critique item below is incorporated in the plan text (quote the line) or rejected with a verifiable reason; (2) the Abdeckung table matches the phases and covers every confirmed blocker/major in the findings (count both sides); (3) every phase carries Depends on, Modell, Review-Tier, Gate and ≥1 testable criterion; (4) no placeholder markers; (5) the "Fable-Kritik" bullet list exists under Resolved decisions. Return ok=true only if all five hold; list what remains; copy the Open questions verbatim into openQuestions; put the coverage counts in coverage.\n\nCRITIQUE ITEMS:\n${JSON.stringify(all.filter((f) => f.severity !== 'minor'), null, 0)}\n\nPLANNER RESPONSE:\n${revision || '(no revision needed: zero critique items)'}`, {
  label: 'recheck', phase: 'Recheck', schema: RECHECK, model: 'opus', effort: 'high',
})

return { draft, sweep, critique: { counts, items: all }, revision, recheck }
