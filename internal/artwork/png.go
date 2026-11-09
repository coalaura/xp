// Package artwork decodes embedded artwork once, before any windows are drawn.
package artwork

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
)

// Decode validates an embedded atlas and converts it to straight-alpha RGBA.
func Decode(data []byte, width, height int) *image.NRGBA {
	decoded, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}

	if decoded.Bounds() != image.Rect(0, 0, width, height) {
		panic("xp: unexpected embedded artwork dimensions")
	}

	if pixels, ok := decoded.(*image.NRGBA); ok {
		return pixels
	}

	pixels := image.NewNRGBA(decoded.Bounds())

	draw.Draw(pixels, pixels.Bounds(), decoded, image.Point{}, draw.Src)

	return pixels
}
