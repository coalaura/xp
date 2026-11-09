// Command reference extracts text-free Luna artwork from the XP UI state atlas.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

type crop struct {
	file   string
	bounds image.Rectangle
}

type artwork struct {
	path    string
	crops   []crop
	columns []int
}

var artworks = []artwork{
	{
		path: "theme/icon.png",
		crops: []crop{
			{file: "09_frames", bounds: image.Rect(33, 94, 49, 110)},
			{file: "00_inactive", bounds: image.Rect(33, 94, 49, 110)},
		},
	},
	{
		path: "theme/caption.png",
		crops: []crop{
			{file: "09_frames", bounds: image.Rect(923, 92, 944, 113)},
			{file: "08_caption", bounds: image.Rect(492, 296, 513, 317)},
			{file: "08_caption", bounds: image.Rect(676, 296, 697, 317)},
			{file: "09_frames", bounds: image.Rect(946, 92, 967, 113)},
			{file: "08_caption", bounds: image.Rect(492, 368, 513, 389)},
			{file: "08_caption", bounds: image.Rect(676, 368, 697, 389)},
			{file: "08_caption", bounds: image.Rect(308, 512, 329, 533)},
			{file: "08_caption", bounds: image.Rect(492, 512, 513, 533)},
			{file: "08_caption", bounds: image.Rect(676, 512, 697, 533)},
			{file: "09_frames", bounds: image.Rect(969, 92, 990, 113)},
			{file: "08_caption", bounds: image.Rect(492, 404, 513, 425)},
			{file: "08_caption", bounds: image.Rect(676, 404, 697, 425)},
		},
	},
	{
		path: "theme/inactive.png",
		crops: []crop{
			{file: "00_inactive", bounds: image.Rect(923, 92, 944, 113)},
			{file: "00_inactive", bounds: image.Rect(946, 92, 967, 113)},
			{file: "00_inactive", bounds: image.Rect(946, 92, 967, 113)},
			{file: "00_inactive", bounds: image.Rect(969, 92, 990, 113)},
		},
	},
	{
		path: "ui/checkbox.png",
		crops: []crop{
			{file: "01_buttons", bounds: image.Rect(252, 293, 265, 306)},
			{file: "01_buttons", bounds: image.Rect(399, 293, 412, 306)},
			{file: "01_buttons", bounds: image.Rect(546, 293, 559, 306)},
			{file: "01_buttons", bounds: image.Rect(693, 293, 706, 306)},
			{file: "01_buttons", bounds: image.Rect(252, 331, 265, 344)},
			{file: "01_buttons", bounds: image.Rect(399, 331, 412, 344)},
			{file: "01_buttons", bounds: image.Rect(546, 331, 559, 344)},
			{file: "01_buttons", bounds: image.Rect(693, 331, 706, 344)},
		},
	},
	{
		path: "ui/button.png",
		crops: []crop{
			{file: "01_buttons", bounds: image.Rect(244, 249, 356, 275)},
			{file: "01_buttons", bounds: image.Rect(391, 249, 503, 275)},
			{file: "01_buttons", bounds: image.Rect(538, 249, 650, 275)},
			{file: "01_buttons", bounds: image.Rect(685, 249, 797, 275)},
			{file: "01_buttons", bounds: image.Rect(832, 249, 944, 275)},
		},
	},
	{
		path: "theme/frame.png",
		crops: []crop{
			{file: "09_frames", bounds: image.Rect(28, 116, 996, 780)},
			{file: "00_inactive", bounds: image.Rect(28, 116, 996, 780)},
		},
		columns: []int{0, 1, 2, 3, 472, 964, 965, 966, 967},
	},
	{
		path: "theme/dialog.png",
		// The gallery has no native dialog: retain the inner three frame columns.
		crops: []crop{
			{file: "09_frames", bounds: image.Rect(29, 116, 995, 780)},
			{file: "00_inactive", bounds: image.Rect(29, 116, 995, 780)},
		},
		columns: []int{0, 1, 2, 471, 963, 964, 965},
	},
	{
		path: "theme/bottom.png",
		crops: []crop{
			{file: "09_frames", bounds: image.Rect(28, 776, 996, 780)},
			{file: "00_inactive", bounds: image.Rect(28, 776, 996, 780)},
		},
	},
	{
		path: "theme/dialog_bottom.png",
		crops: []crop{
			{file: "09_frames", bounds: image.Rect(29, 777, 995, 780)},
			{file: "00_inactive", bounds: image.Rect(29, 777, 995, 780)},
		},
	},
	{
		path: "theme/title.png",
		crops: []crop{
			{file: "09_frames", bounds: image.Rect(28, 86, 996, 116)},
			{file: "00_inactive", bounds: image.Rect(28, 86, 996, 116)},
		},
		// Preserve the native corner shading and icon, then a text-free center
		// and the right edge beyond the close button.
		columns: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 472, 962, 963, 964, 965, 966, 967},
	},
	{
		path: "theme/title_maximized.png",
		crops: []crop{
			{file: "09_frames", bounds: image.Rect(350, 353, 351, 377)},
			{file: "09_frames", bounds: image.Rect(596, 353, 597, 377)},
		},
	},
}

