#!/usr/bin/env python3
"""crm.py — checks and reports on the private club register in .crm/ (never tracked).

Usage:
  python3 scripts/crm/crm.py validate

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


class RegisterError(Exception):
    """The register violates the schema; the message lists every violation."""


def default_crm_dir() -> Path:
    env = os.environ.get("CRM_DIR")
    if env:
        return Path(env)
    try:
        common = subprocess.run(
            ["git", "rev-parse", "--path-format=absolute", "--git-common-dir"],
            check=True, capture_output=True, text=True,
        ).stdout.strip()
    except (OSError, subprocess.CalledProcessError):
        return Path(".crm")
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
    for status, count in lage.get("deleted", {}).items():
        if status not in ORG_STATUSES or not isinstance(count, int) or count < 0:
            errors.append(f"[lage].deleted.{status}: must be an org status with a count >= 0")
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
        seen: set[tuple[str, str]] = set()
        for j, mention in enumerate(mentions):
            mwhere = f"{where} mention[{j}]"
            _unknown(mwhere, mention, MENTION_KEYS, errors)
            if mention.get("org") not in org_ids:
                errors.append(f"{mwhere}: unknown org '{mention.get('org')}'")
            source = str(mention.get("source", ""))
            if not SOURCE_RE.match(source):
                errors.append(f"{mwhere}: source must be gmail:<16 hex> or gh:<number>")
            key = (str(mention.get("org")), source)
            if key in seen:
                errors.append(f"{mwhere}: duplicate source for this org")
            seen.add(key)
            if not _is_date(mention.get("on")):
                errors.append(f"{mwhere}: on must be a date")
            if not isinstance(mention.get("answered"), bool):
                errors.append(f"{mwhere}: answered must be a boolean")
    return errors


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
    parser.add_argument("command", choices=["validate"])
    args = parser.parse_args(argv)
    crm_dir = default_crm_dir()
    if not (crm_dir / "register.toml").is_file():
        print(f"no register at {crm_dir}", file=sys.stderr)
        return 1
    return {"validate": cmd_validate}[args.command](crm_dir)


if __name__ == "__main__":
    sys.exit(main())
