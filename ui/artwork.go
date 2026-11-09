package ui

import (
	_ "embed"

	"github.com/coalaura/xp/internal/artwork"
)

// VM captures: unchecked, checked, hot; each checkbox occupies 13 x 13 pixels.
//
//go:embed checkbox.png
var checkboxPNG []byte

// Default and hot buttons, reduced to four left columns, a center and four right.
//
//go:embed button.png
var buttonPNG []byte

var (
	checkboxArtwork = artwork.Decode(checkboxPNG, 13, 39)
	buttonArtwork   = artwork.Decode(buttonPNG, 9, 42)
)
