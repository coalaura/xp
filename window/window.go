// Package window creates custom-drawn desktop windows with Windows XP's blue Luna frame.
package window

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

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
	Text(id string) (string, error)
	SetText(id, text string) error
	Checked(id string) (bool, error)
	// SetChecked updates a checkbox or radio without invoking callbacks.
	SetChecked(id string, checked bool) error
	CheckState(id string) (ui.CheckState, error)
	SetCheckState(id string, state ui.CheckState) error
	Value(id string) (int, error)
	SetValue(id string, value int) error
	Selected(id string) (int, error)
	SetSelected(id string, index int) error
	SetDisabled(id string, disabled bool) error
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
	// OnClick receives button, link, checkbox and radio activations, after state changes.
	OnClick func(Window, string)
	// OnChange runs after a user changes an input, choice, slider or combo selection.
	// Programmatic setters do not invoke callbacks.
	OnChange func(Window, string)
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
		if control.Kind > ui.Group || control.Bounds.Empty() {
			return errors.New("xp: controls need a valid kind and nonempty bounds")
		}

		if !validText(control.Text) {
			return errors.New("xp: control text must be valid UTF-8 without NUL characters")
		}

		if control.Checked && control.Kind != ui.Checkbox && control.Kind != ui.Radio {
			return errors.New("xp: only checkboxes and radios can be checked")
		}

		if control.Mixed && control.Checked || (control.Mixed || control.ThreeState) && control.Kind != ui.Checkbox {
			return errors.New("xp: mixed and three-state apply only to checkboxes; mixed and checked are exclusive")
		}

		if control.Kind == ui.Radio {
			for previous := range index {
				peer := &options.Controls[previous]
				if control.Checked && peer.Kind == ui.Radio && peer.RadioGroup == control.RadioGroup && peer.Checked {
					return errors.New("xp: a radio group can have only one selected item")
				}
			}
		}

		if control.Kind == ui.Slider {
			if control.Min < -2147483648 || control.Max > 2147483647 || control.Max <= control.Min || control.Value < control.Min || control.Value > control.Max || control.Step < 0 || int64(control.Step) > int64(control.Max)-int64(control.Min) {
				return errors.New("xp: sliders need a valid 32-bit range, value and nonnegative step")
			}
		}

		if control.Kind == ui.Input && !validSingleLine(control.Text) {
			return errors.New("xp: inputs are single-line")
		}

		if control.Kind == ui.ComboBox {
			if control.Selected < -1 || control.Selected >= len(control.Items) && (len(control.Items) != 0 || control.Selected != 0) {
				return errors.New("xp: combo selection is outside its item list")
			}

			for _, item := range control.Items {
				if !validSingleLine(item) {
					return errors.New("xp: combo items must be valid single-line text")
				}
			}
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

func validText(text string) bool {
	return utf8.ValidString(text) && !strings.ContainsRune(text, 0)
}

func validSingleLine(text string) bool {
	if !validText(text) {
		return false
	}

	for _, current := range text {
		if unicode.IsControl(current) || current == '\u2028' || current == '\u2029' {
			return false
		}
	}

	return true
}
