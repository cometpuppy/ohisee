package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// -----------------------------------------------------------------------
// Keybinds
// -----------------------------------------------------------------------

// validKeys is the set of key strings Bubble Tea / terminals can produce.
var validKeys = map[string]bool{
	// Letters
	"a": true, "b": true, "c": true, "d": true, "e": true,
	"f": true, "g": true, "h": true, "i": true, "j": true,
	"k": true, "l": true, "m": true, "n": true, "o": true,
	"p": true, "q": true, "r": true, "s": true, "t": true,
	"u": true, "v": true, "w": true, "x": true, "y": true, "z": true,
	"A": true, "B": true, "C": true, "D": true, "E": true,
	"F": true, "G": true, "H": true, "I": true, "J": true,
	"K": true, "L": true, "M": true, "N": true, "O": true,
	"P": true, "Q": true, "R": true, "S": true, "T": true,
	"U": true, "V": true, "W": true, "X": true, "Y": true, "Z": true,
	// Digits
	"0": true, "1": true, "2": true, "3": true, "4": true,
	"5": true, "6": true, "7": true, "8": true, "9": true,
	// Special
	"enter": true, "esc": true, " ": true, "space": true,
	"backspace": true, "delete": true, "tab": true,
	"up": true, "down": true, "left": true, "right": true,
	"home": true, "end": true, "pgup": true, "pgdown": true,
	"f1": true, "f2": true, "f3": true, "f4": true,
	"f5": true, "f6": true, "f7": true, "f8": true,
	"f9": true, "f10": true, "f11": true, "f12": true,
	// ctrl+letter
	"ctrl+a": true, "ctrl+b": true, "ctrl+c": true, "ctrl+d": true,
	"ctrl+e": true, "ctrl+f": true, "ctrl+g": true, "ctrl+h": true,
	"ctrl+i": true, "ctrl+j": true, "ctrl+k": true, "ctrl+l": true,
	"ctrl+m": true, "ctrl+n": true, "ctrl+o": true, "ctrl+p": true,
	"ctrl+q": true, "ctrl+r": true, "ctrl+s": true, "ctrl+t": true,
	"ctrl+u": true, "ctrl+v": true, "ctrl+w": true, "ctrl+x": true,
	"ctrl+y": true, "ctrl+z": true,
}

// KeyBinds holds all configurable actions → list of keys
type KeyBinds struct {
	Up        []string `json:"up"`
	Down      []string `json:"down"`
	Check     []string `json:"check"`
	Uncheck   []string `json:"uncheck"`
	Reset     []string `json:"reset"`
	Config    []string `json:"config"`
	Preset    []string `json:"preset"`
	Quit      []string `json:"quit"`
	ToggleAll []string `json:"toggleAll"`
	Confirm   []string `json:"confirm"`
	Back      []string `json:"back"`
}

func defaultKeys() KeyBinds {
	return KeyBinds{
		Up:        []string{"up", "k"},
		Down:      []string{"down", "j"},
		Check:     []string{"l", "right", " "},
		Uncheck:   []string{"h", "left"},
		Reset:     []string{"R"},
		Config:    []string{"c"},
		Preset:    []string{"p"},
		Quit:      []string{"q"},
		ToggleAll: []string{"A"},
		Confirm:   []string{"enter"},
		Back:      []string{"esc"},
	}
}

// matches returns true if key is in the action's list
func (kb KeyBinds) matches(action string, key string) bool {
	var list []string
	switch action {
	case "up":        list = kb.Up
	case "down":      list = kb.Down
	case "check":     list = kb.Check
	case "uncheck":   list = kb.Uncheck
	case "reset":     list = kb.Reset
	case "config":    list = kb.Config
	case "preset":    list = kb.Preset
	case "quit":      list = kb.Quit
	case "toggleAll": list = kb.ToggleAll
	case "confirm":   list = kb.Confirm
	case "back":      list = kb.Back
	}
	for _, k := range list {
		if k == key {
			return true
		}
	}
	return false
}

