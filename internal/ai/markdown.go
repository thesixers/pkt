package ai

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var (
	reBold         = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reItalic       = regexp.MustCompile(`\*([^*]+)\*`)
	reCode         = regexp.MustCompile("`([^`]+)`")
	reLink         = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reSkillMention = regexp.MustCompile(`(^|\s)([#$][a-zA-Z0-9_\-]+)`)
	reFileMention  = regexp.MustCompile(`(^|\s)(@[a-zA-Z0-9_\-./]+)`)
	reBullet       = regexp.MustCompile(`^(\s*)[-*+]\s+(.*)$`)
	reNumbered     = regexp.MustCompile(`^(\s*)(\d+\.)\s+(.*)$`)
	reCheckbox     = regexp.MustCompile(`^(\s*)[-*+]\s+\[([ xX])\]\s+(.*)$`)
)

// RenderMarkdown parses and renders terminal-optimized markdown with syntax-highlighted
// code blocks, styled rounded Unicode tables, and ANSI formatted prose.
func RenderMarkdown(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	termWidth := getTerminalWidth()
	lines := strings.Split(content, "\n")
	var out []string

	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// 1. Detect Fenced Code Block: ``` or ~~~
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fence := trimmed[:3]
			lang := strings.TrimSpace(trimmed[3:])
			var codeLines []string
			i++
			for i < len(lines) {
				if strings.HasPrefix(strings.TrimSpace(lines[i]), fence) {
					i++
					break
				}
				codeLines = append(codeLines, lines[i])
				i++
			}
			renderedCode := RenderCodeBlock(lang, strings.Join(codeLines, "\n"), termWidth)
			out = append(out, renderedCode)
			continue
		}

		// 2. Detect Markdown Table
		if strings.Contains(line, "|") && i+1 < len(lines) && isTableSeparator(lines[i+1]) {
			headers := splitTableRow(line)
			sepLine := lines[i+1]
			aligns := parseTableAlignments(sepLine)
			i += 2

			var dataRows [][]string
			for i < len(lines) {
				rowLine := strings.TrimSpace(lines[i])
				if rowLine == "" || !strings.Contains(rowLine, "|") {
					break
				}
				dataRows = append(dataRows, splitTableRow(rowLine))
				i++
			}

			renderedTable := RenderTable(headers, aligns, dataRows, termWidth)
			out = append(out, renderedTable)
			continue
		}

		// 3. Render Prose line
		renderedLine := renderMarkdownLine(line, termWidth)
		out = append(out, renderedLine)
		i++
	}

	return strings.TrimSpace(strings.Join(out, "\n"))
}

// RenderCodeBlock formats code with Chroma syntax highlighting, framed box borders,
// language badge, and aligned line numbers.
func RenderCodeBlock(lang, code string, width int) string {
	code = strings.Trim(code, "\r\n")
	if code == "" {
		return ""
	}

	highlighted := highlightCode(lang, code)
	codeLines := strings.Split(highlighted, "\n")
	if len(codeLines) > 0 && codeLines[len(codeLines)-1] == "" {
		codeLines = codeLines[:len(codeLines)-1]
	}

	boxWidth := width
	if boxWidth < 50 {
		boxWidth = 60
	}
	if boxWidth > 100 {
		boxWidth = 100
	}

	langLabel := strings.TrimSpace(lang)
	if langLabel == "" {
		langLabel = "code"
	}
	badge := fmt.Sprintf("\033[1;38;5;75m[%s]\033[0m", langLabel)
	badgeLen := len(langLabel) + 2

	dashCount := boxWidth - badgeLen - 6
	if dashCount < 4 {
		dashCount = 4
	}

	var b strings.Builder
	// Top border: ╭─── [go] ──────────────────────────────────────────╮
	fmt.Fprintf(&b, "\n\033[38;5;238m╭───\033[0m %s \033[38;5;238m%s╮\033[0m\n", badge, strings.Repeat("─", dashCount))

	lineCount := len(codeLines)
	numWidth := len(fmt.Sprintf("%d", lineCount))
	if numWidth < 2 {
		numWidth = 2
	}

	for i, l := range codeLines {
		lineNum := fmt.Sprintf("\033[38;5;241m%*d\033[0m", numWidth, i+1)
		fmt.Fprintf(&b, "\033[38;5;238m│\033[0m %s \033[38;5;238m│\033[0m %s\n", lineNum, l)
	}

	// Bottom border: ╰────────────────────────────────────────────────╯
	bottomDashes := boxWidth - 2
	fmt.Fprintf(&b, "\033[38;5;238m╰%s╯\033[0m\n", strings.Repeat("─", bottomDashes))

	return b.String()
}

