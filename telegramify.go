// Package telegramify 将 Markdown 转换为 Telegram plain text + MessageEntity 对
//
// 这个包提供了将原始 Markdown（包括 LLM 输出、GitHub README 等）转换为
// Telegram Bot API 所需格式的功能。
//
// 核心功能：
//   - 将 Markdown 转换为纯文本 + MessageEntity 列表
//   - 支持 LaTeX 公式转 Unicode
//   - 智能拆分长消息
//   - 提取代码块为文件
//   - 渲染 Mermaid 图表（可选）
//
// 主要 API：
//   - Convert(): 同步转换，返回 (text, entities)
//   - Telegramify(): 异步完整处理，返回可发送的内容列表
//
// 示例：
//
//	// 简单转换
//	text, entities := telegramify.Convert(markdown, true, nil)
//
//	// 完整处理（含拆分、文件提取）
//	contents, err := telegramify.Telegramify(ctx, markdown, 4096, true, nil)
//	for _, content := range contents {
//	    switch c := content.(type) {
//	    case *telegramify.Text:
//	        // 发送文本消息
//	    case *telegramify.File:
//	        // 发送文件
//	    case *telegramify.Photo:
//	        // 发送图片
//	    }
//	}
package telegramify

import (
	"context"
)

// Telegramify 将 Markdown 转换为 Telegram 就绪的内容片段
//
// 这是主要的异步 API，用于完整的 Markdown 处理，包括消息拆分、
// 代码块提取和 Mermaid 渲染。对于较低级别的纯文本转换，使用 Convert()。
//
// 参数：
//   - ctx: 上下文
//   - content: 原始 Markdown 文本
//   - maxMessageLength: 每条文本消息的最大 UTF-16 code units（Telegram 限制为 4096）
//   - latexEscape: 是否将 LaTeX \(...\) 和 \[...\] 转换为 Unicode
//   - config: 渲染配置，如为 nil 则使用默认配置
//
// 返回：
//   - []Content: Text、File 或 Photo 对象的有序列表，可直接用于 Telegram Bot API
//   - error: 错误信息
func Telegramify(
	ctx context.Context,
	content string,
	maxMessageLength int,
	latexEscape bool,
	config *RenderConfig,
) ([]Content, error) {
	return ProcessMarkdown(ctx, content, maxMessageLength, latexEscape, config)
}

