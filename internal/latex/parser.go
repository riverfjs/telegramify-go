package latex

import (
	"regexp"
	"strings"
	"unicode"
)

// Parser 递归下降 LaTeX→Unicode 转换引擎
//
// 设计原则：
// 1. 数据驱动 — 符号映射集中在 symbols.go
// 2. 鲁棒降级 — 未知命令返回原文，不崩溃
// 3. 标准 LaTeX 语法 — 可选参数用 [...]
// 4. Unicode 优先 — 尽量用 Unicode，无法表示时用可读 ASCII 近似
type Parser struct{}

// NewParser 创建新的 LaTeX 解析器
func NewParser() *Parser {
	return &Parser{}
}

// ──────────────────────────────────────────────
// 静态工具方法
// ──────────────────────────────────────────────

// TranslateCombining 将组合字符应用于文本
func TranslateCombining(command, text string) string {
	sample, ok := Combining[command]
	if !ok {
		return text
	}
	
	combiningChar := sample.Char
	combiningType := sample.Type
	
	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}
	
	switch combiningType {
	case FirstChar:
		// 应用到第一个字符（跳过空格和组合字符）
		i := 1
		for i < len(runes) && (unicode.IsSpace(runes[i]) || isCombiningChar(runes[i])) {
			i++
		}
		if i >= len(runes) {
			return string(runes) + string(combiningChar)
		}
		return string(runes[:i]) + string(combiningChar) + string(runes[i:])
		
	case LastChar:
		return text + string(combiningChar)
		
	case AllChars:
		var result strings.Builder
		for _, r := range runes {
			result.WriteRune(r)
			result.WriteRune(combiningChar)
		}
		return result.String()
	}
	
	return text
}

// MakeNot 生成带否定符号的字符
func MakeNot(negated string) string {
	trimmed := strings.TrimSpace(negated)
	if trimmed == "" {
		return " "
	}
	if notSymbol, ok := NotMap[trimmed]; ok {
		return notSymbol
	}
	// 默认：添加组合长斜线
	runes := []rune(trimmed)
	if len(runes) > 0 {
		return string(runes[0]) + "\u0338" + string(runes[1:])
	}
	return trimmed
}

// ──────────────────────────────────────────────
// 上下标
// ──────────────────────────────────────────────

// TryMakeSubscript 尝试将文本完整转换为 Unicode 下标，失败返回空字符串
func TryMakeSubscript(text string) string {
	if text == "" {
		return ""
	}
	var result strings.Builder
	for _, ch := range text {
		if sub, ok := Subscripts[ch]; ok {
			result.WriteRune(sub)
		} else {
			return "" // 不是所有字符都能转下标
		}
	}
	return result.String()
}

// MakeSubscript 生成下标表示
func MakeSubscript(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	subscript := TryMakeSubscript(text)
	if subscript != "" {
		return subscript
	}
	if len([]rune(text)) == 1 {
		r := []rune(text)[0]
		if sub, ok := Subscripts[r]; ok {
			return string(sub)
		}
		return "_" + text
	}
	return "_(" + text + ")"
}

// TryMakeSuperscript 尝试将文本完整转换为 Unicode 上标，失败返回空字符串
func TryMakeSuperscript(text string) string {
	if text == "" {
		return ""
	}
	var result strings.Builder
	for _, ch := range text {
		if sup, ok := Superscripts[ch]; ok {
			result.WriteRune(sup)
		} else {
			return "" // 不是所有字符都能转上标
		}
	}
	return result.String()
}

// MakeSuperscript 生成上标表示
func MakeSuperscript(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	superscript := TryMakeSuperscript(text)
	if superscript != "" {
		return superscript
	}
	if len([]rune(text)) == 1 {
		r := []rune(text)[0]
		if sup, ok := Superscripts[r]; ok {
			return string(sup)
		}
		return "^" + text
	}
	return "^(" + text + ")"
}

