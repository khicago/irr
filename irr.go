package irr

//func (ir *BasicIrr) Format(s fmt.State, verb rune) {
//	switch verb {
//	case 'q':
//		io.WriteString(s, fmt.Sprintf("%q", ir.Error()))
//	case 's', 'v':
//		io.WriteString(s, ir.Error())
//	}
//}

func Error(formatOrMsg string, args ...interface{}) IRR {
	err := newBasicIrr(formatOrMsg, args...)
	return err
}

func Wrap(innerErr error, formatOrMsg string, args ...interface{}) IRR {
	err := newBasicIrr(formatOrMsg, args...)
	err.inner = innerErr
	return err
}

func TraceSkip(skip int, formatOrMsg string, args ...interface{}) IRR {
	err := newBasicIrr(formatOrMsg, args...)
	err.Trace = createTraceInfo(skip+1, nil)
	return err
}

func Trace(formatOrMsg string, args ...interface{}) IRR {
	return TraceSkip(1, formatOrMsg, args...)
}

func TrackSkip(skip int, innerErr error, formatOrMsg string, args ...interface{}) IRR {
	err := newBasicIrr(formatOrMsg, args...)
	err.inner = innerErr
	err.Trace = createTraceInfo(skip+1, innerErr)
	return err
}

func Track(innerErr error, formatOrMsg string, args ...interface{}) IRR {
	return TrackSkip(1, innerErr, formatOrMsg, args...)
}

func CatchFailure(set func(err error)) {
	r := recover()
	if r == nil {
		return
	}
	if e, ok := r.(error); ok {
		set(e)
		return
	}
	set(Wrap(ErrUntypedExecutionFailure, "err = %v", r))
}
