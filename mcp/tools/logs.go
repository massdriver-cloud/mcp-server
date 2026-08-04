package tools

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// defaultLogTailBytes caps get_deployment_logs output when the caller doesn't
// pass tail_lines. Provisioner logs for a single deployment can run to
// hundreds of KB of provider noise, which overflows MCP clients' tool-result
// limits; the tail keeps the plan summary and errors (both emitted at the
// end) while staying well under those limits.
const defaultLogTailBytes = 40 * 1024

// tailLogs bounds logs for tool output. tailLines < 0 returns everything;
// tailLines > 0 returns exactly the last N lines; tailLines == 0 (omitted by
// the caller) returns the last defaultLogTailBytes bytes, cut at a line
// boundary. Log content is never rewritten — the returned text (after the
// truncation note, when present) is always a verbatim suffix of the input.
func tailLogs(logs string, tailLines int) string {
	switch {
	case tailLines < 0 || logs == "":
		return logs

	case tailLines > 0:
		lines := strings.Split(strings.TrimSuffix(logs, "\n"), "\n")
		if len(lines) <= tailLines {
			return logs
		}
		note := fmt.Sprintf("(showing the last %d of %d log lines; pass tail_lines=-1 for the complete log)\n", tailLines, len(lines))
		return note + strings.Join(lines[len(lines)-tailLines:], "\n") + "\n"

	default:
		if len(logs) <= defaultLogTailBytes {
			return logs
		}
		tail := logs[len(logs)-defaultLogTailBytes:]
		// Start the tail on a whole line when one is available; otherwise
		// (a single line longer than the cap) start on a rune boundary.
		if i := strings.IndexByte(tail, '\n'); i >= 0 && i+1 < len(tail) {
			tail = tail[i+1:]
		} else {
			for len(tail) > 0 && !utf8.RuneStart(tail[0]) {
				tail = tail[1:]
			}
		}
		note := fmt.Sprintf("(showing the last %d bytes of %d total; pass tail_lines=-1 for the complete log)\n", len(tail), len(logs))
		return note + tail
	}
}
