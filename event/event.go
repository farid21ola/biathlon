package event

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"biathlon/model"
)

const (
	Registered      = 1
	StartTimeSet    = 2
	OnStartLine     = 3
	Started         = 4
	OnFiringRange   = 5
	TargetHit       = 6
	LeftFiringRange = 7
	EnteredPenalty  = 8
	LeftPenalty     = 9
	EndedLap        = 10
	CannotContinue  = 11
	Disqualified    = 32
	Finished        = 33
)

type Event struct {
	Time         string
	EventID      int
	CompetitorID int
	ExtraParams  string
}

func (e Event) ParsedTime() (time.Time, error) {
	return model.ParseTime(e.Time)
}

func Parse(line string) (Event, error) {
	timeStart := strings.Index(line, "[")
	timeEnd := strings.Index(line, "]")
	if timeStart == -1 || timeEnd == -1 || timeStart >= timeEnd {
		return Event{}, fmt.Errorf("invalid event format: %s", line)
	}

	timeStr := line[timeStart+1 : timeEnd]
	if !isTimeValid(timeStr) {
		return Event{}, fmt.Errorf("invalid event time format: %s", timeStr)
	}

	parts := strings.Fields(line[timeEnd+1:])
	if len(parts) < 2 {
		return Event{}, fmt.Errorf("not enough event parts: %s", line)
	}

	eventID, err := strconv.Atoi(parts[0])
	if err != nil {
		return Event{}, fmt.Errorf("invalid event ID: %s", parts[0])
	}

	competitorID, err := strconv.Atoi(parts[1])
	if err != nil {
		return Event{}, fmt.Errorf("invalid competitor ID: %s", parts[1])
	}

	extraParams := ""
	if len(parts) > 2 {
		extraParams = strings.Join(parts[2:], " ")
	}

	return Event{
		Time:         timeStr,
		EventID:      eventID,
		CompetitorID: competitorID,
		ExtraParams:  extraParams,
	}, nil
}

func LoadFromFile(path string) ([]Event, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	events := make([]Event, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		event, err := Parse(line)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

func FormatLogEntry(event Event) string {
	switch event.EventID {
	case Registered:
		return fmt.Sprintf("The competitor(%d) registered", event.CompetitorID)
	case StartTimeSet:
		return fmt.Sprintf("The start time for the competitor(%d) was set by a draw to %s", event.CompetitorID, event.ExtraParams)
	case OnStartLine:
		return fmt.Sprintf("The competitor(%d) is on the start line", event.CompetitorID)
	case Started:
		return fmt.Sprintf("The competitor(%d) has started", event.CompetitorID)
	case OnFiringRange:
		firingRange, _ := strconv.Atoi(event.ExtraParams)
		return fmt.Sprintf("The competitor(%d) is on the firing range(%d)", event.CompetitorID, firingRange)
	case TargetHit:
		target, _ := strconv.Atoi(event.ExtraParams)
		return fmt.Sprintf("The target(%d) has been hit by competitor(%d)", target, event.CompetitorID)
	case LeftFiringRange:
		return fmt.Sprintf("The competitor(%d) left the firing range", event.CompetitorID)
	case EnteredPenalty:
		return fmt.Sprintf("The competitor(%d) entered the penalty laps", event.CompetitorID)
	case LeftPenalty:
		return fmt.Sprintf("The competitor(%d) left the penalty laps", event.CompetitorID)
	case EndedLap:
		return fmt.Sprintf("The competitor(%d) ended the main lap", event.CompetitorID)
	case CannotContinue:
		return fmt.Sprintf("The competitor(%d) can`t continue: %s", event.CompetitorID, event.ExtraParams)
	case Disqualified:
		return fmt.Sprintf("The competitor(%d) is disqualified", event.CompetitorID)
	case Finished:
		return fmt.Sprintf("The competitor(%d) has finished", event.CompetitorID)
	default:
		return fmt.Sprintf("Unknown event: %d for competitor(%d) with params: %s", event.EventID, event.CompetitorID, event.ExtraParams)
	}
}

func isTimeValid(timeStr string) bool {
	if len(timeStr) != 12 {
		return false
	}
	_, err := time.Parse("15:04:05.000", timeStr)
	return err == nil
}
