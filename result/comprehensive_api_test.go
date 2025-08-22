package result

import (
	"errors"
	"testing"

	"github.com/khicago/irr"
	"github.com/stretchr/testify/assert"
)

// ========== 测试未覆盖的构造函数 ==========

func TestOkIRR(t *testing.T) {
	value := "success"
	r := OkIRR(value)

	assert.True(t, r.IsOk())
	assert.False(t, r.IsErr())
	assert.Equal(t, value, r.Unwrap())
}

func TestErrIRR(t *testing.T) {
	err := irr.New("test error").Build()
	r := ErrIRR[string](err)

	assert.False(t, r.IsOk())
	assert.True(t, r.IsErr())
	assert.Equal(t, err, r.UnwrapErr())
}

func TestOkError(t *testing.T) {
	value := 42
	r := OkError(value)

	assert.True(t, r.IsOk())
	assert.Equal(t, value, r.Unwrap())
}

func TestErrError(t *testing.T) {
	err := errors.New("test error")
	r := ErrError[int](err)

	assert.True(t, r.IsErr())
	assert.Equal(t, err, r.UnwrapErr())
}

// ========== 测试状态检查方法 ==========

func TestIsOkIsErr(t *testing.T) {
	t.Run("Ok result", func(t *testing.T) {
		r := Ok[string, error]("success")
		assert.True(t, r.IsOk())
		assert.False(t, r.IsErr())
	})

	t.Run("Err result", func(t *testing.T) {
		r := Err[string, error](errors.New("fail"))
		assert.False(t, r.IsOk())
		assert.True(t, r.IsErr())
	})
}

// ========== 测试值提取方法 ==========

func TestValue(t *testing.T) {
	t.Run("Ok result", func(t *testing.T) {
		expected := "success"
		r := Ok[string, error](expected)

		value, ok := r.Value()
		assert.True(t, ok)
		assert.Equal(t, expected, value)
	})

	t.Run("Err result", func(t *testing.T) {
		r := Err[string, error](errors.New("fail"))

		value, ok := r.Value()
		assert.False(t, ok)
		assert.Empty(t, value)
	})
}

func TestError(t *testing.T) {
	t.Run("Ok result", func(t *testing.T) {
		r := Ok[string, error]("success")

		err, hasErr := r.Error()
		assert.False(t, hasErr)
		assert.Nil(t, err)
	})

	t.Run("Err result", func(t *testing.T) {
		expected := errors.New("fail")
		r := Err[string, error](expected)

		err, hasErr := r.Error()
		assert.True(t, hasErr)
		assert.Equal(t, expected, err)
	})
}

func TestExpectWithPanicBehavior(t *testing.T) {
	t.Run("Ok result", func(t *testing.T) {
		expected := "success"
		r := Ok[string, error](expected)

		result := r.Expect("should not panic")
		assert.Equal(t, expected, result)
	})

	t.Run("Err result panics", func(t *testing.T) {
		r := Err[string, error](errors.New("fail"))

		assert.Panics(t, func() {
			r.Expect("expected error")
		})
	})
}

// ========== 测试模式匹配 ==========

func TestMatchPatternHandling(t *testing.T) {
	t.Run("Ok result", func(t *testing.T) {
		expected := "success"
		r := Ok[string, error](expected)

		value, err := r.Match()
		assert.Equal(t, expected, value)
		assert.Nil(t, err)
	})

	t.Run("Err result", func(t *testing.T) {
		expected := errors.New("fail")
		r := Err[string, error](expected)

		value, err := r.Match()
		assert.Empty(t, value)
		assert.Equal(t, expected, err)
	})
}

func TestMatchFunc(t *testing.T) {
	t.Run("Ok result", func(t *testing.T) {
		expected := "success"
		r := Ok[string, error](expected)

		var capturedValue string
		r.MatchFunc(
			func(v string) { capturedValue = v },
			func(e error) { t.Error("should not call error handler") },
		)

		assert.Equal(t, expected, capturedValue)
	})

	t.Run("Err result", func(t *testing.T) {
		expected := errors.New("fail")
		r := Err[string, error](expected)

		var capturedError error
		r.MatchFunc(
			func(v string) { t.Error("should not call success handler") },
			func(e error) { capturedError = e },
		)

		assert.Equal(t, expected, capturedError)
	})
}

