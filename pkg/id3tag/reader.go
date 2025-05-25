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
