package event

import (
	"testing"
)

func TestParseEvent(t *testing.T) {
	tests := []struct {
		input string
		want  Event
		err   bool
	}{
		{
			"[10:00:00.000] 1 123",
			Event{"10:00:00.000", 1, 123, ""},
			false,
		},
		{
			"[invalid] 2 456 extra",
			Event{},
			true,
		},
	}

	for _, tt := range tests {
		got, err := Parse(tt.input)
		if (err != nil) != tt.err {
			t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.err)
			continue
		}
		if !tt.err && got != tt.want {
			t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParsedTime(t *testing.T) {
	evt := Event{Time: "10:05:30.500"}
	pt, err := evt.ParsedTime()
	if err != nil {
		t.Fatal(err)
	}

	if pt.Hour() != 10 || pt.Minute() != 5 || pt.Second() != 30 || pt.Nanosecond() != 500000000 {
		t.Errorf("Parsed time mismatch: %v", pt)
	}
}
