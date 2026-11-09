//go:build windows || linux

package window

import (
	"image"

	"github.com/coalaura/xp/theme"
	"github.com/coalaura/xp/ui"
)

const (
	textLeft textMode = iota
	textCentered
	textWrapped
	textUnclipped
)

type textMode uint8

func (app *application) drawControls() error {
	err := app.prepareControlText()
	if err != nil {
		return err
	}

	surface := app.controlSurface().Clipped(app.layout.Content)

	for index := range app.controls {
		control := &app.controls[index]

		rectangle := app.controlRectangle(index)
		if rectangle.Empty() {
			continue
		}

		pressed := app.controlPressed == index && (app.keyboardPress || app.controlHot == index || control.Kind == ui.Slider)
		activeFocus := app.controlFocus == index && app.state.Active && !control.Disabled

		state := ui.State{
			Hot:         app.controlHot == index,
			Pressed:     pressed || app.comboOpen == index,
			Focused:     activeFocus && app.showFocus,
			ActiveFocus: activeFocus,
		}

		ui.Draw(surface, app.layout, rectangle, *control, state)

		if control.Kind == ui.Input {
			app.drawInput(surface, index, state)
		} else {
			app.drawControlCaption(surface, index, rectangle, state)
		}
	}

	app.drawComboList(surface)

	return nil
}

func (app *application) drawControlCaption(surface theme.Surface, index int, rectangle image.Rectangle, state ui.State) {
	control := &app.controls[index]
	text := control.Text
	bold := control.Bold || control.Kind == ui.Heading

	var (
		foreground uint32
		mode       textMode
	)

	if control.Kind == ui.Link {
		foreground = 0x0000FF
	}

	if control.Kind == ui.Group {
		foreground = 0x0046D5
	}

	if control.Disabled {
		foreground = 0xACA899
	}

	switch control.Kind {
	case ui.Label, ui.Heading:
		mode = textWrapped
	case ui.Link:
	case ui.Button:
		mode = textCentered
		rectangle = rectangle.Inset(app.layout.Scale(3))

		if state.Pressed && !control.Disabled {
			rectangle = rectangle.Add(image.Pt(app.layout.Scale(1), app.layout.Scale(1)))
		}
	case ui.Checkbox, ui.Radio:
		rectangle.Min.X = ui.CheckboxBounds(app.layout, rectangle).Max.X + app.layout.Scale(4)
	case ui.ComboBox:
		rectangle = ui.FieldTextBounds(app.layout, rectangle, ui.ComboBox)
		text = ""

		if control.Selected >= 0 {
			text = control.Items[control.Selected]
		}

		if state.ActiveFocus && !control.Disabled && app.comboOpen != index {
			surface.Fill(rectangle, 0x316AC5)
			foreground = 0xFFFFFF

			if state.Focused {
				ui.DrawFocus(surface, app.layout, rectangle)
			}
		}
	case ui.Group:
		rectangle.Min.X += app.layout.Scale(9)
		rectangle.Max.X -= app.layout.Scale(9)
		rectangle.Max.Y = rectangle.Min.Y + app.layout.Scale(13)
		rectangle.Max.X = min(rectangle.Max.X, rectangle.Min.X+app.textWidth(text, bold))

		if text != "" && !rectangle.Empty() {
			surface.Fill(rectangle.Inset(-app.layout.Scale(2)), theme.Background)
		}
	default:
		return
	}

	if rectangle.Empty() {
		return
	}

	clip := rectangle.Intersect(app.layout.Content)

	app.drawControlString(surface, text, rectangle, clip, foreground, bold, mode)

	if text == "" {
		return
	}

	if control.Kind == ui.Checkbox || control.Kind == ui.Radio || control.Kind == ui.Link {
		width := min(rectangle.Dx(), app.textWidth(text, bold))
		height := app.textHeight(bold)

		top := rectangle.Min.Y + (rectangle.Dy()-height)/2
		textBounds := image.Rect(rectangle.Min.X, top, rectangle.Min.X+width, top+height)

		if control.Kind == ui.Link {
			underline := image.Rect(textBounds.Min.X, textBounds.Max.Y-app.layout.Scale(2), textBounds.Max.X, textBounds.Max.Y-app.layout.Scale(1))

			surface.Clipped(clip).Fill(underline, foreground)
		}

		if state.Focused {
			ui.DrawFocus(surface.Clipped(app.controlRectangle(index)), app.layout, textBounds.Inset(-app.layout.Scale(1)))
		}
	}
}

