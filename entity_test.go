package telegramify

import (
	"testing"
)

// TestUTF16Len_Empty 测试空字符串
func TestUTF16Len_Empty(t *testing.T) {
	if got := UTF16Len(""); got != 0 {
		t.Errorf("UTF16Len(\"\") = %d, want 0", got)
	}
}

// TestUTF16Len_ASCII 测试 ASCII 字符
func TestUTF16Len_ASCII(t *testing.T) {
	if got := UTF16Len("hello"); got != 5 {
		t.Errorf("UTF16Len(\"hello\") = %d, want 5", got)
	}
}

// TestUTF16Len_CJK 测试中日韩字符（BMP 内，每个 1 个 UTF-16 code unit）
func TestUTF16Len_CJK(t *testing.T) {
	if got := UTF16Len("你好"); got != 2 {
		t.Errorf("UTF16Len(\"你好\") = %d, want 2", got)
	}
}

// TestUTF16Len_EmojiBMP 测试 BMP 内的 emoji
func TestUTF16Len_EmojiBMP(t *testing.T) {
	// ☑️ is U+2611 (BMP) + U+FE0F (BMP) = 2 code units
	if got := UTF16Len("☑️"); got != 2 {
		t.Errorf("UTF16Len(\"☑️\") = %d, want 2", got)
	}
}

// TestUTF16Len_EmojiSupplementary 测试补充平面的 emoji
func TestUTF16Len_EmojiSupplementary(t *testing.T) {
	// 📌 is U+1F4CC (supplementary plane) = 2 UTF-16 code units
	if got := UTF16Len("📌"); got != 2 {
		t.Errorf("UTF16Len(\"📌\") = %d, want 2", got)
	}
}

// TestUTF16Len_Mixed 测试混合字符
func TestUTF16Len_Mixed(t *testing.T) {
	// "A📌B" = 1 + 2 + 1 = 4
	if got := UTF16Len("A📌B"); got != 4 {
		t.Errorf("UTF16Len(\"A📌B\") = %d, want 4", got)
	}
}

// TestUTF16Len_FlagEmoji 测试旗帜 emoji
func TestUTF16Len_FlagEmoji(t *testing.T) {
	// 🇺🇸 is two regional indicator symbols, each supplementary
	if got := UTF16Len("🇺🇸"); got != 4 {
		t.Errorf("UTF16Len(\"🇺🇸\") = %d, want 4", got)
	}
}

// TestUTF16Len_MatchesEncode 测试 UTF16Len 是否匹配 UTF-16LE 编码长度
func TestUTF16Len_MatchesEncode(t *testing.T) {
	testStrings := []string{
		"",
		"hello",
		"你好世界",
		"📌✅🔗",
		"A📌B你好C",
		"test 🇺🇸 flag",
	}
	for _, s := range testStrings {
		t.Run(s, func(t *testing.T) {
			expected := len([]rune(s)) // 简化版：实际应计算 UTF-16LE
			// 精确计算
			expected = 0
			for _, r := range s {
				if r > 0xFFFF {
					expected += 2
				} else {
					expected++
				}
			}
			got := UTF16Len(s)
			if got != expected {
				t.Errorf("UTF16Len(%q) = %d, want %d", s, got, expected)
			}
		})
	}
}

// TestMessageEntity_ToDict 测试 MessageEntity.ToDict
func TestMessageEntity_ToDict(t *testing.T) {
	e := MessageEntity{Type: "bold", Offset: 0, Length: 5}
	d := e.ToDict()
	if d["type"] != "bold" || d["offset"] != 0 || d["length"] != 5 {
		t.Errorf("ToDict() = %v, want type=bold offset=0 length=5", d)
	}
	if _, exists := d["url"]; exists {
		t.Error("ToDict() should not include empty url")
	}
}

func TestMessageEntity_ToDictWithURL(t *testing.T) {
	e := MessageEntity{Type: "text_link", Offset: 0, Length: 5, URL: "https://example.com"}
	d := e.ToDict()
	if d["url"] != "https://example.com" {
		t.Errorf("ToDict() url = %v, want https://example.com", d["url"])
	}
	if _, exists := d["language"]; exists {
		t.Error("ToDict() should not include language when not set")
	}
}

func TestMessageEntity_ToDictWithLanguage(t *testing.T) {
	e := MessageEntity{Type: "pre", Offset: 0, Length: 10, Language: "python"}
	d := e.ToDict()
	if d["language"] != "python" {
		t.Errorf("ToDict() language = %v, want python", d["language"])
	}
	if _, exists := d["url"]; exists {
		t.Error("ToDict() should not include url when not set")
	}
}

