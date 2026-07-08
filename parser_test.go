package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name      string
		input     string
		output    string
		shouldErr bool
	}{
		{
			name:      "Already an absolute path",
			input:     "/home/cooluser/.zshrc",
			output:    "/home/cooluser/.zshrc",
			shouldErr: false,
		},
		{
			name:      "Full path && Not a user directory",
			input:     "/var/log/syslog",
			output:    "/var/log/syslog",
			shouldErr: false,
		},
		{
			name:      "Relative path",
			input:     "~/.vimrc",
			output:    filepath.Join(home, ".vimrc"),
			shouldErr: false,
		},
		{
			name:      "Empty path",
			input:     "",
			output:    "",
			shouldErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ResolvePath(test.input)
			if (err != nil) != test.shouldErr {
				t.Fatalf("ResolvePath(%q): expected [%v] but got [%v]", test.input, err, test.shouldErr)
			}

			if got != test.output {
				t.Errorf("ResolvePath(%q): expected [%v] but got [%v]", test.input, test.output, got)
			}
		})
	}
}

func TestParseConfigFile_Valid(t *testing.T) {
	validTests := []struct {
		name        string
		jsonContent string
		expectedLen int
	}{
		{
			name:        "Correct Single Job Object",
			jsonContent: `[{"source_data": "/etc", "backup_dest": "etc_bak", "storage_type": "archive"}]`,
			expectedLen: 1,
		},
		{
			name: "Correct Multiple Job Objects",
			jsonContent: `[
				{"source_data": "/etc", "backup_dest": "etc_bak", "storage_type": "archive"},
				{"source_data": "/home", "backup_dest": "home_bak", "storage_type": "secure"}
			]`,
			expectedLen: 2,
		},
	}

	t.Run("Valid Configs", func(t *testing.T) {
		for _, tt := range validTests {
			t.Run(tt.name, func(t *testing.T) {
				tmpFile, err := os.CreateTemp("", "valid_*.json")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer func() {
					err = os.Remove(tmpFile.Name())
					if err != nil {
						t.Fatalf("Failed to remove temp file: %v", err)
					}
				}()

				_, err = tmpFile.WriteString(tt.jsonContent)
				if err != nil {
					t.Fatalf("Failed to write temp file: %v", err)
				}
				CloseSafe(tmpFile)

				jobs, err := ParseConfigFile(tmpFile.Name())
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}

				if len(jobs) != tt.expectedLen {
					t.Errorf("Expected %d jobs, got %d", tt.expectedLen, len(jobs))
				}
			})
		}
	})
}

func TestParseConfigFile_Invalid(t *testing.T) {
	invalidTests := []struct {
		name        string
		jsonContent string
	}{
		{
			name:        "Broken JSON syntax (no closing quote)",
			jsonContent: `[{"source_data": "/etc"`,
		},
		{
			name:        "Unknown field (Strict mode)",
			jsonContent: `[{"source_data": "/etc", "unknown_field_xyz": "value"}]`,
		},
		{
			name:        "Plain text instead of JSON",
			jsonContent: `Hello world, I am not a JSON at all!`,
		},
		{
			name:        "Valid JSON object but not an array (no quotes)",
			jsonContent: `{"source_data": "/etc", "backup_dest": "etc_bak"}`,
		},
	}

	t.Run("Invalid Configs", func(t *testing.T) {
		for _, tt := range invalidTests {
			t.Run(tt.name, func(t *testing.T) {
				tmpFile, err := os.CreateTemp("", "invalid_*.json")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer func() {
					err = os.Remove(tmpFile.Name())
					if err != nil {
						t.Fatalf("Failed to remove temp file: %v", err)
					}
				}()

				_, err = tmpFile.WriteString(tt.jsonContent)
				if err != nil {
					t.Fatalf("Failed to write to temp file: %v", err)
				}

				CloseSafe(tmpFile)

				_, err = ParseConfigFile(tmpFile.Name())

				if err == nil {
					t.Errorf("Expected error for case '%s', but got nil", tt.name)
				}
			})
		}
	})
}
