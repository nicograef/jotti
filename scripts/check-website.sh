#!/usr/bin/env bash

# Integritäts-Checks für website/ — dependency-frei (bash, grep, sed, awk, comm, find).
#
# Prüft:
#   1. SSI-Includes:  referenzierte Partials existieren, jedes Partial wird eingebunden
#   2. Links & Anker: interne Links sind absolut und zeigen auf existierende Seiten/IDs
#   3. Assets:        beidseitig — referenziert ⇒ existiert, existiert ⇒ referenziert
#   4. CSS-Klassen:   beidseitig — im Markup benutzt ⇆ im Stylesheet definiert

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WEBSITE_DIR="$ROOT_DIR/website"
CSS_FILE="$WEBSITE_DIR/css/base.css"

ERRORS=0

fail() {
  echo "  ✗ $1"
  ERRORS=$((ERRORS + 1))
}

mapfile -t HTML_FILES < <(find "$WEBSITE_DIR" -name '*.html' | sort)

# ── 1. SSI-Includes ──────────────────────────────────────────────

echo "Prüfe SSI-Includes …"

for file in "${HTML_FILES[@]}"; do
  rel="${file#"$WEBSITE_DIR/"}"
  while read -r inc; do
    [ -f "$WEBSITE_DIR$inc" ] || fail "$rel: Include-Ziel fehlt: $inc"
  done < <(grep -oE 'include virtual="[^"]+"' "$file" | sed 's/^include virtual="//;s/"$//' || true)
done

