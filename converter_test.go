package telegramify

import (
	"strings"
	"testing"
)

// findEntity 查找指定类型的第一个 entity
func findEntity(entities []MessageEntity, etype string) *MessageEntity {
	for i := range entities {
		if entities[i].Type == etype {
			return &entities[i]
		}
	}
	return nil
}

// findEntities 查找指定类型的所有 entities
func findEntities(entities []MessageEntity, etype string) []MessageEntity {
	result := []MessageEntity{}
	for _, e := range entities {
		if e.Type == etype {
			result = append(result, e)
		}
	}
	return result
}

// extractEntityText 从纯文本中提取 entity 覆盖的子串
func extractEntityText(text string, entity *MessageEntity) string {
	utf16Offset := 0
	pyStart := -1
	pyEnd := -1
	for i, ch := range text {
		if utf16Offset == entity.Offset && pyStart == -1 {
			pyStart = i
		}
		if utf16Offset == entity.Offset+entity.Length && pyEnd == -1 {
			pyEnd = i
			break
		}
		if ch > 0xFFFF {
			utf16Offset += 2
		} else {
			utf16Offset++
		}
	}
	if pyStart != -1 && pyEnd == -1 {
		pyEnd = len(text)
	}
	if pyStart == -1 {
		return ""
	}
	return text[pyStart:pyEnd]
}

// TestBold_Simple 测试简单的粗体
func TestBold_Simple(t *testing.T) {
	text, entities := Convert("**hello**", false, nil)
	if !strings.Contains(text, "hello") {
		t.Errorf("Convert() text should contain 'hello'")
	}
	bold := findEntity(entities, "bold")
	if bold == nil {
		t.Fatal("Convert() should have bold entity")
	}
	if extractEntityText(text, bold) != "hello" {
		t.Errorf("Bold entity text = %q, want 'hello'", extractEntityText(text, bold))
	}
}

// TestBold_InSentence 测试句子中的粗体
func TestBold_InSentence(t *testing.T) {
	text, entities := Convert("foo **bar** baz", false, nil)
	bold := findEntity(entities, "bold")
	if bold == nil {
		t.Fatal("Convert() should have bold entity")
	}
	if extractEntityText(text, bold) != "bar" {
		t.Errorf("Bold entity text = %q, want 'bar'", extractEntityText(text, bold))
	}
}

// TestItalic_Simple 测试简单的斜体
func TestItalic_Simple(t *testing.T) {
	text, entities := Convert("*hello*", false, nil)
	italic := findEntity(entities, "italic")
	if italic == nil {
		t.Fatal("Convert() should have italic entity")
	}
	if extractEntityText(text, italic) != "hello" {
		t.Errorf("Italic entity text = %q, want 'hello'", extractEntityText(text, italic))
	}
}

// TestStrikethrough_Simple 测试简单的删除线
func TestStrikethrough_Simple(t *testing.T) {
	text, entities := Convert("~~hello~~", false, nil)
	s := findEntity(entities, "strikethrough")
	if s == nil {
		t.Fatal("Convert() should have strikethrough entity")
	}
	if extractEntityText(text, s) != "hello" {
		t.Errorf("Strikethrough entity text = %q, want 'hello'", extractEntityText(text, s))
	}
}

// TestNestedFormatting_BoldItalic 测试嵌套格式
func TestNestedFormatting_BoldItalic(t *testing.T) {
	text, entities := Convert("**bold *italic* bold**", false, nil)
	bold := findEntity(entities, "bold")
	italic := findEntity(entities, "italic")
	if bold == nil {
		t.Fatal("Convert() should have bold entity")
	}
	if italic == nil {
		t.Fatal("Convert() should have italic entity")
	}
	// Italic 应该包含在 bold 内
	if italic.Offset < bold.Offset {
		t.Errorf("Italic offset %d < bold offset %d", italic.Offset, bold.Offset)
	}
	if italic.Offset+italic.Length > bold.Offset+bold.Length {
		t.Errorf("Italic end %d > bold end %d", italic.Offset+italic.Length, bold.Offset+bold.Length)
	}
	if extractEntityText(text, italic) != "italic" {
		t.Errorf("Italic entity text = %q, want 'italic'", extractEntityText(text, italic))
	}
}

