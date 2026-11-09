package theme

import "image"

// DrawImage scales a region of straight-alpha artwork using nearest-neighbor
// sampling. Clipping never changes its scale or origin, and drawing allocates nothing.
func (surface Surface) DrawImage(destination image.Rectangle, source *image.NRGBA, bounds image.Rectangle) {
	if destination.Empty() || bounds.Empty() || !bounds.In(source.Rect) {
		return
	}

	clipped := destination.Intersect(image.Rect(0, 0, surface.Width, surface.Height))

	if surface.clipped {
		clipped = clipped.Intersect(surface.clip)
	}

	// Sample pixel centers so fractional scaling distributes repeated pixels
	// symmetrically instead of duplicating the top and left edges first.
	for vertical := clipped.Min.Y; vertical < clipped.Max.Y; vertical++ {
		row := bounds.Min.Y + (2*(vertical-destination.Min.Y)+1)*bounds.Dy()/(2*destination.Dy())

		for horizontal := clipped.Min.X; horizontal < clipped.Max.X; horizontal++ {
			column := bounds.Min.X + (2*(horizontal-destination.Min.X)+1)*bounds.Dx()/(2*destination.Dx())
			offset := source.PixOffset(column, row)

			color := uint32(source.Pix[offset])<<16 | uint32(source.Pix[offset+1])<<8 | uint32(source.Pix[offset+2])
			alpha := int(source.Pix[offset+3])

			pixel := vertical*surface.Stride + horizontal

			surface.Pixels[pixel] = blend(surface.Pixels[pixel], color, alpha)
		}
	}
}
