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
	threshold := flag.Int("binarize", -1, "Binarization threshold percentage (0-100)")
	resize := flag.String("resize", "1236x1648", "New size of the images (<width>x<height>)")
	max_inner_rects := flag.Int("max-inner-rects", 9, "If algorithm finds more than max-inner-rects inner rectangles, the original image is used")
	min_content_percent := flag.Int("min-content", 10, "If algorithm returns less than min-content % of image, it uses the original image (0-100)")
	outputDir := flag.String("output", ".", "Directory to save modified CBZ files")

	// Parse command-line flags
	flag.Parse()

	// Validate inputs
	if *files == "" {
		log.Fatal("You must provide a list of CBZ files using the --files flag")
	}
	if *threshold < -1 || *threshold > 100 {
		log.Fatal("Threshold must be between 0 and 100")
	}

	resizeWidth := 0
	resizeHeight := 0
	if len(*resize) > 0 {
		_, err := fmt.Sscanf(*resize, "%dx%d", &resizeWidth, &resizeHeight)
		if err != nil {
			log.Fatal("Invalid format for resize argument", err)
		}
	}

	// Split input files into a slice
	filesList := strings.Split(*files, ",")
	for _, cbzFile := range filesList {
		cbzFile = strings.TrimSpace(cbzFile)
		fmt.Printf("Processing CBZ file: %s\n", cbzFile)
		outputPath, err := mangagompressor.ProcessCBZFile(cbzFile, *threshold, resizeWidth, resizeHeight, *max_inner_rects, *min_content_percent, *outputDir)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", cbzFile, err)
		} else {
			fmt.Printf("Successfully created: %s\n", outputPath)
		}
	}
}
