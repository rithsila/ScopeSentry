package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// ParseCommaSeparated splits a comma-separated string into trimmed non-empty tokens.
func ParseCommaSeparated(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}

// ParseJSONOrFile parses either a raw JSON string, or if prefixed with @ or is a file, reads the file.
// If s is "-", it reads from os.Stdin.
func ParseJSONOrFile(s string) (any, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	var data []byte
	var err error

	if s == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}
	} else if strings.HasPrefix(s, "@") {
		filePath := strings.TrimPrefix(s, "@")
		data, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
	} else if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		data = []byte(s)
	} else if _, statErr := os.Stat(s); statErr == nil {
		data, err = os.ReadFile(s)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", s, err)
		}
	} else {
		data = []byte(s)
	}

	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	return parsed, nil
}

// ParseFilterMap parses a string into map[string][]string.
// Accepts JSON map: {"project":["id1","id2"], "task":["name1"]} or comma lists.
func ParseFilterMap(s string) (map[string][]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return nil, fmt.Errorf("invalid filter json: %w", err)
	}

	res := make(map[string][]string, len(raw))
	for k, v := range raw {
		switch val := v.(type) {
		case []any:
			strs := make([]string, 0, len(val))
			for _, item := range val {
				strs = append(strs, fmt.Sprint(item))
			}
			res[k] = strs
		case []string:
			res[k] = val
		case string:
			res[k] = ParseCommaSeparated(val)
		default:
			res[k] = []string{fmt.Sprint(val)}
		}
	}
	return res, nil
}

// ParseSortMap parses sort string like `length:asc` or `{"length":"ascending"}`.
func ParseSortMap(s string) (map[string]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	var m map[string]string
	if err := json.Unmarshal([]byte(s), &m); err == nil {
		return m, nil
	}

	parts := strings.Split(s, ":")
	if len(parts) == 2 {
		return map[string]string{
			strings.TrimSpace(parts[0]): strings.TrimSpace(parts[1]),
		}, nil
	}

	return nil, fmt.Errorf("invalid sort format, expected JSON or key:direction (e.g. length:ascending)")
}

// ParseFlags reorders positional arguments after flags so flag.FlagSet parses all flags
// regardless of where non-flag positional arguments were specified on the CLI.
func ParseFlags(fs *flag.FlagSet, args []string) error {
	var flags []string
	var positionals []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				name := strings.TrimLeft(arg, "-")
				f := fs.Lookup(name)
				if f != nil {
					if b, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && b.IsBoolFlag() {
						continue
					}
					flags = append(flags, args[i+1])
					i++
					continue
				}
			}
		} else {
			positionals = append(positionals, arg)
		}
	}

	reordered := append(flags, positionals...)
	return fs.Parse(reordered)
}
