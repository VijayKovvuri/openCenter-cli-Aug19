#!/usr/bin/env python3
"""Check local Markdown link destinations without accessing the network.

Only paths in Markdown link destinations are checked.  HTTP(S) and other
non-file URLs, as well as fragment-only anchors, are deliberately ignored.
Fragments on a local path are ignored after the path itself is checked.
"""

from __future__ import annotations

import argparse
from dataclasses import dataclass
from pathlib import Path
import re
import sys
from typing import Iterable
from urllib.parse import unquote, urlsplit


_INLINE_LINK = re.compile(r"(?<!\!)\[[^\]\n]+\]\(\s*(?:<([^>\n]*)>|([^\s)]+))")
_REFERENCE_LINK = re.compile(
    r"^\s{0,3}\[[^\]\n]+\]:\s*(?:<([^>\n]*)>|([^\s]+))"
)
_FENCE = re.compile(r"^\s{0,3}(`{3,}|~{3,})(.*)$")
_SCHEME = re.compile(r"^[A-Za-z][A-Za-z0-9+.-]*:")


@dataclass(frozen=True)
class Issue:
    """One invalid local link."""

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


def _destination(value: str) -> str:
    """Remove Markdown escapes and return a link destination."""
    return value.replace(r"\(", "(").replace(r"\)", ")").strip()


def _local_target(destination: str) -> str | None:
    """Return the local path portion, or ``None`` for an ignored URL/anchor."""
    destination = _destination(destination)
    if not destination or destination.startswith("#"):
        return None
    if destination.startswith("//") or _SCHEME.match(destination):
        return None

    parsed = urlsplit(destination)
    # ``urlsplit`` treats a fragment as metadata but leaves a query in the
    # path.  Neither is part of a repository filename.
    return unquote(parsed.path)


def _link_destinations(line: str) -> Iterable[str]:
    """Yield inline and reference-definition destinations from one line."""
    for match in _INLINE_LINK.finditer(line):
        yield match.group(1) if match.group(1) is not None else match.group(2)
    reference = _REFERENCE_LINK.match(line)
    if reference:
        yield reference.group(1) if reference.group(1) is not None else reference.group(2)


def validate_file(path: Path, repo_root: Path) -> list[Issue]:
    """Validate local link paths in one Markdown file."""
    root = repo_root.resolve()
    issues: list[Issue] = []
    try:
        lines = path.read_text(encoding="utf-8").splitlines()
    except OSError as exc:
        return [Issue(path, 1, f"could not read file: {exc}")]

    fence: tuple[str, int] | None = None
    for line_number, line in enumerate(lines, 1):
        fence_match = _FENCE.match(line)
        if fence_match:
            marker, info = fence_match.groups()
            if fence is None:
                fence = (marker[0], len(marker))
            elif marker[0] == fence[0] and len(marker) >= fence[1] and not info.strip():
                fence = None
            continue
        if fence is not None:
            continue

        for raw_destination in _link_destinations(line):
            target = _local_target(raw_destination)
            if target is None:
                continue
            if not target:
                issues.append(Issue(path, line_number, "local link has an empty target"))
                continue

            candidate = Path(target)
            if candidate.is_absolute():
                candidate = root / candidate.as_posix().lstrip("/")
            else:
                candidate = path.parent / candidate
            candidate = candidate.resolve()
            try:
                candidate.relative_to(root)
            except ValueError:
                issues.append(
                    Issue(path, line_number, f"local link targets outside repository: {target}")
                )
                continue
            if not candidate.exists():
                issues.append(Issue(path, line_number, f"local link target does not exist: {target}"))
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

    issues = [issue for path in sorted(set(files)) for issue in validate_file(path, root)]
    if issues:
        for issue in issues:
            print(issue.render(root))
        print(f"FAILED: {len(issues)} invalid local link(s).")
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
