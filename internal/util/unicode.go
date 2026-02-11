package util

import (
	"path/filepath"
	"regexp"
	"strings"
)

// DefaultLanguageToExt maps programming language names to file extensions.
var DefaultLanguageToExt = map[string]string{
	"python":     "py",
	"javascript": "js",
	"typescript": "ts",
	"java":       "java",
	"c++":        "cpp",
	"c":          "c",
	"html":       "html",
	"css":        "css",
	"bash":       "sh",
	"shell":      "sh",
	"php":        "php",
	"markdown":   "md",
	"dotenv":     "env",
	"json":       "json",
	"yaml":       "yaml",
	"xml":        "xml",
	"dockerfile": "dockerfile",
	"plaintext":  "txt",
	"toml":       "toml",
	"go":         "go",
	"ruby":       "rb",
	"rust":       "rs",
	"perl":       "pl",
	"swift":      "swift",
	"kotlin":     "kt",
	"sql":        "sql",
	"jsx":        "jsx",
	"tsx":        "tsx",
	"graphql":    "graphql",
	"r":          "r",
	"dart":       "dart",
	"scala":      "scala",
	"groovy":     "groovy",
}

var filenamePattern = regexp.MustCompile(`([a-zA-Z0-9_\-\.]+\.[a-zA-Z0-9]+)`)

// ExtractValidFilename extracts a valid filename (with extension) from a line of text.
func ExtractValidFilename(line string) string {
	matches := filenamePattern.FindAllString(line, -1)
	for _, match := range matches {
		if filepath.Ext(match) != "" {
			return match
		}
	}
	return ""
}

// GetExt returns the file extension for a given language.
func GetExt(language string) string {
	ext, ok := DefaultLanguageToExt[strings.ToLower(language)]
	if !ok {
		return "txt"
	}
	return ext
}

// GetFilename generates a filename for a code block.
//
// Tries to extract a filename from the first line of the code.
// Falls back to 'readable.<ext>' based on the language.
func GetFilename(code string, language string) string {
	// Take the first two lines
	lines := strings.Split(strings.TrimSpace(code), "\n")
	sample := ""
	if len(lines) > 0 {
		sample = lines[0]
		if len(lines) > 1 {
			sample += lines[1]
		}
	}
	sample = strings.ReplaceAll(sample, "\\", "")

	extractedFilename := ExtractValidFilename(sample)
	ext := GetExt(language)

	if extractedFilename != "" {
		// Check if it already has the correct extension and is reasonably short
		if strings.HasSuffix(extractedFilename, "."+ext) && len(extractedFilename) <= 24 {
			return extractedFilename
		}
		return extractedFilename + "." + ext
	}

	return "readable." + ext
}

