package mangagompressor

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// binarizeImage takes an image and a threshold percentage, then binarizes the image.
func binarizeImage(img image.Image, thresholdPercentage int) (image.Image, error) {
	threshold := uint8(thresholdPercentage * 255 / 100)
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	newImage := image.NewGray(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			originalColor := img.At(x, y)
			r, g, b, _ := originalColor.RGBA()
			grayValue := uint8((r + g + b) / 3) // Convert to grayscale by averaging the RGB values
			if grayValue > threshold {
				newImage.SetGray(x, y, color.Gray{Y: 255}) // White
			} else {
				newImage.SetGray(x, y, color.Gray{Y: 0}) // Black
			}
		}
	}

	return newImage, nil
}

func copyImage(src image.Image) image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, src.Bounds().Dx(), src.Bounds().Dy()))
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Rect.Dx(); x++ {
			img.Set(x, y, src.At(x, y))
		}
	}
	return *img
}

// cropImage takes an image and a rectangle, and returns a new image cropped to that rectangle.
func cropImage(src image.Image, rect image.Rectangle) image.Image {
	// Create a new RGBA image to store the cropped image
	cropped := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))

	// Loop through the pixels of the rectangle and copy them to the new image
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			// Get the pixel color from the source image at (x, y)
			color := src.At(x, y)

			// Set the pixel color in the new image
			cropped.Set(x-rect.Min.X, y-rect.Min.Y, color)
		}
	}

	return cropped
}

func cropRectangle(src image.Image, rect image.Rectangle) image.Image {
	return nil
}

// removeRectangle removes a rectangle from an RGBA image by filling it with the specified background color
func removeRectangle(src image.Image, rect image.Rectangle, background color.Color) image.Image {
	// Create a new RGBA image to store the modified image
	bounds := src.Bounds()
	newImage := image.NewRGBA(bounds)

	// Loop through the original image and copy pixels to the new image
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Check if the current pixel is within the rectangle to be removed
			if rect.Min.X <= x && x < rect.Max.X && rect.Min.Y <= y && y < rect.Max.Y {
				// Set the pixel in the rectangle to the background color
				newImage.Set(x, y, background)
			} else {
				// Otherwise, copy the pixel from the original image
				newImage.Set(x, y, src.At(x, y))
			}
		}
	}

	return newImage
}

// getRemainingRectangles takes an image and an array of rectangles, and returns a list
// of rectangles that are not covered by the input rectangles. The remaining rectangles
// represent the areas of the image that do not overlap with any of the input rectangles.
func getRemainingRectangles(img image.Image, rects []image.Rectangle) []image.Rectangle {
	// Get the dimensions of the image
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a grid to mark which pixels are covered by the input rectangles
	covered := make([][]bool, height)
	for i := range covered {
		covered[i] = make([]bool, width)
	}

	// Mark the covered pixels in the grid
	for _, rect := range rects {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				covered[y][x] = true
			}
		}
	}

	// Now we need to find the remaining areas that are not covered
	var remainingRects []image.Rectangle

	// Iterate through the image and identify the remaining uncovered rectangles
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// If the current pixel is not covered, it might be part of a new rectangle
			if !covered[y][x] {
				// Start a new rectangle
				startX, startY := x, y

				// Find the right boundary of this rectangle
				for x < width && !covered[y][x] {
					x++
				}

				// Find the bottom boundary of this rectangle
				for y < height && !covered[y][startX] {
					y++
				}

				// Create a new rectangle from the found bounds
				remainingRects = append(remainingRects, image.Rect(startX, startY, x, y))

				// Move to the next y position
				y--
			}
		}
	}

	return remainingRects
}

// func subtractRectangles1(src image.Image, rects []image.Rectangle) []image.Rectangle {
// 	// Create a matrix that represents the image.
// 	page := make([][]bool, src.Bounds().Dy())
// 	for i := range page {
// 		page[i] = make([]bool, src.Bounds().Dx())
// 	}

// 	// Add the rectangles to the page
// 	for _, rect := range rects {
// 		for y := rect.Min.Y; y < rect.Max.Y; y++ {
// 			for x := rect.Min.X; x < rect.Max.X; x++ {
// 				page[y][x] = true
// 			}
// 		}
// 	}

// 	// Find all the rectangles in the page that do not overlap with the input rectangles
// 	var remainingRects []image.Rectangle

// 	for y := 0; y < page.Dy(); y++ {
// 		for x := 0; x < page.Dx(); x++ {

// 		}
// 	}
// 	return nil
// }