func TestMessageEntity_ToDictWithCustomEmoji(t *testing.T) {
	e := MessageEntity{Type: "custom_emoji", Offset: 0, Length: 2, CustomEmojiID: "5368324170671202286"}
	d := e.ToDict()
	if d["custom_emoji_id"] != "5368324170671202286" {
		t.Errorf("ToDict() custom_emoji_id = %v, want 5368324170671202286", d["custom_emoji_id"])
	}
}

// TestSplitEntities_NoSplitNeeded 测试不需要拆分的情况
func TestSplitEntities_NoSplitNeeded(t *testing.T) {
	text := "hello"
	entities := []MessageEntity{{Type: "bold", Offset: 0, Length: 5}}
	result := SplitEntities(text, entities, 100)
	if len(result) != 1 {
		t.Errorf("SplitEntities() returned %d chunks, want 1", len(result))
	}
	if result[0].Text != "hello" || len(result[0].Entities) != 1 {
		t.Errorf("SplitEntities() result = %v, want text=hello with 1 entity", result[0])
	}
}

// TestSplitEntities_EmptyText 测试空文本
func TestSplitEntities_EmptyText(t *testing.T) {
	result := SplitEntities("", []MessageEntity{}, 100)
	if len(result) != 1 {
		t.Errorf("SplitEntities() returned %d chunks, want 1", len(result))
	}
	if result[0].Text != "" || len(result[0].Entities) != 0 {
		t.Errorf("SplitEntities() result = %v, want empty", result[0])
	}
}

// TestSplitEntities_SplitAtNewline 测试在换行符处拆分
func TestSplitEntities_SplitAtNewline(t *testing.T) {
	text := "aaa\nbbb\nccc"
	entities := []MessageEntity{}
	result := SplitEntities(text, entities, 5)
	// "aaa\n" = 4 code units, "bbb\n" = 4, "ccc" = 3
	if len(result) < 2 {
		t.Errorf("SplitEntities() returned %d chunks, want >= 2", len(result))
	}
	// 合并所有文本应该等于原文本
	combined := ""
	for _, chunk := range result {
		combined += chunk.Text
	}
	if combined != text {
		t.Errorf("SplitEntities() combined text = %q, want %q", combined, text)
	}
}

// TestSplitEntities_EntityFullyInFirstChunk 测试 entity 完全在第一个块中
func TestSplitEntities_EntityFullyInFirstChunk(t *testing.T) {
	text := "bold\nnormal"
	entities := []MessageEntity{{Type: "bold", Offset: 0, Length: 4}}
	result := SplitEntities(text, entities, 5)
	if len(result) < 2 {
		t.Errorf("SplitEntities() returned %d chunks, want >= 2", len(result))
	}
	// 第一个块应该有 bold entity
	if len(result[0].Entities) != 1 || result[0].Entities[0].Type != "bold" {
		t.Errorf("First chunk should have bold entity, got %v", result[0].Entities)
	}
}

// TestSplitEntities_PreservesTotalText 测试拆分保留完整文本
func TestSplitEntities_PreservesTotalText(t *testing.T) {
	text := "line1\nline2\nline3\nline4\nline5"
	entities := []MessageEntity{{Type: "italic", Offset: 0, Length: 5}}
	result := SplitEntities(text, entities, 12)
	combined := ""
	for _, chunk := range result {
		combined += chunk.Text
	}
	if combined != text {
		t.Errorf("SplitEntities() combined = %q, want %q", combined, text)
	}
}

// TestSplitEntities_WithEmoji 测试包含 emoji 的拆分
func TestSplitEntities_WithEmoji(t *testing.T) {
	// 📌 = 2 UTF-16 code units
	text := "📌\n📌\n📌"
	entities := []MessageEntity{}
	result := SplitEntities(text, entities, 4)
	combined := ""
	for _, chunk := range result {
		combined += chunk.Text
	}
	if combined != text {
		t.Errorf("SplitEntities() combined = %q, want %q", combined, text)
	}
}

// TestSplitEntities_HardSplitNoNewlines 测试没有换行符的硬拆分
func TestSplitEntities_HardSplitNoNewlines(t *testing.T) {
	text := "abcdefghij"
	entities := []MessageEntity{}
	result := SplitEntities(text, entities, 4)
	combined := ""
	for _, chunk := range result {
		combined += chunk.Text
		if UTF16Len(chunk.Text) > 4 {
			t.Errorf("Chunk %q exceeds max length 4", chunk.Text)
		}
	}
	if combined != text {
		t.Errorf("SplitEntities() combined = %q, want %q", combined, text)
	}
}

