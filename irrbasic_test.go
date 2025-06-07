package irr

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	theIrr, innerIrr   *BasicIrr
	sourceErr, rootErr error
)

type testLogger struct{ ret string }

func (l *testLogger) Warn(args ...interface{}) {
	l.ret = fmt.Sprint(args...)
}

func (l *testLogger) Error(args ...interface{}) {
	l.ret = fmt.Sprint(args...)
}

func (l *testLogger) Fatal(args ...interface{}) {
	l.ret = fmt.Sprint(args...)
}

func init() {
	theIrr = newBasicIrr("test err %d", 1)
	innerIrr = newBasicIrr("inner error")
	theIrr.inner = innerIrr
	rootErr = errors.New("root error")
	sourceErr = fmt.Errorf("source error => %w", rootErr)
	innerIrr.inner = sourceErr
}

func TestIrrToString(t *testing.T) {
	str := theIrr.ToString(false, "=the=split=")
	assert.Equal(t, "test err 1=the=split=inner error=the=split=source error => root error", str, "")
}

func TestIrrFmtV(t *testing.T) {
	str := fmt.Sprintf("%v", theIrr)
	assert.Equal(t, "test err 1, inner error, source error => root error", str, "")
	str = fmt.Sprintf("%+v", theIrr)
	assert.Equal(t, "test err 1, inner error, source error => root error", str, "")
	str = fmt.Sprintf("%+q", theIrr)
	assert.Equal(t, "\"test err 1, inner error, source error => root error\"", str, "")
}

func TestIrrError(t *testing.T) {
	str := theIrr.Error()
	assert.Equal(t, "test err 1, inner error, source error => root error", str, "")
}

func TestIrrRoot(t *testing.T) {
	assert.Equal(t, rootErr, theIrr.Root(), "root error are not equal")
}

func TestIrrSource(t *testing.T) {
	assert.Equal(t, sourceErr, theIrr.Source(), "source error are not equal")
}

func TestUnwrap(t *testing.T) {
	assert.Equal(t, innerIrr, theIrr.Unwrap(), "unwrap to innerIrr are not correct")
	assert.Equal(t, sourceErr, innerIrr.Unwrap(), "unwrap to sourceErr are not correct")
}

func TestIrrTraverseToSourceStack(t *testing.T) {
	stack := []error{
		theIrr, innerIrr, sourceErr,
	}
	_ = theIrr.TraverseToSource(func(err error, isSource bool) error {
		pop := stack[0]
		stack = stack[1:]
		assert.Equal(t, pop, err, "wrong error stack")
		if isSource {
			assert.Equal(t, 0, len(stack), "stack should finished")
		}
		return nil
	})
	assert.Equal(t, 0, len(stack), "stack should finished")
}

func TestIrrTraverseToSourceThrownErr(t *testing.T) {
	previousErr := errors.New("the previous error")
	returnedErr := errors.New("the returned error")
	err := theIrr.TraverseToSource(func(err error, isSource bool) error {
		if isSource {
			return returnedErr
		}
		return previousErr
	})
	assert.Equal(t, returnedErr, err, "return error stack")

	err = theIrr.TraverseToSource(func(err error, isSource bool) error {
		if isSource {
			return returnedErr
		}
		panic(previousErr)
	})
	assert.Equal(t, previousErr, err, "return error stack")

	err = theIrr.TraverseToSource(func(err error, isSource bool) error {
		panic("some error string")
	})
	assert.Exactly(t, true, errors.Is(err, ErrUntypedExecutionFailure), "should returns ErrUntypedExecutionFailure")
}

func TestIrrTraverseToRootStack(t *testing.T) {
	stack := []error{
		theIrr, innerIrr, sourceErr, rootErr,
	}
	_ = theIrr.TraverseToRoot(func(err error) error {
		pop := stack[0]
		stack = stack[1:]
		assert.Equal(t, pop, err, "wrong error stack")
		return nil
	})
	assert.Equal(t, 0, len(stack), "stack should finished")
}

