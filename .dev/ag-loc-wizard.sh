#!/usr/bin/env bash
# ag-loc-wizard.sh — Interactive localisation workflow guide for AGS3D
#
# A step-by-step terminal wizard that teaches new developers how the AGS3D
# localisation pipeline works and how to use it.
#
# Usage:
#   .dev/ag-loc-wizard.sh [project]
#
# If no project is given, defaults to game_prototype.
#
# What it covers:
#   1. Understanding the localisation data model
#   2. How language: headers work in .agdlg and .agcut files
#   3. Exporting strings for translation (ag export --locale)
#   4. Multi-locale export (ag export without --locale)
#   5. Using the interactive TUI (ag loc tui)
#   6. Importing completed translations
#   7. Validation and reporting

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AG_BIN="${REPO_ROOT}/bin/ag"
PROJECT="${1:-${REPO_ROOT}/game_prototype}"

# ── Colours ──────────────────────────────────────────────────────────────────

BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
RED='\033[0;31m'
DIM='\033[2m'
RESET='\033[0m'

header() {
  echo -e "\n${BOLD}${BLUE}━━━ $1 ━━━${RESET}"
}

step() {
  echo -e "\n${BOLD}${GREEN}▶ STEP $1:${RESET} ${BOLD}$2${RESET}"
}

info() {
  echo -e "  ${DIM}$1${RESET}"
}

cmd() {
  echo -e "  ${CYAN}$ $1${RESET}"
}

output() {
  echo -e "  ${MAGENTA}$1${RESET}"
}

warn() {
  echo -e "  ${YELLOW}⚠ $1${RESET}"
}

success() {
  echo -e "  ${GREEN}✓ $1${RESET}"
}

error() {
  echo -e "  ${RED}✗ $1${RESET}"
}

press_enter() {
  echo -e "\n${DIM}  Press Enter to continue…${RESET}"
  read -r
}

clear_screen() {
  echo -e "\033[2J\033[H"
}

wait_key() {
  echo -e "\n${DIM}  Press any key to continue…${RESET}"
  read -n 1 -s
}

# ── Check prerequisites ───────────────────────────────────────────────────────

check_prereqs() {
  header "Prerequisites Check"

  if [[ ! -x "$AG_BIN" ]]; then
    warn "ag binary not found at $AG_BIN"
    info "Building ag first…"
    (cd "$REPO_ROOT/.dev" && ./build-ag.sh)
  fi
  success "ag binary found"

  if [[ ! -f "$PROJECT/game.agp" ]]; then
    error "No game.agp found in $PROJECT"
    exit 1
  fi
  success "Project found: $PROJECT"

  # Check for supported locales
  if grep -q "supported_locales" "$PROJECT/game.agp" 2>/dev/null; then
    local SUPPORTED
    SUPPORTED=$(grep "supported_locales" "$PROJECT/game.agp" | cut -d'"' -f2 || echo "")
    if [[ -n "$SUPPORTED" ]]; then
      success "Supported locales: $SUPPORTED"
    fi
  fi

  wait_key
}

# ── Data model overview ────────────────────────────────────────────────────

explain_model() {
  clear_screen
  header "Localisation Data Model"

  echo -e "
${BOLD}How localisation works in AGS3D:${RESET}

  ${BOLD}1. Source files${RESET} — .agdlg (dialogue) and .agcut (cutscene) files
     contain strings in the author's language.

  ${BOLD}2. Language headers${RESET} — each file or dialogue node can declare:
     ${MAGENTA}language: en${RESET}   (project default if omitted)
     ${MAGENTA}language: fr${RESET}   (per-file override for French-authored content)

  ${BOLD}3. Locale entries${RESET} — strings are extracted with stable keys:
     ${MAGENTA}#loc:my_key_001${RESET}  ← annotation in source files

  ${BOLD}4. .agstrings files${RESET} — one per locale in locale/ directory:
     ${MAGENTA}locale/en.agstrings${RESET}  — source language (filled)
     ${MAGENTA}locale/fr.agstrings${RESET}  — French translation (filled/empty)
     ${MAGENTA}locale/de.agstrings${RESET}  — German translation (filled/empty)

  ${BOLD}5. game.agp declaration${RESET}:
     ${MAGENTA}default_author_locale = \"en\"${RESET}   ← source language
     ${MAGENTA}supported_locales = \"en fr de\"${RESET}  ← all locales
"

  echo -e "
${BOLD}Key insight:${RESET} When ${CYAN}ag export${RESET} (without --locale) runs,
  it writes the source text to the author's locale file and
  creates empty stubs for all other supported locales.
  "
  wait_key
}

