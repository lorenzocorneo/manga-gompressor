package mangagompressor

import (
	"fmt"
	"image"

	// "image/color"
	// "image/color"

	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestDetecInnerBorders(t *testing.T) {

	rc, _ := os.Open("../test/7.png")

	// if err != nil {
	// 	return "", fmt.Errorf("failed to open file inside CBZ: %v", err)
	// }
	defer rc.Close()

	// Decode the image (it could be either PNG or JPEG)
	var img image.Image
	img, _ = png.Decode(rc)
	img, _ = binarizeImage(img, 65)
	rects := detectInnerBorders(img)
	for _, rect := range rects {
		img = removeRectangle(img, rect, color.RGBA{255, 0, 0, 150})
	}

	remaining := getRemainingRectangles(img, rects)
	for _, rect := range remaining {
		img = removeRectangle(img, rect, color.RGBA{0, 255, 0, 150})
	}
	fmt.Println(remaining)
	// img = assembleImageFromRectangles(img, remaining)
	modifiedImageFile, _ := os.Create("/tmp/test.png")
	defer modifiedImageFile.Close()

	png.Encode(modifiedImageFile, img)
}
