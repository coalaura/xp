//go:build !windows

package window

func run(_ Options) error {
	return ErrUnsupportedPlatform
}
