package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// -----------------------------------------------------------------------
// Data
// -----------------------------------------------------------------------

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Found    bool   `json:"found"`
	Enabled  bool   `json:"enabled"`
}

var presets = map[string][]string{
	"Key Items": {"Progress", "Key Items"},
	"Standard":  {"Progress", "Key Items", "Utility"},
	"Full Run":  {"Progress", "Key Items", "Utility", "Archive Keys"},
}

var presetOrder = []string{"Key Items", "Standard", "Full Run"}

var presetDesc = map[string]string{
	"Key Items": "Progress + Key Items only",
	"Standard":  "Progress + Key Items + Utility",
	"Full Run":  "Everything including Archive Keys",
}

func defaultItems() []Item {
	type entry struct {
		name     string
		category string
	}
	entries := []entry{
		{"Lordvessel", "Progress"},
		{"Soul of Gravelord Nito", "Progress"},
		{"Soul of Bed of Chaos", "Progress"},
		{"Soul of Four Kings", "Progress"},
		{"Soul of Seath the Scaleless", "Progress"},

		{"Master Key", "Key Items"},
		{"Key to the Seal", "Key Items"},
		{"Basement Key", "Key Items"},
		{"Key to the Depths", "Key Items"},
		{"Blighttown Key", "Key Items"},
		{"Key to New Londo Ruins", "Key Items"},
		{"Crest of Artorias", "Key Items"},
		{"Peculiar Doll", "Key Items"},
		{"Broken Pendant", "Key Items"},
		{"Crest Key", "Key Items"},
		{"Undead Asylum F2 West Key", "Key Items"},
		{"Residence Key", "Key Items"},
		{"Mystery Key", "Key Items"},
		{"Sewer Chamber Key", "Key Items"},
		{"Watchtower Basement Key", "Key Items"},
		{"Cage Key", "Key Items"},
		{"Annex Key", "Key Items"},

		{"Weapon Smithbox", "Utility"},
		{"Armor Smithbox", "Utility"},
		{"Repairbox", "Utility"},
		{"Bottomless Box", "Utility"},
		{"Rite of Kindling", "Utility"},

		{"Archive Tower Cell Key", "Archive Keys"},
		{"Archive Prison Extra Key", "Archive Keys"},
		{"Archive Tower Giant Door Key", "Archive Keys"},
		{"Archive Tower Giant Cell Key", "Archive Keys"},
	}

	items := make([]Item, len(entries))
	for i, e := range entries {
		items[i] = Item{Name: e.name, Category: e.category, Found: false, Enabled: true}
	}
	return items
}

func applyPreset(items []Item, preset string) []Item {
	cats, ok := presets[preset]
	if !ok {
		return items
	}
	catSet := make(map[string]bool)
	for _, c := range cats {
		catSet[c] = true
	}
	for i := range items {
		items[i].Enabled = catSet[items[i].Category]
	}
	return items
}

// -----------------------------------------------------------------------
// Save / Load
// -----------------------------------------------------------------------

func savePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "ohisee", "save.json")
}

func saveItems(items []Item) error {
	path := savePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func loadItems() []Item {
	data, err := os.ReadFile(savePath())
	if err != nil {
		return defaultItems()
	}
	var items []Item
	if err := json.Unmarshal(data, &items); err != nil {
		return defaultItems()
	}
	if len(items) > 0 && items[0].Category == "" {
		return defaultItems()
	}
	return items
}

// -----------------------------------------------------------------------
// Styles
// -----------------------------------------------------------------------

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E8B86D")).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	// Normal category header — blank line above
	categoryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A90D9")).
			Bold(true).
			MarginTop(1)

	// Compact category header — no margin
	categoryCompactStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4A90D9")).
				Bold(true)

	checkedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Strikethrough(true)

	disabledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E06C75")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4A90D9")).
			Padding(1, 3)

	confirmStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E06C75")).
			Bold(true)

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	selectedPresetStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E8B86D")).
				Bold(true)

	presetStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cccccc"))

	activeConfigStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#98C379"))

	inactiveConfigStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666"))
)

// -----------------------------------------------------------------------
// Screen type
// -----------------------------------------------------------------------

type screen int

const (
	screenPreset screen = iota
	screenMain
	screenConfig
)

// -----------------------------------------------------------------------
// Model
// -----------------------------------------------------------------------

type model struct {
	items         []Item
	visibleIdx    []int
	cursor        int
	configCursor  int
	confirming    bool
	width         int
	height        int
	currentScreen screen
	presetCursor  int
	activePreset  string
}

func initialModel() model {
	items := loadItems()
	m := model{
		items:         items,
		activePreset:  "",
		currentScreen: screenPreset,
	}
	m.rebuildVisible()
	return m
}

func (m *model) rebuildVisible() {
	m.visibleIdx = m.visibleIdx[:0]
	for i, item := range m.items {
		if item.Enabled {
			m.visibleIdx = append(m.visibleIdx, i)
		}
	}
	if m.cursor >= len(m.visibleIdx) {
		m.cursor = len(m.visibleIdx) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m model) Init() tea.Cmd { return nil }

// -----------------------------------------------------------------------
// Update
// -----------------------------------------------------------------------

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch m.currentScreen {
		case screenPreset:
			return m.updatePreset(msg)
		case screenMain:
			return m.updateMain(msg)
		case screenConfig:
			return m.updateConfig(msg)
		}
	}
	return m, nil
}

func (m model) updatePreset(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.presetCursor > 0 {
			m.presetCursor--
		}
	case "down", "j":
		if m.presetCursor < len(presetOrder)-1 {
			m.presetCursor++
		}
	case "enter", " ":
		chosen := presetOrder[m.presetCursor]
		m.activePreset = chosen
		m.items = applyPreset(m.items, chosen)
		saveItems(m.items)
		m.cursor = 0
		m.rebuildVisible()
		m.currentScreen = screenMain
	}
	return m, nil
}

