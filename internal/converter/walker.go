package converter

import (
	"fmt"
	"io"
	"strings"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"

	"github.com/riverfjs/telegramify-go/internal/buffer"
	"github.com/riverfjs/telegramify-go/internal/latex"
)

var latexHelper = latex.NewParser()

// EventWalker 遍历 goldmark AST 并生成 (text, entities, segments)
type EventWalker struct {
	buf          *buffer.TextBuffer
	source       []byte
	entityStack  []EntityScope
	entities     []MessageEntity
	segments     []Segment
	config       *RenderConfig

	// Block-level state
	blockCount int // 用于段落间距
	listStack  []interface{} // nil=unordered, *int=ordered(next_number)
	itemStarted bool
	itemIndent string // 当前 item 的缩进，用于 task list marker 替换

	// Table state
	inTable         bool
	tableAlignments []east.Alignment
	tableRows       [][]string
	currentRow      []string
	cellParts       []string
	inTableCell     bool

	// Code block state
	inCodeBlock      bool
	codeBlockLang    string
	codeBlockParts   []string

	// Heading state
	inHeading        bool
	headingEntities  []string

	// Blockquote state
	blockquoteScopes []EntityScope
}

// NewEventWalker 创建新的 EventWalker
func NewEventWalker(source []byte, config *RenderConfig) *EventWalker {
	return &EventWalker{
		buf:          buffer.New(),
		source:       source,
		entityStack:  make([]EntityScope, 0),
		entities:     make([]MessageEntity, 0),
		segments:     make([]Segment, 0),
		config:       config,
		listStack:    make([]interface{}, 0),
		tableRows:    make([][]string, 0),
		currentRow:   make([]string, 0),
		cellParts:    make([]string, 0),
		blockquoteScopes: make([]EntityScope, 0),
		headingEntities:  make([]string, 0),
	}
}

// Walk 遍历 AST 节点
func (w *EventWalker) Walk(node ast.Node, entering bool) (ast.WalkStatus, error) {
	switch n := node.(type) {
	// --- Document ---
	case *ast.Document:
		if !entering {
			// Post-process: upgrade long blockquotes to expandable
			if w.config.CiteExpandable {
				for i := range w.entities {
					if w.entities[i].Type == "blockquote" && w.entities[i].Length > 200 {
						w.entities[i].Type = "expandable_blockquote"
					}
				}
			}
		}

	// --- Inline elements ---
	case *ast.Text:
		if entering {
			w.onText(n.Segment, n.SoftLineBreak(), n.HardLineBreak())
		}

	case *ast.String:
		if entering {
			w.onTextString(n.Value)
		}

	case *ast.CodeSpan:
		if entering {
			w.onInlineCode(n)
			// Skip children to avoid processing the text content twice
			return ast.WalkSkipChildren, nil
		}

	case *ast.Emphasis:
		if entering {
			// Level 1 = italic, Level 2 = bold
			if n.Level == 2 {
				w.pushEntity("bold", "")
			} else {
				w.pushEntity("italic", "")
			}
		} else {
			if n.Level == 2 {
				w.popEntity("bold")
			} else {
				w.popEntity("italic")
			}
		}

	case *east.Strikethrough:
		if entering {
			w.pushEntity("strikethrough", "")
		} else {
			w.popEntity("strikethrough")
		}

	// --- Links & Images ---
	case *ast.Link:
		if entering {
			w.onStartLink(n)
		} else {
			w.popEntity("text_link")
		}

	case *ast.Image:
		if entering {
			w.onStartImage(n)
		} else {
			w.popEntityAny()
		}

	case *ast.AutoLink:
		if entering {
			url := string(n.URL(w.source))
			w.pushEntity("text_link", url)
			w.buf.Write(url)
			return ast.WalkSkipChildren, nil
		}

	// --- Block elements ---
	case *ast.Paragraph:
		if entering {
			w.onStartParagraph()
		} else {
			w.onEndParagraph()
		}

	case *ast.Heading:
		if entering {
			w.onStartHeading(n)
		} else {
			w.onEndHeading()
		}

	case *ast.Blockquote:
		if entering {
			w.onStartBlockquote()
		} else {
			w.onEndBlockquote()
		}

	case *ast.List:
		if entering {
			w.onStartList(n)
		} else {
			w.onEndList()
		}

	case *ast.ListItem:
		if entering {
			w.onStartItem()
		} else {
			w.onEndItem()
		}

	case *east.TaskCheckBox:
		if entering {
			w.onTaskCheckBox(n.IsChecked)
		}

	case *ast.FencedCodeBlock, *ast.CodeBlock:
		if entering {
			w.onStartCodeBlock(n)
			return ast.WalkSkipChildren, nil
		}

	case *ast.ThematicBreak:
		if entering {
			w.onRule()
		}

	case *ast.HTMLBlock:
		// Block HTML ignored
		return ast.WalkSkipChildren, nil

	case *ast.RawHTML:
		if entering {
			w.onInlineHTML(n)
		}

	// --- Table ---
	case *east.Table:
		if entering {
			w.onStartTable(n)
		} else {
			w.onEndTable()
		}

	case *east.TableHeader:
		if entering {
			w.currentRow = make([]string, 0)
		} else {
			w.onEndTableRow()
		}

	case *east.TableRow:
		if entering {
			w.currentRow = make([]string, 0)
		} else {
			w.onEndTableRow()
		}

	case *east.TableCell:
		if entering {
			w.cellParts = make([]string, 0)
			w.inTableCell = true
		} else {
			w.onEndTableCell()
		}
	}

	return ast.WalkContinue, nil
}

