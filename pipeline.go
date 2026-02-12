package telegramify

import (
	"bytes"
	"context"
	"strings"

	"github.com/riverfjs/telegramify-go/internal/converter"
	"github.com/riverfjs/telegramify-go/internal/mermaid"
	"github.com/riverfjs/telegramify-go/internal/util"
)

// stripNewlinesAdjust 去除首尾换行符并调整 entity 偏移量
func stripNewlinesAdjustInternal(text string, entities []MessageEntity) (string, []MessageEntity) {
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
	
	end := len(text) - trailing
	if trailing > 0 {
		end = len(text) - trailing
	} else {
		end = len(text)
	}
	
	// Check bounds before slicing to avoid panic
	if leading >= end {
		return "", []MessageEntity{}
	}
	
	stripped := text[leading:end]
	if stripped == "" {
		return stripped, []MessageEntity{}
	}
	
	// Newlines are each 1 UTF-16 code unit
	leadingUTF16 := leading
	newUTF16Len := UTF16Len(stripped)
	
	adjusted := make([]MessageEntity, 0)
	for _, ent := range entities {
		newOffset := ent.Offset - leadingUTF16
		newEnd := newOffset + ent.Length
		
		// Skip entities entirely outside the stripped range
		if newEnd <= 0 || newOffset >= newUTF16Len {
			continue
		}
		
		// Clip to boundaries
		if newOffset < 0 {
			newOffset = 0
		}
		if newEnd > newUTF16Len {
			newEnd = newUTF16Len
		}
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

// sliceTextEntities 提取子串及其重叠的实体，调整偏移量
func sliceTextEntities(
	fullText string,
	fullEntities []MessageEntity,
	pyStart int,
	pyEnd int,
	utf16Start int,
	utf16End int,
) (string, []MessageEntity) {
	chunkText := fullText[pyStart:pyEnd]
	chunkEntities := make([]MessageEntity, 0)
	
	for _, ent := range fullEntities {
		entStart := ent.Offset
		entEnd := ent.Offset + ent.Length
		
		// Check overlap with [utf16Start, utf16End)
		if entEnd <= utf16Start || entStart >= utf16End {
			continue
		}
		
		clippedStart := entStart
		if clippedStart < utf16Start {
			clippedStart = utf16Start
		}
		
		clippedEnd := entEnd
		if clippedEnd > utf16End {
			clippedEnd = utf16End
		}
		
		clippedLength := clippedEnd - clippedStart
		if clippedLength <= 0 {
			continue
		}
		
		chunkEntities = append(chunkEntities, MessageEntity{
			Type:          ent.Type,
			Offset:        clippedStart - utf16Start,
			Length:        clippedLength,
			URL:           ent.URL,
			Language:      ent.Language,
			CustomEmojiID: ent.CustomEmojiID,
		})
	}
	
	return chunkText, chunkEntities
}

// ProcessMarkdown 完整异步管道：markdown → 可发送的内容列表
//
// 步骤：
// 1. 通过 converter 转换 markdown 为 (text, entities, segments)
// 2. 按顺序遍历 segments：
//    - mermaid → 渲染为 Photo（或失败时为 File）
//    - code_block → 提取为 File
//    - text regions → 收集并按 max_message_length 拆分
// 3. 返回 Text | File | Photo 的有序列表
func ProcessMarkdown(
	ctx context.Context,
	content string,
	maxMessageLength int,
	latexEscape bool,
	config *RenderConfig,
) ([]Content, error) {
	if maxMessageLength <= 0 {
		maxMessageLength = 4096
	}
	if config == nil {
		config = DefaultConfig()
	}
	
	fullText, fullEntities, segments := ConvertWithSegments(content, latexEscape, config)
	
	result := make([]Content, 0)
	
	// First pass: identify which code blocks should be extracted as files
	// Only segments that are extracted as files/photos will split the text
	extractableSegments := make([]converter.Segment, 0)
	for _, s := range segments {
		if s.Kind == "mermaid" {
			// Mermaid always extracted as photo/file
			extractableSegments = append(extractableSegments, s)
		} else if s.Kind == "code_block" {
			// Only extract code blocks > 50 lines
			lineCount := strings.Count(s.RawCode, "\n") + 1
			if lineCount > 50 {
				extractableSegments = append(extractableSegments, s)
			}
		}
	}
	
	// Walk through the text, splitting only at extractable segments
	cursorPy := 0
	cursorUTF16 := 0
	
	for _, seg := range extractableSegments {
		// Emit text before this segment
		if seg.TextStart > cursorPy {
			textChunk, textEntities := sliceTextEntities(
				fullText, fullEntities,
				cursorPy, seg.TextStart,
				cursorUTF16, seg.UTF16Start,
			)
			textChunk, textEntities = stripNewlinesAdjustInternal(textChunk, textEntities)
			if textChunk != "" {
				appendTextChunks(&result, textChunk, textEntities, maxMessageLength)
			}
		}
		
		// Extract the segment as file/photo
		if seg.Kind == "mermaid" {
			handleMermaid(ctx, &result, seg)
		} else if seg.Kind == "code_block" {
			handleCodeBlockAsFile(&result, seg)
		}
		
		// Move cursor past the segment
		cursorPy = seg.TextEnd
		cursorUTF16 = seg.UTF16End
	}
	
	// Emit remaining text after last special segment
	if cursorPy < len(fullText) {
		textChunk, textEntities := sliceTextEntities(
			fullText, fullEntities,
			cursorPy, len(fullText),
			cursorUTF16, UTF16Len(fullText),
		)
		textChunk, textEntities = stripNewlinesAdjust(textChunk, textEntities)
		if textChunk != "" {
			appendTextChunks(&result, textChunk, textEntities, maxMessageLength)
		}
	}
	
	// If no output was generated, emit empty text
	if len(result) == 0 && strings.TrimSpace(fullText) != "" {
		appendTextChunks(&result, strings.TrimSpace(fullText), fullEntities, maxMessageLength)
	}
	
	return result, nil
}

// appendTextChunks 按 max_message_length 拆分文本并发送 Text 对象
func appendTextChunks(
	result *[]Content,
	text string,
	entities []MessageEntity,
	maxMessageLength int,
) {
	chunks := SplitEntities(text, entities, maxMessageLength)
	for _, chunk := range chunks {
		chunkText, chunkEntities := stripNewlinesAdjust(chunk.Text, chunk.Entities)
		if chunkText != "" {
			*result = append(*result, &Text{
				Text:     chunkText,
				Entities: chunkEntities,
				ContentTrace: ContentTrace{
					SourceType: "text",
				},
			})
		}
	}
}

// handleCodeBlockAsFile 将大代码块提取为 File（仅当代码超过 50 行时调用）
func handleCodeBlockAsFile(result *[]Content, seg converter.Segment) {
	rawCode := seg.RawCode
	lang := seg.Language
	if lang == "" {
		lang = "txt"
	}
	fileName := util.GetFilename(rawCode, lang)
	
	*result = append(*result, &File{
		FileName: fileName,
		FileData: []byte(rawCode),
		ContentTrace: ContentTrace{
			SourceType: "file",
			Extra: map[string]interface{}{
				"language": lang,
			},
		},
	})
}

// handleMermaid 渲染 mermaid 图表为 Photo，或回退到 File
func handleMermaid(ctx context.Context, result *[]Content, seg converter.Segment) {
	rawCode := seg.RawCode
	
	// 尝试渲染 Mermaid
	imgData, caption, err := renderMermaid(ctx, rawCode)
	if err != nil {
		// 渲染失败，作为文件发送
		Logger.Printf("Mermaid rendering failed: %v", err)
		*result = append(*result, &File{
			FileName: "invalid_mermaid.txt",
			FileData: []byte(rawCode),
			ContentTrace: ContentTrace{
				SourceType: ContentTypeMermaid,
			},
		})
		return
	}
	
	// 渲染成功，作为图片发送
	*result = append(*result, &Photo{
		FileName: "mermaid.webp",
		FileData: imgData.Bytes(),
		Caption:  caption,
		ContentTrace: ContentTrace{
			SourceType: ContentTypeMermaid,
		},
	})
}

// renderMermaid 内部渲染函数
func renderMermaid(ctx context.Context, code string) (*bytes.Buffer, string, error) {
	return mermaid.RenderMermaid(ctx, code, nil)
}

