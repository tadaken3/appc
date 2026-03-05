package validation

import (
	"testing"
)

func TestAppID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid numeric", "123456789", false},
		{"valid single digit", "1", false},
		{"empty", "", true},
		{"letters", "abc", true},
		{"mixed", "123abc", true},
		{"with spaces", "123 456", true},
		{"with special chars", "123?fields=name", true},
		{"with newline", "123\n456", true},
		{"with null byte", "123\x00456", true},
		{"comma separated (invalid single)", "123,456", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AppID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("AppID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestAppIDs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"single valid", "123", false},
		{"multiple valid", "123,456,789", false},
		{"with spaces", "123, 456", false},
		{"one invalid", "123,abc", true},
		{"empty string", "", true},
		{"only comma", ",", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AppIDs(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("AppIDs(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestCountryCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"valid lowercase", "jp", false},
		{"valid uppercase", "US", false},
		{"valid mixed", "Jp", false},
		{"empty", "", true},
		{"too short", "j", true},
		{"too long", "jpn", true},
		{"digits", "12", true},
		{"special chars", "j!", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CountryCode(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("CountryCode(%q) error = %v, wantErr %v", tt.code, err, tt.wantErr)
			}
		})
	}
}

func TestDateString(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{"valid date", "2024-01-15", false},
		{"valid month", "2024-01", false},
		{"empty", "", true},
		{"invalid format", "2024/01/15", true},
		{"with injection", "2024-01-15; rm -rf /", true},
		{"too long", "2024-01-15T00:00:00", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DateString(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("DateString(%q) error = %v, wantErr %v", tt.date, err, tt.wantErr)
			}
		})
	}
}

func TestFilePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid absolute", "/path/to/key.p8", false},
		{"valid home", "~/key.p8", false},
		{"valid relative", "key.p8", false},
		{"empty", "", true},
		{"path traversal", "/path/../../../etc/passwd", true},
		{"null byte", "/path/to/\x00key.p8", true},
		{"control chars", "/path/to/\x01key.p8", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}
