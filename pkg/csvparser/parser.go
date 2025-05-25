package csvparser

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type MarkerEntry struct {
	Name      string
	StartTime time.Duration
}

// ParseAuditionCSV parses marker CSV file from Adobe Audition
func ParseAuditionCSV(filepath string) ([]MarkerEntry, error) {
	// Open the CSV file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Read CSV data
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	reader.LazyQuotes = true // Handle quotes more flexibly
	reader.TrimLeadingSpace = true

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV data: %w", err)
	}

	// Check if file is empty
	if len(records) <= 1 {
		return []MarkerEntry{}, nil
	}

	// Find header row to determine column indexes
	var nameIdx, startTimeIdx int
	nameIdx, startTimeIdx = -1, -1

	// Look for header row
	for i, row := range records {
		if len(row) > 0 {
			for j, cell := range row {
				cellLower := strings.ToLower(strings.TrimSpace(cell))
				if strings.Contains(cellLower, "name") {
					nameIdx = j
				} else if strings.Contains(cellLower, "start") {
					startTimeIdx = j
				}
			}

			// If we found the header row, start parsing from next row
			if nameIdx >= 0 && startTimeIdx >= 0 {
				records = records[i+1:]
				break
			}
		}
	}

	// Check if we found necessary columns
	if nameIdx < 0 || startTimeIdx < 0 {
		return nil, fmt.Errorf("CSV format error: could not find Name and Start columns")
	}

	// Parse all markers
	var markers []MarkerEntry
	for _, row := range records {
		if len(row) <= max(nameIdx, startTimeIdx) {
			continue // Skip rows that don't have enough columns
		}

		// Get marker name
		name := strings.TrimSpace(row[nameIdx])
		if name == "" {
			continue // Skip entries without a name
		}

		// Parse start time
		startTimeStr := strings.TrimSpace(row[startTimeIdx])
		startTime, err := parseTimeString(startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start time '%s': %w", startTimeStr, err)
		}

		// Add the marker to our list with zero EndTime
		markers = append(markers, MarkerEntry{
			Name:      name,
			StartTime: startTime,
		})
	}

	return markers, nil
}

// parseTimeString converts time strings in various formats to time.Duration
func parseTimeString(timeStr string) (time.Duration, error) {
	// Try to parse as decimal seconds
	if seconds, err := strconv.ParseFloat(timeStr, 64); err == nil {
		return time.Duration(seconds * float64(time.Second)), nil
	}

	// Try to parse as MM:SS.mmm format
	if strings.Contains(timeStr, ":") {
		parts := strings.Split(timeStr, ":")
		if len(parts) == 2 {
			// MM:SS.mmm format
			minutes, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid minutes format: %v", err)
			}

			seconds, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid seconds format: %v", err)
			}

			totalSeconds := minutes*60 + seconds
			return time.Duration(totalSeconds * float64(time.Second)), nil
		} else if len(parts) == 3 {
			// HH:MM:SS.mmm format
			hours, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid hours format: %v", err)
			}

			minutes, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid minutes format: %v", err)
			}

			seconds, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid seconds format: %v", err)
			}

			totalSeconds := hours*3600 + minutes*60 + seconds
			return time.Duration(totalSeconds * float64(time.Second)), nil
		}
	}

	return 0, fmt.Errorf("unsupported time format: %s", timeStr)
}

// max returns the maximum value among the provided integers
func max(values ...int) int {
	if len(values) == 0 {
		return 0
	}
	maxVal := values[0]
	for _, v := range values[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}
