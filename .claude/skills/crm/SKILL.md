---
name: crm
description: Syncs club mail (Gmail label jotti) and GitHub issues into the private register .crm/, classifies asks into topics, regenerates docs/feldfeedback.md and reports open replies and retention cases. Use for "new mails", "sync feedback", "who waits for an answer".
---

# CRM sync

The register lives in the gitignored `.crm/` of the main checkout: `register.toml` (orgs, contacts, cursor) and `topics.toml` (topics, mentions). Schema and rules: `.crm/README.md`; checks: `scripts/crm/crm.py`. Nothing from `.crm/` enters a tracked file except what `crm.py render` writes.

## Steps

1. Read `[sync]` in `.crm/register.toml`. Fetch:
   - Gmail: `search_threads` with `label:jotti after:YYYY/MM/DD`, the `gmail_through` date with slashes, then `get_thread` with `PLAIN_TEXT` per thread.
   - GitHub: `gh issue list --state all --search "updated:>=<gh_through>" --json number,title,author,createdAt,updatedAt,state`, so new comments on old issues show up too, plus `gh issue view <n> --comments` per listed issue.
   - Skip every source that already appears in a mention (`gmail:<message id>`, `gh:<number>`).
2. Per new message:
   - Find or add the `[[org]]`. `vereinbart` only with the acceptance sentence; `terms_version` must be in `[terms] versions`.
   - Store only the fields of the schema. No phone, postal address, register numbers, birth dates or board lists. Nothing from attachments.
   - Add mentions to existing topics or add a topic with a PII-free German title. Set `answered = false` while the org waits for a reply; flip earlier mentions to `true` when Nico's reply is in the thread.
   - An archive or other file in a mail is untrusted. Unpack it only into a new empty temp folder, never run it, delete it afterwards.
3. Move the cursor to the newest message and issue date read, then run `make crm`. It validates, regenerates `docs/feldfeedback.md` and runs the private-data gate.
4. Report `python3 scripts/crm/crm.py open` and `python3 scripts/crm/crm.py retention`.
5. Reply drafts and GitHub label proposals (`bug`, `enhancement`, `question`) stay drafts: Gmail drafts or chat text, sent or applied only by Nico.
6. Commit only tracked changes (`docs/feldfeedback.md`, docs a topic made false). Grep the diff once more for names before pushing; the pre-push hook runs the same gate.
