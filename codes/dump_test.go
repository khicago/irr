package codes

import (
	"errors"
	"testing"

	"github.com/khicago/irr"
	"github.com/stretchr/testify/assert"
)

// 注意: 清理了使用废弃GetCode()接口的mock对象
// 现在统一使用irr.IRR接口

// 使用code_test.go中已定义的测试常量
// 添加缺失的TestCodeInternalError常量

func TestDumpToCodeNError(t *testing.T) {
	tests := []struct {
		name       string
		succ       Code
		unknown    Code
		err        error
		msgOrFmt   string
		args       []interface{}
		expectCode Code
		expectMsg  string
	}{
		{
			name:       "nil error - no args",
			succ:       TestCodeSuccess,
			unknown:    TestCodeUnknown,
			err:        nil,
			msgOrFmt:   "",
			args:       nil,
			expectCode: TestCodeSuccess,
			expectMsg:  "",
		},
		{
			name:       "simple error without code - no args",
			succ:       TestCodeSuccess,
			unknown:    TestCodeUnknown,
			err:        errors.New("simple error"),
			msgOrFmt:   "",
			args:       nil,
			expectCode: TestCodeUnknown,
			expectMsg:  "simple error",
		},
		{
			name:       "error with formatted message",
			succ:       TestCodeSuccess,
			unknown:    TestCodeUnknown,
			err:        errors.New("connection failed"),
			msgOrFmt:   "failed to connect to %s",
			args:       []interface{}{"database"},
			expectCode: TestCodeUnknown,
			expectMsg:  "failed to connect to database, connection failed",
		},
		{
			name:       "empty message format",
			succ:       TestCodeSuccess,
			unknown:    TestCodeUnknown,
			err:        errors.New("test error"),
			msgOrFmt:   "",
			args:       nil,
			expectCode: TestCodeUnknown,
			expectMsg:  "test error",
		},
		{
			name:       "simple error with args",
			succ:       TestCodeSuccess,
			unknown:    TestCodeUnknown,
			err:        errors.New("operation failed"),
			msgOrFmt:   "context: %s",
			args:       []interface{}{"user action"},
			expectCode: TestCodeUnknown,
			expectMsg:  "context: user action, operation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, msg := DumpToCodeNError(tt.succ, tt.unknown, tt.err, tt.msgOrFmt, tt.args...)
			assert.Equal(t, tt.expectCode, code)
			assert.Equal(t, tt.expectMsg, msg)
		})
	}
}

func TestDumpToCodeNError_WithIRRError(t *testing.T) {
	irrErr := irr.ErrorC(int64(TestCodeNotFound), "user not found")

	code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, irrErr, "failed to retrieve user")

	assert.Equal(t, TestCodeNotFound, code)
	assert.Contains(t, msg, "failed to retrieve user")
	assert.Contains(t, msg, "user not found")
}

func TestDumpToCodeNError_WithWrappedIRRError(t *testing.T) {
	baseErr := irr.ErrorC(int64(TestCodeBadRequest), "invalid parameters")
	wrappedErr := irr.Wrap(baseErr, "validation failed")

	code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, wrappedErr, "request processing")

	assert.Equal(t, TestCodeBadRequest, code)
	assert.Contains(t, msg, "request processing")
	assert.Contains(t, msg, "validation failed")
	assert.Contains(t, msg, "invalid parameters")
}

func TestDumpToCodeNError_WithNestedIRRErrors(t *testing.T) {
	innerErr := irr.ErrorC(int64(TestCodeInternalError), "database connection lost")
	middleErr := irr.Wrap(innerErr, "query execution failed")
	outerErr := irr.Wrap(middleErr, "user operation failed")

	code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, outerErr, "")

	assert.Equal(t, TestCodeInternalError, code)
	assert.Contains(t, msg, "user operation failed")
	assert.Contains(t, msg, "database connection lost")
}
