package irr

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	err := Error("%d %d", 1, 2)
	assert.Equal(t, "1 2", err.Error(), "err message unmatched")
}

func TestWrap(t *testing.T) {
	inner := Error("%d %d", 1, 2)
	ir := Wrap(inner, "wrap msg")
	assert.Equal(t, "wrap msg, 1 2", ir.Error(), "err message unmatched")
}

func TestTrace(t *testing.T) {
	ir := Trace("%d %d", 1, 2)
	tracePrint := ir.ToString(true, ";")
	prefix := "1 2 irr.TestTrace"
	assert.Exactly(t, true,
		tracePrint[:len(prefix)] == prefix &&
			strings.Contains(tracePrint, "/irr_test.go:"), tracePrint)
}

func TestTraceSkip(t *testing.T) {
	ir := TraceSkip(1, "%d %d", 1, 2)
	tracePrint := ir.ToString(true, ";")
	prefix := "1 2 testing.tRunner"
	assert.Exactly(t, true,
		tracePrint[:len(prefix)] == prefix &&
			strings.Contains(tracePrint, "/testing.go:"), tracePrint)
}

func TestTrack(t *testing.T) {
	inner := Trace("%d %d", 1, 2)
	ir := Track(inner, "%d %d", 1, 2)
	tracePrint := ir.ToString(true, "\n")
	traceOut := strings.Split(tracePrint, "\n")

	assert.Equal(t, 2, len(traceOut), "trace out length not match, "+tracePrint)

	prefix := "1 2 irr.TestTrack"
	assert.Exactly(t, true,
		traceOut[0][:len(prefix)] == prefix &&
			strings.Contains(traceOut[0], "/irr_test.go:"), "traceOut[0] not match\n"+tracePrint)
	assert.Exactly(t, true,
		traceOut[1][:len(prefix)] == prefix &&
			strings.Contains(traceOut[1], "/irr_test.go:"), "traceOut[1] not match\n"+tracePrint)
}

func TestTrackSkip(t *testing.T) {
	inner := Trace("%d %d", 1, 2)
	ir := TrackSkip(1, inner, "%d %d", 1, 2)
	tracePrint := ir.ToString(true, "\n")
	traceOut := strings.Split(tracePrint, "\n")

	assert.Equal(t, 2, len(traceOut), "trace out length not match\n"+tracePrint)

	prefix, prefixOuter := "1 2 irr.TestTrackSkip", "1 2 testing.tRunner"
	assert.Exactly(t, true,
		traceOut[0][:len(prefixOuter)] == prefixOuter &&
			strings.Contains(traceOut[0], "/testing.go:"), "outer trace not match\n"+tracePrint)
	assert.Exactly(t, true,
		traceOut[1][:len(prefix)] == prefix &&
			strings.Contains(traceOut[1], "/irr_test.go:"), "inner trace not match\n"+tracePrint)
}
