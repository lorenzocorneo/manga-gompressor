package main

import (
	"flag"
	"fmt"
	"log"
	"mangagompressor/mangagompressor"
	"strings"
)



func main() {
	// Define command-line flags
	files := flag.String("files", "", "Comma-separated list of CBZ files to modify (e.g., file1.cbz,file2.cbz)")
	threshold := flag.Int("threshold", 128, "Binarization threshold percentage (0-100)")
	outputDir := flag.String("output", ".", "Directory to save modified CBZ files")
	// resize := flag.String("resize", "", "New size maximum size of the images")

	// Parse command-line flags
	flag.Parse()

	// Validate inputs
	if *files == "" {
		log.Fatal("You must provide a list of CBZ files using the --files flag")
	}
	if *threshold < 0 || *threshold > 100 {
		log.Fatal("Threshold must be between 0 and 100")
	}

	// Split input files into a slice
	filesList := strings.Split(*files, ",")
	for _, cbzFile := range filesList {
		cbzFile = strings.TrimSpace(cbzFile)
		fmt.Printf("Processing CBZ file: %s\n", cbzFile)
		outputPath, err := mangagompressor.ProcessCBZFile(cbzFile, *threshold, *outputDir)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", cbzFile, err)
		} else {
			fmt.Printf("Successfully created: %s\n", outputPath)
		}
	}
}