func main() {
	err := generate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generate() error {
	captures := make(map[string]image.Image, 4)

	for _, specification := range artworks {
		atlas, err := extractArtwork(specification, captures)
		if err != nil {
			return fmt.Errorf("%s: %w", specification.path, err)
		}

		if specification.path == "theme/inactive.png" {
			restoreInactive(atlas, captures)
		}

		if specification.path == "theme/icon.png" {
			atlas = transparentIcon(atlas)
		}

		if specification.path == "ui/button.png" {
			removeButtonLabels(atlas)
		}

		err = writeArtwork(specification.path, atlas)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractArtwork(specification artwork, captures map[string]image.Image) (*image.NRGBA, error) {
	if len(specification.crops) == 0 {
		return nil, fmt.Errorf("no crops")
	}

	size := specification.crops[0].bounds.Size()
	width := size.X

	if specification.columns != nil {
		width = len(specification.columns)
	}

	for _, column := range specification.columns {
		if column < 0 || column >= size.X {
			return nil, fmt.Errorf("column %d outside crop", column)
		}
	}

	atlas := image.NewNRGBA(image.Rect(0, 0, width, size.Y*len(specification.crops)))

	for index, source := range specification.crops {
		capture := captures[source.file]
		if capture == nil {
			var err error

			capture, err = readCapture(source.file)
			if err != nil {
				return nil, err
			}

			captures[source.file] = capture
		}

		if source.bounds.Size() != size || source.bounds.Empty() || !source.bounds.In(capture.Bounds()) {
			return nil, fmt.Errorf("reference %s: invalid crop %v", source.file, source.bounds)
		}

		for vertical := range size.Y {
			for horizontal := range width {
				column := horizontal

				if specification.columns != nil {
					column = specification.columns[horizontal]
				}

				atlas.Set(horizontal, index*size.Y+vertical, capture.At(source.bounds.Min.X+column, source.bounds.Min.Y+vertical))
			}
		}
	}

	return atlas, nil
}

func removeButtonLabels(atlas *image.NRGBA) {
	// Only the flat interior behind the labels repeats horizontally. Keep the
	// full borders and shadows, whose shading differs across the button width.
	for index := range 5 {
		for vertical := 4; vertical < 22; vertical++ {
			row := index*26 + vertical
			pixel := atlas.NRGBAAt(12, row)

			for horizontal := 30; horizontal < 82; horizontal++ {
				atlas.SetNRGBA(horizontal, row, pixel)
			}
		}
	}
}

func transparentIcon(atlas *image.NRGBA) *image.NRGBA {
	icon := image.NewNRGBA(image.Rect(0, 0, 16, 16))

	// The stock icon is unchanged between active and inactive captions; the
	// differing pixels are its transparent background, not part of the icon.
	for vertical := range 16 {
		for horizontal := range 16 {
			pixel := atlas.NRGBAAt(horizontal, vertical)
			if pixel == atlas.NRGBAAt(horizontal, vertical+16) {
				icon.SetNRGBA(horizontal, vertical, pixel)
			}
		}
	}

	return icon
}

func restoreInactive(atlas *image.NRGBA, captures map[string]image.Image) {
	active := captures["09_frames"]
	restore := captures["08_caption"]
	inactive := captures["00_inactive"]

	// No inactive restore is captured. Transfer the normal restore glyph onto
	// the inactive face, matching the paired native buttons on each scanline.
	for vertical := 3; vertical < 18; vertical++ {
		for horizontal := 3; horizontal < 18; horizontal++ {
			pixel := color.NRGBAModel.Convert(restore.At(308+horizontal, 512+vertical)).(color.NRGBA)
			if pixel.R == 255 && pixel.G == 255 && pixel.B == 255 {
				atlas.SetNRGBA(horizontal, 42+vertical, pixel)

				continue
			}

			bestDistance := int(^uint(0) >> 1)

			var matched color.Color

			for candidate := 3; candidate < 18; candidate++ {
				for button := range 2 {
					left := 923 + button*23
					original := color.NRGBAModel.Convert(active.At(left+candidate, 92+vertical)).(color.NRGBA)
					red := int(original.R) - int(pixel.R)
					green := int(original.G) - int(pixel.G)
					blue := int(original.B) - int(pixel.B)
					distance := red*red + green*green + blue*blue

					if distance < bestDistance || distance == bestDistance && candidate == horizontal {
						bestDistance = distance
						matched = inactive.At(left+candidate, 92+vertical)
					}
				}
			}

			atlas.Set(horizontal, 42+vertical, matched)
		}
	}
}

func writeArtwork(path string, atlas image.Image) error {
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
