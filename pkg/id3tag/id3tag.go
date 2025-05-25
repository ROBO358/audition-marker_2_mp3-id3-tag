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
	// If output path is not specified, create a new filename with "_with_chapters" suffix
	if outputPath == "" {
		outputPath = generateOutputPath(mp3Path)
	}

	// If input and output file paths are the same
	if mp3Path == outputPath {
		// Modify the file directly
		return addChaptersInPlace(mp3Path, markers)
	} else {
		// Copy to a new file and add tags
		return addChaptersToNewFile(mp3Path, markers, outputPath)
	}
}

// generateOutputPath generates an output file path from the input file path
func generateOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	baseName := inputPath[:len(inputPath)-len(ext)]
	return baseName + "_with_chapters" + ext
}

// addChaptersInPlace adds chapter tags directly to an existing MP3 file
func addChaptersInPlace(mp3Path string, markers []csvparser.MarkerEntry) error {
	// Confirm before modifying the original file
	if err := confirmOperation(fmt.Sprintf("This will modify the original file '%s'. Continue? (y/n): ", mp3Path)); err != nil {
		return err
	}

	// Open MP3 file
	tag, err := id3v2.Open(mp3Path, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("Cannot open MP3 file: %w", err)
	}
	defer tag.Close()

	// Add chapter tags
	if err = addChapterFrames(tag, markers); err != nil {
		return err
	}

	// Save changes
	return tag.Save()
}

// confirmOperation asks for user confirmation before proceeding with an operation
func confirmOperation(prompt string) error {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("Error reading input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		return fmt.Errorf("Operation cancelled by user")
	}

	return nil
}

// addChaptersToNewFile adds chapter tags to a new MP3 file
func addChaptersToNewFile(mp3Path string, markers []csvparser.MarkerEntry, outputPath string) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("Failed to create output directory: %w", err)
	}

	// If output file already exists, ask for confirmation
	if fileExists(outputPath) {
		if err := confirmOperation(fmt.Sprintf("File '%s' already exists. Overwrite? (y/n): ", outputPath)); err != nil {
			return err
		}
	}

	// Create a temporary file for processing
	tempPath := outputPath + ".tmp"
	if err := copyFile(mp3Path, tempPath); err != nil {
		return err
	}

	// Clean up temporary file in case of failure
	defer func() {
		if fileExists(tempPath) {
			os.Remove(tempPath)
		}
	}()

	// Add ID3 tags to the temporary file
	tag, err := id3v2.Open(tempPath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("Cannot open temporary file: %w", err)
	}

	// Add chapter tags
	if err = addChapterFrames(tag, markers); err != nil {
		tag.Close()
		return err
	}

	// Save and close the tags
	err = tag.Save()
	tag.Close()
	if err != nil {
		return fmt.Errorf("Failed to save tags: %w", err)
	}

	// On success, move the temporary file to the final output file
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("Failed to create final file: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Open input file
	inputFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Cannot open input file: %w", err)
	}
	defer inputFile.Close()

	// Create output file
	outputFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("Failed to create temporary file: %w", err)
	}
	defer outputFile.Close()

	// Copy content from input file to output file
	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return fmt.Errorf("Failed to copy file: %w", err)
	}

	return nil
}

// addChapterFrames adds chapter frames to ID3 tags
func addChapterFrames(tag *id3v2.Tag, markers []csvparser.MarkerEntry) error {
	// Delete existing chapter and CTOC frames (to avoid duplicates)
	tag.DeleteFrames("CHAP")
	tag.DeleteFrames("CTOC")

	if len(markers) == 0 {
		return nil // Do nothing if there are no markers
	}

	// Generate chapter frames and collect their element IDs
	var chapterElementIDs []string

	for i, marker := range markers {
		// Skip markers with empty names
		if strings.TrimSpace(marker.Name) == "" {
			continue
		}

		// Unique ID for chapter element
		elementID := fmt.Sprintf("chp%d", i)
		chapterElementIDs = append(chapterElementIDs, elementID)

		// Create chapter frame
		chapterFrame := createChapterFrame(elementID, marker.Name, marker.StartTime)

		// Add chapter frame to the tag
		tag.AddFrame("CHAP", chapterFrame)
	}

	// Exit if there are no valid chapters
	if len(chapterElementIDs) == 0 {
		return nil
	}

	// Create a table of contents frame referencing all chapters
	tocFrameID := "toc"
	tocTitle := "Table of Contents"
	tocFrame := createCTOCFrame(tocFrameID, true, true, chapterElementIDs, tocTitle)

	// Add CTOC frame to the tag
	tag.AddFrame("CTOC", tocFrame)

	return nil
}

// createChapterFrame creates a new chapter frame with the given parameters
func createChapterFrame(elementID string, title string, startTime time.Duration) id3v2.ChapterFrame {
	return id3v2.ChapterFrame{
		ElementID:   elementID,
		StartTime:   startTime,
		EndTime:     id3v2.IgnoredOffset, // Ignore end time
		StartOffset: id3v2.IgnoredOffset, // Ignore start offset
		EndOffset:   id3v2.IgnoredOffset, // Ignore end offset
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
