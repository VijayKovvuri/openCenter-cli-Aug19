"""Focused tests for collect_changed_docs.py."""

from __future__ import annotations

import json
from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).parent))

from collect_changed_docs import ZERO_SHA, collect, decode_nul_paths, is_docs_input, write_outputs


class ChangedDocsCollectorTests(unittest.TestCase):
    def setUp(self) -> None:
        self.directory = tempfile.TemporaryDirectory()
        self.root = Path(self.directory.name)

    def tearDown(self) -> None:
        self.directory.cleanup()

    def test_decode_nul_paths_preserves_whitespace_and_unicode(self) -> None:
        self.assertEqual(
            decode_nul_paths("docs/a b.md\0docs/café.md\0".encode()),
            ("docs/a b.md", "docs/café.md"),
        )

    def test_documentation_input_contract(self) -> None:
        for path in (
            "README.md",
            ".mise.toml",
            ".pre-commit-config.yaml",
            ".github/workflows/docs-p0.yml",
            ".vale/styles/Documentation.yml",
            "cmd/docs/main.go",
            "internal/config/v2schema/schema.go",
            "hack/scripts/check_doc_links.py",
        ):
            self.assertTrue(is_docs_input(path), path)
        self.assertFalse(is_docs_input("internal/cluster/bootstrap_service.go"))

    def test_pull_request_uses_merge_base_and_collects_changed_paths(self) -> None:
        calls: list[tuple[str, ...]] = []

        def git(_root: Path, *args: str) -> bytes:
            calls.append(args)
            if args[0] == "merge-base":
                return b"base-sha\n"
            return b"docs/guide.md\0internal/cluster/service.go\0"

        result = collect(
            self.root,
            "pull_request",
            "head-sha",
            base_ref="main",
            git=git,
        )
        self.assertEqual(result.base, "base-sha")
        self.assertEqual(result.markdown_files, ("docs/guide.md",))
        self.assertTrue(result.has_inputs)
        self.assertEqual(calls[0], ("merge-base", "origin/main", "head-sha"))
        self.assertEqual(calls[1][0], "diff")

    def test_push_uses_before_sha(self) -> None:
        calls: list[tuple[str, ...]] = []

        def git(_root: Path, *args: str) -> bytes:
            calls.append(args)
            return b"internal/cluster/service.go\0"

        result = collect(
            self.root,
            "push",
            "head-sha",
            before="before-sha",
            git=git,
        )
        self.assertEqual(result.base, "before-sha")
        self.assertFalse(result.has_inputs)
        self.assertEqual(calls[0][0], "diff")

    def test_initial_push_and_manual_dispatch_use_full_scan(self) -> None:
        for event_name, before in (("push", ZERO_SHA), ("workflow_dispatch", "")):
            with self.subTest(event_name=event_name):
                calls: list[tuple[str, ...]] = []

                def git(_root: Path, *args: str) -> bytes:
                    calls.append(args)
                    return b"docs/all.md\0"

                result = collect(
                    self.root,
                    event_name,
                    "head-sha",
                    before=before,
                    git=git,
                )
                self.assertIsNone(result.base)
                self.assertTrue(result.has_inputs)
                self.assertEqual(calls[0], ("ls-files", "-z"))

    def test_outputs_use_single_line_json_and_nul_path_file(self) -> None:
        files = ("docs/path with spaces.md", 'docs/quote"and\nline.md')

        def git(_root: Path, *args: str) -> bytes:
            if args[0] == "merge-base":
                return b"base-sha\n"
            return b"\0".join(path.encode() for path in files) + b"\0"

        result = collect(
            self.root,
            "pull_request",
            "head-sha",
            base_ref="main",
            git=git,
        )
        github_output = self.root / "github-output"
        list_file = self.root / "changed-files"
        summary = self.root / "summary"
        write_outputs(result, github_output, list_file, summary)

        output_lines = dict(
            line.split("=", 1)
            for line in github_output.read_text(encoding="utf-8").splitlines()
        )
        self.assertEqual(tuple(json.loads(output_lines["files"])), files)
        self.assertEqual(output_lines["has_files"], "true")
        self.assertEqual(output_lines["has_inputs"], "true")
        self.assertEqual(decode_nul_paths(list_file.read_bytes()), files)
        self.assertIn("Docs P0 change collection", summary.read_text(encoding="utf-8"))

    def test_pull_request_requires_base_ref(self) -> None:
        with self.assertRaisesRegex(ValueError, "requires --base-ref"):
            collect(self.root, "pull_request", "head-sha", git=lambda *_: b"")


if __name__ == "__main__":
    unittest.main()