// TestInlineCode 测试行内代码
func TestInlineCode(t *testing.T) {
	text, entities := Convert("use `print()` here", false, nil)
	code := findEntity(entities, "code")
	if code == nil {
		t.Fatal("Convert() should have code entity")
	}
	if extractEntityText(text, code) != "print()" {
		t.Errorf("Code entity text = %q, want 'print()'", extractEntityText(text, code))
	}
}

// TestCodeBlock_Fenced 测试围栏代码块
func TestCodeBlock_Fenced(t *testing.T) {
	md := "```python\nprint('hello')\n```"
	text, entities := Convert(md, false, nil)
	pre := findEntity(entities, "pre")
	if pre == nil {
		t.Fatal("Convert() should have pre entity")
	}
	if pre.Language != "python" {
		t.Errorf("Pre entity language = %q, want 'python'", pre.Language)
	}
	extracted := extractEntityText(text, pre)
	if !strings.Contains(extracted, "print('hello')") {
		t.Errorf("Pre entity should contain print('hello'), got %q", extracted)
	}
}

// TestCodeBlock_NoLanguage 测试无语言标识的代码块
func TestCodeBlock_NoLanguage(t *testing.T) {
	md := "```\nsome code\n```"
	_, entities := Convert(md, false, nil)
	pre := findEntity(entities, "pre")
	if pre == nil {
		t.Fatal("Convert() should have pre entity")
	}
	if pre.Language != "" {
		t.Errorf("Pre entity language = %q, want empty", pre.Language)
	}
}

// TestHeading_H1 测试 H1 标题
func TestHeading_H1(t *testing.T) {
	text, entities := Convert("# Title", false, nil)
	if !strings.Contains(text, "📌") {
		t.Errorf("H1 should contain emoji 📌")
	}
	if findEntity(entities, "bold") == nil {
		t.Error("H1 should have bold entity")
	}
	if findEntity(entities, "underline") == nil {
		t.Error("H1 should have underline entity")
	}
}

// TestHeading_H2 测试 H2 标题
func TestHeading_H2(t *testing.T) {
	text, entities := Convert("## Subtitle", false, nil)
	if !strings.Contains(text, "📝") {
		t.Errorf("H2 should contain emoji 📝")
	}
	if findEntity(entities, "bold") == nil {
		t.Error("H2 should have bold entity")
	}
	if findEntity(entities, "underline") == nil {
		t.Error("H2 should have underline entity")
	}
}

// TestHeading_H3 测试 H3 标题
func TestHeading_H3(t *testing.T) {
	text, entities := Convert("### Section", false, nil)
	if !strings.Contains(text, "📋") {
		t.Errorf("H3 should contain emoji 📋")
	}
	if findEntity(entities, "bold") == nil {
		t.Error("H3 should have bold entity")
	}
	// H3 无下划线
	if findEntity(entities, "underline") != nil {
		t.Error("H3 should not have underline entity")
	}
}

// TestLink_Inline 测试行内链接
func TestLink_Inline(t *testing.T) {
	text, entities := Convert("[Google](https://google.com)", false, nil)
	link := findEntity(entities, "text_link")
	if link == nil {
		t.Fatal("Convert() should have text_link entity")
	}
	if link.URL != "https://google.com" {
		t.Errorf("Link URL = %q, want 'https://google.com'", link.URL)
	}
	if extractEntityText(text, link) != "Google" {
		t.Errorf("Link text = %q, want 'Google'", extractEntityText(text, link))
	}
}

// TestBlockquote_Simple 测试简单引用
func TestBlockquote_Simple(t *testing.T) {
	text, entities := Convert("> quoted text", false, nil)
	bq := findEntity(entities, "blockquote")
	if bq == nil {
		// 可能是 expandable_blockquote
		bq = findEntity(entities, "expandable_blockquote")
	}
	if bq == nil {
		t.Fatal("Convert() should have blockquote entity")
	}
	extracted := extractEntityText(text, bq)
	if !strings.Contains(extracted, "quoted text") {
		t.Errorf("Blockquote should contain 'quoted text', got %q", extracted)
	}
}

// TestList_Unordered 测试无序列表
func TestList_Unordered(t *testing.T) {
	md := "- item1\n- item2"
	text, _ := Convert(md, false, nil)
	if !strings.Contains(text, "item1") {
		t.Errorf("Unordered list should contain 'item1'")
	}
	if !strings.Contains(text, "item2") {
		t.Errorf("Unordered list should contain 'item2'")
	}
}

