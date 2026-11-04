"""Tests for crm.py on invented fixtures; run with python3 -m unittest scripts/crm/test_crm.py."""

import contextlib
import copy
import datetime as dt
import io
import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

sys.path.insert(0, str(Path(__file__).resolve().parent))
import crm  # noqa: E402

REGISTER = {
    "sync": {"gmail_through": dt.date(2026, 1, 31), "gh_through": dt.date(2026, 1, 31)},
    "terms": {"versions": [dt.date(2026, 7, 14), dt.date(2026, 9, 7)]},
    "org": [
        {
            "id": "beispielverein",
            "kind": "verein",
            "name": "Beispielverein Musterstadt e.V.",
            "legal_form": "e.V.",
            "seat": "Musterstadt",
            "country": "DE",
            "status": "vereinbart",
            "agreed_on": dt.date(2026, 1, 15),
            "terms_version": dt.date(2026, 9, 7),
            "agreement_source": "gmail:0000000000000001",
            "public_name_ok": False,
            "contact": [
                {"name": "Erika Mustermann", "role": "Vorsitzende", "emails": ["vorstand@example.org"]}
            ],
        },
        {
            "id": "offener-verein",
            "kind": "verein",
            "name": "Offener Verein e.V.",
            "status": "anfrage",
            "public_name_ok": True,
            "contact": [
                {"name": "Max Beispiel", "emails": ["max@example.net"]},
                {"name": "Ida", "emails": ["ida@example.net"]},
                {"name": "Al B", "emails": ["a@bc"]},
            ],
        },
        {
            "id": "ticketdienst",
            "kind": "dienstleister",
            "name": "Ticketdienst Support",
            "status": "erledigt",
            "contact": [{"name": "Support Team", "emails": ["support@example.com"]}],
        },
    ],
}

TOPICS = {
    "lage": {"notes": ["Ein Satz zur Lage."], "deleted": {"zurueckgetreten": 1}},
    "topic": [
        {
            "id": "pfand",
            "title": "Gläserpfand",
            "kind": "wunsch",
            "status": "offen",
            "ref": "K-99",
            "mention": [
                {"org": "beispielverein", "source": "gmail:0000000000000002",
                 "on": dt.date(2026, 1, 20), "answered": False},
                {"org": "offener-verein", "source": "gh:12",
                 "on": dt.date(2026, 1, 22), "answered": True},
            ],
        },
        {
            "id": "vereinbarung-beispiel",
            "title": "Nutzungsvereinbarung",
            "kind": "verwaltung",
            "status": "beantwortet",
            "mention": [
                {"org": "beispielverein", "source": "gmail:0000000000000001",
                 "on": dt.date(2026, 1, 15), "answered": True},
            ],
        },
    ],
}


def fixture():
    return copy.deepcopy(REGISTER), copy.deepcopy(TOPICS)


class ValidateTest(unittest.TestCase):
    def test_valid_fixture_passes(self):
        self.assertEqual(crm.validate(*fixture()), [])

    def test_rejects_fields_outside_the_schema(self):
        register, topics = fixture()
        register["org"][0]["contact"][0]["phone"] = "0123 456"
        register["org"][1]["address"] = "Hauptstraße 1"
        errors = crm.validate(register, topics)
        self.assertIn("org[0] beispielverein contact[0]: unknown field 'phone'", errors)
        self.assertIn("org[1] offener-verein: unknown field 'address'", errors)

    def test_rejects_unknown_mention_org(self):
        register, topics = fixture()
        topics["topic"][0]["mention"][0]["org"] = "gibt-es-nicht"
        self.assertIn(
            "topic[0] pfand mention[0]: unknown org 'gibt-es-nicht'", crm.validate(register, topics)
        )

    def test_rejects_unknown_terms_version(self):
        register, topics = fixture()
        register["org"][0]["terms_version"] = dt.date(2026, 6, 1)
        self.assertIn(
            "org[0] beispielverein: terms_version must be one of [terms].versions",
            crm.validate(register, topics),
        )

    def test_agreement_needs_date_version_and_source(self):
        register, topics = fixture()
        del register["org"][0]["agreed_on"]
        register["org"][0]["agreement_source"] = "mail:42"
        errors = crm.validate(register, topics)
        self.assertIn("org[0] beispielverein: vereinbart needs agreed_on", errors)
        self.assertIn("org[0] beispielverein: vereinbart needs agreement_source", errors)

    def test_rejects_contact_without_email_and_bad_source(self):
        register, topics = fixture()
        register["org"][1]["contact"][0]["emails"] = []
        topics["topic"][0]["mention"][1]["source"] = "gh:abc"
        errors = crm.validate(register, topics)
        self.assertIn("org[1] offener-verein contact[0]: emails must list at least one address", errors)
        self.assertIn("topic[0] pfand mention[1]: source must be gmail:<16 hex> or gh:<number>", errors)

    def test_rejects_shared_source_finished_org_with_open_reply_and_flat_deleted(self):
        register, topics = fixture()
        topics["topic"][0]["mention"][1]["source"] = "gmail:0000000000000002"
        register["org"][0]["status"] = "erledigt"
        topics["lage"]["deleted"] = 3
        errors = crm.validate(register, topics)
        self.assertIn("topic[0] pfand mention[1]: duplicate source in this topic", errors)
        self.assertIn("org[0] beispielverein: erledigt while a mention still waits for a reply", errors)
        self.assertIn("[lage].deleted: must be a table of org status to count", errors)


