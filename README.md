# telegramify-go

English | [简体中文](README.zh-CN.md)

Convert Markdown to plain text + MessageEntity format required by Telegram Bot API.

## Features

- ✅ **Full Markdown Support**: Headings, lists, tables, code blocks, quotes, and more
- ✅ **LaTeX to Unicode**: Automatically converts LaTeX math formulas to Unicode symbols
- ✅ **Smart Message Splitting**: Intelligently splits long messages by UTF-16 length
- ✅ **Code Block Extraction**: Automatically extracts code blocks as files
- ✅ **Mermaid Rendering**: Supports rendering Mermaid diagrams as images
- ✅ **Zero Dependencies Core**: Core conversion has no external dependencies (except Mermaid rendering)

## Installation

```bash
go get github.com/riverfjs/telegramify-go
```

## Quick Start

### Basic Conversion

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
    
    // Convert to plain text + entities
    text, entities := tg.Convert(markdown, true, nil)
    
    fmt.Println("Text:", text)
    fmt.Println("Entities:", len(entities))
}
```

### Full Processing (with splitting and file extraction)

```go
package main

import (
    "context"
    "fmt"
    tg "github.com/riverfjs/telegramify-go"
)

func main() {
    markdown := `# Long Document Example

This is a very long document...

\`\`\`go
func main() {
    fmt.Println("Code will be extracted as files")
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
            fmt.Printf("Text message: %d characters\n", len(c.Text))
        case *tg.File:
            fmt.Printf("File: %s (%d bytes)\n", c.FileName, len(c.FileData))
        case *tg.Photo:
            fmt.Printf("Photo: %s\n", c.FileName)
        }
    }
}
```

## API Reference

### Convert

```go
func Convert(markdown string, latexEscape bool, config *RenderConfig) (string, []MessageEntity)
```

Converts Markdown to (plain text, entities).

**Parameters:**
- `markdown`: Raw Markdown text
- `latexEscape`: Whether to convert LaTeX to Unicode
- `config`: Render configuration, nil uses default config

**Returns:**
- `string`: Plain text
- `[]MessageEntity`: Entity list

### Telegramify

```go
func Telegramify(ctx context.Context, content string, maxMessageLength int, latexEscape bool, config *RenderConfig) ([]Content, error)
```

Full processing pipeline: conversion, splitting, file extraction, Mermaid rendering.

**Parameters:**
- `ctx`: Context
- `content`: Raw Markdown text
- `maxMessageLength`: Maximum UTF-16 length per message (Telegram limit is 4096)
- `latexEscape`: Whether to convert LaTeX to Unicode
- `config`: Render configuration

**Returns:**
- `[]Content`: List of Text, File, or Photo objects

### Configuration

```go
type RenderConfig struct {
    MarkdownSymbol *Symbol
    CiteExpandable bool
}

type Symbol struct {
    HeadingLevel1   string  // Default: 📌
    HeadingLevel2   string  // Default: 📝
    HeadingLevel3   string  // Default: 📋
    HeadingLevel4   string  // Default: 📄
    HeadingLevel5   string  // Default: 📃
    HeadingLevel6   string  // Default: 🔖
    Quote           string  // Default: 💬
    Image           string  // Default: 🖼
    TaskCompleted   string  // Default: ✅
    TaskUncompleted string  // Default: ☑️
}
```

## Supported Markdown Features

- **Headings**: H1-H6, with custom prefix symbols
- **Emphasis**: **bold**, *italic*, ~~strikethrough~~
- **Lists**: Ordered lists, unordered lists, task lists
- **Code**: Inline code, code blocks (with language identifiers)
- **Quotes**: Single-line and multi-line quotes
- **Links**: [text](URL)
- **Images**: ![alt](URL)
- **Tables**: GitHub-flavored tables
- **Math**: LaTeX to Unicode conversion
- **Custom Emoji**: `tg://emoji?id=...`
- **Spoilers**: ||hidden text||

## UTF-16 Calculation

Telegram requires entity offsets and lengths to be calculated in UTF-16 code units. This library handles it automatically:

```go
text := "Hello 世界 🌍"
length := tg.UTF16Len(text)  // 10 (not 9 runes)
```

## Project Structure

```
telegramify-go/
├── entity.go              # MessageEntity and UTF-16 utilities
├── content.go             # Output type definitions
├── config.go              # Configuration system
├── converter.go           # Converter public API
├── pipeline.go            # Processing pipeline
├── telegramify.go         # Main entry point
├── internal/
│   ├── types/            # Shared type definitions
│   ├── buffer/           # Text buffer
│   ├── converter/        # Core converter
│   │   ├── walker.go    # AST walker
│   │   ├── preprocess.go # Preprocessing
│   │   └── segment.go   # Segment definitions
│   ├── parser/           # Markdown parser
│   ├── latex/            # LaTeX to Unicode
│   │   ├── symbols.go   # Symbol table
│   │   ├── parser.go    # Recursive descent parser
│   │   └── latex.go     # Public interface
│   ├── mermaid/          # Mermaid rendering
│   └── util/             # Utility functions
└── go.mod
```

## Dependencies

- **Core**: [goldmark](https://github.com/yuin/goldmark) - Markdown parser
- **Optional**: None (Mermaid rendering uses standard library HTTP client)

## Differences from Python Version

1. **Type System**: Go's strong type system provides better type safety
2. **Concurrency**: Go's goroutines support more efficient concurrent processing
3. **Performance**: Compiled language, better performance
4. **Dependencies**: Core functionality has zero external dependencies (Python version depends on pyromark)

## Development

```bash
# Clone repository
git clone https://github.com/riverfjs/telegramify-go.git
cd telegramify-go

# Build
go build ./...

# Test
go test ./...

# Run examples
go run examples/basic/main.go
```

## License

MIT License

## Acknowledgement

This library is inspired by [npm:telegramify-markdown](https://www.npmjs.com/package/telegramify-markdown).
