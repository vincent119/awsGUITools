// Package i18n provides internationalization support for the AWS TUI application.
// Supports English (default) and Traditional Chinese (zh-TW).
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed messages/*.json
var messagesFS embed.FS

// Language represents a supported language.
type Language string

const (
	// English is the default language.
	English Language = "en"
	// TraditionalChinese is Traditional Chinese (Taiwan).
	TraditionalChinese Language = "zh-TW"
)

// SupportedLanguages returns all supported languages.
func SupportedLanguages() []Language {
	return []Language{English, TraditionalChinese}
}

// Translator handles message translation.
type Translator struct {
	mu       sync.RWMutex
	current  Language
	messages map[Language]map[string]string
	fallback Language
}

var (
	globalTranslator *Translator
	once             sync.Once
)

// Global returns the global translator instance.
func Global() *Translator {
	once.Do(func() {
		t, err := New(English)
		if err != nil {
			// Fallback to empty translator if loading fails
			t = &Translator{
				current:  English,
				messages: make(map[Language]map[string]string),
				fallback: English,
			}
		}
		globalTranslator = t
	})
	return globalTranslator
}

// New creates a new Translator with the specified default language.
func New(defaultLang Language) (*Translator, error) {
	t := &Translator{
		current:  defaultLang,
		messages: make(map[Language]map[string]string),
		fallback: English,
	}

	// Load all supported languages
	for _, lang := range SupportedLanguages() {
		if err := t.loadLanguage(lang); err != nil {
			return nil, fmt.Errorf("load language %s: %w", lang, err)
		}
	}

	return t, nil
}

// loadLanguage loads messages for a specific language from embedded files.
func (t *Translator) loadLanguage(lang Language) error {
	filename := fmt.Sprintf("messages/%s.json", lang)
	data, err := messagesFS.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read %s: %w", filename, err)
	}

	var messages map[string]string
	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("parse %s: %w", filename, err)
	}

	t.mu.Lock()
	t.messages[lang] = messages
	t.mu.Unlock()

	return nil
}

// SetLanguage changes the current language.
func (t *Translator) SetLanguage(lang Language) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Only set if the language is loaded
	if _, ok := t.messages[lang]; ok {
		t.current = lang
	}
}

// CurrentLanguage returns the current language.
func (t *Translator) CurrentLanguage() Language {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.current
}

// T translates a message key to the current language.
// If the key is not found, it falls back to English, then returns the key itself.
func (t *Translator) T(key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Try current language
	if msgs, ok := t.messages[t.current]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}

	// Fallback to English
	if t.current != t.fallback {
		if msgs, ok := t.messages[t.fallback]; ok {
			if msg, ok := msgs[key]; ok {
				return msg
			}
		}
	}

	// Return the key as-is if not found
	return key
}

// Tf translates a message key with format arguments.
func (t *Translator) Tf(key string, args ...interface{}) string {
	return fmt.Sprintf(t.T(key), args...)
}

// T is a convenience function that uses the global translator.
func T(key string) string {
	return Global().T(key)
}

// Tf is a convenience function that uses the global translator with format arguments.
func Tf(key string, args ...interface{}) string {
	return Global().Tf(key, args...)
}

// SetLanguage sets the language for the global translator.
func SetLanguage(lang Language) {
	Global().SetLanguage(lang)
}

// CurrentLanguage returns the current language of the global translator.
func CurrentLanguage() Language {
	return Global().CurrentLanguage()
}

// NextLanguage returns the next language in the cycle.
func NextLanguage() Language {
	current := CurrentLanguage()
	langs := SupportedLanguages()
	for i, lang := range langs {
		if lang == current {
			return langs[(i+1)%len(langs)]
		}
	}
	return English
}

// LanguageDisplayName returns the display name for a language.
func LanguageDisplayName(lang Language) string {
	switch lang {
	case English:
		return "English"
	case TraditionalChinese:
		return "繁體中文"
	default:
		return string(lang)
	}
}

