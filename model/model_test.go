package model

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	_, err := ParseTime("invalid")
	if err == nil {
		t.Error("Expected error for invalid time")
	}

	tm, err := ParseTime("12:34:56.789")
	if err != nil {
		t.Fatal(err)
	}

	if tm.Hour() != 12 || tm.Second() != 56 {
		t.Errorf("Time parsing mismatch: %v", tm)
	}
}

func TestFormatDuration(t *testing.T) {
	d := 2*time.Hour + 30*time.Minute + 5*time.Second + 123456789
	formatted := FormatDuration(d)
	expected := "02:30:05.123"

	if formatted != expected {
		t.Errorf("FormatDuration() = %v, want %v", formatted, expected)
	}
}
