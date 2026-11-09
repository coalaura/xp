//go:build windows || linux

package window

import (
	"errors"
	"image"
	"slices"

	"github.com/coalaura/xp/ui"
)

type controlState struct {
	controls       []ui.Control
	inputs         []inputState
	controlHot     int
	controlPressed int
	controlFocus   int
	keyboardPress  bool
	showFocus      bool
	sliderOffset   int
	comboOpen      int
	comboHover     int
	comboFirst     int
	comboDragging  bool
}

func (app *application) Text(id string) (string, error) {
	index, err := app.controlIndex(id)
	if err != nil {
		return "", err
	}

	control := &app.controls[index]
	if control.Kind == ui.ComboBox {
		if control.Selected < 0 {
			return "", nil
		}

		return control.Items[control.Selected], nil
	}

	return control.Text, nil
}

func (app *application) SetText(id, text string) error {
	index, err := app.controlIndex(id)
	if err != nil {
		return err
	}

	control := &app.controls[index]
	if !validText(text) || control.Kind == ui.Input && !validSingleLine(text) {
		return errors.New("xp: invalid text for control")
	}

	if control.Kind == ui.ComboBox {
		return errors.New("xp: use SetSelected to select a combo item")
	}

	control.Text = text

	if control.Kind == ui.Input {
		app.inputs[index] = inputState{caret: len(text), anchor: len(text)}

		app.refreshInput(index)
	}

	app.invalidate()

	return nil
}

func (app *application) Checked(id string) (bool, error) {
	state, err := app.CheckState(id)

	return state == ui.Checked, err
}

func (app *application) SetChecked(id string, checked bool) error {
	state := ui.Unchecked

	if checked {
		state = ui.Checked
	}

	return app.SetCheckState(id, state)
}

func (app *application) CheckState(id string) (ui.CheckState, error) {
	index, err := app.choiceIndex(id)
	if err != nil {
		return ui.Unchecked, err
	}

	control := &app.controls[index]
	if control.Mixed {
		return ui.Mixed, nil
	}

	if control.Checked {
		return ui.Checked, nil
	}

	return ui.Unchecked, nil
}

func (app *application) SetCheckState(id string, state ui.CheckState) error {
	index, err := app.choiceIndex(id)
	if err != nil {
		return err
	}

	control := &app.controls[index]
	if state > ui.Mixed || control.Kind == ui.Radio && state == ui.Mixed {
		return errors.New("xp: invalid check state")
	}

	control.Checked = state == ui.Checked
	control.Mixed = state == ui.Mixed

	if control.Kind == ui.Radio && control.Checked {
		app.selectRadio(index)
	}

	app.invalidate()

	return nil
}

func (app *application) Value(id string) (int, error) {
	index, err := app.kindIndex(id, ui.Slider)
	if err != nil {
		return 0, err
	}

	return app.controls[index].Value, nil
}

func (app *application) SetValue(id string, value int) error {
	index, err := app.kindIndex(id, ui.Slider)
	if err != nil {
		return err
	}

	control := &app.controls[index]
	if value < control.Min || value > control.Max {
		return errors.New("xp: slider value outside range")
	}

	control.Value = value

	app.invalidate()

	return nil
}

func (app *application) Selected(id string) (int, error) {
	index, err := app.kindIndex(id, ui.ComboBox)
	if err != nil {
		return -1, err
	}

	return app.controls[index].Selected, nil
}

func (app *application) SetSelected(id string, selected int) error {
	index, err := app.kindIndex(id, ui.ComboBox)
	if err != nil {
		return err
	}

	control := &app.controls[index]
	if selected < -1 || selected >= len(control.Items) {
		return errors.New("xp: combo selection outside item list")
	}

	control.Selected = selected

	if app.comboOpen == index {
		app.closeCombo()
	}

	app.invalidate()

	return nil
}

func (app *application) SetDisabled(id string, disabled bool) error {
	index, err := app.controlIndex(id)
	if err != nil {
		return err
	}

	app.controls[index].Disabled = disabled

	if disabled {
		if app.controlFocus == index {
			app.controlFocus = -1
		}

		if app.controlHot == index {
			app.controlHot = -1
		}

		if app.controlPressed == index {
			app.cancelPress()
		}

		if app.comboOpen == index {
			app.closeCombo()
		}
	}

	app.invalidate()

	return nil
}

