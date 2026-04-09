package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ags3d/ag/internal/loc"
	"github.com/ags3d/ag/internal/project"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type entryView struct {
	LocKey     string
	Source     string
	Translated string
	Char       string
	Scene      string
	LineType   string
	Ctx        string
	Stale      bool
	Orphan     bool
}

type model struct {
	root             string
	localeFile       string
	locale           string
	isSourceLocale   bool
	sf               *loc.StringsFile
	sourceEntries    []loc.LocaleEntryFull
	availableLocales []string

	entries  []entryView
	filtered []entryView
	selected int

	filterMode string

	ti        textinput.Model
	tiBlink   tea.Cmd
	isEditing bool

	savePath string
	changed  bool

	width  int
	height int

	showLocalePicker bool
	pickerSelected   int
}

func RunTUIMain(root, locale string) error {
	manifest, err := project.Load(root)
	if err != nil {
		return fmt.Errorf("load project manifest: %w", err)
	}

	if locale == "" {
		locale = manifest.Localisation.DefaultAuthorLocale
	}
	if locale == "" {
		locale = "en"
	}

	availableLocales := manifest.Localisation.SupportedLocales
	if len(availableLocales) == 0 {
		availableLocales = []string{"en"}
	}

	localePath := findLocaleFile(root, locale)
	if localePath == "" {
		return fmt.Errorf("no locale file found for %q (try ag export --locale %s first)", locale, locale)
	}

	sf, err := loadStringsFile(localePath)
	if err != nil {
		return fmt.Errorf("parse locale file: %w", err)
	}

	isSourceLocale := manifest.Localisation.DefaultAuthorLocale == locale

	sourceEntries, _ := loc.CollectAllLocaleEntriesWithTranslations(root, locale)
	entries := buildEntryViews(sf, sourceEntries, isSourceLocale)

	m := newModel(root, localePath, locale, isSourceLocale, availableLocales, sf, sourceEntries, entries)
	p := tea.NewProgram(&m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run TUI: %w", err)
	}
	return nil
}

func newModel(root, path, locale string, isSource bool, availableLocales []string, sf *loc.StringsFile, src []loc.LocaleEntryFull, entries []entryView) model {
	ti := textinput.NewModel()
	ti.Prompt = ""
	ti.Placeholder = ""
	ti.CharLimit = 0
	if len(entries) > 0 {
		ti.SetValue(bufferForEntry(entries[0], isSource))
	}
	m := model{
		root:             root,
		localeFile:       path,
		locale:           locale,
		isSourceLocale:   isSource,
		availableLocales: availableLocales,
		sf:               sf,
		sourceEntries:    src,
		entries:          entries,
		filtered:         entries,
		selected:         0,
		filterMode:       "all",
		ti:               ti,
		savePath:         path,
		changed:          false,
		showLocalePicker: false,
		pickerSelected:   0,
	}
	return m
}

func bufferForEntry(e entryView, isSource bool) string {
	if isSource && e.Source != "" {
		return e.Source
	}
	return e.Translated
}

func (m *model) reloadLocale(newLocale string) {
	manifest, _ := project.Load(m.root)
	isSource := manifest != nil && manifest.Localisation.DefaultAuthorLocale == newLocale

	localePath := findLocaleFile(m.root, newLocale)
	if localePath == "" {
		return
	}
	sf, err := loadStringsFile(localePath)
	if err != nil {
		return
	}
	sourceEntries, _ := loc.CollectAllLocaleEntriesWithTranslations(m.root, newLocale)
	entries := buildEntryViews(sf, sourceEntries, isSource)

	m.locale = newLocale
	m.localeFile = localePath
	m.isSourceLocale = isSource
	m.sf = sf
	m.sourceEntries = sourceEntries
	m.entries = entries
	m.filtered = entries
	m.selected = 0
	m.filterMode = "all"
	m.isEditing = false
	m.changed = false
	m.savePath = localePath
	m.ti.SetValue("")
	m.ti.Blur()
	if len(m.filtered) > 0 {
		m.ti.SetValue(bufferForEntry(m.filtered[0], isSource))
	}
}

func (m *model) bufferFor(idx int) string {
	if m.isSourceLocale && idx < len(m.filtered) && m.filtered[idx].Source != "" {
		return m.filtered[idx].Source
	}
	if idx < len(m.filtered) {
		return m.filtered[idx].Translated
	}
	return ""
}

