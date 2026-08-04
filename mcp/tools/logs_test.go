package tools

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTailLogs(t *testing.T) {
	bigLog := func(lines int) string {
		var sb strings.Builder
		for i := 1; i <= lines; i++ {
			fmt.Fprintf(&sb, "line %d\n", i)
		}
		return sb.String()
	}

	t.Run("negative returns everything", func(t *testing.T) {
		logs := bigLog(20000)
		if got := tailLogs(logs, -1); got != logs {
			t.Errorf("tail_lines=-1 modified the log")
		}
	})

	t.Run("short log passes through untouched", func(t *testing.T) {
		logs := bigLog(10)
		if got := tailLogs(logs, 0); got != logs {
			t.Errorf("default cap truncated a short log: %q", got)
		}
	})

	t.Run("default cap truncates long log with note, keeping a verbatim suffix", func(t *testing.T) {
		logs := bigLog(20000)
		got := tailLogs(logs, 0)
		if !strings.HasPrefix(got, "(showing the last ") {
			t.Fatalf("missing truncation note, got prefix: %q", got[:60])
		}
		body := got[strings.IndexByte(got, '\n')+1:]
		if !strings.HasSuffix(logs, body) {
			t.Errorf("truncated output is not a verbatim suffix of the input")
		}
		if !strings.HasPrefix(body, "line ") {
			t.Errorf("tail does not start on a whole line: %q", body[:20])
		}
		if len(body) > defaultLogTailBytes {
			t.Errorf("tail is %d bytes, want <= %d", len(body), defaultLogTailBytes)
		}
	})

	t.Run("single line exceeding the byte cap keeps its tail", func(t *testing.T) {
		logs := strings.Repeat("x", 100*1024) + "THE-END"
		got := tailLogs(logs, 0)
		if !strings.HasPrefix(got, "(showing the last ") {
			t.Errorf("missing truncation note, got prefix: %q", got[:60])
		}
		if !strings.HasSuffix(got, "THE-END") {
			t.Errorf("tail lost the end of the log")
		}
	})

	t.Run("byte truncation lands on a rune boundary", func(t *testing.T) {
		logs := strings.Repeat("héllo wörld ", 8*1024) // multibyte, ~100KB single line
		got := tailLogs(logs, 0)
		if !utf8.ValidString(got) {
			t.Errorf("byte-truncated output contains invalid UTF-8")
		}
	})

	t.Run("explicit tail_lines returns exactly N", func(t *testing.T) {
		got := tailLogs(bigLog(100), 3)
		want := "(showing the last 3 of 100 log lines; pass tail_lines=-1 for the complete log)\nline 98\nline 99\nline 100\n"
		if got != want {
			t.Errorf("tailLogs(_, 3) = %q, want %q", got, want)
		}
	})

	t.Run("explicit tail_lines larger than log is a no-op", func(t *testing.T) {
		logs := bigLog(5)
		if got := tailLogs(logs, 100); got != logs {
			t.Errorf("oversized tail_lines modified the log: %q", got)
		}
	})
}