class ReportTest(unittest.TestCase):
    def test_open_lists_only_unanswered_mentions(self):
        _, topics = fixture()
        self.assertEqual(
            crm.open_mentions(topics),
            [(dt.date(2026, 1, 20), "beispielverein", "gmail:0000000000000002", "pfand")],
        )

    def test_retention_follows_the_status(self):
        register, topics = fixture()
        for status in ("zurueckgetreten", "bestritten", "erledigt", "anfrage"):
            register["org"].append(
                {"id": f"org-{status}", "kind": "verein", "name": status, "status": status}
            )
        topics["topic"][0]["mention"].append(
            {"org": "org-bestritten", "source": "gh:7", "on": dt.date(2026, 2, 1), "answered": True}
        )
        self.assertEqual(
            crm.retention(register, topics),
            [
                ("ticketdienst", "-", []),
                ("org-zurueckgetreten", "-", []),
                ("org-bestritten", "2026-02-01", ["gh:7"]),
                ("org-erledigt", "-", []),
            ],
        )


EXPECTED_RENDER = """# Feldfeedback

Wünsche, Fehlerberichte und Fragen aus den Mails und GitHub-Issues der Vereine, anonymisiert und gezählt. Keine Namen, Orte, Termine oder Zitate: Die Zuordnung zu Vereinen liegt nur im gitignorierten Register `.crm/`.

`make crm` erzeugt diese Datei aus dem Register; Änderungen von Hand gehen beim nächsten Lauf verloren. „Vereine" zählt die Vereine und Einzelpersonen, die einen Punkt unabhängig voneinander genannt haben.

## Lage

- 2 Organisationen haben die Nutzungsvereinbarung geschlossen, 1 davon zurückgezogen; 1 Vereinbarung gilt.
- Die geltenden Annahmesätze nennen die TERMS-Fassungen vom 7. September 2026 (1).
- 1 weitere Organisation und 0 Einzelpersonen haben angefragt, ohne eine Vereinbarung zu schließen.
- Ein Satz zur Lage.

## Umgesetzt

Keine Einträge.

## Abgelehnt

Keine Einträge.

## Offen

| Wunsch oder Fehler | Vereine | Stand |
| ------------------ | ------: | ----- |
| Gläserpfand        |       2 | K-99  |

## Wiederkehrende Fragen

Keine Einträge.
"""


class RenderTest(unittest.TestCase):
    def test_render_counts_without_names_or_admin_topics(self):
        text = crm.render(*fixture())
        self.assertEqual(text, EXPECTED_RENDER)
        self.assertNotIn("Mustermann", text)
        self.assertNotIn("Nutzungsvereinbarung |", text)


SCRIPTS = Path(__file__).resolve().parent.parent

REGISTER_TOML = """
[terms]
versions = [2026-09-07]

[[org]]
id = "beispielverein"
kind = "verein"
name = "Beispielverein Musterstadt e.V."
status = "anfrage"
  [[org.contact]]
  name = "Erika Mustermann"
  emails = ["vorstand@example.org"]
"""


