package irr

func CatchFailure(set func(err error)) {
	r := recover()
	if e, ok := r.(error); ok {
		set(e)
	}
	set(Wrap(ErrUntypedExecutionFailure, "err= %v", r))
}