func assembleImageFromRectangles(src image.Image, rects []image.Rectangle) image.Image {
	if len(rects) == 0 {
		return src
	}

	fmt.Println("src: ", src.Bounds())
	// The assumption is that all the rectangles have same width, as it should be safe
	// to assume given the previous step.
	width := rects[0].Dx()
	height := 0
	// Calculate smallest bound to fit all the rectangles
	for _, rect := range rects {
		height += rect.Dy()
	}

	fmt.Println(image.Rect(0, 0, width, height))

	newImage := image.NewRGBA(image.Rect(0, 0, width, height))
	x := 0
	y := 0
	fmt.Println(len(rects))
	for _, rect := range rects {
		for i := rect.Min.Y; i < rect.Min.Y+rect.Dy(); i++ {
			for j := rect.Min.X; j < rect.Min.X+rect.Dx(); j++ {
				newImage.Set(x, y, src.At(j, i))
				// fmt.Printf("img: (%d,%d) -- src: (%d,%d)\n", x, y, j, i)
				x += 1
			}
			x = 0
			y += 1
		}
	}

	return newImage
}

// detectOuterBorder detect the outer borders of a manga page
func detectOuterBorder(img image.Image) image.Rectangle {
	// Assume the background color is white (255, 255, 255)
	backgroundColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// Get image bounds
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Initialize return values
	xmin := width
	xmax := 0
	ymin := height
	ymax := 0

	// Loop through all pixels in the image
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Get the color of the current pixel
			r, g, b, _ := img.At(x, y).RGBA()

			// Normalize the RGBA values to 0-255 range (since RGBA returns a 16-bit value)
			normalizedR := uint8(r >> 8)
			normalizedG := uint8(g >> 8)
			normalizedB := uint8(b >> 8)

			// If the pixel is not white (background color), return its coordinates
			if normalizedR != backgroundColor.R || normalizedG != backgroundColor.G || normalizedB != backgroundColor.B {
				xmin = min(xmin, x)
				xmax = max(xmax, x)
				ymin = min(ymin, y)
				ymax = max(ymax, y)
				// fmt.Println(normalizedR, backgroundColor.R, normalizedG, backgroundColor.G)
				// fmt.Println(x, xmin, xmax)
			}
		}
	}
	// fmt.Println(xmin, xmax, ymin, ymax)
	return image.Rect(xmin, ymin, xmax, ymax)
}

func normalizeRGB(red, green, blue uint32) (r, g, b uint8) {
	// Normalize the RGBA values to 0-255 range (since RGBA returns a 16-bit value)
	r = uint8(r >> 8)
	g = uint8(g >> 8)
	b = uint8(b >> 8)

	return r, g, b
}

// lineDiff returns a slice of points where the line has different colors than the input color.
// The number of returned points is always even, as points go in pairs.
// func lineDiff(src image.Image, color color.RGBA, line int, horizontal bool) (points []image.Point) {
// 	points = []image.Point{}
// 	// Horizontal line
// 	if horizontal {
// 		start := src.Bounds().Max.X
// 		end := 0
// 		for i := 0; i < src.Bounds().Dx(); i++ {
// 			// Get the color of the current pixel
// 			r, g, b, _ := src.At(i, line).RGBA()

// 			// Normalize the RGBA values to 0-255 range (since RGBA returns a 16-bit value)
// 			normalizedR, normalizedG, normalizedB := normalizeRGB(r,g,b)
// 			if normalizedR != color.R || normalizedG != color.G || normalizedB != color.B {
// 				start = min(start, i)
// 			}
// 		}
// 	} else {
// 		// Vertical line
// 		for i := 0; i < src.Bounds().Dy(); i++ {
// 			// Get the color of the current pixel
// 			r, g, b, _ := src.At(line, i).RGBA()

// 			// Normalize the RGBA values to 0-255 range (since RGBA returns a 16-bit value)
// 			normalizedR := uint8(r >> 8)
// 			normalizedG := uint8(g >> 8)
// 			normalizedB := uint8(b >> 8)
// 			if normalizedR != color.R || normalizedG != color.G || normalizedB != color.B {

// 			}
// 		}
// 	}
// 	return points
// }

