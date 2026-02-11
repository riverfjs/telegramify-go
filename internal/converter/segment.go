package converter

// Segment 记录代码块或 Mermaid 图的位置信息
type Segment struct {
	Kind       string // "code_block" or "mermaid"
	TextStart  int    // 文本起始位置（字节）
	TextEnd    int    // 文本结束位置（字节）
	UTF16Start int    // UTF-16 起始位置
	UTF16End   int    // UTF-16 结束位置
	Language   string // 编程语言或 "mermaid"
	RawCode    string // 原始代码内容
}

// EntityScope 用于跟踪未闭合的实体
type EntityScope struct {
	EntityType    string
	StartOffset   int
	URL           string
	Language      string
	CustomEmojiID string
}

