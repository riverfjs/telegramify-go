package telegramify

import (
	"github.com/riverfjs/telegramify-go/internal/converter"
	"github.com/riverfjs/telegramify-go/internal/latex"
	"github.com/riverfjs/telegramify-go/internal/parser"
)

// Convert 将 Markdown 转换为 (plain_text, entities) 用于 Telegram
//
// 参数:
//   - markdown: 原始 Markdown 文本
//   - latexEscape: 是否将 LaTeX 转换为 Unicode
//   - config: 渲染配置，如为 nil 则使用默认配置
//
// 返回:
//   - string: 纯文本
//   - []MessageEntity: 实体列表
func Convert(markdown string, latexEscape bool, config *RenderConfig) (string, []MessageEntity) {
	text, entities, _ := ConvertWithSegments(markdown, latexEscape, config)
	return text, entities
}

// ConvertWithSegments 将 Markdown 转换为 (plain_text, entities, segments)
//
// 类似 Convert()，但还返回 segment 信息供管道使用
//
// 参数:
//   - markdown: 原始 Markdown 文本
//   - latexEscape: 是否将 LaTeX 转换为 Unicode
//   - config: 渲染配置，如为 nil 则使用默认配置
//
// 返回:
//   - string: 纯文本
//   - []MessageEntity: 实体列表
//   - []converter.Segment: 代码块/Mermaid 片段信息
func ConvertWithSegments(markdown string, latexEscape bool, config *RenderConfig) (string, []MessageEntity, []converter.Segment) {
	if config == nil {
		config = DefaultConfig()
	}
	
	// 预处理
	preprocessed := markdown
	if latexEscape {
		latexHelper := latex.NewParser()
		preprocessed = converter.EscapeLatex(preprocessed, latexHelper)
	}
	preprocessed = converter.PreprocessSpoilers(preprocessed)
	
	// 解析（类型已通过别名统一）
	text, entities, segments := parser.Parse(preprocessed, config)
	return text, entities, segments
}