// ──────────────────────────────────────────────
// 样式、分数、根号
// ──────────────────────────────────────────────

// TranslateStyles 翻译样式命令（粗体、斜体、正体等）
func TranslateStyles(command, text string) string {
	styleMap, ok := LatexStyles[command]
	if !ok {
		return text // 未知命令
	}
	if styleMap == nil {
		// mathrm, mathsf: 原样返回
		return text
	}
	
	var result strings.Builder
	for _, ch := range text {
		if styled, ok := styleMap[ch]; ok {
			result.WriteRune(styled)
		} else {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

// MakeSqrt 生成根号的 Unicode 表示
func MakeSqrt(index, radicand string) string {
	var radix string
	switch index {
	case "", "2":
		radix = "√"
	case "3":
		radix = "∛"
	case "4":
		radix = "∜"
	default:
		sup := TryMakeSuperscript(index)
		if sup != "" {
			radix = sup + "√"
		} else {
			radix = "(" + index + ")√"
		}
	}
	return radix + TranslateCombining("\\overline", radicand)
}

// TranslateSqrt 翻译 \sqrt 命令
func TranslateSqrt(command, option, param string) string {
	if command != "\\sqrt" {
		return command
	}
	return MakeSqrt(strings.TrimSpace(option), strings.TrimSpace(param))
}

// MakeFraction 生成分数的 Unicode 表示
func MakeFraction(numerator, denominator string) string {
	n, d := strings.TrimSpace(numerator), strings.TrimSpace(denominator)
	if n == "" && d == "" {
		return ""
	}
	key := [2]string{n, d}
	if frac, ok := FracMap[key]; ok {
		return frac
	}
	return maybeParenthesize(n) + "/" + maybeParenthesize(d)
}

// maybeParenthesize adds parentheses if text contains special characters.
func maybeParenthesize(text string) string {
	for _, r := range text {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && !isCombiningChar(r) && r != '_' {
			return "(" + text + ")"
		}
	}
	return text
}

// isCombiningChar returns true if the rune is a Unicode combining character.
func isCombiningChar(r rune) bool {
	return (r >= '\u0300' && r <= '\u036F') ||
		(r >= '\u1AB0' && r <= '\u1AFF') ||
		(r >= '\u1DC0' && r <= '\u1DFF') ||
		(r >= '\u20D0' && r <= '\u20FF') ||
		(r >= '\uFE20' && r <= '\uFE2F')
}

// TranslateFrac 翻译 \frac 命令
func TranslateFrac(command, numerator, denominator string) string {
	if command != "\\frac" {
		return command
	}
	return MakeFraction(numerator, denominator)
}

// TranslateEscape 查找 LaTeX 符号
func TranslateEscape(name string) string {
	if symbol, ok := LatexSymbols[name]; ok {
		return symbol
	}
	return name
}

// ──────────────────────────────────────────────
// 解析器核心
// ──────────────────────────────────────────────

// Parse 递归下降解析 LaTeX 字符串，转换为 Unicode
func (p *Parser) Parse(latex string) string {
	var result []string
	i := 0
	
	for i < len(latex) {
		if latex[i] == '\\' {
			command, newIdx := p.parseCommand(latex, i)
			// 混合分数格式（数字后紧跟 \frac）
			if command == "\\frac" && len(result) > 0 && len(result[len(result)-1]) > 0 {
				lastChar := result[len(result)-1][len(result[len(result)-1])-1]
				if lastChar >= '0' && lastChar <= '9' {
					result[len(result)-1] += " "
				}
			}
			handled, newIdx := p.handleCommand(command, latex, newIdx)
			result = append(result, handled)
			i = newIdx
			
		} else if latex[i] == '{' {
			block, newIdx := p.parseBlock(latex, i)
			result = append(result, block)
			i = newIdx
			
		} else if latex[i] == '_' || latex[i] == '^' {
			sym := latex[i]
			arg := ""
			i++
			
			if i < len(latex) && latex[i] == '{' {
				arg, i = p.parseBlock(latex, i)
			} else if i < len(latex) && latex[i] == '\\' {
				command, newIdx := p.parseCommand(latex, i)
				if command == "\\frac" && len(result) > 0 && len(result[len(result)-1]) > 0 {
					lastChar := result[len(result)-1][len(result[len(result)-1])-1]
					if lastChar >= '0' && lastChar <= '9' {
						result[len(result)-1] += " "
					}
				}
				arg, i = p.handleCommand(command, latex, newIdx)
			} else if i < len(latex) {
				arg = string(latex[i])
				i++
			}
			
			if sym == '_' {
				result = append(result, MakeSubscript(arg))
			} else {
				result = append(result, MakeSuperscript(arg))
			}
			
		} else if unicode.IsSpace(rune(latex[i])) {
			spaces, newIdx := p.parseSpaces(latex, i)
			result = append(result, spaces)
			i = newIdx
			
		} else {
			result = append(result, string(latex[i]))
			i++
		}
	}
	
	return strings.Join(result, "")
}

// ──────────────────────────────────────────────
// 命令分派（有序优先级）
// ──────────────────────────────────────────────

func (p *Parser) handleCommand(command, latex string, index int) (string, int) {
	// 1. 符号表直查（最常见路径）
	if _, ok := LatexSymbols[command]; ok {
		return TranslateEscape(command), index
	}
	
	// 2. \not 前缀否定
	if command == "\\not" {
		if index < len(latex) {
			if latex[index] == '\\' {
				nextCmd, nextIdx := p.parseCommand(latex, index)
				symbol := LatexSymbols[nextCmd]
				if symbol == "" {
					symbol = nextCmd
				}
				return MakeNot(symbol), nextIdx
			}
			return MakeNot(string(latex[index])), index + 1
		}
		return "\u0338", index
	}
	
	// 3. 组合字符命令（\hat, \bar, \vec, \dot 等）
	if _, ok := Combining[command]; ok {
		arg, newIdx := p.parseBlock(latex, index)
		return TranslateCombining(command, arg), newIdx
	}
	
	// 4. \frac{num}{den}
	if command == "\\frac" {
		numer, idx1 := p.parseBlock(latex, index)
		denom, idx2 := p.parseBlock(latex, idx1)
		return MakeFraction(numer, denom), idx2
	}
	
	// 5. \sqrt[n]{x} — 可选参数用 []
	if command == "\\sqrt" {
		option, idx1 := p.parseOptional(latex, index)
		param, idx2 := p.parseBlock(latex, idx1)
		return TranslateSqrt(command, option, param), idx2
	}
	
	// 6. 样式命令（\mathbb, \mathbf, \mathrm, \mathit 等）
	if _, ok := LatexStyles[command]; ok {
		text, newIdx := p.parseBlock(latex, index)
		return TranslateStyles(command, text), newIdx
	}
	
	// 7. 文本直通命令
	textCommands := map[string]bool{
		"\\text": true, "\\operatorname": true, "\\mbox": true,
		"\\textrm": true, "\\textup": true, "\\mathop": true,
	}
	if textCommands[command] {
		text, newIdx := p.parseBlock(latex, index)
		return text, newIdx
	}
	
	// 8. \left / \right 定界符
	if command == "\\left" || command == "\\right" {
		delim, newIdx := p.parseDelimiter(latex, index)
		return delim, newIdx
	}
	
	// 9. \binom{n}{k} / \tbinom / \dbinom
	if command == "\\binom" || command == "\\tbinom" || command == "\\dbinom" {
		nVal, idx1 := p.parseBlock(latex, index)
		kVal, idx2 := p.parseBlock(latex, idx1)
		return "C(" + nVal + "," + kVal + ")", idx2
	}
	
	// 10. \boxed{x}
	if command == "\\boxed" {
		text, newIdx := p.parseBlock(latex, index)
		return "[" + text + "]", newIdx
	}
	
	// 11. \pmod{p}
	if command == "\\pmod" {
		text, newIdx := p.parseBlock(latex, index)
		return " (mod " + text + ")", newIdx
	}
	
	// 12. \phantom / \hphantom / \vphantom — 等宽空白
	if command == "\\phantom" || command == "\\hphantom" || command == "\\vphantom" {
		text, newIdx := p.parseBlock(latex, index)
		length := len(text)
		if length < 1 {
			length = 1
		}
		return strings.Repeat(" ", length), newIdx
	}
	
	// 13. \overset{over}{base}
	if command == "\\overset" {
		over, idx1 := p.parseBlock(latex, index)
		base, idx2 := p.parseBlock(latex, idx1)
		sup := TryMakeSuperscript(over)
		if sup != "" {
			return base + sup, idx2
		}
		return base + "^(" + over + ")", idx2
	}
	
	// 14. \underset{under}{base}
	if command == "\\underset" {
		under, idx1 := p.parseBlock(latex, index)
		base, idx2 := p.parseBlock(latex, idx1)
		sub := TryMakeSubscript(under)
		if sub != "" {
			return base + sub, idx2
		}
		return base + "_(" + under + ")", idx2
	}
	
	// 15. \stackrel{over}{base}
	if command == "\\stackrel" {
		over, idx1 := p.parseBlock(latex, index)
		base, idx2 := p.parseBlock(latex, idx1)
		sup := TryMakeSuperscript(over)
		if sup != "" {
			return base + sup, idx2
		}
		return base + "^(" + over + ")", idx2
	}
	
	// 16. \substack{...} — 多行下标
	if command == "\\substack" {
		text, newIdx := p.parseBlock(latex, index)
		lines := strings.Split(text, "\\\\")
		var parsed []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				parsed = append(parsed, p.Parse(trimmed))
			}
		}
		return strings.Join(parsed, ", "), newIdx
	}
	
	// 17. \color{...} — 忽略颜色参数
	if command == "\\color" {
		_, newIdx := p.parseBlock(latex, index)
		return "", newIdx
	}
	
	// 18. \cancel / \bcancel / \xcancel / \sout — 删除线效果
	cancelCommands := map[string]bool{
		"\\cancel": true, "\\bcancel": true, "\\xcancel": true, "\\sout": true,
	}
	if cancelCommands[command] {
		text, newIdx := p.parseBlock(latex, index)
		return TranslateCombining("\\underline", text), newIdx
	}
	
	// 19. \overbrace / \underbrace
	if command == "\\overbrace" {
		text, newIdx := p.parseBlock(latex, index)
		return TranslateCombining("\\overline", text), newIdx
	}
	if command == "\\underbrace" {
		text, newIdx := p.parseBlock(latex, index)
		return TranslateCombining("\\underline", text), newIdx
	}
	
	// 20. \xrightarrow / \xleftarrow
	if command == "\\xrightarrow" {
		text, newIdx := p.parseBlock(latex, index)
		if strings.TrimSpace(text) != "" {
			return "→(" + text + ")", newIdx
		}
		return "→", newIdx
	}
	if command == "\\xleftarrow" {
		text, newIdx := p.parseBlock(latex, index)
		if strings.TrimSpace(text) != "" {
			return "←(" + text + ")", newIdx
		}
		return "←", newIdx
	}
	
	// 21. \begin{...}\end{...} 环境
	if command == "\\begin" {
		envName, idx1 := p.parseEnvName(latex, index)
		content, idx2 := p.parseEnvironment(latex, idx1, envName)
		return p.renderEnvironment(envName, content), idx2
	}
	if command == "\\end" {
		envName, newIdx := p.parseEnvName(latex, index)
		_ = envName
		return "", newIdx
	}
	
	// 22. 兜底：返回原始命令文本
	return command, index
}

// ──────────────────────────────────────────────
// 底层解析方法
// ──────────────────────────────────────────────

var commandRegex = regexp.MustCompile(`^\\([a-zA-Z]+|.)`)

func (p *Parser) parseCommand(latex string, start int) (string, int) {
	match := commandRegex.FindStringSubmatch(latex[start:])
	if match != nil {
		return match[0], start + len(match[0])
	}
	return "\\", start + 1
}

func (p *Parser) parseBlock(latex string, start int) (string, int) {
	if start >= len(latex) {
		return "", start
	}
	if latex[start] != '{' {
		// 无 {} 包裹 — 读取单个 token（标准 LaTeX 行为）
		if latex[start] == '\\' {
			cmd, newIdx := p.parseCommand(latex, start)
			return p.handleCommand(cmd, latex, newIdx)
		}
		return string(latex[start]), start + 1
	}
	
	// 标准 {...} 块解析
	level, pos := 1, start+1
	for pos < len(latex) && level > 0 {
		if latex[pos] == '{' {
			level++
		} else if latex[pos] == '}' {
			level--
		}
		pos++
	}
	return p.Parse(latex[start+1 : pos-1]), pos
}

func (p *Parser) parseOptional(latex string, start int) (string, int) {
	if start >= len(latex) || latex[start] != '[' {
		return "", start
	}
	level, pos := 1, start+1
	for pos < len(latex) && level > 0 {
		if latex[pos] == '[' {
			level++
		} else if latex[pos] == ']' {
			level--
		}
		pos++
	}
	return p.Parse(latex[start+1 : pos-1]), pos
}

func (p *Parser) parseSpaces(latex string, start int) (string, int) {
	end := start
	hasNewline := false
	for end < len(latex) && unicode.IsSpace(rune(latex[end])) {
		if latex[end] == '\n' {
			hasNewline = true
		}
		end++
	}
	if hasNewline {
		return "\n\n", end
	}
	return " ", end
}

// ──────────────────────────────────────────────
// 定界符解析
// ──────────────────────────────────────────────

func (p *Parser) parseDelimiter(latex string, index int) (string, int) {
	if index >= len(latex) {
		return "", index
	}
	ch := latex[index]
	if ch == '\\' {
		cmdMatch := commandRegex.FindStringSubmatch(latex[index:])
		if cmdMatch != nil {
			cmd := cmdMatch[0]
			symbol := LatexSymbols[cmd]
			if symbol == "" {
				symbol = strings.TrimPrefix(cmd, "\\")
			}
			return symbol, index + len(cmdMatch[0])
		}
		return "\\", index + 1
	}
	if ch == '.' {
		return "", index + 1 // 不可见定界符
	}
	return string(ch), index + 1
}

// ──────────────────────────────────────────────
// 环境解析与渲染
// ──────────────────────────────────────────────

func (p *Parser) parseEnvName(latex string, index int) (string, int) {
	if index < len(latex) && latex[index] == '{' {
		close := strings.IndexByte(latex[index:], '}')
		if close != -1 {
			return latex[index+1 : index+close], index + close + 1
		}
	}
	return "", index
}

func (p *Parser) parseEnvironment(latex string, index int, envName string) (string, int) {
	endMarker := "\\end{" + envName + "}"
	endPos := strings.Index(latex[index:], endMarker)
	if endPos == -1 {
		return latex[index:], len(latex)
	}
	return latex[index : index+endPos], index + endPos + len(endMarker)
}

// 矩阵类环境类型 → (左定界符, 右定界符)
var matrixTypes = map[string][2]string{
	"matrix":      {"", ""},
	"pmatrix":     {"(", ")"},
	"bmatrix":     {"[", "]"},
	"Bmatrix":     {"{", "}"},
	"vmatrix":     {"|", "|"},
	"Vmatrix":     {"‖", "‖"},
	"smallmatrix": {"", ""},
}

// align 类环境
var alignTypes = map[string]bool{
	"align": true, "aligned": true, "gather": true, "gathered": true,
	"equation": true, "equation*": true, "multline": true, "multline*": true,
	"split": true, "flalign": true, "flalign*": true,
}

func (p *Parser) renderEnvironment(envName, content string) string {
	if delims, ok := matrixTypes[envName]; ok {
		compact := (envName == "smallmatrix")
		return p.renderMatrix(content, delims[0], delims[1], compact)
	}
	if envName == "cases" {
		return p.renderCases(content)
	}
	if alignTypes[envName] {
		return p.renderAlign(content)
	}
	if envName == "array" {
		return p.renderArray(content)
	}
	// 未知环境 — 直接解析内容
	return p.Parse(content)
}

func (p *Parser) renderMatrix(content, left, right string, compact bool) string {
	rows := strings.Split(content, "\\\\")
	var rendered []string
	for _, row := range rows {
		trimmed := strings.TrimSpace(row)
		if trimmed == "" {
			continue
		}
		cells := strings.Split(trimmed, "&")
		var parsedCells []string
		for _, cell := range cells {
			parsedCells = append(parsedCells, p.Parse(strings.TrimSpace(cell)))
		}
		sep := "  "
		if compact {
			sep = ", "
		}
		rendered = append(rendered, strings.Join(parsedCells, sep))
	}
	joiner := "\n"
	if compact {
		joiner = "; "
	}
	body := strings.Join(rendered, joiner)
	if left != "" || right != "" {
		return left + body + right
	}
	return body
}

func (p *Parser) renderCases(content string) string {
	rows := strings.Split(content, "\\\\")
	var parts []string
	for _, row := range rows {
		trimmed := strings.TrimSpace(row)
		if trimmed == "" {
			continue
		}
		segments := strings.SplitN(trimmed, "&", 2)
		val := p.Parse(strings.TrimSpace(segments[0]))
		cond := ""
		if len(segments) > 1 {
			cond = p.Parse(strings.TrimSpace(segments[1]))
		}
		if cond != "" {
			parts = append(parts, val+", "+cond)
		} else {
			parts = append(parts, val)
		}
	}
	
	n := len(parts)
	if n == 0 {
		return ""
	}
	if n == 1 {
		return "\u23A7 " + parts[0]
	}
	
	var lines []string
	for i, part := range parts {
		if i == 0 {
			lines = append(lines, "\u23A7 "+part)
		} else if i == n-1 {
			lines = append(lines, "\u23A9 "+part)
		} else {
			lines = append(lines, "\u23A8 "+part)
		}
	}
	return strings.Join(lines, "\n")
}

func (p *Parser) renderAlign(content string) string {
	rows := strings.Split(content, "\\\\")
	var rendered []string
	for _, row := range rows {
		trimmed := strings.TrimSpace(row)
		if trimmed == "" {
			continue
		}
		// 移除 & 符号
		cleaned := strings.ReplaceAll(trimmed, "&", " ")
		rendered = append(rendered, p.Parse(cleaned))
	}
	return strings.Join(rendered, "\n")
}

func (p *Parser) renderArray(content string) string {
	// array 第一个 {} 是列格式说明（如 {ccc}），跳过
	stripped := strings.TrimSpace(content)
	if strings.HasPrefix(stripped, "{") {
		close := strings.IndexByte(stripped, '}')
		if close != -1 {
			content = stripped[close+1:]
		}
	}
	return p.renderMatrix(content, "", "", false)
}

// ──────────────────────────────────────────────
// 公开接口
// ──────────────────────────────────────────────

// Convert 将 LaTeX 字符串转换为 Unicode 文本。出错时返回原文。
func (p *Parser) Convert(latex string) string {
	defer func() {
		if r := recover(); r != nil {
			// 出错时返回原文
		}
	}()
	return p.Parse(latex)
}

