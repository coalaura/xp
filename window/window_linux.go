package window

import (
	"errors"
	"fmt"
	"runtime"
	"slices"

	"github.com/coalaura/xp/theme"
	"github.com/coalaura/xp/ui"
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

const maxXDimension = 16384

type application struct {
	options        Options
	connection     *xgb.Conn
	screen         *xproto.ScreenInfo
	handle         xproto.Window
	gc             xproto.Gcontext
	atoms          map[string]xproto.Atom
	supported      map[xproto.Atom]bool
	layout         theme.Layout
	state          theme.State
	controls       []ui.Control
	controlHot     int
	controlPressed int
	controlFocus   int
	keyboardPress  bool
	showFocus      bool
	keys           [256]bool
	keysyms        []xproto.Keysym
	keysymsPerKey  int
	firstKeycode   int
	fonts          linuxFonts
	pixels         pixelEncoder
	surface        theme.Surface
	dirty          bool
	closed         bool
	shape          bool
	cursors        [9]xproto.Cursor
	cursor         int
	lastClick      xproto.Timestamp
	lastClickX     int
	lastClickY     int
	drag           windowDrag
}

func (app *application) createWindow() error {
	var err error

	app.handle, err = xproto.NewWindowId(app.connection)
	if err != nil {
		return err
	}

	eventMask := xproto.EventMaskExposure | xproto.EventMaskStructureNotify | xproto.EventMaskPropertyChange | xproto.EventMaskFocusChange | xproto.EventMaskKeyPress | xproto.EventMaskKeyRelease | xproto.EventMaskButtonPress | xproto.EventMaskButtonRelease | xproto.EventMaskPointerMotion | xproto.EventMaskLeaveWindow

	width := app.layout.Width
	height := app.layout.Height

	left := max(0, (int(app.screen.WidthInPixels)-width)/2)
	top := max(0, (int(app.screen.HeightInPixels)-height)/2)

	err = xproto.CreateWindowChecked(app.connection, app.screen.RootDepth, app.handle, app.screen.Root, int16(left), int16(top), uint16(width), uint16(height), 0, xproto.WindowClassInputOutput, app.screen.RootVisual, xproto.CwEventMask, []uint32{uint32(eventMask)}).Check()
	if err != nil {
		return fmt.Errorf("xp: create X11 window: %w", err)
	}

	app.setWindowProperties()

	app.gc, err = xproto.NewGcontextId(app.connection)
	if err != nil {
		return err
	}

	err = xproto.CreateGCChecked(app.connection, app.gc, xproto.Drawable(app.handle), xproto.GcGraphicsExposures, []uint32{0}).Check()
	if err != nil {
		return err
	}

	app.initializeShape()
	app.initializeCursors()

	return xproto.MapWindowChecked(app.connection, app.handle).Check()
}

func (app *application) runEvents() error {
	var event xgb.Event

	for !app.closed {
		if event == nil {
			var eventError xgb.Error

			event, eventError = app.connection.WaitForEvent()
			if eventError != nil {
				return fmt.Errorf("xp: X11 event: %w", eventError)
			}

			if event == nil {
				return errors.New("xp: X11 connection closed")
			}
		}

		// Bound each batch so a busy event queue cannot starve painting.
		for count := 0; event != nil && count < 256; count++ {
			var err error

			event, err = app.dispatchEvent(event)
			if err != nil || app.closed {
				return err
			}
		}

		if app.dirty {
			err := app.paint()
			if err != nil {
				return err
			}

			app.dirty = false
		}
	}

	return nil
}

func (app *application) dispatchEvent(event xgb.Event) (xgb.Event, error) {
	if release, ok := event.(xproto.KeyReleaseEvent); ok {
		// A round trip ensures the paired press of core X11 autorepeat is queued.
		// Compare events rather than physical key state so rapid real taps are preserved.
		app.connection.Sync()

		next, eventError := app.connection.PollForEvent()
		if eventError != nil {
			return nil, eventError
		}

		if press, repeated := next.(xproto.KeyPressEvent); repeated && press.Detail == release.Detail && press.Time == release.Time {
			return app.connection.PollForEvent()
		}

		app.keyUp(release)

		return next, nil
	}

	err := app.handleEvent(event)
	if err != nil || app.closed {
		return nil, err
	}

	return app.connection.PollForEvent()
}

func (app *application) handleEvent(event xgb.Event) error {
	switch event := event.(type) {
	case xproto.ExposeEvent:
		app.dirty = true
	case xproto.ConfigureNotifyEvent:
		if event.Window == app.handle && (int(event.Width) != app.layout.Width || int(event.Height) != app.layout.Height) {
			if event.Width > maxXDimension || event.Height > maxXDimension {
				return errors.New("xp: X11 window exceeds the supported pixel dimensions")
			}

			app.resize(int(event.Width), int(event.Height))
		}
	case xproto.DestroyNotifyEvent:
		if event.Window == app.handle {
			app.closed = true
		}
	case xproto.ClientMessageEvent:
		app.clientMessage(event)
	case xproto.PropertyNotifyEvent:
		if event.Window == app.handle && event.Atom == app.atoms["_NET_WM_STATE"] {
			app.updateWindowState()
		}
	case xproto.FocusInEvent:
		if event.Mode == xproto.NotifyModeNormal && event.Detail != xproto.NotifyDetailInferior {
			app.state.Active = true
			app.dirty = true
		}
	case xproto.FocusOutEvent:
		if event.Mode == xproto.NotifyModeNormal && event.Detail != xproto.NotifyDetailInferior {
			app.state.Active = false

			app.cancelPress()

			clear(app.keys[:])
		}
	case xproto.ButtonPressEvent:
		app.buttonDown(event)
	case xproto.ButtonReleaseEvent:
		app.buttonUp(event)
	case xproto.MotionNotifyEvent:
		app.pointerMotion(event)
	case xproto.LeaveNotifyEvent:
		app.controlHot = -1
		app.state.Hot = theme.NoButton
		app.dirty = true
	case xproto.KeyPressEvent:
		app.keyDown(event)
	case xproto.KeyReleaseEvent:
		app.keyUp(event)
	case xproto.MappingNotifyEvent:
		if event.Request == xproto.MappingKeyboard {
			return app.loadKeyboard()
		}
	}

	return nil
}

func (app *application) resize(width, height int) {
	app.layout = theme.NewLayout(width, height, app.layout.DPI, app.options.Style == Dialog)
	if app.state.Maximized {
		app.layout = app.layout.Maximized()
	}

	app.updateShape()

	app.dirty = true
}

func run(options Options) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	connection, err := xgb.NewConn()
	if err != nil {
		return fmt.Errorf("xp: connect to X11 (Wayland sessions require XWayland and DISPLAY): %w", err)
	}

	defer connection.Close()

	setup := xproto.Setup(connection)

	screen := setup.DefaultScreen(connection)
	if screen == nil {
		return errors.New("xp: no default X11 screen")
	}

	dpi := linuxDPI(connection, screen)
	layout := theme.NewLayout(options.Width, options.Height, dpi, options.Style == Dialog)

	width := layout.Scale(options.Width)
	height := layout.Scale(options.Height)

	if width > maxXDimension || height > maxXDimension {
		return errors.New("xp: scaled window dimensions exceed 16384 pixels")
	}

	app := &application{
		options:        options,
		connection:     connection,
		screen:         screen,
		layout:         theme.NewLayout(width, height, dpi, options.Style == Dialog),
		controls:       slices.Clone(options.Controls),
		controlHot:     -1,
		controlPressed: -1,
		controlFocus:   -1,
		cursor:         -1,
		dirty:          true,
	}

	err = app.loadAtoms()
	if err != nil {
		return err
	}

	err = app.pixels.initialize(setup, screen)
	if err != nil {
		return err
	}

	err = app.fonts.initialize(dpi)
	if err != nil {
		return err
	}

	defer app.fonts.close()

	err = app.loadKeyboard()
	if err != nil {
		return err
	}

	err = app.createWindow()
	if err != nil {
		return err
	}

	// Closing the connection releases the window, GC, cursors and server resources.
	return app.runEvents()
}