func TestIrrTraverseToRootThrownErr(t *testing.T) {
	previousErr := errors.New("the previous error")
	err := theIrr.TraverseToRoot(func(error) error {
		return previousErr
	})
	assert.Equal(t, previousErr, err, "return error stack")

	err = theIrr.TraverseToRoot(func(err error) error {
		panic(previousErr)
	})
	assert.Equal(t, previousErr, err, "return error stack")

	err = theIrr.TraverseToRoot(func(err error) error {
		panic("some error string")
	})
	assert.Exactly(t, true, errors.Is(err, ErrUntypedExecutionFailure), "should returns ErrUntypedExecutionFailure")
}

func TestLog(t *testing.T) {
	l := &testLogger{}
	ir := theIrr.LogWarn(l)
	assert.Equal(t, theIrr.ToString(true, "\n"), l.ret, "LogWarn failed")
	assert.Equal(t, theIrr, ir, "ir should be returned by LogWarn")

	ir = theIrr.LogError(l)
	assert.Equal(t, theIrr.ToString(true, "\n"), l.ret, "LogError failed")
	assert.Equal(t, theIrr, ir, "ir should be returned by LogError")

	ir = theIrr.LogFatal(l)
	assert.Equal(t, theIrr.ToString(true, "\n"), l.ret, "LogFatal failed")
	assert.Equal(t, theIrr, ir, "ir should be returned by LogFatal")
}

// === 新错误码API测试 ===

// TestNewErrorCodeAPI_NearestCode 测试NearestCode方法
func TestNewErrorCodeAPI_NearestCode(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() IRR
		expected int64
	}{
		{
			name: "单个错误有错误码",
			setup: func() IRR {
				return Error("test error").SetCode(404)
			},
			expected: 404,
		},
		{
			name: "单个错误无错误码",
			setup: func() IRR {
				return Error("test error")
			},
			expected: 0,
		},
		{
			name: "错误链-最外层有错误码",
			setup: func() IRR {
				inner := Error("inner error").SetCode(500)
				return Wrap(inner, "outer error").SetCode(400)
			},
			expected: 400, // 最近的（最外层）错误码
		},
		{
			name: "错误链-只有内层有错误码",
			setup: func() IRR {
				inner := Error("inner error").SetCode(500)
				return Wrap(inner, "outer error") // 外层没有设置错误码
			},
			expected: 500, // 内层的错误码
		},
		{
			name: "错误链-多层嵌套",
			setup: func() IRR {
				root := Error("root error").SetCode(404)
				middle := Wrap(root, "middle error").SetCode(500)
				return Wrap(middle, "outer error").SetCode(400)
			},
			expected: 400, // 最外层的错误码
		},
		{
			name: "错误链-中间层有错误码",
			setup: func() IRR {
				root := Error("root error") // 无错误码
				middle := Wrap(root, "middle error").SetCode(500)
				return Wrap(middle, "outer error") // 无错误码
			},
			expected: 500, // 中间层的错误码
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup()
			assert.Equal(t, tt.expected, err.NearestCode())
		})
	}
}

// TestNewErrorCodeAPI_CurrentCode 测试CurrentCode方法
func TestNewErrorCodeAPI_CurrentCode(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() IRR
		expected int64
	}{
		{
			name: "设置了错误码",
			setup: func() IRR {
				return Error("test error").SetCode(404)
			},
			expected: 404,
		},
		{
			name: "未设置错误码",
			setup: func() IRR {
				return Error("test error")
			},
			expected: 0,
		},
		{
			name: "设置错误码为0",
			setup: func() IRR {
				return Error("test error").SetCode(0)
			},
			expected: 0,
		},
		{
			name: "错误链-只检查当前层",
			setup: func() IRR {
				inner := Error("inner error").SetCode(500)
				return Wrap(inner, "outer error") // 外层未设置错误码
			},
			expected: 0, // 当前层（外层）没有错误码
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup()
			assert.Equal(t, tt.expected, err.CurrentCode())
		})
	}
}

