package race

import (
	"biathlon/config"
	"biathlon/event"
	"biathlon/model"
	"testing"
)

func TestRegistration(t *testing.T) {
	cfg := &config.Config{Laps: 2}
	ctrl := NewController(cfg)

	evt := event.Event{
		Time:         "10:00:00.000",
		EventID:      event.Registered,
		CompetitorID: 1,
	}

	ctrl.processEvent(evt)

	if _, exists := ctrl.Competitors[1]; !exists {
		t.Fatal("Competitor not registered")
	}
}

func TestStartTimeSet(t *testing.T) {
	cfg := &config.Config{Laps: 2}
	ctrl := NewController(cfg)

	evt := event.Event{
		Time:         "10:00:00.000",
		EventID:      event.Registered,
		CompetitorID: 1,
	}
	ctrl.processEvent(evt)

	evt = event.Event{
		Time:         "10:01:00.000",
		EventID:      event.StartTimeSet,
		CompetitorID: 1,
		ExtraParams:  "10:05:00.000",
	}
	ctrl.processEvent(evt)

	if ctrl.Competitors[1].PlannedStartTime != "10:05:00.000" {
		t.Error("Start time not set")
	}
}

func TestFullRaceFlow(t *testing.T) {
	cfg := &config.Config{
		Laps:        2,
		LapLen:      4000,
		PenaltyLen:  150,
		FiringLines: 1,
	}
	ctrl := NewController(cfg)

	events := []event.Event{
		{Time: "10:00:00.000", EventID: event.Registered, CompetitorID: 1},
		{Time: "10:01:00.000", EventID: event.Registered, CompetitorID: 2},
		{Time: "10:05:00.000", EventID: event.StartTimeSet, CompetitorID: 1, ExtraParams: "10:10:00.000"},
		{Time: "10:06:00.000", EventID: event.StartTimeSet, CompetitorID: 2, ExtraParams: "10:15:00.000"},
		{Time: "10:09:55.000", EventID: event.OnStartLine, CompetitorID: 1},
		{Time: "10:10:01.000", EventID: event.Started, CompetitorID: 1},
		{Time: "10:11:55.000", EventID: event.OnStartLine, CompetitorID: 2},
		{Time: "10:18:01.000", EventID: event.Started, CompetitorID: 2},
		{Time: "10:19:00.000", EventID: event.Disqualified, CompetitorID: 2},
		{Time: "10:20:00.000", EventID: event.OnFiringRange, CompetitorID: 1, ExtraParams: "1"},
		{Time: "10:20:01.000", EventID: event.TargetHit, CompetitorID: 1, ExtraParams: "3"},
		{Time: "10:20:03.000", EventID: event.LeftFiringRange, CompetitorID: 1},
		{Time: "10:25:00.000", EventID: event.EndedLap, CompetitorID: 1},
		{Time: "10:35:00.000", EventID: event.EndedLap, CompetitorID: 1},
	}

	_, err := ctrl.ProcessEvents(events)
	if err != nil {
		t.Fatal(err)
	}

	// Проверка финиша
	comp := ctrl.Competitors[1]
	if comp.Status != "Finished" {
		t.Errorf("Expected Finished status for comp1, got %s", comp.Status)
	}
	comp2 := ctrl.Competitors[2]
	if comp2.Status != "NotStarted" {
		t.Errorf("Expected NotStarted status for comp2, got %s", comp.Status)
	}
	// Проверка скорости первого круга
	if comp.LapTimes[0].Speed < 2.0 {
		t.Error("Invalid lap speed calculation")
	}
}

func TestPenaltyCalculation(t *testing.T) {
	cfg := &config.Config{PenaltyLen: 150}
	ctrl := NewController(cfg)

	// Участник с 2 попаданиями из 5 выстрелов
	ctrl.Competitors[1] = &model.Competitor{
		CurrentLap:       1,
		HitsCount:        2,
		ShotsCount:       5,
		PenaltyStartTime: "10:00:00.000",
	}

	// Событие выхода с штрафных кругов
	evt := event.Event{
		Time:         "10:05:00.000",
		EventID:      event.LeftPenalty,
		CompetitorID: 1,
	}
	ctrl.processEvent(evt)

	comp := ctrl.Competitors[1]
	expectedDistance := float64(3 * 150)      // 3 промаха
	expectedSpeed := expectedDistance / 300.0 // 5 минут = 300 сек

	if comp.PenaltySpeed != expectedSpeed {
		t.Errorf("Expected speed %.2f, got %.2f", expectedSpeed, comp.PenaltySpeed)
	}
}

func TestDisqualification(t *testing.T) {
	cfg := &config.Config{}
	ctrl := NewController(cfg)

	// Участник с установленным временем старта, но без события Started
	ctrl.Competitors[1] = &model.Competitor{
		PlannedStartTime: "10:00:00.000",
	}

	ctrl.ProcessEvents([]event.Event{}) // Запуск пост-обработки

	if ctrl.Competitors[1].Status != "NotStarted" {
		t.Error("Competitor should be disqualified")
	}
}
