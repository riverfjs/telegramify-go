package telegramify

import (
	"context"
	"strings"
	"testing"

	"github.com/riverfjs/telegramify-go/internal/converter"
)

// TestStripNewlinesAdjustInternal_OnlyNewlines 测试只包含换行符的情况
// 这是导致 panic: slice bounds out of range [2:0] 的边界情况
func TestStripNewlinesAdjustInternal_OnlyNewlines(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		entities []MessageEntity
		wantText string
		wantLen  int
	}{
		{
			name:     "single newline",
			text:     "\n",
			entities: []MessageEntity{},
			wantText: "",
			wantLen:  0,
		},
		{
			name:     "two newlines",
			text:     "\n\n",
			entities: []MessageEntity{},
			wantText: "",
			wantLen:  0,
		},
		{
			name:     "multiple newlines",
			text:     "\n\n\n\n",
			entities: []MessageEntity{},
			wantText: "",
			wantLen:  0,
		},
		{
			name:     "leading newlines only",
			text:     "\n\ntest",
			entities: []MessageEntity{},
			wantText: "test",
			wantLen:  0,
		},
		{
			name:     "trailing newlines only",
			text:     "test\n\n",
			entities: []MessageEntity{},
			wantText: "test",
			wantLen:  0,
		},
		{
			name:     "both leading and trailing",
			text:     "\n\ntest\n\n",
			entities: []MessageEntity{},
			wantText: "test",
			wantLen:  0,
		},
		{
			name:     "empty string",
			text:     "",
			entities: []MessageEntity{},
			wantText: "",
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText, gotEntities := stripNewlinesAdjustInternal(tt.text, tt.entities)
			if gotText != tt.wantText {
				t.Errorf("stripNewlinesAdjustInternal() text = %q, want %q", gotText, tt.wantText)
			}
			if len(gotEntities) != tt.wantLen {
				t.Errorf("stripNewlinesAdjustInternal() entities len = %d, want %d", len(gotEntities), tt.wantLen)
			}
		})
	}
}

// TestStripNewlinesAdjustInternal_WithEntities 测试带实体的情况
func TestStripNewlinesAdjustInternal_WithEntities(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		entities     []MessageEntity
		wantText     string
		wantEntities []MessageEntity
	}{
		{
			name: "entity in middle",
			text: "\n\nhello world\n\n",
			entities: []MessageEntity{
				{Type: "bold", Offset: 2, Length: 5}, // "hello" after leading newlines
			},
			wantText: "hello world",
			wantEntities: []MessageEntity{
				{Type: "bold", Offset: 0, Length: 5}, // adjusted to start
			},
		},
		{
			name: "entity clipped by leading newlines",
			text: "\n\nhello\n\n",
			entities: []MessageEntity{
				{Type: "bold", Offset: 0, Length: 7}, // starts before text
			},
			wantText: "hello",
			wantEntities: []MessageEntity{
				{Type: "bold", Offset: 0, Length: 5}, // clipped
			},
		},
		{
			name: "entity entirely in stripped area",
			text: "\n\nhello\n\n",
			entities: []MessageEntity{
				{Type: "bold", Offset: 0, Length: 1}, // in leading newlines
			},
			wantText:     "hello",
			wantEntities: []MessageEntity{}, // entity removed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText, gotEntities := stripNewlinesAdjustInternal(tt.text, tt.entities)
			if gotText != tt.wantText {
				t.Errorf("stripNewlinesAdjustInternal() text = %q, want %q", gotText, tt.wantText)
			}
			if len(gotEntities) != len(tt.wantEntities) {
				t.Errorf("stripNewlinesAdjustInternal() entities len = %d, want %d", len(gotEntities), len(tt.wantEntities))
				return
			}
			for i := range gotEntities {
				if gotEntities[i].Type != tt.wantEntities[i].Type ||
					gotEntities[i].Offset != tt.wantEntities[i].Offset ||
					gotEntities[i].Length != tt.wantEntities[i].Length {
					t.Errorf("stripNewlinesAdjustInternal() entity[%d] = %+v, want %+v",
						i, gotEntities[i], tt.wantEntities[i])
				}
			}
		})
	}
}