class PiiTermsTest(unittest.TestCase):
    def test_terms_skip_public_names_providers_first_names_and_short_terms(self):
        register, _ = fixture()
        self.assertEqual(
            crm.pii_terms(register),
            [
                "Beispielverein Musterstadt",
                "Beispielverein Musterstadt e.V.",
                "Erika Mustermann",
                "Max Beispiel",
                "ida@example.net",
                "max@example.net",
                "vorstand@example.org",
            ],
        )

    def test_invalid_register_yields_no_terms(self):
        crm_dir = Path(tempfile.mkdtemp())
        self.addCleanup(shutil.rmtree, crm_dir)
        broken = REGISTER_TOML.replace('["vorstand@example.org"]', '"vorstand@example.org"')
        (crm_dir / "register.toml").write_text(broken, encoding="utf-8")
        (crm_dir / "topics.toml").write_text("", encoding="utf-8")
        out, err = io.StringIO(), io.StringIO()
        with mock.patch.dict(os.environ, {"CRM_DIR": str(crm_dir)}), \
                contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
            self.assertEqual(crm.main(["pii-terms"]), 1)
        self.assertEqual(out.getvalue(), "")
        self.assertEqual(err.getvalue(), "register invalid; run crm.py validate\n")


class CheckPrivateDataTest(unittest.TestCase):
    """Runs scripts/check-private-data.sh against a throwaway git repository."""

    def setUp(self):
        self.tmp = Path(tempfile.mkdtemp())
        self.addCleanup(shutil.rmtree, self.tmp)
        repo = self.tmp / "repo"
        (repo / "scripts" / "crm").mkdir(parents=True)
        for rel in ("check-private-data.sh", "lib.sh", "crm/crm.py"):
            shutil.copy(SCRIPTS / rel, repo / "scripts" / rel)
        (repo / "README.md").write_text("Ein öffentliches Dokument.\n", encoding="utf-8")
        crm_dir = self.tmp / "crm"
        crm_dir.mkdir()
        (crm_dir / "register.toml").write_text(REGISTER_TOML, encoding="utf-8")
        (crm_dir / "topics.toml").write_text("", encoding="utf-8")
        self.repo, self.env = repo, {**os.environ, "CRM_DIR": str(crm_dir)}

    def run_gate(self):
        git = ["git", "-C", str(self.repo)]
        subprocess.run(git + ["init", "-q"], check=True)
        subprocess.run(git + ["add", "-A"], check=True)
        return subprocess.run(
            ["bash", str(self.repo / "scripts" / "check-private-data.sh")],
            env=self.env, capture_output=True, text=True,
        )

    def test_clean_repo_passes(self):
        result = self.run_gate()
        self.assertEqual(result.returncode, 0)
        self.assertIn("No tracked file contains any of the 4 private terms.", result.stderr)

    def test_leaked_contact_email_fails_without_printing_it(self):
        (self.repo / "notes.txt").write_text("Kontakt: Vorstand@Example.org\n", encoding="utf-8")
        result = self.run_gate()
        self.assertEqual(result.returncode, 1)
        self.assertIn("notes.txt:1", result.stderr)
        self.assertNotIn("example.org", (result.stdout + result.stderr).lower())

    def test_term_in_an_unpushed_commit_fails_after_its_removal(self):
        git = ["git", "-C", str(self.repo), "-c", "user.name=Test", "-c", "user.email=test@example.com",
               "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null"]
        remote = self.tmp / "remote.git"
        subprocess.run(["git", "init", "-q", "--bare", str(remote)], check=True)

        def commit(message):
            subprocess.run(git + ["add", "-A"], check=True)
            subprocess.run(git + ["commit", "-q", "-m", message], check=True)
            return subprocess.run(git + ["rev-parse", "HEAD"], check=True, capture_output=True,
                                  text=True).stdout.strip()

        subprocess.run(git + ["init", "-q"], check=True)
        commit("base")
        subprocess.run(git + ["remote", "add", "origin", str(remote)], check=True)
        subprocess.run(git + ["push", "-q", "-u", "origin", "HEAD"], check=True, capture_output=True)
        (self.repo / "notes.txt").write_text("Erika Mustermann\n", encoding="utf-8")
        leaked = commit("add notes")
        (self.repo / "notes.txt").unlink()
        commit("remove notes")
        result = self.run_gate()
        self.assertEqual(result.returncode, 1)
        self.assertIn(f"commit {leaked}: adds a name or e-mail", result.stderr)
        self.assertEqual(result.stderr.count("commit "), 1)
        self.assertNotIn("mustermann", (result.stdout + result.stderr).lower())

    def test_missing_register_passes_with_notice(self):
        self.env["CRM_DIR"] = str(self.tmp / "nonexistent")
        result = self.run_gate()
        self.assertEqual(result.returncode, 0)
        self.assertIn("private-data check skipped", result.stderr)


if __name__ == "__main__":
    unittest.main()
