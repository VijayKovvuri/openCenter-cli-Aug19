#!/usr/bin/env python3
"""Check Markdown Mermaid fence structure, not Mermaid diagram syntax.

The checker only verifies that Mermaid fences use the exact opening marker,
have a matching closing fence, and contain a first non-empty statement.
"""

from __future__ import annotations

import argparse
from dataclasses import dataclass
from pathlib import Path
import re
import sys


_FENCE = re.compile(r"^\s{0,3}(`{3,}|~{3,})(.*)$")


@dataclass(frozen=True)
class Issue:
    """One Mermaid fence structure issue."""

    path: Path
    line: int
    message: str

    def render(self, repo_root: Path) -> str:
        try:
            filename = self.path.relative_to(repo_root).as_posix()
        except ValueError:
            filename = self.path.as_posix()
        return f"{filename}:{self.line}: {self.message}"


def collect_markdown(repo_root: Path) -> list[Path]:
    """Return Markdown files in the repository, in stable order."""
    root = repo_root.resolve()
    return sorted(
        path
        for path in root.rglob("*.md")
        if not any(part.startswith(".") for part in path.relative_to(root).parts)
        and path.is_file()
    )


def _is_mermaid_candidate(marker: str, info: str) -> bool:
    """Return whether a fence appears intended to be a Mermaid fence."""
    normalized_info = info.strip().lower()
    return normalized_info.startswith("mermaid") or (
        marker.startswith("~") and normalized_info == "mermaid"
    )


def validate_file(path: Path) -> list[Issue]:
    """Validate Mermaid fence pairing and the first diagram statement."""
    issues: list[Issue] = []
    try:
        lines = path.read_text(encoding="utf-8").splitlines()
    except OSError as exc:
        return [Issue(path, 1, f"could not read file: {exc}")]

    fence: tuple[str, int, int, bool, bool] | None = None
    # (character, length, opening line, is Mermaid, has statement)
    for line_number, line in enumerate(lines, 1):
        match = _FENCE.match(line)
        if match is None:
            if fence is not None and line.strip():
                fence = (*fence[:4], True)  # type: ignore[misc]
            continue

        marker, info = match.groups()
        if fence is None:
            is_mermaid = _is_mermaid_candidate(marker, info)
            if is_mermaid and line.strip() != "```mermaid":
                issues.append(Issue(path, line_number, "Mermaid opening fence must be exactly ```mermaid"))
            fence = (marker[0], len(marker), line_number, is_mermaid, False)
            continue

        character, length, opening_line, is_mermaid, has_statement = fence
        if marker[0] != character or len(marker) < length or info.strip():
            if is_mermaid and line.strip() == "```mermaid":
                # A Mermaid marker inside an open block is content, not a close.
                if line.strip() == "```mermaid":
                    fence = (*fence[:4], True)  # type: ignore[misc]
                continue
            if is_mermaid:
                # Another fence-like line is content unless it closes this
                # block according to Markdown's matching-fence rules.
                if line.strip():
                    fence = (*fence[:4], True)  # type: ignore[misc]
            continue

        if is_mermaid and not has_statement:
            issues.append(Issue(path, opening_line + 1, "Mermaid diagram must have a non-empty first statement"))
        fence = None

    if fence is not None:
        _character, _length, opening_line, is_mermaid, _has_statement = fence
        if is_mermaid:
            issues.append(Issue(path, opening_line, "Mermaid opening fence has no matching closing fence"))
    return issues


def check(repo_root: Path, paths: list[str] | None = None) -> int:
    """Check repository Markdown files and return a process exit status."""
    root = repo_root.resolve()
    if paths:
        files = []
        for raw_path in paths:
            candidate = Path(raw_path)
            if not candidate.is_absolute():
                candidate = root / candidate
            candidate = candidate.resolve()
            if candidate.is_file() and candidate.suffix.lower() == ".md":
                files.append(candidate)
    else:
        files = collect_markdown(root)

    issues = [issue for path in sorted(set(files)) for issue in validate_file(path)]
    if issues:
        for issue in issues:
            print(issue.render(root))
        print(f"FAILED: {len(issues)} Mermaid fence issue(s).")
        return 1
    print(f"OK: {len(files)} Markdown file(s) passed.")
    return 0


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--repo-root", default=".", type=Path, help="Repository root (default: current directory)"
    )
    parser.add_argument("paths", nargs="*", metavar="PATH", help="Markdown files to check")
    args = parser.parse_args(argv)
    return check(args.repo_root, args.paths or None)


if __name__ == "__main__":
    sys.exit(main())
