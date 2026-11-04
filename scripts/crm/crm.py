#!/usr/bin/env python3
"""crm.py — checks and reports on the private club register in .crm/ (never tracked).

Usage:
  python3 scripts/crm/crm.py validate     # schema check, prints the counts
  python3 scripts/crm/crm.py pii-terms    # one search term per line for check-private-data.sh
  python3 scripts/crm/crm.py render [--check]  # write docs/feldfeedback.md, or compare with it
  python3 scripts/crm/crm.py open         # mentions still waiting for a reply, oldest first
  python3 scripts/crm/crm.py retention    # orgs whose correspondence is due for deletion

The register is edited by hand or by Claude; this script only reads it. CRM_DIR
overrides the register location, which defaults to the main checkout's .crm/ so
that worktrees under ../jotti-wt/ find it too.
"""

from __future__ import annotations

import sys

if sys.version_info < (3, 11):
    sys.exit("crm.py needs Python >= 3.11 (tomllib); found " + sys.version.split()[0])

import argparse
import datetime as dt
import os
import re
import subprocess
import tempfile
import tomllib
from pathlib import Path

ORG_KINDS = {"verein", "person", "dienstleister"}
ORG_STATUSES = {"anfrage", "vereinbart", "zurueckgetreten", "bestritten", "erledigt"}
TOPIC_KINDS = {"wunsch", "fehler", "frage", "einsatzbericht", "verwaltung"}
TOPIC_STATUSES = {"offen", "umgesetzt", "abgelehnt", "beantwortet"}
SOURCE_RE = re.compile(r"^(gmail:[0-9a-f]{16}|gh:[0-9]+)$")

# Allowed keys per table; anything else (phone, address, …) is rejected so the
# register cannot grow data categories the privacy policy does not name.
REGISTER_KEYS = {"sync", "terms", "org"}
SYNC_KEYS = {"gmail_through", "gh_through"}
TERMS_KEYS = {"versions"}
ORG_KEYS = {
    "id", "kind", "name", "legal_form", "seat", "country", "status", "agreed_on",
    "terms_version", "agreement_source", "public_name_ok", "gh", "notes", "contact",
}
CONTACT_KEYS = {"name", "role", "emails"}
TOPICS_KEYS = {"lage", "topic"}
LAGE_KEYS = {"notes", "deleted"}
TOPIC_KEYS = {"id", "title", "kind", "status", "ref", "mention"}
MENTION_KEYS = {"org", "source", "on", "answered"}

ROOT = Path(__file__).resolve().parents[2]


class RegisterError(Exception):
    """The register violates the schema; the message lists every violation."""


def default_crm_dir() -> Path:
    env = os.environ.get("CRM_DIR")
    if env:
        return Path(env)
    try:
        common = subprocess.run(
            ["git", "rev-parse", "--path-format=absolute", "--git-common-dir"],
            cwd=ROOT, check=True, capture_output=True, text=True,
        ).stdout.strip()
    except (OSError, subprocess.CalledProcessError):
        return ROOT / ".crm"
    return Path(common).parent / ".crm"


def load(crm_dir: Path) -> tuple[dict, dict]:
    with open(crm_dir / "register.toml", "rb") as f:
        register = tomllib.load(f)
    with open(crm_dir / "topics.toml", "rb") as f:
        topics = tomllib.load(f)
    return register, topics


def _unknown(where: str, table: dict, allowed: set[str], errors: list[str]) -> None:
    for key in sorted(set(table) - allowed):
        errors.append(f"{where}: unknown field '{key}'")


def _is_date(value: object) -> bool:
    return isinstance(value, dt.date) and not isinstance(value, dt.datetime)


