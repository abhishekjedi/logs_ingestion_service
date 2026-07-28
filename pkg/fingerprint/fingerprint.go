package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"

	dbdto "error-logging/db/dto"
)

const topFrames = 5

var (
	reQuoted = regexp.MustCompile(`'[^']*'|"[^"]*"`)
	reUUID   = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	reHex    = regexp.MustCompile(`0x[0-9a-fA-F]+`)
	reNum    = regexp.MustCompile(`\d+`)

	reFrameJS = regexp.MustCompile(`at\s+(\S+)\s+\(([\w./\\-]+\.\w+):(\d+)`)

	reFrame = regexp.MustCompile(`([\w./\\-]+\.\w+):(\d+)`)

	reFramePy = regexp.MustCompile(`File "([^"]+)", line (\d+)(?:, in (\S+))?`)
)

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

func NormalizeMessage(msg string) string {
	msg = reQuoted.ReplaceAllString(msg, "<str>")
	msg = reUUID.ReplaceAllString(msg, "<uuid>")
	msg = reHex.ReplaceAllString(msg, "<hex>")
	msg = reNum.ReplaceAllString(msg, "<n>")
	return strings.TrimSpace(msg)
}

func ParseStacktrace(raw string) []dbdto.StackFrame {
	if raw == "" {
		return nil
	}
	var frames []dbdto.StackFrame

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

	for _, m := range reFrameJS.FindAllStringSubmatch(raw, -1) {
		line, _ := strconv.ParseUint(m[3], 10, 32)
		frames = append(frames, dbdto.StackFrame{Function: m[1], File: m[2], Line: uint32(line), InApp: true})
	}
	if len(frames) > 0 {
		return frames
	}

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