func isTranslated(e entryView) bool {
	return strings.TrimSpace(e.Translated) != ""
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) applyFilter() {
	m.filtered = make([]entryView, 0)
	for _, e := range m.entries {
		switch m.filterMode {
		case "all":
			m.filtered = append(m.filtered, e)
		case "untranslated":
			if !isTranslated(e) && !e.Orphan {
				if !m.isSourceLocale {
					m.filtered = append(m.filtered, e)
				} else if e.Source != "" {
					m.filtered = append(m.filtered, e)
				}
			}
		case "translated":
			if isTranslated(e) && !e.Stale && !e.Orphan && !m.isSourceLocale {
				m.filtered = append(m.filtered, e)
			}
		case "stale":
			if e.Stale {
				m.filtered = append(m.filtered, e)
			}
		case "orphan":
			if e.Orphan {
				m.filtered = append(m.filtered, e)
			}
		}
	}
	if m.selected >= len(m.filtered) {
		m.selected = max(0, len(m.filtered)-1)
	}
}

func (m *model) commitEdit() {
	if m.selected >= len(m.filtered) {
		return
	}
	ev := &m.filtered[m.selected]
	ev.Translated = m.ti.Value()

	for i := range m.entries {
		if m.entries[i].LocKey == ev.LocKey {
			m.entries[i].Translated = m.ti.Value()
			break
		}
	}

	sfKey := ev.LocKey
	if idx, ok := m.sf.Index()[sfKey]; ok {
		m.sf.Entries[idx].Value = m.ti.Value()
		m.sf.Entries[idx].Stale = false
	}
	m.changed = true
}

func (m *model) save() {
	content := loc.Write(m.sf)
	if err := os.WriteFile(m.savePath, []byte(content), 0644); err != nil {
		return
	}
	m.changed = false
}

