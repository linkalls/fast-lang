package linter

// Issue represents a single linting issue found.
type Issue struct {
	Filepath string // The path to the file where the issue was found.
	Line     int    // The line number of the issue.
	Column   int    // The column number of the issue (can be 0 if not applicable).
	RuleName string // The name of the rule that was violated.
	Message  string // A descriptive message for the issue.
	// Severity string // e.g., "error", "warning", "info" (optional for now, can default to warning)
}