func (app *application) drawInput(surface theme.Surface, index int, state ui.State) {
	control := &app.controls[index]
	input := &app.inputs[index]

	viewport := app.revealCaret(index)
	if viewport.Empty() {
		return
	}

	clip := viewport.Intersect(app.layout.Content)

	bounds := viewport
	bounds.Min.X -= input.scroll
	bounds.Max.X = max(viewport.Max.X, bounds.Min.X+app.textWidth(input.display, control.Bold)+app.layout.Scale(1))

	foreground := uint32(0x000000)

	if control.Disabled {
		foreground = 0xACA899
	}

	app.drawControlString(surface, input.display, bounds, clip, foreground, control.Bold, textUnclipped)

	if !state.ActiveFocus {
		return
	}

	start := app.inputDisplayOffset(index, min(input.caret, input.anchor))
	end := app.inputDisplayOffset(index, max(input.caret, input.anchor))
	left := bounds.Min.X + app.textWidth(input.display[:start], control.Bold)
	right := bounds.Min.X + app.textWidth(input.display[:end], control.Bold)
	height := min(viewport.Dy(), app.textHeight(control.Bold))
	top := viewport.Min.Y + (viewport.Dy()-height)/2
	selection := image.Rect(left, top, right, top+height).Intersect(clip)

	if start != end {
		surface.Fill(selection, 0x316AC5)
		app.drawControlString(surface, input.display, bounds, selection, 0xFFFFFF, control.Bold, textUnclipped)
	}

	offset := app.inputDisplayOffset(index, input.caret)
	left = bounds.Min.X + app.textWidth(input.display[:offset], control.Bold)
	caret := image.Rect(left, top, left+app.layout.Scale(1), top+height)

	surface.Clipped(clip).Fill(caret, 0x000000)
}

func (app *application) drawComboList(surface theme.Surface) {
	if app.comboOpen < 0 {
		return
	}

	bounds, rows := app.comboBounds()
	if rows == 0 {
		return
	}

	control := &app.controls[app.comboOpen]

	one := app.layout.Scale(1)
	surface.Fill(bounds, 0x000000)
	inner := bounds.Inset(one)
	surface.Fill(inner, 0xFFFFFF)
	textRight := inner.Max.X

	if len(control.Items) > rows {
		textRight -= app.layout.Scale(12)
		track := image.Rect(textRight, inner.Min.Y, inner.Max.X, inner.Max.Y)
		surface.Fill(track, 0xF1EFE2)
		height := max(app.layout.Scale(8), track.Dy()*rows/len(control.Items))
		travel := max(0, track.Dy()-height)
		top := track.Min.Y + travel*app.comboFirst/(len(control.Items)-rows)
		thumb := image.Rect(track.Min.X+one, top, track.Max.X-one, top+height)
		surface.Fill(thumb, 0xACA899)
	}

	for row := range rows {
		index := app.comboFirst + row
		if index >= len(control.Items) {
			break
		}

		top := inner.Min.Y + row*app.layout.Scale(18)
		rectangle := image.Rect(inner.Min.X, top, textRight, top+app.layout.Scale(18))
		foreground := uint32(0x000000)

		if index == app.comboHover {
			surface.Fill(rectangle, 0x316AC5)
			foreground = 0xFFFFFF
		}

		rectangle.Min.X += app.layout.Scale(3)
		rectangle.Max.X -= app.layout.Scale(3)

		app.drawControlString(surface, control.Items[index], rectangle, rectangle, foreground, control.Bold, textLeft)
	}
}
