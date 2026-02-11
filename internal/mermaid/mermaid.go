package mermaid

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"time"

	_ "golang.org/x/image/webp"
)

// Config Mermaid 配置
type Config struct {
	Theme string `json:"theme"`
}

// DefaultConfig 返回默认 Mermaid 配置
func DefaultConfig() *Config {
	return &Config{
		Theme: "default",
	}
}

// compressToDeflate 使用 DEFLATE 算法压缩数据
func compressToDeflate(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	
	if err := writer.Close(); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// safeBase64Encode URL-safe base64 编码
func safeBase64Encode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// GeneratePako 生成 Mermaid 图表的 pako URL
func GeneratePako(graphMarkdown string, config *Config) (string, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	graphData := map[string]interface{}{
		"code":    graphMarkdown,
		"mermaid": config,
	}
	
	jsonBytes, err := json.Marshal(graphData)
	if err != nil {
		return "", err
	}
	
	compressedData, err := compressToDeflate(jsonBytes)
	if err != nil {
		return "", err
	}
	
	base64Encoded := safeBase64Encode(compressedData)
	return fmt.Sprintf("pako:%s", base64Encoded), nil
}

// GetMermaidLiveURL 获取 Mermaid Live 编辑器 URL
// 可用于在浏览器中编辑图表
func GetMermaidLiveURL(graphMarkdown string) (string, error) {
	pako, err := GeneratePako(graphMarkdown, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://mermaid.live/edit/#%s", pako), nil
}

// GetMermaidInkURL 获取 Mermaid Ink 图片 URL
// 可用于下载图片
func GetMermaidInkURL(graphMarkdown string) (string, error) {
	pako, err := GeneratePako(graphMarkdown, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://mermaid.ink/img/%s?theme=default&width=500&scale=2&type=webp", pako), nil
}

// DownloadImage 异步下载图片
func DownloadImage(ctx context.Context, url string, client *http.Client) (*bytes.Buffer, error) {
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}
	
	return &buf, nil
}

// IsImage 检查数据是否为有效图片
// 使用 Go 标准库 image 包验证图片格式
func IsImage(data *bytes.Buffer) bool {
	if data.Len() == 0 {
		return false
	}
	
	// 使用 image.DecodeConfig 验证图片
	// 这会尝试解码图片头部，支持 PNG, JPEG, GIF, WebP
	reader := bytes.NewReader(data.Bytes())
	_, _, err := image.DecodeConfig(reader)
	return err == nil
}

// RenderMermaid 渲染 Mermaid 图表
// 返回图片数据和编辑 URL
func RenderMermaid(ctx context.Context, diagram string, client *http.Client) (*bytes.Buffer, string, error) {
	// 生成 URL
	imgURL, err := GetMermaidInkURL(diagram)
	if err != nil {
		return nil, "", err
	}
	
	caption, err := GetMermaidLiveURL(diagram)
	if err != nil {
		return nil, "", err
	}
	
	// 下载图片
	imgData, err := DownloadImage(ctx, imgURL, client)
	if err != nil {
		return nil, "", err
	}
	
	// 验证图片
	if !IsImage(imgData) {
		return nil, "", fmt.Errorf("downloaded data is not a valid image")
	}
	
	return imgData, caption, nil
}

// SupportMermaid 检查是否支持 Mermaid 渲染
// Go 版本总是返回 true，因为不需要额外依赖
func SupportMermaid() bool {
	return true
}

