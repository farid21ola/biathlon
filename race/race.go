package race

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"biathlon/config"
	"biathlon/event"
	"biathlon/model"
	"biathlon/report"
)

type Controller struct {
	Config      *config.Config
	Competitors map[int]*model.Competitor
	OutputLog   []string
}

func NewController(cfg *config.Config) *Controller {
	return &Controller{
		Config:      cfg,
		Competitors: make(map[int]*model.Competitor),
		OutputLog:   []string{},
	}
}

func (c *Controller) ProcessEvents(events []event.Event) (string, error) {
	for _, evt := range events {
		if err := c.processEvent(evt); err != nil {
			return "", err
		}
	}

	for id, competitor := range c.Competitors {
		if competitor.PlannedStartTime != "" && competitor.ActualStartTime == "" {
			c.disqualifyCompetitor(id, nil)
		}
	}

	return strings.Join(c.OutputLog, "\n"), nil
}

func (c *Controller) processEvent(evt event.Event) error {
	logEntry := event.FormatLogEntry(evt)
	c.OutputLog = append(c.OutputLog, fmt.Sprintf("[%s] %s", evt.Time, logEntry))

	switch evt.EventID {
	case event.Registered:
		c.registerCompetitor(evt)
	case event.StartTimeSet:
		c.setCompetitorStartTime(evt)
	case event.OnStartLine:
		// Просто логируем, особой обработки не требуется
	case event.Started:
		c.startCompetitor(evt)
	case event.OnFiringRange:
		firingRange, _ := strconv.Atoi(evt.ExtraParams)
		c.competitorOnFiringRange(evt.CompetitorID, firingRange)
	case event.TargetHit:
		target, _ := strconv.Atoi(evt.ExtraParams)
		c.targetHit(evt.CompetitorID, target)
	case event.LeftFiringRange:
		c.competitorLeftFiringRange(evt.CompetitorID)
	case event.EnteredPenalty:
		c.competitorEnteredPenaltyLaps(evt.CompetitorID, evt.Time)
	case event.LeftPenalty:
		c.competitorLeftPenaltyLaps(evt.CompetitorID, evt.Time)
	case event.EndedLap:
		c.competitorEndedMainLap(evt.CompetitorID, evt.Time)
	case event.CannotContinue:
		c.competitorCannotContinue(evt.CompetitorID, evt.ExtraParams)
	case event.Disqualified:
		c.disqualifyCompetitor(evt.CompetitorID, &evt.Time)
	}

	return nil
}

func (c *Controller) registerCompetitor(evt event.Event) {
	c.Competitors[evt.CompetitorID] = model.NewCompetitor(evt.CompetitorID, evt.Time, c.Config.Laps)
}

func (c *Controller) setCompetitorStartTime(evt event.Event) {
	if competitor, exists := c.Competitors[evt.CompetitorID]; exists {
		competitor.PlannedStartTime = evt.ExtraParams
	}
}

func (c *Controller) startCompetitor(evt event.Event) {
	if competitor, exists := c.Competitors[evt.CompetitorID]; exists {
		competitor.ActualStartTime = evt.Time
	}
}

func (c *Controller) competitorOnFiringRange(competitorID, firingRange int) {
	if competitor, exists := c.Competitors[competitorID]; exists {
		competitor.IsOnFiringRange = true
		competitor.FiringRangeVisits[firingRange] = true
	}
}

func (c *Controller) targetHit(competitorID, target int) {
	if competitor, exists := c.Competitors[competitorID]; exists && competitor.IsOnFiringRange {
		competitor.HitsCount++
	}
}

func (c *Controller) competitorLeftFiringRange(competitorID int) {
	if competitor, exists := c.Competitors[competitorID]; exists {
		competitor.IsOnFiringRange = false
		competitor.ShotsCount += 5
	}
}

