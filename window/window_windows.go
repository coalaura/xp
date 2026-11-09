package window

import (
	"sync"
	"syscall"
	"unsafe"

	"github.com/coalaura/xp/theme"
)

type application struct {
	controlState
	handle           uintptr
	options          Options
	title            []uint16
	dpi              int
	layout           theme.Layout
	state            theme.State
	regionLayout     theme.Layout
	regionMaximized  bool
	tracking         bool
	buffer           drawingBuffer
	font             uintptr
	controlFonts     [2]uintptr
	textBuffer       []uint16
	pendingSurrogate uint16
	wheelDelta       int
	err              error
}

var (
	classOnce      sync.Once
	classError     error
	className, _   = syscall.UTF16PtrFromString("Coalaura.XP.Luna.Window")
	captionFace, _ = syscall.UTF16PtrFromString("Trebuchet MS")
	controlFace, _ = syscall.UTF16PtrFromString("Tahoma")
	instance       uintptr
	arrowCursor    uintptr
	textCursor     uintptr
	linkCursor     uintptr
	applications   sync.Map
	windowCallback = syscall.NewCallback(windowProcedure)
)

func (app *application) invalidate() {
	if app.handle != 0 {
		invalidateRect.Call(app.handle, 0, 0)
	}
}

func (app *application) fail(err error) {
	if app.err == nil {
		app.err = err
	}

	destroyWindow.Call(app.handle)
}

func (app *application) resize() {
	app.closeCombo()

	var rectangle nativeRect

	getClientRect.Call(app.handle, uintptr(unsafe.Pointer(&rectangle)))

	app.layout = theme.NewLayout(int(rectangle.Right), int(rectangle.Bottom), app.dpi, app.options.Style == Dialog)
	maximized, _, _ := isZoomed.Call(app.handle)

	app.state.Maximized = maximized != 0
	if app.state.Maximized {
		app.layout = app.layout.Maximized()
	}

	err := app.updateRegion()
	if err != nil {
		app.fail(err)

		return
	}

	app.invalidate()
}

func (app *application) updateRegion() error {
	if app.regionLayout == app.layout && app.regionMaximized == app.state.Maximized {
		return nil
	}

	// SetWindowRgn can synchronously cause another WM_SIZE; publish the geometry first.
	app.regionLayout = app.layout
	app.regionMaximized = app.state.Maximized

	if app.state.Maximized {
		setWindowRgn.Call(app.handle, 0, 1)

		return nil
	}

	width := app.layout.Width
	height := app.layout.Height
	radius := app.layout.Scale(16)

	region, _, lastError := createRoundRectRgn.Call(0, 0, uintptr(width+1), uintptr(height+1), uintptr(radius), uintptr(radius))
	if region == 0 {
		return nativeError("CreateRoundRectRgn", lastError)
	}

	bottom, _, lastError := createRectRgn.Call(0, uintptr(app.layout.Scale(8)), uintptr(width), uintptr(height))
	if bottom == 0 {
		deleteObject.Call(region)

		return nativeError("CreateRectRgn", lastError)
	}

	result, _, lastError := combineRgn.Call(region, region, bottom, 2)
	deleteObject.Call(bottom)

	if result == 0 {
		deleteObject.Call(region)

		return nativeError("CombineRgn", lastError)
	}

	result, _, lastError = setWindowRgn.Call(app.handle, region, 1)
	if result == 0 {
		deleteObject.Call(region)

		return nativeError("SetWindowRgn", lastError)
	}

	// Windows owns the region after a successful SetWindowRgn call.
	return nil
}

