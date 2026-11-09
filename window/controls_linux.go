package window

import (
	"errors"
	"image"
	"strings"

	"github.com/coalaura/xp/ui"
)

func (app *application) Close() {
	app.closed = true
}

func (app *application) SetText(id, text string) error {
	if strings.ContainsRune(text, 0) {
		return errors.New("xp: control text contains a NUL character")
	}

	for index := range app.controls {
		control := &app.controls[index]
		if id != "" && control.ID == id {
			control.Text = text
			app.dirty = true

			return nil
		}
	}

	return errors.New("xp: no control with that ID")
}

func (app *application) Checked(id string) (bool, error) {
	control, err := app.checkbox(id)
	if err != nil {
		return false, err
	}

	return control.Checked, nil
}

func (app *application) SetChecked(id string, checked bool) error {
	control, err := app.checkbox(id)
	if err != nil {
		return err
	}

	control.Checked = checked
	app.dirty = true

	return nil
}

func (app *application) checkbox(id string) (*ui.Control, error) {
	for index := range app.controls {
		control := &app.controls[index]
		if id != "" && control.ID == id && control.Kind == ui.Checkbox {
			return control, nil
		}
	}

	return nil, errors.New("xp: no checkbox with that ID")
}

func (app *application) controlAt(point image.Point) int {
	if !point.In(app.layout.Content) {
		return -1
	}

	for index := len(app.controls) - 1; index >= 0; index-- {
		control := &app.controls[index]
		if !control.Interactive() {
			continue
		}

		rectangle := control.Rectangle(app.layout, app.options.Width, app.options.Height)
		if point.In(rectangle) {
			return index
		}
	}

	return -1
}

func (app *application) activateControl(index int) {
	if index < 0 || index >= len(app.controls) {
		return
	}

	control := &app.controls[index]
	if !control.Interactive() {
		return
	}

	if control.Kind == ui.Checkbox {
		control.Checked = !control.Checked
	}

	app.dirty = true

	if app.options.OnClick != nil {
		app.options.OnClick(app, control.ID)
	}
}

func (app *application) focusNext(backward bool) {
	app.cancelPress()
	count := len(app.controls)

	index := app.controlFocus
	if index < 0 {
		if backward {
			index = 0
		} else {
			index = count - 1
		}
	}

	for range count {
		if backward {
			index = (index + count - 1) % count
		} else {
			index = (index + 1) % count
		}

		if app.controls[index].Interactive() {
			app.controlFocus = index
			app.showFocus = true
			app.dirty = true

			return
		}
	}
}

func (app *application) activateDefault(cancel bool) {
	if !cancel && app.controlFocus >= 0 && app.controls[app.controlFocus].Kind == ui.Button {
		app.activateControl(app.controlFocus)

		return
	}

	for index := range app.controls {
		control := &app.controls[index]
		if control.Kind != ui.Button || !control.Interactive() {
			continue
		}

		if cancel && control.Cancel || !cancel && control.Default {
			app.activateControl(index)

			return
		}
	}
}
