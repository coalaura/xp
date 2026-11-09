package window

import (
	"runtime"
	"syscall"
	"unsafe"

	"github.com/coalaura/xp/theme"
)

func run(options Options) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Scope DPI awareness to this UI thread instead of changing the host process.
	dpiContext := uintptr(0)

	if setThreadDPIAwarenessContext.Find() == nil {
		dpiContext, _, _ = setThreadDPIAwarenessContext.Call(^uintptr(3))
	}

	if dpiContext != 0 {
		defer setThreadDPIAwarenessContext.Call(dpiContext)
	}

	classOnce.Do(registerWindowClass)

	if classError != nil {
		return classError
	}

	title, err := syscall.UTF16FromString(options.Title)
	if err != nil {
		return err
	}

	app := application{
		options: options,
		title:   title,
		dpi:     systemDPI(),
		state:   theme.State{Active: true},
	}

	defer app.buffer.close()

	defer app.deleteFonts()

	app.initializeControls()

	err = app.create()
	if err != nil {
		return err
	}

	defer func() {
		if app.handle != 0 {
			destroyWindow.Call(app.handle)
		}
	}()

	showWindow.Call(app.handle, 1)

	var message nativeMessage

	for {
		result, _, lastError := getMessage.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(result) == -1 {
			return nativeError("GetMessageW", lastError)
		}

		if result == 0 {
			return app.err
		}

		translateMessage.Call(uintptr(unsafe.Pointer(&message)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&message)))
	}
}
