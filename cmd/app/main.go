package main

import (
	"fmt"
	"os"
	"path/filepath"

	"biathlon/config"
	"biathlon/event"
	"biathlon/race"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: ./cmd/app/main.go config.json events output_prefix")
		return
	}

	configPath := os.Args[1]
	eventsPath := os.Args[2]
	outputPrefix := os.Args[3]

	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		return
	}

	raceCtrl := race.NewController(cfg)

	events, err := event.LoadFromFile(eventsPath)
	if err != nil {
		fmt.Printf("Error loading events: %v\n", err)
		return
	}

	outputLog, err := raceCtrl.ProcessEvents(events)
	if err != nil {
		fmt.Printf("Error processing events: %v\n", err)
		return
	}

	outputDir := filepath.Dir(outputPrefix)
	if outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			return
		}
	}

	logFile := outputPrefix + "_log.txt"
	if err := os.WriteFile(logFile, []byte(outputLog), 0644); err != nil {
		fmt.Printf("Error writing output log: %v\n", err)
		return
	}

	report := raceCtrl.GenerateReport()
	reportFile := outputPrefix + "_report.txt"
	if err := os.WriteFile(reportFile, []byte(report), 0644); err != nil {
		fmt.Printf("Error writing final report: %v\n", err)
		return
	}

	fmt.Println("Processing completed successfully!")
}