func (m model) localePickerView() string {
	boxW := 40
	boxH := len(m.availableLocales) + 5

	var lines []string
	lines = append(lines, bold("  Choose a locale  "))
	lines = append(lines, "")
	for i, loc := range m.availableLocales {
		prefix := "    "
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
		if i == m.pickerSelected {
			prefix = "  >"
			style = selectedStyle
		}
		tag := ""
		if m.isSourceLocale && loc == m.locale {
			tag = dim(" (source)")
		}
		lines = append(lines, style.Render(fmt.Sprintf("%s %s%s", prefix, loc, tag)))
	}
	lines = append(lines, "")
	lines = append(lines, dim("  ↑↓ navigate  ·  enter select  ·  esc cancel"))

	boxContent := strings.Join(lines, "\n")
	borderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(boxW)
	rendered := lipgloss.Place(boxW, boxH, lipgloss.Left, lipgloss.Top, borderStyle.Render(boxContent))
	full := lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, rendered)

	help := fmt.Sprintf("  l: switch locale  ·  esc cancel  [%s]", m.locale)
	return full + "\n" + statusStyle.Width(m.width).Render(help)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.showLocalePicker {
			switch msg.String() {
			case "up", "k":
				if m.pickerSelected > 0 {
					m.pickerSelected--
				}
			case "down", "j":
				if m.pickerSelected < len(m.availableLocales)-1 {
					m.pickerSelected++
				}
			case "enter":
				chosen := m.availableLocales[m.pickerSelected]
				m.showLocalePicker = false
				if chosen != m.locale {
					m.reloadLocale(chosen)
				}
			case "esc", "ctrl+c", "q":
				m.showLocalePicker = false
			}
			return m, nil
		}

		if m.isEditing {
			if msg.String() == "ctrl+s" {
				m.commitEdit()
				m.isEditing = false
				m.ti.Blur()
				m.save()
				return m, nil
			}
			if msg.String() == "esc" {
				m.isEditing = false
				m.ti.Blur()
				return m, nil
			}
			updated, cmd := m.ti.Update(msg)
			m.ti = updated
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			if m.changed {
				m.changed = false
			} else {
				return m, tea.Quit
			}
			return m, nil

		case "tab":
			m.filterMode = nextFilter(m.filterMode)
			m.applyFilter()
			if m.selected >= len(m.filtered) {
				m.selected = max(0, len(m.filtered)-1)
			}
			if len(m.filtered) > 0 {
				m.ti.SetValue(bufferForEntry(m.filtered[m.selected], m.isSourceLocale))
			}
			return m, nil

		case "l":
			m.showLocalePicker = true
			curIdx := -1
			for i, loc := range m.availableLocales {
				if loc == m.locale {
					curIdx = i
					break
				}
			}
			m.pickerSelected = curIdx
			if m.pickerSelected < 0 {
				m.pickerSelected = 0
			}
			return m, nil

		case "up", "k":
			if m.selected > 0 {
				m.selected--
				if m.selected < len(m.filtered) {
					m.ti.SetValue(bufferForEntry(m.filtered[m.selected], m.isSourceLocale))
				}
			}
			return m, nil

		case "down", "j":
			if m.selected < len(m.filtered)-1 {
				m.selected++
				if m.selected < len(m.filtered) {
					m.ti.SetValue(bufferForEntry(m.filtered[m.selected], m.isSourceLocale))
				}
			}
			return m, nil

		case "enter":
			if len(m.filtered) > 0 {
				m.isEditing = true
				m.ti.SetValue(bufferForEntry(m.filtered[m.selected], m.isSourceLocale))
				return m, m.ti.Focus()
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		m.width = 120
	}
	if m.height == 0 {
		m.height = 30
	}

	if m.showLocalePicker {
		return m.localePickerView()
	}

	listW := min(44, m.width/3)
	detailW := m.width - listW - 1

	listLines := make([]string, 0, m.height)
	listLines = append(listLines, "\n "+headerStyle.Render("Strings")+"\n")
	listLines = append(listLines, " "+filterBar(m.filterMode)+"\n")
	listLines = append(listLines, " "+strings.Repeat("─", listW-2)+"\n")

	viewH := m.height - 6
	start := max(0, m.selected-viewH/2)
	end := min(len(m.filtered), start+viewH)
	if end < start+viewH && start > 0 {
		start = max(0, end-viewH)
	}

	for i := start; i < end; i++ {
		e := m.filtered[i]
		prefix := "  "
		style := lipgloss.NewStyle()
		if i == m.selected {
			prefix = " >"
			style = selectedStyle
		}
		status := "   "
		if e.Orphan {
			status = red(" X ")
		} else if e.Stale {
			status = yellow(" ! ")
		} else if isTranslated(e) {
			status = green(" ✓ ")
		} else {
			status = dim(" · ")
		}

		trunc := e.Source
		if len(trunc) > listW-10 {
			trunc = trunc[:listW-13] + "..."
		}
		line := fmt.Sprintf("%s%s %s", prefix, status, trunc)
		listLines = append(listLines, " "+style.Render(padRight(line, listW-2)))
	}

	if len(m.filtered) == 0 {
		listLines = append(listLines, dim("  (no entries)"))
	}

	countSummary := fmt.Sprintf("  %d/%d", len(m.filtered), len(m.entries))
	listLines = append(listLines, "")
	listLines = append(listLines, dim(countSummary))

	listContent := strings.Join(listLines, "\n")

	detailLines := make([]string, 0, m.height)
	detailLines = append(detailLines, "\n "+detailHeaderStyle.Render("Detail")+"\n")
	detailLines = append(detailLines, " "+strings.Repeat("─", detailW-2)+"\n")

	if len(m.filtered) > 0 && m.selected < len(m.filtered) {
		e := m.filtered[m.selected]
		detailLines = append(detailLines, fmt.Sprintf("  Key:     %s\n", e.LocKey))
		detailLines = append(detailLines, fmt.Sprintf("  Type:    %s\n", e.LineType))
		if e.Char != "" {
			detailLines = append(detailLines, fmt.Sprintf("  Char:    %s\n", e.Char))
		}
		if e.Scene != "" {
			detailLines = append(detailLines, fmt.Sprintf("  Scene:   %s\n", e.Scene))
		}
		if e.Ctx != "" {
			detailLines = append(detailLines, fmt.Sprintf("  Context: %s\n", e.Ctx))
		}
		if e.Stale {
			detailLines = append(detailLines, "\n  "+staleStyle.Render("⚠ stale — source text changed")+"\n")
		}
		if e.Orphan {
			detailLines = append(detailLines, "\n  "+orphanStyle.Render("✗ orphan — key removed from source")+"\n")
		}
		detailLines = append(detailLines, "\n  Source:\n")
		for _, l := range wrapText(e.Source, detailW-6) {
			detailLines = append(detailLines, "  "+sourceStyle.Render(l))
		}
		detailLines = append(detailLines, "\n  Translation:\n")
		if m.isEditing {
			tiView := m.ti.View()
			tiView = strings.TrimPrefix(tiView, m.ti.Prompt)
			for _, l := range wrapText(tiView, detailW-6) {
				detailLines = append(detailLines, "  "+editFieldStyle.Render(l))
				break
			}
		} else {
			if e.Translated != "" {
				for _, l := range wrapText(e.Translated, detailW-6) {
					detailLines = append(detailLines, "  "+translatedStyle.Render(l))
				}
			} else {
				detailLines = append(detailLines, "  "+untranslatedStyle.Render("(untranslated)"))
			}
		}
	} else {
		detailLines = append(detailLines, "\n  (no entry selected)")
	}

	if m.changed && !m.isEditing {
		detailLines = append(detailLines, "\n")
		detailLines = append(detailLines, "  "+changedStyle.Render("[unsaved changes — ctrl+s to save]"))
	}

	detailContent := strings.Join(detailLines, "\n")

	help := helpText(m.filterMode, m.locale, m.isEditing, m.changed)
	statusBar := statusStyle.Width(m.width).Render(help)

	left := lipgloss.Place(listW, m.height-1, lipgloss.Left, lipgloss.Top, listContent)
	right := lipgloss.Place(detailW, m.height-1, lipgloss.Left, lipgloss.Top, detailContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right) + "\n" + statusBar
}

