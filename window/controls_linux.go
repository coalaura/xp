package window

import (
	"image"

	"github.com/coalaura/xp/theme"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func (app *application) Close() {
	app.closed = true
}

func (app *application) controlSurface() theme.Surface {
	return app.surface
}

func (app *application) invalidate() {
	app.dirty = true
}

func (app *application) prepareControlText() error {
	return nil
}

func (app *application) textWidth(text string, bold bool) int {
	face := app.fonts.regular

	if bold {
		face = app.fonts.bold
	}

	return font.MeasureString(face, text).Ceil()
}

func (app *application) textHeight(bold bool) int {
	face := app.fonts.regular

	if bold {
		face = app.fonts.bold
	}

	metrics := face.Metrics()

	return (metrics.Ascent + metrics.Descent).Ceil()
}

func (app *application) drawControlString(surface theme.Surface, text string, bounds, clip image.Rectangle, foreground uint32, bold bool, mode textMode) {
	clip = clip.Intersect(app.layout.Content)
	if bounds.Empty() || clip.Empty() {
		return
	}

	face := app.fonts.regular

	if bold {
		face = app.fonts.bold
	}

	switch mode {
	case textWrapped:
		drawWrappedText(surface, face, text, bounds, clip, foreground)
	case textUnclipped:
		metrics := face.Metrics()
		height := (metrics.Ascent + metrics.Descent).Ceil()
		baseline := bounds.Min.Y + (bounds.Dy()-height)/2 + metrics.Ascent.Ceil()

		drawGlyphs(surface, face, text, fixed.P(bounds.Min.X, baseline), clip, foreground)
	default:
		drawSingleLine(surface, face, text, bounds, clip, foreground, mode == textCentered)
	}
}
