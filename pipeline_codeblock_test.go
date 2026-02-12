package telegramify

import (
	"context"
	"strings"
	"testing"
)

// TestSmallCodeBlocksNotSplit 测试小代码块（≤50行）不应分割消息
func TestSmallCodeBlocksNotSplit(t *testing.T) {
	markdown := `完成！配置文件和备份都已就绪。

## 已完成的工作

1. 配置文件已创建: ~/.myclaw/workspace/.claude/skills/todo/config.json

` + "```json" + `
{
  "channel": "telegram",
  "chat_id": "5821086579"
}
` + "```" + `

2. 原文件已备份: scripts/main.go.backup
3. Cron 数据结构已详细说明 (见前面的说明)`

	ctx := context.Background()
	contents, err := ProcessMarkdown(ctx, markdown, 4096, false, nil)
	if err != nil {
		t.Fatalf("ProcessMarkdown failed: %v", err)
	}

	// 统计消息数量
	textCount := 0
	fileCount := 0
	for _, c := range contents {
		switch c.(type) {
		case *Text:
			textCount++
		case *File:
			fileCount++
		}
	}

	t.Logf("Total contents: %d (text: %d, file: %d)", len(contents), textCount, fileCount)
	
	// 小代码块（4行）不应提取为文件，应该保留在文本中
	// 期望：1条文本消息，0个文件
	if fileCount != 0 {
		t.Errorf("Expected 0 files, got %d", fileCount)
	}
	
	if textCount != 1 {
		t.Errorf("Expected 1 text message, got %d", textCount)
		for i, c := range contents {
			switch v := c.(type) {
			case *Text:
				t.Logf("Text[%d]: %d chars, %d entities", i, len(v.Text), len(v.Entities))
				t.Logf("Content: %s", v.Text[:min(100, len(v.Text))])
			case *File:
				t.Logf("File[%d]: %s", i, v.FileName)
			}
		}
	}
	
	// 验证文本中包含代码块内容
	if textCount > 0 {
		text := contents[0].(*Text)
		if !strings.Contains(text.Text, `"channel": "telegram"`) {
			t.Error("Code block content not found in text")
		}
	}
}

// TestLargeCodeBlockExtracted 测试大代码块（>50行）应该提取为文件
func TestLargeCodeBlockExtracted(t *testing.T) {
	// 生成一个超过50行的代码块
	largeCode := "package main\n\nfunc main() {\n"
	for i := 0; i < 60; i++ {
		largeCode += "\t// line " + string(rune('0'+i%10)) + "\n"
	}
	largeCode += "}\n"

	markdown := "这是一个大代码块：\n\n```go\n" + largeCode + "```\n\n后续文本"

	ctx := context.Background()
	contents, err := ProcessMarkdown(ctx, markdown, 4096, false, nil)
	if err != nil {
		t.Fatalf("ProcessMarkdown failed: %v", err)
	}

	// 统计消息数量
	textCount := 0
	fileCount := 0
	for _, c := range contents {
		switch c.(type) {
		case *Text:
			textCount++
		case *File:
			fileCount++
		}
	}

	t.Logf("Total contents: %d (text: %d, file: %d)", len(contents), textCount, fileCount)

	// 期望：2条文本消息（前后文本）+ 1个文件（大代码块）
	if fileCount != 1 {
		t.Errorf("Expected 1 file, got %d", fileCount)
	}
	
	if textCount != 2 {
		t.Errorf("Expected 2 text messages, got %d", textCount)
	}
}

// TestMultipleSmallCodeBlocks 测试多个小代码块
func TestMultipleSmallCodeBlocks(t *testing.T) {
	markdown := `文本1

` + "```json" + `
{"key": "value1"}
` + "```" + `

文本2

` + "```json" + `
{"key": "value2"}
` + "```" + `

文本3`

	ctx := context.Background()
	contents, err := ProcessMarkdown(ctx, markdown, 4096, false, nil)
	if err != nil {
		t.Fatalf("ProcessMarkdown failed: %v", err)
	}

	// 统计消息数量
	textCount := 0
	fileCount := 0
	for _, c := range contents {
		switch c.(type) {
		case *Text:
			textCount++
		case *File:
			fileCount++
		}
	}

	t.Logf("Total contents: %d (text: %d, file: %d)", len(contents), textCount, fileCount)

	// 两个小代码块都应该保留在文本中
	// 期望：1条文本消息，0个文件
	if fileCount != 0 {
		t.Errorf("Expected 0 files, got %d", fileCount)
	}
	
	if textCount != 1 {
		t.Errorf("Expected 1 text message (all content in one), got %d", textCount)
		for i, c := range contents {
			switch v := c.(type) {
			case *Text:
				t.Logf("Text[%d]: %d chars", i, len(v.Text))
			case *File:
				t.Logf("File[%d]: %s", i, v.FileName)
			}
		}
	}
}

