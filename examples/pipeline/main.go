package main

import (
	"context"
	"fmt"
	tg "github.com/riverfjs/telegramify-go"
)

func main() {
	// 包含代码块和 Mermaid 图的 Markdown
	markdown := `# 项目文档

## 架构概述

这是我们的系统架构。

## 代码示例

这是一个 Python 示例：

` + "```python\n" + `import requests

def fetch_data(url):
    """从 API 获取数据"""
    response = requests.get(url)
    return response.json()

if __name__ == "__main__":
    data = fetch_data("https://api.example.com/data")
    print(f"获取到 {len(data)} 条记录")
` + "```" + `

这是一个 Go 示例：

` + "```go\n" + `package main

import (
    "fmt"
    "net/http"
)

func main() {
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    fmt.Println("请求成功")
}
` + "```" + `

## 流程图

` + "```mermaid\n" + `graph TD
    A[开始] --> B{检查条件}
    B -->|是| C[执行操作]
    B -->|否| D[跳过]
    C --> E[结束]
    D --> E
` + "```" + `

## 总结

以上就是完整的示例。
`

	fmt.Println("=== 完整处理管道示例 ===\n")
	
	ctx := context.Background()
	
	// 使用 Telegramify 完整处理（转换、拆分、提取文件）
	// maxMessageLength: 4096 是 Telegram 的限制
	contents, err := tg.Telegramify(ctx, markdown, 4096, false, nil)
	if err != nil {
		fmt.Printf("处理失败: %v\n", err)
		return
	}
	
	fmt.Printf("处理结果: %d 个内容项\n\n", len(contents))
	
	for i, content := range contents {
		switch c := content.(type) {
		case *tg.Text:
			utf16Len := tg.UTF16Len(c.Text)
			fmt.Printf("%d. 文本消息\n", i+1)
			fmt.Printf("   长度: %d 字符 (%d UTF-16 code units)\n", len(c.Text), utf16Len)
			fmt.Printf("   实体数: %d\n", len(c.Entities))
			fmt.Printf("   来源: %s\n", c.ContentTrace.SourceType)
			
			// 显示前 150 个字符
			preview := c.Text
			if len(preview) > 150 {
				preview = preview[:150] + "..."
			}
			fmt.Printf("   预览: %s\n\n", preview)
			
		case *tg.File:
			fmt.Printf("%d. 文件附件\n", i+1)
			fmt.Printf("   文件名: %s\n", c.FileName)
			fmt.Printf("   大小: %d 字节\n", len(c.FileData))
			if c.CaptionText != "" {
				fmt.Printf("   标题: %s\n", c.CaptionText)
			}
			fmt.Printf("   来源: %s\n", c.ContentTrace.SourceType)
			
			// 显示文件内容的前 100 字节
			preview := string(c.FileData)
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			fmt.Printf("   内容预览:\n%s\n\n", preview)
			
		case *tg.Photo:
			fmt.Printf("%d. 图片\n", i+1)
			fmt.Printf("   文件名: %s\n", c.FileName)
			fmt.Printf("   大小: %d 字节\n", len(c.FileData))
			if c.Caption != "" {
				fmt.Printf("   标题: %s\n", c.Caption)
			}
			fmt.Printf("   来源: %s\n\n", c.ContentTrace.SourceType)
		}
	}
}

