"""Tests for crm.py on invented fixtures; run with python3 -m unittest scripts/crm/test_crm.py."""

import copy
import datetime as dt
import sys
import unittest
from pathlib import Path

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
            "contact": [{"name": "Max Beispiel", "emails": ["max@example.net"]}],
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
            "ref": "[backlog.md](backlog.md)",
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


if __name__ == "__main__":
    unittest.main()
