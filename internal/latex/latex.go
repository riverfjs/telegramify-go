package latex

import (
	"regexp"
	"strings"
)

var (
	latexMathBlockRegex  = regexp.MustCompile(`\\\[(.*?)\\\]`)
	latexInlineRegex     = regexp.MustCompile(`\\\((.*?)\\\)`)
)

// ContainsLatexSymbols 检查内容是否包含 LaTeX 符号
func ContainsLatexSymbols(content string) bool {
	if len(content) < 5 {
		return false
	}
	// 检查常见 LaTeX 命令
	patterns := []string{`\frac`, `\sqrt`, `\begin`}
	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	// 检查符号表中的符号
	for cmd := range LatexSymbols {
		if len(cmd) > 2 && strings.Contains(content, cmd) {
			return true
		}
	}
	return false
}

// EscapeLaTeX 预处理 LaTeX \[...\] 和 \(...\) 块转为 Unicode
func EscapeLaTeX(text string) string {
	parser := NewParser()
	
	// 处理块级数学
	text = latexMathBlockRegex.ReplaceAllStringFunc(text, func(match string) string {
		content := latexMathBlockRegex.FindStringSubmatch(match)[1]
		if !ContainsLatexSymbols(content) {
			return match
		}
		converted := parser.Convert(content)
		return "$$" + strings.TrimSpace(converted) + "$$"
	})
	
	// 处理行内数学
	text = latexInlineRegex.ReplaceAllStringFunc(text, func(match string) string {
		content := latexInlineRegex.FindStringSubmatch(match)[1]
		if !ContainsLatexSymbols(content) {
			return match
		}
		converted := parser.Convert(content)
		return "$" + strings.TrimSpace(strings.TrimRight(converted, "\n")) + "$"
	})
	
	return text
}

