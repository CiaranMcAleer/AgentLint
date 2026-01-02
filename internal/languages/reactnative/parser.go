package reactnative

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/CiaranMcAleer/AgentLint/internal/core"
	"github.com/CiaranMcAleer/AgentLint/internal/languages/reactnative/rules"
)

// Parser is the JavaScript/TypeScript file parser
type Parser struct {
	config core.Config
	cache  *Cache

	funcPattern       *regexp.Regexp
	arrowFuncPattern  *regexp.Regexp
	classPattern      *regexp.Regexp
	importPattern     *regexp.Regexp
	exportPattern     *regexp.Regexp
	constPattern      *regexp.Regexp
	letPattern        *regexp.Regexp
	varPattern        *regexp.Regexp
	componentPattern  *regexp.Regexp
	lineCommentPattern *regexp.Regexp
	blockCommentStart *regexp.Regexp
	blockCommentEnd   *regexp.Regexp
}

func NewParser(config core.Config) *Parser {
	return &Parser{
		config:            config,
		cache:             NewCache(0),
		funcPattern:       regexp.MustCompile(`^(\s*)(?:async\s+)?function\s+(\w+)\s*\(`),
		arrowFuncPattern:  regexp.MustCompile(`^(\s*)(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?(?:\([^)]*\)|[\w]+)\s*=>`),
		classPattern:      regexp.MustCompile(`^(\s*)(?:export\s+)?(?:default\s+)?class\s+(\w+)(?:\s+extends\s+(\w+))?`),
		importPattern:     regexp.MustCompile(`^import\s+(.+)\s+from\s+['"]([^'"]+)['"]`),
		exportPattern:     regexp.MustCompile(`^export\s+(?:(default)\s+)?(?:const|let|var|function|class)\s*(\w*)`),
		constPattern:      regexp.MustCompile(`^(\s*)const\s+(\w+)\s*=`),
		letPattern:        regexp.MustCompile(`^(\s*)let\s+(\w+)\s*=`),
		varPattern:        regexp.MustCompile(`^(\s*)var\s+(\w+)\s*=`),
		componentPattern:  regexp.MustCompile(`(?:function|const)\s+([A-Z]\w+)`),
		lineCommentPattern: regexp.MustCompile(`^\s*//`),
		blockCommentStart: regexp.MustCompile(`/\*`),
		blockCommentEnd:   regexp.MustCompile(`\*/`),
	}
}

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
	state := &parseState{}

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

func (p *Parser) newParsedFile() *ParsedFile {
	return &ParsedFile{
		Lines:      make([]string, 0),
		Functions:  make([]FunctionDef, 0),
		Classes:    make([]ClassDef, 0),
		Components: make([]ComponentDef, 0),
		Imports:    make([]ImportStmt, 0),
		Exports:    make([]ExportStmt, 0),
		Comments:   make([]Comment, 0),
		Variables:  make([]VariableDef, 0),
	}
}

type parseState struct {
	lineNum        int
	inBlockComment bool
	braceDepth     int
}

func (p *Parser) processLine(line string, state *parseState, parsed *ParsedFile) {
	trimmed := strings.TrimSpace(line)

	if trimmed == "" {
		parsed.BlankLines++
		return
	}

	if p.handleBlockComment(line, trimmed, state, parsed) {
		return
	}

	if p.handleLineComment(trimmed, state, parsed) {
		return
	}

	p.handleInlineComment(line, state, parsed)

	if p.handleImport(line, state, parsed) {
		return
	}

	if p.handleExport(line, state, parsed) {
		return
	}

	if p.handleClass(line, state, parsed) {
		return
	}

	if p.handleFunction(line, state, parsed) {
		return
	}

	if p.handleArrowFunction(line, state, parsed) {
		return
	}

	p.handleVariable(line, state, parsed)
	parsed.CodeLines++
}

