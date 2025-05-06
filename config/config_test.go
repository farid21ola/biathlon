package config

import (
	"os"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	// Create test config file
	const testFile = "test_config.json"
	const testData = `{
		"laps": 3,
		"lapLen": 4000,
		"penaltyLen": 150,
		"firingLines": 2,
		"start": "10:00:00",
		"startDelta": "00:01:00"
	}`

	os.WriteFile(testFile, []byte(testData), 0644)
	defer os.Remove(testFile)

	// Test valid config
	cfg, err := LoadFromFile(testFile)
	if err != nil {
		t.Fatalf("Failed to load valid config: %v", err)
	}

	if cfg.Laps != 3 || cfg.LapLen != 4000 {
		t.Errorf("Config parsing mismatch. Got %+v", cfg)
	}

	// Test invalid file
	_, err = LoadFromFile("nonexistent.json")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}