func highlightCode(lang, source string) string {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Analyse(source)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("dracula")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, source)
	if err != nil {
		return source
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return source
	}
	return buf.String()
}

func isTableSeparator(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.Contains(trimmed, "|") || !strings.Contains(trimmed, "-") {
		return false
	}
	parts := splitTableRow(trimmed)
	if len(parts) == 0 {
		return false
	}
	for _, p := range parts {
		cell := strings.TrimSpace(p)
		if cell == "" {
			continue
		}
		cell = strings.Trim(cell, ":")
		if strings.Trim(cell, "-") != "" {
			return false
		}
	}
	return true
}

func splitTableRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	cells := strings.Split(trimmed, "|")
	var result []string
	for _, c := range cells {
		result = append(result, strings.TrimSpace(c))
	}
	return result
}

func parseTableAlignments(sepLine string) []text.Align {
	cells := splitTableRow(sepLine)
	var aligns []text.Align
	for _, cell := range cells {
		trimmed := strings.TrimSpace(cell)
		hasLeft := strings.HasPrefix(trimmed, ":")
		hasRight := strings.HasSuffix(trimmed, ":")
		if hasLeft && hasRight {
			aligns = append(aligns, text.AlignCenter)
		} else if hasRight {
			aligns = append(aligns, text.AlignRight)
		} else {
			aligns = append(aligns, text.AlignLeft)
		}
	}
	return aligns
}

// RenderTable builds a rounded Unicode box table using go-pretty.
func RenderTable(headers []string, aligns []text.Align, rows [][]string, termWidth int) string {
	if len(headers) == 0 {
		return ""
	}

	t := table.NewWriter()

	// Style: StyleRounded with dark border and bright cyan headers
	style := table.StyleRounded
	style.Options.SeparateRows = false
	style.Options.DrawBorder = true
	style.Color.Header = text.Colors{text.Bold, text.FgHiCyan}
	style.Color.Border = text.Colors{text.FgHiBlack}
	style.Color.Separator = text.Colors{text.FgHiBlack}
	t.SetStyle(style)

	// Format headers with inline markdown
	headerRow := make(table.Row, len(headers))
	for i, h := range headers {
		headerRow[i] = formatInlineMarkdown(h)
	}
	t.AppendHeader(headerRow)

	// Configure column alignments and max wrapping width
	colCount := len(headers)
	availableWidth := termWidth - (colCount * 3) - 5
	colMax := availableWidth / colCount
	if colMax < 12 {
		colMax = 12
	}
	if colMax > 45 {
		colMax = 45
	}

	colConfigs := make([]table.ColumnConfig, colCount)
	for i := range headers {
		align := text.AlignLeft
		if i < len(aligns) {
			align = aligns[i]
		}
		colConfigs[i] = table.ColumnConfig{
			Number:   i + 1,
			Align:    align,
			WidthMax: colMax,
		}
	}
	t.SetColumnConfigs(colConfigs)

	// Append data rows
	for _, row := range rows {
		tRow := make(table.Row, colCount)
		for i := 0; i < colCount; i++ {
			if i < len(row) {
				tRow[i] = formatInlineMarkdown(row[i])
			} else {
				tRow[i] = ""
			}
		}
		t.AppendRow(tRow)
	}

	return "\n" + t.Render() + "\n"
}