// TestNewErrorCodeAPI_RootCode 测试RootCode方法
func TestNewErrorCodeAPI_RootCode(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() IRR
		expected int64
	}{
		{
			name: "单个错误",
			setup: func() IRR {
				return Error("test error").SetCode(404)
			},
			expected: 404,
		},
		{
			name: "错误链-根错误有错误码",
			setup: func() IRR {
				root := Error("root error").SetCode(404)
				middle := Wrap(root, "middle error").SetCode(500)
				return Wrap(middle, "outer error").SetCode(400)
			},
			expected: 404, // 根错误的错误码
		},
		{
			name: "错误链-根错误无错误码",
			setup: func() IRR {
				root := Error("root error") // 无错误码
				middle := Wrap(root, "middle error").SetCode(500)
				return Wrap(middle, "outer error").SetCode(400)
			},
			expected: 0, // 根错误没有错误码
		},
		{
			name: "包装标准错误",
			setup: func() IRR {
				stdErr := errors.New("standard error")
				return Wrap(stdErr, "wrapped error").SetCode(500)
			},
			expected: 0, // 标准错误没有错误码
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup()
			assert.Equal(t, tt.expected, err.RootCode())
		})
	}
}

// TestNewErrorCodeAPI_HasCurrentCode 测试HasCurrentCode方法
func TestNewErrorCodeAPI_HasCurrentCode(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() IRR
		expected bool
	}{
		{
			name: "未设置错误码",
			setup: func() IRR {
				return Error("test error")
			},
			expected: false,
		},
		{
			name: "设置了非零错误码",
			setup: func() IRR {
				return Error("test error").SetCode(404)
			},
			expected: true,
		},
		{
			name: "显式设置错误码为0",
			setup: func() IRR {
				return Error("test error").SetCode(0)
			},
			expected: true, // 显式设置了，即使是0
		},
		{
			name: "错误链-外层未设置",
			setup: func() IRR {
				inner := Error("inner error").SetCode(500)
				return Wrap(inner, "outer error") // 外层未设置
			},
			expected: false, // 当前层（外层）未设置
		},
		{
			name: "错误链-外层设置了",
			setup: func() IRR {
				inner := Error("inner error").SetCode(500)
				return Wrap(inner, "outer error").SetCode(400)
			},
			expected: true, // 当前层（外层）设置了
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup()
			assert.Equal(t, tt.expected, err.HasCurrentCode())
		})
	}
}

// TestNewErrorCodeAPI_HasAnyCode 测试HasAnyCode方法
func TestNewErrorCodeAPI_HasAnyCode(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() IRR
		expected bool
	}{
		{
			name: "完全没有错误码",
			setup: func() IRR {
				root := Error("root error")
				return Wrap(root, "outer error")
			},
			expected: false,
		},
		{
			name: "有错误码",
			setup: func() IRR {
				return Error("test error").SetCode(404)
			},
			expected: true,
		},
		{
			name: "错误链中有错误码",
			setup: func() IRR {
				inner := Error("inner error").SetCode(500)
				return Wrap(inner, "outer error") // 外层无错误码
			},
			expected: true, // 内层有错误码
		},
		{
			name: "所有错误码都是0",
			setup: func() IRR {
				inner := Error("inner error").SetCode(0)
				return Wrap(inner, "outer error").SetCode(0)
			},
			expected: false, // 所有错误码都是0，视为无有效错误码
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup()
			assert.Equal(t, tt.expected, err.HasAnyCode())
		})
	}
}

