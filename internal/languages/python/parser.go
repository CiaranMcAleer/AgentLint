package python

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	"github.com/CiaranMcAleer/AgentLint/internal/languages/python/rules"
)

// Parser is the Python file parser
type Parser struct {
	config core.Config
	cache  *Cache

	// Compiled regex patterns for parsing
	funcPattern     *regexp.Regexp
	classPattern    *regexp.Regexp
	importPattern   *regexp.Regexp
	fromPattern     *regexp.Regexp
	decoratorPattern *regexp.Regexp
	variablePattern *regexp.Regexp
}

// NewParser creates a new Python parser
func NewParser(config core.Config) *Parser {
	return &Parser{
		config:           config,
		cache:            NewCache(0),
		funcPattern:      regexp.MustCompile(`^(\s*)def\s+(\w+)\s*\(`),
		classPattern:     regexp.MustCompile(`^(\s*)class\s+(\w+)\s*(?:\(([^)]*)\))?:`),
		importPattern:    regexp.MustCompile(`^import\s+(.+)`),
		fromPattern:      regexp.MustCompile(`^from\s+(\S+)\s+import\s+(.+)`),
		decoratorPattern: regexp.MustCompile(`^(\s*)@(\w+)`),
		variablePattern:  regexp.MustCompile(`^(\s*)(\w+)\s*=`),
	}
}

// lineParseState holds mutable state during line-by-line parsing
type lineParseState struct {
	lineNum              int
	pendingDecorators    []string
	inMultilineString    bool
	multilineStringDelim string
}

// ParseFile parses a Python file
func (p *Parser) ParseFile(ctx context.Context, filePath string) (*ParsedFile, error) {
	if cached, ok := p.cache.Get(filePath); ok {
		return cached, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	parsed := p.newParsedFile()
	state := &lineParseState{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		state.lineNum++
		line := scanner.Text()
		parsed.Lines = append(parsed.Lines, line)
		parsed.TotalLines++

		p.processLine(line, state, parsed)
	}

	p.calculateFunctionEndLines(parsed)
	p.cache.Set(filePath, parsed)

	return parsed, scanner.Err()
}

// newParsedFile creates a new initialized ParsedFile
func (p *Parser) newParsedFile() *ParsedFile {
	return &ParsedFile{
		Lines:     make([]string, 0),
		Functions: make([]FunctionDef, 0),
		Classes:   make([]ClassDef, 0),
		Imports:   make([]ImportStmt, 0),
		Comments:  make([]Comment, 0),
	}
}

// processLine processes a single line of Python code
func (p *Parser) processLine(line string, state *lineParseState, parsed *ParsedFile) {
	trimmed := strings.TrimSpace(line)

	if trimmed == "" {
		parsed.BlankLines++
		return
	}

	if p.handleMultilineString(line, trimmed, state, parsed) {
		return
	}

	if p.handleComment(line, trimmed, state, parsed) {
		return
	}

	p.handleInlineComment(line, state, parsed)

	if p.handleDecorator(line, state) {
		return
	}

	if p.handleClass(line, state, parsed) {
		return
	}

	if p.handleFunction(line, state, parsed) {
		return
	}

	if p.handleImport(line, state, parsed) {
		return
	}

	p.handleVariable(line, state, parsed)
	parsed.CodeLines++
}

// handleMultilineString handles multiline string (docstring) parsing
func (p *Parser) handleMultilineString(line, trimmed string, state *lineParseState, parsed *ParsedFile) bool {
	if state.inMultilineString {
		if strings.Contains(line, state.multilineStringDelim) {
			state.inMultilineString = false
		}
		return true
	}

	if strings.Contains(trimmed, `"""`) || strings.Contains(trimmed, `'''`) {
		delim := `"""`
		if strings.Contains(trimmed, `'''`) {
			delim = `'''`
		}
		if strings.Count(trimmed, delim) == 1 {
			state.inMultilineString = true
			state.multilineStringDelim = delim
		}
		parsed.CommentLines++
		return true
	}

	return false
}

// handleComment handles standalone comment lines
func (p *Parser) handleComment(line, trimmed string, state *lineParseState, parsed *ParsedFile) bool {
	if strings.HasPrefix(trimmed, "#") {
		parsed.Comments = append(parsed.Comments, Comment{
			Text:     trimmed,
			Line:     state.lineNum,
			IsInline: false,
		})
		parsed.CommentLines++
		return true
	}
	return false
}

// handleInlineComment handles inline comments at the end of code lines
func (p *Parser) handleInlineComment(line string, state *lineParseState, parsed *ParsedFile) {
	if idx := strings.Index(line, " #"); idx != -1 {
		commentText := strings.TrimSpace(line[idx+1:])
		parsed.Comments = append(parsed.Comments, Comment{
			Text:     commentText,
			Line:     state.lineNum,
			IsInline: true,
		})
	}
}

// handleDecorator handles decorator lines
func (p *Parser) handleDecorator(line string, state *lineParseState) bool {
	if matches := p.decoratorPattern.FindStringSubmatch(line); matches != nil {
		state.pendingDecorators = append(state.pendingDecorators, matches[2])
		return true
	}
	return false
}

// handleClass handles class definition lines
func (p *Parser) handleClass(line string, state *lineParseState, parsed *ParsedFile) bool {
	matches := p.classPattern.FindStringSubmatch(line)
	if matches == nil {
		return false
	}

	var bases []string
	if matches[3] != "" {
		bases = splitAndTrim(matches[3])
	}

	parsed.Classes = append(parsed.Classes, ClassDef{
		Name:       matches[2],
		StartLine:  state.lineNum,
		Bases:      bases,
		Decorators: state.pendingDecorators,
	})
	state.pendingDecorators = nil
	return true
}

// handleFunction handles function definition lines
func (p *Parser) handleFunction(line string, state *lineParseState, parsed *ParsedFile) bool {
	matches := p.funcPattern.FindStringSubmatch(line)
	if matches == nil {
		return false
	}

	indent := len(matches[1])
	funcName := matches[2]

	funcDef := FunctionDef{
		Name:       funcName,
		StartLine:  state.lineNum,
		Decorators: state.pendingDecorators,
		IsPrivate:  strings.HasPrefix(funcName, "_"),
		Indent:     indent,
	}

	if len(parsed.Classes) > 0 && indent > 0 {
		funcDef.IsMethod = true
		funcDef.ClassName = parsed.Classes[len(parsed.Classes)-1].Name
	}

	parsed.Functions = append(parsed.Functions, funcDef)
	state.pendingDecorators = nil
	return true
}

// handleImport handles import statement lines
func (p *Parser) handleImport(line string, state *lineParseState, parsed *ParsedFile) bool {
	if matches := p.fromPattern.FindStringSubmatch(line); matches != nil {
		parsed.Imports = append(parsed.Imports, ImportStmt{
			Module: matches[1],
			Names:  splitAndTrim(matches[2]),
			IsFrom: true,
			Line:   state.lineNum,
		})
		return true
	}

	if matches := p.importPattern.FindStringSubmatch(line); matches != nil {
		for _, mod := range splitAndTrim(matches[1]) {
			parsed.Imports = append(parsed.Imports, ImportStmt{
				Module: strings.Split(mod, " as ")[0],
				IsFrom: false,
				Line:   state.lineNum,
			})
		}
		return true
	}

	return false
}

// handleVariable handles variable definition lines at module level
func (p *Parser) handleVariable(line string, state *lineParseState, parsed *ParsedFile) {
	matches := p.variablePattern.FindStringSubmatch(line)
	if matches == nil {
		return
	}

	indent := len(matches[1])
	if indent != 0 {
		return
	}

	varName := matches[2]
	if !strings.HasPrefix(varName, "_") || !unicode.IsUpper(rune(varName[0])) {
		parsed.Variables = append(parsed.Variables, VariableDef{
			Name:     varName,
			Line:     state.lineNum,
			IsGlobal: true,
		})
	}
}

// calculateFunctionEndLines determines where each function ends based on indentation
func (p *Parser) calculateFunctionEndLines(parsed *ParsedFile) {
	for i := range parsed.Functions {
		funcDef := &parsed.Functions[i]
		funcIndent := funcDef.Indent

		// Find the end of the function by looking for the next line with same or less indentation
		for j := funcDef.StartLine; j < len(parsed.Lines); j++ {
			line := parsed.Lines[j]
			if strings.TrimSpace(line) == "" {
				continue
			}

			currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))
			// Convert tabs to spaces (assuming 4 spaces per tab)
			currentIndent = strings.Count(line[:currentIndent], "\t")*4 + strings.Count(line[:currentIndent], " ")

			if j > funcDef.StartLine && currentIndent <= funcIndent && strings.TrimSpace(line) != "" {
				funcDef.EndLine = j
				break
			}
		}

		// If we didn't find an end, use the last line
		if funcDef.EndLine == 0 {
			funcDef.EndLine = len(parsed.Lines)
		}
	}
}