func (app *application) initializeControls() {
	app.controlState = controlState{
		controls:       slices.Clone(app.options.Controls),
		inputs:         make([]inputState, len(app.options.Controls)),
		controlHot:     -1,
		controlPressed: -1,
		controlFocus:   -1,
		comboOpen:      -1,
		comboHover:     -1,
	}

	for index := range app.controls {
		control := &app.controls[index]
		control.Items = slices.Clone(control.Items)

		if control.Kind == ui.ComboBox && len(control.Items) == 0 {
			control.Selected = -1
		}

		if control.Kind == ui.Input {
			app.inputs[index] = inputState{caret: len(control.Text), anchor: len(control.Text)}

			app.refreshInput(index)
		}
	}
}

func (app *application) controlIndex(id string) (int, error) {
	for index := range app.controls {
		if id != "" && app.controls[index].ID == id {
			return index, nil
		}
	}

	return -1, errors.New("xp: no control with that ID")
}

func (app *application) kindIndex(id string, kind ui.Kind) (int, error) {
	index, err := app.controlIndex(id)
	if err != nil {
		return -1, err
	}

	if app.controls[index].Kind != kind {
		return -1, errors.New("xp: control has a different kind")
	}

	return index, nil
}

func (app *application) choiceIndex(id string) (int, error) {
	index, err := app.controlIndex(id)
	if err != nil {
		return -1, err
	}

	kind := app.controls[index].Kind
	if kind != ui.Checkbox && kind != ui.Radio {
		return -1, errors.New("xp: control is not a checkbox or radio")
	}

	return index, nil
}

func (app *application) controlAt(point image.Point) int {
	if !point.In(app.layout.Content) {
		return -1
	}

	for index := len(app.controls) - 1; index >= 0; index-- {
		control := &app.controls[index]
		if control.Interactive() && point.In(app.controlRectangle(index)) {
			return index
		}
	}

	return -1
}

func (app *application) controlRectangle(index int) image.Rectangle {
	return app.controls[index].Rectangle(app.layout, app.options.Width, app.options.Height)
}

func (app *application) activateControl(index int) {
	if index < 0 || index >= len(app.controls) || !app.controls[index].Interactive() {
		return
	}

	control := &app.controls[index]
	changed := false

	switch control.Kind {
	case ui.Checkbox:
		switch {
		case control.Mixed:
			control.Mixed = false
			control.Checked = !control.ThreeState
		case control.Checked && control.ThreeState:
			control.Checked = false
			control.Mixed = true
		default:
			control.Checked = !control.Checked
		}

		changed = true
	case ui.Radio:
		changed = !control.Checked

		app.selectRadio(index)
	case ui.ComboBox:
		app.toggleCombo(index)

		return
	case ui.Button, ui.Link:
	default:
		return
	}

	app.invalidate()

	if changed {
		app.changed(index)
	}

	if app.options.OnClick != nil {
		app.options.OnClick(app, control.ID)
	}
}

func (app *application) selectRadio(index int) {
	group := app.controls[index].RadioGroup

	for peer := range app.controls {
		control := &app.controls[peer]
		if control.Kind == ui.Radio && control.RadioGroup == group {
			control.Checked = peer == index
		}
	}
}

func (app *application) changed(index int) {
	app.invalidate()

	if app.options.OnChange != nil {
		app.options.OnChange(app, app.controls[index].ID)
	}
}

func (app *application) clearControlFocus() {
	app.controlFocus = -1
	app.showFocus = false

	app.closeCombo()
}

func (app *application) moveFocus(backward bool) {
	app.cancelPress()
	app.closeCombo()

	count := len(app.controls)
	index := app.controlFocus
	direction := 1

	if backward {
		direction = -1

		if index < 0 {
			index = 0
		}
	}

	for range count {
		index = (index + direction + count) % count

		if app.tabStop(index) {
			app.controlFocus = index
			app.showFocus = true

			if app.controls[index].Kind == ui.Input {
				app.inputs[index].anchor = 0
				app.inputs[index].caret = len(app.controls[index].Text)
			}

			app.invalidate()

			return
		}
	}
}

func (app *application) tabStop(index int) bool {
	control := &app.controls[index]
	if !control.Interactive() {
		return false
	}

	if control.Kind != ui.Radio || control.Checked {
		return true
	}

	for peer := range app.controls {
		candidate := &app.controls[peer]
		if peer != index && candidate.Kind == ui.Radio && candidate.RadioGroup == control.RadioGroup && candidate.Interactive() && (candidate.Checked || peer < index) {
			return false
		}
	}

	return true
}

func (app *application) activateDefault(cancel bool) {
	if !cancel && app.controlFocus >= 0 && app.controls[app.controlFocus].Kind == ui.Button {
		app.activateControl(app.controlFocus)

		return
	}

	for index := range app.controls {
		control := &app.controls[index]
		if control.Kind == ui.Button && control.Interactive() && (cancel && control.Cancel || !cancel && control.Default) {
			app.activateControl(index)

			return
		}
	}
}