def validate(register: dict, topics: dict) -> list[str]:
    """Returns every schema violation; an empty list means the register is valid."""
    errors: list[str] = []
    _unknown("register", register, REGISTER_KEYS, errors)
    _unknown("[sync]", register.get("sync", {}), SYNC_KEYS, errors)
    for key in SYNC_KEYS:
        if key in register.get("sync", {}) and not _is_date(register["sync"][key]):
            errors.append(f"[sync].{key}: must be a date")
    _unknown("[terms]", register.get("terms", {}), TERMS_KEYS, errors)
    versions = register.get("terms", {}).get("versions", [])
    if not versions or not all(_is_date(v) for v in versions):
        errors.append("[terms].versions: must be a non-empty list of dates")

    org_ids: set[str] = set()
    for i, org in enumerate(register.get("org", [])):
        where = f"org[{i}] {org.get('id', '?')}"
        _unknown(where, org, ORG_KEYS, errors)
        oid = org.get("id")
        if not isinstance(oid, str) or not re.fullmatch(r"[a-z0-9-]+", oid):
            errors.append(f"{where}: id must be a lowercase slug")
        elif oid in org_ids:
            errors.append(f"{where}: duplicate id")
        else:
            org_ids.add(oid)
        if not isinstance(org.get("name"), str) or not org["name"].strip():
            errors.append(f"{where}: name is required")
        if org.get("kind") not in ORG_KINDS:
            errors.append(f"{where}: kind must be one of {sorted(ORG_KINDS)}")
        status = org.get("status")
        if status not in ORG_STATUSES:
            errors.append(f"{where}: status must be one of {sorted(ORG_STATUSES)}")
        if status == "vereinbart":
            if not _is_date(org.get("agreed_on")):
                errors.append(f"{where}: vereinbart needs agreed_on")
            if org.get("terms_version") not in versions:
                errors.append(f"{where}: terms_version must be one of [terms].versions")
            if not SOURCE_RE.match(str(org.get("agreement_source", ""))):
                errors.append(f"{where}: vereinbart needs agreement_source")
        elif "terms_version" in org and org["terms_version"] not in versions:
            errors.append(f"{where}: terms_version must be one of [terms].versions")
        country = org.get("country")
        if country is not None and not re.fullmatch(r"[A-Z]{2}", str(country)):
            errors.append(f"{where}: country must be an ISO 3166 alpha-2 code")
        if not isinstance(org.get("public_name_ok", False), bool):
            errors.append(f"{where}: public_name_ok must be a boolean")
        for j, contact in enumerate(org.get("contact", [])):
            cwhere = f"{where} contact[{j}]"
            _unknown(cwhere, contact, CONTACT_KEYS, errors)
            if not isinstance(contact.get("name"), str) or not contact["name"].strip():
                errors.append(f"{cwhere}: name is required")
            emails = contact.get("emails")
            if not isinstance(emails, list) or not emails or not all(
                isinstance(e, str) and "@" in e for e in emails
            ):
                errors.append(f"{cwhere}: emails must list at least one address")

    _unknown("topics", topics, TOPICS_KEYS, errors)
    lage = topics.get("lage", {})
    _unknown("[lage]", lage, LAGE_KEYS, errors)
    deleted = lage.get("deleted", {})
    if not isinstance(deleted, dict):
        errors.append("[lage].deleted: must be a table of org status to count")
        deleted = {}
    for status, count in deleted.items():
        if status not in ORG_STATUSES or not isinstance(count, int) or count < 0:
            errors.append(f"[lage].deleted.{status}: must be an org status with a count >= 0")
    waiting: set[str] = set()
    topic_ids: set[str] = set()
    for i, topic in enumerate(topics.get("topic", [])):
        where = f"topic[{i}] {topic.get('id', '?')}"
        _unknown(where, topic, TOPIC_KEYS, errors)
        tid = topic.get("id")
        if not isinstance(tid, str) or not re.fullmatch(r"[a-z0-9-]+", tid):
            errors.append(f"{where}: id must be a lowercase slug")
        elif tid in topic_ids:
            errors.append(f"{where}: duplicate id")
        else:
            topic_ids.add(tid)
        if not isinstance(topic.get("title"), str) or not topic["title"].strip():
            errors.append(f"{where}: title is required")
        if topic.get("kind") not in TOPIC_KINDS:
            errors.append(f"{where}: kind must be one of {sorted(TOPIC_KINDS)}")
        if topic.get("status") not in TOPIC_STATUSES:
            errors.append(f"{where}: status must be one of {sorted(TOPIC_STATUSES)}")
        if topic.get("kind") != "verwaltung" and not str(topic.get("ref", "")).strip():
            errors.append(f"{where}: ref is required")
        mentions = topic.get("mention", [])
        if not mentions:
            errors.append(f"{where}: needs at least one mention")
        seen: set[str] = set()
        for j, mention in enumerate(mentions):
            mwhere = f"{where} mention[{j}]"
            _unknown(mwhere, mention, MENTION_KEYS, errors)
            if mention.get("org") not in org_ids:
                errors.append(f"{mwhere}: unknown org '{mention.get('org')}'")
            source = str(mention.get("source", ""))
            if not SOURCE_RE.match(source):
                errors.append(f"{mwhere}: source must be gmail:<16 hex> or gh:<number>")
            if source in seen:
                errors.append(f"{mwhere}: duplicate source in this topic")
            seen.add(source)
            if not _is_date(mention.get("on")):
                errors.append(f"{mwhere}: on must be a date")
            if not isinstance(mention.get("answered"), bool):
                errors.append(f"{mwhere}: answered must be a boolean")
            elif not mention["answered"]:
                waiting.add(str(mention.get("org")))
    for i, org in enumerate(register.get("org", [])):
        if org.get("status") == "erledigt" and org.get("id") in waiting:
            errors.append(f"org[{i}] {org['id']}: erledigt while a mention still waits for a reply")
    return errors


