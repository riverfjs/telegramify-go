package telegramify

import (
	"testing"
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