func isLineContinuous(src image.Image, color color.RGBA, line int, horizontal bool) bool {
	// Get image bounds
	bounds := src.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	lineCounter := 0

	// Horizontal line
	if horizontal {
		for i := 0; i < width; i++ {
			// Get the color of the current pixel
			r, g, b, _ := src.At(i, line).RGBA()

			// Normalize the RGBA values to 0-255 range (since RGBA returns a 16-bit value)
			normalizedR := uint8(r >> 8)
			normalizedG := uint8(g >> 8)
			normalizedB := uint8(b >> 8)
			if normalizedR != color.R || normalizedG != color.G || normalizedB != color.B {
				lineCounter += 1
				// return false
			}
		}
	} else {
		// Vertical line
		for i := 0; i < height; i++ {
			// Get the color of the current pixel
			r, g, b, _ := src.At(line, i).RGBA()

			// Normalize the RGBA values to 0-255 range (since RGBA returns a 16-bit value)
			normalizedR := uint8(r >> 8)
			normalizedG := uint8(g >> 8)
			normalizedB := uint8(b >> 8)
			if normalizedR != color.R || normalizedG != color.G || normalizedB != color.B {
				lineCounter += 1
				// return false
			}
		}
	}

	// This statement takes care of some random pixel
	if lineCounter >= 20 {
		// fmt.Println("LineCount: ", lineCounter)
		return false
	}

	return true
}

// detectInnerBorders detect empty spaces (left to right) between manga tables
func detectInnerBorders(src image.Image) []image.Rectangle {
	rects := []image.Rectangle{}

	// Assume the background color is white (255, 255, 255)
	backgroundColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// Get image bounds
	bounds := src.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Initialize return values
	ymin := height

	// Loop through all the rows in the image
	for y := 0; y < height; y++ {
		if isLineContinuous(src, backgroundColor, y, true) {
			// fmt.Println("Continuous: ", y)
			ymin = min(ymin, y)
		} else {
			// If ymin < y, the rectangle has broken. So append the rectangle to rects and reassign ymin to height
			if ymin < y && (y-ymin) > 15 {
				rects = append(rects, image.Rect(0, ymin, width, y))
			}
			ymin = height
		}
	}

	if ymin < height {
		rects = append(rects, image.Rect(0, ymin, width, height))
	}

	ymin = width

	// Loop through all the columns in the image
	for y := 0; y < width; y++ {
		if isLineContinuous(src, backgroundColor, y, false) {
			ymin = min(ymin, y)
		} else {
			// If ymin < y, the rectangle has broken. So append the rectangle to rects and reassign ymin to height
			if ymin < y && (y-ymin) > 15 {
				rects = append(rects, image.Rect(ymin, 0, y, height))
			}
			ymin = width
		}
	}

	if ymin < width {
		rects = append(rects, image.Rect(ymin, 0, width, height))
	}

	fmt.Println(rects)
	return rects
}

// rotateImage rotates the input image by 90 degrees clockwise.
func rotateImage(img image.Image) image.Image {
	// Get the bounds of the original image
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a new image with swapped dimensions (width becomes height, height becomes width)
	rotatedImage := image.NewRGBA(image.Rect(0, 0, height, width))

	// Loop through each pixel in the original image and copy it to the rotated position
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Original pixel is at (x, y)
			// After 90-degree clockwise rotation, it will move to (y, height-1-x)
			color := img.At(x, y)
			rotatedImage.Set(y, width-1-x, color)
		}
	}

	return rotatedImage
}

func resizeImage(src image.Image, newWidth, newHeight int) image.Image {
	// Create a new empty RGBA image with the new dimensions
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Get the dimensions of the original image
	srcWidth := src.Bounds().Dx()
	srcHeight := src.Bounds().Dy()

	// Iterate over each pixel of the new image and find the corresponding pixel in the source image
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Calculate the corresponding coordinates in the original image using nearest neighbor
			originalX := (x * srcWidth) / newWidth
			originalY := (y * srcHeight) / newHeight
			// originalX := int(float32(x) * float32(srcWidth) / float32(newWidth))
			// originalY := int(float32(y) * float32(srcHeight) / float32(newHeight))

			// Get the pixel color from the original image
			color := src.At(originalX, originalY)

			// Set the color of the corresponding pixel in the new image
			dst.Set(x, y, color)
		}
	}
	// fmt.Println(src.Bounds(), dst.Bounds())

	return dst
}

// rgbaToGray converts an RGBA image to grayscale
func rgbaToGray(src image.Image) image.Image {
	// Create a new grayscale image with the same dimensions as the source image
	bounds := src.Bounds()
	grayImage := image.NewGray(bounds)

	// Iterate through each pixel in the original image
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Get the color of the pixel at (x, y) in the RGBA image
			rgbaColor := src.At(x, y)
			r, g, b, _ := rgbaColor.RGBA()

			// Normalize RGB values to [0, 255] range
			r = r >> 8
			g = g >> 8
			b = b >> 8

			// Convert to grayscale using the luminosity method
			grayValue := uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))

			// Set the pixel in the new grayscale image
			grayImage.Set(x, y, color.Gray{Y: grayValue})
		}
	}

	return grayImage
}