func renderMarkdownLine(line string, width int) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return ""
	}

	// Headings
	if strings.HasPrefix(trimmed, "# ") {
		return fmt.Sprintf("\n\033[1;38;5;39m◈ %s\033[0m", formatInlineMarkdown(strings.TrimPrefix(trimmed, "# ")))
	}
	if strings.HasPrefix(trimmed, "## ") {
		return fmt.Sprintf("\n\033[1;38;5;75m◆ %s\033[0m", formatInlineMarkdown(strings.TrimPrefix(trimmed, "## ")))
	}
	if strings.HasPrefix(trimmed, "### ") {
		return fmt.Sprintf("\n\033[1;38;5;111m◇ %s\033[0m", formatInlineMarkdown(strings.TrimPrefix(trimmed, "### ")))
	}
	if strings.HasPrefix(trimmed, "#### ") {
		return fmt.Sprintf("\n\033[1;38;5;147m%s\033[0m", formatInlineMarkdown(strings.TrimPrefix(trimmed, "#### ")))
	}

	// Horizontal Rule
	if trimmed == "---" || trimmed == "***" || trimmed == "___" {
		hrWidth := width - 4
		if hrWidth < 20 {
			hrWidth = 20
		}
		if hrWidth > 80 {
			hrWidth = 80
		}
		return fmt.Sprintf("\033[38;5;238m%s\033[0m", strings.Repeat("─", hrWidth))
	}

	// Checkbox task list: - [ ] or - [x]
	if match := reCheckbox.FindStringSubmatch(line); len(match) == 4 {
		indent := match[1]
		checked := strings.ToLower(match[2]) == "x"
		box := "\033[38;5;244m☐\033[0m"
		if checked {
			box = "\033[1;38;5;78m☑\033[0m"
		}
		return fmt.Sprintf("%s  %s %s", indent, box, formatInlineMarkdown(match[3]))
	}

	// Bullet List
	if match := reBullet.FindStringSubmatch(line); len(match) == 3 {
		indent := match[1]
		return fmt.Sprintf("%s  \033[38;5;75m•\033[0m %s", indent, formatInlineMarkdown(match[2]))
	}

	// Numbered List
	if match := reNumbered.FindStringSubmatch(line); len(match) == 4 {
		indent := match[1]
		return fmt.Sprintf("%s  \033[38;5;75m%s\033[0m %s", indent, match[2], formatInlineMarkdown(match[3]))
	}

	// Blockquote
	if strings.HasPrefix(trimmed, "> ") {
		content := formatInlineMarkdown(strings.TrimPrefix(trimmed, "> "))
		return fmt.Sprintf("  \033[38;5;240m│\033[0m \033[3;38;5;248m%s\033[0m", content)
	}

	// Paragraph text: soft-wrap to terminal width
	return text.WrapSoft(formatInlineMarkdown(line), width)
}

func formatInlineMarkdown(s string) string {
	// 1. Inline code: styled with bold cyan
	s = reCode.ReplaceAllString(s, "\033[1;38;5;81m`$1`\033[0m")
	// 2. Bold
	s = reBold.ReplaceAllString(s, "\033[1m$1\033[0m")
	// 3. Italic
	s = reItalic.ReplaceAllString(s, "\033[3m$1\033[0m")
	// 4. Links
	s = reLink.ReplaceAllString(s, "\033[4;38;5;75m$1\033[0m \033[2m($2)\033[0m")
	// 5. Skill mentions: #skill or $skill in purple/magenta
	s = reSkillMention.ReplaceAllString(s, "$1\033[1;38;5;141m$2\033[0m")
	// 6. File mentions: @file in blue
	s = reFileMention.ReplaceAllString(s, "$1\033[38;5;75m$2\033[0m")
	return s
}
