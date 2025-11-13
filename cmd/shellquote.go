package cmd

import "strings"

// shellQuote wraps a value in single quotes and escapes any embedded single quotes.
// It allows us to safely interpolate user-supplied values in remote shell commands.
func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

// joinCommand builds a shell string by concatenating non-empty parts with spaces.
func joinCommand(parts ...string) string {
	trimmed := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		trimmed = append(trimmed, part)
	}
	return strings.Join(trimmed, " ")
}
