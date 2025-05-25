package irc

import (
	"errors"
	"fmt"
	"testing"

	"github.com/khicago/irr"
	"github.com/stretchr/testify/assert"
)

// 定义测试用的错误码常量
const (
	TestCodeSuccess     Code = 0
	TestCodeNotFound    Code = 404
	TestCodeServerError Code = 500
	TestCodeBadRequest  Code = 400
	TestCodeUnknown     Code = 9999
)

func TestCode_I64(t *testing.T) {
	tests := []struct {
		name     string
		code     Code
		expected int64
	}{
		{"zero code", TestCodeSuccess, 0},
		{"positive code", TestCodeNotFound, 404},
		{"large code", TestCodeUnknown, 9999},
		{"negative code", Code(-1), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.code.I64()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCode_String(t *testing.T) {
	tests := []struct {
		name     string
		code     Code
		expected string
	}{
		{"zero code", TestCodeSuccess, "0"},
		{"positive code", TestCodeNotFound, "404"},
		{"large code", TestCodeUnknown, "9999"},
		{"negative code", Code(-1), "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.code.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCode_Error(t *testing.T) {
	tests := []struct {
		name        string
		code        Code
		formatOrMsg string
		args        []interface{}
		expectCode  int64
		expectMsg   string
	}{
		{
			name:        "simple message",
			code:        TestCodeNotFound,
			formatOrMsg: "user not found",
			args:        nil,
			expectCode:  404,
			expectMsg:   "user not found",
		},
		{
			name:        "formatted message",
			code:        TestCodeBadRequest,
			formatOrMsg: "invalid user ID: %d",
			args:        []interface{}{12345},
			expectCode:  400,
			expectMsg:   "invalid user ID: 12345",
		},
		{
			name:        "zero code",
			code:        TestCodeSuccess,
			formatOrMsg: "operation completed",
			args:        nil,
			expectCode:  0,
			expectMsg:   "operation completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.code.Error(tt.formatOrMsg, tt.args...)

			// 验证错误码
			assert.Equal(t, tt.expectCode, err.GetCode())

			// 验证错误消息包含预期内容
			errStr := err.Error()
			assert.Contains(t, errStr, tt.expectMsg)

			// 验证错误码在消息中的格式
			if tt.expectCode != 0 {
				expectedCodeStr := fmt.Sprintf("code(%d)", tt.expectCode)
				assert.Contains(t, errStr, expectedCodeStr)
			}
		})
	}
}

func TestCode_Wrap(t *testing.T) {
	innerErr := errors.New("connection timeout")

	tests := []struct {
		name        string
		code        Code
		innerErr    error
		formatOrMsg string
		args        []interface{}
		expectCode  int64
	}{
		{
			name:        "wrap with simple message",
			code:        TestCodeServerError,
			innerErr:    innerErr,
			formatOrMsg: "database operation failed",
			args:        nil,
			expectCode:  500,
		},
		{
			name:        "wrap with formatted message",
			code:        TestCodeNotFound,
			innerErr:    innerErr,
			formatOrMsg: "user %s not found",
			args:        []interface{}{"john_doe"},
			expectCode:  404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.code.Wrap(tt.innerErr, tt.formatOrMsg, tt.args...)

			// 验证错误码
			assert.Equal(t, tt.expectCode, err.GetCode())

			// 验证可以解包到内部错误
			assert.Equal(t, tt.innerErr, errors.Unwrap(err))

			// 验证错误消息包含包装信息
			errStr := err.Error()
			assert.Contains(t, errStr, tt.innerErr.Error())
		})
	}
}

func TestCode_Trace(t *testing.T) {
	code := TestCodeNotFound
	err := code.Trace("user not found")

	// 验证错误不为nil
	assert.NotNil(t, err)

	// 验证错误消息
	assert.Contains(t, err.Error(), "user not found")

	// 验证错误码
	assert.Equal(t, int64(404), err.GetCode())

	// 验证有堆栈信息 - 使用ToString(true, split)来获取堆栈跟踪
	traceStr := err.ToString(true, "\n")
	assert.Contains(t, traceStr, "irc.Code.Trace") // 应该包含Code.Trace方法
	assert.Contains(t, traceStr, "/irc/code.go:")  // 应该包含code.go文件

	// 验证有堆栈跟踪信息
	assert.NotNil(t, err.GetTraceInfo())
}

func TestCode_TraceSkip(t *testing.T) {
	code := TestCodeServerError
	err := code.TraceSkip(1, "server error occurred")

	// 验证错误不为nil
	assert.NotNil(t, err)

	// 验证错误消息
	assert.Contains(t, err.Error(), "server error occurred")

	// 验证错误码
	assert.Equal(t, int64(500), err.GetCode())

	// 验证有堆栈信息，但跳过了一层
	traceStr := err.ToString(true, "\n")
	assert.NotEmpty(t, traceStr)
	// 由于跳过了一层，可能不包含当前函数名，但应该有堆栈信息
	assert.Contains(t, traceStr, ".go:") // 应该包含某个文件的行号信息
}

func TestCode_Track(t *testing.T) {
	code := TestCodeBadRequest
	innerErr := errors.New("validation failed")
	err := code.Track(innerErr, "request validation error")

	// 验证错误不为nil
	assert.NotNil(t, err)

	// 验证错误消息包含原始错误和新消息
	errStr := err.Error()
	assert.Contains(t, errStr, "request validation error")
	assert.Contains(t, errStr, "validation failed")

	// 验证错误码
	assert.Equal(t, int64(400), err.GetCode())

	// 验证有堆栈信息 - 使用ToString(true, split)来获取堆栈跟踪
	traceStr := err.ToString(true, "\n")
	assert.Contains(t, traceStr, "irc.Code.Track") // 应该包含Code.Track方法
	assert.Contains(t, traceStr, "/irc/code.go:")  // 应该包含code.go文件

	// 验证有堆栈跟踪信息
	assert.NotNil(t, err.GetTraceInfo())
}

func TestCode_TrackSkip(t *testing.T) {
	innerErr := errors.New("inner error")

	// 创建一个辅助函数来测试skip功能
	createTrackErrorWithSkip := func(skip int) irr.IRR {
		return TestCodeBadRequest.TrackSkip(skip, innerErr, "tracked error with skip %d", skip)
	}

	err0 := createTrackErrorWithSkip(0)
	err1 := createTrackErrorWithSkip(1)

	// 验证错误码
	assert.Equal(t, int64(400), err0.GetCode())
	assert.Equal(t, int64(400), err1.GetCode())

	// 验证都可以解包到内部错误
	assert.Equal(t, innerErr, errors.Unwrap(err0))
	assert.Equal(t, innerErr, errors.Unwrap(err1))

	// 验证都有堆栈跟踪
	assert.NotNil(t, err0.GetTraceInfo())
	assert.NotNil(t, err1.GetTraceInfo())
}

func TestCode_ErrorChaining(t *testing.T) {
	// 测试错误链中的错误码传播
	originalErr := TestCodeNotFound.Error("user not found")
	wrappedErr := TestCodeServerError.Wrap(originalErr, "service unavailable")
	trackedErr := TestCodeBadRequest.Track(wrappedErr, "request processing failed")

	// 验证最外层错误码
	assert.Equal(t, int64(400), trackedErr.GetCode())

	// 验证可以获取最近的错误码（应该是最外层的）
	assert.Equal(t, int64(400), trackedErr.ClosestCode())

	// 验证错误链
	assert.Equal(t, wrappedErr, errors.Unwrap(trackedErr))
	assert.Equal(t, originalErr, errors.Unwrap(wrappedErr))
}

func TestCode_WithZeroCode(t *testing.T) {
	// 测试零错误码的特殊情况
	err := TestCodeSuccess.Error("operation completed successfully")

	// 零错误码应该正常工作
	assert.Equal(t, int64(0), err.GetCode())

	// 错误消息中不应该包含code(0)
	errStr := err.Error()
	assert.NotContains(t, errStr, "code(0)")
	assert.Contains(t, errStr, "operation completed successfully")
}

func TestCode_InterfaceCompliance(t *testing.T) {
	// 验证Code类型实现了预期的接口
	var _ irr.Spawner = Code(0)

	// 测试通过接口使用
	var spawner irr.Spawner = TestCodeNotFound
	err := spawner.Error("interface test")

	// 验证通过接口创建的错误仍然有正确的错误码
	if coder, ok := err.(irr.ICoder[int64]); ok {
		assert.Equal(t, int64(404), coder.GetCode())
	} else {
		t.Error("Error should implement ICoder interface")
	}
}
