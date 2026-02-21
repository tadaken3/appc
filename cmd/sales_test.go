package cmd

import (
	"testing"
)

func TestDateRange(t *testing.T) {
	dates, err := dateRange("2025-01-01", "2025-01-03")
	if err != nil {
		t.Fatalf("dateRange() error: %v", err)
	}
	if len(dates) != 3 {
		t.Fatalf("len = %d, want 3", len(dates))
	}
	if dates[0] != "2025-01-01" {
		t.Errorf("dates[0] = %q, want 2025-01-01", dates[0])
	}
	if dates[1] != "2025-01-02" {
		t.Errorf("dates[1] = %q, want 2025-01-02", dates[1])
	}
	if dates[2] != "2025-01-03" {
		t.Errorf("dates[2] = %q, want 2025-01-03", dates[2])
	}
}

func TestDateRangeSingleDay(t *testing.T) {
	dates, err := dateRange("2025-01-15", "2025-01-15")
	if err != nil {
		t.Fatalf("dateRange() error: %v", err)
	}
	if len(dates) != 1 {
		t.Fatalf("len = %d, want 1", len(dates))
	}
	if dates[0] != "2025-01-15" {
		t.Errorf("dates[0] = %q, want 2025-01-15", dates[0])
	}
}

func TestDateRangeInvalidFrom(t *testing.T) {
	_, err := dateRange("invalid", "2025-01-03")
	if err == nil {
		t.Fatal("dateRange() should return error for invalid --from")
	}
}

func TestDateRangeInvalidTo(t *testing.T) {
	_, err := dateRange("2025-01-01", "invalid")
	if err == nil {
		t.Fatal("dateRange() should return error for invalid --to")
	}
}

func TestDateRangeToBeforeFrom(t *testing.T) {
	_, err := dateRange("2025-01-10", "2025-01-01")
	if err == nil {
		t.Fatal("dateRange() should return error when --to is before --from")
	}
}
