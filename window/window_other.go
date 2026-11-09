//go:build !windows && !linux

package window

func run(_ Options) error {
	return ErrUnsupportedPlatform
}
