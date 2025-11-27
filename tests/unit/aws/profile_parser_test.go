package aws_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vin/ck123gogo/internal/aws/profile"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name            string
		configContent   string
		credsContent    string
		wantProfiles    []string
		wantRegions     map[string]string
		wantErr         bool
	}{
		{
			name: "basic config with default and named profiles",
			configContent: `
[default]
region = us-east-1
output = json

[profile dev]
region = ap-northeast-1

[profile prod]
region = us-west-2
`,
			credsContent: `
[default]
aws_access_key_id = AKIA123
aws_secret_access_key = secret

[dev]
aws_access_key_id = AKIA456
aws_secret_access_key = secret
`,
			wantProfiles: []string{"default", "dev", "prod"},
			wantRegions: map[string]string{
				"default": "us-east-1",
				"dev":     "ap-northeast-1",
				"prod":    "us-west-2",
			},
			wantErr: false,
		},
		{
			name: "credentials only profiles",
			configContent: `
[default]
region = us-east-1
`,
			credsContent: `
[default]
aws_access_key_id = AKIA123
aws_secret_access_key = secret

[staging]
aws_access_key_id = AKIA789
aws_secret_access_key = secret
`,
			wantProfiles: []string{"default", "staging"},
			wantRegions: map[string]string{
				"default": "us-east-1",
				"staging": "", // no region in config
			},
			wantErr: false,
		},
		{
			name: "config with comments and empty lines",
			configContent: `
# This is a comment
[default]
region = us-east-1

; Another comment style
[profile dev]
# inline comment won't work but this is separate line
region = eu-west-1
`,
			credsContent: "",
			wantProfiles: []string{"default", "dev"},
			wantRegions: map[string]string{
				"default": "us-east-1",
				"dev":     "eu-west-1",
			},
			wantErr: false,
		},
		{
			name: "profile without region",
			configContent: `
[default]
output = json

[profile no-region]
output = table
`,
			credsContent: "",
			wantProfiles: []string{"default", "no-region"},
			wantRegions: map[string]string{
				"default":   "",
				"no-region": "",
			},
			wantErr: false,
		},
		{
			name:          "empty files",
			configContent: "",
			credsContent:  "",
			wantProfiles:  []string{},
			wantRegions:   map[string]string{},
			wantErr:       false,
		},
		{
			name: "whitespace handling",
			configContent: `
[  default  ]
  region   =   ap-southeast-1  

[ profile   spaced  ]
region=no-spaces
`,
			credsContent: "",
			wantProfiles: []string{"default", "spaced"},
			wantRegions: map[string]string{
				"default": "ap-southeast-1",
				"spaced":  "no-spaces",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config")
			credsPath := filepath.Join(tmpDir, "credentials")

			// Write test files
			if tt.configContent != "" {
				if err := os.WriteFile(configPath, []byte(tt.configContent), 0600); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
			}
			if tt.credsContent != "" {
				if err := os.WriteFile(credsPath, []byte(tt.credsContent), 0600); err != nil {
					t.Fatalf("failed to write credentials file: %v", err)
				}
			}

			// Parse
			parser := profile.NewParserWithPaths(configPath, credsPath)
			list, err := parser.Parse()

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check profile count
			gotNames := list.Names()
			if len(gotNames) != len(tt.wantProfiles) {
				t.Errorf("got %d profiles, want %d. got: %v, want: %v",
					len(gotNames), len(tt.wantProfiles), gotNames, tt.wantProfiles)
				return
			}

			// Check each expected profile exists with correct region
			for _, wantName := range tt.wantProfiles {
				info, found := list.GetProfile(wantName)
				if !found {
					t.Errorf("profile %q not found", wantName)
					continue
				}
				wantRegion := tt.wantRegions[wantName]
				if info.Region != wantRegion {
					t.Errorf("profile %q region = %q, want %q", wantName, info.Region, wantRegion)
				}
			}
		})
	}
}

func TestParser_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent", "config")
	credsPath := filepath.Join(tmpDir, "nonexistent", "credentials")

	parser := profile.NewParserWithPaths(configPath, credsPath)
	list, err := parser.Parse()

	// Should not error on missing files
	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
		return
	}

	if list.HasProfiles() {
		t.Errorf("expected no profiles, got %v", list.Names())
	}
}

func TestList_GetProfile(t *testing.T) {
	list := &profile.List{
		Profiles: []profile.Info{
			{Name: "default", Region: "us-east-1"},
			{Name: "dev", Region: "ap-northeast-1"},
		},
		Default: "default",
	}

	t.Run("existing profile", func(t *testing.T) {
		info, found := list.GetProfile("dev")
		if !found {
			t.Error("expected to find profile 'dev'")
		}
		if info.Region != "ap-northeast-1" {
			t.Errorf("region = %q, want %q", info.Region, "ap-northeast-1")
		}
	})

	t.Run("non-existing profile", func(t *testing.T) {
		_, found := list.GetProfile("nonexistent")
		if found {
			t.Error("expected not to find profile 'nonexistent'")
		}
	})
}

func TestList_Names(t *testing.T) {
	list := &profile.List{
		Profiles: []profile.Info{
			{Name: "default"},
			{Name: "dev"},
			{Name: "prod"},
		},
	}

	names := list.Names()
	if len(names) != 3 {
		t.Errorf("got %d names, want 3", len(names))
	}

	expected := []string{"default", "dev", "prod"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("names[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestList_HasProfiles(t *testing.T) {
	t.Run("with profiles", func(t *testing.T) {
		list := &profile.List{
			Profiles: []profile.Info{{Name: "default"}},
		}
		if !list.HasProfiles() {
			t.Error("expected HasProfiles() = true")
		}
	})

	t.Run("empty", func(t *testing.T) {
		list := &profile.List{}
		if list.HasProfiles() {
			t.Error("expected HasProfiles() = false")
		}
	})
}

func TestParser_EnvVarOverride(t *testing.T) {
	// Create temp files
	tmpDir := t.TempDir()
	customConfigPath := filepath.Join(tmpDir, "custom-config")
	customCredsPath := filepath.Join(tmpDir, "custom-credentials")

	configContent := `
[default]
region = custom-region
`
	credsContent := `
[default]
aws_access_key_id = AKIA123
`

	if err := os.WriteFile(customConfigPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	if err := os.WriteFile(customCredsPath, []byte(credsContent), 0600); err != nil {
		t.Fatalf("failed to write credentials: %v", err)
	}

	// Set environment variables
	t.Setenv("AWS_CONFIG_FILE", customConfigPath)
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", customCredsPath)

	parser, err := profile.NewParser()
	if err != nil {
		t.Fatalf("NewParser() error: %v", err)
	}

	// Verify paths
	if parser.GetConfigPath() != customConfigPath {
		t.Errorf("config path = %q, want %q", parser.GetConfigPath(), customConfigPath)
	}
	if parser.GetCredentialsPath() != customCredsPath {
		t.Errorf("credentials path = %q, want %q", parser.GetCredentialsPath(), customCredsPath)
	}

	// Parse and verify
	list, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	info, found := list.GetProfile("default")
	if !found {
		t.Fatal("expected to find 'default' profile")
	}
	if info.Region != "custom-region" {
		t.Errorf("region = %q, want %q", info.Region, "custom-region")
	}
}