MIN_TERM_LENGTH = 5


def pii_terms(register: dict) -> list[str]:
    """Terms that must never appear in a tracked file.

    Contact e-mails, full contact names (a single first name is no identifier and
    would fire on ordinary prose), and org names unless the org consented to be
    named. Service providers are skipped: docs may name them. GitHub logins are
    public on the issues anyway.
    """
    terms: set[str] = set()
    for org in register.get("org", []):
        if org.get("kind") == "dienstleister":
            continue
        if not org.get("public_name_ok", False):
            terms.add(org["name"].strip())
        for contact in org.get("contact", []):
            name = contact["name"].strip()
            if " " in name:
                terms.add(name)
            terms.update(e.strip() for e in contact.get("emails", []))
    return sorted(t for t in terms if len(t) >= MIN_TERM_LENGTH)


def cmd_pii_terms(crm_dir: Path) -> int:
    register, topics = load(crm_dir)
    if validate(register, topics):
        print("register invalid; run crm.py validate", file=sys.stderr)
        return 1
    print("\n".join(pii_terms(register)))
    return 0


FELDFEEDBACK = ROOT / "docs" / "feldfeedback.md"
PRETTIER = ROOT / "frontend" / "node_modules" / ".bin" / "prettier"
MONTHS = ["Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August",
          "September", "Oktober", "November", "Dezember"]
COUNTRIES = {"AT": "Österreich", "CH": "der Schweiz", "LI": "Liechtenstein", "LU": "Luxemburg"}
INTRO = """# Feldfeedback

Wünsche, Fehlerberichte und Fragen aus den Mails und GitHub-Issues der Vereine, anonymisiert und gezählt. Keine Namen, Orte, Termine oder Zitate: Die Zuordnung zu Vereinen liegt nur im gitignorierten Register `.crm/`.

`make crm` erzeugt diese Datei aus dem Register; Änderungen von Hand gehen beim nächsten Lauf verloren. „Vereine" zählt die Vereine und Einzelpersonen, die einen Punkt unabhängig voneinander genannt haben.
"""
# Section title, first column header, last column header; frage topics form the last section.
SECTIONS = [
    ("umgesetzt", "Umgesetzt", "Wunsch", "Umsetzung"),
    ("abgelehnt", "Abgelehnt", "Wunsch", "Entscheidung"),
    ("offen", "Offen", "Wunsch oder Fehler", "Stand"),
    ("frage", "Wiederkehrende Fragen", "Frage", "Antwort steht in"),
]


def _german_date(d: dt.date) -> str:
    return f"{d.day}. {MONTHS[d.month - 1]} {d.year}"


def _plural(n: int, one: str, many: str) -> str:
    return f"{n} {one if n == 1 else many}"


def _join(parts: list[str]) -> str:
    return parts[0] if len(parts) == 1 else ", ".join(parts[:-1]) + " und " + parts[-1]


