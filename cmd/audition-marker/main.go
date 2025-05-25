package auditionmarker

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ROBO358/audition-marker_2_mp3-id3-tag/pkg/csvparser"
	"github.com/ROBO358/audition-marker_2_mp3-id3-tag/pkg/id3tag"
)

// Config holds the application settings
type Config struct {
	CSVPath   string // Path to the marker CSV file
	InputMP3  string // Path to the original MP3 file
	OutputMP3 string // Path for the output MP3 with chapters
}

// Execute runs the main application logic
func Execute() {
	// Parse and validate command line arguments
	config, err := parseAndValidateArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		flag.Usage()
		os.Exit(1)
	}

	// Parse markers from CSV file
	fmt.Printf("Parsing CSV file '%s'...\n", config.CSVPath)
	markers, err := csvparser.ParseAuditionCSV(config.CSVPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred while parsing CSV: %v\n", err)
		os.Exit(1)
	}

	// Display marker information
	showMarkerInfo(markers)

	// Add chapter tags to MP3 file
	fmt.Println("Adding chapter tags to MP3 file...")
	err = id3tag.AddChapters(config.InputMP3, markers, config.OutputMP3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred while adding chapter tags: %v\n", err)
		os.Exit(1)
	}

	// Determine output file path
	targetFile := determineOutputPath(config.InputMP3, config.OutputMP3)

	// Display success message
	showSuccessMessage(targetFile)

	// Verify and display chapters from output file
	verifyAndShowChapters(targetFile)
}

// parseAndValidateArgs parses and validates command line arguments
func parseAndValidateArgs() (*Config, error) {
	// Define command line options
	csvPath := flag.String("csv", "", "Path to CSV file containing Adobe Audition markers (required)")
	inputMP3 := flag.String("input", "", "Path to original MP3 file to add chapters to (required)")
	outputMP3 := flag.String("output", "", "Path for output MP3 file with chapters (if not specified, will output as filename_with_chapters.mp3)")

	// Customize help message
	customizeHelpMessage()

	flag.Parse()

	// Create configuration
	config := &Config{
		CSVPath:   *csvPath,
		InputMP3:  *inputMP3,
		OutputMP3: *outputMP3,
	}

	// Validate required options
	if config.CSVPath == "" || config.InputMP3 == "" {
		return nil, fmt.Errorf("CSV file path and input MP3 path are required")
	}

	// Check file existence
	if !fileExists(config.CSVPath) {
		return nil, fmt.Errorf("CSV file '%s' not found", config.CSVPath)
	}

	if !fileExists(config.InputMP3) {
		return nil, fmt.Errorf("Input MP3 file '%s' not found", config.InputMP3)
	}

	// Check file extensions
	if !strings.EqualFold(filepath.Ext(config.InputMP3), ".mp3") {
		return nil, fmt.Errorf("Input file '%s' is not an MP3 file", config.InputMP3)
	}

	if config.OutputMP3 != "" && !strings.EqualFold(filepath.Ext(config.OutputMP3), ".mp3") {
		return nil, fmt.Errorf("Output file '%s' does not have MP3 extension", config.OutputMP3)
	}

	return config, nil
}

// customizeHelpMessage customizes the help message
func customizeHelpMessage() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -csv <CSV file path> -input <input MP3 path> [-output <output MP3 path>]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  Add chapters and save as podcast_with_chapters.mp3:\n")
		fmt.Fprintf(os.Stderr, "  %s -csv \"marker.csv\" -input \"podcast.mp3\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Save with custom output filename:\n")
		fmt.Fprintf(os.Stderr, "  %s -csv \"marker.csv\" -input \"podcast.mp3\" -output \"custom_filename.mp3\"\n", os.Args[0])
	}
}

// showMarkerInfo displays marker information
func showMarkerInfo(markers []csvparser.MarkerEntry) {
	if len(markers) == 0 {
		fmt.Println("Warning: No markers found in CSV file")
	} else {
		fmt.Printf("Loaded %d markers\n", len(markers))
	}
}

// determineOutputPath determines the output file path
func determineOutputPath(inputPath, outputPath string) string {
	if outputPath != "" {
		return outputPath
	}

	// Calculate automatically generated output path
	ext := filepath.Ext(inputPath)
	baseName := inputPath[:len(inputPath)-len(ext)]
	return baseName + "_with_chapters" + ext
}

// showSuccessMessage displays success message
func showSuccessMessage(outputPath string) {
	fmt.Printf("Done! MP3 file with chapter tags has been saved to '%s'\n", outputPath)
}

// verifyAndShowChapters reads and displays chapters from the output file
func verifyAndShowChapters(filePath string) {
	fmt.Println("\nVerifying chapters in output file:")

	// Get chapter information
	chapters, err := id3tag.ReadChapters(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not read chapters from output file: %v\n", err)
		return
	}

	if len(chapters) == 0 {
		fmt.Println("No chapters found in output file.")
		return
	}

	// Read table of contents information
	tocInfo, tocErr := id3tag.ReadTOC(filePath)
	if tocErr == nil {
		fmt.Println("Table of Contents information:")
		fmt.Printf("Title: %s\n", tocInfo.Title)
		fmt.Printf("Top level: %t\n", tocInfo.IsTopLevel)
		fmt.Printf("Ordered: %t\n", tocInfo.IsOrdered)
		fmt.Printf("Child elements: %d\n", len(tocInfo.ChildIDs))
		fmt.Println("------------------------------------------------------------")
	}

	// Display chapter list
	fmt.Printf("Found %d chapters in output file:\n", len(chapters))
	fmt.Println("------------------------------------------------------------")
	fmt.Printf("%-4s | %-12s | %s\n", "No.", "Start Time", "Title")
	fmt.Println("------------------------------------------------------------")
	for i, chapter := range chapters {
		fmt.Printf("%-4d | %-12s | %s\n", i+1, id3tag.FormatDuration(chapter.StartTime), chapter.Title)
	}
	fmt.Println("------------------------------------------------------------")
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return !info.IsDir()
}