func (c *Controller) competitorEnteredPenaltyLaps(competitorID int, timeStr string) {
	if competitor, exists := c.Competitors[competitorID]; exists {
		competitor.PenaltyStartTime = timeStr
	}
}

func (c *Controller) competitorLeftPenaltyLaps(competitorID int, timeStr string) {
	if competitor, exists := c.Competitors[competitorID]; exists {
		eventTime, _ := model.ParseTime(timeStr)
		penaltyStart, _ := model.ParseTime(competitor.PenaltyStartTime)

		competitor.PenaltyDuration += eventTime.Sub(penaltyStart)

		penaltyDistance := float64(5*competitor.CurrentLap-competitor.HitsCount) * float64(c.Config.PenaltyLen)
		speed := 0.0
		if competitor.PenaltyDuration.Seconds() > 0 {
			speed = penaltyDistance / competitor.PenaltyDuration.Seconds()
		}

		competitor.PenaltySpeed = speed
	}
}

func (c *Controller) competitorEndedMainLap(competitorID int, timeStr string) {
	if competitor, exists := c.Competitors[competitorID]; exists {
		eventTime, _ := model.ParseTime(timeStr)

		var startTime time.Time
		if competitor.CurrentLap == 1 {
			startTime, _ = model.ParseTime(competitor.PlannedStartTime)
		} else {
			startTime, _ = model.ParseTime(competitor.LapTimes[competitor.CurrentLap-2].EndTime)
		}

		lapDuration := eventTime.Sub(startTime)

		speed := 0.0
		if lapDuration.Seconds() > 0 {
			speed = float64(c.Config.LapLen) / lapDuration.Seconds()
		}

		if competitor.CurrentLap <= len(competitor.LapTimes) {
			competitor.LapTimes[competitor.CurrentLap-1] = model.LapInfo{
				Time:    model.FormatDuration(lapDuration),
				Speed:   speed,
				EndTime: timeStr,
			}
		}

		competitor.CurrentLap++

		if competitor.CurrentLap > c.Config.Laps {
			c.finishCompetitor(competitorID, timeStr)
		}
	}
}

func (c *Controller) finishCompetitor(competitorID int, timeStr string) {
	if competitor, exists := c.Competitors[competitorID]; exists {
		competitor.Status = "Finished"
		competitor.EndTime = timeStr

		evt := event.Event{
			Time:         timeStr,
			EventID:      event.Finished,
			CompetitorID: competitorID,
		}
		logEntry := event.FormatLogEntry(evt)
		c.OutputLog = append(c.OutputLog, fmt.Sprintf("[%s] %s", timeStr, logEntry))
	}
}

func (c *Controller) disqualifyCompetitor(competitorID int, evtTime *string) {
	if competitor, exists := c.Competitors[competitorID]; exists {
		competitor.Status = "NotStarted"
		if evtTime == nil {
			plannedStart, _ := model.ParseTime(competitor.PlannedStartTime)
			startDelta, _ := time.ParseDuration(strings.Replace(c.Config.StartDelta, ":", "h", 1) + "m")
			disqualificationTime := plannedStart.Add(startDelta)
			t := model.FormatDuration(time.Since(time.Time{}.Add(disqualificationTime.Sub(time.Time{}))))
			evtTime = &t
		}

		evt := event.Event{
			Time:         *evtTime,
			EventID:      event.Disqualified,
			CompetitorID: competitorID,
		}
		logEntry := event.FormatLogEntry(evt)
		c.OutputLog = append(c.OutputLog, fmt.Sprintf("[%s] %s", evt.Time, logEntry))
	}
}

func (c *Controller) competitorCannotContinue(competitorID int, reason string) {
	if competitor, exists := c.Competitors[competitorID]; exists {
		competitor.Status = "NotFinished"
		competitor.CannotContinue = reason
	}
}

func (c *Controller) GenerateReport() string {
	return report.GenerateFinalReport(c.Competitors, c.Config)
}
