//go:build windows || linux

package window

import (
	"image"

	"github.com/coalaura/xp/ui"
)

const (
	controlTab controlKey = -1 - iota
	controlEnter
	controlEscape
	controlLeft
	controlRight
	controlUp
	controlDown
	controlHome
	controlEnd
	controlPageUp
	controlPageDown
	controlBackspace
	controlDelete
	controlF4
)

type controlKey int32

type keyModifiers struct {
	shift   bool
	control bool
	alt     bool
}

func (app *application) controlDown(point image.Point, shift bool) {
	if app.comboOpen >= 0 {
		bounds, rows := app.comboBounds()
		if point.In(bounds) {
			app.controlPressed = app.comboOpen

			if app.comboScrollbar(point, bounds, rows) {
				app.comboDragging = true

				app.scrollComboPointer(point, bounds, rows)
			} else {
				app.comboHover = app.comboItemAt(point)
			}

			app.invalidate()

			return
		}

		if !point.In(app.controlRectangle(app.comboOpen)) {
			app.closeCombo()
		}
	}

	index := app.controlAt(point)
	if index < 0 {
		app.clearControlFocus()

		return
	}

	app.controlFocus = index
	app.controlPressed = index
	app.controlHot = index
	app.showFocus = false

	control := &app.controls[index]

	switch control.Kind {
	case ui.Input:
		position := app.inputPosition(index, point.X)
		app.inputs[index].caret = position

		if !shift {
			app.inputs[index].anchor = position
		}
	case ui.Slider:
		thumb := ui.SliderThumb(app.layout, app.controlRectangle(index), *control)
		app.sliderOffset = thumb.Dx() / 2

		if control.Vertical {
			app.sliderOffset = thumb.Dy() / 2
		}

		if point.In(thumb) {
			app.sliderOffset = point.X - thumb.Min.X

			if control.Vertical {
				app.sliderOffset = point.Y - thumb.Min.Y
			}
		} else {
			app.dragSlider(index, point)
		}
	}

	app.invalidate()
}

func (app *application) controlMove(point image.Point) {
	if app.comboOpen >= 0 {
		bounds, rows := app.comboBounds()

		if app.comboDragging {
			app.scrollComboPointer(point, bounds, rows)

			return
		}

		item := app.comboItemAt(point)
		if item >= 0 && item != app.comboHover {
			app.comboHover = item

			app.invalidate()
		}
	}

	index := app.controlPressed
	if index < 0 || app.keyboardPress || !app.controls[index].Interactive() {
		return
	}

	switch app.controls[index].Kind {
	case ui.Slider:
		app.dragSlider(index, point)
	case ui.Input:
		app.inputs[index].caret = app.inputPosition(index, point.X)

		app.invalidate()
	}
}

func (app *application) controlUp(point image.Point) {
	if app.keyboardPress {
		return
	}

	pressed := app.controlPressed
	if pressed < 0 {
		app.cancelPress()

		return
	}

	scrolling := app.comboDragging
	app.comboDragging = false

	app.cancelPress()

	if scrolling {
		return
	}

	if app.comboOpen == pressed {
		item := app.comboItemAt(point)
		if item >= 0 {
			app.chooseCombo(pressed, item)
			app.closeCombo()

			return
		}
	}

	if app.controlAt(point) == pressed {
		app.activateControl(pressed)
	}
}

func (app *application) dragSlider(index int, point image.Point) {
	control := &app.controls[index]
	position := point.X - app.sliderOffset

	if control.Vertical {
		position = point.Y - app.sliderOffset
	}

	value := ui.SliderValue(app.layout, app.controlRectangle(index), *control, position)
	app.changeSlider(index, value)
}

func (app *application) changeSlider(index, value int) {
	control := &app.controls[index]
	value = max(control.Min, min(control.Max, value))

	if value != control.Value {
		control.Value = value

		app.changed(index)
	}
}

