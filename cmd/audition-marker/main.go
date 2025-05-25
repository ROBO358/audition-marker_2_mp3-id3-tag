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

// Execute runs the main logic of the application
func Execute() {
	// Define command line options
	csvPath := flag.String("csv", "", "Path to CSV file containing Adobe Audition marker information (required)")
	inputMP3 := flag.String("input", "", "Path to the original MP3 file to add chapters to (required)")
	outputMP3 := flag.String("output", "", "Output path for MP3 file with chapters added (creates filename_with_chapters.mp3 if not specified)")

	// Customize help message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -csv <CSV file path> -input <input MP3 path> [-output <output MP3 path>]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  Add chapters and save as podcast_with_chapters.mp3:\n")
		fmt.Fprintf(os.Stderr, "  %s -csv \"marker.csv\" -input \"podcast.mp3\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Save with a custom output filename:\n")
		fmt.Fprintf(os.Stderr, "  %s -csv \"marker.csv\" -input \"podcast.mp3\" -output \"custom_filename.mp3\"\n", os.Args[0])
	}

	flag.Parse()

	// Validate required options
	if *csvPath == "" || *inputMP3 == "" {
		fmt.Fprintln(os.Stderr, "Error: CSV file path and input MP3 path are required")
		flag.Usage()
		os.Exit(1)
	}

	// Check if files exist
	if !fileExists(*csvPath) {
		fmt.Fprintf(os.Stderr, "Error: CSV file '%s' not found\n", *csvPath)
		os.Exit(1)
	}

	if !fileExists(*inputMP3) {
		fmt.Fprintf(os.Stderr, "Error: Input MP3 file '%s' not found\n", *inputMP3)
		os.Exit(1)
	}

	// Check file extensions
	if !strings.EqualFold(filepath.Ext(*inputMP3), ".mp3") {
		fmt.Fprintf(os.Stderr, "Error: Input file '%s' is not an MP3 file\n", *inputMP3)
		os.Exit(1)
	}

	if *outputMP3 != "" && !strings.EqualFold(filepath.Ext(*outputMP3), ".mp3") {
		fmt.Fprintf(os.Stderr, "Error: Output file '%s' does not have an MP3 extension\n", *outputMP3)
		os.Exit(1)
	}

	// Parse CSV file
	fmt.Printf("Parsing CSV file '%s'...\n", *csvPath)
	markers, err := csvparser.ParseAuditionCSV(*csvPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred while parsing CSV: %v\n", err)
		os.Exit(1)
	}

	if len(markers) == 0 {
		fmt.Println("Warning: No markers found in the CSV file")
	} else {
		fmt.Printf("Loaded %d markers\n", len(markers))
	}

	// Add ID3 tags to MP3 file
	fmt.Println("Adding chapter tags to MP3 file...")
	err = id3tag.AddChapters(*inputMP3, markers, *outputMP3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error occurred while adding chapter tags: %v\n", err)
		os.Exit(1)
	}

	// Determine which file has the chapters
	var targetFile string
	if *outputMP3 != "" {
		targetFile = *outputMP3
	} else {
		// Calculate the auto-generated output path
		ext := filepath.Ext(*inputMP3)
		baseName := (*inputMP3)[:len(*inputMP3)-len(ext)]
		targetFile = baseName + "_with_chapters" + ext
	}

	// Success message
	if *outputMP3 == "" {
		// Determine the auto-generated output path
		ext := filepath.Ext(*inputMP3)
		baseName := (*inputMP3)[:len(*inputMP3)-len(ext)]
		generatedOutput := baseName + "_with_chapters" + ext
		fmt.Printf("Complete! MP3 file with chapter tags output to '%s'\n", generatedOutput)
	} else {
		fmt.Printf("Complete! MP3 file with chapter tags output to '%s'\n", *outputMP3)
	}

	// Read and display the chapters from the output file
	fmt.Println("\nVerifying chapters in the output file:")
	chapters, err := id3tag.ReadChapters(targetFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not read chapters from the output file: %v\n", err)
		return
	}

	if len(chapters) == 0 {
		fmt.Println("No chapters found in the output file.")
		return
	}

	// Try to read TOC information
	tocInfo, tocErr := id3tag.ReadTOC(targetFile)
	if tocErr == nil {
		fmt.Println("Table of Contents information:")
		fmt.Printf("Title: %s\n", tocInfo.Title)
		fmt.Printf("Is Top Level: %t\n", tocInfo.IsTopLevel)
		fmt.Printf("Is Ordered: %t\n", tocInfo.IsOrdered)
		fmt.Printf("Child Elements: %d\n", len(tocInfo.ChildIDs))
		fmt.Println("------------------------------------------------------------")
	}

	fmt.Printf("Found %d chapters in the output file:\n", len(chapters))
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
