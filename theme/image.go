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

	for vertical := clipped.Min.Y; vertical < clipped.Max.Y; vertical++ {
		row := bounds.Min.Y + (vertical-destination.Min.Y)*bounds.Dy()/destination.Dy()

		for horizontal := clipped.Min.X; horizontal < clipped.Max.X; horizontal++ {
			column := bounds.Min.X + (horizontal-destination.Min.X)*bounds.Dx()/destination.Dx()
			offset := source.PixOffset(column, row)

			color := uint32(source.Pix[offset])<<16 | uint32(source.Pix[offset+1])<<8 | uint32(source.Pix[offset+2])
			alpha := int(source.Pix[offset+3])

			pixel := vertical*surface.Stride + horizontal

			surface.Pixels[pixel] = blend(surface.Pixels[pixel], color, alpha)
		}
	}
}