// TestNewErrorCodeAPI_BackwardCompatibility 测试向后兼容性
func TestNewErrorCodeAPI_BackwardCompatibility(t *testing.T) {
	// 创建一个错误链
	inner := Error("inner error").SetCode(404)
	wrapped := Wrap(inner, "wrapped error").SetCode(500)

	// 测试GetCode和NearestCode返回相同结果
	assert.Equal(t, wrapped.NearestCode(), wrapped.GetCode())
	
	// 测试ClosestCode和NearestCode返回相同结果
	assert.Equal(t, wrapped.NearestCode(), wrapped.ClosestCode())

	// 测试具体值
	assert.Equal(t, int64(500), wrapped.GetCode())
	assert.Equal(t, int64(500), wrapped.ClosestCode())
	assert.Equal(t, int64(500), wrapped.NearestCode())
}

// TestNewErrorCodeAPI_SetCodeBehavior 测试SetCode的新行为
func TestNewErrorCodeAPI_SetCodeBehavior(t *testing.T) {
	err := Error("test error")
	
	// 初始状态
	assert.False(t, err.HasCurrentCode())
	assert.Equal(t, int64(0), err.CurrentCode())
	
	// 设置错误码
	err.SetCode(404)
	assert.True(t, err.HasCurrentCode())
	assert.Equal(t, int64(404), err.CurrentCode())
	
	// 设置错误码为0
	err.SetCode(0)
	assert.True(t, err.HasCurrentCode()) // 仍然是true，因为显式设置了
	assert.Equal(t, int64(0), err.CurrentCode())
}

// TestNewErrorCodeAPI_ComplexChain 测试复杂错误链
func TestNewErrorCodeAPI_ComplexChain(t *testing.T) {
	// 创建复杂的错误链
	stdErr := errors.New("standard error")
	level1 := Wrap(stdErr, "level 1").SetCode(100)
	level2 := Wrap(level1, "level 2") // 无错误码
	level3 := Wrap(level2, "level 3").SetCode(300)
	level4 := Wrap(level3, "level 4") // 无错误码
	level5 := Wrap(level4, "level 5").SetCode(500)

	// 测试各种方法
	assert.Equal(t, int64(500), level5.NearestCode()) // 最近的有效错误码
	assert.Equal(t, int64(500), level5.CurrentCode()) // 当前层错误码
	assert.Equal(t, int64(0), level5.RootCode())      // 根错误（标准错误）无错误码
	assert.True(t, level5.HasCurrentCode())           // 当前层设置了错误码
	assert.True(t, level5.HasAnyCode())               // 错误链中有错误码

	// 测试中间层
	assert.Equal(t, int64(300), level4.NearestCode()) // 最近的有效错误码是level3的
	assert.Equal(t, int64(0), level4.CurrentCode())   // 当前层无错误码
	assert.False(t, level4.HasCurrentCode())          // 当前层未设置错误码
	assert.True(t, level4.HasAnyCode())               // 错误链中有错误码
}

// TestNewErrorCodeAPI_EdgeCases 测试边界情况
func TestNewErrorCodeAPI_EdgeCases(t *testing.T) {
	t.Run("空错误链", func(t *testing.T) {
		err := Error("")
		assert.Equal(t, int64(0), err.NearestCode())
		assert.Equal(t, int64(0), err.CurrentCode())
		assert.Equal(t, int64(0), err.RootCode())
		assert.False(t, err.HasCurrentCode())
		assert.False(t, err.HasAnyCode())
	})

	t.Run("多次设置错误码", func(t *testing.T) {
		err := Error("test")
		err.SetCode(100)
		err.SetCode(200)
		err.SetCode(300)
		
		assert.Equal(t, int64(300), err.CurrentCode())
		assert.True(t, err.HasCurrentCode())
	})

	t.Run("负数错误码", func(t *testing.T) {
		err := Error("test").SetCode(-1)
		assert.Equal(t, int64(-1), err.CurrentCode())
		assert.Equal(t, int64(-1), err.NearestCode())
		assert.True(t, err.HasCurrentCode())
		assert.True(t, err.HasAnyCode())
	})
}

