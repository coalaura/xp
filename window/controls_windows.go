package window

import (
	"image"
	"unicode/utf16"
	"unsafe"

	"github.com/coalaura/xp/theme"
)

func (app *application) Close() {
	postMessage.Call(app.handle, wmClose, 0, 0)
}

func (app *application) controlSurface() theme.Surface {
	return app.buffer.surface
}

func (app *application) prepareControlText() error {
	for index := range app.controlFonts {
		if app.controlFonts[index] != 0 {
			continue
		}

		weight := uintptr(400 + index*300)
		height := -int32(app.layout.Scale(11))

		font, _, lastError := createFont.Call(uintptr(height), 0, 0, 0, weight, 0, 0, 0, 1, 0, 0, 3, 0, uintptr(unsafe.Pointer(controlFace)))
		if font == 0 {
			return nativeError("CreateFontW", lastError)
		}

		app.controlFonts[index] = font
	}

	return nil
}

func (app *application) deleteFonts() {
	if app.font != 0 {
		deleteObject.Call(app.font)
		app.font = 0
	}

	for index, font := range app.controlFonts {
		if font != 0 {
			deleteObject.Call(font)
			app.controlFonts[index] = 0
		}
	}
}

func (app *application) encodeText(text string) {
	app.textBuffer = app.textBuffer[:0]

	for _, current := range text {
		app.textBuffer = utf16.AppendRune(app.textBuffer, current)
	}

	app.textBuffer = append(app.textBuffer, 0)
}

func (app *application) textWidth(text string, bold bool) int {
	if text == "" {
		return 0
	}

	err := app.prepareControlText()
	if err != nil {
		app.fail(err)

		return 0
	}

	index := 0

	if bold {
		index = 1
	}

	app.encodeText(text)

	// Measuring can precede the first paint, when the DIB DC does not exist yet.
	dc, _, _ := getDC.Call(app.handle)
	previous, _, _ := selectObject.Call(dc, app.controlFonts[index])
	bounds := nativeRect{}

	drawText.Call(dc, uintptr(unsafe.Pointer(&app.textBuffer[0])), uintptr(len(app.textBuffer)-1), uintptr(unsafe.Pointer(&bounds)), 0xC20)
	selectObject.Call(dc, previous)
	releaseDC.Call(app.handle, dc)

	return int(bounds.Right - bounds.Left)
}

func (app *application) textHeight(bold bool) int {
	return app.layout.Scale(13)
}

func (app *application) drawControlString(surface theme.Surface, text string, bounds, clip image.Rectangle, foreground uint32, bold bool, mode textMode) {
	clip = clip.Intersect(app.layout.Content)
	if text == "" || bounds.Empty() || clip.Empty() {
		return
	}

	index := 0

	if bold {
		index = 1
	}

	dc := app.buffer.dc

	saved, _, _ := saveDC.Call(dc)
	if saved == 0 {
		return
	}

	selectObject.Call(dc, app.controlFonts[index])
	setBackgroundMode.Call(dc, 1)
	setTextColor.Call(dc, colorRef(foreground))
	intersectClipRect.Call(dc, uintptr(clip.Min.X), uintptr(clip.Min.Y), uintptr(clip.Max.X), uintptr(clip.Max.Y))

	app.encodeText(text)

	flags := uintptr(0x8824)

	switch mode {
	case textWrapped:
		flags = 0x810
	case textCentered:
		flags = 0x8825
	case textUnclipped:
		flags = 0x924 // Single line without ellipsis; the DC owns the viewport clip.
	}

	rectangle := nativeRect{Left: int32(bounds.Min.X), Top: int32(bounds.Min.Y), Right: int32(bounds.Max.X), Bottom: int32(bounds.Max.Y)}

	drawText.Call(dc, uintptr(unsafe.Pointer(&app.textBuffer[0])), uintptr(len(app.textBuffer)-1), uintptr(unsafe.Pointer(&rectangle)), flags)
	restoreDC.Call(dc, saved)
	// Subsequent control surfaces and selections write directly to the DIB.
	gdiFlush.Call()
}