# ── Show language header examples ──────────────────────────────────────────

show_headers() {
  clear_screen
  header "Language Headers"

  step 1 "Per-file language declaration (.agcut)"
  echo -e "
  Add ${MAGENTA}language:${RESET} to the header of any .agcut file:

${DIM}  title: tavern_scene
  language: fr          ← author wrote this in French
  voice_session: vs_tavern
  sequence:
    <<fade_in #duration:1.5>>
    <<line barkeep \"Bonjour!\" #loc:tavern_bonjour>>
    ...${RESET}
"

  step 2 "Per-node language declaration (.agdlg)"
  echo -e "
  Add ${MAGENTA}language:${RESET} to a specific dialogue node:

${DIM}  title: merchant_greeting
  character: merchant
  language: de          ← author wrote this node in German
  ---
  Merchant: Willkommen in meinem Laden! #loc:merchant_welcome_de
  ...${RESET}
"

  step 3 "Omitting language (inherits project default)"
  echo -e "
  If ${MAGENTA}language:${RESET} is omitted, the project's
  ${CYAN}default_author_locale${RESET} is used (from game.agp).
  "
  wait_key
}

# ── Show loc key annotation ────────────────────────────────────────────────

show_loc_keys() {
  clear_screen
  header "Localisation Key Annotations"

  step 1 "Adding a loc key to dialogue lines"
  echo -e "
  Use ${MAGENTA}#loc:${RESET} annotation at the end of any spoken line:

${DIM}  Guard: Halt! Who goes there?  #loc:guard_hello_001
  Player: It's me, a traveler.       #loc:guard_player_intro_001
  Guard: State your business.         #loc:guard_purpose_001
  -> I'm looking for the merchant.     #loc:guard_opt_merchant_001
     Guard: He's in the eastern market. #loc:guard_merchant_directions_001
  -> I'm just passing through.        #loc:guard_opt_passing_001${RESET}
"

  step 2 "Adding a loc key to cutscene commands"
  echo -e "
  Use ${MAGENTA}#loc:${RESET} as a named parameter on any localizable command:

${DIM}  <<title_card \"Chapter One\" #duration:3.0 #loc:tc_chapter_1>>
  <<line narrator \"Once upon a time.\" #loc:narration_intro>>
  <<subtitle \"The city gates stood open.\" #loc:subtitle_gate>>
  <<choice \"Continue\" #loc:choice_continue>>${RESET}
"

  step 3 "Automatic key generation"
  echo -e "
  If ${MAGENTA}#loc:${RESET} is omitted, a stable auto-key is generated:
  ${DIM}<node_title>:<line_index>:<hash8>${RESET}

  Example: ${DIM}tavern_greeting:line0:a1b2c3d4${RESET}
  "
  wait_key
}

# ── Export workflow ────────────────────────────────────────────────────────