// Result 返回转换结果
func (w *EventWalker) Result() (string, []MessageEntity, []Segment) {
	return w.buf.String(), w.entities, w.segments
}

// --- Text handling ---

func (w *EventWalker) onText(seg text.Segment, softBreak bool, hardBreak bool) {
	textContent := string(seg.Value(w.source))
	
	if softBreak {
		textContent += "\n"
	}
	if hardBreak {
		textContent += "\n"
	}
	
	if w.inCodeBlock {
		w.codeBlockParts = append(w.codeBlockParts, textContent)
		return
	}
	if w.inTableCell {
		// Table cells: soft breaks become spaces
		if softBreak {
			w.cellParts = append(w.cellParts, textContent[:len(textContent)-1], " ")
		} else {
			w.cellParts = append(w.cellParts, textContent)
		}
		return
	}
	
	w.buf.Write(textContent)
}

func (w *EventWalker) onTextString(value []byte) {
	textContent := string(value)
	if w.inCodeBlock {
		w.codeBlockParts = append(w.codeBlockParts, textContent)
		return
	}
	if w.inTableCell {
		w.cellParts = append(w.cellParts, textContent)
		return
	}
	w.buf.Write(textContent)
}

func (w *EventWalker) onInlineCode(n *ast.CodeSpan) {
	code := extractCodeSpanText(n, w.source)
	if w.inTableCell {
		w.cellParts = append(w.cellParts, code)
		return
	}
	start := w.buf.UTF16Offset()
	w.buf.Write(code)
	length := w.buf.UTF16Offset() - start
	if length > 0 {
		w.entities = append(w.entities, MessageEntity{
			Type:   "code",
			Offset: start,
			Length: length,
		})
	}
}

func (w *EventWalker) onInlineHTML(n *ast.RawHTML) {
	html := string(n.Segments.Value(w.source))
	tag := strings.TrimSpace(strings.ToLower(html))
	
	if tag == "<tg-spoiler>" {
		w.pushEntity("spoiler", "")
	} else if tag == "</tg-spoiler>" {
		w.popEntity("spoiler")
	}
	// Other inline HTML is ignored
}

func (w *EventWalker) onRule() {
	w.ensureBlockSpacing()
	w.buf.Write("————————")
	w.blockCount++
}

// --- Paragraph ---

func (w *EventWalker) onStartParagraph() {
	if len(w.listStack) == 0 {
		w.ensureBlockSpacing()
	}
}

func (w *EventWalker) onEndParagraph() {
	if len(w.listStack) == 0 {
		w.blockCount++
	} else if w.buf.TrailingNewlineCount() == 0 {
		// loose list 中段落结束时写入换行，避免多段落粘连
		w.buf.Write("\n")
	}
}