// label returns a display string for an action e.g. "k/↑"
func (kb KeyBinds) label(action string) string {
	var list []string
	switch action {
	case "up":        list = kb.Up
	case "down":      list = kb.Down
	case "check":     list = kb.Check
	case "uncheck":   list = kb.Uncheck
	case "reset":     list = kb.Reset
	case "config":    list = kb.Config
	case "preset":    list = kb.Preset
	case "quit":      list = kb.Quit
	case "toggleAll": list = kb.ToggleAll
	case "confirm":   list = kb.Confirm
	case "back":      list = kb.Back
	}
	display := make([]string, len(list))
	for i, k := range list {
		switch k {
		case " ":     display[i] = "space"
		case "up":    display[i] = "↑"
		case "down":  display[i] = "↓"
		case "left":  display[i] = "←"
		case "right": display[i] = "→"
		default:      display[i] = k
		}
	}
	return strings.Join(display, "/")
}

func keysPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "ohisee", "keys.json")
}

// keybindError holds a non-fatal error to show in the TUI
type keybindError struct {
	msg string
}

// loadKeys loads keys.json, validates it, and returns the binds + any error.
// If no file exists it writes the defaults and returns them.
func loadKeys() (KeyBinds, *keybindError) {
	defaults := defaultKeys()
	path := keysPath()

	data, err := os.ReadFile(path)
	if err != nil {
		// No file — write defaults so the user can see/edit them
		writeDefaultKeys(path, defaults)
		return defaults, nil
	}

	var kb KeyBinds
	if err := json.Unmarshal(data, &kb); err != nil {
		return defaults, &keybindError{
			msg: fmt.Sprintf("keys.json: invalid JSON: %v — using defaults", err),
		}
	}

	// Validate all keys
	allActions := map[string][]string{
		"up": kb.Up, "down": kb.Down, "check": kb.Check,
		"uncheck": kb.Uncheck, "reset": kb.Reset, "config": kb.Config,
		"preset": kb.Preset, "quit": kb.Quit, "toggleAll": kb.ToggleAll,
		"confirm": kb.Confirm, "back": kb.Back,
	}

	// Check for unknown key names
	for action, keys := range allActions {
		for _, k := range keys {
			if !validKeys[k] {
				return defaults, &keybindError{
					msg: fmt.Sprintf("keys.json: unknown key %q in action %q — using defaults", k, action),
				}
			}
		}
	}

	// Check for conflicts — same key assigned to two different actions
	seen := make(map[string]string) // key → first action that claimed it
	for action, keys := range allActions {
		for _, k := range keys {
			if k == "" || k == "ctrl+c" {
				continue
			}
			if first, conflict := seen[k]; conflict && first != action {
				return defaults, &keybindError{
					msg: fmt.Sprintf("keys.json: key %q is used by both %q and %q — using defaults", k, first, action),
				}
			}
			seen[k] = action
		}
	}

	return kb, nil
}

func writeDefaultKeys(path string, kb KeyBinds) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}
	data, err := json.MarshalIndent(kb, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(path, data, 0644)
}

// -----------------------------------------------------------------------
// Data
// -----------------------------------------------------------------------

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

var minimalItems = []string{
	"Lordvessel",
	"Soul of Gravelord Nito",
	"Soul of Bed of Chaos",
	"Soul of Four Kings",
	"Soul of Seath the Scaleless",
	"Master Key",
	"Key to the Seal",
	"Key to New Londo Ruins",
	"Crest of Artorias",
	"Peculiar Doll",
	"Broken Pendant",
	"Weapon Smithbox",
}

var presets = map[string][]string{
	"Minimal":   {"Progress", "Key Items", "Utility"},
	"Key Items": {"Progress", "Key Items"},
	"Standard":  {"Progress", "Key Items", "Utility"},
	"Full Run":  {"Progress", "Key Items", "Utility", "Archive Keys"},
	"Custom":    {},
}