demo_export() {
  clear_screen
  header "Exporting Strings for Translation"

  step 1 "Single locale export (--locale flag)"
  echo -e "
  Export strings for a specific locale:

${CYAN}  ag export --locale fr --format po${RESET}

  This creates or updates ${MAGENTA}locale/fr.po${RESET} with all strings
  from the source (author's) language.
  "

  step 2 "Multi-locale export (no --locale flag)"
  echo -e "
  Export to ALL supported locales simultaneously:

${CYAN}  ag export${RESET}

  Reads ${CYAN}supported_locales${RESET} from game.agp and generates:
  ${MAGENTA}  locale/strings.en.agstrings  ← source text (filled)
  ${MAGENTA}  locale/strings.fr.agstrings  ← empty stubs
  ${MAGENTA}  locale/strings.de.agstrings  ← empty stubs
  ${MAGENTA}  locale/strings.he.agstrings  ← empty stubs (RTL)
  ${MAGENTA}  locale/strings.ru.agstrings  ← empty stubs (Cyrillic)${RESET}
  "

  step 3 "Agstrings format (recommended)"
  echo -e "
  Use ${MAGENTA}--format agstrings${RESET} for the native AGS3D format:

${CYAN}  ag export --locale fr --format agstrings${RESET}

  This uses the ${MAGENTA}Diff/Apply${RESET} pipeline:
  • New keys → added as empty stubs
  • Removed keys → marked as [orphan]
  • Changed source → marked as [stale]
  "
  wait_key
}

# ── Interactive TUI demo ─────────────────────────────────────────────────

demo_tui() {
  clear_screen
  header "Interactive Translation TUI"

  step 1 "Launch the TUI"
  echo -e "
  The TUI provides an interactive translation editor:

${CYAN}  ag loc tui game_prototype --locale fr${RESET}

  Features:
  • Browse all strings by character, node, or type
  • Mark strings as translated
  • See untranslated / stale / orphan status
  • Search by key or source text
  • Inline editing with keyboard navigation
  "

  step 2 "Navigation keys"
  echo -e "
  ${BOLD}j/k${RESET} or ${BOLD}↑/↓${RESET}    — move between entries
  ${BOLD}Enter${RESET}            — edit selected entry
  ${BOLD}e${RESET}               — edit entry
  ${BOLD}s${RESET}               — save changes
  ${BOLD}q${RESET}               — quit
  ${BOLD}/${RESET}               — search
  ${BOLD}f${RESET}               — filter (untranslated / stale / orphan)
  ${BOLD}r${RESET}               — refresh
  "

  step 3 "Try it now!"
  echo -e "
  ${YELLOW}This will launch the TUI for game_prototype French locale:${RESET}

${CYAN}  ag loc tui ${PROJECT} --locale fr${RESET}
  "
  wait_key

  "$AG_BIN" loc tui "$PROJECT" --locale fr || true
}

# ── Validation and reporting ───────────────────────────────────────────────

demo_validate() {
  clear_screen
  header "Validation and Reporting"

  step 1 "Check for issues"
  echo -e "
  Validate all locale files:

${CYAN}  ag loc check game_prototype${RESET}

  Checks for:
  • Orphan keys (in locale file but not in source)
  • Missing translations (empty values)
  • Stale entries (source text changed)
  "

  step 2 "Find specific strings"
  echo -e "
  Search for strings by key or text:

${CYAN}  ag loc find game_prototype --locale fr --pattern \"*guard*\"${RESET}
${CYAN}  ag loc find game_prototype --locale fr --group-by character${RESET}
  "

  step 3 "Filter by status"
  echo -e "
  Show only untranslated strings:

${CYAN}  ag loc filter game_prototype --locale fr --untranslated${RESET}
  "

  step 4 "Generate a report"
  echo -e "
  Full locale status report:

${CYAN}  ag loc report game_prototype --locale fr --group-by character${RESET}
  "
  wait_key
}

# ── Import workflow ────────────────────────────────────────────────────────

demo_import() {
  clear_screen
  header "Importing Translations"

  step 1 "Import from PO file"
  echo -e "
  After a translator completes the .po file:

${CYAN}  ag loc import game_prototype --locale fr --file translations/fr.po${RESET}

  Keys in the PO that don't exist in the project are reported as invalid.
  "

  step 2 "Import from CSV"
  echo -e "
  For spreadsheets, export as CSV and import:

${CYAN}  ag loc import game_prototype --locale fr --file translations/fr.csv${RESET}
  "

  step 3 "Update workflow"
  echo -e "
  Typical translation workflow:

  1. ${CYAN}ag export --locale fr --format po${RESET}   → send fr.po to translator
  2. Translator edits fr.po
  3. ${CYAN}ag loc import --locale fr --file fr.po${RESET}   → merge into .agstrings
  4. ${CYAN}ag loc check${RESET}                        → verify no issues
  "
  wait_key
}

# ── RTL languages ──────────────────────────────────────────────────────────

demo_rtl() {
  clear_screen
  header "RTL Language Support"

  echo -e "
  AGS3D supports right-to-left (RTL) languages:

  ${BOLD}Hebrew${RESET} (he) and ${BOLD}Arabic${RESET} (ar) require:
  1. ${MAGENTA}rtl = true${RESET} in the [locale.xx] section of game.agp:

${DIM}  [locale.he]
  name = \"Hebrew\"
  rtl = true${RESET}

  2. The runtime reads ${MAGENTA}rtl = true${RESET} and mirrors the UI layout.

  ${BOLD}Note:${RESET} game_prototype includes ${CYAN}he.agstrings${RESET} as a stub
  for Hebrew. When filled in with Hebrew text, the game will
  automatically render UI elements right-to-left.
  "
  wait_key
}

# ── Live demo: export ──────────────────────────────────────────────────────

live_export() {
  clear_screen
  header "Live Demo: Running ag export"

  echo -e "
  ${BOLD}Let's export all locale files for game_prototype:${RESET}

  This will:
  • Read game.agp → find supported_locales
  • Scan all .agdlg and .agcut files
  • Extract strings with their #loc: keys
  • Write one .agstrings file per supported locale
  • Source-language file gets filled strings
  • Other locale files get empty stubs
  "
  wait_key

  echo -e "\n${CYAN}$ $AG_BIN export --project \"$PROJECT\"${RESET}"
  if "$AG_BIN" export --project "$PROJECT" 2>&1; then
    success "Export completed!"
  else
    error "Export failed — check your .agdlg and .agcut files"
  fi

  if [[ -d "$PROJECT/locale" ]]; then
    echo -e "\n  ${BOLD}Generated locale files:${RESET}"
    ls -la "$PROJECT/locale/" | grep "\.agstrings$" | while read -r line; do
      echo -e "  ${DIM}$line${RESET}"
    done
  fi
  wait_key
}

# ── Summary ────────────────────────────────────────────────────────────────

summary() {
  clear_screen
  header "Summary: AGS3D Localisation Pipeline"

  echo -e "
  ${BOLD}Quick Reference:${RESET}

  ${GREEN}Writing source strings:${RESET}
    .agdlg / .agcut files with ${MAGENTA}language:${RESET} header
    Dialogue lines annotated with ${MAGENTA}#loc:my_key_001${RESET}

  ${GREEN}Exporting:${RESET}
    ag export                         → all locales
    ag export --locale fr             → single locale

  ${GREEN}Translating:${RESET}
    ag loc tui <project> --locale fr  → interactive editor
    ag loc find <project> --locale fr → search
    ag loc filter <project> --locale fr --untranslated → untranslated only

  ${GREEN}Importing:${RESET}
    ag loc import <project> --locale fr --file fr.po

  ${GREEN}Validation:${RESET}
    ag loc check <project>             → validate all
    ag loc report <project> --locale fr → full report

  ${GREEN}RTL languages:${RESET}
    Set ${MAGENTA}rtl = true${RESET} in [locale.xx] section of game.agp

  ${BOLD}More help:${RESET}
    ag --help
    ag loc --help
    ag export --help
  "
}

# ── Main menu ─────────────────────────────────────────────────────────────

main_menu() {
  clear_screen
  header "AGS3D Localisation Wizard"

  echo -e "
  Welcome! This wizard walks you through the AGS3D localisation pipeline.

  ${BOLD}Project:${RESET} $PROJECT
  "
  echo -e "
  ${BOLD}Choose a topic:${RESET}

    ${CYAN}1${RESET}  How localisation works (data model)
    ${CYAN}2${RESET}  Language headers in .agcut / .agdlg files
    ${CYAN}3${RESET}  #loc: key annotations
    ${CYAN}4${RESET}  Exporting strings for translation
    ${CYAN}5${RESET}  Interactive translation TUI (demo)
    ${CYAN}6${RESET}  Validation and reporting
    ${CYAN}7${RESET}  Importing completed translations
    ${CYAN}8${RESET}  RTL language support
    ${CYAN}9${RESET}  Live demo: run ag export
   ${CYAN}10${RESET}  Summary & quick reference

    ${RED}q${RESET}  Quit
  "

  echo -en "  ${BOLD}Choice:${RESET} "
  read -r choice
  echo

  case "$choice" in
    1) explain_model ;;
    2) show_headers ;;
    3) show_loc_keys ;;
    4) demo_export ;;
    5) demo_tui ;;
    6) demo_validate ;;
    7) demo_import ;;
    8) demo_rtl ;;
    9) live_export ;;
   10) summary ;;
    q|Q) exit 0 ;;
    *) warn "Invalid choice '$choice'" ;;
  esac

  main_menu
}

# ── Entry point ────────────────────────────────────────────────────────────

if [[ ! -f "$PROJECT/game.agp" ]]; then
  echo -e "${RED}Error: $PROJECT/game.agp not found${RESET}" >&2
  echo -e "Usage: $0 [project_dir]" >&2
  exit 1
fi

check_prereqs
main_menu