func (m model) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirming {
		switch msg.String() {
		case "y", "Y":
			m.items = defaultItems()
			m.items = applyPreset(m.items, m.activePreset)
			saveItems(m.items)
			m.cursor = 0
			m.confirming = false
			m.rebuildVisible()
		case "n", "N", "esc":
			m.confirming = false
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.visibleIdx)-1 {
			m.cursor++
		}
	case " ", "right", "l":
		if len(m.visibleIdx) > 0 {
			idx := m.visibleIdx[m.cursor]
			m.items[idx].Found = !m.items[idx].Found
			saveItems(m.items)
		}
	case "left", "h":
		if len(m.visibleIdx) > 0 {
			idx := m.visibleIdx[m.cursor]
			m.items[idx].Found = false
			saveItems(m.items)
		}
	case "R":
		m.confirming = true
	case "c":
		m.currentScreen = screenConfig
		m.configCursor = 0
	case "p":
		m.currentScreen = screenPreset
		for i, p := range presetOrder {
			if p == m.activePreset {
				m.presetCursor = i
			}
		}
	}
	return m, nil
}

func (m model) updateConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "c":
		m.currentScreen = screenMain
	case "up", "k":
		if m.configCursor > 0 {
			m.configCursor--
		}
	case "down", "j":
		if m.configCursor < len(m.items)-1 {
			m.configCursor++
		}
	case " ", "enter":
		m.items[m.configCursor].Enabled = !m.items[m.configCursor].Enabled
		saveItems(m.items)
		m.rebuildVisible()
	}
	return m, nil
}

// -----------------------------------------------------------------------
// View
// -----------------------------------------------------------------------

func (m model) View() string {
	switch m.currentScreen {
	case screenPreset:
		return m.viewPreset()
	case screenConfig:
		return m.viewConfig()
	default:
		return m.viewMain()
	}
}

func (m model) viewPreset() string {
	content := titleStyle.Render("OhISee") + "\n"
	content += subtitleStyle.Render("DS1 Item Randomizer Tracker") + "\n\n"
	content += subtitleStyle.Render("Select a preset to begin:") + "\n\n"

	for i, p := range presetOrder {
		cursor := "  "
		if m.presetCursor == i {
			cursor = cursorStyle.Render("▶ ")
		}
		var name string
		if p == m.activePreset {
			name = selectedPresetStyle.Render(p + " ✓")
		} else {
			name = presetStyle.Render(p)
		}
		content += fmt.Sprintf("%s%s\n   %s\n\n", cursor, name, hintStyle.Render(presetDesc[p]))
	}

	content += hintStyle.Render("↑/↓ k/j navigate • enter select • q quit")
	box := boxStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m model) viewMain() string {
	compact := m.activePreset == "Full Run"

	found, total := 0, 0
	for _, item := range m.items {
		if item.Enabled {
			total++
			if item.Found {
				found++
			}
		}
	}

	content := titleStyle.Render(fmt.Sprintf("OhISee  [%d/%d]  %s", found, total, m.activePreset)) + "\n"

	catHeader := categoryStyle
	newline := "\n"
	if compact {
		catHeader = categoryCompactStyle
		newline = "" // no blank line before first header in compact mode
	}

	lastCat := ""
	for vi, idx := range m.visibleIdx {
		item := m.items[idx]

		if item.Category != lastCat {
			lastCat = item.Category
			if compact {
				// No leading newline for very first header
				if vi == 0 {
					content += catHeader.Render("── "+item.Category+" ──") + "\n"
				} else {
					content += "\n" + catHeader.Render("── "+item.Category+" ──") + "\n"
				}
			} else {
				_ = newline
				content += catHeader.Render("── "+item.Category+" ──") + "\n"
			}
		}

		cursor := "  "
		if m.cursor == vi {
			cursor = cursorStyle.Render("▶ ")
		}
		checkbox := "[ ]"
		name := item.Name
		if item.Found {
			checkbox = "[✓]"
			name = checkedStyle.Render(item.Name)
		}
		content += fmt.Sprintf("%s%s %s\n", cursor, checkbox, name)
	}

	if m.confirming {
		content += "\n" + confirmStyle.Render("Reset all items? (y/n)")
	} else {
		sep := "\n"
		if compact {
			sep = "\n"
		}
		content += sep + hintStyle.Render("↑/↓ k/j • l/→ check • h/← uncheck • R reset • p preset • c config • q quit")
	}

	box := boxStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m model) viewConfig() string {
	content := titleStyle.Render("OhISee — Item Config") + "\n"
	content += hintStyle.Render("Toggle individual items on/off") + "\n\n"

	currentCat := ""
	for i, item := range m.items {
		if item.Category != currentCat {
			currentCat = item.Category
			content += categoryStyle.Render("── "+currentCat+" ──") + "\n"
		}
		cursor := "  "
		if m.configCursor == i {
			cursor = cursorStyle.Render("▶ ")
		}
		toggle := inactiveConfigStyle.Render("[off]")
		name := disabledStyle.Render(item.Name)
		if item.Enabled {
			toggle = activeConfigStyle.Render("[ on]")
			name = item.Name
		}
		content += fmt.Sprintf("%s%s %s\n", cursor, toggle, name)
	}

	content += "\n" + hintStyle.Render("↑/↓ k/j navigate • space/enter toggle • esc/c back")
	box := boxStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// -----------------------------------------------------------------------
// Main
// -----------------------------------------------------------------------

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
