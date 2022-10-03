package irr

type (
	IWarnLogger  interface{ Warn(args ...any) }
	IErrorLogger interface{ Error(args ...any) }
	IFatalLogger interface{ Fatal(args ...any) }
)
