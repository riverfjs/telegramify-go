package mermaid

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"
	"time"
)

// TestGeneratePako 测试 Pako 生成
func TestGeneratePako(t *testing.T) {
	tests := []struct {
		name     string
		diagram  string
		wantErr  bool
		contains string
	}{
		{
			name:     "simple graph",
			diagram:  "graph LR\n    A-->B",
			wantErr:  false,
			contains: "pako:",
		},
		{
			name:     "empty diagram",
			diagram:  "",
			wantErr:  false,
			contains: "pako:",
		},
		{
			name:     "complex diagram",
			diagram:  "flowchart TD\n    A[Start] --> B{Check}\n    B -->|Yes| C[OK]\n    B -->|No| D[Cancel]",
			wantErr:  false,
			contains: "pako:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GeneratePako(tt.diagram, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePako() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.HasPrefix(got, tt.contains) {
				t.Errorf("GeneratePako() = %v, should contain %v", got, tt.contains)
			}
		})
	}
}

// TestGetMermaidLiveURL 测试获取编辑器 URL
func TestGetMermaidLiveURL(t *testing.T) {
	diagram := "graph LR\n    A-->B"
	url, err := GetMermaidLiveURL(diagram)
	if err != nil {
		t.Fatalf("GetMermaidLiveURL() error = %v", err)
	}
	
	if !strings.HasPrefix(url, "https://mermaid.live/edit/#pako:") {
		t.Errorf("GetMermaidLiveURL() = %v, should start with https://mermaid.live/edit/#pako:", url)
	}
}

// TestGetMermaidInkURL 测试获取图片 URL
func TestGetMermaidInkURL(t *testing.T) {
	diagram := "graph LR\n    A-->B"
	url, err := GetMermaidInkURL(diagram)
	if err != nil {
		t.Fatalf("GetMermaidInkURL() error = %v", err)
	}
	
	if !strings.HasPrefix(url, "https://mermaid.ink/img/pako:") {
		t.Errorf("GetMermaidInkURL() = %v, should start with https://mermaid.ink/img/pako:", url)
	}
	
	if !strings.Contains(url, "theme=default") {
		t.Errorf("GetMermaidInkURL() = %v, should contain theme=default", url)
	}
}

// TestIsImage 测试图片验证
func TestIsImage(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		setup func() *bytes.Buffer
		want  bool
	}{
		{
			name: "valid PNG",
			setup: func() *bytes.Buffer {
				// 创建一个简单的 PNG 图片
				img := image.NewRGBA(image.Rect(0, 0, 10, 10))
				img.Set(5, 5, color.RGBA{255, 0, 0, 255})
				
				var buf bytes.Buffer
				png.Encode(&buf, img)
				return &buf
			},
			want: true,
		},
		{
			name: "valid JPEG",
			setup: func() *bytes.Buffer {
				img := image.NewRGBA(image.Rect(0, 0, 10, 10))
				var buf bytes.Buffer
				jpeg.Encode(&buf, img, nil)
				return &buf
			},
			want: true,
		},
		{
			name: "valid GIF",
			setup: func() *bytes.Buffer {
				img := image.NewPaletted(image.Rect(0, 0, 10, 10), color.Palette{
					color.RGBA{0, 0, 0, 255},
					color.RGBA{255, 255, 255, 255},
				})
				var buf bytes.Buffer
				gif.Encode(&buf, img, nil)
				return &buf
			},
			want: true,
		},
		{
			name: "empty data",
			setup: func() *bytes.Buffer {
				return &bytes.Buffer{}
			},
			want: false,
		},
		{
			name: "invalid data",
			setup: func() *bytes.Buffer {
				return bytes.NewBufferString("not an image")
			},
			want: false,
		},
		{
			name: "corrupted PNG header",
			setup: func() *bytes.Buffer {
				return bytes.NewBuffer([]byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00})
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := tt.setup()
			if got := IsImage(buf); got != tt.want {
				t.Errorf("IsImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDownloadImage 测试图片下载（需要网络）
func TestDownloadImage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 使用一个简单的 Mermaid 图表测试
	diagram := "graph LR\n    A-->B"
	testURL, err := GetMermaidInkURL(diagram)
	if err != nil {
		t.Fatalf("GetMermaidInkURL() error = %v", err)
	}
	
	data, err := DownloadImage(ctx, testURL, nil)
	if err != nil {
		t.Skipf("DownloadImage() error = %v (mermaid.ink may be unavailable, skipping)", err)
		return
	}
	
	if data.Len() == 0 {
		t.Error("DownloadImage() returned empty data")
	}
	
	if !IsImage(data) {
		t.Error("DownloadImage() returned invalid image data")
	}
	
	t.Logf("Downloaded image size: %d bytes", data.Len())
}

// TestCompressToDeflate 测试压缩功能
func TestCompressToDeflate(t *testing.T) {
	testData := []byte("Hello, World! This is a test string for compression.")
	
	compressed, err := compressToDeflate(testData)
	if err != nil {
		t.Fatalf("compressToDeflate() error = %v", err)
	}
	
	if len(compressed) == 0 {
		t.Error("compressToDeflate() returned empty data")
	}
	
	// 压缩后的数据应该比原始数据小（对于大多数文本）
	// 但对于很短的字符串可能会更大，所以我们只检查不为空
	if len(compressed) > len(testData)*2 {
		t.Errorf("compressed data too large: got %d, want less than %d", len(compressed), len(testData)*2)
	}
}

// TestSafeBase64Encode 测试 base64 编码
func TestSafeBase64Encode(t *testing.T) {
	testData := []byte("test data")
	encoded := safeBase64Encode(testData)
	
	if encoded == "" {
		t.Error("safeBase64Encode() returned empty string")
	}
	
	// URL-safe base64 不应该包含 + 或 /
	if strings.Contains(encoded, "+") || strings.Contains(encoded, "/") {
		t.Error("safeBase64Encode() should return URL-safe encoding")
	}
}

// TestSupportMermaid 测试 Mermaid 支持检查
func TestSupportMermaid(t *testing.T) {
	if !SupportMermaid() {
		t.Error("SupportMermaid() should return true")
	}
}

// BenchmarkGeneratePako 基准测试 Pako 生成
func BenchmarkGeneratePako(b *testing.B) {
	diagram := "graph TD\n    A[Start] --> B[Process]\n    B --> C[End]"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GeneratePako(diagram, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkIsImage 基准测试图片验证
func BenchmarkIsImage(b *testing.B) {
	// 创建一个测试图片
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	png.Encode(&buf, img)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsImage(&buf)
	}
}

