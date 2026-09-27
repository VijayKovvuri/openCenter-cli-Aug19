#!/usr/bin/env python3
"""Focused tests for check_mermaid_fences.py."""

from __future__ import annotations

from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).parent))

from check_mermaid_fences import validate_file


class MermaidFenceTests(unittest.TestCase):
    def setUp(self) -> None:
        self.directory = tempfile.TemporaryDirectory()
        self.root = Path(self.directory.name)

    def tearDown(self) -> None:
        self.directory.cleanup()

    def write(self, text: str) -> Path:
        path = self.root / "diagram.md"
        path.write_text(text, encoding="utf-8")
        return path

    def test_valid_fence_only_requires_nonempty_content(self) -> None:
        path = self.write("```mermaid\nnot Mermaid-validated syntax\n```\n")
        self.assertEqual(validate_file(path), [])

    def test_exact_opening_syntax_is_required(self) -> None:
        path = self.write("``` mermaid\ngraph TD\n```\n")
        issues = validate_file(path)
        self.assertEqual(len(issues), 1)
        self.assertIn("exactly ```mermaid", issues[0].message)

    def test_empty_and_unclosed_fences_are_reported_with_lines(self) -> None:
        path = self.write("```mermaid\n\n```\n\n```mermaid\ngraph TD\n")
        issues = validate_file(path)
        self.assertEqual([issue.line for issue in issues], [2, 5])
        self.assertIn("non-empty", issues[0].message)
        self.assertIn("closing", issues[1].message)

    def test_non_mermaid_code_fences_are_not_syntax_checked(self) -> None:
        path = self.write("```text\n```mermaid\n```\n")
        self.assertEqual(validate_file(path), [])


if __name__ == "__main__":
    unittest.main()
