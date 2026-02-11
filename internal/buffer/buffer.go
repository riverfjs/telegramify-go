package buffer

// UTF16Len returns the length of text measured in UTF-16 code units.
func utf16Len(text string) int {
	count := 0
	for _, r := range text {
		if r > 0xFFFF {
			count += 2
		} else {
			count++
		}
	}
	return count
}

// TextBuffer accumulates plain text and tracks the current UTF-16 offset.
type TextBuffer struct {
	parts       []string
	utf16Offset int
}

// New creates a new TextBuffer.
func New() *TextBuffer {
	return &TextBuffer{
		parts:       make([]string, 0),
		utf16Offset: 0,
	}
}

// Write appends text to the buffer.
func (tb *TextBuffer) Write(text string) {
	tb.parts = append(tb.parts, text)
	tb.utf16Offset += utf16Len(text)
}

// UTF16Offset returns the current UTF-16 offset.
func (tb *TextBuffer) UTF16Offset() int {
	return tb.utf16Offset
}

// ByteOffset returns the current byte offset (total string length).
func (tb *TextBuffer) ByteOffset() int {
	total := 0
	for _, p := range tb.parts {
		total += len(p)
	}
	return total
}

// TrailingNewlineCount counts trailing newline characters in the buffer.
func (tb *TextBuffer) TrailingNewlineCount() int {
	count := 0
	for i := len(tb.parts) - 1; i >= 0; i-- {
		part := tb.parts[i]
		for j := len(part) - 1; j >= 0; j-- {
			if part[j] == '\n' {
				count++
			} else {
				return count
			}
		}
	}
	return count
}

// PopLast removes and returns the last written part.
// Used for replacing just-written bullet prefixes in task lists.
func (tb *TextBuffer) PopLast() string {
	if len(tb.parts) == 0 {
		return ""
	}
	last := tb.parts[len(tb.parts)-1]
	tb.parts = tb.parts[:len(tb.parts)-1]
	tb.utf16Offset -= utf16Len(last)
	return last
}

// String returns the accumulated text.
func (tb *TextBuffer) String() string {
	if len(tb.parts) == 0 {
		return ""
	}
	// Calculate total length
	totalLen := 0
	for _, p := range tb.parts {
		totalLen += len(p)
	}
	// Build result efficiently
	result := make([]byte, 0, totalLen)
	for _, p := range tb.parts {
		result = append(result, []byte(p)...)
	}
	return string(result)
}

// Reset clears the buffer.
func (tb *TextBuffer) Reset() {
	tb.parts = tb.parts[:0]
	tb.utf16Offset = 0
}

