package window

import (
	"errors"
	"fmt"
	"syscall"
)

const (
	wmDestroy         = 0x0002
	wmSize            = 0x0005
	wmActivate        = 0x0006
	wmSetFocus        = 0x0007
	wmKillFocus       = 0x0008
	wmPaint           = 0x000F
	wmClose           = 0x0010
	wmEraseBackground = 0x0014
	wmCancelMode      = 0x001F
	wmGetMinMaxInfo   = 0x0024
	wmSetCursor       = 0x0020
	wmNCDestroy       = 0x0082
	wmNCCalcSize      = 0x0083
	wmNCHitTest       = 0x0084
	wmNCPaint         = 0x0085
	wmNCActivate      = 0x0086
	wmSysCommand      = 0x0112
	wmKeyDown         = 0x0100
	wmKeyUp           = 0x0101
	wmMouseMove       = 0x0200
	wmLButtonDown     = 0x0201
	wmLButtonUp       = 0x0202
	wmCaptureChanged  = 0x0215
	wmMouseLeave      = 0x02A3
	wmDPIChanged      = 0x02E0

	hitClient      = 1
	hitCaption     = 2
	hitSystemMenu  = 3
	hitLeft        = 10
	hitRight       = 11
	hitTop         = 12
	hitTopLeft     = 13
	hitTopRight    = 14
	hitBottom      = 15
	hitBottomLeft  = 16
	hitBottomRight = 17

	systemMinimize = 0xF020
	systemMaximize = 0xF030
	systemClose    = 0xF060
	systemRestore  = 0xF120

	windowStyle             = 0x00CF0000 // Caption, system menu, thick frame, minimize/maximize.
	setPositionNoZOrder     = 0x0004
	setPositionNoActivate   = 0x0010
	setPositionFrameChanged = 0x0020
)

type nativePoint struct {
	X int32
	Y int32
}

type nativeRect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type windowClass struct {
	Size        uint32
	Style       uint32
	Procedure   uintptr
	ClassExtra  int32
	WindowExtra int32
	Instance    uintptr
	Icon        uintptr
	Cursor      uintptr
	Background  uintptr
	MenuName    *uint16
	ClassName   *uint16
	SmallIcon   uintptr
}

type nativeMessage struct {
	Window  uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Point   nativePoint
	Private uint32
}

type paintStruct struct {
	DC                uintptr
	Erase             int32
	Paint             nativeRect
	Restore           int32
	IncrementalUpdate int32
	Reserved          [32]byte
}

type minMaxInfo struct {
	Reserved     nativePoint
	MaxSize      nativePoint
	MaxPosition  nativePoint
	MinTrackSize nativePoint
	MaxTrackSize nativePoint
}

type monitorInfo struct {
	Size    uint32
	Monitor nativeRect
	Work    nativeRect
	Flags   uint32
}

type trackMouseEvent struct {
	Size      uint32
	Flags     uint32
	Window    uintptr
	HoverTime uint32
}

type bitmapInfo struct {
	Size            uint32
	Width           int32
	Height          int32
	Planes          uint16
	BitCount        uint16
	Compression     uint32
	SizeImage       uint32
	XPelsPerMeter   int32
	YPelsPerMeter   int32
	ColorsUsed      uint32
	ColorsImportant uint32
}

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	dwmapi   = syscall.NewLazyDLL("dwmapi.dll")

	beginPaint                   = user32.NewProc("BeginPaint")
	createWindowEx               = user32.NewProc("CreateWindowExW")
	defWindowProc                = user32.NewProc("DefWindowProcW")
	destroyWindow                = user32.NewProc("DestroyWindow")
	dispatchMessage              = user32.NewProc("DispatchMessageW")
	drawText                     = user32.NewProc("DrawTextW")
	endPaint                     = user32.NewProc("EndPaint")
	getClientRect                = user32.NewProc("GetClientRect")
	getDPIForSystem              = user32.NewProc("GetDpiForSystem")
	getDPIForWindow              = user32.NewProc("GetDpiForWindow")
	getMessage                   = user32.NewProc("GetMessageW")
	getKeyState                  = user32.NewProc("GetKeyState")
	getMonitorInfo               = user32.NewProc("GetMonitorInfoW")
	invalidateRect               = user32.NewProc("InvalidateRect")
	isZoomed                     = user32.NewProc("IsZoomed")
	loadCursor                   = user32.NewProc("LoadCursorW")
	loadIcon                     = user32.NewProc("LoadIconW")
	monitorFromWindow            = user32.NewProc("MonitorFromWindow")
	postMessage                  = user32.NewProc("PostMessageW")
	postQuitMessage              = user32.NewProc("PostQuitMessage")
	registerClassEx              = user32.NewProc("RegisterClassExW")
	releaseCapture               = user32.NewProc("ReleaseCapture")
	screenToClient               = user32.NewProc("ScreenToClient")
	setCapture                   = user32.NewProc("SetCapture")
	setFocus                     = user32.NewProc("SetFocus")
	setCursor                    = user32.NewProc("SetCursor")
	setThreadDPIAwarenessContext = user32.NewProc("SetThreadDpiAwarenessContext")
	setWindowPos                 = user32.NewProc("SetWindowPos")
	setWindowRgn                 = user32.NewProc("SetWindowRgn")
	showWindow                   = user32.NewProc("ShowWindow")
	systemParametersInfo         = user32.NewProc("SystemParametersInfoW")
	trackMouse                   = user32.NewProc("TrackMouseEvent")
	translateMessage             = user32.NewProc("TranslateMessage")

	bitBlt             = gdi32.NewProc("BitBlt")
	combineRgn         = gdi32.NewProc("CombineRgn")
	createCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	createDIBSection   = gdi32.NewProc("CreateDIBSection")
	createFont         = gdi32.NewProc("CreateFontW")
	createRectRgn      = gdi32.NewProc("CreateRectRgn")
	createRoundRectRgn = gdi32.NewProc("CreateRoundRectRgn")
	deleteDC           = gdi32.NewProc("DeleteDC")
	deleteObject       = gdi32.NewProc("DeleteObject")
	gdiFlush           = gdi32.NewProc("GdiFlush")
	selectObject       = gdi32.NewProc("SelectObject")
	saveDC             = gdi32.NewProc("SaveDC")
	restoreDC          = gdi32.NewProc("RestoreDC")
	intersectClipRect  = gdi32.NewProc("IntersectClipRect")
	setBackgroundMode  = gdi32.NewProc("SetBkMode")
	setTextColor       = gdi32.NewProc("SetTextColor")

	getModuleHandle       = kernel32.NewProc("GetModuleHandleW")
	dwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
)

func nativeError(operation string, lastError error) error {
	if lastError == nil || errors.Is(lastError, syscall.Errno(0)) {
		return fmt.Errorf("xp: %s failed", operation)
	}

	return fmt.Errorf("xp: %s: %w", operation, lastError)
}