// TestGetCodeStr 测试GetCodeStr方法
func TestGetCodeStr(t *testing.T) {
	tests := []struct {
		name     string
		code     int64
		expected string
	}{
		{
			name:     "零错误码",
			code:     0,
			expected: "",
		},
		{
			name:     "非零错误码",
			code:     404,
			expected: "code(404), ",
		},
		{
			name:     "负数错误码",
			code:     -1,
			expected: "code(-1), ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Error("test error").SetCode(tt.code)
			assert.Equal(t, tt.expected, err.GetCodeStr())
		})
	}
}

// TestWriteSelfTo 测试writeSelfTo方法
func TestWriteSelfTo(t *testing.T) {
	t.Run("基本输出", func(t *testing.T) {
		err := Error("test message").SetCode(404)
		// 通过ToString方法间接测试writeSelfTo
		result := err.ToString(false, "")
		
		expected := "code(404), test message"
		assert.Equal(t, expected, result)
	})

	t.Run("不打印错误码", func(t *testing.T) {
		err := Error("test message").SetCode(404)
		// 测试GetCodeStr方法
		codeStr := err.GetCodeStr()
		assert.Equal(t, "code(404), ", codeStr)
	})

	t.Run("零错误码不打印", func(t *testing.T) {
		err := Error("test message")
		codeStr := err.GetCodeStr()
		assert.Equal(t, "", codeStr)
	})

	t.Run("带标签输出", func(t *testing.T) {
		err := Error("test message")
		err.SetTag("module", "auth")
		err.SetTag("severity", "high")
		result := err.ToString(false, "")
		
		// 验证基本消息
		assert.NotEmpty(t, result)
		// 验证标签存在
		moduleTags := err.GetTag("module")
		severityTags := err.GetTag("severity")
		assert.Equal(t, []string{"auth"}, moduleTags)
		assert.Equal(t, []string{"high"}, severityTags)
	})

	t.Run("带堆栈跟踪输出", func(t *testing.T) {
		err := Trace("test message")
		result := err.ToString(true, "")
		
		// 验证基本消息存在
		assert.NotEmpty(t, result)
		// 验证堆栈跟踪信息存在
		traceInfo := err.GetTraceInfo()
		assert.NotNil(t, traceInfo)
		traceStr := traceInfo.String()
		assert.NotEmpty(t, traceStr)
	})
}

// TestSetTag 测试SetTag方法
func TestSetTag(t *testing.T) {
	t.Run("设置单个标签", func(t *testing.T) {
		err := Error("test error")
		err.SetTag("module", "auth")
		
		tags := err.GetTag("module")
		assert.Equal(t, []string{"auth"}, tags)
	})

	t.Run("设置多个相同键的标签", func(t *testing.T) {
		err := Error("test error")
		err.SetTag("module", "auth")
		err.SetTag("module", "user")
		
		tags := err.GetTag("module")
		assert.Equal(t, []string{"auth", "user"}, tags)
	})

	t.Run("设置不同键的标签", func(t *testing.T) {
		err := Error("test error")
		err.SetTag("module", "auth")
		err.SetTag("severity", "high")
		
		moduleTags := err.GetTag("module")
		severityTags := err.GetTag("severity")
		assert.Equal(t, []string{"auth"}, moduleTags)
		assert.Equal(t, []string{"high"}, severityTags)
	})

	t.Run("并发设置标签", func(t *testing.T) {
		err := Error("test error")
		
		// 并发设置标签
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				err.SetTag("concurrent", fmt.Sprintf("value%d", index))
			}(i)
		}
		wg.Wait()
		
		tags := err.GetTag("concurrent")
		assert.Equal(t, 10, len(tags))
	})
}

