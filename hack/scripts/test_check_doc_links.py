#!/usr/bin/env python3
"""Focused tests for check_doc_links.py."""

from __future__ import annotations

from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).parent))

from check_doc_links import validate_file


class DocumentationLinkTests(unittest.TestCase):
    def setUp(self) -> None:
        self.directory = tempfile.TemporaryDirectory()
        self.root = Path(self.directory.name)
        (self.root / "docs").mkdir()
        (self.root / "docs" / "guide.md").write_text("# Guide\n", encoding="utf-8")

    def tearDown(self) -> None:
        self.directory.cleanup()

    def write(self, name: str, text: str) -> Path:
        path = self.root / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(text, encoding="utf-8")
        return path

    def test_existing_local_target_passes_and_external_and_anchor_are_skipped(self) -> None:
        page = self.write(
            "docs/page.md",
            "[guide](guide.md#intro) [anchor](#intro) [web](https://example.test/nope)\n",
        )
        self.assertEqual(validate_file(page, self.root), [])

    def test_missing_target_has_source_line_number(self) -> None:
        page = self.write("docs/page.md", "Intro\n\n[missing](missing.md)\n")
        issues = validate_file(page, self.root)
        self.assertEqual(len(issues), 1)
        self.assertEqual(issues[0].line, 3)
        self.assertIn("missing.md", issues[0].message)

    def test_reference_definitions_and_code_fences_are_handled(self) -> None:
        page = self.write(
            "docs/page.md",
            "[guide][g]\n\n[g]: guide.md\n\n```text\n[not-a-link](missing.md)\n```\n",
        )
        self.assertEqual(validate_file(page, self.root), [])

    def test_target_cannot_escape_repository(self) -> None:
        page = self.write("docs/page.md", "[outside](../../outside.md)\n")
        issues = validate_file(page, self.root)
        self.assertEqual(len(issues), 1)
        self.assertIn("outside repository", issues[0].message)


if __name__ == "__main__":
    unittest.main()
