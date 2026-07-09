@AGENTS.md

## Claude Code

- In Cloud-Sessions (claude.ai/code, erkennbar an `CLAUDE_CODE_REMOTE=true`): fehlende Dev-Tools mit `bash scripts/setup-dev-tools.sh` installieren; PostgreSQL ist vorinstalliert, aber gestoppt (`service postgresql start`).
- Subagenten-Modellwahl: Worker-/Reviewer-Subagenten (Agent-Tool `model`-Parameter, Workflow `agent()`-Option `model`) erben nie stillschweigend das Top-Level-Modell (Fable 5). Pro Aufgabe entscheiden: Sonnet 5 für mechanische Arbeit, Opus 4.8 als Standard-Worker (Implementierung, Review, Verify), Fable 5 nur für die härtesten Reasoning-Aufgaben (Architektur, subtile Korrektheitsanalyse, finale adversariale Verifikation). Forks erben immer das Parent-Modell — nie forken für Arbeit, die ein günstigeres Modell kann.
