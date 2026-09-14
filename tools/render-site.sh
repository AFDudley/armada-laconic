#!/usr/bin/env bash
#
# render-site.sh — render every site page from Markdown.
#
# Root pages render from their .md via tools/render.sh; the engineering tree
# renders via tools/render-engineering.sh. Every published page is generated
# from Markdown; none is hand-written.
#
#   tools/render-site.sh
#
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# Root pages: "<md>|<out.html>|<title>"
PAGES=(
  "index.md|index.html|Armada × Laconic — private settlement, mobile support, and yield"
  "architecture.md|architecture.html|Architecture — Armada × Laconic tier stack"
  "execution-platform.md|execution-platform.html|Execution platform (ex_net) — Armada × Laconic"
  "yield-clearing.md|yield-clearing.html|Yield & clearing — Armada × Laconic"
  "mobile-privacy.md|mobile-privacy.html|Mobile end-to-end privacy — Armada × Laconic"
  "glossary.md|glossary.html|Glossary — Armada × Laconic"
  "laconic_ethereum_privacy_via_armada.md|laconic_ethereum_privacy_via_armada.html|Leveraging Laconic for Ethereum Privacy (via Armada)"
  "builder-codes.md|builder-codes.html|Builder codes, attribution & liquidity — Armada × Laconic"
  "mobile-proving-research.md|mobile-proving-research.html|Mobile ZK proving — research note — Armada × Laconic"
)

for page in "${PAGES[@]}"; do
  IFS='|' read -r md out title <<< "$page"
  if [ -f "$md" ]; then
    tools/render.sh "$md" "$out" "$title"
  else
    echo "render-site.sh: SKIP $out (missing $md)" >&2
  fi
done

# The root build-plan page is a redirect to the engineering build plan.
cat > build-plan.html <<'HTML'
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta http-equiv="refresh" content="0; url=engineering/build-plan.html" />
<link rel="canonical" href="engineering/build-plan.html" />
<title>Build plan — moved</title>
</head>
<body>
<p>The build plan has moved to <a href="engineering/build-plan.html">engineering/build-plan.html</a>.</p>
</body>
</html>
HTML
echo "render-site.sh: wrote build-plan.html (redirect)"

# Engineering tree.
tools/render-engineering.sh

echo "render-site.sh: done"
