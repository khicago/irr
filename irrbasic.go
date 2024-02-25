package irr

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type (
	BasicIrr struct {
		inner error

		Code  int64      `json:"code"`
		Tags  []string   `json:"tags"`
		Msg   string     `json:"msg"`
		Trace *traceInfo `json:"trace"`

		mu sync.RWMutex
	}
)

func newBasicIrr(formatOrMsg string, args ...any) *BasicIrr {
	err := &BasicIrr{}
	if len(args) > 0 {
		err.Msg = fmt.Sprintf(formatOrMsg, args...)
	} else {
		err.Msg = formatOrMsg
	}
	return err
}

var _ IRR = newBasicIrr("")

// Error
// the implementation of error
func (ir *BasicIrr) Error() string {
	return ir.ToString(false, ", ")
}

// Unwrap
// the implementation of IUnwrap
func (ir *BasicIrr) Unwrap() error {
	return ir.inner
}

// Root
// the implementation of ITraverseError
func (ir *BasicIrr) Root() error {
	var err error = ir
	for {
		inner := errors.Unwrap(err)
		if inner == nil {
			return err
		}
		err = inner
	}
}

// TraverseToRoot
// the implementation of ITraverseError
func (ir *BasicIrr) TraverseToRoot(fn func(err error) error) (err error) {
	defer CatchFailure(func(e error) { err = e })
	for inner := error(ir); inner != nil; inner = errors.Unwrap(inner) {
		if err = fn(inner); err != nil {
			return
		}
	}
	return
}

// Source
// the implementation of ITraverseIrr
func (ir *BasicIrr) Source() (err error) {
	_ = ir.TraverseToSource(func(e error, isSource bool) error {
		if isSource {
			err = e
		}
		return nil
	})
	return
}

// TraverseToSource
// the implementation of ITraverseIrr
func (ir *BasicIrr) TraverseToSource(fn func(err error, isSource bool) error) (err error) {
	defer CatchFailure(func(e error) { err = e })
	for cur := ir; cur != nil; {
		isCurSource := cur.inner == nil
		err = fn(cur, isCurSource)
		if isCurSource {
			break
		}
		if next, ok := cur.inner.(*BasicIrr); ok {
			cur = next
			continue
		}
		err = fn(cur.inner, true)
		break
	}
	return
}

// GetCodeStr
// Determines how the code is written to the message,
// so that this method can input an empty string to
// avoid outputting the code in the message
func (ir *BasicIrr) GetCodeStr() string {
	if ir.Code == 0 {
		return ""
	}
	return fmt.Sprintf("code(%d), ", ir.Code)
}

func (ir *BasicIrr) writeSelfTo(sb *strings.Builder, printTrace bool, printCode bool) {
	if printCode {
		if codeStr := ir.GetCodeStr(); codeStr != "" {
			sb.WriteString(codeStr)
		}
	}
	sb.WriteString(ir.Msg)

	if ir.Tags != nil && len(ir.Tags) > 0 {
		for _, str := range ir.Tags {
			sb.WriteRune('[')
			sb.WriteString(str)
			sb.WriteString("] ")
		}
	}
	if printTrace && ir.Trace != nil {
		sb.WriteRune(' ')
		ir.Trace.writeTo(sb)
	}
}

// ToString
// consecutive equal codes will be printed only once during the traceback process
func (ir *BasicIrr) ToString(printTrace bool, split string) string {
	sb := strings.Builder{}
	lastCode := int64(0)
	_ = ir.TraverseToSource(func(err error, isSource bool) error {
		if irr, ok := err.(*BasicIrr); ok {
			// since have to continue traversing, irr only output itself
			irr.writeSelfTo(&sb, printTrace, lastCode != irr.Code)
			lastCode = irr.Code
		} else {
			sb.WriteString(err.Error())
		}
		if !isSource {
			sb.WriteString(split)
		}
		return nil
	})
	return sb.String()
}

// LogWarn
// the implementation of ILogCaller
func (ir *BasicIrr) LogWarn(logger IWarnLogger) IRR {
	logger.Warn(ir.ToString(true, "\n"))
	return ir
}

// LogError
// the implementation of ILogCaller
func (ir *BasicIrr) LogError(logger IErrorLogger) IRR {
	logger.Error(ir.ToString(true, "\n"))
	return ir
}

// LogFatal
// the implementation of ILogCaller
func (ir *BasicIrr) LogFatal(logger IFatalLogger) IRR {
	str := ir.ToString(true, "\n")
	logger.Fatal(str)
	// to make sure it has been printed to std output stream
	fmt.Println(str)
	return ir
}

// SetCode
// the implementation of ICoder[int64]
func (ir *BasicIrr) SetCode(val int64) IRR {
	ir.Code = val
	return ir
}

// GetCode
// the implementation of ICoder[int64]
func (ir *BasicIrr) GetCode() (val int64) {
	return ir.Code
}

// ClosestCode
// the implementation of ITraverseCoder[int64]
func (ir *BasicIrr) ClosestCode() (val int64) {
	eExit := errors.New("stop")
	if err := ir.TraverseCode(func(_ error, code int64) error {
		if code != 0 {
			val = code
			return eExit
		}
		return nil
	}); err != nil && err != eExit {
		panic("traverse panic")
	}
	return val
}

// TraverseCode
// the implementation of ITraverseCoder[int64]
func (ir *BasicIrr) TraverseCode(fn func(err error, code int64) error) (err error) {
	return ir.TraverseToRoot(func(err error) error {
		if t, ok := err.(ICoder[int64]); ok {
			if err = fn(err, t.GetCode()); err != nil {
				return err
			}
		}
		return nil
	})
}

// SetTag
// the implementation of ITagger
func (ir *BasicIrr) SetTag(key, val string) {
	ir.mu.Lock()
	if ir.Tags == nil {
		ir.Tags = make([]string, 0)
	}
	ir.Tags = append(ir.Tags, fmt.Sprintf("%s:%s", key, val))
	ir.mu.Unlock()
}

// GetTag
// the implementation of ITagger
func (ir *BasicIrr) GetTag(key string) (val []string) {
	ir.mu.RLock()
	val = make([]string, 0)
	if ir.Tags == nil {
		ir.mu.RUnlock()
		return val
	}
	lenK := len(key)
	for _, str := range ir.Tags {
		if len(str) < lenK+2 {
			continue
		}
		if str[:lenK] == key {
			val = append(val, str[lenK+1:])
		}
	}
	ir.mu.RUnlock()
	return val
}

func (ir *BasicIrr) GetTraceInfo() *traceInfo {
	return ir.Trace
}
