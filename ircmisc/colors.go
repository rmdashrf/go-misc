package ircmisc

import (
	"strconv"
)

type FormattedToken struct {
	Text   string
	Format Formatting
}

// BIts [0, 6)  are formatting flags
// Bits [6, 10) are the foreground color
// Bits [10, 14) are the background color
type Formatting uint32

func (f Formatting) IsBold() bool {
	return f&FLAG_BOLD != 0
}

func (f Formatting) IsItalic() bool {
	return f&FLAG_ITALICS != 0
}

func (f Formatting) IsUnderlined() bool {
	return f&FLAG_UNDERLINE != 0
}

func (f Formatting) IsStrikeThrough() bool {
	return f&FLAG_STRIKETHROUGH != 0
}

func (f Formatting) IsMonospace() bool {
	return f&FLAG_MONOSPACE != 0
}

func (f Formatting) ForegroundColor() int {
	return int(uint32((f & (15 << 6)) >> 6))
}

func (f Formatting) BackgroundColor() int {
	return int(uint32((f & (15 << 10)) >> 10))
}

func (f *Formatting) SetForegroundColor(c int) {
	if c < 0 {
		c = 0
	}
	if c > 15 {
		c = 15
	}

	mask := uint32(15) << 6
	*f &= Formatting(^mask)
	*f = Formatting(uint32(*f) | (uint32(c) << 6))
}

func (f *Formatting) SetBackgroundColor(c int) {
	if c < 0 {
		c = 0
	}
	if c > 15 {
		c = 15
	}

	mask := uint32(15) << 10
	*f &= Formatting(^mask)
	*f = Formatting(uint32(*f) | (uint32(c) << 10))
}

type FormattingFlags uint32

const (
	FLAG_COLOR         FormattingFlags = (1 << 0)
	FLAG_BOLD                          = (1 << 1)
	FLAG_ITALICS                       = (1 << 2)
	FLAG_UNDERLINE                     = (1 << 3)
	FLAG_STRIKETHROUGH                 = (1 << 4)
	FLAG_MONOSPACE                     = (1 << 5)
)

const (
	FORMAT_BOLD          = 0x02
	FORMAT_ITALICS       = 0x1D
	FORMAT_UNDERLINE     = 0x1F
	FORMAT_STRIKETHROUGH = 0x1E
	FORMAT_MONOSPACE     = 0x11
	FORMAT_COLOR         = 0x03
	FORMAT_RESET         = 0x0F
)

func TokenizeByFormatting(s string) (ret []FormattedToken) {
	var (
		remaining string = s
		consumed  string
		format    Formatting
	)

	for remaining != "" {
		remaining = consumeFormatting(remaining, &format)
		consumed, remaining = consumeText(remaining)

		if consumed != "" {
			ret = append(ret, FormattedToken{
				Text:   consumed,
				Format: format,
			})
		}
	}

	return
}

func isFormattingChar(c byte) bool {
	return c == FORMAT_BOLD || c == FORMAT_ITALICS || c == FORMAT_UNDERLINE || c == FORMAT_STRIKETHROUGH || c == FORMAT_MONOSPACE || c == FORMAT_COLOR || c == FORMAT_RESET
}

func consumeFormatting(s string, format *Formatting) (remaining string) {
	if len(s) < 1 {
		return
	}

	var i int

LOOP:
	for i = 0; i < len(s); {
		c := s[i]
		switch c {
		case FORMAT_BOLD:
			*format ^= FLAG_BOLD
			i++
		case FORMAT_ITALICS:
			*format ^= FLAG_ITALICS
			i++
		case FORMAT_UNDERLINE:
			*format ^= FLAG_UNDERLINE
			i++
		case FORMAT_STRIKETHROUGH:
			*format ^= FLAG_STRIKETHROUGH
			i++
		case FORMAT_MONOSPACE:
			*format ^= FLAG_MONOSPACE
			i++
		case FORMAT_RESET:
			*format = Formatting(0)
			i++
		case FORMAT_COLOR:
			// If we can't peek at the next character, or if the next character
			// isn't numeric, then we reset the color formatting
			if i+1 >= len(s) || !isNumeric(s[i+1]) {
				format.SetBackgroundColor(0)
				format.SetForegroundColor(0)
				i++
				break
			}

			i++
			// Read a number
			consumed := readColors(s[i:], format)
			i += consumed

		default:
			break LOOP

		}
	}

	if i < len(s) {
		remaining = s[i:]
	}

	return
}

func isNumeric(c byte) bool {
	return c >= '0' && c <= '9'
}

// Reads an IRC color spec e.g. 03,05. Returns the number of consumed characters
// We consume even invalid colors, like 69
func readColors(s string, f *Formatting) int {
	// IRC color codes are incredibly stupid. Right now, when this function is
	// called, we know that s[0] is a numeric.

	// IRC clients should not send 16-99, because that is ambiguous. ircdocs.horse
	// says that clients (such as us) may choose to process these or ignore them.
	// For simplicity, we will just process these

	consumed := 0
	color, c := readColor(s)
	consumed += c
	f.SetForegroundColor(color)

	s = s[consumed:]

	// Check if there is a background code, there can only be one
	if len(s) >= 2 && s[0] == ',' && isNumeric(s[1]) {
		// Consume the comma
		consumed += 1

		color, c := readColor(s[1:])
		consumed += c
		f.SetBackgroundColor(color)
	}

	return consumed
}

func readColor(s string) (color, consumed int) {
	if len(s) < 1 || !isNumeric(s[0]) {
		return
	}

	// We know that s[0] is a numeric char. Check if s[1] is also a numeric char
	if len(s) >= 2 && isNumeric(s[1]) {
		// Two digit color
		consumed = 2
		color, _ = strconv.Atoi(s[:2])
	} else {
		// One digit color
		consumed = 1
		color, _ = strconv.Atoi(s[:1])
	}

	return
}

func consumeText(s string) (consumed, remaining string) {
	var (
		i int
	)
	for i = 0; i < len(s); i++ {
		if isFormattingChar(s[i]) {
			break
		}
	}

	consumed = s[:i]

	if i < len(s) {
		remaining = s[i:]
	}

	return
}
