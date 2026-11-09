package window

import (
	"encoding/binary"
	"math"
	"strconv"
	"strings"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func linuxDPI(connection *xgb.Conn, screen *xproto.ScreenInfo) int {
	dpi := xsettingsDPI(connection)
	if dpi > 0 {
		return dpi
	}

	property, err := xproto.GetProperty(connection, false, screen.Root, xproto.AtomResourceManager, xproto.AtomString, 0, 65536).Reply()
	if err == nil && property != nil && property.Format == 8 {
		dpi = resourceDPI(string(property.Value))
		if dpi > 0 {
			return dpi
		}
	}

	// XWayland's physical screen dimensions are often synthetic; 96 is the fallback.
	return 96
}

func xsettingsDPI(connection *xgb.Conn) int {
	selection := "_XSETTINGS_S" + strconv.Itoa(connection.DefaultScreen)

	atom, err := xproto.InternAtom(connection, true, uint16(len(selection)), selection).Reply()
	if err != nil || atom == nil || atom.Atom == 0 {
		return 0
	}

	owner, err := xproto.GetSelectionOwner(connection, atom.Atom).Reply()
	if err != nil || owner == nil || owner.Owner == 0 {
		return 0
	}

	atom, err = xproto.InternAtom(connection, true, uint16(len("_XSETTINGS_SETTINGS")), "_XSETTINGS_SETTINGS").Reply()
	if err != nil || atom == nil || atom.Atom == 0 {
		return 0
	}

	property, err := xproto.GetProperty(connection, false, owner.Owner, atom.Atom, atom.Atom, 0, 65536).Reply()
	if err != nil || property == nil || property.Format != 8 || property.BytesAfter != 0 {
		return 0
	}

	return parseXsettingsDPI(property.Value)
}

func parseXsettingsDPI(data []byte) int {
	if len(data) < 12 || data[0] > 1 {
		return 0
	}

	var order binary.ByteOrder = binary.LittleEndian

	if data[0] == 1 {
		order = binary.BigEndian
	}

	count := order.Uint32(data[8:12])
	offset := 12

	for range count {
		if len(data)-offset < 4 {
			return 0
		}

		kind := data[offset]
		nameLength := int(order.Uint16(data[offset+2 : offset+4]))

		offset += 4
		paddedLength := (nameLength + 3) &^ 3

		if len(data)-offset < paddedLength+4 {
			return 0
		}

		isDPI := string(data[offset:offset+nameLength]) == "Xft/DPI"
		offset += paddedLength + 4 // Each setting includes a last-change serial.

		switch kind {
		case 0:
			if len(data)-offset < 4 {
				return 0
			}

			value := int32(order.Uint32(data[offset : offset+4]))
			offset += 4

			if isDPI && value > 0 {
				return clampDPI(float64(value) / 1024)
			}
		case 1:
			if len(data)-offset < 4 {
				return 0
			}

			length := uint64(order.Uint32(data[offset : offset+4]))

			offset += 4
			padded := (length + 3) &^ 3

			if padded > uint64(len(data)-offset) {
				return 0
			}

			offset += int(padded)
		case 2:
			if len(data)-offset < 8 {
				return 0
			}

			offset += 8
		default:
			return 0
		}
	}

	return 0
}

func resourceDPI(resources string) int {
	for line := range strings.Lines(resources) {
		name, value, found := strings.Cut(line, ":")
		if !found || strings.TrimSpace(name) != "Xft.dpi" {
			continue
		}

		dpi, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err == nil && !math.IsNaN(dpi) && !math.IsInf(dpi, 0) && dpi > 0 {
			return clampDPI(dpi)
		}
	}

	return 0
}

func clampDPI(dpi float64) int {
	return int(math.Round(max(72, min(dpi, 384))))
}
