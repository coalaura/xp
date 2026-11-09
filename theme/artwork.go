package theme

import (
	_ "embed"

	"github.com/coalaura/xp/internal/artwork"
)

// Bundled caption artwork: minimize, maximize, restore, close; normal, hot, pressed each.
//
//go:embed caption.png
var captionPNG []byte

//go:embed inactive.png
var inactivePNG []byte

//go:embed inactive_maximized.png
var inactiveMaximizedPNG []byte

// Body-frame atlases contain the left edge, one center column and right edge.
//
//go:embed frame.png
var framePNG []byte

//go:embed dialog.png
var dialogPNG []byte

//go:embed bottom.png
var bottomPNG []byte

//go:embed dialog_bottom.png
var dialogBottomPNG []byte

var (
	captionArtwork           = artwork.Decode(captionPNG, 21, 252)
	inactiveArtwork          = artwork.Decode(inactivePNG, 21, 84)
	inactiveMaximizedArtwork = artwork.Decode(inactiveMaximizedPNG, 21, 63)
	frameArtwork             = artwork.Decode(framePNG, 9, 1140)
	dialogArtwork            = artwork.Decode(dialogPNG, 7, 161)
	bottomArtwork            = artwork.Decode(bottomPNG, 800, 8)
	dialogBottomArtwork      = artwork.Decode(dialogBottomPNG, 381, 3)
)