var presetOrder = []string{"Minimal", "Key Items", "Standard", "Full Run", "Custom"}

var presetDesc = map[string]string{
	"Minimal":   "Essential items only",
	"Key Items": "Progress + Key Items only",
	"Standard":  "Progress + Key Items + Utility",
	"Full Run":  "Everything including Archive Keys",
	"Custom":    "Your own selection",
}

func allItems() []Item {
	type entry struct{ name, category string }
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
		items[i] = Item{Name: e.name, Category: e.category}
	}
	return items
}

func enabledForPreset(preset string, customEnabled []string) map[string]bool {
	enabled := make(map[string]bool)
	if preset == "Custom" {
		for _, name := range customEnabled {
			enabled[name] = true
		}
		return enabled
	}
	if preset == "Minimal" {
		for _, name := range minimalItems {
			enabled[name] = true
		}
		return enabled
	}
	cats, ok := presets[preset]
	if !ok {
		return enabled
	}
	catSet := make(map[string]bool)
	for _, c := range cats {
		catSet[c] = true
	}
	for _, item := range allItems() {
		if catSet[item.Category] {
			enabled[item.Name] = true
		}
	}
	return enabled
}

// -----------------------------------------------------------------------
// Save / Load
// -----------------------------------------------------------------------

type SaveData struct {
	ActivePreset  string   `json:"activePreset"`
	FoundItems    []string `json:"foundItems"`
	CustomEnabled []string `json:"customEnabled"`
}

func savePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "ohisee", "save.json")
}

