//go:build darwin || linux

package terminal

// Unix terminal constants.
// The actual values differ between macOS and Linux but are set
// correctly in platform-specific files.
var (
	_TCGETS uintptr
	_TCSETS uintptr
)

// Termios contains terminal attributes for Unix-like systems.
type Termios struct {
	Iflag  uint64
	Oflag  uint64
	Cflag  uint64
	Lflag  uint64
	Cc     [20]byte
	Ispeed uint64
	Ospeed uint64
}
