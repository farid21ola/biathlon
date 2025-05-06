package model

import (
	"fmt"
	"time"
)

const TimeFormat = "15:04:05.000"

type LapInfo struct {
	Time    string
	Speed   float64
	EndTime string
}

type Competitor struct {
	ID                int
	RegisterTime      string
	PlannedStartTime  string
	ActualStartTime   string
	PenaltyStartTime  string
	PenaltyDuration   time.Duration
	PenaltySpeed      float64
	EndTime           string
	Status            string // "Finished", "NotStarted", "NotFinished"
	LapTimes          []LapInfo
	HitsCount         int
	ShotsCount        int
	FiringRangeVisits map[int]bool
	IsOnFiringRange   bool
	CurrentLap        int
	LastEvent         time.Time
	CannotContinue    string
}

func NewCompetitor(id int, registerTime string, laps int) *Competitor {
	return &Competitor{
		ID:                id,
		RegisterTime:      registerTime,
		Status:            "",
		LapTimes:          make([]LapInfo, laps),
		FiringRangeVisits: make(map[int]bool),
		CurrentLap:        1,
	}
}

func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse(TimeFormat, timeStr)
}

func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := int(d.Milliseconds()) % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}
