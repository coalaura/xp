package window

import (
	"strings"
	"unicode/utf8"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

const maxClipboardBytes = 16 << 20

type clipboardState struct {
	text        string
	pending     bool
	incremental bool
	index       int
	caret       int
	anchor      int
	original    string
	data        []byte
}

func (app *application) copyText(text string) bool {
	maximum := int(xproto.Setup(app.connection).MaximumRequestLength)*4 - 24
	if len(text) > maximum {
		return false
	}

	err := xproto.SetSelectionOwnerChecked(app.connection, app.handle, app.atoms["CLIPBOARD"], xproto.TimeCurrentTime).Check()
	if err != nil {
		return false
	}

	owner, err := xproto.GetSelectionOwner(app.connection, app.atoms["CLIPBOARD"]).Reply()
	if err != nil || owner.Owner != app.handle {
		return false
	}

	app.clipboard.text = text

	return true
}

func (app *application) pasteText(index int) {
	if app.clipboard.pending {
		return
	}

	input := &app.inputs[index]

	app.clipboard.pending = true
	app.clipboard.incremental = false
	app.clipboard.index = index
	app.clipboard.caret = input.caret
	app.clipboard.anchor = input.anchor
	app.clipboard.original = app.controls[index].Text
	app.clipboard.data = app.clipboard.data[:0]

	xproto.ConvertSelection(app.connection, app.handle, app.atoms["CLIPBOARD"], app.atoms["UTF8_STRING"], app.atoms["_XP_CLIPBOARD"], xproto.TimeCurrentTime)
}

func (app *application) selectionRequest(event xproto.SelectionRequestEvent) {
	property := event.Property
	if property == xproto.AtomNone {
		property = event.Target
	}

	accepted := false

	if event.Selection == app.atoms["CLIPBOARD"] {
		switch event.Target {
		case app.atoms["TARGETS"]:
			data := make([]byte, 12)

			xgb.Put32(data, uint32(app.atoms["TARGETS"]))
			xgb.Put32(data[4:], uint32(app.atoms["UTF8_STRING"]))
			xgb.Put32(data[8:], uint32(xproto.AtomString))

			err := xproto.ChangePropertyChecked(app.connection, xproto.PropModeReplace, event.Requestor, property, xproto.AtomAtom, 32, 3, data).Check()
			accepted = err == nil
		case app.atoms["UTF8_STRING"], xproto.AtomString:
			data := []byte(app.clipboard.text)

			if event.Target == xproto.AtomString {
				data = make([]byte, 0, len(app.clipboard.text))

				for _, character := range app.clipboard.text {
					if character > 255 {
						character = '?'
					}

					data = append(data, byte(character))
				}
			}

			maximum := int(xproto.Setup(app.connection).MaximumRequestLength)*4 - 24
			if len(data) <= maximum {
				err := xproto.ChangePropertyChecked(app.connection, xproto.PropModeReplace, event.Requestor, property, event.Target, 8, uint32(len(data)), data).Check()
				accepted = err == nil
			}
		}
	}

	if !accepted {
		property = xproto.AtomNone
	}

	notify := xproto.SelectionNotifyEvent{
		Time:      event.Time,
		Requestor: event.Requestor,
		Selection: event.Selection,
		Target:    event.Target,
		Property:  property,
	}

	xproto.SendEvent(app.connection, false, event.Requestor, 0, string(notify.Bytes()))
}

func (app *application) selectionNotify(event xproto.SelectionNotifyEvent) {
	if !app.clipboard.pending || event.Requestor != app.handle || event.Selection != app.atoms["CLIPBOARD"] {
		return
	}

	if event.Property == xproto.AtomNone {
		if event.Target == app.atoms["UTF8_STRING"] {
			xproto.ConvertSelection(app.connection, app.handle, app.atoms["CLIPBOARD"], xproto.AtomString, app.atoms["_XP_CLIPBOARD"], xproto.TimeCurrentTime)
		} else {
			app.clipboard.pending = false
		}

		return
	}

	property, err := xproto.GetProperty(app.connection, true, app.handle, event.Property, xproto.GetPropertyTypeAny, 0, maxClipboardBytes/4).Reply()
	if err != nil || property.BytesAfter != 0 {
		app.clipboard.pending = false

		return
	}

	if property.Type == app.atoms["INCR"] {
		app.clipboard.incremental = true

		return
	}

	app.finishPaste(property.Value, property.Type)
}

func (app *application) clipboardProperty(event xproto.PropertyNotifyEvent) {
	if !app.clipboard.pending || !app.clipboard.incremental || event.Window != app.handle || event.Atom != app.atoms["_XP_CLIPBOARD"] || event.State != xproto.PropertyNewValue {
		return
	}

	property, err := xproto.GetProperty(app.connection, true, app.handle, event.Atom, xproto.GetPropertyTypeAny, 0, maxClipboardBytes/4).Reply()
	if err != nil || property.BytesAfter != 0 || len(app.clipboard.data)+len(property.Value) > maxClipboardBytes {
		app.clipboard.pending = false
		app.clipboard.incremental = false

		return
	}

	if len(property.Value) == 0 {
		app.finishPaste(app.clipboard.data, property.Type)

		return
	}

	app.clipboard.data = append(app.clipboard.data, property.Value...)
}

func (app *application) finishPaste(data []byte, format xproto.Atom) {
	clipboard := &app.clipboard
	clipboard.pending = false
	clipboard.incremental = false

	index := clipboard.index
	input := &app.inputs[index]

	if app.controlFocus != index || app.controls[index].Text != clipboard.original || input.caret != clipboard.caret || input.anchor != clipboard.anchor {
		return
	}

	text := ""

	switch format {
	case app.atoms["UTF8_STRING"]:
		if !utf8.Valid(data) {
			return
		}

		text = string(data)
	case xproto.AtomString:
		var converted strings.Builder

		converted.Grow(len(data) * 2)

		for _, character := range data {
			converted.WriteRune(rune(character))
		}

		text = converted.String()
	default:
		return
	}

	app.replaceInput(index, cleanInput(text))
}
