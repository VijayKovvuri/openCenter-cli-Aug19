"""Collect documentation inputs for stable pull request, push, and manual CI runs."""

from __future__ import annotations

import argparse
import html
import json
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Callable

ZERO_SHA = "0" * 40
DOC_INPUT_FILES = {
    ".github/workflows/docs-p0.yml",
    ".mise.toml",
    ".pre-commit-config.yaml",
    ".vale.ini",
    "cmd/docs_drift_test.go",
}
DOC_INPUT_PREFIXES = (
    ".vale/styles/",
    "cmd/docs/",
    "docs/",
    "hack/scripts/",
    "internal/config/v2schema/",
    "schema/",
)

GitRunner = Callable[..., bytes]


@dataclass(frozen=True)
class Collection:
    """Documentation-relevant paths and metadata for one workflow event."""

    event_name: str
    head: str
    base: str | None
    paths: tuple[str, ...]
    markdown_files: tuple[str, ...]
    has_inputs: bool


def run_git(repo_root: Path, *args: str) -> bytes:
    """Run a read-only Git command and return its raw output."""
    result = subprocess.run(
        ("git", *args),
        cwd=repo_root,
        check=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    return result.stdout


def decode_nul_paths(data: bytes) -> tuple[str, ...]:
    """Decode Git's NUL-delimited path output without whitespace splitting."""
    if not data:
        return ()
    parts = data.split(b"\0")
    if parts[-1] == b"":
        parts.pop()
    return tuple(part.decode("utf-8") for part in parts)


def is_docs_input(path: str) -> bool:
    """Return whether a changed path should activate documentation checks."""
    return (
        path.endswith(".md")
        or path in DOC_INPUT_FILES
        or any(path.startswith(prefix) for prefix in DOC_INPUT_PREFIXES)
    )


def resolve_base(
    repo_root: Path,
    event_name: str,
    head: str,
    base_ref: str,
    before: str,
    git: GitRunner = run_git,
) -> str | None:
    """Resolve the comparison base, or None when the event needs a full scan."""
    if event_name == "pull_request":
        if not base_ref:
            raise ValueError("pull_request event requires --base-ref")
        return git(repo_root, "merge-base", f"origin/{base_ref}", head).decode().strip()
    if event_name == "push":
        return before if before and before != ZERO_SHA else None
    if event_name == "workflow_dispatch":
        return None
    raise ValueError(f"unsupported event: {event_name}")


def collect(
    repo_root: Path,
    event_name: str,
    head: str,
    base_ref: str = "",
    before: str = "",
    git: GitRunner = run_git,
) -> Collection:
    """Collect changed paths and Markdown files for one supported event."""
    base = resolve_base(repo_root, event_name, head, base_ref, before, git)
    if base is None:
        raw_paths = git(repo_root, "ls-files", "-z")
    else:
        raw_paths = git(
            repo_root,
            "diff",
            "--name-only",
            "-z",
            "--diff-filter=ACMR",
            base,
            head,
            "--",
        )
    paths = decode_nul_paths(raw_paths)
    markdown_files = tuple(path for path in paths if path.endswith(".md"))
    return Collection(
        event_name=event_name,
        head=head,
        base=base,
        paths=paths,
        markdown_files=markdown_files,
        has_inputs=event_name == "workflow_dispatch"
        or any(is_docs_input(path) for path in paths),
    )


def write_outputs(
    collection: Collection,
    github_output: Path,
    list_file: Path,
    step_summary: Path | None = None,
) -> None:
    """Write GitHub outputs, a NUL path file, and an optional safe summary."""
    encoded_files = json.dumps(
        collection.markdown_files, ensure_ascii=False, separators=(",", ":")
    )
    with github_output.open("a", encoding="utf-8") as output:
        output.write(f"has_inputs={str(collection.has_inputs).lower()}\n")
        output.write(f"has_files={str(bool(collection.markdown_files)).lower()}\n")
        output.write(f"files={encoded_files}\n")
        output.write(f"base={collection.base or 'full-scan'}\n")

    list_file.parent.mkdir(parents=True, exist_ok=True)
    with list_file.open("wb") as output:
        for path in collection.markdown_files:
            output.write(path.encode("utf-8") + b"\0")

    if step_summary is not None:
        summary_payload = html.escape(
            json.dumps(collection.markdown_files, ensure_ascii=False, indent=2)
        )
        with step_summary.open("a", encoding="utf-8") as summary:
            summary.write("### Docs P0 change collection\n")
            summary.write(f"- Event: `{collection.event_name}`\n")
            summary.write(f"- Base: `{collection.base or 'full-scan'}`\n")
            summary.write(f"- Head: `{collection.head}`\n")
            summary.write(f"- Documentation inputs: `{collection.has_inputs}`\n")
            summary.write(f"- Markdown files: `{len(collection.markdown_files)}`\n")
            summary.write(f"<pre>{summary_payload}</pre>\n")


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", type=Path, default=Path("."))
    parser.add_argument("--event-name", required=True)
    parser.add_argument("--head", required=True)
    parser.add_argument("--base-ref", default="")
    parser.add_argument("--before", default="")
    parser.add_argument("--github-output", required=True, type=Path)
    parser.add_argument("--list-file", required=True, type=Path)
    parser.add_argument("--step-summary", type=Path)
    args = parser.parse_args(argv)

    try:
        collection = collect(
            args.repo_root.resolve(),
            args.event_name,
            args.head,
            args.base_ref,
            args.before,
        )
        write_outputs(
            collection,
            args.github_output,
            args.list_file,
            args.step_summary,
        )
    except (OSError, UnicodeError, ValueError, subprocess.CalledProcessError) as exc:
        print(f"documentation change collection failed: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