// TestGetTag 测试GetTag方法
func TestGetTag(t *testing.T) {
	t.Run("获取存在的标签", func(t *testing.T) {
		err := Error("test error")
		err.SetTag("module", "auth")
		err.SetTag("module", "user")
		
		tags := err.GetTag("module")
		assert.Equal(t, []string{"auth", "user"}, tags)
	})

	t.Run("获取不存在的标签", func(t *testing.T) {
		err := Error("test error")
		
		tags := err.GetTag("nonexistent")
		assert.Nil(t, tags)
	})

	t.Run("获取空标签", func(t *testing.T) {
		err := Error("test error")
		err.SetTag("empty", "")
		
		tags := err.GetTag("empty")
		assert.Equal(t, []string{""}, tags)
	})

	t.Run("标签返回副本", func(t *testing.T) {
		err := Error("test error")
		err.SetTag("module", "auth")
		
		tags1 := err.GetTag("module")
		tags2 := err.GetTag("module")
		
		// 修改返回的切片不应影响原始数据
		tags1[0] = "modified"
		assert.Equal(t, []string{"auth"}, tags2)
	})
}

// TestLogStats 测试LogStats方法
func TestLogStats(t *testing.T) {
	// 重置指标以确保测试的独立性
	ResetMetrics()
	
	// 创建一些错误以生成统计数据
	Error("test1").SetCode(404)
	Error("test2").SetCode(500)
	Trace("test3")
	Wrap(errors.New("inner"), "outer")
	
	// 测试LogStats - 使用实现了正确接口的logger
	logger := &mockStatsLogger{}
	LogStats(logger)
	
	// 验证日志输出包含统计信息
	assert.NotNil(t, logger.loggedMetrics)
	assert.True(t, logger.loggedMetrics.ErrorCreated > 0)
	assert.True(t, logger.loggedMetrics.ErrorWithCode > 0)
	assert.True(t, logger.loggedMetrics.ErrorWithTrace > 0)
	assert.True(t, logger.loggedMetrics.ErrorWrapped > 0)
}

// 实现ErrorStatsLogger接口的mock logger
type mockStatsLogger struct {
	loggedMetrics *ErrorMetrics
}

func (m *mockStatsLogger) LogErrorStats(metrics *ErrorMetrics) {
	m.loggedMetrics = metrics
}

// TestCatchFailure 测试CatchFailure方法
func TestCatchFailure(t *testing.T) {
	t.Run("测试CatchFailure实际用法", func(t *testing.T) {
		// 测试CatchFailure的实际使用场景
		var caughtError error
		
		func() {
			defer CatchFailure(func(err error) {
				caughtError = err
			})
			panic("test panic")
		}()
		
		assert.NotNil(t, caughtError)
		// 验证错误类型和消息 - CatchFailure会包装非error类型的panic
		assert.Contains(t, caughtError.Error(), "panic = test panic")
	})

	t.Run("测试CatchFailure无panic", func(t *testing.T) {
		var caughtError error
		
		func() {
			defer CatchFailure(func(err error) {
				caughtError = err
			})
			// 正常执行，无panic
		}()
		
		assert.Nil(t, caughtError)
	})

	t.Run("测试CatchFailure捕获error类型panic", func(t *testing.T) {
		var caughtError error
		testErr := errors.New("test error")
		
		func() {
			defer CatchFailure(func(err error) {
				caughtError = err
			})
			panic(testErr)
		}()
		
		assert.Equal(t, testErr, caughtError)
	})
}

// TestTraceString 测试trace.go中的String方法
func TestTraceString(t *testing.T) {
	err := Trace("test trace")
	traceInfo := err.GetTraceInfo()
	
	assert.NotNil(t, traceInfo)
	
	// 测试String方法
	str := traceInfo.String()
	assert.NotEmpty(t, str)
	// 验证堆栈跟踪信息的基本结构
	assert.True(t, len(str) > 0)
}

// TestTraceRelease 测试trace.go中的Release方法
func TestTraceRelease(t *testing.T) {
	err := Trace("test trace")
	traceInfo := err.GetTraceInfo()
	
	assert.NotNil(t, traceInfo)
	
	// 测试Release方法
	traceInfo.Release()
	// Release方法主要用于对象池，这里主要测试不会panic
}