func (m *model) UpdateEntryFromLocaleFile() {
	content, err := os.ReadFile(m.savePath)
	if err != nil {
		return
	}
	sf, err := loc.Parse(m.savePath, string(content))
	if err != nil {
		return
	}
	m.sf = sf
	srcMap := make(map[string]loc.LocaleEntryFull)
	for _, e := range m.sourceEntries {
		srcMap[e.LocKey] = e
	}
	m.entries = make([]entryView, 0, len(sf.Entries))
	for _, e := range sf.Entries {
		ev := entryView{LocKey: e.Key, Translated: e.Value, Stale: e.Stale, Orphan: e.Orphan}
		if se, ok := srcMap[e.Key]; ok {
			ev.Source = se.Source
			ev.Char = se.Character
			ev.Scene = se.NodeTitle
			ev.LineType = se.LineType
		}
		m.entries = append(m.entries, ev)
	}
	m.applyFilter()
}

func loadStringsFile(path string) (*loc.StringsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return loc.Parse(path, string(data))
}

func findLocaleFile(root, locale string) string {
	patterns := []string{
		filepath.Join(root, "locale", "strings."+locale+".agstrings"),
		filepath.Join(root, "locale", locale+".agstrings"),
		filepath.Join(root, "locale", locale+".po"),
		filepath.Join(root, "locale", locale+".csv"),
	}
	for _, p := range patterns {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func buildEntryViews(sf *loc.StringsFile, src []loc.LocaleEntryFull, isSourceLocale bool) []entryView {
	srcMap := make(map[string]loc.LocaleEntryFull)
	for _, e := range src {
		srcMap[e.LocKey] = e
	}
	var views []entryView
	for _, e := range sf.Entries {
		ev := entryView{LocKey: e.Key, Translated: e.Value, Stale: e.Stale, Orphan: e.Orphan}
		if se, ok := srcMap[e.Key]; ok {
			ev.Source = se.Source
			ev.Char = se.Character
			ev.Scene = se.NodeTitle
			ev.LineType = se.LineType
		}
		if isSourceLocale {
			ev.Translated = ""
		}
		views = append(views, ev)
	}
	return views
}

func nextFilter(current string) string {
	switch current {
	case "all":
		return "untranslated"
	case "untranslated":
		return "translated"
	case "translated":
		return "stale"
	case "stale":
		return "orphan"
	default:
		return "all"
	}
}

func filterBar(mode string) string {
	all, untrans, trans, stale, orphan := "[all]", "[untranslated]", "[translated]", "[stale]", "[orphan]"
	switch mode {
	case "all":
		all = bold("[all]")
	case "untranslated":
		untrans = bold("[untranslated]")
	case "translated":
		trans = bold("[translated]")
	case "stale":
		stale = bold("[stale]")
	case "orphan":
		orphan = bold("[orphan]")
	}
	return fmt.Sprintf("%s %s %s %s %s  (tab)", all, untrans, trans, stale, orphan)
}

func helpText(filter, locale string, editing, changed bool) string {
	if editing {
		return "ENTER: confirm  |  ESC: cancel  |  ctrl+s: save  |  q: discard & quit"
	}
	if changed {
		return fmt.Sprintf("ctrl+s: save  |  q: quit  |  l: switch locale  |  ↑↓: navigate  |  tab: filter  |  enter: edit  [%s]", locale)
	}
	return fmt.Sprintf("↑↓: navigate  |  tab: filter  |  enter: edit  |  l: switch locale  |  q: quit  [%s]", locale)
}

func wrapText(s string, width int) []string {
	if width <= 0 {
		width = 60
	}
	var lines []string
	for _, paragraph := range strings.Split(s, "\n") {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}
		words := strings.Fields(paragraph)
		var line string
		for _, w := range words {
			if line == "" {
				line = w
			} else if len(line)+1+len(w) <= width {
				line += " " + w
			} else {
				lines = append(lines, line)
				line = w
			}
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		lines = append(lines, "")
	}
	return lines
}

func padRight(s string, w int) string {
	if w <= 0 {
		return s
	}
	if len(s) >= w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

func cursorBlink() string {
	return "▌"
}

func green(s string) string  { return lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render(s) }
func red(s string) string    { return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(s) }
func yellow(s string) string { return lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render(s) }
func dim(s string) string    { return lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(s) }
func bold(s string) string   { return lipgloss.NewStyle().Bold(true).Render(s) }