func saveState(data SaveData) error {
	path := savePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func loadState() SaveData {
	data, err := os.ReadFile(savePath())
	if err != nil {
		return SaveData{}
	}
	var save SaveData
	if err := json.Unmarshal(data, &save); err == nil && save.ActivePreset != "" {
		return save
	}
	return SaveData{}
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

	categoryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A90D9")).
			Bold(true).
			MarginTop(1)

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

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E06C75")).
			Bold(true)

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

type viewItem struct {
	Item
	found   bool
	enabled bool
}

type model struct {
	items         []viewItem
	visibleIdx    []int
	cursor        int
	configCursor  int
	confirming    bool
	width         int
	height        int
	currentScreen screen
	presetCursor  int
	activePreset  string
	customEnabled []string
	fromCustom    bool
	keys          KeyBinds
	keyError      *keybindError
}

func initialModel() model {
	save := loadState()
	kb, kbErr := loadKeys()
	base := allItems()

	foundSet := make(map[string]bool)
	for _, name := range save.FoundItems {
		foundSet[name] = true
	}

	items := make([]viewItem, len(base))
	for i, it := range base {
		items[i] = viewItem{Item: it, found: foundSet[it.Name]}
	}

	m := model{
		items:         items,
		customEnabled: save.CustomEnabled,
		currentScreen: screenPreset,
		keys:          kb,
		keyError:      kbErr,
	}

	if save.ActivePreset != "" {
		for i, p := range presetOrder {
			if p == save.ActivePreset {
				m.presetCursor = i
			}
		}
	}

	return m
}

func (m *model) applyPresetEnabled(preset string) {
	enabled := enabledForPreset(preset, m.customEnabled)
	for i := range m.items {
		m.items[i].enabled = enabled[m.items[i].Name]
	}
}

func (m *model) rebuildVisible() {
	m.visibleIdx = m.visibleIdx[:0]
	for i, item := range m.items {
		if item.enabled {
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

func (m model) currentSaveData() SaveData {
	var found []string
	for _, item := range m.items {
		if item.found {
			found = append(found, item.Name)
		}
	}
	return SaveData{
		ActivePreset:  m.activePreset,
		FoundItems:    found,
		CustomEnabled: m.customEnabled,
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
		// ctrl+c is hardcoded and reserved — not configurable
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
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
	k := msg.String()
	switch {
	case m.keys.matches("quit", k):
		return m, tea.Quit
	case m.keys.matches("up", k):
		if m.presetCursor > 0 {
			m.presetCursor--
		}
	case m.keys.matches("down", k):
		if m.presetCursor < len(presetOrder)-1 {
			m.presetCursor++
		}
	case m.keys.matches("confirm", k):
		chosen := presetOrder[m.presetCursor]
		m.activePreset = chosen
		if chosen == "Custom" {
			for i := range m.items {
				m.items[i].enabled = false
			}
			if len(m.customEnabled) > 0 {
				savedEnabled := enabledForPreset("Custom", m.customEnabled)
				for i := range m.items {
					m.items[i].enabled = savedEnabled[m.items[i].Name]
				}
			}
			m.rebuildVisible()
			m.configCursor = 0
			m.fromCustom = true
			m.currentScreen = screenConfig
		} else {
			m.applyPresetEnabled(chosen)
			m.cursor = 0
			m.rebuildVisible()
			saveState(m.currentSaveData())
			m.currentScreen = screenMain
		}
	}
	return m, nil
}

func (m model) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	if m.confirming {
		switch {
		case k == "y" || k == "Y":
			for i := range m.items {
				m.items[i].found = false
			}
			m.cursor = 0
			m.confirming = false
			saveState(m.currentSaveData())
		case k == "n" || k == "N" || m.keys.matches("back", k):
			m.confirming = false
		}
		return m, nil
	}

	switch {
	case m.keys.matches("quit", k):
		return m, tea.Quit
	case m.keys.matches("up", k):
		if m.cursor > 0 {
			m.cursor--
		}
	case m.keys.matches("down", k):
		if m.cursor < len(m.visibleIdx)-1 {
			m.cursor++
		}
	case m.keys.matches("check", k):
		if len(m.visibleIdx) > 0 {
			idx := m.visibleIdx[m.cursor]
			m.items[idx].found = !m.items[idx].found
			saveState(m.currentSaveData())
		}
	case m.keys.matches("uncheck", k):
		if len(m.visibleIdx) > 0 {
			idx := m.visibleIdx[m.cursor]
			m.items[idx].found = false
			saveState(m.currentSaveData())
		}
	case m.keys.matches("reset", k):
		m.confirming = true
	case m.keys.matches("config", k):
		m.currentScreen = screenConfig
		m.configCursor = 0
		m.fromCustom = false
	case m.keys.matches("preset", k):
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
	k := msg.String()

	switch {
	case m.keys.matches("back", k):
		if m.fromCustom {
			m.fromCustom = false
			m.activePreset = ""
			m.currentScreen = screenPreset
		} else {
			m.currentScreen = screenMain
		}

	case m.keys.matches("config", k):
		if !m.fromCustom {
			m.currentScreen = screenMain
		}

	case m.keys.matches("confirm", k):
		if m.fromCustom {
			var enabled []string
			for _, item := range m.items {
				if item.enabled {
					enabled = append(enabled, item.Name)
				}
			}
			m.customEnabled = enabled
			m.fromCustom = false
			m.rebuildVisible()
			saveState(m.currentSaveData())
			m.currentScreen = screenMain
		} else {
			m.toggleConfigItem()
		}

	case k == " ":
		m.toggleConfigItem()

	case m.keys.matches("toggleAll", k):
		allEnabled := true
		for _, item := range m.items {
			if !item.enabled {
				allEnabled = false
				break
			}
		}
		for i := range m.items {
			m.items[i].enabled = !allEnabled
		}
		if m.activePreset == "Custom" || m.fromCustom {
			var enabled []string
			for _, item := range m.items {
				if item.enabled {
					enabled = append(enabled, item.Name)
				}
			}
			m.customEnabled = enabled
		}
		m.rebuildVisible()
		saveState(m.currentSaveData())

	case m.keys.matches("up", k):
		if m.configCursor > 0 {
			m.configCursor--
		}
	case m.keys.matches("down", k):
		if m.configCursor < len(m.items)-1 {
			m.configCursor++
		}
	}
	return m, nil
}

func (m *model) toggleConfigItem() {
	m.items[m.configCursor].enabled = !m.items[m.configCursor].enabled
	if m.activePreset == "Custom" || m.fromCustom {
		var enabled []string
		for _, item := range m.items {
			if item.enabled {
				enabled = append(enabled, item.Name)
			}
		}
		m.customEnabled = enabled
	}
	m.rebuildVisible()
	saveState(m.currentSaveData())
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

	if m.keyError != nil {
		content += errorStyle.Render("⚠ "+m.keyError.msg) + "\n\n"
	}

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

	upDown := m.keys.label("up") + "/" + m.keys.label("down")
	content += hintStyle.Render(fmt.Sprintf(
		"%s navigate • %s select • %s quit",
		upDown,
		m.keys.label("confirm"),
		m.keys.label("quit"),
	))

	box := boxStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m model) viewMain() string {
	compact := m.activePreset == "Full Run"

	found, total := 0, 0
	for _, item := range m.items {
		if item.enabled {
			total++
			if item.found {
				found++
			}
		}
	}

	content := titleStyle.Render(fmt.Sprintf("OhISee  [%d/%d]  %s", found, total, m.activePreset)) + "\n"

	catHeader := categoryStyle
	if compact {
		catHeader = categoryCompactStyle
	}

	lastCat := ""
	for vi, idx := range m.visibleIdx {
		item := m.items[idx]
		if item.Category != lastCat {
			lastCat = item.Category
			if compact {
				if vi == 0 {
					content += catHeader.Render("── "+item.Category+" ──") + "\n"
				} else {
					content += "\n" + catHeader.Render("── "+item.Category+" ──") + "\n"
				}
			} else {
				content += catHeader.Render("── "+item.Category+" ──") + "\n"
			}
		}
		cursor := "  "
		if m.cursor == vi {
			cursor = cursorStyle.Render("▶ ")
		}
		checkbox := "[ ]"
		name := item.Name
		if item.found {
			checkbox = "[✓]"
			name = checkedStyle.Render(item.Name)
		}
		content += fmt.Sprintf("%s%s %s\n", cursor, checkbox, name)
	}

	if m.confirming {
		content += "\n" + confirmStyle.Render("Reset all items? (y/n)")
	} else {
		upDown := m.keys.label("up") + "/" + m.keys.label("down")
		content += "\n" + hintStyle.Render(fmt.Sprintf(
			"%s navigate • %s check • %s uncheck • %s reset • %s preset • %s config • %s quit",
			upDown,
			m.keys.label("check"),
			m.keys.label("uncheck"),
			m.keys.label("reset"),
			m.keys.label("preset"),
			m.keys.label("config"),
			m.keys.label("quit"),
		))
	}

	box := boxStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m model) viewConfig() string {
	title := "OhISee — Item Config"
	hint := "Toggle individual items on/off"
	var backHint string
	if m.fromCustom {
		title = "OhISee — Custom Preset"
		hint = "Toggle the items you want to track"
		backHint = fmt.Sprintf("%s confirm • %s cancel",
			m.keys.label("confirm"),
			m.keys.label("back"),
		)
	} else {
		backHint = fmt.Sprintf("%s back", m.keys.label("back"))
	}

	content := titleStyle.Render(title) + "\n"
	content += hintStyle.Render(hint) + "\n\n"

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
		if item.enabled {
			toggle = activeConfigStyle.Render("[ on]")
			name = item.Name
		}
		content += fmt.Sprintf("%s%s %s\n", cursor, toggle, name)
	}

	upDown := m.keys.label("up") + "/" + m.keys.label("down")
	content += "\n" + hintStyle.Render(fmt.Sprintf(
		"%s navigate • space toggle • %s toggle all • %s",
		upDown,
		m.keys.label("toggleAll"),
		backHint,
	))

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
