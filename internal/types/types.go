package types

// MessageEntity 表示 Telegram 消息实体
type MessageEntity struct {
	Type          string `json:"type"`
	Offset        int    `json:"offset"`
	Length        int    `json:"length"`
	URL           string `json:"url,omitempty"`
	Language      string `json:"language,omitempty"`
	CustomEmojiID string `json:"custom_emoji_id,omitempty"`
}

// ToDict 将 MessageEntity 转换为 map
func (e MessageEntity) ToDict() map[string]interface{} {
	result := map[string]interface{}{
		"type":   e.Type,
		"offset": e.Offset,
		"length": e.Length,
	}
	if e.URL != "" {
		result["url"] = e.URL
	}
	if e.Language != "" {
		result["language"] = e.Language
	}
	if e.CustomEmojiID != "" {
		result["custom_emoji_id"] = e.CustomEmojiID
	}
	return result
}

// Symbol 定义 Markdown 元素的显示符号
type Symbol struct {
	HeadingLevel1   string
	HeadingLevel2   string
	HeadingLevel3   string
	HeadingLevel4   string
	HeadingLevel5   string
	HeadingLevel6   string
	Quote           string
	Image           string
	TaskCompleted   string
	TaskUncompleted string
}

// DefaultSymbol 返回默认符号配置
func DefaultSymbol() *Symbol {
	return &Symbol{
		HeadingLevel1:   "📌",
		HeadingLevel2:   "📝",
		HeadingLevel3:   "📋",
		HeadingLevel4:   "📄",
		HeadingLevel5:   "📃",
		HeadingLevel6:   "🔖",
		Quote:           "💬",
		Image:           "🖼",
		TaskCompleted:   "✅",
		TaskUncompleted: "☑️",
	}
}

// RenderConfig 渲染配置
type RenderConfig struct {
	MarkdownSymbol *Symbol
	CiteExpandable bool
}

// DefaultRenderConfig 返回默认渲染配置
func DefaultRenderConfig() *RenderConfig {
	return &RenderConfig{
		MarkdownSymbol: DefaultSymbol(),
		CiteExpandable: true,
	}
}

