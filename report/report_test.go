package report

import (
	"biathlon/config"
	"biathlon/model"
	"strings"
	"testing"
)

func TestReportSorting(t *testing.T) {
	cfg := &config.Config{Laps: 2}

	competitors := map[int]*model.Competitor{
		1: {ID: 1, Status: "Finished"},
		2: {ID: 2, Status: "NotStarted"},
		3: {ID: 3, Status: "Finished"},
		4: {ID: 4, Status: "NotFinished"},
	}

	report := GenerateFinalReport(competitors, cfg)

	// Проверяем порядок: Finished сначала, затем NotStarted
	real1, imposter1, real2, imposter2 := strings.Index(report, "1"), strings.Index(report, "2"),
		strings.Index(report, "3"), strings.Index(report, "4")
	if real1 == -1 || imposter1 == -1 || real2 == -1 || imposter2 == -1 {
		t.Error("Incorrect output")
	}
	if !(real1 < imposter1 && real2 < imposter1 && real1 < imposter2) {
		t.Error("Incorrect sorting order")
	}
}