// CalculateFileMetrics calculates metrics for a parsed file
func (p *Parser) CalculateFileMetrics(ctx context.Context, filePath string, parsed *ParsedFile) *rules.FileMetrics {
	var commentRatio float64
	if parsed.CodeLines > 0 {
		commentRatio = float64(parsed.CommentLines) / float64(parsed.CodeLines)
	}

	return &rules.FileMetrics{
		Path:          filePath,
		TotalLines:    parsed.TotalLines,
		CodeLines:     parsed.CodeLines,
		CommentLines:  parsed.CommentLines,
		BlankLines:    parsed.BlankLines,
		CommentRatio:  commentRatio,
		FunctionCount: len(parsed.Functions),
		ImportCount:   len(parsed.Imports),
		ClassCount:    len(parsed.Classes),
	}
}

// CalculateFunctionMetrics calculates metrics for all functions in a parsed file
func (p *Parser) CalculateFunctionMetrics(ctx context.Context, parsed *ParsedFile) []*rules.FunctionMetrics {
	metrics := make([]*rules.FunctionMetrics, 0, len(parsed.Functions))

	for _, fn := range parsed.Functions {
		lineCount := fn.EndLine - fn.StartLine
		if lineCount < 0 {
			lineCount = 0
		}

		// Calculate nesting depth
		nestingDepth := 0
		for i := fn.StartLine - 1; i < fn.EndLine && i < len(parsed.Lines); i++ {
			line := parsed.Lines[i]
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "if ") || strings.HasPrefix(trimmed, "for ") ||
				strings.HasPrefix(trimmed, "while ") || strings.HasPrefix(trimmed, "with ") ||
				strings.HasPrefix(trimmed, "try:") || strings.HasPrefix(trimmed, "except") {
				depth := countLeadingSpaces(line) / 4
				if depth > nestingDepth {
					nestingDepth = depth
				}
			}
		}

		metrics = append(metrics, &rules.FunctionMetrics{
			Name:         fn.Name,
			IsMethod:     fn.IsMethod,
			ClassName:    fn.ClassName,
			IsPrivate:    fn.IsPrivate,
			LineCount:    lineCount,
			StartLine:    fn.StartLine,
			NestingDepth: nestingDepth,
			Decorators:   fn.Decorators,
		})
	}

	return metrics
}

// splitAndTrim splits a string by comma and trims each part
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// countLeadingSpaces counts the number of leading spaces in a line
func countLeadingSpaces(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4
		} else {
			break
		}
	}
	return count
}
