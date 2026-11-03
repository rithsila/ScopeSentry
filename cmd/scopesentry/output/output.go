package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	ExitSuccess        = 0
	ExitExecutionError = 1
	ExitUsageError     = 2
	ExitTimeout        = 3
)

// PrintJSON formats any Go value as indented JSON and writes to stdout.
func PrintJSON(w io.Writer, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

// PrintRawJSON takes a string, verifies/indents if valid JSON, and writes to stdout.
func PrintRawJSON(w io.Writer, s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		_, err := fmt.Fprintln(w, "{}")
		return err
	}

	var parsed any
	if err := json.Unmarshal([]byte(s), &parsed); err == nil {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		if err := enc.Encode(parsed); err == nil {
			_, err = fmt.Fprint(w, buf.String())
			return err
		}
	}

	// If not JSON object, print as-is
	_, err := fmt.Fprintln(w, s)
	return err
}

// PrintDiagnostic writes informative messages to stderr.
func PrintDiagnostic(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, format+"\n", args...)
}

// PrintError writes error messages to stderr.
func PrintError(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "error: "+format+"\n", args...)
}
