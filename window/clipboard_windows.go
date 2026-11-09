package window

import (
	"syscall"
	"unsafe"
)

func (app *application) copyText(text string) bool {
	opened, _, _ := openClipboard.Call(app.handle)
	if opened == 0 {
		return false
	}

	defer closeClipboard.Call()

	app.encodeText(text)

	encoded := app.textBuffer

	memory, _, _ := globalAlloc.Call(0x0042, uintptr(len(encoded)*2))
	if memory == 0 {
		return false
	}

	address, _, _ := globalLock.Call(memory)
	if address == 0 {
		globalFree.Call(memory)

		return false
	}

	moveMemory.Call(address, uintptr(unsafe.Pointer(&encoded[0])), uintptr(len(encoded)*2))
	globalUnlock.Call(memory)
	emptyClipboard.Call()

	result, _, _ := setClipboardData.Call(13, memory)
	if result == 0 {
		globalFree.Call(memory)

		return false
	}

	// A successful SetClipboardData transfers ownership to the system.
	return true
}

func (app *application) pasteText(index int) {
	opened, _, _ := openClipboard.Call(app.handle)
	if opened == 0 {
		return
	}

	defer closeClipboard.Call()

	memory, _, _ := getClipboardData.Call(13)
	if memory == 0 {
		return
	}

	size, _, _ := globalSize.Call(memory)
	if size < 2 || size > 16<<20 {
		return
	}

	address, _, _ := globalLock.Call(memory)
	if address == 0 {
		return
	}

	encoded := make([]uint16, int(size/2))

	moveMemory.Call(uintptr(unsafe.Pointer(&encoded[0])), address, size)
	globalUnlock.Call(memory)

	text := syscall.UTF16ToString(encoded)

	app.replaceInput(index, cleanInput(text))
}
