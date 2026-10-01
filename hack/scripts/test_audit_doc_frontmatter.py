"""Focused tests for audit_doc_frontmatter.py."""

from __future__ import annotations

from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).parent))

from audit_doc_frontmatter import audit_file


class FrontmatterYamlSafetyTests(unittest.TestCase):
    def setUp(self) -> None:
        self.directory = tempfile.TemporaryDirectory()
        self.path = Path(self.directory.name) / "page.md"

    def tearDown(self) -> None:
        self.directory.cleanup()

    def audit(self, sidebar_label: str, tags: str = "[services, templates]") -> list[str]:
        self.path.write_text(
            "\n".join(
                (
                    "---",
                    "last_updated: 2026-09-24",
                    "id: services-templates",
                    'title: "Service Templates: How They Work"',
                    f"sidebar_label: {sidebar_label}",
                    "description: Explains how service templates work.",
                    "doc_type: explanation",
                    'audience: "platform engineers, operators"',
                    f"tags: {tags}",
                    "---",
                    "# Service Templates",
                    "",
                )
            ),
            encoding="utf-8",
        )
        return [issue.message for issue in audit_file(self.path)]

    def test_unquoted_mapping_separator_is_rejected(self) -> None:
        issues = self.audit("Service Templates: How They Work")
        self.assertTrue(any("contains ': '" in issue for issue in issues))

    def test_quoted_mapping_separator_passes(self) -> None:
        self.assertEqual(self.audit('"Service Templates: How They Work"'), [])

    def test_unquoted_comment_marker_is_rejected(self) -> None:
        issues = self.audit("Service Templates # overview")
        self.assertTrue(any("YAML treats as a comment" in issue for issue in issues))

    def test_inline_tags_sequence_passes(self) -> None:
        self.assertEqual(self.audit("Service Templates", "[services, templates]"), [])

    def test_unterminated_quote_is_rejected(self) -> None:
        issues = self.audit('"Service Templates')
        self.assertTrue(any("unterminated quoted string" in issue for issue in issues))


if __name__ == "__main__":
    unittest.main()
