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
	reNum = regexp.MustCompile(`\d+`)
	// JS/Node with function: at Func.name (path/to/file.js:12:5)
	reFrameJS = regexp.MustCompile(`at\s+(\S+)\s+\(([\w./\\-]+\.\w+):(\d+)`)
	// Bare "path/to/file.ext:123" (Go, or JS without a named function)
	reFrame = regexp.MustCompile(`([\w./\\-]+\.\w+):(\d+)`)
	// Python: File "path/to/file.py", line 123, in func
	reFramePy = regexp.MustCompile(`File "([^"]+)", line (\d+)(?:, in (\S+))?`)
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
// It recognizes Python ("File ..., line N, in func"), JS/Node ("at func (file:line)")
// and bare "file.ext:line" forms, capturing function names where the format exposes
// them and leaving Function empty otherwise.
func ParseStacktrace(raw string) []dbdto.StackFrame {
	if raw == "" {
		return nil
	}
	var frames []dbdto.StackFrame

	// Python: File "path", line N, in func
	for _, m := range reFramePy.FindAllStringSubmatch(raw, -1) {
		line, _ := strconv.ParseUint(m[2], 10, 32)
		fn := ""
		if len(m) > 3 {
			fn = m[3]
		}
		frames = append(frames, dbdto.StackFrame{File: m[1], Function: fn, Line: uint32(line), InApp: true})
	}
	if len(frames) > 0 {
		return frames
	}

	// JS/Node: at func (file:line)
	for _, m := range reFrameJS.FindAllStringSubmatch(raw, -1) {
		line, _ := strconv.ParseUint(m[3], 10, 32)
		frames = append(frames, dbdto.StackFrame{Function: m[1], File: m[2], Line: uint32(line), InApp: true})
	}
	if len(frames) > 0 {
		return frames
	}

	// Bare file.ext:line
	for _, m := range reFrame.FindAllStringSubmatch(raw, -1) {
		line, _ := strconv.ParseUint(m[2], 10, 32)
		frames = append(frames, dbdto.StackFrame{File: m[1], Line: uint32(line), InApp: true})
	}
	return frames
}

func sum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