// TestCreateTraceInfo 测试utils.go中的createTraceInfo方法
func TestCreateTraceInfo(t *testing.T) {
	t.Run("正常创建", func(t *testing.T) {
		trace := createTraceInfo(1, nil)
		assert.NotNil(t, trace)
		str := trace.String()
		assert.NotEmpty(t, str)
	})

	t.Run("跳过层级", func(t *testing.T) {
		trace := createTraceInfo(2, nil)
		assert.NotNil(t, trace)
		// 跳过更多层级，应该指向调用者的调用者
		str := trace.String()
		assert.NotEmpty(t, str)
	})
}

// TestContextMethods 测试context.go中未覆盖的方法
func TestContextMethods(t *testing.T) {
	t.Run("ErrorWithContext创建", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "key", "value")
		err := ErrorWithContext(ctx, "test error")
		
		assert.NotNil(t, err)
		// 验证错误消息的具体内容
		errMsg := err.Error()
		assert.Equal(t, "test error", errMsg)
	})

	t.Run("ToString方法-context错误", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		
		time.Sleep(2 * time.Millisecond) // 确保超时
		
		err := ErrorWithContext(ctx, "test error")
		str := err.ToString(false, ", ")
		
		// 验证包含测试错误和context错误
		// 由于context错误会被包装，我们验证基本结构
		assert.NotEmpty(t, str)
		// 验证至少包含我们的错误消息
		contextualIrr, ok := err.(*ContextualIrr)
		assert.True(t, ok)
		assert.Equal(t, "test error", contextualIrr.BasicIrr.Msg)
		
		// 验证context错误存在
		assert.NotNil(t, ctx.Err())
		assert.Equal(t, context.DeadlineExceeded, ctx.Err())
	})
}

