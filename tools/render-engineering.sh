#!/usr/bin/env bash
#
# render-engineering.sh — render every engineering/**/*.md into a site HTML page.
#
# Same markdown-driven pipeline as tools/render.sh (Python-Markdown wrapped in the
# shared site shell: same <head>, the site <nav class="nav">, style.css, footer),
# but walks the engineering/ tree, keeps each page at the same relative path with a
# .html extension, fixes up style.css / nav hrefs for the page's directory depth,
# and rewrites intra-doc .md links to the rendered .html paths.
#
# The engineering/**/*.md files stay the single source of truth — this only renders
# them; do not hand-edit the generated .html.
#
#   tools/render-engineering.sh
#
set -euo pipefail

MARKDOWN=/opt/local/bin/markdown_py-3.13
SED=/opt/local/bin/gsed

[ -x "$MARKDOWN" ] || { echo "render-engineering.sh: markdown renderer not found at $MARKDOWN" >&2; exit 1; }
[ -x "$SED" ]      || { echo "render-engineering.sh: gsed not found at $SED" >&2; exit 1; }

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# Site navigation, as (root-relative-href|label) pairs. The engineering pages get an
# extra "Engineering" entry, marked active on every rendered engineering page.
NAV_LINKS=(
  "index.html|Overview"
  "architecture.html|Architecture"
  "engineering/build-plan.html|Build plan"
  "execution-platform.html|Execution platform"
  "yield-clearing.html|Yield &amp; clearing"
  "mobile-privacy.html|Mobile privacy"
  "glossary.html|Glossary"
  "laconic_ethereum_privacy_via_armada.html|Thesis"
  "builder-codes.html|Builder codes"
  "engineering/README.html|Engineering"
)

render_one() {
  local md="$1"                      # e.g. engineering/05-building-block-view.md
  local out="${md%.md}.html"
  local reldir; reldir="$(dirname "$md")"

  # Path prefix from the output page back to the site root.
  local prefix
  case "$reldir" in
    engineering)                    prefix="../" ;;
    engineering/A-nitro-on-railgun) prefix="../../" ;;
    *) echo "render-engineering.sh: unexpected dir: $reldir" >&2; exit 1 ;;
  esac

  # Title from the first Markdown H1, else the filename.
  local heading
  heading="$("$SED" -n 's/^#[[:space:]]\+//p' "$md" | head -n1)"
  [ -n "$heading" ] || heading="$(basename "${md%.md}")"
  local title="${heading} — Armada × Laconic"

  # Render body, strip trailing whitespace, then rewrite intra-doc .md links
  # (href="...md" / href="...md#anchor") to their rendered .html targets.
  local body
  body="$("$MARKDOWN" -x tables -x toc -x attr_list -x fenced_code -x sane_lists "$md" \
    | "$SED" -e 's/[[:space:]]*$//' \
             -e 's@href="\([^"]\+\)\.md\(#[^"]*\)\?"@href="\1.html\2"@g')"

  # Build the nav with hrefs rewritten to this page's depth; mark Engineering active.
  local nav="" entry href label
  for entry in "${NAV_LINKS[@]}"; do
    href="${prefix}${entry%%|*}"
    label="${entry##*|}"
    if [ "$label" = "Engineering" ]; then
      nav+="  <a href=\"$href\" class=\"active\">$label</a>"$'\n'
    else
      nav+="  <a href=\"$href\">$label</a>"$'\n'
    fi
  done

  cat > "$out" <<HTML
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>${title}</title>
<link rel="stylesheet" href="${prefix}style.css" />
</head>
<body>
<main>
<nav class="nav">
${nav}</nav>

${body}

<hr/>
<p class="small">Generated from <code>${md}</code> by <code>tools/render-engineering.sh</code> — do not hand-edit. Internal Google Docs require Laconic/Vulcanize access.</p>

</main>
<script type="module">
import mermaid from "https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs";
for (const code of document.querySelectorAll("code.language-mermaid")) {
  const div = document.createElement("div");
  div.className = "mermaid";
  div.textContent = code.textContent;
  (code.closest("pre") || code).replaceWith(div);
}
mermaid.initialize({ startOnLoad: false, theme: "dark", securityLevel: "loose" });
mermaid.run();
</script>

</body>
</html>
HTML

  echo "render-engineering.sh: wrote $out from $md"
}

# Walk the engineering tree in a stable order.
while IFS= read -r md; do
  render_one "$md"
done < <(find engineering -type f -name '*.md' | LC_ALL=C sort)
