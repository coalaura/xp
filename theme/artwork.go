package theme

import (
	_ "embed"

	"github.com/coalaura/xp/internal/artwork"
)

//go:embed icon.png
var iconPNG []byte

// Bundled caption artwork: minimize, maximize, restore, close; normal, hot, pressed each.
//
//go:embed caption.png
var captionPNG []byte

//go:embed inactive.png
var inactivePNG []byte

//go:embed title.png
var titlePNG []byte

//go:embed title_maximized.png
var titleMaximizedPNG []byte

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
	iconArtwork           = artwork.Decode(iconPNG, 16, 16)
	captionArtwork        = artwork.Decode(captionPNG, 21, 252)
	inactiveArtwork       = artwork.Decode(inactivePNG, 21, 84)
	titleArtwork          = artwork.Decode(titlePNG, 33, 60)
	titleMaximizedArtwork = artwork.Decode(titleMaximizedPNG, 1, 48)
	frameArtwork          = artwork.Decode(framePNG, 9, 1328)
	dialogArtwork         = artwork.Decode(dialogPNG, 7, 1328)
	bottomArtwork         = artwork.Decode(bottomPNG, 968, 8)
	dialogBottomArtwork   = artwork.Decode(dialogBottomPNG, 966, 6)
)