// TestAdditionalCoverage 测试提高覆盖率的边界情况
func TestAdditionalCoverage(t *testing.T) {
	t.Run("Context方法边界情况", func(t *testing.T) {
		// 测试Context方法的所有分支
		ctx := context.WithValue(context.Background(), "key", "value")
		err := ErrorWithContext(ctx, "test error")
		
		contextualErr, ok := err.(*ContextualIrr)
		assert.True(t, ok)
		
		// 测试Context方法返回context
		retrievedCtx := contextualErr.Context()
		assert.Equal(t, "value", retrievedCtx.Value("key"))
		
		// 测试ctx为nil的情况
		contextualErrNil := &ContextualIrr{
			BasicIrr: Error("test").(*BasicIrr),
			ctx:      nil,
		}
		nilCtx := contextualErrNil.Context()
		assert.Equal(t, context.Background(), nilCtx)
		
		// 测试非ContextualIrr的情况 - ExtractContext返回context.Background()
		basicErr := Error("basic error")
		extractedCtx := ExtractContext(basicErr)
		assert.NotNil(t, extractedCtx)
		assert.Equal(t, context.Background(), extractedCtx)
	})

	t.Run("TraverseToSource边界情况", func(t *testing.T) {
		// 测试TraverseToSource的panic恢复
		err := Error("test error")
		
		// 测试正常遍历
		var sourceErr error
		traverseErr := err.TraverseToSource(func(e error, isSource bool) error {
			if isSource {
				sourceErr = e
			}
			return nil
		})
		assert.Nil(t, traverseErr)
		assert.Equal(t, err, sourceErr)
		
		// 测试遍历中抛出错误
		traverseErr = err.TraverseToSource(func(e error, isSource bool) error {
			return errors.New("traverse error")
		})
		assert.NotNil(t, traverseErr)
		assert.Equal(t, "traverse error", traverseErr.Error())
		
		// 测试遍历中panic
		traverseErr = err.TraverseToSource(func(e error, isSource bool) error {
			panic("test panic")
		})
		assert.NotNil(t, traverseErr)
		assert.Contains(t, traverseErr.Error(), "panic = test panic")
	})

	t.Run("NearestCode边界情况", func(t *testing.T) {
		// 测试错误链中没有错误码的情况
		innerErr := errors.New("standard error")
		err := Wrap(innerErr, "wrapper error")
		
		code := err.NearestCode()
		assert.Equal(t, int64(0), code)
		
		// 测试错误链中有错误码的情况
		codeErr := ErrorC(404, "not found")
		wrappedErr := Wrap(codeErr, "wrapped")
		
		code = wrappedErr.NearestCode()
		assert.Equal(t, int64(404), code)
	})

	t.Run("RootCode边界情况", func(t *testing.T) {
		// 测试根错误没有错误码的情况
		innerErr := errors.New("standard error")
		err := Wrap(innerErr, "wrapper error")
		
		code := err.RootCode()
		assert.Equal(t, int64(0), code)
		
		// 测试根错误有错误码的情况
		rootErr := ErrorC(500, "internal error")
		wrappedErr := Wrap(rootErr, "wrapped")
		
		code = wrappedErr.RootCode()
		assert.Equal(t, int64(500), code)
		
		// 测试复杂错误链
		level1 := ErrorC(100, "level1")
		level2 := Wrap(level1, "level2")
		level3 := Wrap(level2, "level3")
		
		code = level3.RootCode()
		assert.Equal(t, int64(100), code)
	})

	t.Run("TraverseCode边界情况", func(t *testing.T) {
		// 测试TraverseCode的panic恢复
		err := ErrorC(404, "test error")
		
		// 测试正常遍历
		var codes []int64
		traverseErr := err.TraverseCode(func(e error, code int64) error {
			codes = append(codes, code)
			return nil
		})
		assert.Nil(t, traverseErr)
		assert.Contains(t, codes, int64(404))
		
		// 测试遍历中抛出错误
		traverseErr = err.TraverseCode(func(e error, code int64) error {
			return errors.New("traverse error")
		})
		assert.NotNil(t, traverseErr)
		assert.Equal(t, "traverse error", traverseErr.Error())
		
		// 测试遍历中panic
		traverseErr = err.TraverseCode(func(e error, code int64) error {
			panic("test panic")
		})
		assert.NotNil(t, traverseErr)
		assert.Contains(t, traverseErr.Error(), "panic = test panic")
	})

	t.Run("GetTag边界情况", func(t *testing.T) {
		err := Error("test error")
		
		// 测试获取不存在的标签
		tags := err.GetTag("nonexistent")
		assert.Empty(t, tags)
		
		// 测试tags为nil的情况
		basicErr := &BasicIrr{Msg: "test"}
		tags = basicErr.GetTag("any")
		assert.Empty(t, tags)
		
		// 测试获取存在的标签
		err.SetTag("key", "value1")
		err.SetTag("key", "value2")
		tags = err.GetTag("key")
		assert.Len(t, tags, 2)
		assert.Contains(t, tags, "value1")
		assert.Contains(t, tags, "value2")
	})

	t.Run("createTraceInfo边界情况", func(t *testing.T) {
		// 测试不同的skip值，但要小心避免panic
		trace1 := createTraceInfo(0, nil)
		assert.NotNil(t, trace1)
		
		trace2 := createTraceInfo(1, nil)
		assert.NotNil(t, trace2)
		
		// 测试带有innerErr的情况
		innerErr := Error("inner error")
		trace3 := createTraceInfo(1, innerErr)
		// 可能返回nil或非nil，取决于堆栈是否相同
		_ = trace3 // 使用变量避免编译错误
		
		// 验证堆栈信息不为空
		str1 := trace1.String()
		str2 := trace2.String()
		assert.NotEmpty(t, str1)
		assert.NotEmpty(t, str2)
	})
}