// --- Heading ---

var headingEntitiesMap = map[int][]string{
	1: {"bold", "underline"},
	2: {"bold", "underline"},
	3: {"bold"},
	4: {"bold"},
	5: {"italic"},
	6: {"italic"},
}

func (w *EventWalker) onStartHeading(n *ast.Heading) {
	w.ensureBlockSpacing()
	
	// 获取标题符号
	var symbol string
	switch n.Level {
	case 1:
		symbol = w.config.MarkdownSymbol.HeadingLevel1
	case 2:
		symbol = w.config.MarkdownSymbol.HeadingLevel2
	case 3:
		symbol = w.config.MarkdownSymbol.HeadingLevel3
	case 4:
		symbol = w.config.MarkdownSymbol.HeadingLevel4
	case 5:
		symbol = w.config.MarkdownSymbol.HeadingLevel5
	case 6:
		symbol = w.config.MarkdownSymbol.HeadingLevel6
	}
	
	if symbol != "" {
		w.buf.Write(symbol + " ")
	}
	
	// 推送标题实体
	w.headingEntities = headingEntitiesMap[n.Level]
	if w.headingEntities == nil {
		w.headingEntities = []string{"bold"}
	}
	
	for _, etype := range w.headingEntities {
		w.pushEntity(etype, "")
	}
	w.inHeading = true
}

func (w *EventWalker) onEndHeading() {
	// 反向弹出
	for i := len(w.headingEntities) - 1; i >= 0; i-- {
		w.popEntity(w.headingEntities[i])
	}
	w.headingEntities = nil
	w.inHeading = false
	w.blockCount++
}

// --- Code block ---

func (w *EventWalker) onStartCodeBlock(n ast.Node) {
	w.inCodeBlock = true
	w.codeBlockParts = make([]string, 0)
	
	if fenced, ok := n.(*ast.FencedCodeBlock); ok {
		w.codeBlockLang = string(fenced.Language(w.source))
	} else {
		w.codeBlockLang = ""
	}
	
	// 提取代码块内容
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		w.codeBlockParts = append(w.codeBlockParts, string(line.Value(w.source)))
	}
	
	w.onEndCodeBlock()
}

func (w *EventWalker) onEndCodeBlock() {
	w.inCodeBlock = false
	rawCode := strings.Join(w.codeBlockParts, "")
	
	// Strip single trailing newline
	if strings.HasSuffix(rawCode, "\n") {
		rawCode = rawCode[:len(rawCode)-1]
	}
	
	w.ensureBlockSpacing()
	
	// Record segment
	segTextStart := w.buf.ByteOffset()
	segUTF16Start := w.buf.UTF16Offset()
	
	start := w.buf.UTF16Offset()
	w.buf.Write(rawCode)
	length := w.buf.UTF16Offset() - start
	
	lang := strings.Split(w.codeBlockLang, ",")[0]
	lang = strings.TrimSpace(lang)
	
	if length > 0 {
		entity := MessageEntity{
			Type:   "pre",
			Offset: start,
			Length: length,
		}
		if lang != "" {
			entity.Language = lang
		}
		w.entities = append(w.entities, entity)
	}
	
	// Determine segment kind
	segKind := "code_block"
	if strings.ToLower(lang) == "mermaid" {
		segKind = "mermaid"
	}
	
	w.segments = append(w.segments, Segment{
		Kind:       segKind,
		TextStart:  segTextStart,
		TextEnd:    w.buf.ByteOffset(),
		UTF16Start: segUTF16Start,
		UTF16End:   w.buf.UTF16Offset(),
		Language:   lang,
		RawCode:    rawCode,
	})
	
	w.blockCount++
	w.codeBlockLang = ""
	w.codeBlockParts = nil
}

// --- Blockquote ---

func (w *EventWalker) onStartBlockquote() {
	w.ensureBlockSpacing()
	scope := EntityScope{
		EntityType:  "blockquote",
		StartOffset: w.buf.UTF16Offset(),
	}
	w.blockquoteScopes = append(w.blockquoteScopes, scope)
}

