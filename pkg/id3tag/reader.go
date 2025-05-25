package id3tag

import (
	"fmt"
	"sort"
	"time"

	"github.com/bogem/id3v2/v2"
)

// Chapter represents a single chapter from an MP3 file's ID3 tags
type Chapter struct {
	Title     string
	StartTime time.Duration
}

// CTOCInfo represents CTOC (Table of Contents) information from an MP3 file
type CTOCInfo struct {
	Title      string
	IsTopLevel bool
	IsOrdered  bool
	ChildIDs   []string
}

// ReadChapters reads chapter information from an MP3 file
func ReadChapters(mp3Path string) ([]Chapter, error) {
	tag, err := id3v2.Open(mp3Path, id3v2.Options{Parse: true})
	if err != nil {
		return nil, fmt.Errorf("failed to open MP3 file: %w", err)
	}
	defer tag.Close()

	var chapters []Chapter

	// Get all chapter frames
	for _, frame := range tag.GetFrames("CHAP") {
		chapterFrame, ok := frame.(id3v2.ChapterFrame)
		if !ok {
			continue // Skip non-chapter frames
		}

		// Extract title from the chapter frame
		var title string
		if chapterFrame.Title != nil {
			title = chapterFrame.Title.Text
		}

		// Add to our chapters slice
		chapters = append(chapters, Chapter{
			Title:     title,
			StartTime: chapterFrame.StartTime,
		})
	}

	// Sort chapters by start time
	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].StartTime < chapters[j].StartTime
	})

	return chapters, nil
}

// ReadTOC reads the Table of Contents information from an MP3 file
func ReadTOC(mp3Path string) (*CTOCInfo, error) {
	tag, err := id3v2.Open(mp3Path, id3v2.Options{Parse: true})
	if err != nil {
		return nil, fmt.Errorf("failed to open MP3 file: %w", err)
	}
	defer tag.Close()

	// Get all CTOC frames
	for _, frame := range tag.GetFrames("CTOC") {
		// We only need the first CTOC frame
		ctocInfo := &CTOCInfo{}

		// Try to extract information from the raw frame bytes
		unknownFrame, ok := frame.(id3v2.UnknownFrame)
		if !ok {
			continue
		}

		// Extract Element ID (null-terminated string)
		rawData := unknownFrame.Body
		if len(rawData) < 3 { // Need at least element ID, null, flags, count
			continue
		}

		// Find the first null byte for ElementID
		idEnd := 0
		for i, b := range rawData {
			if b == 0 {
				idEnd = i
				break
			}
		}

		if idEnd == 0 {
			continue
		}

		// Extract flags (1 byte after the null terminator of ElementID)
		flagsPos := idEnd + 1
		if len(rawData) <= flagsPos {
			continue
		}
		flags := rawData[flagsPos]
		ctocInfo.IsTopLevel = (flags & 1) != 0
		ctocInfo.IsOrdered = (flags & 2) != 0

		// Extract entry count (1 byte after flags)
		countPos := flagsPos + 1
		if len(rawData) <= countPos {
			continue
		}
		entryCount := int(rawData[countPos])

		// Extract child element IDs
		ctocInfo.ChildIDs = make([]string, 0, entryCount)
		pos := countPos + 1

		// Iterate through the child element IDs
		for i := 0; i < entryCount && pos < len(rawData); i++ {
			// Find the next null byte for this child ID
			startPos := pos
			endPos := startPos
			for endPos < len(rawData) && rawData[endPos] != 0 {
				endPos++
			}

			if endPos >= len(rawData) {
				break // End of data
			}

			// Extract the child ID
			childID := string(rawData[startPos:endPos])
			ctocInfo.ChildIDs = append(ctocInfo.ChildIDs, childID)
			pos = endPos + 1 // Move past the null terminator
		}

		// Look for TIT2 frame after the child IDs
		// Simple heuristic: Look for "TIT2" bytes in the remaining data
		for i := pos; i < len(rawData)-4; i++ {
			if string(rawData[i:i+4]) == "TIT2" {
				// Found TIT2 frame, skip the frame header (10 bytes) to get to the text content
				textPos := i + 10
				if textPos >= len(rawData) {
					break
				}

				// Text frames start with encoding byte
				encoding := rawData[textPos]
				textPos++

				// Extract text based on encoding
				var title string
				if encoding == 0 || encoding == 3 { // ISO-8859-1 (0) or UTF-8 (3)
					// UTF-8 or ISO - read until end or null
					endPos := textPos
					for endPos < len(rawData) && rawData[endPos] != 0 {
						endPos++
					}
					title = string(rawData[textPos:endPos])
				} else {
					// Other encodings, just try to get something readable
					endPos := textPos
					for endPos < len(rawData) && endPos < textPos+50 && rawData[endPos] != 0 {
						endPos++
					}
					title = string(rawData[textPos:endPos])
				}

				ctocInfo.Title = title
				break
			}
		}

		return ctocInfo, nil
	}

	return nil, fmt.Errorf("no CTOC frame found")
}

// FormatDuration formats a time.Duration into a readable string format (HH:MM:SS.mmm)
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	millis := int(d.Milliseconds()) % 1000

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%03d", hours, minutes, seconds, millis)
	}
	return fmt.Sprintf("%d:%02d.%03d", minutes, seconds, millis)
}
