package irr

func CatchFailure(set func(err error)) {
	r := recover()
	if r == nil {
		return
	}
	if e, ok := r.(error); ok {
		set(e)
		return
	}
	set(Wrap(ErrUntypedExecutionFailure, "err= %v", r))
}
