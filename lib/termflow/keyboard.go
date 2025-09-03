package termflow

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// Key represents a keyboard key
type Key struct {
	Type KeyType
	Rune rune
}

// KeyType represents the type of key pressed
type KeyType int

const (
	KeyUnknown KeyType = iota
	KeyRune
	KeyEnter
	KeyTab
	KeyBackspace
	KeyDelete
	KeyArrowUp
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	KeyCtrlC
	KeyCtrlD
	KeyEscape
)

// String returns a string representation of the key
func (k Key) String() string {
	switch k.Type {
	case KeyRune:
		return string(k.Rune)
	case KeyEnter:
		return "Enter"
	case KeyTab:
		return "Tab"
	case KeyBackspace:
		return "Backspace"
	case KeyDelete:
		return "Delete"
	case KeyArrowUp:
		return "ArrowUp"
	case KeyArrowDown:
		return "ArrowDown"
	case KeyArrowLeft:
		return "ArrowLeft"
	case KeyArrowRight:
		return "ArrowRight"
	case KeyCtrlC:
		return "Ctrl+C"
	case KeyCtrlD:
		return "Ctrl+D"
	case KeyEscape:
		return "Escape"
	default:
		return "Unknown"
	}
}

// KeyboardReader handles raw keyboard input
type KeyboardReader struct {
	fd       int
	oldState *term.State
	rawMode  bool
}

// NewKeyboardReader creates a new keyboard reader
func NewKeyboardReader() (*KeyboardReader, error) {
	fd := int(os.Stdin.Fd())
	// Skip terminal check in test mode to allow PTY testing
	isTestMode := os.Getenv("RIGEL_TEST_MODE") == "1"
	if !isTestMode && !term.IsTerminal(fd) {
		return nil, fmt.Errorf("stdin is not a terminal")
	}

	return &KeyboardReader{
		fd:      fd,
		rawMode: false,
	}, nil
}

// EnableRawMode enables raw terminal mode for key-by-key input
func (kr *KeyboardReader) EnableRawMode() error {
	if kr.rawMode {
		return nil
	}

	oldState, err := term.MakeRaw(kr.fd)
	if err != nil {
		return fmt.Errorf("failed to enable raw mode: %w", err)
	}

	kr.oldState = oldState
	kr.rawMode = true
	return nil
}

// DisableRawMode restores normal terminal mode
func (kr *KeyboardReader) DisableRawMode() error {
	if !kr.rawMode {
		return nil
	}

	err := term.Restore(kr.fd, kr.oldState)
	kr.rawMode = false
	return err
}

// ReadKey reads a single key press
func (kr *KeyboardReader) ReadKey() (Key, error) {
	if !kr.rawMode {
		return Key{}, fmt.Errorf("raw mode not enabled")
	}

	buf := make([]byte, 1)
	_, err := os.Stdin.Read(buf)
	if err != nil {
		return Key{}, err
	}

	b := buf[0]

	// Handle special keys
	switch b {
	case 3: // Ctrl+C
		return Key{Type: KeyCtrlC}, nil
	case 4: // Ctrl+D
		return Key{Type: KeyCtrlD}, nil
	case 9: // Tab
		return Key{Type: KeyTab}, nil
	case 13, 10: // Enter (CR or LF)
		return Key{Type: KeyEnter}, nil
	case 127, 8: // Backspace (DEL or BS)
		return Key{Type: KeyBackspace}, nil
	case 27: // Escape sequence
		return kr.readEscapeSequence()
	default:
		// Regular character
		if b >= 32 && b < 127 { // Printable ASCII
			return Key{Type: KeyRune, Rune: rune(b)}, nil
		}
		return Key{Type: KeyUnknown}, nil
	}
}

// readEscapeSequence reads and parses escape sequences (like arrow keys)
func (kr *KeyboardReader) readEscapeSequence() (Key, error) {
	// Read next byte to see if it's part of a sequence
	buf := make([]byte, 1)
	_, err := os.Stdin.Read(buf)
	if err != nil {
		return Key{Type: KeyEscape}, nil // Just escape key
	}

	if buf[0] != '[' {
		// Not an ANSI escape sequence, just escape
		return Key{Type: KeyEscape}, nil
	}

	// Read the final byte of the sequence
	_, err = os.Stdin.Read(buf)
	if err != nil {
		return Key{Type: KeyEscape}, nil
	}

	// Parse arrow keys and other sequences
	switch buf[0] {
	case 'A':
		return Key{Type: KeyArrowUp}, nil
	case 'B':
		return Key{Type: KeyArrowDown}, nil
	case 'C':
		return Key{Type: KeyArrowRight}, nil
	case 'D':
		return Key{Type: KeyArrowLeft}, nil
	case '3':
		// Delete key sends ESC[3~, read the ~
		buf2 := make([]byte, 1)
		os.Stdin.Read(buf2)
		if buf2[0] == '~' {
			return Key{Type: KeyDelete}, nil
		}
		return Key{Type: KeyUnknown}, nil
	default:
		return Key{Type: KeyUnknown}, nil
	}
}
