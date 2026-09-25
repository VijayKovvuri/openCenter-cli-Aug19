#!/usr/bin/env python3
"""Focused tests for check_staged_docs.py."""

from __future__ import annotations

import contextlib
import io
from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).parent))

from check_staged_docs import check


class StagedDocumentationPolicyTests(unittest.TestCase):
    def setUp(self) -> None:
        self.directory = tempfile.TemporaryDirectory()
        self.root = Path(self.directory.name)

    def tearDown(self) -> None:
        self.directory.cleanup()

    def add(self, *filenames: str) -> None:
        for filename in filenames:
            path = self.root / filename
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text("placeholder\n", encoding="utf-8")

    def run_check(self, *filenames: str) -> int:
        with contextlib.redirect_stdout(io.StringIO()):
            return check(list(filenames), self.root)

    def test_no_qualifying_code_passes(self) -> None:
        self.add("cmd/root_test.go", "scripts/check.sh")
        self.assertEqual(self.run_check("cmd/root_test.go", "scripts/check.sh"), 0)

    def test_qualifying_code_without_docs_fails(self) -> None:
        self.add("cmd/cluster_status.go")
        self.assertEqual(self.run_check("cmd/cluster_status.go"), 1)

    def test_qualifying_code_with_maintained_docs_passes(self) -> None:
        self.add("internal/config/loader.go", "docs/contributing/pre-commit-hooks.md")
        self.assertEqual(
            self.run_check(
                "internal/config/loader.go", "docs/contributing/pre-commit-hooks.md"
            ),
            0,
        )

    def test_test_only_changes_pass(self) -> None:
        self.add("cmd/cluster_status_test.go", "internal/config/loader_test.go")
        self.assertEqual(
            self.run_check("cmd/cluster_status_test.go", "internal/config/loader_test.go"),
            0,
        )

    def test_excluded_internal_or_generated_docs_do_not_satisfy_policy(self) -> None:
        self.add(
            "main.go",
            "docs/CODEMAPS/runtime.md",
            "docs/superpowers/spec.md",
            "docs/work/plan.md",
            "docs/reference/opencenter/opencenter_root.md",
        )
        self.assertEqual(
            self.run_check(
                "main.go",
                "docs/CODEMAPS/runtime.md",
                "docs/superpowers/spec.md",
                "docs/work/plan.md",
                "docs/reference/opencenter/opencenter_root.md",
            ),
            1,
        )

    def test_root_readme_satisfies_policy(self) -> None:
        self.add("schema/cluster.schema.json", "README.md")
        self.assertEqual(self.run_check("schema/cluster.schema.json", "README.md"), 0)

    def test_deleted_or_missing_paths_are_ignored(self) -> None:
        self.add("docs/contributing/pre-commit-hooks.md")
        self.assertEqual(
            self.run_check("cmd/removed.go", "docs/removed.md"), 0
        )


if __name__ == "__main__":
    unittest.main()
