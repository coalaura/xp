![xp banner](.github/banner.png)

A small Go library for desktop windows with Windows XP's blue Luna look, plus a demo application. The frame and controls are custom-drawn, with DPI scaling, keyboard navigation and window-manager integration.

![demo screenshot](.github/screenshot.png)

- Resizable windows and fixed-size dialogs.
- Headings, text, links, buttons, three-state checkboxes, grouped radios, sliders, single-line inputs, list-only combo boxes, separators, sunken/raised edges and group frames, with anchors for resizing.
- Windows support through Win32/GDI and Linux support through X11, including Wayland sessions with XWayland.

## Try the demo

Requires Go 1.27.3 or newer. From the repository root:

```sh
go run .
```

To build a fully static Linux binary, run on Linux with CGO disabled:

```sh
CGO_ENABLED=0 go build -o xp .
```

The Linux backend embeds its fonts and connects to the display through `DISPLAY`. Wayland sessions require XWayland.

## Use as a library

```sh
go get github.com/coalaura/xp
```

```go
package main

import (
	"image"
	"log"

	"github.com/coalaura/xp/ui"
	"github.com/coalaura/xp/window"
)

func main() {
	err := window.Run(window.Options{
		Title:  "Hello, Luna",
		Style:  window.Dialog,
		Width:  320,
		Height: 160,
		Controls: []ui.Control{
			{Kind: ui.Label, Bounds: image.Rect(20, 20, 290, 45), Text: "A little nostalgia, written in Go."},
			{ID: "close", Kind: ui.Button, Bounds: image.Rect(210, 80, 290, 105), Text: "OK", Default: true, Cancel: true},
		},
		OnClick: func(current window.Window, id string) {
			if id == "close" {
				current.Close()
			}
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

`window.Run` blocks until the window closes. Window dimensions include the frame; control bounds are relative to the content area. Both use logical pixels at 96 DPI.

Use `Anchor` to move or stretch controls when resizing, such as `ui.Right | ui.Bottom` for a bottom-right button. `Default` and `Cancel` assign Enter and Escape actions; Tab moves focus and Space activates the focused control.

`OnClick` receives button, link, checkbox and radio activations. Links invoke your callback; they do not open URLs automatically. `OnChange` receives user changes to inputs, checkboxes, radios, sliders and combo selections after the new value is stored. Radio changes report the newly selected control. Programmatic setters do not invoke either callback. Keep callbacks brief and call window methods only on the callback's UI thread.

| Control | Configuration and access |
| --- | --- |
| `ui.Heading`, `ui.Text`, `ui.Link` | `Text`, `Disabled`, optional `Bold`; `ui.Label` remains an alias for text. Use `Text` / `SetText` to read or update captions. |
| `ui.Checkbox` | `Checked` or `Mixed`; `ThreeState` cycles unchecked, checked, mixed. Use `Checked` / `SetChecked` or `CheckState` / `SetCheckState` with `ui.Unchecked`, `ui.Checked`, `ui.Mixed`. |
| `ui.Radio` | Shared `RadioGroup` values make a mutually exclusive group; an empty group name is also a group. Use `Checked` / `SetChecked`. Arrow keys move within a group; Tab visits its selected or first enabled item. |
| `ui.Slider` | Set `Min`, `Max` and `Value`; `Max` must exceed `Min`. `Step` defaults to 1. `Vertical` rotates the slider, with values increasing downwards. Use `Value` / `SetValue`; arrows, Page Up/Down and Home/End also adjust it. |
| `ui.Input` | Single-line `Text`, optional `Password` or `ReadOnly`. Supports mouse selection, Shift/Ctrl navigation, Ctrl+A/C/X/V and single-level undo/redo with Ctrl+Z/Y. Password inputs mask text and disable copying/cutting. Use `Text` / `SetText`. |
| `ui.ComboBox` | Supply `Items` and a `Selected` index; -1 means no selection. Use `Selected` / `SetSelected`; `Text` returns the selected item. F4 or Alt+Down opens the list, arrows navigate, Enter commits and Escape cancels. Typing finds the next matching prefix. |
| `ui.Separator`, `ui.SunkenEdge`, `ui.RaisedEdge`, `ui.Group` | Separators support `Vertical`. Group frames use `Text` and `Disabled`; they are visual decorations, not containers. Position their controls separately and draw the group first. |

All controls support `Disabled`; `SetDisabled` updates it at runtime. Combo lists are drawn inside the window content area, above or below the field as space allows, and scroll when there are more items than fit. Text editing uses Unicode code points; IME composition is not implemented.

See [`main.go`](main.go) for the full demo. [`window`](window) provides the desktop API, [`ui`](ui) defines controls and [`theme`](theme) contains the shared Luna renderer.

## License

[Mozilla Public License 2.0](LICENSE).
