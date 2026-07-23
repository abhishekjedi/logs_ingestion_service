// Package fingerprint groups error events: it normalizes an exception into a stable
// hash so the same logical error (ignoring dynamic tokens) maps to one issue, and
// parses raw stack traces into structured frames.
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"

	dbdto "error-logging/db/dto"
)

// topFrames bounds how many frames feed the fingerprint (deepest call sites first).
const topFrames = 5

var (
	reQuoted = regexp.MustCompile(`'[^']*'|"[^"]*"`)
	reUUID   = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	reHex    = regexp.MustCompile(`0x[0-9a-fA-F]+`)
	reNum    = regexp.MustCompile(`\d+`)
	// Go/JS style: "path/to/file.ext:123"
	reFrame = regexp.MustCompile(`([\w./\\-]+\.\w+):(\d+)`)
	// Python style: File "path/to/file.py", line 123
	reFramePy = regexp.MustCompile(`File "([^"]+)", line (\d+)`)
)

// Compute returns the fingerprint for an error. A caller-supplied hint (the
// log.fingerprint attribute) overrides automatic grouping. Otherwise the hash is
// built from the exception type, a normalized message template, and the top frames'
// file+function — so the same logical error groups despite dynamic values.
func Compute(exceptionType, message string, frames []dbdto.StackFrame, hint string) string {
	if hint != "" {
		return sum(hint)
	}

	var b strings.Builder
	b.WriteString(exceptionType)
	b.WriteByte('|')
	b.WriteString(NormalizeMessage(message))

	n := len(frames)
	if n > topFrames {
		n = topFrames
	}
	for _, f := range frames[:n] {
		b.WriteByte('|')
		b.WriteString(f.File)
		b.WriteByte(':')
		b.WriteString(f.Function)
	}

	return sum(b.String())
}

// NormalizeMessage strips dynamic tokens (quoted strings, UUIDs, hex, numbers) so
// messages differing only by runtime values collapse to one template.
func NormalizeMessage(msg string) string {
	msg = reQuoted.ReplaceAllString(msg, "<str>")
	msg = reUUID.ReplaceAllString(msg, "<uuid>")
	msg = reHex.ReplaceAllString(msg, "<hex>")
	msg = reNum.ReplaceAllString(msg, "<n>")
	return strings.TrimSpace(msg)
}

// ParseStacktrace best-effort-extracts structured frames from a raw stack trace.
// It is language-agnostic: it pulls out "file.ext:line" occurrences. Function names
// are left empty when not confidently identifiable.
func ParseStacktrace(raw string) []dbdto.StackFrame {
	if raw == "" {
		return nil
	}
	// Python "File \"x\", line N" first; if none, fall back to Go/JS "file.ext:line".
	matches := reFramePy.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		matches = reFrame.FindAllStringSubmatch(raw, -1)
	}

	var frames []dbdto.StackFrame
	for _, m := range matches {
		line, _ := strconv.ParseUint(m[2], 10, 32)
		frames = append(frames, dbdto.StackFrame{
			File:  m[1],
			Line:  uint32(line),
			InApp: true,
		})
	}
	return frames
}

func sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
