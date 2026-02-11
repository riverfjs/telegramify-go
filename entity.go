package telegramify

import (
	"strings"
	"github.com/riverfjs/telegramify-go/internal/types"
)

// 导出类型别名
type MessageEntity = types.MessageEntity

// UTF16Len returns the length of text measured in UTF-16 code units.
//
// Telegram measures entity offsets and lengths in UTF-16 code units,
// not Go string bytes or runes. Characters outside the BMP (codepoint > 0xFFFF)
// take 2 UTF-16 code units (a surrogate pair); all others take 1.
func UTF16Len(text string) int {
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

// TextChunk represents a chunk of text with its entities.
type TextChunk struct {
	Text     string
	Entities []MessageEntity
}

// findNewlinePositions finds newline positions in text suitable for splitting.
// Returns a list of string indices right after each newline.
func findNewlinePositions(text string) []int {
	var points []int
	for i, ch := range text {
		if ch == '\n' {
			// Convert rune index to byte position
			bytePos := 0
			for j, r := range text {
				if j == i {
					break
				}
				bytePos += len(string(r))
			}
			points = append(points, bytePos+1)
		}
	}
	return points
}

// buildUTF16OffsetTable builds a cumulative UTF-16 offset table for each byte position.
// Returns a slice where result[i] is the UTF-16 offset at byte position i.
func buildUTF16OffsetTable(text string) []int {
	offsets := make([]int, len(text)+1)
	cum := 0
	bytePos := 0
	for _, r := range text {
		offsets[bytePos] = cum
		if r > 0xFFFF {
			cum += 2
		} else {
			cum++
		}
		bytePos += len(string(r))
	}
	offsets[len(text)] = cum
	return offsets
}

// SplitEntities splits (text, entities) into chunks not exceeding maxUTF16Len UTF-16 code units.
//
// Tries to split at newline boundaries. Entities that span a split boundary
// are clipped into both chunks.
func SplitEntities(text string, entities []MessageEntity, maxUTF16Len int) []TextChunk {
	total := UTF16Len(text)
	if total <= maxUTF16Len {
		return []TextChunk{{Text: text, Entities: entities}}
	}

	offsets := buildUTF16OffsetTable(text)

	// Build list of candidate split points (newline positions)
	splitPoints := findNewlinePositions(text)

	// Determine actual split positions using greedy packing
	var chunksRanges [][2]int // [byteStart, byteEnd]
	byteStart := 0

	for byteStart < len(text) {
		utf16Start := offsets[byteStart]
		utf16Budget := utf16Start + maxUTF16Len

		if offsets[len(text)] <= utf16Budget {
			// Remaining text fits
			chunksRanges = append(chunksRanges, [2]int{byteStart, len(text)})
			break
		}

		// Find the last split point that fits within budget
		bestSplit := -1
		for _, sp := range splitPoints {
			if sp <= byteStart {
				continue
			}
			if sp < len(offsets) && offsets[sp] <= utf16Budget {
				bestSplit = sp
			} else {
				break
			}
		}

		if bestSplit == -1 || bestSplit == byteStart {
			// No newline split fits -- hard split at maxUTF16Len boundary
			bestSplit = byteStart
			for i := byteStart + 1; i <= len(text); i++ {
				if i < len(offsets) && offsets[i] > utf16Budget {
					bestSplit = i - 1
					break
				}
			}
			if bestSplit == byteStart {
				bestSplit = byteStart + 1 // Force progress
			}
		}

		chunksRanges = append(chunksRanges, [2]int{byteStart, bestSplit})
		byteStart = bestSplit
	}

	// Assign entities to chunks, clipping as needed
	var result []TextChunk
	for _, chunkRange := range chunksRanges {
		chunkByteStart, chunkByteEnd := chunkRange[0], chunkRange[1]
		chunkText := text[chunkByteStart:chunkByteEnd]
		chunkUTF16Start := offsets[chunkByteStart]
		chunkUTF16End := offsets[chunkByteEnd]
		var chunkEntities []MessageEntity

		for _, ent := range entities {
			entStart := ent.Offset
			entEnd := ent.Offset + ent.Length

			// Check overlap
			if entEnd <= chunkUTF16Start || entStart >= chunkUTF16End {
				continue // No overlap
			}

			// Clip to chunk boundaries
			clippedStart := max(entStart, chunkUTF16Start)
			clippedEnd := min(entEnd, chunkUTF16End)
			clippedLength := clippedEnd - clippedStart

			if clippedLength <= 0 {
				continue
			}

			newEnt := MessageEntity{
				Type:          ent.Type,
				Offset:        clippedStart - chunkUTF16Start,
				Length:        clippedLength,
				URL:           ent.URL,
				Language:      ent.Language,
				CustomEmojiID: ent.CustomEmojiID,
			}
			chunkEntities = append(chunkEntities, newEnt)
		}

		result = append(result, TextChunk{
			Text:     chunkText,
			Entities: chunkEntities,
		})
	}

	return result
}

// stripNewlinesAdjust strips leading/trailing newlines from text and adjusts entity offsets.
func stripNewlinesAdjust(text string, entities []MessageEntity) (string, []MessageEntity) {
	// Count leading newlines
	leading := 0
	for _, ch := range text {
		if ch == '\n' {
			leading++
		} else {
			break
		}
	}

	// Count trailing newlines
	trailing := 0
	runes := []rune(text)
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] == '\n' {
			trailing++
		} else {
			break
		}
	}

	if leading == 0 && trailing == 0 {
		return text, entities
	}

	end := len(text)
	if trailing > 0 {
		end = len(text) - trailing
	}
	stripped := text[leading:end]
	if stripped == "" {
		return stripped, nil
	}

	// Newlines are each 1 UTF-16 code unit
	leadingUTF16 := leading
	newUTF16Len := UTF16Len(stripped)

	var adjusted []MessageEntity
	for _, ent := range entities {
		newOffset := ent.Offset - leadingUTF16
		newEnd := newOffset + ent.Length
		// Skip entities entirely outside the stripped range
		if newEnd <= 0 || newOffset >= newUTF16Len {
			continue
		}
		// Clip to boundaries
		newOffset = max(0, newOffset)
		newEnd = min(newEnd, newUTF16Len)
		newLength := newEnd - newOffset
		if newLength <= 0 {
			continue
		}
		adjusted = append(adjusted, MessageEntity{
			Type:          ent.Type,
			Offset:        newOffset,
			Length:        newLength,
			URL:           ent.URL,
			Language:      ent.Language,
			CustomEmojiID: ent.CustomEmojiID,
		})
	}

	return stripped, adjusted
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TrimSpace removes leading and trailing whitespace while adjusting entities.
func TrimSpace(text string, entities []MessageEntity) (string, []MessageEntity) {
	trimmed := strings.TrimSpace(text)
	if trimmed == text {
		return text, entities
	}

	// Find the start offset
	startOffset := 0
	for i, ch := range text {
		if !isSpace(ch) {
			startOffset = i
			break
		}
	}

	utf16Start := UTF16Len(text[:startOffset])
	utf16Len := UTF16Len(trimmed)

	var adjusted []MessageEntity
	for _, ent := range entities {
		newOffset := ent.Offset - utf16Start
		newEnd := newOffset + ent.Length

		if newEnd <= 0 || newOffset >= utf16Len {
			continue
		}

		newOffset = max(0, newOffset)
		newEnd = min(newEnd, utf16Len)
		newLength := newEnd - newOffset

		if newLength <= 0 {
			continue
		}

		adjusted = append(adjusted, MessageEntity{
			Type:          ent.Type,
			Offset:        newOffset,
			Length:        newLength,
			URL:           ent.URL,
			Language:      ent.Language,
			CustomEmojiID: ent.CustomEmojiID,
		})
	}

	return trimmed, adjusted
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

