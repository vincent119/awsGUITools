// Package i18n 提供國際化功能的單元測試。
package i18n_test

import (
	"testing"

	"github.com/vin/ck123gogo/internal/i18n"
)

func TestTranslator_BasicTranslation(t *testing.T) {
	translator, err := i18n.New(i18n.English)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{"app.loading", "Loading..."},
		{"action.confirm", "Confirm"},
		{"action.cancel", "Cancel"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := translator.T(tt.key)
			if got != tt.want {
				t.Errorf("T(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestTranslator_TraditionalChinese(t *testing.T) {
	translator, err := i18n.New(i18n.TraditionalChinese)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{"app.loading", "載入中..."},
		{"action.confirm", "確認"},
		{"action.cancel", "取消"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := translator.T(tt.key)
			if got != tt.want {
				t.Errorf("T(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestTranslator_Fallback(t *testing.T) {
	translator, err := i18n.New(i18n.TraditionalChinese)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	// Non-existent key should return the key itself
	key := "nonexistent.key"
	got := translator.T(key)
	if got != key {
		t.Errorf("T(%q) = %q, want %q (fallback to key)", key, got, key)
	}
}

func TestTranslator_SetLanguage(t *testing.T) {
	translator, err := i18n.New(i18n.English)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	// Start with English
	if translator.CurrentLanguage() != i18n.English {
		t.Errorf("CurrentLanguage() = %v, want %v", translator.CurrentLanguage(), i18n.English)
	}

	// Switch to Traditional Chinese
	translator.SetLanguage(i18n.TraditionalChinese)
	if translator.CurrentLanguage() != i18n.TraditionalChinese {
		t.Errorf("CurrentLanguage() = %v, want %v", translator.CurrentLanguage(), i18n.TraditionalChinese)
	}

	// Verify translation uses new language
	got := translator.T("app.loading")
	want := "載入中..."
	if got != want {
		t.Errorf("T(\"app.loading\") = %q, want %q", got, want)
	}
}

func TestTranslator_Tf(t *testing.T) {
	translator, err := i18n.New(i18n.English)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	got := translator.Tf("app.loaded", 42)
	want := "Loaded (42)"
	if got != want {
		t.Errorf("Tf(\"app.loaded\", 42) = %q, want %q", got, want)
	}
}

func TestGlobalTranslator(t *testing.T) {
	// Reset to English
	i18n.SetLanguage(i18n.English)

	got := i18n.T("app.loading")
	want := "Loading..."
	if got != want {
		t.Errorf("T(\"app.loading\") = %q, want %q", got, want)
	}

	// Switch language
	i18n.SetLanguage(i18n.TraditionalChinese)
	got = i18n.T("app.loading")
	want = "載入中..."
	if got != want {
		t.Errorf("T(\"app.loading\") = %q, want %q", got, want)
	}

	// Reset back to English for other tests
	i18n.SetLanguage(i18n.English)
}

func TestNextLanguage(t *testing.T) {
	i18n.SetLanguage(i18n.English)
	next := i18n.NextLanguage()
	if next != i18n.TraditionalChinese {
		t.Errorf("NextLanguage() from English = %v, want %v", next, i18n.TraditionalChinese)
	}

	i18n.SetLanguage(i18n.TraditionalChinese)
	next = i18n.NextLanguage()
	if next != i18n.English {
		t.Errorf("NextLanguage() from TraditionalChinese = %v, want %v", next, i18n.English)
	}

	// Reset
	i18n.SetLanguage(i18n.English)
}

func TestLanguageDisplayName(t *testing.T) {
	tests := []struct {
		lang i18n.Language
		want string
	}{
		{i18n.English, "English"},
		{i18n.TraditionalChinese, "繁體中文"},
	}

	for _, tt := range tests {
		t.Run(string(tt.lang), func(t *testing.T) {
			got := i18n.LanguageDisplayName(tt.lang)
			if got != tt.want {
				t.Errorf("LanguageDisplayName(%v) = %q, want %q", tt.lang, got, tt.want)
			}
		})
	}
}

func TestSupportedLanguages(t *testing.T) {
	langs := i18n.SupportedLanguages()
	if len(langs) != 2 {
		t.Errorf("SupportedLanguages() returned %d languages, want 2", len(langs))
	}

	hasEnglish := false
	hasTraditionalChinese := false
	for _, lang := range langs {
		if lang == i18n.English {
			hasEnglish = true
		}
		if lang == i18n.TraditionalChinese {
			hasTraditionalChinese = true
		}
	}

	if !hasEnglish {
		t.Error("SupportedLanguages() missing English")
	}
	if !hasTraditionalChinese {
		t.Error("SupportedLanguages() missing TraditionalChinese")
	}
}

