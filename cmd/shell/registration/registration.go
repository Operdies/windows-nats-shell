package registration

import (
	"syscall"
	"unsafe"
)

var (
	user32                = syscall.MustLoadDLL("user32.dll")
	systemParametersInfoA = user32.MustFindProc("SystemParametersInfoA")
)

type tagMINIMIZEDMETRICS struct {
	cbSize   uint32
	iWidth   int32
	iHorzGap int32
	iVertGap int32
	iArrange int32
}

const (
	ARW_HIDE                = 0x0008
	SPI_SETMINIMIZEDMETRICS = 0x002C
)

func RegisterThisProcessAsShell() {
	var min tagMINIMIZEDMETRICS
	min.iArrange = ARW_HIDE
	min.cbSize = uint32(unsafe.Sizeof(min))

	minptr := unsafe.Pointer(&min)

	// This call is required in order to receive shell events.
	// It also hides minimized windows so there is no pseudo-taskbar
	systemParametersInfoA.Call(SPI_SETMINIMIZEDMETRICS, 0, uintptr(minptr), 0)
}
