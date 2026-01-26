package main

import (
	"strings"
	"testing"

	"github.com/andriyg76/go-hbars/internal/compiler"
)

func TestParseImports(t *testing.T) {
	tests := []struct {
		name        string
		flags       importFlag
		wantMap     map[string]string
		wantDefault string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "empty flags",
			flags:       importFlag{},
			wantMap:     map[string]string{},
			wantDefault: "",
			wantErr:     false,
		},
		{
			name:        "default import only",
			flags:       importFlag{"github.com/example/helpers"},
			wantMap:     map[string]string{},
			wantDefault: "github.com/example/helpers",
			wantErr:     false,
		},
		{
			name: "aliased import",
			flags: importFlag{"extra:github.com/example/extra"},
			wantMap: map[string]string{
				"extra": "github.com/example/extra",
			},
			wantDefault: "",
			wantErr:     false,
		},
		{
			name: "multiple aliased imports",
			flags: importFlag{
				"extra:github.com/example/extra",
				"utils:github.com/example/utils",
			},
			wantMap: map[string]string{
				"extra": "github.com/example/extra",
				"utils": "github.com/example/utils",
			},
			wantDefault: "",
			wantErr:     false,
		},
		{
			name: "default and aliased imports",
			flags: importFlag{
				"github.com/example/helpers",
				"extra:github.com/example/extra",
			},
			wantMap: map[string]string{
				"extra": "github.com/example/extra",
			},
			wantDefault: "github.com/example/helpers",
			wantErr:     false,
		},
		{
			name:        "multiple default imports error",
			flags:       importFlag{"github.com/example/helpers", "github.com/example/other"},
			wantErr:     true,
			errMsg:      "multiple default imports",
		},
		{
			name:        "duplicate alias error",
			flags:       importFlag{"extra:github.com/example/extra", "extra:github.com/example/other"},
			wantErr:     true,
			errMsg:      "duplicate import alias",
		},
		{
			name:        "invalid format - empty alias",
			flags:       importFlag{":github.com/example/helpers"},
			wantErr:     true,
			errMsg:      "invalid import flag",
		},
		{
			name:        "invalid format - empty path",
			flags:       importFlag{"extra:"},
			wantErr:     true,
			errMsg:      "invalid import flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMap, gotDefault, err := parseImports(tt.flags)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseImports() expected error but got none")
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Fatalf("parseImports() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseImports() error = %v, want no error", err)
			}
			if gotDefault != tt.wantDefault {
				t.Errorf("parseImports() defaultImport = %q, want %q", gotDefault, tt.wantDefault)
			}
			if len(gotMap) != len(tt.wantMap) {
				t.Errorf("parseImports() importMap length = %d, want %d", len(gotMap), len(tt.wantMap))
			}
			for k, v := range tt.wantMap {
				if gotMap[k] != v {
					t.Errorf("parseImports() importMap[%q] = %q, want %q", k, gotMap[k], v)
				}
			}
		})
	}
}

