package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

type testRecord struct {
	Name  string `json:"name" csv:"name"`
	Score int    `json:"score" csv:"score"`
	Note  string `json:"note" csv:"note"`
}

func (r testRecord) CSVHeaders() []string {
	return []string{"name", "score", "note"}
}

func (r testRecord) CSVRow() []string {
	return []string{r.Name, intToStr(r.Score), r.Note}
}

func TestFormatJSON(t *testing.T) {
	records := []testRecord{
		{Name: "Alice", Score: 90, Note: "good"},
		{Name: "Bob", Score: 85, Note: "ok"},
	}

	var buf bytes.Buffer
	if err := FormatJSON(&buf, records); err != nil {
		t.Fatalf("FormatJSON() error: %v", err)
	}

	var parsed []testRecord
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(parsed) != 2 {
		t.Errorf("len = %d, want 2", len(parsed))
	}
	if parsed[0].Name != "Alice" {
		t.Errorf("name = %q, want Alice", parsed[0].Name)
	}
}

func TestFormatCSV(t *testing.T) {
	records := []testRecord{
		{Name: "Alice", Score: 90, Note: "good"},
		{Name: "Bob", Score: 85, Note: "has, comma"},
	}

	var buf bytes.Buffer
	if err := FormatCSV(&buf, records); err != nil {
		t.Fatalf("FormatCSV() error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3 (header + 2 rows)", len(lines))
	}
	if lines[0] != "name,score,note" {
		t.Errorf("header = %q, want %q", lines[0], "name,score,note")
	}
	if lines[1] != "Alice,90,good" {
		t.Errorf("row1 = %q, want %q", lines[1], "Alice,90,good")
	}
	// CSV should quote fields containing commas
	if !strings.Contains(lines[2], `"has, comma"`) {
		t.Errorf("row2 = %q, want comma-containing field to be quoted", lines[2])
	}
}

func TestFormatEmptyRecords(t *testing.T) {
	var records []testRecord
	var buf bytes.Buffer

	if err := FormatJSON(&buf, records); err != nil {
		t.Fatalf("FormatJSON() error: %v", err)
	}
	if buf.String() != "[]\n" {
		t.Errorf("empty JSON = %q, want []", buf.String())
	}

	buf.Reset()
	if err := FormatCSV(&buf, records); err != nil {
		t.Fatalf("FormatCSV() error: %v", err)
	}
	// Empty CSV should still have no output (no headers if no records)
	if buf.String() != "" {
		t.Errorf("empty CSV = %q, want empty", buf.String())
	}
}

func TestWriteOutput(t *testing.T) {
	records := []testRecord{{Name: "Test", Score: 1, Note: ""}}

	tests := []struct {
		name   string
		format string
	}{
		{"json", "json"},
		{"csv", "csv"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := Write(&buf, tt.format, records); err != nil {
				t.Fatalf("Write(%s) error: %v", tt.format, err)
			}
			if buf.Len() == 0 {
				t.Errorf("Write(%s) produced empty output", tt.format)
			}
		})
	}
}

func TestWriteInvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, "xml", []testRecord{{Name: "Test"}})
	if err == nil {
		t.Fatal("Write(xml) should return error")
	}
}
