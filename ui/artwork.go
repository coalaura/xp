package ui

import (
	_ "embed"

	"github.com/coalaura/xp/internal/artwork"
)

// Unchecked then checked; normal, hot, pressed, disabled in each group.
//
//go:embed checkbox.png
var checkboxPNG []byte

// Text-free normal, hot, pressed, disabled, and default button surfaces.
//
//go:embed button.png
var buttonPNG []byte

var (
	checkboxArtwork = artwork.Decode(checkboxPNG, 13, 104)
	buttonArtwork   = artwork.Decode(buttonPNG, 112, 130)
)
