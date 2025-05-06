package report

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"biathlon/config"
	"biathlon/model"
)

func GenerateFinalReport(competitors map[int]*model.Competitor, cfg *config.Config) string {
	var sortedCompetitors []*model.Competitor
	for _, competitor := range competitors {
		sortedCompetitors = append(sortedCompetitors, competitor)
	}

	sort.Slice(sortedCompetitors, func(i, j int) bool {
		c1, c2 := sortedCompetitors[i], sortedCompetitors[j]

		if c1.Status != c2.Status {
			if c1.Status == "Finished" && c2.Status != "Finished" {
				return true
			}
			if c1.Status != "Finished" && c2.Status == "Finished" {
				return false
			}
			return c1.Status < c2.Status
		}

		if c1.Status == "Finished" && c2.Status == "Finished" {
			end1, _ := model.ParseTime(c1.EndTime)
			planned1, _ := model.ParseTime(c1.PlannedStartTime)

			end2, _ := model.ParseTime(c2.EndTime)
			planned2, _ := model.ParseTime(c2.PlannedStartTime)

			totalTime1 := end1.Sub(planned1)
			for _, lap := range c1.LapTimes {
				lapDuration, _ := time.ParseDuration(strings.Replace(lap.Time, ":", "h", 1) + "m")
				totalTime1 += lapDuration
			}

			totalTime2 := end2.Sub(planned2)
			for _, lap := range c2.LapTimes {
				lapDuration, _ := time.ParseDuration(strings.Replace(lap.Time, ":", "h", 1) + "m")
				totalTime2 += lapDuration
			}

			return totalTime1 < totalTime2
		}

		return c1.ID < c2.ID
	})

	var report strings.Builder

	for _, competitor := range sortedCompetitors {
		lapTimesStr := "["
		for i, lap := range competitor.LapTimes {
			if lap.Time != "" {
				lapTimesStr += fmt.Sprintf("{%s, %.3f}", lap.Time, math.Floor(lap.Speed*1000)/1000)
			} else {
				lapTimesStr += "{,}"
			}

			if i < len(competitor.LapTimes)-1 {
				lapTimesStr += ", "
			}
		}
		lapTimesStr += "]"

		penaltyStr := "{,}"
		if competitor.PenaltyStartTime != "" {
			penaltyTime := model.FormatDuration(competitor.PenaltyDuration)
			penaltyStr = fmt.Sprintf("{%s, %.3f}", penaltyTime, math.Floor(competitor.PenaltySpeed*1000)/1000)

		}

		hitsStr := fmt.Sprintf("%d/%d", competitor.HitsCount, competitor.ShotsCount)

		var statusStr string
		if competitor.Status == "" || competitor.Status == "Finished" {
			end, _ := model.ParseTime(competitor.EndTime)
			planned, _ := model.ParseTime(competitor.PlannedStartTime)
			totalTime := end.Sub(planned)
			for _, lap := range competitor.LapTimes {
				if lap.Time != "" {
					lapDuration, _ := time.ParseDuration(strings.Replace(lap.Time, ":", "h", 1) + "m")
					totalTime += lapDuration
				}
			}
			statusStr = fmt.Sprintf("[%s]", model.FormatDuration(totalTime))
		} else {
			statusStr = fmt.Sprintf("[%s]", competitor.Status)
		}

		report.WriteString(fmt.Sprintf("%s %d %s %s %s\n", statusStr, competitor.ID, lapTimesStr, penaltyStr, hitsStr))
	}

	return report.String()
}
