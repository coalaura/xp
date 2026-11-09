package window

import (
	"image"
	"strings"
	"unicode/utf8"

	"github.com/coalaura/xp/theme"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type linuxFonts struct {
	regular font.Face
	bold    font.Face
	title   font.Face
}

func (fonts *linuxFonts) initialize(dpi int) error {
	regular, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return err
	}

	bold, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return err
	}

	options := opentype.FaceOptions{Size: 8.25, DPI: float64(dpi), Hinting: font.HintingFull}

	fonts.regular, err = opentype.NewFace(regular, &options)
	if err != nil {
		return err
	}

	fonts.bold, err = opentype.NewFace(bold, &options)
	if err != nil {
		fonts.close()

		return err
	}

	options.Size = 9.75

	fonts.title, err = opentype.NewFace(bold, &options)
	if err != nil {
		fonts.close()

		return err
	}

	return nil
}

func (fonts *linuxFonts) close() {
	if fonts.regular != nil {
		fonts.regular.Close()
	}

	if fonts.bold != nil {
		fonts.bold.Close()
	}

	if fonts.title != nil {
		fonts.title.Close()
	}
}

func drawSingleLine(surface theme.Surface, face font.Face, text string, bounds, clip image.Rectangle, foreground uint32, centered bool) {
	if bounds.Empty() || clip.Empty() || text == "" {
		return
	}

	text, _, _ = strings.Cut(text, "\n")
	text = strings.TrimSuffix(text, "\r")
	width := font.MeasureString(face, text).Ceil()
	ellipsis := width > bounds.Dx()

	if ellipsis {
		available := max(0, bounds.Dx()-font.MeasureString(face, "...").Ceil())
		text = text[:textPrefix(face, text, available)]
		width = font.MeasureString(face, text).Ceil() + font.MeasureString(face, "...").Ceil()
	}

	left := bounds.Min.X

	if centered {
		left += max(0, (bounds.Dx()-width)/2)
	}

	metrics := face.Metrics()
	height := (metrics.Ascent + metrics.Descent).Ceil()
	baseline := bounds.Min.Y + (bounds.Dy()-height)/2 + metrics.Ascent.Ceil()

	dot := fixed.P(left, baseline)
	dot = drawGlyphs(surface, face, text, dot, clip, foreground)

	if ellipsis {
		drawGlyphs(surface, face, "...", dot, clip, foreground)
	}
}

func drawWrappedText(surface theme.Surface, face font.Face, text string, bounds, clip image.Rectangle, foreground uint32) {
	if bounds.Empty() || clip.Empty() {
		return
	}

	metrics := face.Metrics()
	lineHeight := max(1, metrics.Height.Ceil())
	baseline := bounds.Min.Y + metrics.Ascent.Ceil()

	for paragraph := range strings.SplitSeq(text, "\n") {
		paragraph = strings.TrimSuffix(paragraph, "\r")

		for {
			if baseline-metrics.Ascent.Ceil() >= clip.Max.Y {
				return
			}

			end := textPrefix(face, paragraph, bounds.Dx())
			if end < len(paragraph) {
				space := strings.LastIndexAny(paragraph[:end], " \t")
				if space > 0 {
					end = space
				}
			}

			if end == 0 && paragraph != "" {
				_, end = utf8.DecodeRuneInString(paragraph)
			}

			drawGlyphs(surface, face, paragraph[:end], fixed.P(bounds.Min.X, baseline), clip, foreground)

			baseline += lineHeight

			paragraph = strings.TrimLeft(paragraph[end:], " \t")
			if paragraph == "" {
				break
			}
		}
	}
}

func textPrefix(face font.Face, text string, width int) int {
	limit := fixed.I(width)

	var (
		advance  fixed.Int26_6
		previous rune
	)

	for index, current := range text {
		if index > 0 {
			advance += face.Kern(previous, current)
		}

		glyphAdvance, _ := face.GlyphAdvance(current)

		advance += glyphAdvance
		if advance > limit {
			return index
		}

		previous = current
	}

	return len(text)
}

func drawGlyphs(surface theme.Surface, face font.Face, text string, dot fixed.Point26_6, clip image.Rectangle, foreground uint32) fixed.Point26_6 {
	clip = clip.Intersect(image.Rect(0, 0, surface.Width, surface.Height))

	var previous rune

	for index, current := range text {
		if index > 0 {
			dot.X += face.Kern(previous, current)
		}

		rectangle, mask, maskPoint, advance, _ := face.Glyph(dot, current)
		dot.X += advance
		previous = current

		visible := rectangle.Intersect(clip)
		if visible.Empty() || mask == nil {
			continue
		}

		maskPoint = maskPoint.Add(visible.Min.Sub(rectangle.Min))
		alpha, isAlpha := mask.(*image.Alpha)

		for row := visible.Min.Y; row < visible.Max.Y; row++ {
			for column := visible.Min.X; column < visible.Max.X; column++ {
				maskX := maskPoint.X + column - visible.Min.X
				maskY := maskPoint.Y + row - visible.Min.Y

				var coverage uint32

				if isAlpha {
					coverage = uint32(alpha.Pix[alpha.PixOffset(maskX, maskY)])
				} else {
					_, _, _, coverage = mask.At(maskX, maskY).RGBA()
					coverage >>= 8
				}

				offset := row*surface.Stride + column
				surface.Pixels[offset] = blendText(surface.Pixels[offset], foreground, coverage)
			}
		}
	}

	return dot
}

func blendText(background, foreground, alpha uint32) uint32 {
	inverse := 255 - alpha
	red := ((background>>16&255)*inverse + (foreground>>16&255)*alpha + 127) / 255
	green := ((background>>8&255)*inverse + (foreground>>8&255)*alpha + 127) / 255
	blue := ((background&255)*inverse + (foreground&255)*alpha + 127) / 255

	return red<<16 | green<<8 | blue
}