def _table(header: list[str], rows: list[list[str]], right: set[int]) -> list[str]:
    """A Markdown table aligned the way Prettier aligns it, so `make crm` stays stable."""
    widths = [max(3, len(header[i]), *(len(r[i]) for r in rows)) for i in range(len(header))]

    def line(cells: list[str]) -> str:
        padded = [c.rjust(w) if i in right else c.ljust(w) for i, (c, w) in enumerate(zip(cells, widths))]
        return "| " + " | ".join(padded) + " |"

    rule = ["-" * (w - 1) + ":" if i in right else "-" * w for i, w in enumerate(widths)]
    return [line(header), "| " + " | ".join(rule) + " |", *(line(r) for r in rows)]


def render(register: dict, topics: dict) -> str:
    """The full docs/feldfeedback.md; only counts, titles, statuses and refs leave the register."""
    orgs = register.get("org", [])
    deleted = topics.get("lage", {}).get("deleted", {})
    agreed = [o for o in orgs if o["status"] == "vereinbart"]
    withdrawn = sum(o["status"] == "zurueckgetreten" for o in orgs) + deleted.get("zurueckgetreten", 0)
    disowned = sum(o["status"] == "bestritten" for o in orgs) + deleted.get("bestritten", 0)
    inquiring_clubs = sum(o["status"] == "anfrage" and o["kind"] == "verein" for o in orgs)
    inquiring_people = sum(o["status"] == "anfrage" and o["kind"] == "person" for o in orgs)

    lage = [
        f"{_plural(len(agreed) + withdrawn, 'Organisation hat', 'Organisationen haben')} die "
        f"Nutzungsvereinbarung geschlossen, {withdrawn} davon zurückgezogen; "
        f"{_plural(len(agreed), 'Vereinbarung gilt', 'Vereinbarungen gelten')}."
    ]
    versions = sorted({o["terms_version"] for o in agreed})
    if versions:
        cited = [f"{_german_date(v)} ({sum(o['terms_version'] == v for o in agreed)})" for v in versions]
        lage.append(f"Die geltenden Annahmesätze nennen die TERMS-Fassungen vom {_join(cited)}.")
    asked = f"{_plural(inquiring_clubs, 'weitere Organisation', 'weitere Organisationen')} und " \
            f"{_plural(inquiring_people, 'Einzelperson', 'Einzelpersonen')} haben angefragt, ohne eine Vereinbarung zu schließen."
    if disowned:
        asked += f" {_plural(disowned, 'Absender hat seine Mail', 'Absender haben ihre Mail')} bestritten."
    lage.append(asked)
    abroad = {}
    for o in orgs:
        if o["kind"] == "verein" and o.get("country") not in (None, "DE"):
            abroad[o["country"]] = abroad.get(o["country"], 0) + 1
    if abroad:
        parts = [f"{n} aus {COUNTRIES.get(c, c)}" for c, n in sorted(abroad.items())]
        lage.append(f"Vereine aus dem Ausland: {_join(parts)}.")
    reporters = {m["org"] for tp in topics.get("topic", []) if tp["kind"] == "einsatzbericht"
                 for m in tp["mention"]}
    if reporters:
        lage.append(f"{_plural(len(reporters), 'Verein berichtet', 'Vereine berichten')} vom Einsatz am Fest.")
    lage.extend(topics.get("lage", {}).get("notes", []))

    out = [INTRO, "## Lage", "", *(f"- {s}" for s in lage), ""]
    for key, title, first, last in SECTIONS:
        rows = []
        for tp in topics.get("topic", []):
            if tp["kind"] in ("verwaltung", "einsatzbericht"):
                continue
            section = "frage" if tp["kind"] == "frage" else tp["status"]
            if section == "beantwortet":
                section = "umgesetzt"
            if section != key:
                continue
            count = len({m["org"] for m in tp["mention"]})
            rows.append((count, tp["title"], tp.get("ref", "")))
        rows.sort(key=lambda r: (-r[0], r[1].casefold()))
        out += [f"## {title}", ""]
        if rows:
            out += _table([first, "Vereine", last], [[t, str(c), ref] for c, t, ref in rows], right={1})
        else:
            out.append("Keine Einträge.")
        out.append("")
    return "\n".join(out).rstrip("\n") + "\n"