func (w *EventWalker) onEndBlockquote() {
	if len(w.blockquoteScopes) > 0 {
		scope := w.blockquoteScopes[len(w.blockquoteScopes)-1]
		w.blockquoteScopes = w.blockquoteScopes[:len(w.blockquoteScopes)-1]
		
		length := w.buf.UTF16Offset() - scope.StartOffset
		if length > 0 {
			w.entities = append(w.entities, MessageEntity{
				Type:   "blockquote",
				Offset: scope.StartOffset,
				Length: length,
			})
		}
	}
	w.blockCount++
}

// --- Links & Images ---

func (w *EventWalker) onStartLink(n *ast.Link) {
	destURL := string(n.Destination)
	emojiID := validateTelegramEmoji(destURL)
	
	if emojiID != "" {
		w.pushEntity("custom_emoji", emojiID)
	} else if destURL != "" {
		w.pushEntity("text_link", destURL)
	}
	// Empty URL links are rendered as plain text (no entity)
}

func (w *EventWalker) onStartImage(n *ast.Image) {
	destURL := string(n.Destination)
	emojiID := validateTelegramEmoji(destURL)
	
	if emojiID != "" {
		w.pushEntity("custom_emoji", emojiID)
	} else {
		w.buf.Write(w.config.MarkdownSymbol.Image)
		w.pushEntity("text_link", destURL)
	}
}

// --- Lists ---

func (w *EventWalker) onStartList(n *ast.List) {
	if len(w.listStack) == 0 {
		w.ensureBlockSpacing()
	}
	
	if n.IsOrdered() {
		start := n.Start
		w.listStack = append(w.listStack, &start)
	} else {
		w.listStack = append(w.listStack, nil)
	}
}

func (w *EventWalker) onStartItem() {
	depth := len(w.listStack)
	indent := strings.Repeat("  ", depth-1)
	
	// 嵌套列表：父项文本后没有换行时，插入换行确保子项独占一行
	if w.buf.ByteOffset() > 0 && w.buf.TrailingNewlineCount() == 0 {
		w.buf.Write("\n")
	}
	
	w.itemIndent = indent
	
	if len(w.listStack) > 0 {
		currentList := w.listStack[len(w.listStack)-1]
		if currentList != nil {
			// Ordered list
			num := *(currentList.(*int))
			w.buf.Write(fmt.Sprintf("%s%d. ", indent, num))
			*(currentList.(*int)) = num + 1
		} else {
			// Unordered list - 先写 bullet，如果后面遇到 TaskCheckBox 会被替换
			w.buf.Write(fmt.Sprintf("%s⦁ ", indent))
		}
	}
	
	w.itemStarted = true
}

func (w *EventWalker) onEndItem() {
	if w.buf.TrailingNewlineCount() == 0 {
		w.buf.Write("\n")
	}
	w.itemStarted = false
}

// onTaskCheckBox 处理任务列表复选框
// 对应 Python 的 _on_task_list_marker
func (w *EventWalker) onTaskCheckBox(checked bool) {
	// 移除 _on_start_item 刚写入的 bullet 前缀（"⦁ " 或缩进+⦁）
	w.buf.PopLast()
	
	// 写入任务标记
	symbol := w.config.MarkdownSymbol.TaskUncompleted
	if checked {
		symbol = w.config.MarkdownSymbol.TaskCompleted
	}
	w.buf.Write(fmt.Sprintf("%s%s ", w.itemIndent, symbol))
}

func (w *EventWalker) onEndList() {
	if len(w.listStack) > 0 {
		w.listStack = w.listStack[:len(w.listStack)-1]
	}
	if len(w.listStack) == 0 {
		w.blockCount++
	}
}

// --- Tables ---

func (w *EventWalker) onStartTable(n *east.Table) {
	w.ensureBlockSpacing()
	w.inTable = true
	w.tableAlignments = n.Alignments
	w.tableRows = make([][]string, 0)
}

func (w *EventWalker) onEndTableCell() {
	cellText := strings.Join(w.cellParts, "")
	w.currentRow = append(w.currentRow, cellText)
	w.cellParts = nil
	w.inTableCell = false
}

func (w *EventWalker) onEndTableRow() {
	w.tableRows = append(w.tableRows, w.currentRow)
	w.currentRow = nil
}

