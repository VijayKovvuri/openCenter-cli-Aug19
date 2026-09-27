# `hack/scripts`

Small repository-maintenance and test-support utilities. Run commands from the
repository root unless noted otherwise.

## Scripts

- `audit_doc_frontmatter.py` audits maintained pages under `docs/` and can
  update `last_updated` on explicitly supplied maintained pages.
- `add_purpose_line.py` adds the required purpose line to maintained docs; it
  skips generated command references.
- `check_staged_docs.py` checks whether implementation changes include a
  maintained documentation change.
- `test_check_staged_docs.py` tests the staged-documentation checker.
- `check_doc_links.py` checks local Markdown link paths without network access;
  external URLs and anchors are skipped.
- `test_check_doc_links.py` tests the local-link checker.
- `check_mermaid_fences.py` checks Mermaid fence structure without validating
  Mermaid diagram syntax.
- `test_check_mermaid_fences.py` tests the Mermaid fence checker.
- `openstack-reset.sh` removes OpenStack resources used by lifecycle tests.
- `tf2yaml.py` converts the supported Terraform locals/module subset used by
  GitOps helpers.

Examples:

```bash
python3 hack/scripts/audit_doc_frontmatter.py
python3 hack/scripts/add_purpose_line.py
python3 hack/scripts/check_staged_docs.py cmd/cluster_generate.go README.md
python3 hack/scripts/check_doc_links.py --repo-root .
python3 hack/scripts/check_mermaid_fences.py --repo-root .
```

The documentation-audit scripts intentionally operate on `docs/`; repository
root and source-adjacent Markdown files are not maintained site pages.