def _prettier(text: str) -> str | None:
    """text as `make crm` leaves it after Prettier; None when Prettier is missing or fails."""
    if not PRETTIER.is_file():
        print(f"Prettier missing at {PRETTIER}; run pnpm install in frontend/", file=sys.stderr)
        return None
    # Inside docs/, so Prettier resolves the same config and ignore files as for the real file.
    fd, name = tempfile.mkstemp(prefix="feldfeedback-check-", suffix=".md", dir=FELDFEEDBACK.parent)
    tmp = Path(name)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as f:
            f.write(text)
        result = subprocess.run([str(PRETTIER), "--write", str(tmp)], cwd=ROOT, capture_output=True, text=True)
        if result.returncode != 0:
            print(f"Prettier failed: {result.stderr.strip()}", file=sys.stderr)
            return None
        return tmp.read_text(encoding="utf-8")
    finally:
        tmp.unlink(missing_ok=True)


def cmd_render(crm_dir: Path, check: bool) -> int:
    register, topics = load(crm_dir)
    if validate(register, topics):
        print("register invalid; run crm.py validate", file=sys.stderr)
        return 1
    text = render(register, topics)
    rel = FELDFEEDBACK.relative_to(ROOT)
    if check:
        formatted = _prettier(text)
        if formatted is None:
            return 1
        current = FELDFEEDBACK.read_text(encoding="utf-8") if FELDFEEDBACK.exists() else ""
        if current != formatted:
            print(f"{rel} is out of date; run make crm", file=sys.stderr)
            return 1
        print(f"{rel} is up to date")
        return 0
    FELDFEEDBACK.write_text(text, encoding="utf-8")
    print(f"wrote {rel}")
    return 0


RETENTION_STATUSES = {"zurueckgetreten", "bestritten", "erledigt"}


def open_mentions(topics: dict) -> list[tuple[dt.date, str, str, str]]:
    """(date, org id, source, topic id) of every mention still waiting for a reply."""
    rows = [(m["on"], m["org"], m["source"], tp["id"])
            for tp in topics.get("topic", []) for m in tp["mention"] if not m["answered"]]
    return sorted(rows)


def retention(register: dict, topics: dict) -> list[tuple[str, str, list[str]]]:
    """(org id, last date or '-', sources) of every org whose correspondence is to be deleted.

    Mirrors the website privacy policy: withdrawn, disowned and finished contacts go
    now; agreements stay for their term and inquiries until Nico marks them erledigt.
    """
    rows = []
    for org in register.get("org", []):
        if org["status"] not in RETENTION_STATUSES:
            continue
        mentions = [m for tp in topics.get("topic", []) for m in tp["mention"] if m["org"] == org["id"]]
        sources = sorted({m["source"] for m in mentions})
        last = max((m["on"] for m in mentions), default=None)
        rows.append((org["id"], last.isoformat() if last else "-", sources))
    return rows


def cmd_open(crm_dir: Path) -> int:
    _, topics = load(crm_dir)
    rows = open_mentions(topics)
    for on, org, source, topic in rows:
        print(f"{on.isoformat()}\t{org}\t{source}\t{topic}")
    print(f"{len(rows)} mention(s) waiting for a reply", file=sys.stderr)
    return 0


def cmd_retention(crm_dir: Path) -> int:
    register, topics = load(crm_dir)
    rows = retention(register, topics)
    for org, last, sources in rows:
        print(f"{org}\t{last}\t{','.join(sources) or '-'}")
    print(f"{len(rows)} organisation(s) due for deletion", file=sys.stderr)
    return 0


def cmd_validate(crm_dir: Path) -> int:
    register, topics = load(crm_dir)
    errors = validate(register, topics)
    for e in errors:
        print(e, file=sys.stderr)
    if errors:
        print(f"{len(errors)} violation(s) in {crm_dir}", file=sys.stderr)
        return 1
    print(f"{len(register.get('org', []))} organisations, {len(topics.get('topic', []))} topics: valid")
    return 0


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("command", choices=["validate", "pii-terms", "render", "open", "retention"])
    parser.add_argument("--check", action="store_true", help="render: compare instead of writing")
    args = parser.parse_args(argv)
    crm_dir = default_crm_dir()
    if not (crm_dir / "register.toml").is_file():
        print(f"no register at {crm_dir}", file=sys.stderr)
        return 1
    if args.command == "render":
        return cmd_render(crm_dir, args.check)
    return {"validate": cmd_validate, "pii-terms": cmd_pii_terms, "open": cmd_open,
            "retention": cmd_retention}[args.command](crm_dir)


if __name__ == "__main__":
    sys.exit(main())
