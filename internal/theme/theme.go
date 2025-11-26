package theme

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

//go:embed themes/*.json
var builtinThemes embed.FS

// Definition 描述單一主題的色彩配置。
type Definition struct {
	Name                   string `json:"name"`
	PrimitiveBackgroundHex string `json:"primitive_background"`
	ContrastBackgroundHex  string `json:"contrast_background"`
	TextHex                string `json:"text"`
	SecondaryTextHex       string `json:"secondary_text"`
	BorderHex              string `json:"border"`
	SelectionHex           string `json:"selection"`
}

// Manager 管理主題載入與套用。
type Manager struct {
	defs    map[string]Definition
	current string
	mu      sync.RWMutex
	names   []string
}

// NewManager 建立並載入內建主題。
func NewManager() (*Manager, error) {
	entries, err := builtinThemes.ReadDir("themes")
	if err != nil {
		return nil, fmt.Errorf("list themes: %w", err)
	}

	defs := make(map[string]Definition)
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		raw, err := builtinThemes.ReadFile("themes/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read theme %s: %w", entry.Name(), err)
		}

		var def Definition
		if err := json.Unmarshal(raw, &def); err != nil {
			return nil, fmt.Errorf("decode theme %s: %w", entry.Name(), err)
		}
		key := strings.ToLower(def.Name)
		defs[key] = def
		names = append(names, def.Name)
	}

	return &Manager{
		defs:  defs,
		names: names,
	}, nil
}

// Apply 更新 tview 樣式為指定主題。
func (m *Manager) Apply(name string) error {
	if name == "" {
		name = "dark"
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	def, ok := m.defs[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("theme %s not found", name)
	}

	tview.Styles.PrimitiveBackgroundColor = parseColor(def.PrimitiveBackgroundHex)
	tview.Styles.ContrastBackgroundColor = parseColor(def.ContrastBackgroundHex)
	tview.Styles.PrimaryTextColor = parseColor(def.TextHex)
	tview.Styles.SecondaryTextColor = parseColor(def.SecondaryTextHex)
	tview.Styles.BorderColor = parseColor(def.BorderHex)
	tview.Styles.TitleColor = parseColor(def.TextHex)
	tview.Styles.GraphicsColor = parseColor(def.SelectionHex)

	m.current = def.Name
	return nil
}

// Current 回傳目前主題名稱。
func (m *Manager) Current() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// Names 回傳所有可用主題名稱。
func (m *Manager) Names() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, len(m.names))
	copy(names, m.names)
	return names
}

func parseColor(hex string) tcell.Color {
	if hex == "" {
		return tcell.ColorWhite
	}
	// tcell 支援 "#RRGGBB" 形式
	return tcell.GetColor(hex)
}
