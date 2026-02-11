package main

import (
	"fmt"
	"strings"
	tg "github.com/riverfjs/telegramify-go"
)

func main() {
	// 生成一个很长的 Markdown 文本
	var sb strings.Builder
	sb.WriteString("# 长文档拆分示例\n\n")
	
	for i := 1; i <= 50; i++ {
		sb.WriteString(fmt.Sprintf("## 第 %d 节\n\n", i))
		sb.WriteString(fmt.Sprintf("这是第 %d 节的内容。", i))
		sb.WriteString("包含一些 **重要** 的信息和 *注释*。\n\n")
		
		if i%10 == 0 {
			sb.WriteString("```python\n")
			sb.WriteString(fmt.Sprintf("# 代码示例 %d\n", i))
			sb.WriteString("def process():\n")
			sb.WriteString("    return True\n")
			sb.WriteString("```\n\n")
		}
	}
	
	markdown := sb.String()
	
	fmt.Println("=== 消息拆分示例 ===\n")
	fmt.Printf("原始 Markdown 长度: %d 字符\n\n", len(markdown))
	
	// 先转换
	text, entities := tg.Convert(markdown, false, nil)
	fmt.Printf("转换后文本长度: %d 字符\n", len(text))
	fmt.Printf("转换后 UTF-16 长度: %d code units\n", tg.UTF16Len(text))
	fmt.Printf("实体数量: %d\n\n", len(entities))
	
	// 拆分为符合 Telegram 限制的块（4096 UTF-16 code units）
	chunks := tg.SplitEntities(text, entities, 4096)
	
	fmt.Printf("拆分结果: %d 个消息块\n\n", len(chunks))
	
	for i, chunk := range chunks {
		utf16Len := tg.UTF16Len(chunk.Text)
		fmt.Printf("块 %d:\n", i+1)
		fmt.Printf("  文本长度: %d 字符\n", len(chunk.Text))
		fmt.Printf("  UTF-16 长度: %d code units\n", utf16Len)
		fmt.Printf("  实体数量: %d\n", len(chunk.Entities))
		
		// 显示前 100 个字符
		preview := chunk.Text
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		fmt.Printf("  预览: %s\n\n", preview)
	}
}