func (app *application) setMinimumSize(info *minMaxInfo) {
	if app.options.Style == Dialog {
		info.MinTrackSize.X = int32(app.layout.Scale(app.options.Width))
		info.MinTrackSize.Y = int32(app.layout.Scale(app.options.Height))
		info.MaxTrackSize = info.MinTrackSize

		return
	}

	info.MinTrackSize.X = int32(app.layout.Scale(app.options.MinWidth))
	info.MinTrackSize.Y = int32(app.layout.Scale(app.options.MinHeight))

	monitor, _, _ := monitorFromWindow.Call(app.handle, 2)
	details := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}

	result, _, _ := getMonitorInfo.Call(monitor, uintptr(unsafe.Pointer(&details)))
	if result == 0 {
		return
	}

	info.MaxPosition.X = details.Work.Left - details.Monitor.Left
	info.MaxPosition.Y = details.Work.Top - details.Monitor.Top
	info.MaxSize.X = details.Work.Right - details.Work.Left
	info.MaxSize.Y = details.Work.Bottom - details.Work.Top
	info.MaxTrackSize.X = max(info.MaxTrackSize.X, info.MaxSize.X)
	info.MaxTrackSize.Y = max(info.MaxTrackSize.Y, info.MaxSize.Y)
}

func (app *application) changeDPI(dpi int, suggested *nativeRect) {
	app.dpi = dpi
	app.layout.DPI = dpi

	app.deleteFonts()

	flags := uintptr(setPositionNoZOrder | setPositionNoActivate)
	setWindowPos.Call(app.handle, 0, uintptr(suggested.Left), uintptr(suggested.Top), uintptr(suggested.Right-suggested.Left), uintptr(suggested.Bottom-suggested.Top), flags)

	app.resize()
}

func (app *application) create() error {
	app.layout = theme.NewLayout(0, 0, app.dpi, app.options.Style == Dialog)

	width := app.layout.Scale(app.options.Width)
	height := app.layout.Scale(app.options.Height)

	work := nativeRect{Right: 1024, Bottom: 768}
	systemParametersInfo.Call(0x0030, 0, uintptr(unsafe.Pointer(&work)), 0)

	left := int(work.Left) + max(0, (int(work.Right-work.Left)-width)/2)
	top := int(work.Top) + max(0, (int(work.Bottom-work.Top)-height)/2)

	style := uintptr(windowStyle)

	if app.options.Style == Dialog {
		style = 0x00C80000 // Caption and system menu, without resizing or min/max buttons.
	}

	created, _, lastError := createWindowEx.Call(0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(&app.title[0])), style, uintptr(left), uintptr(top), uintptr(width), uintptr(height), 0, 0, instance, 0)
	if created == 0 {
		return nativeError("CreateWindowExW", lastError)
	}

	app.handle = created
	applications.Store(created, app)

	if getDPIForWindow.Find() == nil {
		dpi, _, _ := getDPIForWindow.Call(created)
		if dpi != 0 {
			app.dpi = int(dpi)
		}
	}

	// Suppress modern non-client decoration, including Windows 11 corner rounding.
	if dwmSetWindowAttribute.Find() == nil {
		disabled := int32(1)
		dwmSetWindowAttribute.Call(created, 2, uintptr(unsafe.Pointer(&disabled)), 4)
		dwmSetWindowAttribute.Call(created, 33, uintptr(unsafe.Pointer(&disabled)), 4)
	}

	app.layout.DPI = app.dpi

	width = app.layout.Scale(app.options.Width)
	height = app.layout.Scale(app.options.Height)

	left = int(work.Left) + max(0, (int(work.Right-work.Left)-width)/2)
	top = int(work.Top) + max(0, (int(work.Bottom-work.Top)-height)/2)

	flags := uintptr(setPositionNoZOrder | setPositionNoActivate | setPositionFrameChanged)

	result, _, lastError := setWindowPos.Call(created, 0, uintptr(left), uintptr(top), uintptr(width), uintptr(height), flags)
	if result == 0 {
		destroyWindow.Call(created)

		return nativeError("SetWindowPos", lastError)
	}

	app.resize()

	return app.err
}

func registerWindowClass() {
	instance, _, _ = getModuleHandle.Call(0)
	arrowCursor, _, _ = loadCursor.Call(0, 32512)
	textCursor, _, _ = loadCursor.Call(0, 32513)
	linkCursor, _, _ = loadCursor.Call(0, 32649)
	icon, _, _ := loadIcon.Call(0, 32512)

	class := windowClass{
		Size:      uint32(unsafe.Sizeof(windowClass{})),
		Style:     0x0008,
		Procedure: windowCallback,
		Instance:  instance,
		Icon:      icon,
		Cursor:    arrowCursor,
		ClassName: className,
		SmallIcon: icon,
	}

	result, _, lastError := registerClassEx.Call(uintptr(unsafe.Pointer(&class)))
	if result == 0 {
		classError = nativeError("RegisterClassExW", lastError)
	}
}

