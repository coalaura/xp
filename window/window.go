// Package window creates custom-drawn desktop windows with Windows XP's blue Luna frame.
package window

import (
	"errors"
	"strings"

	"github.com/coalaura/xp/ui"
)

const (
	// Standard has an icon, all three caption buttons, and a resizable border.
	Standard Style = iota
	// Dialog has a compact caption, only a close button, and a fixed-size border.
	Dialog
)

// Style selects the window's decorations and native sizing behavior.
type Style uint8

// Window is available to event callbacks. Its methods must be called on the
// callback's UI thread. Callbacks should return promptly to keep the UI responsive.
type Window interface {
	Close()
	SetText(id, text string) error
	Checked(id string) (bool, error)
	// SetChecked updates a checkbox without invoking OnClick.
	SetChecked(id string, checked bool) error
}

// Options describes a window. Dimensions include the frame and are in logical
// pixels at 96 DPI. Zero dimensions select the defaults.
type Options struct {
	Title     string
	Style     Style
	Width     int
	Height    int
	MinWidth  int
	MinHeight int
	Controls  []ui.Control
	// OnClick receives the control ID after activation. Checkboxes toggle before the callback.
	OnClick func(Window, string)
}

// ErrUnsupportedPlatform is returned on platforms without a window backend.
var ErrUnsupportedPlatform = errors.New("xp: this platform does not have a window backend")

// Run opens a window and blocks until it closes. Each call owns its event loop
// and OS thread. Concurrent calls can create independent windows.
func Run(options Options) error {
	if options.Style != Standard && options.Style != Dialog {
		return errors.New("xp: invalid window style")
	}

	if strings.ContainsRune(options.Title, 0) {
		return errors.New("xp: title contains a NUL character")
	}

	if options.Width == 0 {
		options.Width = 640
	}

	if options.Height == 0 {
		options.Height = 420
	}

	if options.MinWidth == 0 {
		options.MinWidth = 160
	}

	if options.MinHeight == 0 {
		options.MinHeight = 64
	}

	if options.MinWidth < 120 || options.MinHeight < 40 {
		return errors.New("xp: minimum dimensions must be at least 120 x 40")
	}

	if options.Width < options.MinWidth || options.Height < options.MinHeight {
		return errors.New("xp: window dimensions are smaller than the minimum dimensions")
	}

	if options.Width > 16384 || options.Height > 16384 {
		return errors.New("xp: window dimensions exceed 16384 logical pixels")
	}

	ids := make(map[string]struct{}, len(options.Controls))

	var (
		defaultButton bool
		cancelButton  bool
	)

	for index := range options.Controls {
		control := &options.Controls[index]
		if control.Kind > ui.Checkbox || control.Bounds.Empty() {
			return errors.New("xp: controls need a valid kind and nonempty bounds")
		}

		if strings.ContainsRune(control.Text, 0) {
			return errors.New("xp: control text contains a NUL character")
		}

		if control.Checked && control.Kind != ui.Checkbox {
			return errors.New("xp: only checkboxes can be checked")
		}

		if control.ID != "" {
			if _, exists := ids[control.ID]; exists {
				return errors.New("xp: control IDs must be unique")
			}

			ids[control.ID] = struct{}{}
		}

		if control.Default || control.Cancel {
			if control.Kind != ui.Button || control.Default && defaultButton || control.Cancel && cancelButton {
				return errors.New("xp: default and cancel actions must each identify a single button")
			}

			defaultButton = defaultButton || control.Default
			cancelButton = cancelButton || control.Cancel
		}
	}

	return run(options)
}
