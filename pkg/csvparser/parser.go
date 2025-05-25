package csvparser

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// MarkerEntry represents a single chapter marker
type MarkerEntry struct {
	Name      string        // Marker name (chapter title)
	StartTime time.Duration // Start time of the marker
}

// ParseAuditionCSV parses Adobe Audition marker CSV file
func ParseAuditionCSV(filepath string) ([]MarkerEntry, error) {
	// Open CSV file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("Cannot open CSV file: %w", err)
	}
	defer file.Close()

	// Read CSV data
	reader := csv.NewReader(file)
	reader.Comma = '\t'            // Process tab-delimited CSV file
	reader.LazyQuotes = true       // Process quotes flexibly
	reader.TrimLeadingSpace = true // Remove leading whitespace

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("Failed to read CSV data: %w", err)
	}

	// Check if file is empty
	if len(records) <= 1 {
		return []MarkerEntry{}, nil
	}

	// Find header row and determine column indices
	nameIdx, startTimeIdx, err := findHeaderColumns(records)
	if err != nil {
		return nil, err
	}

	// Parse all markers
	markers, err := parseMarkers(records, nameIdx, startTimeIdx)
	if err != nil {
		return nil, err
	}

	return markers, nil
}

// findHeaderColumns searches for the header row in CSV records and returns the required column indices
func findHeaderColumns(records [][]string) (nameIdx int, startTimeIdx int, err error) {
	nameIdx, startTimeIdx = -1, -1

	// Search for header row
	for _, row := range records {
		if len(row) > 0 {
			for j, cell := range row {
				cellLower := strings.ToLower(strings.TrimSpace(cell))
				if strings.Contains(cellLower, "name") {
					nameIdx = j
				} else if strings.Contains(cellLower, "start") {
					startTimeIdx = j
				}
			}

			// If header row is found, start parsing from the next row
			if nameIdx >= 0 && startTimeIdx >= 0 {
				return nameIdx, startTimeIdx, nil
			}
		}
	}

	// If required columns are not found
	return -1, -1, fmt.Errorf("CSV format error: 'Name' and 'Start' columns not found")
}

// parseMarkers extracts marker information from data after the header row
func parseMarkers(records [][]string, nameIdx int, startTimeIdx int) ([]MarkerEntry, error) {
	var markers []MarkerEntry

	// Skip header row and process only data rows
	dataStart := 0
	for rowIdx, row := range records {
		if len(row) > 0 {
			for _, cell := range row {
				cellLower := strings.ToLower(strings.TrimSpace(cell))
				if strings.Contains(cellLower, "name") || strings.Contains(cellLower, "start") {
					dataStart = rowIdx + 1
					break
				}
			}
			if dataStart > 0 {
				break
			}
		}
	}

	// Parse each marker
	for _, row := range records[dataStart:] {
		if len(row) <= max(nameIdx, startTimeIdx) {
			continue // Skip rows with insufficient columns
		}

		// Get marker name
		name := strings.TrimSpace(row[nameIdx])
		if name == "" {
			continue // Skip items without a name
		}

		// Parse start time
		startTimeStr := strings.TrimSpace(row[startTimeIdx])
		startTime, err := parseTimeString(startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse start time '%s': %w", startTimeStr, err)
		}

		// Add marker to the list
		markers = append(markers, MarkerEntry{
			Name:      name,
			StartTime: startTime,
		})
	}

	return markers, nil
}

// parseTimeString converts various time string formats to time.Duration
func parseTimeString(timeStr string) (time.Duration, error) {
	// Try to parse as decimal seconds
	if seconds, err := strconv.ParseFloat(timeStr, 64); err == nil {
		return time.Duration(seconds * float64(time.Second)), nil
	}

	// Try to parse as MM:SS.mmm format
	if strings.Contains(timeStr, ":") {
		parts := strings.Split(timeStr, ":")

		switch len(parts) {
		case 2:
			// MM:SS.mmm format
			return parseMinutesSeconds(parts)
		case 3:
			// HH:MM:SS.mmm format
			return parseHoursMinutesSeconds(parts)
		}
	}

	return 0, fmt.Errorf("Unsupported time format: %s", timeStr)
}

// parseMinutesSeconds parses time strings in MM:SS.mmm format
func parseMinutesSeconds(parts []string) (time.Duration, error) {
	minutes, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid minutes format: %v", err)
	}

	seconds, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid seconds format: %v", err)
	}

	totalSeconds := minutes*60 + seconds
	return time.Duration(totalSeconds * float64(time.Second)), nil
}

// parseHoursMinutesSeconds parses time strings in HH:MM:SS.mmm format
func parseHoursMinutesSeconds(parts []string) (time.Duration, error) {
	hours, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid hours format: %v", err)
	}

	minutes, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid minutes format: %v", err)
	}

	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid seconds format: %v", err)
	}

	totalSeconds := hours*3600 + minutes*60 + seconds
	return time.Duration(totalSeconds * float64(time.Second)), nil
}

// max returns the maximum value of the provided integers
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
