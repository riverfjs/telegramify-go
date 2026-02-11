package telegramify

import (
	"log"
	"os"
)

// Logger 全局日志记录器
var Logger = log.New(os.Stderr, "[telegramify] ", log.LstdFlags)

// SetLogger 设置自定义日志记录器
func SetLogger(logger *log.Logger) {
	Logger = logger
}