// ========== 测试类型别名 ==========

func TestResultIRR(t *testing.T) {
	t.Run("Success case", func(t *testing.T) {
		var r ResultIRR[string] = OkIRR("success")
		assert.True(t, r.IsOk())
		assert.Equal(t, "success", r.Unwrap())
	})

	t.Run("Error case", func(t *testing.T) {
		err := irr.New("test error").Build()
		var r ResultIRR[string] = ErrIRR[string](err)
		assert.True(t, r.IsErr())
		assert.Equal(t, err, r.UnwrapErr())
	})
}

func TestResultError(t *testing.T) {
	t.Run("Success case", func(t *testing.T) {
		var r ResultError[int] = OkError(42)
		assert.True(t, r.IsOk())
		assert.Equal(t, 42, r.Unwrap())
	})

	t.Run("Error case", func(t *testing.T) {
		err := errors.New("test error")
		var r ResultError[int] = ErrError[int](err)
		assert.True(t, r.IsErr())
		assert.Equal(t, err, r.UnwrapErr())
	})
}

// ========== 边界条件测试 ==========

func TestEdgeCases(t *testing.T) {
	t.Run("Zero values", func(t *testing.T) {
		// 测试零值
		r := Ok[int, error](0)
		assert.True(t, r.IsOk())
		assert.Equal(t, 0, r.Unwrap())

		// 测试空字符串
		r2 := Ok[string, error]("")
		assert.True(t, r2.IsOk())
		assert.Equal(t, "", r2.Unwrap())
	})

	t.Run("Nil error", func(t *testing.T) {
		// 测试nil错误
		r := Err[string, error](nil)
		assert.True(t, r.IsErr())
		assert.Nil(t, r.UnwrapErr())
	})

	t.Run("Complex types", func(t *testing.T) {
		// 测试复杂类型
		type ComplexType struct {
			Name string
			Data map[string]int
		}

		complex := ComplexType{
			Name: "test",
			Data: map[string]int{"key": 42},
		}

		r := Ok[ComplexType, error](complex)
		assert.True(t, r.IsOk())
		result := r.Unwrap()
		assert.Equal(t, "test", result.Name)
		assert.Equal(t, 42, result.Data["key"])
	})
}

// ========== 性能基准测试 ==========

func BenchmarkResultCreation(b *testing.B) {
	b.Run("Ok creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Ok[string, error]("benchmark")
		}
	})

	b.Run("Err creation", func(b *testing.B) {
		err := errors.New("benchmark error")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Err[string, error](err)
		}
	})
}

func BenchmarkResultOperations(b *testing.B) {
	r := Ok[string, error]("benchmark")

	b.Run("IsOk check", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = r.IsOk()
		}
	})

	b.Run("Unwrap", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = r.Unwrap()
		}
	})

	b.Run("Match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = r.Match()
		}
	})
}

// ========== 示例测试（作为文档） ==========

func ExampleOkIRR() {
	result := OkIRR("success")
	if result.IsOk() {
		value := result.Unwrap()
		println("Got value:", value)
	}
	// Output:
}

func ExampleErrIRR() {
	err := irr.New("database connection failed").Build()
	result := ErrIRR[string](err)
	if result.IsErr() {
		err := result.UnwrapErr()
		println("Got error:", err.Error())
	}
	// Output:
}

func ExampleResult_Match() {
	result := Ok[string, error]("hello")

	value, err := result.Match()
	if err == nil {
		println("Success:", value)
	} else {
		println("Error:", err.Error())
	}
	// Output:
}

func ExampleResult_MatchFunc() {
	result := Err[string, error](errors.New("failed"))

	result.MatchFunc(
		func(value string) {
			println("Success:", value)
		},
		func(err error) {
			println("Error:", err.Error())
		},
	)
	// Output:
}
