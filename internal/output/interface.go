package output

import "github.com/agentlint/agentlint/internal/core"

// Formatter interface for output formatters
type Formatter interface {
	Format(results []core.Result) error
	FormatError(err error) error
	PrintHeader()
	PrintFooter()
}