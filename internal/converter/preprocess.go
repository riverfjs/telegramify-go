package converter

import (
	"regexp"
	"strings"

	"github.com/riverfjs/telegramify-go/internal/latex"
)

var (
	// _SPOILER_RE 匹配 ||...|| (非转义的 ||)
	spoilerRe = regexp.MustCompile(`(?:[^\\]|^)\|\|(.+?)\|\|`)
	
	// _CODE_REGION_RE 匹配代码块和行内代码
	codeRegionRe = regexp.MustCompile("(```[\\s\\S]*?```|`[^`\\n]+`)")
	
	// LaTeX 块级公式：\[...\]
	latexMathRe = regexp.MustCompile(`\\\[(.*?)\\\]`)
	
	// LaTeX 行内公式：\(...\)
	latexInlineRe = regexp.MustCompile(`\\\((.*?)\\\)`)
)

// PreprocessSpoilers 将 ||spoiler|| 替换为 <tg-spoiler>spoiler</tg-spoiler>
// 跳过代码块和行内代码中的内容
func PreprocessSpoilers(text string) string {
	parts := codeRegionRe.Split(text, -1)
	matches := codeRegionRe.FindAllString(text, -1)
	
	var result strings.Builder
	for i, part := range parts {
		// 偶数索引：非代码区域，进行替换
		if i%2 == 0 || i >= len(matches) {
			// 替换 ||...||
			part = replaceSpoilerTags(part)
		}
		result.WriteString(part)
		
		// 添加回代码块（奇数索引）
		if i < len(matches) {
			result.WriteString(matches[i])
		}
	}
	
	return result.String()
}

// replaceSpoilerTags 将 ||...|| 替换为 <tg-spoiler>...</tg-spoiler>
func replaceSpoilerTags(text string) string {
	// 使用状态机手动处理，避免转义字符问题
	var result strings.Builder
	i := 0
	inSpoiler := false
	
	for i < len(text) {
		// 检查是否是转义的 ||
		if i > 0 && text[i-1] == '\\' && i+1 < len(text) && text[i] == '|' && text[i+1] == '|' {
			result.WriteByte(text[i])
			i++
			continue
		}
		
		// 检查是否是 ||
		if i+1 < len(text) && text[i] == '|' && text[i+1] == '|' {
			if inSpoiler {
				result.WriteString("</tg-spoiler>")
			} else {
				result.WriteString("<tg-spoiler>")
			}
			inSpoiler = !inSpoiler
			i += 2
		} else {
			result.WriteByte(text[i])
			i++
		}
	}
	
	return result.String()
}

// validateTelegramEmoji 如果 URL 是 tg://emoji?id=<19位数字>，返回 id，否则返回空
func validateTelegramEmoji(url string) string {
	const prefix = "tg://emoji?id="
	if !strings.HasPrefix(url, prefix) {
		return ""
	}
	emojiID := strings.TrimPrefix(url, prefix)
	if len(emojiID) == 19 && isDigits(emojiID) {
		return emojiID
	}
	return ""
}

// isDigits 检查字符串是否全为数字
func isDigits(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return len(s) > 0
}

// containsLatexSymbols 检查内容是否包含 LaTeX 符号
func containsLatexSymbols(content string) bool {
	if len(content) < 5 {
		return false
	}
	
	// 检查常见的 LaTeX 命令
	checkSymbols := []string{`\frac`, `\sqrt`, `\begin`}
	for _, sym := range checkSymbols {
		if strings.Contains(content, sym) {
			return true
		}
	}
	
	// 检查 latex.LATEX_SYMBOLS 中的符号
	for key := range latex.LatexSymbols {
		if strings.Contains(content, key) {
			return true
		}
	}
	
	// 检查 NOT_MAP 中的符号
	for key := range latex.NotMap {
		if strings.Contains(content, key) {
			return true
		}
	}
	
	// 检查 LATEX_STYLES 中的符号
	for key := range latex.LatexStyles {
		if strings.Contains(content, key) {
			return true
		}
	}
	
	return false
}

// EscapeLatex 预处理 LaTeX \[...\] 和 \(...\) 块转换为 Unicode
func EscapeLatex(text string, latexHelper *latex.Parser) string {
	// 按段落分割（\n\n）
	lines := strings.Split(text, "\n\n")
	processed := make([]string, len(lines))
	
	for i, line := range lines {
		// 处理块级公式 \[...\]
		line = latexMathRe.ReplaceAllStringFunc(line, func(match string) string {
			return convertLatexMatch(match, true, latexHelper)
		})
		
		// 处理行内公式 \(...\)
		line = latexInlineRe.ReplaceAllStringFunc(line, func(match string) string {
			return convertLatexMatch(match, false, latexHelper)
		})
		
		processed[i] = line
	}
	
	return strings.Join(processed, "\n\n")
}

// convertLatexMatch 转换单个 LaTeX 匹配
func convertLatexMatch(match string, isBlock bool, latexHelper *latex.Parser) string {
	// 提取内容
	var content string
	if isBlock {
		// \[...\] 格式
		content = strings.TrimPrefix(match, `\[`)
		content = strings.TrimSuffix(content, `\]`)
	} else {
		// \(...\) 格式
		content = strings.TrimPrefix(match, `\(`)
		content = strings.TrimSuffix(content, `\)`)
	}
	
	// 检查是否包含 LaTeX 符号
	if !containsLatexSymbols(content) {
		return match
	}
	
	// 转换
	converted := latexHelper.Convert(content)
	converted = strings.TrimSpace(converted)
	converted = strings.Trim(converted, "\n")
	
	// 返回对应格式
	if isBlock {
		return "$$" + strings.TrimSpace(converted) + "$$"
	}
	return "$" + strings.TrimSpace(strings.Trim(converted, "\n")) + "$"
}

