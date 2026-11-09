// Command reference extracts text-free Luna artwork from the XP VM captures.
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
)

type crop struct {
	file   string
	bounds image.Rectangle
}

func main() {
	err := generate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generate() error {
	caption := []crop{
		{file: "04", bounds: image.Rect(820, 113, 841, 134)},
		{file: "04", bounds: image.Rect(843, 113, 864, 134)},
		{file: "03", bounds: image.Rect(978, 66, 999, 87)},
		{file: "04", bounds: image.Rect(866, 113, 887, 134)},
	}

	err := writeArtwork("theme/inactive.png", caption, nil)
	if err != nil {
		return err
	}

	maximized := []crop{
		{file: "03", bounds: image.Rect(955, 66, 976, 87)},
		{file: "03", bounds: image.Rect(978, 66, 999, 87)},
		{file: "03", bounds: image.Rect(1001, 66, 1022, 87)},
	}

	err = writeArtwork("theme/inactive_maximized.png", maximized, nil)
	if err != nil {
		return err
	}

	checkbox := []crop{
		{file: "10", bounds: image.Rect(361, 486, 374, 499)},
		{file: "12", bounds: image.Rect(361, 486, 374, 499)},
		{file: "07", bounds: image.Rect(361, 486, 374, 499)},
	}

	err = writeArtwork("ui/checkbox.png", checkbox, nil)
	if err != nil {
		return err
	}

	button := []crop{
		{file: "08", bounds: image.Rect(478, 489, 551, 510)},
		{file: "09", bounds: image.Rect(478, 489, 551, 510)},
	}

	// Keep the ends and a text-free center column for nine-slice rendering.
	buttonColumns := []int{0, 1, 2, 3, 12, 69, 70, 71, 72}

	err = writeArtwork("ui/button.png", button, buttonColumns)
	if err != nil {
		return err
	}

	frame := []crop{
		{file: "01", bounds: image.Rect(93, 137, 893, 707)},
		{file: "04", bounds: image.Rect(93, 137, 893, 707)},
	}

	frameColumns := []int{0, 1, 2, 3, 400, 796, 797, 798, 799}

	err = writeArtwork("theme/frame.png", frame, frameColumns)
	if err != nil {
		return err
	}

	dialog := []crop{
		{file: "10", bounds: image.Rect(304, 355, 685, 516)},
	}

	dialogColumns := []int{0, 1, 2, 190, 378, 379, 380}

	err = writeArtwork("theme/dialog.png", dialog, dialogColumns)
	if err != nil {
		return err
	}

	bottom := []crop{
		{file: "01", bounds: image.Rect(93, 703, 893, 707)},
		{file: "04", bounds: image.Rect(93, 703, 893, 707)},
	}

	err = writeArtwork("theme/bottom.png", bottom, nil)
	if err != nil {
		return err
	}

	dialogBottom := []crop{
		{file: "10", bounds: image.Rect(304, 513, 685, 516)},
	}

	return writeArtwork("theme/dialog_bottom.png", dialogBottom, nil)
}

func writeArtwork(path string, crops []crop, columns []int) error {
	width := crops[0].bounds.Dx()
	height := crops[0].bounds.Dy()

	if columns != nil {
		width = len(columns)
	}

	atlas := image.NewNRGBA(image.Rect(0, 0, width, height*len(crops)))

	for index, source := range crops {
		capture, err := readCapture(source.file)
		if err != nil {
			return err
		}

		if !source.bounds.In(capture.Bounds()) {
			return fmt.Errorf("reference %s: crop outside image", source.file)
		}

		for vertical := range height {
			for horizontal := range width {
				column := horizontal

				if columns != nil {
					column = columns[horizontal]
				}

				atlas.Set(horizontal, index*height+vertical, capture.At(source.bounds.Min.X+column, source.bounds.Min.Y+vertical))
			}
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	err = png.Encode(file, atlas)
	closeError := file.Close()

	if err != nil {
		return err
	}

	return closeError
}

func readCapture(name string) (image.Image, error) {
	file, err := os.Open("references/" + name + ".png")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	return png.Decode(file)
}