func (p *Parser) handleBlockComment(line, trimmed string, state *parseState, parsed *ParsedFile) bool {
	if state.inBlockComment {
		parsed.CommentLines++
		if p.blockCommentEnd.MatchString(line) {
			state.inBlockComment = false
		}
		return true
	}

	if p.blockCommentStart.MatchString(trimmed) && !p.blockCommentEnd.MatchString(trimmed) {
		state.inBlockComment = true
		parsed.Comments = append(parsed.Comments, Comment{
			Text:    trimmed,
			Line:    state.lineNum,
			IsBlock: true,
			IsJSDoc: strings.HasPrefix(trimmed, "/**"),
		})
		parsed.CommentLines++
		return true
	}

	return false
}

func (p *Parser) handleLineComment(trimmed string, state *parseState, parsed *ParsedFile) bool {
	if p.lineCommentPattern.MatchString(trimmed) {
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

func (p *Parser) handleInlineComment(line string, state *parseState, parsed *ParsedFile) {
	if idx := strings.Index(line, "//"); idx > 0 {
		commentText := strings.TrimSpace(line[idx:])
		parsed.Comments = append(parsed.Comments, Comment{
			Text:     commentText,
			Line:     state.lineNum,
			IsInline: true,
		})
	}
}

func (p *Parser) handleImport(line string, state *parseState, parsed *ParsedFile) bool {
	if !strings.HasPrefix(strings.TrimSpace(line), "import") {
		return false
	}

	matches := p.importPattern.FindStringSubmatch(line)
	if matches == nil {
		return false
	}

	importSpec := matches[1]
	module := matches[2]

	isDefault := !strings.Contains(importSpec, "{")
	var names []string
	if strings.Contains(importSpec, "{") {
		start := strings.Index(importSpec, "{")
		end := strings.Index(importSpec, "}")
		if start != -1 && end != -1 {
			namesStr := importSpec[start+1 : end]
			for _, n := range strings.Split(namesStr, ",") {
				names = append(names, strings.TrimSpace(n))
			}
		}
	}

	parsed.Imports = append(parsed.Imports, ImportStmt{
		Module:    module,
		Names:     names,
		IsDefault: isDefault,
		IsNamed:   !isDefault,
		Line:      state.lineNum,
	})

	return true
}

func (p *Parser) handleExport(line string, state *parseState, parsed *ParsedFile) bool {
	matches := p.exportPattern.FindStringSubmatch(line)
	if matches == nil {
		return false
	}

	parsed.Exports = append(parsed.Exports, ExportStmt{
		Name:      matches[2],
		IsDefault: matches[1] == "default",
		Line:      state.lineNum,
	})

	return false // Continue processing for function/class
}

func (p *Parser) handleClass(line string, state *parseState, parsed *ParsedFile) bool {
	matches := p.classPattern.FindStringSubmatch(line)
	if matches == nil {
		return false
	}

	className := matches[2]
	extends := ""
	if len(matches) > 3 {
		extends = matches[3]
	}

	isExported := strings.Contains(line, "export")

	parsed.Classes = append(parsed.Classes, ClassDef{
		Name:       className,
		StartLine:  state.lineNum,
		Extends:    extends,
		IsExported: isExported,
	})

	// Check if it's a React component
	if extends == "Component" || extends == "PureComponent" || extends == "React.Component" {
		parsed.Components = append(parsed.Components, ComponentDef{
			Name:       className,
			StartLine:  state.lineNum,
			IsClass:    true,
			IsExported: isExported,
		})
	}

	return true
}

func (p *Parser) handleFunction(line string, state *parseState, parsed *ParsedFile) bool {
	matches := p.funcPattern.FindStringSubmatch(line)
	if matches == nil {
		return false
	}

	indent := len(matches[1])
	funcName := matches[2]
	isAsync := strings.Contains(line, "async")
	isExported := strings.Contains(line, "export")

	parsed.Functions = append(parsed.Functions, FunctionDef{
		Name:       funcName,
		StartLine:  state.lineNum,
		IsAsync:    isAsync,
		IsExported: isExported,
		Indent:     indent,
	})

	// Check if it's a functional component (starts with uppercase)
	if len(funcName) > 0 && funcName[0] >= 'A' && funcName[0] <= 'Z' {
		parsed.Components = append(parsed.Components, ComponentDef{
			Name:         funcName,
			StartLine:    state.lineNum,
			IsFunctional: true,
			IsExported:   isExported,
		})
	}

	return true
}

func (p *Parser) handleArrowFunction(line string, state *parseState, parsed *ParsedFile) bool {
	matches := p.arrowFuncPattern.FindStringSubmatch(line)
	if matches == nil {
		return false
	}

	indent := len(matches[1])
	funcName := matches[2]
	isAsync := strings.Contains(line, "async")
	isExported := strings.Contains(line, "export")

	parsed.Functions = append(parsed.Functions, FunctionDef{
		Name:       funcName,
		StartLine:  state.lineNum,
		IsAsync:    isAsync,
		IsArrow:    true,
		IsExported: isExported,
		Indent:     indent,
	})

	// Check if it's a functional component
	if len(funcName) > 0 && funcName[0] >= 'A' && funcName[0] <= 'Z' {
		hasHooks := strings.Contains(line, "useState") || strings.Contains(line, "useEffect")
		parsed.Components = append(parsed.Components, ComponentDef{
			Name:         funcName,
			StartLine:    state.lineNum,
			IsFunctional: true,
			IsExported:   isExported,
			HasHooks:     hasHooks,
		})
	}

	return true
}

func (p *Parser) handleVariable(line string, state *parseState, parsed *ParsedFile) {
	var matches []string
	var kind string

	if matches = p.constPattern.FindStringSubmatch(line); matches != nil {
		kind = "const"
	} else if matches = p.letPattern.FindStringSubmatch(line); matches != nil {
		kind = "let"
	} else if matches = p.varPattern.FindStringSubmatch(line); matches != nil {
		kind = "var"
	}

	if matches == nil {
		return
	}

	indent := len(matches[1])
	if indent > 0 {
		return // Skip local variables
	}

	varName := matches[2]
	isExported := strings.Contains(line, "export")

	parsed.Variables = append(parsed.Variables, VariableDef{
		Name:       varName,
		Line:       state.lineNum,
		Kind:       kind,
		IsExported: isExported,
	})
}

func (p *Parser) calculateFunctionEndLines(parsed *ParsedFile) {
	for i := range parsed.Functions {
		fn := &parsed.Functions[i]
		braceCount := 0
		started := false

		for j := fn.StartLine - 1; j < len(parsed.Lines); j++ {
			line := parsed.Lines[j]
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")

			if strings.Contains(line, "{") {
				started = true
			}

			if started && braceCount <= 0 {
				fn.EndLine = j + 1
				break
			}
		}

		if fn.EndLine == 0 {
			fn.EndLine = len(parsed.Lines)
		}
	}
}

func (p *Parser) CalculateFileMetrics(ctx context.Context, filePath string, parsed *ParsedFile) *rules.FileMetrics {
	var commentRatio float64
	if parsed.CodeLines > 0 {
		commentRatio = float64(parsed.CommentLines) / float64(parsed.CodeLines)
	}

	return &rules.FileMetrics{
		Path:           filePath,
		TotalLines:     parsed.TotalLines,
		CodeLines:      parsed.CodeLines,
		CommentLines:   parsed.CommentLines,
		BlankLines:     parsed.BlankLines,
		CommentRatio:   commentRatio,
		FunctionCount:  len(parsed.Functions),
		ImportCount:    len(parsed.Imports),
		ClassCount:     len(parsed.Classes),
		ComponentCount: len(parsed.Components),
	}
}

func (p *Parser) CalculateFunctionMetrics(ctx context.Context, parsed *ParsedFile) []*rules.FunctionMetrics {
	metrics := make([]*rules.FunctionMetrics, 0, len(parsed.Functions))

	for _, fn := range parsed.Functions {
		lineCount := fn.EndLine - fn.StartLine
		if lineCount < 0 {
			lineCount = 0
		}

		metrics = append(metrics, &rules.FunctionMetrics{
			Name:       fn.Name,
			IsMethod:   fn.IsMethod,
			ClassName:  fn.ClassName,
			IsAsync:    fn.IsAsync,
			IsArrow:    fn.IsArrow,
			IsExported: fn.IsExported,
			LineCount:  lineCount,
			StartLine:  fn.StartLine,
		})
	}

	return metrics
}