// TestHandleCodeBlock_LineFiltering tests that code blocks are only extracted as files if > 50 lines
func TestHandleCodeBlock_LineFiltering(t *testing.T) {
	tests := []struct {
		name      string
		lines     int
		wantFile  bool
	}{
		{
			name:     "1 line - no file",
			lines:    1,
			wantFile: false,
		},
		{
			name:     "10 lines - no file",
			lines:    10,
			wantFile: false,
		},
		{
			name:     "50 lines - no file (boundary)",
			lines:    50,
			wantFile: false,
		},
		{
			name:     "51 lines - extract as file",
			lines:    51,
			wantFile: true,
		},
		{
			name:     "100 lines - extract as file",
			lines:    100,
			wantFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate code with N lines
			codeLines := make([]string, tt.lines)
			for i := 0; i < tt.lines; i++ {
				codeLines[i] = "line" + string(rune('0'+i%10))
			}
			rawCode := strings.Join(codeLines, "\n")

			seg := converter.Segment{
				Kind:     "code_block",
				Language: "go",
				RawCode:  rawCode,
			}

			result := []Content{}
			handleCodeBlock(&result, seg)

			if tt.wantFile && len(result) == 0 {
				t.Errorf("handleCodeBlock() expected file for %d lines, got none", tt.lines)
			}
			if !tt.wantFile && len(result) > 0 {
				t.Errorf("handleCodeBlock() expected no file for %d lines, got %d", tt.lines, len(result))
			}

			// Verify file content if extracted
			if tt.wantFile && len(result) > 0 {
				file, ok := result[0].(*File)
				if !ok {
					t.Errorf("handleCodeBlock() expected File type, got %T", result[0])
					return
				}
				if string(file.FileData) != rawCode {
					t.Errorf("handleCodeBlock() file data mismatch")
				}
			}
		})
	}
}

// TestPathWithTilde tests that file paths with ~ (home directory) are not treated as strikethrough
func TestPathWithTilde(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantText string
	}{
		{
			name:     "single tilde path",
			markdown: "路径是 ~/.myclaw/workspace/file",
			wantText: "路径是 ~/.myclaw/workspace/file",
		},
		{
			name:     "command with tilde path",
			markdown: "执行 ~/.myclaw/workspace/.claude/skills/todo/bin/todo reminders",
			wantText: "执行 ~/.myclaw/workspace/.claude/skills/todo/bin/todo reminders",
		},
		{
			name:     "two separate tilde paths",
			markdown: "路径1: ~/.myclaw/path1 路径2: ~/.myclaw/path2",
			wantText: "路径1: ~/.myclaw/path1 路径2: ~/.myclaw/path2",
		},
		{
			name:     "tilde in middle of line",
			markdown: "你可以使用 ~/.myclaw/workspace/.claude/skills/todo/bin/todo complete 4 来完成任务",
			wantText: "你可以使用 ~/.myclaw/workspace/.claude/skills/todo/bin/todo complete 4 来完成任务",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents, err := ProcessMarkdown(context.Background(), tt.markdown, 4096, false, nil)
			if err != nil {
				t.Fatalf("ProcessMarkdown() error = %v", err)
			}
			
			if len(contents) == 0 {
				t.Fatal("ProcessMarkdown() returned no contents")
			}
			
			text, ok := contents[0].(*Text)
			if !ok {
				t.Fatalf("ProcessMarkdown() first content is not Text, got %T", contents[0])
			}
			
			if text.Text != tt.wantText {
				t.Errorf("ProcessMarkdown() text mismatch\ngot:  %q\nwant: %q", text.Text, tt.wantText)
			}
		})
	}
}
