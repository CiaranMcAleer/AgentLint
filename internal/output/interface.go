package output

import "github.com/CiaranMcAleer/AgentLint/internal/core"

// Formatter interface for output formatters
type Formatter interface {
	Format(results []core.Result) error
	FormatError(err error) error
	PrintHeader()
	PrintFooter()
}
