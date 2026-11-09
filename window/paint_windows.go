package window

import (
	"unsafe"

	"github.com/coalaura/xp/theme"
)

type drawingBuffer struct {
	dc             uintptr
	bitmap         uintptr
	originalBitmap uintptr
	surface        theme.Surface
}

func (buffer *drawingBuffer) close() {
	if buffer.dc == 0 {
		return
	}

	gdiFlush.Call()

	if buffer.originalBitmap != 0 {
		selectObject.Call(buffer.dc, buffer.originalBitmap)
	}

	if buffer.bitmap != 0 {
		deleteObject.Call(buffer.bitmap)
	}

	deleteDC.Call(buffer.dc)
	*buffer = drawingBuffer{}
}

func (buffer *drawingBuffer) resize(width, height int) error {
	if buffer.surface.Width == width && buffer.surface.Height == height {
		return nil
	}

	if buffer.dc == 0 {
		created, _, lastError := createCompatibleDC.Call(0)
		if created == 0 {
			return nativeError("CreateCompatibleDC", lastError)
		}

		buffer.dc = created
	}

	info := bitmapInfo{
		Size:     uint32(unsafe.Sizeof(bitmapInfo{})),
		Width:    int32(width),
		Height:   -int32(height),
		Planes:   1,
		BitCount: 32,
	}

	var pixels unsafe.Pointer

	bitmap, _, lastError := createDIBSection.Call(buffer.dc, uintptr(unsafe.Pointer(&info)), 0, uintptr(unsafe.Pointer(&pixels)), 0, 0)
	if bitmap == 0 {
		return nativeError("CreateDIBSection", lastError)
	}

	gdiFlush.Call()

	previous, _, lastError := selectObject.Call(buffer.dc, bitmap)
	if previous == 0 || previous == ^uintptr(0) {
		deleteObject.Call(bitmap)

		return nativeError("SelectObject", lastError)
	}

	if buffer.originalBitmap == 0 {
		buffer.originalBitmap = previous
	}

	if buffer.bitmap != 0 {
		deleteObject.Call(buffer.bitmap)
	}

	buffer.bitmap = bitmap

	buffer.surface = theme.Surface{
		// This slice is owned by the DIB and is only used until its bitmap is deleted.
		Pixels: unsafe.Slice((*uint32)(pixels), width*height),
		Width:  width,
		Height: height,
		Stride: width,
	}

	return nil
}

func (app *application) paint() error {
	var paint paintStruct

	dc, _, lastError := beginPaint.Call(app.handle, uintptr(unsafe.Pointer(&paint)))
	if dc == 0 {
		return nativeError("BeginPaint", lastError)
	}

	defer endPaint.Call(app.handle, uintptr(unsafe.Pointer(&paint)))

	width := app.layout.Width
	height := app.layout.Height

	if width <= 0 || height <= 0 {
		return nil
	}

	err := app.buffer.resize(width, height)
	if err != nil {
		return err
	}

	// GDI batches operations; finish its previous reads/writes before touching DIB memory.
	gdiFlush.Call()

	theme.Draw(app.buffer.surface, app.layout, app.state)

	err = app.drawCaption()
	if err != nil {
		return err
	}

	err = app.drawControls()
	if err != nil {
		return err
	}

	result, _, lastError := bitBlt.Call(dc, 0, 0, uintptr(width), uintptr(height), app.buffer.dc, 0, 0, 0x00CC0020)
	if result == 0 {
		return nativeError("BitBlt", lastError)
	}

	return nil
}

func (app *application) drawCaption() error {
	if app.font == 0 {
		height := -int32(app.layout.Scale(13))

		font, _, lastError := createFont.Call(uintptr(height), 0, 0, 0, 700, 0, 0, 0, 1, 0, 0, 4, 0, uintptr(unsafe.Pointer(captionFace)))
		if font == 0 {
			return nativeError("CreateFontW", lastError)
		}

		app.font = font
	}

	dc := app.buffer.dc

	previous, _, lastError := selectObject.Call(dc, app.font)
	if previous == 0 || previous == ^uintptr(0) {
		return nativeError("SelectObject", lastError)
	}

	defer selectObject.Call(dc, previous)

	setBackgroundMode.Call(dc, 1)

	bounds := app.layout.Title

	shadowOffset := app.layout.Scale(1)
	shadow := nativeRect{
		Left:   int32(bounds.Min.X + shadowOffset),
		Top:    int32(bounds.Min.Y + shadowOffset),
		Right:  int32(bounds.Max.X + shadowOffset),
		Bottom: int32(bounds.Max.Y + shadowOffset),
	}

	foreground := theme.TitleText
	shadowColor := uint32(0x0F1089)

	if !app.state.Active {
		foreground = theme.InactiveTitleText
	}

	if app.state.Active {
		setTextColor.Call(dc, colorRef(shadowColor))
		drawText.Call(dc, uintptr(unsafe.Pointer(&app.title[0])), uintptr(len(app.title)-1), uintptr(unsafe.Pointer(&shadow)), 0x8824)
	}

	text := nativeRect{
		Left:   int32(bounds.Min.X),
		Top:    int32(bounds.Min.Y),
		Right:  int32(bounds.Max.X),
		Bottom: int32(bounds.Max.Y),
	}

	setTextColor.Call(dc, colorRef(foreground))
	drawText.Call(dc, uintptr(unsafe.Pointer(&app.title[0])), uintptr(len(app.title)-1), uintptr(unsafe.Pointer(&text)), 0x8824)

	return nil
}

func colorRef(color uint32) uintptr {
	return uintptr((color&255)<<16 | color&0xFF00 | color>>16&255)
}
