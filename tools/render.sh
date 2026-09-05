#!/usr/bin/env bash
#
# render.sh — render a repo Markdown page into a site HTML page.
#
# Wraps Python-Markdown output in the shared site shell (same <head>, the site
# <nav class="nav">, style.css, and the footer <hr/><p class="small">…</p>) so
# that site pages are GENERATED from Markdown, never hand-written.
#
#   tools/render.sh <input.md> <output.html> "<title>"
#
# Example:
#   tools/render.sh glossary.md glossary.html "Glossary — Armada × Laconic"
#
set -euo pipefail

MARKDOWN=/opt/local/bin/markdown_py-3.13
SED=/opt/local/bin/gsed

usage() {
  echo "usage: tools/render.sh <input.md> <output.html> \"<title>\"" >&2
  exit 2
}

MD="${1:-}";    [ -n "$MD" ]    || usage
OUT="${2:-}";   [ -n "$OUT" ]   || usage
TITLE="${3:-}"; [ -n "$TITLE" ] || usage

[ -x "$MARKDOWN" ] || { echo "render.sh: markdown renderer not found at $MARKDOWN" >&2; exit 1; }
[ -x "$SED" ]      || { echo "render.sh: gsed not found at $SED" >&2; exit 1; }
[ -f "$MD" ]       || { echo "render.sh: input markdown not found: $MD" >&2; exit 1; }

# Render the Markdown body, then strip trailing whitespace from every line so
# the generated HTML stays diff-clean.
BODY="$("$MARKDOWN" -x tables -x toc -x attr_list -x fenced_code -x sane_lists "$MD" \
  | "$SED" -e 's/[[:space:]]*$//')"

OUTBASE="$(basename "$OUT")"
MDBASE="$(basename "$MD")"

# The shared site navigation. The link matching the output page gets class="active".
NAV_LINKS=(
  "index.html|Overview"
  "architecture.html|Architecture"
  "build-plan.html|Build plan"
  "execution-platform.html|Execution platform"
  "yield-clearing.html|Yield &amp; clearing"
  "mobile-privacy.html|Mobile privacy"
  "glossary.html|Glossary"
  "laconic_ethereum_privacy_via_armada.html|Thesis"
  "builder-codes.html|Builder codes"
)

NAV=""
for entry in "${NAV_LINKS[@]}"; do
  href="${entry%%|*}"
  label="${entry##*|}"
  if [ "$href" = "$OUTBASE" ]; then
    NAV+="  <a href=\"$href\" class=\"active\">$label</a>"$'\n'
  else
    NAV+="  <a href=\"$href\">$label</a>"$'\n'
  fi
done

cat > "$OUT" <<HTML
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>${TITLE}</title>
<link rel="stylesheet" href="style.css" />
</head>
<body>
<main>
<nav class="nav">
${NAV}</nav>

${BODY}

<hr/>
<p class="small">Generated from <code>${MDBASE}</code> by <code>tools/render.sh</code> — do not hand-edit. Internal Google Docs require Laconic/Vulcanize access.</p>

</main>
</body>
</html>
HTML

echo "render.sh: wrote $OUT from $MD"