func (w *EventWalker) onEndTable() {
	w.inTable = false
	tableText := w.formatTable(w.tableRows)
	
	start := w.buf.UTF16Offset()
	w.buf.Write(tableText)
	length := w.buf.UTF16Offset() - start
	
	if length > 0 {
		w.entities = append(w.entities, MessageEntity{
			Type:   "pre",
			Offset: start,
			Length: length,
		})
	}
	
	w.tableRows = nil
	w.blockCount++
}

func (w *EventWalker) formatTable(rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}
	
	// Compute column widths
	numCols := 0
	for _, row := range rows {
		if len(row) > numCols {
			numCols = len(row)
		}
	}
	
	colWidths := make([]int, numCols)
	for _, row := range rows {
		for i, cell := range row {
			if i < numCols && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}
	
	var lines []string
	for rowIdx, row := range rows {
		cells := make([]string, numCols)
		for i := 0; i < numCols; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			// Left-justify
			cells[i] = cell + strings.Repeat(" ", colWidths[i]-len(cell))
		}
		lines = append(lines, strings.Join(cells, " | "))
		
		// Add separator after header
		if rowIdx == 0 && len(rows) > 1 {
			sepCells := make([]string, numCols)
			for i := 0; i < numCols; i++ {
				sepCells[i] = strings.Repeat("-", colWidths[i])
			}
			lines = append(lines, strings.Join(sepCells, "-+-"))
		}
	}
	
	return strings.Join(lines, "\n")
}

// --- Entity helpers ---

func (w *EventWalker) pushEntity(entityType string, urlOrEmojiID string) {
	scope := EntityScope{
		EntityType:  entityType,
		StartOffset: w.buf.UTF16Offset(),
	}
	
	if entityType == "text_link" {
		scope.URL = urlOrEmojiID
	} else if entityType == "custom_emoji" {
		scope.CustomEmojiID = urlOrEmojiID
	}
	
	w.entityStack = append(w.entityStack, scope)
}

func (w *EventWalker) popEntity(entityType string) {
	// Find the matching scope (search from top)
	for i := len(w.entityStack) - 1; i >= 0; i-- {
		if w.entityStack[i].EntityType == entityType {
			scope := w.entityStack[i]
			w.entityStack = append(w.entityStack[:i], w.entityStack[i+1:]...)
			w.finalizeEntity(scope)
			return
		}
	}
}

func (w *EventWalker) popEntityAny() {
	if len(w.entityStack) > 0 {
		scope := w.entityStack[len(w.entityStack)-1]
		w.entityStack = w.entityStack[:len(w.entityStack)-1]
		w.finalizeEntity(scope)
	}
}

func (w *EventWalker) finalizeEntity(scope EntityScope) {
	length := w.buf.UTF16Offset() - scope.StartOffset
	if length <= 0 {
		return
	}
	
	entity := MessageEntity{
		Type:   scope.EntityType,
		Offset: scope.StartOffset,
		Length: length,
	}
	
	if scope.URL != "" {
		entity.URL = scope.URL
	}
	if scope.Language != "" {
		entity.Language = scope.Language
	}
	if scope.CustomEmojiID != "" {
		entity.CustomEmojiID = scope.CustomEmojiID
	}
	
	w.entities = append(w.entities, entity)
}

func (w *EventWalker) ensureBlockSpacing() {
	// Ensure a blank line (\n\n) between blocks, avoiding excess newlines
	if w.blockCount > 0 {
		trailing := w.buf.TrailingNewlineCount()
		needed := 2 - trailing
		if needed > 0 {
			w.buf.Write(strings.Repeat("\n", needed))
		}
	}
}

// --- Utilities ---

func extractCodeSpanText(n *ast.CodeSpan, source []byte) string {
	var buf strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if textNode, ok := c.(*ast.Text); ok {
			_, _ = buf.Write(textNode.Segment.Value(source))
		}
	}
	return buf.String()
}

var _ io.Writer = (*EventWalker)(nil)

func (w *EventWalker) Write(p []byte) (n int, err error) {
	// Implement io.Writer for compatibility
	w.buf.Write(string(p))
	return len(p), nil
}