func windowProcedure(handle uintptr, message uint32, wparam uintptr, lparam unsafe.Pointer) uintptr {
	switch message {
	case wmNCCalcSize, wmNCPaint:
		return 0
	case wmNCActivate, wmEraseBackground:
		return 1
	}

	value, found := applications.Load(handle)
	if !found {
		result, _, _ := defWindowProc.Call(handle, uintptr(message), wparam, uintptr(lparam))

		return result
	}

	app := value.(*application)

	switch message {
	case wmPaint:
		err := app.paint()
		if err != nil {
			app.fail(err)
		}

		return 0
	case wmSize:
		if wparam != 1 {
			app.resize()
		}

		return 0
	case wmActivate:
		app.state.Active = uint16(wparam) != 0
		if !app.state.Active {
			app.cancelPress()
			app.closeCombo()
		}

		app.invalidate()
	case wmGetMinMaxInfo:
		app.setMinimumSize((*minMaxInfo)(lparam))

		return 0
	case wmDPIChanged:
		app.changeDPI(int(uint16(wparam)), (*nativeRect)(lparam))

		return 0
	case wmNCHitTest:
		return app.hitTest(uintptr(lparam))
	case wmNCLButtonDown, wmNCLButtonDblClk:
		app.cancelPress()
		app.clearControlFocus()
	case wmSetCursor:
		if uint16(uintptr(lparam)) == hitClient {
			setCursor.Call(app.clientCursor())

			return 1
		}
	case wmMouseMove:
		app.mouseMove(uintptr(lparam))

		return 0
	case wmMouseLeave:
		app.tracking = false
		app.state.Hot = theme.NoButton
		app.controlHot = -1

		app.invalidate()

		return 0
	case wmLButtonDown, wmLButtonDblClk:
		// CS_DBLCLKS replaces the second mouse-down with a double-click message.
		app.mouseDown(uintptr(lparam))

		return 0
	case wmLButtonUp:
		app.mouseUp(uintptr(lparam))

		return 0
	case wmCancelMode:
		app.cancelPress()
		app.closeCombo()
	case wmCaptureChanged:
		app.state.Pressed = theme.NoButton
		app.controlPressed = -1
		app.comboDragging = false
		app.keyboardPress = false

		app.invalidate()
	case wmKeyDown:
		app.keyDown(wparam, uintptr(lparam))

		return 0
	case wmSysKeyDown, wmSysKeyUp:
		if app.controlFocus >= 0 && (wparam == 0x26 || wparam == 0x28) {
			if message == wmSysKeyDown {
				app.keyDown(wparam, uintptr(lparam))
			}

			return 0
		}
	case wmChar:
		app.character(uint16(wparam))

		return 0
	case wmMouseWheel:
		app.mouseWheel(wparam, uintptr(lparam))

		return 0
	case wmKeyUp:
		app.keyUp(wparam)

		return 0
	case wmSetFocus:
		app.invalidate()
	case wmKillFocus:
		app.cancelPress()
		app.closeCombo()

		app.showFocus = false
		app.pendingSurrogate = 0

		app.invalidate()
	case wmClose:
		destroyWindow.Call(handle)

		return 0
	case wmDestroy:
		postQuitMessage.Call(0)

		return 0
	case wmNCDestroy:
		applications.Delete(handle)

		app.handle = 0
	}

	result, _, _ := defWindowProc.Call(handle, uintptr(message), wparam, uintptr(lparam))
	return result
}

func systemDPI() int {
	if getDPIForSystem.Find() == nil {
		dpi, _, _ := getDPIForSystem.Call()
		if dpi != 0 {
			return int(dpi)
		}
	}

	return 96
}
