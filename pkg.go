package irr

// createTraceInfo will skip some stack layer corresponding to the method it self
// Thus when implement a func `A` which called createTraceInfo, the stack head will
// be `A` when skip= 1 are set, and this is the most general situation.
// There are another cases that use skip > 1, for example, when implement some
// basic lib, you may need the stack starts at a frontier caller.
func createTraceInfo(skip int, innerErr error) *traceInfo {
	t := generateStackTrace(1 + skip)
	if innerErr == nil {
		return t
	}
	if irr, ok := innerErr.(IRR); !ok || irr.GetTraceInfo() == nil || *(irr.GetTraceInfo()) != *t {
		return t
	}
	return nil
}

func Error(formatOrMsg string, args ...interface{}) IRR {
	err := newIrr(formatOrMsg, args...)
	return err
}

func Wrap(innerErr error, formatOrMsg string, args ...interface{}) IRR {
	err := newIrr(formatOrMsg, args...)
	err.inner = innerErr
	return err
}

func TraceSkip(skip int, formatOrMsg string, args ...interface{}) IRR {
	err := newIrr(formatOrMsg, args...)
	err.Trace = createTraceInfo(skip + 1, nil)
	return err
}

func Trace(formatOrMsg string, args ...interface{}) IRR {
	return TraceSkip(1, formatOrMsg, args ...)
}

func TrackSkip(skip int, innerErr error, formatOrMsg string, args ...interface{}) *irr {
	err := newIrr(formatOrMsg, args...)
	err.inner = innerErr
	err.Trace = createTraceInfo(skip + 1, innerErr)
	return err
}


func Track(innerErr error, formatOrMsg string, args ...interface{}) *irr {
	return TrackSkip(1, innerErr, formatOrMsg, args...)
}
