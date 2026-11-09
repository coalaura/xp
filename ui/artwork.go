package ui

import (
	_ "embed"

	"github.com/coalaura/xp/internal/artwork"
)

// Unchecked, checked, then mixed; normal, hot, pressed, disabled in each group.
//
//go:embed checkbox.png
var checkboxPNG []byte

// Text-free normal, hot, pressed, disabled, and default button surfaces.
//
//go:embed button.png
var buttonPNG []byte

//go:embed radio.png
var radioPNG []byte

//go:embed input.png
var inputPNG []byte

//go:embed combo.png
var comboPNG []byte

//go:embed slider.png
var sliderPNG []byte

//go:embed slider_vertical.png
var sliderVerticalPNG []byte

//go:embed track.png
var trackPNG []byte

//go:embed track_vertical.png
var trackVerticalPNG []byte

//go:embed group.png
var groupPNG []byte

var (
	checkboxArtwork       = artwork.Decode(checkboxPNG, 13, 156)
	buttonArtwork         = artwork.Decode(buttonPNG, 112, 130)
	radioArtwork          = artwork.Decode(radioPNG, 13, 104)
	inputArtwork          = artwork.Decode(inputPNG, 111, 115)
	comboArtwork          = artwork.Decode(comboPNG, 17, 88)
	sliderArtwork         = artwork.Decode(sliderPNG, 11, 105)
	sliderVerticalArtwork = artwork.Decode(sliderVerticalPNG, 21, 55)
	trackArtwork          = artwork.Decode(trackPNG, 200, 5)
	trackVerticalArtwork  = artwork.Decode(trackVerticalPNG, 5, 32)
	groupArtwork          = artwork.Decode(groupPNG, 112, 56)
)
