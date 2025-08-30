package tui

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

// applySyntaxHighlighting applies syntax highlighting to code blocks in the content
func applySyntaxHighlighting(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inCodeBlock bool
	var codeLines []string
	var language string

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				// Start of code block
				inCodeBlock = true
				language = strings.TrimPrefix(line, "```")
				language = strings.TrimSpace(language)
				if language == "" {
					language = "text"
				}
				codeLines = []string{}
			} else {
				// End of code block
				inCodeBlock = false
				highlighted := highlightCode(strings.Join(codeLines, "\n"), language)
				result = append(result, "```"+language)
				result = append(result, highlighted)
				result = append(result, "```")
			}
		} else {
			if inCodeBlock {
				codeLines = append(codeLines, line)
			} else {
				result = append(result, line)
			}
		}
	}

	// Handle unclosed code block
	if inCodeBlock {
		result = append(result, "```"+language)
		result = append(result, strings.Join(codeLines, "\n"))
		result = append(result, "```")
	}

	return strings.Join(result, "\n")
}

// highlightCode applies syntax highlighting to a code string
func highlightCode(code, language string) string {
	// Get lexer for language
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Get style
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	// Get formatter
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Tokenize
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	// Format
	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return code
	}

	return buf.String()
}

// HighlightLine applies simple highlighting to a single line
func HighlightLine(line string, theme *Theme) string {
	// Simple keyword highlighting
	keywords := []string{
		"func", "function", "def", "class", "struct", "interface",
		"if", "else", "for", "while", "return", "import", "package",
		"const", "let", "var", "type", "enum",
	}

	result := line
	for _, keyword := range keywords {
		result = strings.ReplaceAll(result,
			" "+keyword+" ",
			" "+theme.KeywordStyle.Render(keyword)+" ",
		)
	}

	// Highlight strings (simple approach)
	if strings.Contains(result, `"`) {
		parts := strings.Split(result, `"`)
		for i := 1; i < len(parts); i += 2 {
			if i < len(parts) {
				parts[i] = theme.StringStyle.Render(`"` + parts[i] + `"`)
			}
		}
		result = strings.Join(parts, "")
	}

	// Highlight comments
	if strings.Contains(result, "//") {
		idx := strings.Index(result, "//")
		result = result[:idx] + theme.CommentStyle.Render(result[idx:])
	}

	return result
}

// FormatMarkdown formats markdown content with basic styling
func FormatMarkdown(content string, theme *Theme) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Headers
		if strings.HasPrefix(trimmed, "# ") {
			text := strings.TrimPrefix(trimmed, "# ")
			result = append(result, theme.HeaderStyle.Render(text))
		} else if strings.HasPrefix(trimmed, "## ") {
			text := strings.TrimPrefix(trimmed, "## ")
			result = append(result, theme.HeaderStyle.Render(text))
		} else if strings.HasPrefix(trimmed, "### ") {
			text := strings.TrimPrefix(trimmed, "### ")
			result = append(result, theme.HeaderStyle.Render(text))
		} else if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			// Bullet points
			result = append(result, line)
		} else if strings.HasPrefix(trimmed, "> ") {
			// Quotes
			text := strings.TrimPrefix(trimmed, "> ")
			result = append(result, theme.CommentStyle.Render("> "+text))
		} else {
			// Regular text
			// Bold
			line = replaceBold(line, theme)
			// Italic
			line = replaceItalic(line, theme)
			// Code
			line = replaceInlineCode(line, theme)

			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func replaceBold(text string, theme *Theme) string {
	parts := strings.Split(text, "**")
	for i := 1; i < len(parts); i += 2 {
		if i < len(parts) {
			parts[i] = lipgloss.NewStyle().Bold(true).Render(parts[i])
		}
	}
	return strings.Join(parts, "")
}

func replaceItalic(text string, theme *Theme) string {
	parts := strings.Split(text, "*")
	for i := 1; i < len(parts); i += 2 {
		if i < len(parts) {
			parts[i] = lipgloss.NewStyle().Italic(true).Render(parts[i])
		}
	}
	return strings.Join(parts, "")
}

func replaceInlineCode(text string, theme *Theme) string {
	parts := strings.Split(text, "`")
	for i := 1; i < len(parts); i += 2 {
		if i < len(parts) {
			parts[i] = theme.CodeBlockStyle.Render(parts[i])
		}
	}
	return strings.Join(parts, "")
}