func TestParseHelpersFlags(t *testing.T) {
	tests := []struct {
		name          string
		flags         helpersFlag
		importMap     map[string]string
		defaultImport string
		wantHelpers   map[string]compiler.HelperRef
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "empty flags",
			flags:         helpersFlag{},
			importMap:     map[string]string{},
			defaultImport: "",
			wantHelpers:   map[string]compiler.HelperRef{},
			wantErr:       false,
		},
		{
			name:          "single helper with default import",
			flags:         helpersFlag{"Upper"},
			importMap:     map[string]string{},
			defaultImport: "github.com/example/helpers",
			wantHelpers: map[string]compiler.HelperRef{
				"upper": {ImportPath: "github.com/example/helpers", Ident: "Upper"},
			},
			wantErr: false,
		},
		{
			name:          "multiple helpers with default import",
			flags:         helpersFlag{"Upper,Lower,Join"},
			importMap:     map[string]string{},
			defaultImport: "github.com/example/helpers",
			wantHelpers: map[string]compiler.HelperRef{
				"upper": {ImportPath: "github.com/example/helpers", Ident: "Upper"},
				"lower": {ImportPath: "github.com/example/helpers", Ident: "Lower"},
				"join":  {ImportPath: "github.com/example/helpers", Ident: "Join"},
			},
			wantErr: false,
		},
		{
			name:      "helper with alias",
			flags:     helpersFlag{"extra:Join"},
			importMap: map[string]string{"extra": "github.com/example/extra"},
			wantHelpers: map[string]compiler.HelperRef{
				"join": {ImportPath: "github.com/example/extra", Ident: "Join"},
			},
			wantErr: false,
		},
		{
			name:          "helper with explicit name",
			flags:         helpersFlag{"myJoin=Join"},
			importMap:     map[string]string{},
			defaultImport: "github.com/example/helpers",
			wantHelpers: map[string]compiler.HelperRef{
				"myJoin": {ImportPath: "github.com/example/helpers", Ident: "Join"},
			},
			wantErr: false,
		},
		{
			name:      "helper with alias and explicit name",
			flags:     helpersFlag{"extra:myJoin=Join"},
			importMap: map[string]string{"extra": "github.com/example/extra"},
			wantHelpers: map[string]compiler.HelperRef{
				"myJoin": {ImportPath: "github.com/example/extra", Ident: "Join"},
			},
			wantErr: false,
		},
		{
			name:          "multiple flag calls",
			flags:         helpersFlag{"Upper", "Lower"},
			importMap:     map[string]string{},
			defaultImport: "github.com/example/helpers",
			wantHelpers: map[string]compiler.HelperRef{
				"upper": {ImportPath: "github.com/example/helpers", Ident: "Upper"},
				"lower": {ImportPath: "github.com/example/helpers", Ident: "Lower"},
			},
			wantErr: false,
		},
		{
			name:          "no import available",
			flags:         helpersFlag{"Upper"},
			importMap:     map[string]string{},
			defaultImport: "",
			wantErr:       true,
			errMsg:        "requires an import alias or default import",
		},
		{
			name:      "unknown alias",
			flags:     helpersFlag{"unknown:Join"},
			importMap: map[string]string{"extra": "github.com/example/extra"},
			wantErr:   true,
			errMsg:    "unknown import alias",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers := make(map[string]compiler.HelperRef)
			err := parseHelpersFlags(tt.flags, tt.importMap, tt.defaultImport, helpers)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseHelpersFlags() expected error but got none")
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Fatalf("parseHelpersFlags() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseHelpersFlags() error = %v, want no error", err)
			}
			if len(helpers) != len(tt.wantHelpers) {
				t.Errorf("parseHelpersFlags() helpers length = %d, want %d", len(helpers), len(tt.wantHelpers))
			}
			for k, v := range tt.wantHelpers {
				got, exists := helpers[k]
				if !exists {
					t.Errorf("parseHelpersFlags() missing helper %q", k)
					continue
				}
				if got.ImportPath != v.ImportPath {
					t.Errorf("parseHelpersFlags() helpers[%q].ImportPath = %q, want %q", k, got.ImportPath, v.ImportPath)
				}
				if got.Ident != v.Ident {
					t.Errorf("parseHelpersFlags() helpers[%q].Ident = %q, want %q", k, got.Ident, v.Ident)
				}
			}
		})
	}
}

