package id3tag

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ROBO358/audition-marker_2_mp3-id3-tag/pkg/csvparser"
	"github.com/bogem/id3v2/v2"
)

// AddChapters adds chapter tags to an MP3 file
func AddChapters(mp3Path string, markers []csvparser.MarkerEntry, outputPath string) error {
	// If output path is not specified, overwrite the input file
	if outputPath == "" {
		outputPath = mp3Path
	}

	// If input and output files are the same
	if mp3Path == outputPath {
		// Modify ID3 tags directly
		return addChaptersInPlace(mp3Path, markers)
	} else {
		// Copy to a new file and add tags
		return addChaptersToNewFile(mp3Path, markers, outputPath)
	}
}

// addChaptersInPlace adds chapter tags directly to an existing MP3 file
func addChaptersInPlace(mp3Path string, markers []csvparser.MarkerEntry) error {
	// Ask for confirmation before modifying the original file
	if err := confirmOperation(fmt.Sprintf("This will modify the original file '%s'. Continue? (y/n): ", mp3Path)); err != nil {
		return err
	}

	// Open the MP3 file
	tag, err := id3v2.Open(mp3Path, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("could not open MP3 file: %w", err)
	}
	defer tag.Close()

	// Add chapter tags
	if err = addChapterFrames(tag, markers); err != nil {
		return err
	}

	// Save changes
	return tag.Save()
}

// confirmOperation asks the user for confirmation before proceeding with an operation
func confirmOperation(prompt string) error {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		return fmt.Errorf("operation cancelled by user")
	}

	return nil
}

// addChaptersToNewFile adds chapter tags to a new MP3 file
func addChaptersToNewFile(mp3Path string, markers []csvparser.MarkerEntry, outputPath string) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Check if output file already exists and ask for confirmation
	if fileExists(outputPath) {
		if err := confirmOperation(fmt.Sprintf("File '%s' already exists. Overwrite? (y/n): ", outputPath)); err != nil {
			return err
		}
	}

	// Create temporary file for processing
	tempPath := outputPath + ".tmp"
	if err := copyFile(mp3Path, tempPath); err != nil {
		return err
	}
	// Make sure to clean up temp file in case of failures
	defer func() {
		if fileExists(tempPath) {
			os.Remove(tempPath)
		}
	}()

	// Add ID3 tags to temporary file
	tag, err := id3v2.Open(tempPath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open temporary file: %w", err)
	}

	// Add chapter tags
	if err = addChapterFrames(tag, markers); err != nil {
		tag.Close()
		return err
	}

	// Save and close tags
	err = tag.Save()
	tag.Close()
	if err != nil {
		return fmt.Errorf("failed to save tags: %w", err)
	}

	// Move the temporary file to the final output file on success
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("failed to create final file: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Open input file
	inputFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Create output file
	outputFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer outputFile.Close()

	// Copy contents from input file to output file
	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// addChapterFrames adds chapter frames to ID3 tags
func addChapterFrames(tag *id3v2.Tag, markers []csvparser.MarkerEntry) error {
	// Remove existing chapter frames (to avoid duplication)
	tag.DeleteFrames("CHAP")

	if len(markers) == 0 {
		return nil // Do nothing if there are no markers
	}

	// Generate and add chapter frames for each marker
	for i, marker := range markers {
		// Skip markers with empty names
		if strings.TrimSpace(marker.Name) == "" {
			continue
		}

		// Unique ID for chapter element
		elementID := fmt.Sprintf("chp%d", i)

		// Create chapter frame
		chapterFrame := createChapterFrame(elementID, marker.Name, marker.StartTime)

		// Add chapter frame to tag
		tag.AddFrame("CHAP", chapterFrame)
	}

	return nil
}

// createChapterFrame creates a new chapter frame with the given parameters
func createChapterFrame(elementID string, title string, startTime time.Duration) id3v2.ChapterFrame {
	return id3v2.ChapterFrame{
		ElementID:   elementID,
		StartTime:   startTime,
		EndTime:     id3v2.IgnoredOffset,
		StartOffset: id3v2.IgnoredOffset,
		EndOffset:   id3v2.IgnoredOffset,
		Title: &id3v2.TextFrame{
			Encoding: id3v2.EncodingUTF8,
			Text:     title,
		},
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
