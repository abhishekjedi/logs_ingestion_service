package fingerprint

import (
	"testing"

	dbdto "error-logging/db/dto"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeMessage(t *testing.T) {
	// Messages differing only by dynamic tokens normalize to the same template.
	a := NormalizeMessage(`cannot find user 12345 with token "abc-def"`)
	b := NormalizeMessage(`cannot find user 99 with token "xyz"`)
	assert.Equal(t, a, b)
	assert.Contains(t, a, "<n>")
	assert.Contains(t, a, "<str>")
}

func TestCompute_StableAcrossDynamicValues(t *testing.T) {
	f1 := Compute("TypeError", "null id 42", nil, "")
	f2 := Compute("TypeError", "null id 7", nil, "")
	assert.Equal(t, f1, f2, "same error, different number → same fingerprint")

	f3 := Compute("ValueError", "null id 42", nil, "")
	assert.NotEqual(t, f1, f3, "different exception type → different fingerprint")
}

func TestCompute_HintOverrides(t *testing.T) {
	withHint := Compute("TypeError", "anything", nil, "custom-group")
	sameHint := Compute("OtherError", "different", nil, "custom-group")
	assert.Equal(t, withHint, sameHint, "explicit hint groups regardless of other fields")
}

func TestCompute_FramesAffectGrouping(t *testing.T) {
	base := Compute("Err", "msg", nil, "")
	withFrame := Compute("Err", "msg", []dbdto.StackFrame{{File: "main.go", Function: "run"}}, "")
	assert.NotEqual(t, base, withFrame)
}

func TestParseStacktrace(t *testing.T) {
	raw := "Traceback:\n  File \"app/handler.py\", line 42, in handle\n  File \"app/db.py\", line 7, in query"
	frames := ParseStacktrace(raw)
	assert.Len(t, frames, 2)
	assert.Equal(t, "app/handler.py", frames[0].File)
	assert.Equal(t, uint32(42), frames[0].Line)
	assert.Empty(t, ParseStacktrace(""))
}
