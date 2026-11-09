//go:build windows || linux

package window

import (
	"image"
	"strings"
	"unicode/utf8"
)

func (app *application) toggleCombo(index int) {
	if app.comboOpen == index {
		app.closeCombo()

		return
	}

	if len(app.controls[index].Items) == 0 {
		return
	}

	app.comboOpen = index
	app.comboHover = max(0, app.controls[index].Selected)
	app.comboFirst = 0

	_, rows := app.comboBounds()
	if rows == 0 {
		app.closeCombo()

		return
	}

	app.revealComboItem()
	app.invalidate()
}

func (app *application) closeCombo() {
	app.comboOpen = -1
	app.comboDragging = false

	app.invalidate()
}

func (app *application) comboBounds() (image.Rectangle, int) {
	if app.comboOpen < 0 {
		return image.Rectangle{}, 0
	}

	control := &app.controls[app.comboOpen]
	rectangle := app.controlRectangle(app.comboOpen)
	content := app.layout.Content
	rowHeight := app.layout.Scale(18)
	border := app.layout.Scale(1)
	wanted := min(8, len(control.Items))
	below := max(0, content.Max.Y-rectangle.Max.Y)
	above := max(0, rectangle.Min.Y-content.Min.Y)
	upwards := below < wanted*rowHeight+2*border && above > below
	available := below

	if upwards {
		available = above
	}

	rows := min(wanted, max(0, (available-2*border)/rowHeight))
	if rows == 0 {
		return image.Rectangle{}, 0
	}

	height := rows*rowHeight + 2*border
	left := max(content.Min.X, min(rectangle.Min.X, content.Max.X-rectangle.Dx()))
	top := rectangle.Max.Y

	if upwards {
		top = rectangle.Min.Y - height
	}

	return image.Rect(left, top, min(content.Max.X, left+rectangle.Dx()), top+height), rows
}

func (app *application) comboItemAt(point image.Point) int {
	bounds, rows := app.comboBounds()
	if rows == 0 || !point.In(bounds.Inset(app.layout.Scale(1))) || app.comboScrollbar(point, bounds, rows) {
		return -1
	}

	return app.comboFirst + (point.Y-bounds.Min.Y-app.layout.Scale(1))/app.layout.Scale(18)
}

func (app *application) comboScrollbar(point image.Point, bounds image.Rectangle, rows int) bool {
	return app.comboOpen >= 0 && len(app.controls[app.comboOpen].Items) > rows && point.X >= bounds.Max.X-app.layout.Scale(13)
}

func (app *application) scrollComboPointer(point image.Point, bounds image.Rectangle, rows int) {
	count := len(app.controls[app.comboOpen].Items)
	maximum := max(0, count-rows)
	height := max(1, bounds.Dy()-app.layout.Scale(2))
	thumb := max(app.layout.Scale(8), height*rows/count)
	travel := max(1, height-thumb)
	position := point.Y - bounds.Min.Y - app.layout.Scale(1) - thumb/2

	app.comboFirst = max(0, min(maximum, (position*maximum+travel/2)/travel))

	app.invalidate()
}

func (app *application) revealComboItem() {
	_, rows := app.comboBounds()
	app.comboFirst = min(app.comboFirst, app.comboHover)

	if app.comboHover >= app.comboFirst+rows {
		app.comboFirst = max(0, app.comboHover-rows+1)
	}

	maximum := max(0, len(app.controls[app.comboOpen].Items)-rows)

	app.comboFirst = max(0, min(maximum, app.comboFirst))
}

func (app *application) chooseCombo(index, selected int) {
	control := &app.controls[index]
	if selected < 0 || selected >= len(control.Items) || selected == control.Selected || control.Disabled {
		return
	}

	control.Selected = selected

	app.changed(index)
}

func (app *application) comboKey(index int, key controlKey, modifiers keyModifiers) bool {
	open := app.comboOpen == index

	if key == controlF4 || modifiers.alt && (key == controlDown || key == controlUp) || !open && (key == ' ' || key == controlEnter) {
		app.toggleCombo(index)

		return true
	}

	if open && key == controlEscape {
		app.closeCombo()

		return true
	}

	if open && (key == controlEnter || key == ' ') {
		selected := app.comboHover

		app.closeCombo()
		app.chooseCombo(index, selected)

		return true
	}

	selected := app.controls[index].Selected

	if open {
		selected = app.comboHover
	}

	switch key {
	case controlUp, controlLeft:
		selected--
	case controlDown, controlRight:
		selected++
	case controlHome:
		selected = 0
	case controlEnd:
		selected = len(app.controls[index].Items) - 1
	case controlPageUp:
		selected -= 8
	case controlPageDown:
		selected += 8
	default:
		return false
	}

	selected = max(0, min(len(app.controls[index].Items)-1, selected))

	if open {
		app.comboHover = selected

		app.revealComboItem()
		app.invalidate()
	} else {
		app.chooseCombo(index, selected)
	}

	return true
}

func (app *application) findComboItem(index int, text string) {
	if text == "" || cleanInput(text) != text {
		return
	}

	control := &app.controls[index]
	count := len(control.Items)
	start := control.Selected

	if app.comboOpen == index {
		start = app.comboHover
	}

	for offset := 1; offset <= count; offset++ {
		selected := (start + offset + count) % count

		item := control.Items[selected]
		prefixEnd := 0

		for range text {
			_, size := utf8.DecodeRuneInString(item[prefixEnd:])
			prefixEnd += size
		}

		if strings.EqualFold(item[:prefixEnd], text) {
			if app.comboOpen == index {
				app.comboHover = selected

				app.revealComboItem()
				app.invalidate()
			} else {
				app.chooseCombo(index, selected)
			}

			return
		}
	}
}
