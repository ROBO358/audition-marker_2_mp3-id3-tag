package id3tag

import (
	"fmt"
	"sort"
	"time"

	"github.com/bogem/id3v2/v2"
)

// Chapter represents a single chapter information contained in the ID3 tags of an MP3 file
type Chapter struct {
	Title     string        // Chapter title
	StartTime time.Duration // Start time of the chapter
}

// CTOCInfo represents the Table of Contents information contained in the ID3 tags of an MP3 file
type CTOCInfo struct {
	Title      string   // Title of the table of contents
	IsTopLevel bool     // Whether this is a top-level table of contents
	IsOrdered  bool     // Whether chapters are in a specific order
	ChildIDs   []string // IDs of child elements (usually CHAP frames)
}

// ReadChapters reads chapter information from an MP3 file
func ReadChapters(mp3Path string) ([]Chapter, error) {
	// Open MP3 file
	tag, err := id3v2.Open(mp3Path, id3v2.Options{Parse: true})
	if err != nil {
		return nil, fmt.Errorf("Cannot open MP3 file: %w", err)
	}
	defer tag.Close()

	var chapters []Chapter

	// Get all chapter frames
	for _, frame := range tag.GetFrames("CHAP") {
		chapterFrame, ok := frame.(id3v2.ChapterFrame)
		if !ok {
			continue // Skip non-chapter frames
		}

		// Extract title from chapter frame
		var title string
		if chapterFrame.Title != nil {
			title = chapterFrame.Title.Text
		}

		// Add to chapter list
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

// ReadTOC reads table of contents information from an MP3 file
func ReadTOC(mp3Path string) (*CTOCInfo, error) {
	// Open MP3 file
	tag, err := id3v2.Open(mp3Path, id3v2.Options{Parse: true})
	if err != nil {
		return nil, fmt.Errorf("Cannot open MP3 file: %w", err)
	}
	defer tag.Close()

	// Get all CTOC frames
	ctocFrames := tag.GetFrames("CTOC")
	if len(ctocFrames) == 0 {
		return nil, fmt.Errorf("No CTOC frame found")
	}

	// Process the first CTOC frame
	return extractCTOCInfo(ctocFrames[0])
}

// extractCTOCInfo extracts CTOC information from an ID3 frame
func extractCTOCInfo(frame id3v2.Framer) (*CTOCInfo, error) {
	ctocInfo := &CTOCInfo{}

	// Get raw data as unknown frame
	unknownFrame, ok := frame.(id3v2.UnknownFrame)
	if !ok {
		return nil, fmt.Errorf("Cannot parse CTOC frame")
	}

	// Get frame data
	rawData := unknownFrame.Body
	if len(rawData) < 3 { // Need at minimum ElementID, null, flags, count
		return nil, fmt.Errorf("CTOC frame data is incomplete")
	}

	// Find end of ElementID (first null byte)
	idEnd := -1
	for i, b := range rawData {
		if b == 0 {
			idEnd = i
			break
		}
	}

	if idEnd < 0 {
		return nil, fmt.Errorf("ElementID not found in CTOC frame")
	}

	// Get position of flags and entry count
	flagsPos := idEnd + 1
	countPos := flagsPos + 1

	if len(rawData) <= countPos {
		return nil, fmt.Errorf("Insufficient data length in CTOC frame")
	}

	// Get information from flags
	flags := rawData[flagsPos]
	ctocInfo.IsTopLevel = (flags & 1) != 0
	ctocInfo.IsOrdered = (flags & 2) != 0

	// Get entry count
	entryCount := int(rawData[countPos])

	// Extract child element IDs
	ctocInfo.ChildIDs, _ = extractChildIDs(rawData[countPos+1:], entryCount)

	// Find TIT2 frame and extract title
	ctocInfo.Title = extractTitleFromCTOC(rawData, countPos+1+len(ctocInfo.ChildIDs)*2)

	return ctocInfo, nil
}

// extractChildIDs extracts child element IDs from CTOC frame data
func extractChildIDs(data []byte, count int) ([]string, int) {
	ids := make([]string, 0, count)
	pos := 0

	for i := 0; i < count && pos < len(data); i++ {
		// Find the end of this ID (null byte)
		startPos := pos
		endPos := startPos

		for endPos < len(data) && data[endPos] != 0 {
			endPos++
		}

		if endPos >= len(data) {
			break // End of data
		}

		// Extract ID
		ids = append(ids, string(data[startPos:endPos]))
		pos = endPos + 1 // Move past null
	}

	return ids, pos
}

// extractTitleFromCTOC finds the TIT2 frame in CTOC data and extracts the title
func extractTitleFromCTOC(data []byte, startPos int) string {
	// Look for "TIT2" byte sequence
	for i := startPos; i < len(data)-4; i++ {
		if string(data[i:i+4]) == "TIT2" {
			// TIT2 frame found, skip frame header (10 bytes) to get text content
			textPos := i + 10
			if textPos >= len(data) {
				break
			}

			// Text frame starts with encoding byte
			encoding := data[textPos]
			textPos++

			// Extract text based on encoding
			if encoding == 0 || encoding == 3 { // ISO-8859-1 (0) or UTF-8 (3)
				// Read until end or null
				endPos := textPos
				for endPos < len(data) && data[endPos] != 0 {
					endPos++
				}
				return string(data[textPos:endPos])
			} else {
				// Other encoding, get readable content
				endPos := textPos
				for endPos < len(data) && endPos < textPos+50 && data[endPos] != 0 {
					endPos++
				}
				return string(data[textPos:endPos])
			}
		}
	}

	return ""
}

// FormatDuration formats a time.Duration as a human-readable string (HH:MM:SS.mmm)
func FormatDuration(d time.Duration) string {
	// Break down into hours, minutes, seconds, milliseconds
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	millis := int(d.Milliseconds()) % 1000

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%03d", hours, minutes, seconds, millis)
	}
	return fmt.Sprintf("%d:%02d.%03d", minutes, seconds, millis)
}
