//go:build windows || linux

package window

import (
	"image"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/coalaura/xp/ui"
)

type inputState struct {
	caret      int
	anchor     int
	scroll     int
	display    string
	undo       string
	undoCaret  int
	undoAnchor int
	hasUndo    bool
}

func (app *application) refreshInput(index int) {
	control := &app.controls[index]
	input := &app.inputs[index]
	input.display = control.Text

	if control.Password {
		input.display = strings.Repeat("*", utf8.RuneCountInString(control.Text))
	}
}

func (app *application) inputDisplayOffset(index, offset int) int {
	control := &app.controls[index]
	if control.Password {
		return utf8.RuneCountInString(control.Text[:offset])
	}

	return offset
}

func (app *application) inputPosition(index, horizontal int) int {
	control := &app.controls[index]
	input := &app.inputs[index]
	bounds := ui.FieldTextBounds(app.layout, app.controlRectangle(index), ui.Input)
	position := horizontal - bounds.Min.X + input.scroll
	previousWidth := 0

	for offset, current := range control.Text {
		next := offset + utf8.RuneLen(current)
		displayEnd := app.inputDisplayOffset(index, next)

		width := app.textWidth(input.display[:displayEnd], control.Bold)
		if position < (previousWidth+width+1)/2 {
			return offset
		}

		previousWidth = width
	}

	return len(control.Text)
}

func (app *application) revealCaret(index int) image.Rectangle {
	control := &app.controls[index]
	input := &app.inputs[index]

	bounds := ui.FieldTextBounds(app.layout, app.controlRectangle(index), ui.Input)
	offset := app.inputDisplayOffset(index, input.caret)
	width := app.textWidth(input.display[:offset], control.Bold)
	available := max(0, bounds.Dx()-app.layout.Scale(1))

	input.scroll = max(0, min(input.scroll, width))

	if width-input.scroll > available {
		input.scroll = width - available
	}

	total := app.textWidth(input.display, control.Bold)
	input.scroll = min(input.scroll, max(0, total-available))

	return bounds
}

func (app *application) replaceInput(index int, inserted string) {
	control := &app.controls[index]
	if control.Disabled || control.ReadOnly {
		return
	}

	input := &app.inputs[index]

	start := min(input.caret, input.anchor)
	end := max(input.caret, input.anchor)

	if start == end && inserted == "" {
		return
	}

	var builder strings.Builder

	builder.Grow(len(control.Text) - (end - start) + len(inserted))
	builder.WriteString(control.Text[:start])
	builder.WriteString(inserted)
	builder.WriteString(control.Text[end:])

	updated := builder.String()

	if updated == control.Text {
		input.caret = start + len(inserted)
		input.anchor = input.caret

		app.invalidate()

		return
	}

	input.undo = control.Text
	input.undoCaret = input.caret
	input.undoAnchor = input.anchor
	input.hasUndo = true

	control.Text = updated

	input.caret = start + len(inserted)
	input.anchor = input.caret

	app.refreshInput(index)
	app.changed(index)
}

func (app *application) typeText(text string) {
	index := app.controlFocus
	if index < 0 || !app.controls[index].Interactive() {
		return
	}

	if app.controls[index].Kind == ui.ComboBox {
		app.findComboItem(index, text)

		return
	}

	if app.controls[index].Kind != ui.Input || app.controls[index].ReadOnly {
		return
	}

	text = cleanInput(text)
	if text != "" {
		app.replaceInput(index, text)
	}
}

func (app *application) editKey(index int, key controlKey, modifiers keyModifiers) bool {
	control := &app.controls[index]
	input := &app.inputs[index]
	position := input.caret

	start := min(input.caret, input.anchor)
	end := max(input.caret, input.anchor)

	if modifiers.control {
		switch key {
		case 'a', 'A':
			input.anchor = 0
			input.caret = len(control.Text)

			app.invalidate()

			return true
		case 'c', 'C', 'x', 'X':
			if !control.Password && start != end {
				copied := app.copyText(control.Text[start:end])
				if copied && (key == 'x' || key == 'X') {
					app.replaceInput(index, "")
				}
			}

			return true
		case 'v', 'V':
			if !control.ReadOnly {
				app.pasteText(index)
			}

			return true
		case 'z', 'Z', 'y', 'Y':
			if !control.ReadOnly && input.hasUndo {
				control.Text, input.undo = input.undo, control.Text
				input.caret, input.undoCaret = input.undoCaret, input.caret
				input.anchor, input.undoAnchor = input.undoAnchor, input.anchor

				app.refreshInput(index)
				app.changed(index)
			}

			return true
		}
	}

	switch key {
	case controlLeft:
		position = previousRune(control.Text, position)

		if modifiers.control {
			position = previousWord(control.Text, input.caret)
		} else if start != end && !modifiers.shift {
			position = start
		}
	case controlRight:
		position = nextRune(control.Text, position)

		if modifiers.control {
			position = nextWord(control.Text, input.caret)
		} else if start != end && !modifiers.shift {
			position = end
		}
	case controlHome:
		position = 0
	case controlEnd:
		position = len(control.Text)
	case controlBackspace, controlDelete:
		if !control.ReadOnly {
			if start == end {
				if key == controlBackspace {
					input.anchor = previousRune(control.Text, input.caret)

					if modifiers.control {
						input.anchor = previousWord(control.Text, input.caret)
					}
				} else {
					input.anchor = nextRune(control.Text, input.caret)

					if modifiers.control {
						input.anchor = nextWord(control.Text, input.caret)
					}
				}
			}

			app.replaceInput(index, "")
		}

		return true
	case ' ':
		return true // Text input is delivered separately by the platform.
	default:
		return false
	}

	input.caret = position

	if !modifiers.shift {
		input.anchor = position
	}

	app.invalidate()

	return true
}

func cleanInput(text string) string {
	return strings.Map(inputRune, text)
}

func inputRune(current rune) rune {
	if unicode.IsControl(current) || current == '\u2028' || current == '\u2029' {
		return -1
	}

	return current
}

func previousRune(text string, position int) int {
	_, size := utf8.DecodeLastRuneInString(text[:position])

	return position - size
}

func nextRune(text string, position int) int {
	_, size := utf8.DecodeRuneInString(text[position:])

	return position + size
}

func previousWord(text string, position int) int {
	for position > 0 {
		current, size := utf8.DecodeLastRuneInString(text[:position])
		if !unicode.IsSpace(current) {
			break
		}

		position -= size
	}

	for position > 0 {
		current, size := utf8.DecodeLastRuneInString(text[:position])
		if unicode.IsSpace(current) {
			break
		}

		position -= size
	}

	return position
}

func nextWord(text string, position int) int {
	for position < len(text) {
		current, size := utf8.DecodeRuneInString(text[position:])
		if unicode.IsSpace(current) {
			break
		}

		position += size
	}

	for position < len(text) {
		current, size := utf8.DecodeRuneInString(text[position:])
		if !unicode.IsSpace(current) {
			break
		}

		position += size
	}

	return position
}
