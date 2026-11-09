package window

import (
	"errors"
	"image"
	"math/bits"
	"slices"

	"github.com/coalaura/xp/theme"
	"github.com/jezek/xgb/xproto"
)

type pixelEncoder struct {
	bytesPerPixel int
	padBytes      int
	byteOrder     byte
	red           [256]uint32
	green         [256]uint32
	blue          [256]uint32
	buffer        []byte
}

func (app *application) paint() error {
	width := app.layout.Width
	height := app.layout.Height
	count := width * height

	app.surface.Pixels = slices.Grow(app.surface.Pixels[:0], count)[:count]
	app.surface.Width = width
	app.surface.Height = height
	app.surface.Stride = width

	theme.Draw(app.surface, app.layout, app.state)

	app.drawTitle()

	err := app.drawControls()
	if err != nil {
		return err
	}

	stride := app.pixels.stride(width)
	maximumPayload := int(xproto.Setup(app.connection).MaximumRequestLength)*4 - 27

	rowsPerRequest := maximumPayload / stride
	if rowsPerRequest == 0 {
		return errors.New("xp: X11 request limit is too small for a rendered row")
	}

	// Reuse a request-sized staging buffer instead of duplicating the whole framebuffer.
	rowsPerRequest = min(rowsPerRequest, height)
	bufferSize := stride * rowsPerRequest

	app.pixels.buffer = slices.Grow(app.pixels.buffer[:0], bufferSize)[:bufferSize]

	for top := 0; top < height; top += rowsPerRequest {
		rows := min(rowsPerRequest, height-top)
		encoded := app.pixels.buffer[:stride*rows]

		app.pixels.encode(encoded, app.surface.Pixels[top*width:(top+rows)*width], width, rows)

		xproto.PutImage(app.connection, xproto.ImageFormatZPixmap, xproto.Drawable(app.handle), app.gc, uint16(width), uint16(rows), 0, int16(top), 0, app.screen.RootDepth, encoded)
	}

	return nil
}

func (app *application) drawTitle() {
	rectangle := app.layout.Title
	foreground := theme.InactiveTitleText

	if app.state.Active {
		foreground = theme.TitleText
		shadow := rectangle.Add(image.Pt(app.layout.Scale(1), app.layout.Scale(1)))

		drawSingleLine(app.surface, app.fonts.title, app.options.Title, shadow, shadow, 0x0F1089, false)
	}

	drawSingleLine(app.surface, app.fonts.title, app.options.Title, rectangle, rectangle, foreground, false)
}

func (encoder *pixelEncoder) initialize(setup *xproto.SetupInfo, screen *xproto.ScreenInfo) error {
	var visual *xproto.VisualInfo

	for depthIndex := range screen.AllowedDepths {
		depth := &screen.AllowedDepths[depthIndex]

		for visualIndex := range depth.Visuals {
			candidate := &depth.Visuals[visualIndex]
			if candidate.VisualId == screen.RootVisual {
				visual = candidate
			}
		}
	}

	if visual == nil || visual.Class != xproto.VisualClassTrueColor {
		return errors.New("xp: X11 requires a TrueColor root visual")
	}

	var format *xproto.Format

	for index := range setup.PixmapFormats {
		candidate := &setup.PixmapFormats[index]
		if candidate.Depth == screen.RootDepth {
			format = candidate

			break
		}
	}

	if format == nil || (format.BitsPerPixel != 16 && format.BitsPerPixel != 24 && format.BitsPerPixel != 32) {
		return errors.New("xp: unsupported X11 pixel format")
	}

	encoder.bytesPerPixel = int(format.BitsPerPixel) / 8
	encoder.padBytes = max(1, int(format.ScanlinePad)/8)
	encoder.byteOrder = setup.ImageByteOrder

	for value := range 256 {
		encoder.red[value] = maskedComponent(byte(value), visual.RedMask)
		encoder.green[value] = maskedComponent(byte(value), visual.GreenMask)
		encoder.blue[value] = maskedComponent(byte(value), visual.BlueMask)
	}

	return nil
}

func (encoder *pixelEncoder) stride(width int) int {
	rowBytes := width * encoder.bytesPerPixel

	return (rowBytes + encoder.padBytes - 1) / encoder.padBytes * encoder.padBytes
}

func (encoder *pixelEncoder) encode(destination []byte, source []uint32, width, height int) {
	stride := encoder.stride(width)

	for row := range height {
		for column := range width {
			color := source[row*width+column]
			pixel := encoder.red[byte(color>>16)] | encoder.green[byte(color>>8)] | encoder.blue[byte(color)]
			offset := row*stride + column*encoder.bytesPerPixel

			for index := range encoder.bytesPerPixel {
				shift := index * 8

				if encoder.byteOrder == xproto.ImageOrderMSBFirst {
					shift = (encoder.bytesPerPixel - index - 1) * 8
				}

				destination[offset+index] = byte(pixel >> shift)
			}
		}

		clear(destination[row*stride+width*encoder.bytesPerPixel : (row+1)*stride])
	}
}

func maskedComponent(value byte, mask uint32) uint32 {
	if mask == 0 {
		return 0
	}

	shift := bits.TrailingZeros32(mask)
	maximum := uint64(1<<bits.OnesCount32(mask)) - 1
	scaled := (uint64(value)*maximum + 127) / 255

	return uint32(scaled<<shift) & mask
}
