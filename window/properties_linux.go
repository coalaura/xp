package window

import (
	"os"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/shape"
	"github.com/jezek/xgb/xproto"
)

var linuxAtomNames = [...]string{
	"UTF8_STRING",
	"WM_PROTOCOLS",
	"WM_DELETE_WINDOW",
	"WM_TAKE_FOCUS",
	"WM_CHANGE_STATE",
	"WM_CLIENT_MACHINE",
	"_MOTIF_WM_HINTS",
	"_NET_SUPPORTED",
	"_NET_WM_NAME",
	"_NET_WM_PID",
	"_NET_WM_PING",
	"_NET_WM_WINDOW_TYPE",
	"_NET_WM_WINDOW_TYPE_NORMAL",
	"_NET_WM_WINDOW_TYPE_DIALOG",
	"_NET_WM_STATE",
	"_NET_WM_STATE_MAXIMIZED_HORZ",
	"_NET_WM_STATE_MAXIMIZED_VERT",
	"_NET_WM_MOVERESIZE",
	"_NET_ACTIVE_WINDOW",
}

func (app *application) loadAtoms() error {
	app.atoms = make(map[string]xproto.Atom, len(linuxAtomNames))
	cookies := make([]xproto.InternAtomCookie, len(linuxAtomNames))

	for index, name := range linuxAtomNames {
		cookies[index] = xproto.InternAtom(app.connection, false, uint16(len(name)), name)
	}

	for index, name := range linuxAtomNames {
		reply, err := cookies[index].Reply()
		if err != nil {
			return err
		}

		app.atoms[name] = reply.Atom
	}

	property, err := xproto.GetProperty(app.connection, false, app.screen.Root, app.atoms["_NET_SUPPORTED"], xproto.AtomAtom, 0, 4096).Reply()
	if err != nil {
		return err
	}

	app.supported = make(map[xproto.Atom]bool, len(property.Value)/4)

	if property.Format == 32 {
		for offset := 0; offset+4 <= len(property.Value); offset += 4 {
			app.supported[xproto.Atom(xgb.Get32(property.Value[offset:]))] = true
		}
	}

	return nil
}

func (app *application) setWindowProperties() {
	app.propertyBytes(xproto.AtomWmName, xproto.AtomString, []byte(app.options.Title))
	app.propertyBytes(app.atoms["_NET_WM_NAME"], app.atoms["UTF8_STRING"], []byte(app.options.Title))
	app.propertyBytes(xproto.AtomWmClass, xproto.AtomString, []byte("xp\x00XP\x00"))
	app.property32(app.atoms["_NET_WM_PID"], xproto.AtomCardinal, uint32(os.Getpid()))
	app.property32(app.atoms["WM_PROTOCOLS"], xproto.AtomAtom, uint32(app.atoms["WM_DELETE_WINDOW"]), uint32(app.atoms["WM_TAKE_FOCUS"]), uint32(app.atoms["_NET_WM_PING"]))
	app.property32(xproto.AtomWmHints, xproto.AtomWmHints, 3, 1, 1, 0, 0, 0, 0, 0, 0)

	hostname, err := os.Hostname()
	if err == nil {
		app.propertyBytes(app.atoms["WM_CLIENT_MACHINE"], xproto.AtomString, []byte(hostname))
	}

	windowType := app.atoms["_NET_WM_WINDOW_TYPE_NORMAL"]
	functions := uint32(1)

	minimumWidth := app.layout.Scale(app.options.MinWidth)
	minimumHeight := app.layout.Scale(app.options.MinHeight)

	maximumWidth := maxXDimension
	maximumHeight := maxXDimension

	if app.options.Style == Dialog {
		windowType = app.atoms["_NET_WM_WINDOW_TYPE_DIALOG"]
		functions = 4 | 32 // Motif move and close functions.

		minimumWidth = app.layout.Width
		minimumHeight = app.layout.Height

		maximumWidth = minimumWidth
		maximumHeight = minimumHeight
	}

	app.property32(app.atoms["_NET_WM_WINDOW_TYPE"], xproto.AtomAtom, uint32(windowType))
	app.property32(app.atoms["_MOTIF_WM_HINTS"], app.atoms["_MOTIF_WM_HINTS"], 3, functions, 0, 0, 0)

	// ICCCM PMinSize | PMaxSize; the first five slots precede the size limits.
	app.property32(xproto.AtomWmNormalHints, xproto.AtomWmSizeHints, 16|32, 0, 0, 0, 0, uint32(minimumWidth), uint32(minimumHeight), uint32(maximumWidth), uint32(maximumHeight), 0, 0, 0, 0, 0, 0, 0, 0, 0)
}

func (app *application) propertyBytes(property, kind xproto.Atom, data []byte) {
	xproto.ChangeProperty(app.connection, xproto.PropModeReplace, app.handle, property, kind, 8, uint32(len(data)), data)
}