func (app *application) controlKeyDown(key controlKey, modifiers keyModifiers, repeat bool) {
	if repeat && (key == controlF4 || key == controlEnter || key == ' ' || modifiers.alt) {
		return
	}

	if key == controlTab {
		app.moveFocus(modifiers.shift)

		return
	}

	index := app.controlFocus

	if app.comboOpen >= 0 {
		app.comboKey(app.comboOpen, key, modifiers)

		return
	}

	if index >= 0 && app.controls[index].Interactive() {
		control := &app.controls[index]
		app.showFocus = true

		switch control.Kind {
		case ui.Input:
			if !modifiers.alt && app.editKey(index, key, modifiers) {
				return
			}
		case ui.ComboBox:
			if app.comboKey(index, key, modifiers) {
				return
			}
		case ui.Slider:
			if !modifiers.alt && !modifiers.control && app.sliderKey(index, key) {
				return
			}
		case ui.Radio:
			direction := key == controlLeft || key == controlUp || key == controlRight || key == controlDown
			if direction && !modifiers.alt && !modifiers.control {
				app.radioKey(index, key == controlLeft || key == controlUp)

				return
			}
		case ui.Link:
			if key == controlEnter && !modifiers.alt && !modifiers.control {
				app.activateControl(index)

				return
			}
		}
	}

	if modifiers.alt || modifiers.control {
		return
	}

	switch key {
	case ' ':
		if !repeat && index >= 0 && app.controls[index].Interactive() {
			kind := app.controls[index].Kind
			if kind == ui.Button || kind == ui.Checkbox || kind == ui.Radio || kind == ui.Link {
				app.keyboardPress = true
				app.controlPressed = index

				app.invalidate()
			}
		}
	case controlEnter:
		if !repeat {
			app.activateDefault(false)
		}
	case controlEscape:
		app.cancelPress()

		if !repeat {
			app.activateDefault(true)
		}
	}
}

func (app *application) controlKeyUp(key controlKey) {
	if key == ' ' && app.keyboardPress {
		index := app.controlPressed

		app.cancelPress()
		app.activateControl(index)
	}
}

func (app *application) sliderKey(index int, key controlKey) bool {
	control := &app.controls[index]
	value := int64(control.Value)
	step := int64(max(1, control.Step))
	page := max(step, (int64(control.Max)-int64(control.Min))/10)

	switch key {
	case controlLeft, controlUp:
		value -= step
	case controlRight, controlDown:
		value += step
	case controlPageUp:
		value -= page
	case controlPageDown:
		value += page
	case controlHome:
		value = int64(control.Min)
	case controlEnd:
		value = int64(control.Max)
	default:
		return false
	}

	value = max(int64(control.Min), min(int64(control.Max), value))
	app.changeSlider(index, int(value))

	return true
}

func (app *application) radioKey(index int, backward bool) {
	group := app.controls[index].RadioGroup
	count := len(app.controls)
	direction := 1

	if backward {
		direction = -1
	}

	for range count {
		index = (index + direction + count) % count
		control := &app.controls[index]

		if control.Kind == ui.Radio && control.RadioGroup == group && control.Interactive() {
			app.controlFocus = index

			app.activateControl(index)

			return
		}
	}
}

func (app *application) controlWheel(point image.Point, steps int) {
	if app.comboOpen >= 0 {
		_, rows := app.comboBounds()
		maximum := max(0, len(app.controls[app.comboOpen].Items)-rows)
		app.comboFirst = max(0, min(maximum, app.comboFirst-steps*3))

		app.invalidate()

		return
	}

	index := app.controlAt(point)
	if index < 0 || index != app.controlFocus {
		return
	}

	control := &app.controls[index]

	switch control.Kind {
	case ui.Slider:
		value := int64(control.Value) + int64(steps)*int64(max(1, control.Step))
		value = max(int64(control.Min), min(int64(control.Max), value))

		app.changeSlider(index, int(value))
	case ui.ComboBox:
		app.chooseCombo(index, max(0, min(len(control.Items)-1, control.Selected-steps)))
	}
}
