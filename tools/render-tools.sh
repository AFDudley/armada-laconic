# render-tools.sh — find the Markdown renderer and GNU sed on macOS or Linux.
#
# Sourced by tools/render.sh and tools/render-engineering.sh; not run directly.
# An explicit MARKDOWN=/SED= in the environment wins. Otherwise the lookup tries
# the MacPorts paths, then the first match on PATH. GNU sed is required: macOS
# ships BSD sed, so on macOS it must be gsed.
#
# Sets MARKDOWN and SED, or exits 1 with a message prefixed by $1.

_render_find_tool() {
  local candidate
  for candidate in "$@"; do
    if [ -x "$candidate" ]; then echo "$candidate"; return 0; fi
    if command -v "$candidate" >/dev/null 2>&1; then command -v "$candidate"; return 0; fi
  done
  return 1
}

_render_caller="${1:-render}"

MARKDOWN="${MARKDOWN:-$(_render_find_tool /opt/local/bin/markdown_py-3.13 markdown_py markdown_py3 || true)}"
SED="${SED:-$(_render_find_tool /opt/local/bin/gsed gsed sed || true)}"

if [ -z "$MARKDOWN" ] || [ ! -x "$MARKDOWN" ]; then
  echo "$_render_caller: no Python-Markdown CLI found (markdown_py); install Python-Markdown or set MARKDOWN=" >&2
  exit 1
fi
if [ -z "$SED" ] || [ ! -x "$SED" ]; then
  echo "$_render_caller: no sed found; set SED=" >&2
  exit 1
fi
if ! "$SED" --version >/dev/null 2>&1; then
  echo "$_render_caller: $SED is not GNU sed; install gsed or set SED=" >&2
  exit 1
fi