func (app *application) property32(property, kind xproto.Atom, values ...uint32) {
	data := make([]byte, 4*len(values))

	for index, value := range values {
		xgb.Put32(data[index*4:], value)
	}

	xproto.ChangeProperty(app.connection, xproto.PropModeReplace, app.handle, property, kind, 32, uint32(len(values)), data)
}

func (app *application) sendRootMessage(kind string, values ...uint32) {
	var data [5]uint32

	copy(data[:], values)

	event := xproto.ClientMessageEvent{
		Format: 32,
		Window: app.handle,
		Type:   app.atoms[kind],
		Data:   xproto.ClientMessageDataUnionData32New(data[:]),
	}

	app.sendRootEvent(event)
}

func (app *application) sendRootEvent(event xproto.ClientMessageEvent) {
	mask := uint32(xproto.EventMaskSubstructureRedirect | xproto.EventMaskSubstructureNotify)
	xproto.SendEvent(app.connection, false, app.screen.Root, mask, string(event.Bytes()))
}

func (app *application) clientMessage(event xproto.ClientMessageEvent) {
	if event.Format != 32 || event.Type != app.atoms["WM_PROTOCOLS"] || event.Window != app.handle {
		return
	}

	protocol := xproto.Atom(event.Data.Data32[0])

	switch protocol {
	case app.atoms["WM_DELETE_WINDOW"]:
		app.Close()
	case app.atoms["WM_TAKE_FOCUS"]:
		xproto.SetInputFocus(app.connection, xproto.InputFocusParent, app.handle, xproto.Timestamp(event.Data.Data32[1]))
	case app.atoms["_NET_WM_PING"]:
		event.Window = app.screen.Root
		app.sendRootEvent(event)
	}
}

func (app *application) updateWindowState() {
	property, err := xproto.GetProperty(app.connection, false, app.handle, app.atoms["_NET_WM_STATE"], xproto.AtomAtom, 0, 1024).Reply()
	if err != nil || property == nil {
		return
	}

	var (
		horizontal bool
		vertical   bool
	)

	for offset := 0; offset+4 <= len(property.Value); offset += 4 {
		atom := xproto.Atom(xgb.Get32(property.Value[offset:]))
		horizontal = horizontal || atom == app.atoms["_NET_WM_STATE_MAXIMIZED_HORZ"]
		vertical = vertical || atom == app.atoms["_NET_WM_STATE_MAXIMIZED_VERT"]
	}

	maximized := horizontal && vertical && app.options.Style != Dialog
	if maximized != app.state.Maximized {
		app.state.Maximized = maximized

		app.resize(app.layout.Width, app.layout.Height)
	}
}

func (app *application) toggleMaximize() {
	if app.options.Style == Dialog {
		return
	}

	app.sendRootMessage("_NET_WM_STATE", 2, uint32(app.atoms["_NET_WM_STATE_MAXIMIZED_HORZ"]), uint32(app.atoms["_NET_WM_STATE_MAXIMIZED_VERT"]), 1)
}

func (app *application) initializeShape() {
	reply, err := xproto.QueryExtension(app.connection, 5, "SHAPE").Reply()
	if err != nil || reply == nil || !reply.Present {
		return
	}

	// shape.Init writes XGB's global event registry while other windows may read it.
	// We never select Shape events, so only the connection-local opcode is needed.
	app.connection.ExtLock.Lock()
	app.connection.Extensions["SHAPE"] = reply.MajorOpcode
	app.connection.ExtLock.Unlock()

	app.shape = true

	app.updateShape()
}

func (app *application) updateShape() {
	if !app.shape {
		return
	}

	var rectangles [6]xproto.Rectangle

	count := 1
	rectangles[0] = xproto.Rectangle{Width: uint16(app.layout.Width), Height: uint16(app.layout.Height)}

	if !app.state.Maximized {
		insets := [...]int{5, 3, 2, 1, 1}

		for index, inset := range insets {
			left := app.layout.Scale(inset)
			top := app.layout.Scale(index)
			height := app.layout.Scale(index+1) - top

			rectangles[index] = xproto.Rectangle{
				X:      int16(left),
				Y:      int16(top),
				Width:  uint16(max(0, app.layout.Width-2*left)),
				Height: uint16(height),
			}
		}

		top := app.layout.Scale(5)
		rectangles[5] = xproto.Rectangle{Y: int16(top), Width: uint16(app.layout.Width), Height: uint16(max(0, app.layout.Height-top))}
		count = len(rectangles)
	}

	shape.Rectangles(app.connection, shape.SoSet, shape.SkBounding, xproto.ClipOrderingYXBanded, app.handle, 0, 0, rectangles[:count])
}
