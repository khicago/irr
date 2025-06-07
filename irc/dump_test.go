package irc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 自定义错误类型，用于测试ICodeGetter接口
type customError struct {
	code int64
	msg  string
}

func (e *customError) Error() string {
	return e.msg
}

func (e *customError) GetCode() int64 {
	return e.code
}

func (e *customError) GetCodeStr() string {
	if e.code == 0 {
		return ""
	}
	return "custom_code(" + string(rune(e.code)) + "), "
}

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
			name:       "nil error",
			succ:       TestCodeSuccess,
			unknown:    TestCodeUnknown,
			err:        nil,
			msgOrFmt:   "operation completed",
			args:       nil,
			expectCode: TestCodeSuccess,
			expectMsg:  "",
		},
		{
			name:       "simple error without code - no args",
			succ:       TestCodeSuccess,
			unknown:    TestCodeUnknown,
			err:        errors.New("simple error"),
			msgOrFmt:   "operation failed",
			args:       nil, // 没有args，所以不会添加前缀
			expectCode: TestCodeUnknown,
			expectMsg:  "simple error",
		},
		{
			name:       "error with formatted message",
			succ:       TestCodeSuccess,
			unknown:    TestCodeUnknown,
			err:        errors.New("connection timeout"),
			msgOrFmt:   "database operation failed for user %s",
			args:       []interface{}{"john_doe"},
			expectCode: TestCodeUnknown,
			expectMsg:  "database operation failed for user john_doe, connection timeout",
		},
		{
			name:       "empty message format",
			succ:       TestCodeSuccess,
			unknown:    TestCodeUnknown,
			err:        errors.New("some error"),
			msgOrFmt:   "",
			args:       nil,
			expectCode: TestCodeUnknown,
			expectMsg:  "some error",
		},
		{
			name:       "simple error with args",
			succ:       TestCodeSuccess,
			unknown:    TestCodeUnknown,
			err:        errors.New("simple error"),
			msgOrFmt:   "operation failed",
			args:       []interface{}{}, // 空args但不是nil
			expectCode: TestCodeUnknown,
			expectMsg:  "simple error", // 仍然不会添加前缀，因为len(args) == 0
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
	// 测试带有错误码的IRR错误
	originalErr := TestCodeNotFound.Error("user not found")

	// 需要提供args才能添加前缀消息
	code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, originalErr, "service error for %s", "user123")

	// 应该提取到原始错误码
	assert.Equal(t, TestCodeNotFound, code)

	// 消息应该包含服务错误信息，但不包含重复的错误码前缀
	assert.Contains(t, msg, "service error for user123")
	assert.Contains(t, msg, "user not found")
	// 不应该包含重复的code(404)前缀
	assert.NotContains(t, msg, "code(404), service error for user123, code(404)")
}

func TestDumpToCodeNError_WithWrappedIRRError(t *testing.T) {
	// 测试包装的IRR错误
	innerErr := errors.New("database connection failed")
	wrappedErr := TestCodeServerError.Wrap(innerErr, "service unavailable")

	code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, wrappedErr, "request failed for %s", "endpoint")

	// 应该提取到最近的错误码
	assert.Equal(t, TestCodeServerError, code)

	// 消息应该正确格式化
	assert.Contains(t, msg, "request failed for endpoint")
	assert.Contains(t, msg, "service unavailable")
}

func TestDumpToCodeNError_WithNestedIRRErrors(t *testing.T) {
	// 测试嵌套的IRR错误，验证ClosestCode的行为
	originalErr := TestCodeNotFound.Error("user not found")
	wrappedErr := TestCodeServerError.Wrap(originalErr, "service error")
	trackedErr := TestCodeBadRequest.Track(wrappedErr, "request processing failed")

	code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, trackedErr, "API error in %s", "handler")

	// 应该获取最外层（最近的）错误码
	assert.Equal(t, TestCodeBadRequest, code)

	// 消息应该包含API错误信息
	assert.Contains(t, msg, "API error in handler")
	assert.Contains(t, msg, "request processing failed")
}

func TestDumpToCodeNError_CodeStripping(t *testing.T) {
	// 测试错误码字符串的剥离功能
	originalErr := TestCodeNotFound.Error("user not found")

	// 获取原始错误消息，应该包含code(404)前缀
	originalMsg := originalErr.Error()
	assert.Contains(t, originalMsg, "code(404)")

	code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, originalErr, "")

	// 提取的消息应该去掉code(404)前缀
	assert.Equal(t, TestCodeNotFound, code)
	assert.NotContains(t, msg, "code(404)")
	assert.Contains(t, msg, "user not found")
}

