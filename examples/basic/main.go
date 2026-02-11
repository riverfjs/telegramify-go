package main

import (
	"fmt"
	tg "github.com/riverfjs/telegramify-go"
)

func main() {
	// 示例 Markdown 文本
	markdown := `# Hello Telegram!

这是一个 **粗体** 和 *斜体* 的示例。

## 功能演示

- 无序列表项 1
- 无序列表项 2

1. 有序列表项 1
2. 有序列表项 2

### 任务列表
- [x] 已完成的任务
- [ ] 未完成的任务

### 代码示例

这是行内代码：` + "`print('Hello')`" + `

代码块：
` + "```python\n" + `def hello():
    print("Hello, Telegram!")
    return True
` + "```" + `

### 引用

> 这是一个引用文本
> 可以有多行

### 链接和强调

访问 [Google](https://google.com) 或使用 ~~删除线~~ 和 ||剧透文本||。

---

**粗体 *斜体嵌套* 粗体**
`

	fmt.Println("=== 基础转换示例 ===\n")
	
	// 转换为纯文本 + entities
	text, entities := tg.Convert(markdown, true, nil)
	
	fmt.Printf("纯文本长度: %d 字符\n", len(text))
	fmt.Printf("UTF-16 长度: %d code units\n", tg.UTF16Len(text))
	fmt.Printf("实体数量: %d\n\n", len(entities))
	
	fmt.Println("前 500 个字符:")
	if len(text) > 500 {
		fmt.Println(text[:500] + "...")
	} else {
		fmt.Println(text)
	}
	
	fmt.Println("\n实体列表:")
	for i, entity := range entities {
		if i >= 10 {
			fmt.Printf("... 还有 %d 个实体\n", len(entities)-10)
			break
		}
		fmt.Printf("  %d. Type: %-20s Offset: %-4d Length: %-4d", 
			i+1, entity.Type, entity.Offset, entity.Length)
		if entity.URL != "" {
			fmt.Printf(" URL: %s", entity.URL)
		}
		if entity.Language != "" {
			fmt.Printf(" Lang: %s", entity.Language)
		}
		fmt.Println()
	}
	
	fmt.Println("\n=== 自定义配置示例 ===\n")
	
	// 使用自定义配置
	config := tg.DefaultConfig()
	config.MarkdownSymbol.HeadingLevel1 = "🌟"
	config.MarkdownSymbol.TaskCompleted = "✓"
	config.MarkdownSymbol.TaskUncompleted = "☐"
	
	text2, _ := tg.Convert("# 自定义标题\n\n- [x] 完成\n- [ ] 待办", false, config)
	fmt.Println("自定义配置输出:")
	fmt.Println(text2)
}