func TestBuildHelpers(t *testing.T) {
	tests := []struct {
		name            string
		noCoreHelpers   bool
		importFlags     importFlag
		helpersFlags    helpersFlag
		legacyFlags     helperFlag
		wantHelperCount int
		wantErr         bool
		checkHelpers    map[string]compiler.HelperRef
	}{
		{
			name:            "core helpers by default",
			noCoreHelpers:   false,
			importFlags:     importFlag{},
			helpersFlags:    helpersFlag{},
			legacyFlags:     helperFlag{},
			wantHelperCount: 57, // All helpers from registry
			wantErr:         false,
			checkHelpers: map[string]compiler.HelperRef{
				"upper": {ImportPath: "github.com/andriyg76/go-hbars/helpers/handlebars", Ident: "Upper"},
				"lower": {ImportPath: "github.com/andriyg76/go-hbars/helpers/handlebars", Ident: "Lower"},
			},
		},
		{
			name:            "no core helpers",
			noCoreHelpers:   true,
			importFlags:     importFlag{},
			helpersFlags:    helpersFlag{},
			legacyFlags:     helperFlag{},
			wantHelperCount: 0,
			wantErr:         false,
		},
		{
			name:            "core helpers + custom helpers",
			noCoreHelpers:   false,
			importFlags:     importFlag{"github.com/example/extra"},
			helpersFlags:    helpersFlag{"CustomHelper"},
			legacyFlags:     helperFlag{},
			wantHelperCount: 58, // 57 core + 1 custom
			wantErr:         false,
			checkHelpers: map[string]compiler.HelperRef{
				"upper":        {ImportPath: "github.com/andriyg76/go-hbars/helpers/handlebars", Ident: "Upper"},
				"customhelper": {ImportPath: "github.com/example/extra", Ident: "CustomHelper"},
			},
		},
		{
			name:            "legacy flag overrides core",
			noCoreHelpers:   false,
			importFlags:     importFlag{},
			helpersFlags:    helpersFlag{},
			legacyFlags:     helperFlag{"upper=github.com/example/custom:MyUpper"},
			wantHelperCount: 57, // Same count, but upper is overridden
			wantErr:         false,
			checkHelpers: map[string]compiler.HelperRef{
				"upper": {ImportPath: "github.com/example/custom", Ident: "MyUpper"},
			},
		},
		{
			name:            "helpers flag overrides core",
			noCoreHelpers:   false,
			importFlags:     importFlag{"github.com/example/custom"},
			helpersFlags:    helpersFlag{"upper=MyUpper"},
			legacyFlags:     helperFlag{},
			wantHelperCount: 57, // Same count, but upper is overridden
			wantErr:         false,
			checkHelpers: map[string]compiler.HelperRef{
				"upper": {ImportPath: "github.com/example/custom", Ident: "MyUpper"},
			},
		},
		{
			name:            "legacy flag overrides helpers flag",
			noCoreHelpers:   false,
			importFlags:     importFlag{"github.com/example/custom"},
			helpersFlags:    helpersFlag{"upper=MyUpper"},
			legacyFlags:     helperFlag{"upper=github.com/example/legacy:LegacyUpper"},
			wantHelperCount: 57,
			wantErr:         false,
			checkHelpers: map[string]compiler.HelperRef{
				"upper": {ImportPath: "github.com/example/legacy", Ident: "LegacyUpper"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers, err := buildHelpers(tt.noCoreHelpers, tt.importFlags, tt.helpersFlags, tt.legacyFlags)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("buildHelpers() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("buildHelpers() error = %v, want no error", err)
			}
			if len(helpers) != tt.wantHelperCount {
				t.Errorf("buildHelpers() helper count = %d, want %d", len(helpers), tt.wantHelperCount)
			}
			for k, v := range tt.checkHelpers {
				got, exists := helpers[k]
				if !exists {
					t.Errorf("buildHelpers() missing helper %q", k)
					continue
				}
				if got.ImportPath != v.ImportPath {
					t.Errorf("buildHelpers() helpers[%q].ImportPath = %q, want %q", k, got.ImportPath, v.ImportPath)
				}
				if got.Ident != v.Ident {
					t.Errorf("buildHelpers() helpers[%q].Ident = %q, want %q", k, got.Ident, v.Ident)
				}
			}
		})
	}
}

func TestParseLegacyHelpers(t *testing.T) {
	tests := []struct {
		name        string
		flags       helperFlag
		wantHelpers map[string]compiler.HelperRef
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "empty flags",
			flags:       helperFlag{},
			wantHelpers: map[string]compiler.HelperRef{},
			wantErr:     false,
		},
		{
			name:  "simple helper",
			flags: helperFlag{"upper=Upper"},
			wantHelpers: map[string]compiler.HelperRef{
				"upper": {ImportPath: "", Ident: "Upper"},
			},
			wantErr: false,
		},
		{
			name:  "helper with import",
			flags: helperFlag{"upper=github.com/example/helpers:Upper"},
			wantHelpers: map[string]compiler.HelperRef{
				"upper": {ImportPath: "github.com/example/helpers", Ident: "Upper"},
			},
			wantErr: false,
		},
		{
			name:        "invalid format - no equals",
			flags:       helperFlag{"upper"},
			wantErr:     true,
			errMsg:      "invalid helper mapping",
		},
		{
			name:        "invalid format - empty name",
			flags:       helperFlag{"=Upper"},
			wantErr:     true,
			errMsg:      "invalid helper mapping",
		},
		{
			name:        "invalid format - empty ref",
			flags:       helperFlag{"upper="},
			wantErr:     true,
			errMsg:      "invalid helper mapping",
		},
		{
			name:        "invalid format - empty import path",
			flags:       helperFlag{"upper=:Upper"},
			wantErr:     true,
			errMsg:      "invalid helper mapping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers := make(map[string]compiler.HelperRef)
			err := parseLegacyHelpers(tt.flags, helpers)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseLegacyHelpers() expected error but got none")
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Fatalf("parseLegacyHelpers() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseLegacyHelpers() error = %v, want no error", err)
			}
			if len(helpers) != len(tt.wantHelpers) {
				t.Errorf("parseLegacyHelpers() helpers length = %d, want %d", len(helpers), len(tt.wantHelpers))
			}
			for k, v := range tt.wantHelpers {
				got, exists := helpers[k]
				if !exists {
					t.Errorf("parseLegacyHelpers() missing helper %q", k)
					continue
				}
				if got.ImportPath != v.ImportPath {
					t.Errorf("parseLegacyHelpers() helpers[%q].ImportPath = %q, want %q", k, got.ImportPath, v.ImportPath)
				}
				if got.Ident != v.Ident {
					t.Errorf("parseLegacyHelpers() helpers[%q].Ident = %q, want %q", k, got.Ident, v.Ident)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

