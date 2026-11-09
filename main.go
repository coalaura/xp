//go:generate go run ./internal/reference

package main

import (
	"image"
	"log"
	"strconv"

	"github.com/coalaura/xp/ui"
	"github.com/coalaura/xp/window"
)

func main() {
	err := window.Run(window.Options{
		Title:     "Windows XP",
		Width:     680,
		Height:    620,
		MinWidth:  680,
		MinHeight: 620,
		Controls: []ui.Control{
			{Kind: ui.Heading, Bounds: image.Rect(20, 18, 320, 38), Text: "Welcome to Windows XP"},
			{Kind: ui.Heading, Bounds: image.Rect(340, 18, 650, 38), Text: "Disabled heading", Disabled: true},
			{Kind: ui.Text, Bounds: image.Rect(20, 46, 650, 65), Text: "Classic Luna controls, drawn from real Windows XP reference captures."},
			{Kind: ui.Text, Bounds: image.Rect(340, 73, 650, 91), Text: "Disabled text", Disabled: true},
			{ID: "link", Kind: ui.Link, Bounds: image.Rect(20, 73, 155, 91), Text: "Explore the controls"},
			{Kind: ui.Link, Bounds: image.Rect(180, 73, 310, 91), Text: "Disabled link", Disabled: true},
			{Kind: ui.Group, Bounds: image.Rect(20, 110, 320, 302), Text: "Choices"},
			{Kind: ui.Group, Bounds: image.Rect(340, 110, 650, 302), Text: "Text and selection"},
			{ID: "remember", Kind: ui.Checkbox, Bounds: image.Rect(34, 136, 300, 157), Text: "Remember my choice"},
			{ID: "mixed", Kind: ui.Checkbox, Bounds: image.Rect(34, 165, 300, 186), Text: "Three-state checkbox", Mixed: true, ThreeState: true},
			{Kind: ui.Checkbox, Bounds: image.Rect(34, 194, 300, 215), Text: "Disabled, checked", Disabled: true, Checked: true},
			{ID: "blue", Kind: ui.Radio, Bounds: image.Rect(34, 234, 160, 255), Text: "Blue", RadioGroup: "color", Checked: true},
			{ID: "silver", Kind: ui.Radio, Bounds: image.Rect(170, 234, 300, 255), Text: "Silver", RadioGroup: "color"},
			{Kind: ui.Radio, Bounds: image.Rect(34, 264, 300, 285), Text: "Unavailable", RadioGroup: "color", Disabled: true},
			{ID: "name", Kind: ui.Input, Bounds: image.Rect(354, 136, 636, 159), Text: "Type something here"},
			{ID: "password", Kind: ui.Input, Bounds: image.Rect(354, 169, 636, 192), Text: "password", Password: true},
			{Kind: ui.Input, Bounds: image.Rect(354, 202, 488, 225), Text: "Read-only", ReadOnly: true},
			{Kind: ui.Input, Bounds: image.Rect(502, 202, 636, 225), Text: "Disabled", Disabled: true},
			{ID: "theme", Kind: ui.ComboBox, Bounds: image.Rect(354, 257, 636, 281), Items: []string{"Windows XP", "Windows Classic", "Silver", "Olive Green"}},
			{ID: "test", Kind: ui.Button, Bounds: image.Rect(20, 321, 132, 347), Text: "Click me"},
			{Kind: ui.Button, Bounds: image.Rect(146, 321, 258, 347), Text: "Disabled", Disabled: true},
			{Kind: ui.Text, Bounds: image.Rect(20, 365, 300, 383), Text: "Volume (arrows, Home/End, or drag)"},
			{ID: "volume", Kind: ui.Slider, Bounds: image.Rect(20, 389, 300, 419), Min: 0, Max: 100, Value: 50, Step: 5},
			{Kind: ui.Slider, Bounds: image.Rect(340, 389, 550, 419), Min: 0, Max: 100, Value: 25, Disabled: true},
			{ID: "vertical", Kind: ui.Slider, Bounds: image.Rect(600, 320, 630, 425), Min: 0, Max: 100, Value: 50, Vertical: true},
			{Kind: ui.SunkenEdge, Bounds: image.Rect(20, 440, 220, 471)},
			{Kind: ui.Text, Bounds: image.Rect(30, 447, 210, 465), Text: "Sunken edge"},
			{Kind: ui.RaisedEdge, Bounds: image.Rect(240, 440, 440, 471)},
			{Kind: ui.Text, Bounds: image.Rect(250, 447, 430, 465), Text: "Raised edge"},
			{Kind: ui.Group, Bounds: image.Rect(460, 438, 650, 471), Text: "Disabled group", Disabled: true},
			{ID: "status", Kind: ui.Text, Bounds: image.Rect(20, 488, 650, 522), Text: "Use Tab to move focus. Inputs support selection, clipboard shortcuts and undo.", Anchor: ui.Left | ui.Right | ui.Bottom},
			{Kind: ui.Separator, Bounds: image.Rect(20, 534, 650, 536), Anchor: ui.Left | ui.Right | ui.Bottom},
			{ID: "ok", Kind: ui.Button, Bounds: image.Rect(484, 550, 562, 573), Text: "OK", Default: true, Anchor: ui.Right | ui.Bottom},
			{ID: "cancel", Kind: ui.Button, Bounds: image.Rect(572, 550, 650, 573), Text: "Cancel", Cancel: true, Anchor: ui.Right | ui.Bottom},
		},
		OnClick:  handleClick,
		OnChange: handleChange,
	})

	if err != nil {
		log.Fatal(err)
	}
}

func handleClick(current window.Window, id string) {
	switch id {
	case "test", "link":
		err := current.SetText("status", "It works! This text was updated through the reusable window API.")
		if err != nil {
			log.Print(err)
		}
	case "ok", "cancel":
		current.Close()
	}
}

func handleChange(current window.Window, id string) {
	text := "Changed: " + id

	switch id {
	case "volume", "vertical":
		value, err := current.Value(id)
		if err != nil {
			log.Print(err)

			return
		}

		text = "Slider value: " + strconv.Itoa(value)
	case "theme":
		selected, err := current.Text(id)
		if err != nil {
			log.Print(err)

			return
		}

		text = "Selected: " + selected
	case "password":
		text = "Password changed."
	}

	err := current.SetText("status", text)
	if err != nil {
		log.Print(err)
	}
}
