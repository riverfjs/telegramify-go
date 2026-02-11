# telegramify-go

[English](README.md) | 简体中文

将 Markdown 转换为 Telegram Bot API 所需的纯文本 + MessageEntity 格式。

## 特性

- ✅ **完整 Markdown 支持**：标题、列表、表格、代码块、引用等
- ✅ **LaTeX 转 Unicode**：自动将 LaTeX 数学公式转换为 Unicode 符号
- ✅ **智能消息拆分**：按 UTF-16 长度智能拆分长消息
- ✅ **代码块提取**：自动提取代码块为文件
- ✅ **Mermaid 渲染**：支持 Mermaid 图表渲染为图片
- ✅ **零依赖核心**：核心转换功能无外部依赖（Mermaid 渲染除外）

## 安装

```bash
go get github.com/riverfjs/telegramify-go
```

## 快速开始

### 基础转换

```go
package main

import (
    "fmt"
    tg "github.com/riverfjs/telegramify-go"
)

func main() {
    markdown := `# Hello World

This is **bold** and *italic* text.

\`\`\`python
print("Hello, Telegram!")
\`\`\`
`
    
    // 转换为纯文本 + entities
    text, entities := tg.Convert(markdown, true, nil)
    
    fmt.Println("Text:", text)
    fmt.Println("Entities:", len(entities))
}
```

### 完整处理（含拆分和文件提取）

```go
package main

import (
    "context"
    "fmt"
    tg "github.com/riverfjs/telegramify-go"
)

func main() {
    markdown := `# 长文档示例

这是一个很长的文档...

\`\`\`go
func main() {
    fmt.Println("代码会被提取为文件")
}
\`\`\`
`
    
    ctx := context.Background()
    contents, err := tg.Telegramify(ctx, markdown, 4096, true, nil)
    if err != nil {
        panic(err)
    }
    
    for _, content := range contents {
        switch c := content.(type) {
        case *tg.Text:
            fmt.Printf("文本消息: %d 字符\n", len(c.Text))
        case *tg.File:
            fmt.Printf("文件: %s (%d 字节)\n", c.FileName, len(c.FileData))
        case *tg.Photo:
            fmt.Printf("图片: %s\n", c.FileName)
        }
    }
}
```

## API 参考

### Convert

```go
func Convert(markdown string, latexEscape bool, config *RenderConfig) (string, []MessageEntity)
```

将 Markdown 转换为 (纯文本, entities)。

**参数：**
- `markdown`: 原始 Markdown 文本
- `latexEscape`: 是否将 LaTeX 转换为 Unicode
- `config`: 渲染配置，nil 使用默认配置

**返回：**
- `string`: 纯文本
- `[]MessageEntity`: 实体列表

### Telegramify

```go
func Telegramify(ctx context.Context, content string, maxMessageLength int, latexEscape bool, config *RenderConfig) ([]Content, error)
```

完整处理管道：转换、拆分、文件提取、Mermaid 渲染。

**参数：**
- `ctx`: 上下文
- `content`: 原始 Markdown 文本
- `maxMessageLength`: 每条消息最大 UTF-16 长度（Telegram 限制 4096）
- `latexEscape`: 是否将 LaTeX 转换为 Unicode
- `config`: 渲染配置

**返回：**
- `[]Content`: Text、File 或 Photo 对象列表

### 配置

```go
type RenderConfig struct {
    MarkdownSymbol *Symbol
    CiteExpandable bool
}

type Symbol struct {
    HeadingLevel1   string  // 默认: 📌
    HeadingLevel2   string  // 默认: 📝
    HeadingLevel3   string  // 默认: 📋
    HeadingLevel4   string  // 默认: 📄
    HeadingLevel5   string  // 默认: 📃
    HeadingLevel6   string  // 默认: 🔖
    Quote           string  // 默认: 💬
    Image           string  // 默认: 🖼
    TaskCompleted   string  // 默认: ✅
    TaskUncompleted string  // 默认: ☑️
}
```

## 支持的 Markdown 特性

- **标题**：H1-H6，带自定义前缀符号
- **强调**：**粗体**、*斜体*、~~删除线~~
- **列表**：有序列表、无序列表、任务列表
- **代码**：行内代码、代码块（带语言标识）
- **引用**：单行和多行引用
- **链接**：[文本](URL)
- **图片**：![alt](URL)
- **表格**：GitHub 风格表格
- **数学公式**：LaTeX 转 Unicode
- **自定义 Emoji**：`tg://emoji?id=...`
- **剧透**：||隐藏文本||

## UTF-16 计算

Telegram 要求 entity 的 offset 和 length 以 UTF-16 code units 计算。本库自动处理：

```go
text := "Hello 世界 🌍"
length := tg.UTF16Len(text)  // 10 (不是 9 个 runes)
```

## 项目结构

```
telegramify-go/
├── entity.go              # MessageEntity 和 UTF-16 工具
├── content.go             # 输出类型定义
├── config.go              # 配置系统
├── converter.go           # 转换器公开 API
├── pipeline.go            # 处理管道
├── telegramify.go         # 主入口
├── internal/
│   ├── types/            # 共享类型定义
│   ├── buffer/           # 文本缓冲
│   ├── converter/        # 核心转换器
│   │   ├── walker.go    # AST 遍历器
│   │   ├── preprocess.go # 预处理
│   │   └── segment.go   # 片段定义
│   ├── parser/           # Markdown 解析器
│   ├── latex/            # LaTeX 转 Unicode
│   │   ├── symbols.go   # 符号表
│   │   ├── parser.go    # 递归下降解析器
│   │   └── latex.go     # 公开接口
│   ├── mermaid/          # Mermaid 渲染
│   └── util/             # 工具函数
└── go.mod
```

## 依赖

- **核心**: [goldmark](https://github.com/yuin/goldmark) - Markdown 解析器
- **可选**: 无（Mermaid 渲染使用标准库 HTTP 客户端）

## 与 Python 版本的差异

1. **类型系统**：Go 的强类型系统提供更好的类型安全
2. **并发**：Go 的 goroutine 支持更高效的并发处理
3. **性能**：编译型语言，性能更优
4. **依赖**：核心功能零外部依赖（Python 版依赖 pyromark）

## 开发

```bash
# 克隆仓库
git clone https://github.com/riverfjs/telegramify-go.git
cd telegramify-go

# 构建
go build ./...

# 测试
go test ./...

# 运行示例
go run examples/basic/main.go
```

## 许可证

MIT License

## 致谢

本库的灵感来源于 [npm:telegramify-markdown](https://www.npmjs.com/package/telegramify-markdown)。