func compressPNG(src image.Image, binThreshold, resizeWidth, resizeHeight, max_inner_rects, min_content int) (image.Image, error) {
	img := src

	if binThreshold >= 0 {
		// Binarize the image
		img, _ = binarizeImage(src, binThreshold)
	}

	// outerBorder := detectOuterBorder(img)
	// img = cropImage(img, outerBorder)

	innerBorders := detectInnerBorders(img)
	fmt.Println(innerBorders)
	// for _, rect := range innerBorders {
	// 	img = removeRectangle(img, rect, color.RGBA{255, 0, 0, 255})
	// }

	if len(innerBorders) < max_inner_rects {
		content := getRemainingRectangles(img, innerBorders)
		// sometimes the algorithm fails and this makes sure the original image is used.
		content_area := 0
		for _, rect := range content {
			content_area += rect.Dx() * rect.Dy()
		}
		if float32(content_area) > float32(min_content/100*resizeWidth*resizeHeight) {
			img = assembleImageFromRectangles(img, content)
		}
	}

	if img.Bounds().Dx() > img.Bounds().Dy() {
		img = rotateImage(img)
	}

	img = resizeImage(img, resizeWidth, resizeHeight)
	return img, nil
}

// ProcessCBZFile processes each image inside a CBZ file, binarizes it, and returns a list of modified image data.
func ProcessCBZFile(cbzPath string, binThreshold int, resizeWidth int, resizeHeight, max_inner_rects int,
	min_content_percent int, outputDir string) (string, error) {
	// Create a temporary folder to store the modified images
	tempDir, err := os.MkdirTemp(outputDir, "binarized_")
	if err != nil {
		return "", fmt.Errorf("unable to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Open the CBZ file (a zip archive)
	r, err := zip.OpenReader(cbzPath)
	if err != nil {
		return "", fmt.Errorf("failed to open CBZ file: %v", err)
	}
	defer r.Close()

	// Prepare a new CBZ file for output
	outputCBZPath := filepath.Join(outputDir, strings.Replace(filepath.Base(cbzPath), ".cbz", "_modified.cbz", 1))
	outputCBZ, err := os.Create(outputCBZPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output CBZ file: %v", err)
	}
	defer outputCBZ.Close()

	// Create a new zip writer to store the new images
	zipWriter := zip.NewWriter(outputCBZ)
	defer zipWriter.Close()

	// Iterate over all files in the CBZ archive
	for _, file := range r.File {
		// Read each image file in the CBZ
		rc, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open file inside CBZ: %v", err)
		}
		defer rc.Close()

		// Decode the image (it could be either PNG or JPEG)
		var img image.Image
		if strings.HasSuffix(file.Name, ".png") {
			img, err = png.Decode(rc)
			img, err = compressPNG(img, binThreshold, resizeWidth, resizeHeight, max_inner_rects, min_content_percent)
		} else if strings.HasSuffix(file.Name, ".jpg") || strings.HasSuffix(file.Name, ".jpeg") {
			img, err = jpeg.Decode(rc)
			img = rgbaToGray(img)
		} else {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("failed to decode image: %v", err)
		}

		// Save the modified image to the temporary directory
		modifiedImagePath := filepath.Join(tempDir, file.Name)
		modifiedImageFile, err := os.Create(modifiedImagePath)
		if err != nil {
			return "", fmt.Errorf("failed to create modified image file: %v", err)
		}
		defer modifiedImageFile.Close()

		// Encode the binarized image and save it as PNG or JPEG
		if strings.HasSuffix(file.Name, ".png") {
			err = png.Encode(modifiedImageFile, img)
		} else if strings.HasSuffix(file.Name, ".jpg") || strings.HasSuffix(file.Name, ".jpeg") {
			err = jpeg.Encode(modifiedImageFile, img, &jpeg.Options{Quality: 20})
		}
		if err != nil {
			return "", fmt.Errorf("failed to encode modified image: %v", err)
		}

		// Add the modified image back to the new CBZ archive
		zipFile, err := zipWriter.Create(file.Name)
		if err != nil {
			return "", fmt.Errorf("failed to create file in new CBZ: %v", err)
		}

		// Open the modified image and copy it to the zip file
		modifiedImageFile, err = os.Open(modifiedImagePath)
		if err != nil {
			return "", fmt.Errorf("failed to open modified image for zipping: %v", err)
		}
		_, err = io.Copy(zipFile, modifiedImageFile)
		if err != nil {
			return "", fmt.Errorf("failed to copy modified image to zip: %v", err)
		}
	}

	return outputCBZPath, nil
}