// TestList_Ordered 测试有序列表
func TestList_Ordered(t *testing.T) {
	md := "1. first\n2. second"
	text, _ := Convert(md, false, nil)
	if !strings.Contains(text, "1. first") {
		t.Errorf("Ordered list should contain '1. first'")
	}
	if !strings.Contains(text, "2. second") {
		t.Errorf("Ordered list should contain '2. second'")
	}
}

// TestList_Task 测试任务列表
func TestList_Task(t *testing.T) {
	md := "- [x] done\n- [ ] todo"
	text, _ := Convert(md, false, nil)
	if !strings.Contains(text, "✅") {
		t.Errorf("Task list should contain completed emoji ✅")
	}
	if !strings.Contains(text, "☑") {
		t.Errorf("Task list should contain uncompleted emoji ☑")
	}
}

// TestSpoiler 测试剧透
func TestSpoiler(t *testing.T) {
	text, entities := Convert("this is ||secret|| text", false, nil)
	spoiler := findEntity(entities, "spoiler")
	if spoiler == nil {
		t.Fatal("Convert() should have spoiler entity")
	}
	if extractEntityText(text, spoiler) != "secret" {
		t.Errorf("Spoiler entity text = %q, want 'secret'", extractEntityText(text, spoiler))
	}
}

// TestRule_HorizontalRule 测试水平线
func TestRule_HorizontalRule(t *testing.T) {
	text, _ := Convert("above\n\n---\n\nbelow", false, nil)
	if !strings.Contains(text, "————————") {
		t.Errorf("Horizontal rule should contain ————————")
	}
}

// TestUTF16Offset_Emoji 测试 emoji 的 UTF-16 偏移
func TestUTF16Offset_Emoji(t *testing.T) {
	// 📌 is 2 UTF-16 code units
	_, entities := Convert("📌 **bold**", false, nil)
	bold := findEntity(entities, "bold")
	if bold == nil {
		t.Fatal("Convert() should have bold entity")
	}
	// "📌 " = 2 + 1 = 3 UTF-16 code units
	if bold.Offset != 3 {
		t.Errorf("Bold offset = %d, want 3", bold.Offset)
	}
	if bold.Length != 4 {
		t.Errorf("Bold length = %d, want 4", bold.Length)
	}
}

// TestUTF16Offset_CJK 测试中日韩字符的 UTF-16 偏移
func TestUTF16Offset_CJK(t *testing.T) {
	_, entities := Convert("你好 **世界**", false, nil)
	bold := findEntity(entities, "bold")
	if bold == nil {
		t.Fatal("Convert() should have bold entity")
	}
	// "你好 " = 2 + 1 = 3 UTF-16 code units (CJK is BMP, 1 each)
	if bold.Offset != 3 {
		t.Errorf("Bold offset = %d, want 3", bold.Offset)
	}
	if bold.Length != 2 {
		t.Errorf("Bold length = %d, want 2", bold.Length)
	}
}

// TestComplexDocument 测试复杂文档
func TestComplexDocument(t *testing.T) {
	md := `# Hello World

This is **bold** and *italic* text.

- item 1
- item 2

> A quote

` + "```python\nprint(\"hello\")\n```"

	text, entities := Convert(md, false, nil)
	
	// 应该有多种类型的 entities
	types := make(map[string]bool)
	for _, e := range entities {
		types[e.Type] = true
	}
	
	if !types["bold"] {
		t.Error("Should have bold entity")
	}
	if !types["italic"] {
		t.Error("Should have italic entity")
	}
	if !types["blockquote"] && !types["expandable_blockquote"] {
		t.Error("Should have blockquote entity")
	}
	if !types["pre"] {
		t.Error("Should have pre entity")
	}
	
	// 文本应该包含所有内容
	if !strings.Contains(text, "Hello World") {
		t.Error("Text should contain 'Hello World'")
	}
	if !strings.Contains(text, "item 1") {
		t.Error("Text should contain 'item 1'")
	}
	if !strings.Contains(text, "A quote") {
		t.Error("Text should contain 'A quote'")
	}
	if !strings.Contains(text, `print("hello")`) {
		t.Error("Text should contain print(\"hello\")")
	}
}

