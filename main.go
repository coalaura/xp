//go:generate go run ./internal/reference

package main

import (
	"image"
	"log"

	"github.com/coalaura/xp/ui"
	"github.com/coalaura/xp/window"
)

func main() {
	err := window.Run(window.Options{
		Title:     "Windows XP",
		Width:     520,
		Height:    410,
		MinWidth:  440,
		MinHeight: 390,
		Controls: []ui.Control{
			{Kind: ui.Label, Bounds: image.Rect(20, 20, 490, 40), Text: "Welcome to Windows XP", Bold: true, Anchor: ui.Left | ui.Right},
			{Kind: ui.Label, Bounds: image.Rect(20, 49, 490, 88), Text: "A familiar blue title bar, warm beige surfaces, and the little details that make Luna feel like home.", Anchor: ui.Left | ui.Right},
			{Kind: ui.Separator, Bounds: image.Rect(20, 102, 494, 104), Anchor: ui.Left | ui.Right},
			{Kind: ui.Label, Bounds: image.Rect(20, 121, 490, 140), Text: "Buttons", Bold: true},
			{ID: "test", Kind: ui.Button, Bounds: image.Rect(20, 151, 120, 176), Text: "Click me"},
			{Kind: ui.Button, Bounds: image.Rect(132, 151, 232, 176), Text: "Disabled", Disabled: true},
			{ID: "remember", Kind: ui.Checkbox, Bounds: image.Rect(20, 193, 232, 214), Text: "Remember my choice"},
			{Kind: ui.Checkbox, Bounds: image.Rect(254, 193, 470, 214), Text: "Disabled option", Disabled: true},
			{Kind: ui.Checkbox, Bounds: image.Rect(254, 219, 470, 240), Text: "Disabled, checked", Disabled: true, Checked: true},
			{ID: "status", Kind: ui.Label, Bounds: image.Rect(20, 258, 490, 298), Text: "Try the buttons and checkboxes, or use Tab and Space to explore keyboard focus.", Anchor: ui.Left | ui.Right},
			{Kind: ui.Separator, Bounds: image.Rect(20, 320, 494, 322), Anchor: ui.Left | ui.Right | ui.Bottom},
			{ID: "ok", Kind: ui.Button, Bounds: image.Rect(328, 336, 406, 359), Text: "OK", Default: true, Anchor: ui.Right | ui.Bottom},
			{ID: "cancel", Kind: ui.Button, Bounds: image.Rect(416, 336, 494, 359), Text: "Cancel", Cancel: true, Anchor: ui.Right | ui.Bottom},
		},
		OnClick: handleClick,
	})

	if err != nil {
		log.Fatal(err)
	}
}

func handleClick(current window.Window, id string) {
	switch id {
	case "test":
		err := current.SetText("status", "It works! This text was updated through the reusable window API.")
		if err != nil {
			log.Print(err)
		}
	case "remember":
		checked, err := current.Checked(id)
		if err != nil {
			log.Print(err)

			return
		}

		text := "The checkbox is unchecked."

		if checked {
			text = "The checkbox is checked."
		}

		err = current.SetText("status", text)
		if err != nil {
			log.Print(err)
		}
	case "ok", "cancel":
		current.Close()
	}
}