func TestDumpToCodeNError_WithICodeGetter(t *testing.T) {
	customErr := &customError{code: 999, msg: "custom error message"}

	code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, customErr, "wrapper message for %s", "test")

	// 应该提取到自定义错误码
	assert.Equal(t, Code(999), code)

	// 消息应该包含包装信息和原始消息
	assert.Contains(t, msg, "wrapper message for test")
	assert.Contains(t, msg, "custom error message")
}

func TestDumpToCodeNError_EdgeCases(t *testing.T) {
	t.Run("空消息格式", func(t *testing.T) {
		err := errors.New("test error")
		code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, err, "")
		
		assert.Equal(t, TestCodeUnknown, code)
		assert.Equal(t, "test error", msg)
	})

	t.Run("消息格式但无参数", func(t *testing.T) {
		err := errors.New("test error")
		code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, err, "prefix message")
		
		assert.Equal(t, TestCodeUnknown, code)
		assert.Equal(t, "test error", msg)
	})

	t.Run("消息格式和参数", func(t *testing.T) {
		err := errors.New("test error")
		code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, err, "prefix %s", "formatted")
		
		assert.Equal(t, TestCodeUnknown, code)
		assert.Equal(t, "prefix formatted, test error", msg)
	})

	t.Run("错误消息包含代码字符串前缀", func(t *testing.T) {
		// 创建一个带有代码字符串前缀的错误
		baseErr := &mockCodeError{
			code:    404,
			codeStr: "code(404), ",
			message: "code(404), not found",
		}
		
		code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, baseErr, "")
		
		assert.Equal(t, Code(404), code)
		assert.Equal(t, "not found", msg)
	})

	t.Run("错误消息不包含代码字符串前缀", func(t *testing.T) {
		baseErr := &mockCodeError{
			code:    404,
			codeStr: "code(404), ",
			message: "not found",
		}
		
		code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, baseErr, "")
		
		assert.Equal(t, Code(404), code)
		assert.Equal(t, "not found", msg)
	})

	t.Run("使用NearestCode接口", func(t *testing.T) {
		baseErr := &mockNearestCodeError{
			code:    500,
			codeStr: "code(500), ",
			message: "code(500), server error",
		}
		
		code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, baseErr, "")
		
		assert.Equal(t, Code(500), code)
		assert.Equal(t, "server error", msg)
	})

	t.Run("使用ICodeTraverse接口", func(t *testing.T) {
		baseErr := &mockTraverseError{
			code:    403,
			codeStr: "code(403), ",
			message: "code(403), forbidden",
		}
		
		code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, baseErr, "")
		
		assert.Equal(t, Code(403), code)
		assert.Equal(t, "forbidden", msg)
	})

	t.Run("使用ICodeGetter接口", func(t *testing.T) {
		baseErr := &mockGetterError{
			code:    401,
			codeStr: "code(401), ",
			message: "code(401), unauthorized",
		}
		
		code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, baseErr, "")
		
		assert.Equal(t, Code(401), code)
		assert.Equal(t, "unauthorized", msg)
	})

	t.Run("复杂格式化消息", func(t *testing.T) {
		err := errors.New("database connection failed")
		code, msg := DumpToCodeNError(TestCodeSuccess, TestCodeUnknown, err, "operation %s failed with %d retries", "connect", 3)
		
		assert.Equal(t, TestCodeUnknown, code)
		assert.Equal(t, "operation connect failed with 3 retries, database connection failed", msg)
	})
}

// 辅助测试结构体
type mockCodeError struct {
	code    int64
	codeStr string
	message string
}

func (e *mockCodeError) Error() string {
	return e.message
}

func (e *mockCodeError) GetCode() int64 {
	return e.code
}

func (e *mockCodeError) GetCodeStr() string {
	return e.codeStr
}

type mockNearestCodeError struct {
	code    int64
	codeStr string
	message string
}

func (e *mockNearestCodeError) Error() string {
	return e.message
}

func (e *mockNearestCodeError) NearestCode() int64 {
	return e.code
}

func (e *mockNearestCodeError) GetCodeStr() string {
	return e.codeStr
}

type mockTraverseError struct {
	code    int64
	codeStr string
	message string
}

func (e *mockTraverseError) Error() string {
	return e.message
}

func (e *mockTraverseError) ClosestCode() int64 {
	return e.code
}

func (e *mockTraverseError) GetCode() int64 {
	return e.code
}

func (e *mockTraverseError) GetCodeStr() string {
	return e.codeStr
}

func (e *mockTraverseError) TraverseCode(fn func(err error, code int64) error) error {
	return fn(e, e.code)
}

type mockGetterError struct {
	code    int64
	codeStr string
	message string
}

func (e *mockGetterError) Error() string {
	return e.message
}

func (e *mockGetterError) GetCode() int64 {
	return e.code
}

func (e *mockGetterError) GetCodeStr() string {
	return e.codeStr
}
