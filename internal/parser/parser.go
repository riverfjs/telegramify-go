package parser

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"

	"github.com/riverfjs/telegramify-go/internal/converter"
	"github.com/riverfjs/telegramify-go/internal/types"
)

// StandardOptions goldmark 扩展配置，对应 pyromark 的 STANDARD_OPTIONS
var StandardOptions = []goldmark.Option{
	goldmark.WithExtensions(
		extension.GFM,            // GitHub Flavored Markdown (tables, strikethrough, tasklists)
		extension.DefinitionList, // 定义列表
		extension.Footnote,       // 脚注
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(), // 自动生成标题 ID
	),
}

// Parse 解析 Markdown 并遍历 AST 生成 (text, entities, segments)
func Parse(markdown string, config *converter.RenderConfig) (string, []converter.MessageEntity, []converter.Segment) {
	if config == nil {
		config = types.DefaultRenderConfig()
	}
	// 创建 goldmark 解析器
	md := goldmark.New(StandardOptions...)
	
	// 解析为 AST
	source := []byte(markdown)
	reader := text.NewReader(source)
	node := md.Parser().Parse(reader)
	
	// 创建 Walker
	walker := converter.NewEventWalker(source, config)
	
	// 遍历 AST
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		return walker.Walk(n, entering)
	})
	
	return walker.Result()
}

// ParseWithCustomRenderer 使用自定义渲染器（预留）
func ParseWithCustomRenderer(markdown string, config *converter.RenderConfig) (string, []converter.MessageEntity, []converter.Segment) {
	md := goldmark.New(StandardOptions...)
	
	source := []byte(markdown)
	reader := text.NewReader(source)
	node := md.Parser().Parse(reader)
	
	walker := converter.NewEventWalker(source, config)
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		return walker.Walk(n, entering)
	})
	
	return walker.Result()
}

// ParseAST 仅解析为 AST，不遍历
func ParseAST(markdown string) ast.Node {
	md := goldmark.New(StandardOptions...)
	source := []byte(markdown)
	reader := text.NewReader(source)
	return md.Parser().Parse(reader)
}

