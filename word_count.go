package telegramify

// CountText 计算文本在 Telegram 中的有效长度（UTF-16 code units）
//
// 由于使用 entity-based 方法，发送给 Telegram 的文本是纯文本
// （没有 MarkdownV2 语法）。URL 存储在 entity 字段中，而不是文本中。
// 因此计数只是文本的 UTF-16 长度。
//
// 参数：
//   - text: 要计数的文本
//
// 返回：
//   - int: UTF-16 code units 数量
func CountText(text string) int {
	return UTF16Len(text)
}

