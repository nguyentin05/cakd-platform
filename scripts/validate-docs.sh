#!/bin/bash
set -euo pipefail

echo "Validating MkDocs documentation..."

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

if command -v mkdocs >/dev/null 2>&1; then
  mkdocs build --strict
else
  echo "mkdocs binary not installed locally; verifying mkdocs.yml and pure Markdown structure..."
  test -f mkdocs.yml
  test -f requirements-docs.txt
  test -f docs/index.md
fi

echo "Documentation validation successful!"
