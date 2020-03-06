package irr

import (
	"errors"
	"fmt"
	"strings"
)

func (ir *irr) Root() error {
	var err error = ir
	for {
		inner := errors.Unwrap(err)
		if inner == nil {
			return err
		}
		err = inner
	}
}

func (ir *irr) Source() error {
	for cur := ir; cur != nil; {
		if cur.inner == nil {
			return cur
		}
		if next, ok := cur.inner.(*irr); ok {
			cur = next
		} else {
			return cur
		}
	}
	return ir
}

func (ir *irr) Unwrap() error {
	return ir.inner
}

func (ir *irr) GetTraceInfo() *traceInfo {
	return ir.Trace
}

func (ir *irr) TraverseToSource(fn func(err error) error) (err error) {
	defer CatchFailure(func(e error) { err = e })
	for cur := ir; cur != nil; {
		err = fn(cur)
		if cur.inner == nil {
			break
		}
		if next, ok := cur.inner.(*irr); ok {
			cur = next
		} else {
			err = fn(cur)
			break
		}
	}
	return
}

func (ir *irr) TraverseToRoot(fn func(err error) error) (err error) {
	defer CatchFailure(func(e error) { err = e })
	var inner error = ir
	for inner != nil {
		err = fn(inner)
		inner = errors.Unwrap(inner)
	}
	return
}

func (ir *irr) writeSelfTo(sb *strings.Builder, printTrace bool) {
	sb.WriteString(ir.Msg)
	if printTrace && ir.Trace != nil {
		sb.WriteRune(' ')
		ir.Trace.writeTo(sb)
	}
}

func (ir *irr) ToString(printTrace bool, split string) string {
	sb := strings.Builder{}
	_ = ir.TraverseToSource(func(err error) error {
		if irr, ok := err.(*irr); ok {
			irr.writeSelfTo(&sb, printTrace)
		} else {
			sb.WriteString(err.Error())
		}
		sb.WriteString(split)
		return nil
	})
	return sb.String()
}

func (ir *irr) Error() string {
	return ir.ToString(false, "; ")
}

func (ir *irr) LogWarn(logger interface{ Warn(args ...interface{}) }) IRR {
	logger.Warn(ir.ToString(true, "\n"))
	return ir
}

func (ir *irr) LogError(logger interface{ Error(args ...interface{}) }) IRR {
	logger.Error(ir.ToString(true, "\n"))
	return ir
}

func (ir *irr) LogFatal(logger interface{ Fatal(args ...interface{}) }) IRR {
	str := ir.ToString(true, "\n")
	logger.Fatal(str)
	fmt.Println(str)
	return ir
}

func newIrr(formatOrMsg string, args ...interface{}) *irr {
	err := &irr{}
	if len(args) > 0 {
		err.Msg = fmt.Sprintf(formatOrMsg, args...)
	} else {
		err.Msg = formatOrMsg
	}
	return err
}