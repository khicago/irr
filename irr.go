package irr

import (
	"errors"
	"fmt"
	"strings"
)

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

func (ir *BasicIrr) Source() (err error) {
	_ = ir.TraverseToSource(func(e error, isSource bool) error {
		if isSource {
			err = e
		}
		return nil
	})
	return
}

func (ir *BasicIrr) Unwrap() error {
	return ir.inner
}

func (ir *BasicIrr) GetTraceInfo() *traceInfo {
	return ir.Trace
}

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
		} else {
			err = fn(cur.inner, true)
			break
		}
	}
	return
}

func (ir *BasicIrr) TraverseToRoot(fn func(err error) error) (err error) {
	defer CatchFailure(func(e error) { err = e })
	var inner error = ir
	for inner != nil {
		err = fn(inner)
		inner = errors.Unwrap(inner)
	}
	return
}

func (ir *BasicIrr) writeSelfTo(sb *strings.Builder, printTrace bool) {
	sb.WriteString(ir.Msg)
	if printTrace && ir.Trace != nil {
		sb.WriteRune(' ')
		ir.Trace.writeTo(sb)
	}
}

func (ir *BasicIrr) ToString(printTrace bool, split string) string {
	sb := strings.Builder{}
	_ = ir.TraverseToSource(func(err error, isSource bool) error {
		if irr, ok := err.(*BasicIrr); ok {
			irr.writeSelfTo(&sb, printTrace)
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

func (ir *BasicIrr) Error() string {
	return ir.ToString(false, "; ")
}

func (ir *BasicIrr) LogWarn(logger interface{ Warn(args ...interface{}) }) IRR {
	logger.Warn(ir.ToString(true, "\n"))
	return ir
}

func (ir *BasicIrr) LogError(logger interface{ Error(args ...interface{}) }) IRR {
	logger.Error(ir.ToString(true, "\n"))
	return ir
}

func (ir *BasicIrr) LogFatal(logger interface{ Fatal(args ...interface{}) }) IRR {
	str := ir.ToString(true, "\n")
	logger.Fatal(str)
	// to make sure it has been printed to std output stream
	fmt.Println(str)
	return ir
}

func newIrr(formatOrMsg string, args ...interface{}) *BasicIrr {
	err := &BasicIrr{}
	if len(args) > 0 {
		err.Msg = fmt.Sprintf(formatOrMsg, args...)
	} else {
		err.Msg = formatOrMsg
	}
	return err
}
