// Package irc provides a customized error code system extending the IRR library.
// It defines a `Code` type which represents an error code as an int64,
// and a set of methods that produce rich errors containing both
// formatted messages and these error codes. The package follows the best
// practices of error handling in Go by leveraging the abilities of the IRR library
// to capture stack traces, wrap errors, and provide context while allowing
// for the classification and handling of errors through unified error codes.
//
// Best practices:
//   - Define domain-specific error codes as constants using the `Code` type
//     to maintain a registry of error codes for the application.
//   - Use the `Code`-related methods to create errors with consistent formatting,
//     additional context, and these predefined error codes.
//   - Wrap errors when catching them to maintain the original error context, providing
//     a clear error chain.
//   - Provide stack traces only when it's necessary for debugging, to avoid overpopulating
//     logs with unnecessary details.
//   - Utilize error codes to handle errors gracefully within application logic or
//     when communicating with clients via API responses.
package irc

import (
	"github.com/khicago/irr"
	"strconv"
)

type (
	// Code defines an int64 type for error codes.
	// It extends the error interface by allowing the association of an error message with a code.
	Code int64

	// ICodeGetter is an interface for types that can return an error code.
	ICodeGetter irr.ICodeGetter[int64]

	// ICodeTraverse is an interface for traversing codes in an error chain.
	ICodeTraverse irr.ITraverseCoder[int64]
)

// Verify that Code implements the irr.Spawner interface.
var _ irr.Spawner = Code(0)

// I64 converts a Code to its int64 representation.
func (c Code) I64() int64 {
	return int64(c)
}

// String converts a Code to its string representation, typically for printing.
func (c Code) String() string {
	return strconv.FormatInt(c.I64(), 10)
}

// Error creates an IRR error object with a formatted message and sets the error code.
func (c Code) Error(formatOrMsg string, args ...interface{}) irr.IRR {
	return irr.Error(formatOrMsg, args...).SetCode(c.I64())
}

// Wrap wraps an existing error object with a formatted message and sets the error code.
func (c Code) Wrap(innerErr error, formatOrMsg string, args ...interface{}) irr.IRR {
	return irr.Wrap(innerErr, formatOrMsg, args...).SetCode(c.I64())
}

// TraceSkip creates an IRR error object with stack trace information, skipping the specified
// number of stack frames, and sets the error code.
func (c Code) TraceSkip(skip int, formatOrMsg string, args ...interface{}) irr.IRR {
	return irr.TraceSkip(skip, formatOrMsg, args...).SetCode(c.I64())
}

// Trace creates an IRR error object with stack trace information and sets the error code.
func (c Code) Trace(formatOrMsg string, args ...interface{}) irr.IRR {
	return irr.Trace(formatOrMsg, args...).SetCode(c.I64())
}

// TrackSkip creates an IRR error object that wraps an inner error with stack trace information,
// skipping the specified number of frames, and sets the error code.
func (c Code) TrackSkip(skip int, innerErr error, formatOrMsg string, args ...interface{}) irr.IRR {
	return irr.TrackSkip(skip, innerErr, formatOrMsg, args...).SetCode(c.I64())
}

// Track creates an IRR error object that wraps an inner error with stack trace information and
// sets the error code.
func (c Code) Track(innerErr error, formatOrMsg string, args ...interface{}) irr.IRR {
	return irr.Track(innerErr, formatOrMsg, args...).SetCode(c.I64())
}
