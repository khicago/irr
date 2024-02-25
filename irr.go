package irr

// ErrorC creates a new IRR error object with an error code and a formatted message.
// code is an int64 type error code used to identify and classify errors.
// formatOrMsg is a string that accepts printf style format specifiers to generate the error message.
// args are variadic parameters that represent the arguments for the formatting string.
// It returns an IRR error object set with the specific error code.
//
// Usage example:
//
// // Define a sample error code
// const ErrCodeInvalidInput int64 = 1001
//
// // ValidateInput checks the input string and returns an error if it is empty
//
//	func ValidateInput(input string) error {
//	    if input == "" {
//	        // Create an error object with a specific error code and formatted message using ErrorC
//	        return irr.ErrorC(ErrCodeInvalidInput, "validation failed: %s", "input cannot be empty")
//	    }
//	    // Other input validation logic...
//	    return nil
//	}
//
// Note: ErrorC is typically used when you wish to categorize errors or define specific status codes for easier error handling and response later on.
func ErrorC[T int64](code T, formatOrMsg string, args ...any) IRR {
	err := newBasicIrr(formatOrMsg, args...)
	return err.SetCode(int64(code))
}

// Error creates a new IRR error object with a formatted message.
// formatOrMsg is a string that accepts printf style format specifiers.
// args are variadic parameters that represent the arguments for the formatting string.
func Error(formatOrMsg string, args ...any) IRR {
	err := newBasicIrr(formatOrMsg, args...)
	return err
}

// Wrap wraps an existing error object with a given message and an inner error.
// innerErr is the error being wrapped.
// formatOrMsg is a string that accepts printf style format specifiers.
// args are variadic parameters that represent the arguments for the formatting string.
func Wrap(innerErr error, formatOrMsg string, args ...any) IRR {
	err := newBasicIrr(formatOrMsg, args...)
	err.inner = innerErr
	return err
}

// TraceSkip creates an error object with stack trace and formatted message, skipping a certain number of stack frames.
// skip indicates the number of call frames to skip in the stack trace.
// formatOrMsg is a string that accepts printf style format specifiers.
// args are variadic parameters that represent the arguments for the formatting string.
func TraceSkip(skip int, formatOrMsg string, args ...any) IRR {
	err := newBasicIrr(formatOrMsg, args...)
	err.Trace = createTraceInfo(skip+1, nil)
	return err
}

// Trace creates an error object with stack trace and a formatted message.
// formatOrMsg is a string that accepts printf style format specifiers.
// args are variadic parameters that represent the arguments for the formatting string.
// It defaults to skipping one call frame, usually the place where Trace is called.
func Trace(formatOrMsg string, args ...any) IRR {
	return TraceSkip(1, formatOrMsg, args...)
}

// TrackSkip creates an error object with a stack trace and wraps an inner error, skipping a specified number of stack frames.
// skip indicates the number of call frames to skip in the stack trace.
// innerErr is the error being wrapped.
// formatOrMsg is a string that accepts printf style format specifiers.
// args are variadic parameters that represent the arguments for the formatting string.
func TrackSkip(skip int, innerErr error, formatOrMsg string, args ...any) IRR {
	err := newBasicIrr(formatOrMsg, args...)
	err.inner = innerErr
	err.Trace = createTraceInfo(skip+1, innerErr)
	return err
}

// Track creates an error object with a stack trace and wraps an inner error.
// innerErr is the error being wrapped.
// formatOrMsg is a string that accepts printf style format specifiers.
// args are variadic parameters that represent the arguments for the formatting string.
// It defaults to skipping one call frame, starting the trace where Track is called.
func Track(innerErr error, formatOrMsg string, args ...any) IRR {
	return TrackSkip(1, innerErr, formatOrMsg, args...)
}

// CatchFailure is used to catch and handle panics within a function, preventing them from causing the program to crash while unifying the encapsulation of non-error information.
// It is declared at the beginning of a function with the defer keyword, ensuring that any panic during function execution can be caught.
// This function takes a callback function as a parameter, which is called when a panic occurs to handle the recovered error.
//
// Usage example:
//
// // A sample function that may cause panic
//
//	func riskyOperation() (err error) {
//	    // Defer calling CatchFailure at the start of riskyOperation
//	    // to ensure any subsequent panics can be caught and handled
//	    defer irr.CatchFailure(func(e error) {
//	        // Convert the recovered panic into a regular error so the function can return it
//	        // err can be set as a side effect, or the caught e can be handled directly (e.g., logging)
//	        // If the panic parameter is nil, e will be nil
//	        // If the panic is triggered with an error, the corresponding err will be passed directly
//	        // If the panic is another value, ErrUntypedExecutionFailure will be passed in, with the panic value attached to the error message
//	        err = e
//	    })
//
//	    // Trigger an out-of-bounds error that will cause a panic
//	    _ = make([]int, 0)[1]
//	    // Due to the panic above, the following code will not execute
//	    fmt.Println("This line of code will not be executed.")
//
//	    // If there is no panic, the function will return a nil error
//	    return nil
//	}
//
// // Calling riskyOperation elsewhere, handling errors returned by it
//
//	func main() {
//	    if err := riskyOperation(); err != nil {
//	        fmt.Printf("Caught error: %v\n", err)
//	    } else {
//	        fmt.Println("Operation successful, no errors occurred")
//	    }
//	}
//
// Note: CatchFailure should only be used to deal with panics caused by unforeseen situations, while regular error handling should be done using the error.
func CatchFailure(set func(err error)) {
	r := recover()

	if r == nil {
		set(nil)
		return
	}

	if e, ok := r.(error); ok {
		set(e)
		return
	}
	set(Wrap(ErrUntypedExecutionFailure, "panic = %v", r))
}

//func (ir *BasicIrr) Format(s fmt.State, verb rune) {
//	switch verb {
//	case 'q':
//		io.WriteString(s, fmt.Sprintf("%q", ir.Error()))
//	case 's', 'v':
//		io.WriteString(s, ir.Error())
//	}
//}
