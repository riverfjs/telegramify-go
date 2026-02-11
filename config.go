package telegramify

import (
	"sync"
	"github.com/riverfjs/telegramify-go/internal/types"
)

// 导出类型别名
type Symbol = types.Symbol
type RenderConfig = types.RenderConfig

var (
	defaultConfig     *RenderConfig
	defaultConfigOnce sync.Once
)

// DefaultConfig returns the default render configuration (singleton).
func DefaultConfig() *RenderConfig {
	defaultConfigOnce.Do(func() {
		defaultConfig = types.DefaultRenderConfig()
	})
	return defaultConfig
}

