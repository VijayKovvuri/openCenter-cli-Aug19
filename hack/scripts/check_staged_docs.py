#!/usr/bin/env python3
"""Require maintained documentation for public behavior changes.

This checker is intentionally read-only.  Pre-commit supplies the paths in its
invocation, so the same command also works when a caller supplies a targeted
file list (for example, in CI).  Paths that no longer exist are ignored: they
represent staged deletions, which are not create/modify/rename changes for the
purposes of this policy.
"""

from __future__ import annotations

import argparse
import os
from pathlib import Path, PurePosixPath
import sys


IMPLEMENTATION_DIRECTORIES = (
    "internal/config/",
    "internal/cloud/",
    "internal/cluster/",
    "internal/plugins/",
    "internal/operations/",
)


def relative_filename(raw: str, repo_root: Path) -> str | None:
    """Return a normalized repository-relative filename, or ``None``."""
    root = repo_root.resolve()
    candidate = Path(raw)
    if not candidate.is_absolute():
        candidate = root / candidate
    try:
        relative = candidate.resolve().relative_to(root)
    except ValueError:
        return None
    return PurePosixPath(relative.as_posix()).as_posix()


def is_qualifying_implementation(filename: str) -> bool:
    """Return whether *filename* is a public-behavior implementation path."""
    path = filename.replace(os.sep, "/").lstrip("./")
    if path == "main.go":
        return True

    if path.startswith("cmd/"):
        return (
            path.endswith(".go")
            and not path.endswith("_test.go")
            and not path.startswith("cmd/docs/")
        )

    return (
        path.endswith(".go")
        and not path.endswith("_test.go")
        and path.startswith(IMPLEMENTATION_DIRECTORIES)
    ) or (path.startswith("schema/") and path.endswith(".json"))


def is_maintained_documentation(filename: str) -> bool:
    """Return whether *filename* is a maintained reader-facing Markdown page."""
    path = filename.replace(os.sep, "/").lstrip("./")
    if path in {"README.md", "SECURITY.md"}:
        return True
    if not path.startswith("docs/") or not path.endswith(".md"):
        return False
    if path == "docs/README.md":
        return False

    parts = PurePosixPath(path).parts
    excluded_parts = {part.lower() for part in parts[1:-1]}
    if excluded_parts & {"codemaps", "superpowers", "work", "reports", "plans"}:
        return False
    if parts[1:3] == ("reference", "opencenter"):
        return False

    # Keep report/plan artifacts out even when they were placed directly under
    # docs/ rather than in one of the conventional artifact directories.
    stem = PurePosixPath(path).stem.lower()
    return not ("report" in stem or stem.endswith("-plan") or "-plan-" in stem)


def existing_paths(
    filenames: list[str], repo_root: Path
) -> tuple[list[str], list[str]]:
    """Normalize supplied paths and retain only files present on disk."""
    existing: list[str] = []
    outside: list[str] = []
    for raw in filenames:
        relative = relative_filename(raw, repo_root)
        if relative is None:
            outside.append(raw)
            continue
        if (repo_root.resolve() / relative).is_file():
            existing.append(relative)
    return sorted(set(existing)), outside


def check(filenames: list[str], repo_root: Path) -> int:
    """Check supplied paths and return a process exit status."""
    paths, outside = existing_paths(filenames, repo_root)
    if outside:
        print("Ignoring paths outside the repository: " + ", ".join(outside))

    implementations = [p for p in paths if is_qualifying_implementation(p)]
    documentation = [p for p in paths if is_maintained_documentation(p)]
    if implementations and not documentation:
        print("Public behavior changes require a staged maintained documentation change:")
        for path in implementations:
            print(f"  implementation: {path}")
        print("Add or update README.md, SECURITY.md, or a maintained docs/*.md page.")
        return 1

    print(
        "OK: "
        f"{len(implementations)} qualifying implementation file(s), "
        f"{len(documentation)} maintained documentation file(s)."
    )
    return 0


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", type=Path, default=Path.cwd())
    parser.add_argument("filenames", nargs="*", help="paths supplied by pre-commit")
    args = parser.parse_args(argv)
    return check(args.filenames, args.repo_root)


if __name__ == "__main__":
    sys.exit(main())
