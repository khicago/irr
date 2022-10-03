package irc

import (
	"strconv"

	"github.com/khicago/irr"
)

type (
	Code int64
)

var _ irr.Spawner = Code(0)

func (c Code) I64() int64 {
	return int64(c)
}

func (c Code) String() string {
	return strconv.FormatInt(c.I64(), 10)
}

func (c Code) Error(formatOrMsg string, args ...interface{}) irr.IRR {
	return irr.Error(formatOrMsg, args...).SetCode(c.I64())
}

func (c Code) Wrap(innerErr error, formatOrMsg string, args ...interface{}) irr.IRR {
	return irr.Wrap(innerErr, formatOrMsg, args...).SetCode(c.I64())
}

func (c Code) TraceSkip(skip int, formatOrMsg string, args ...interface{}) irr.IRR {
	return irr.TraceSkip(skip, formatOrMsg, args...).SetCode(c.I64())
}

func (c Code) Trace(formatOrMsg string, args ...interface{}) irr.IRR {
	return irr.Trace(formatOrMsg, args...).SetCode(c.I64())
}

func (c Code) TrackSkip(skip int, innerErr error, formatOrMsg string, args ...interface{}) irr.IRR {
	return irr.TrackSkip(skip, innerErr, formatOrMsg, args...).SetCode(c.I64())
}

func (c Code) Track(innerErr error, formatOrMsg string, args ...interface{}) irr.IRR {
	return irr.Track(innerErr, formatOrMsg, args...).SetCode(c.I64())
}