for partial in "$WEBSITE_DIR"/partials/*.html; do
  inc="/partials/$(basename "$partial")"
  grep -rlq --include='*.html' "include virtual=\"$inc\"" "$WEBSITE_DIR" ||
    fail "Partial wird nirgends eingebunden: $inc"
done

# ── 2. Interne Links & Anker ─────────────────────────────────────
# Alle href/src-Referenzen: extern wird übersprungen, interne Pfade
# müssen absolut sein und auf existierende Dateien zeigen, Anker (#id)
# auf eine vorhandene id im Zieldokument.

echo "Prüfe interne Links & Anker …"

page_ids() {
  grep -oE 'id="[^"]+"' "$1" | sed 's/^id="//;s/"$//' | sort -u || true
}

for file in "${HTML_FILES[@]}"; do
  rel="${file#"$WEBSITE_DIR/"}"
  while read -r ref; do
    case "$ref" in
      '' | *'<'*) continue ;; # leer oder SSI-Echo im Attribut
      http://* | https://* | mailto:*) continue ;;
    esac

    path="${ref%%#*}"
    anchor=""
    case "$ref" in *'#'*) anchor="${ref#*#}" ;; esac

    target="$file" # Anker ohne Pfad zeigt auf das eigene Dokument
    if [ -n "$path" ]; then
      case "$path" in
        */) target="$WEBSITE_DIR${path}index.html" ;;
        /*) target="$WEBSITE_DIR$path" ;;
        *)
          fail "$rel: relativer Pfad »$ref« — interne Pfade müssen absolut sein"
          continue
          ;;
      esac
      if [ ! -f "$target" ]; then
        fail "$rel: Link-Ziel existiert nicht: $ref"
        continue
      fi
    fi

    if [ -n "$anchor" ]; then
      ids="$(page_ids "$target")"
      printf '%s\n' "$ids" | grep -qFx -- "$anchor" ||
        fail "$rel: Anker-Ziel »#$anchor« fehlt in ${target#"$WEBSITE_DIR/"}"
    fi
  done < <(grep -oE '(href|src)="[^"]*"' "$file" | sed 's/^[a-z]*="//;s/"$//' | sort -u || true)
done

# url()-Referenzen im Stylesheet (relativ zu css/)
while read -r url; do
  case "$url" in
    /*) target="$WEBSITE_DIR$url" ;;
    *) target="$(realpath -m "$WEBSITE_DIR/css/$url")" ;;
  esac
  [ -f "$target" ] || fail "css/base.css: Asset existiert nicht: $url"
done < <(grep -oE 'url\("[^"]+"\)' "$CSS_FILE" | sed 's/^url("//;s/")$//' | sort -u || true)

# ── 3. Assets: existiert ⇒ referenziert ─────────────────────────
# Gegenrichtung zu Check 2: jede Datei unter img/, icons/, fonts/,
# css/ und js/ muss von HTML oder CSS referenziert werden.

echo "Prüfe Asset-Referenzen …"

REFS="$(
  {
    grep -rhoE '(href|src)="[^"]*"' --include='*.html' "$WEBSITE_DIR" | sed 's/^[a-z]*="//;s/"$//'
    grep -oE 'url\("[^"]+"\)' "$CSS_FILE" | sed 's|^url("||;s|")$||;s|^\.\./|/|'
  } | grep -v '<' | sed 's|^https://jotti\.rocks/|/|' | grep '^/' | sort -u || true
)"

while read -r asset; do
  relasset="${asset#"$WEBSITE_DIR"}"
  printf '%s\n' "$REFS" | grep -qFx -- "$relasset" ||
    fail "Asset wird nirgends referenziert: ${relasset#/}"
done < <(find "$WEBSITE_DIR"/{img,icons,fonts,css,js} -type f | sort)

# ── 4. CSS-Klassen-Konsistenz ────────────────────────────────────
# benutzt = class-Attribute im HTML + classList-Aufrufe im JS,
# definiert = Klassen-Selektoren im Stylesheet (Kommentare entfernt,
# mehrzeilige Selektorlisten zusammengeführt, »\:« entschärft).

echo "Prüfe CSS-Klassen-Konsistenz …"

USED="$(
  {
    grep -rhoE 'class="[^"]*"' --include='*.html' "$WEBSITE_DIR" | sed 's/^class="//;s/"$//' | tr ' \t' '\n'
    grep -rhoE 'classList\.(add|remove|toggle)\("[^"]+"' "$WEBSITE_DIR/js" | sed 's/.*("//;s/"$//'
  } | grep -v '^$' | sort -u || true
)"

DEFINED="$(
  awk '
    {
      line = $0; code = ""
      while (length(line) > 0) {
        if (in_comment) {
          i = index(line, "*/")
          if (i == 0) { line = "" } else { line = substr(line, i + 2); in_comment = 0 }
        } else {
          i = index(line, "/*")
          if (i == 0) { code = code line; line = "" }
          else { code = code substr(line, 1, i - 1); line = substr(line, i + 2); in_comment = 1 }
        }
      }
      if (code ~ /\{/) { sel = buf substr(code, 1, index(code, "{") - 1); buf = "" }
      else if (code ~ /,[ \t]*$/) { buf = buf code; next }
      else { buf = ""; next }
      while (match(sel, /\.[A-Za-z_][A-Za-z0-9_-]*(\\:[A-Za-z0-9_-]+)*/)) {
        cls = substr(sel, RSTART + 1, RLENGTH - 1)
        gsub(/\\:/, ":", cls)
        print cls
        sel = substr(sel, RSTART + RLENGTH)
      }
    }
  ' "$CSS_FILE" | sort -u
)"

while read -r cls; do
  [ -n "$cls" ] && fail "CSS-Klasse benutzt, aber nicht definiert: $cls"
done < <(comm -23 <(printf '%s\n' "$USED") <(printf '%s\n' "$DEFINED"))

while read -r cls; do
  [ -n "$cls" ] && fail "CSS-Klasse definiert, aber nirgends benutzt: $cls"
done < <(comm -13 <(printf '%s\n' "$USED") <(printf '%s\n' "$DEFINED"))

# ── Ergebnis ─────────────────────────────────────────────────────

echo ""
if [ "$ERRORS" -gt 0 ]; then
  echo "✗ check-website: $ERRORS Fehler"
  exit 1
fi
echo "✓ check-website: alle Prüfungen bestanden"
