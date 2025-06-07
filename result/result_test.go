package result

import (
	"errors"
	"testing"
	"sync"

	"github.com/stretchr/testify/assert"
)

func TestOK(t *testing.T) {
	r := OK("success")
	if !r.Ok() || r.Err() != nil {
		t.Errorf("Expected OK to produce a Result with no error, got %v", r.Err())
	}
}

func TestErr(t *testing.T) {
	testError := errors.New("fail")
	r := Err[string](testError)
	if r.Ok() || r.Err() == nil {
		t.Errorf("Expected Err to produce a Result with an error, got %v", r.Err())
	}
	if r.Err() != testError {
		t.Errorf("Expected Err to produce the same error, got %v", r.Err())
	}
}

func TestUnwrap(t *testing.T) {
	expected := "success"
	r := OK(expected)
	result := r.Unwrap()
	if result != expected {
		t.Errorf("Expected Unwrap to return %v, got %v", expected, result)
	}
}

func TestUnwrapOr(t *testing.T) {
	expected := "default"
	r := Err[string](errors.New("fail"))
	result := r.UnwrapOr(expected)
	if result != expected {
		t.Errorf("Expected UnwrapOr to return %v, got %v", expected, result)
	}
}

// 注意：因为 UnwrapErr 和 Expect 会 panic，所以在测试中需要特别处理
func TestUnwrapErr(t *testing.T) {
	testError := errors.New("fail")
	r := Err[string](testError)
	if r.UnwrapErr() != testError {
		t.Errorf("Expected UnwrapErr to return %v, got %v", testError, r.Err())
	}

	defer func() {
		if recover() == nil {
			t.Errorf("Expected UnwrapErr to panic on an Ok Result")
		}
	}()
	// 这里应该会 panic，因为是一个 OK 结果
	OK("success").UnwrapErr()
}

// TestUnwrapErrWithOK 测试 UnwrapErr 方法在 Result 为 OK 时触发 panic 的行为
func TestUnwrapErrWithOK(t *testing.T) {
	r := OK("should panic")
	defer func() {
		if rec := recover(); rec == nil {
			t.Errorf("Expected panic when calling UnwrapErr on an Ok Result")
		}
	}()
	_ = r.UnwrapErr() // 这里应该 panic
}

func TestExpect(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("Expected Expect to panic on an error Result")
		}
	}()
	// 这里应该会 panic，因为是一个 Err 结果
	Err[string](errors.New("fail")).Expect("should panic")
}

// TestExpectSuccess 测试 Expect 在没有错误时的行为
func TestExpectSuccess(t *testing.T) {
	expectedResult := "success"
	r := OK(expectedResult)
	res := r.Expect("Expected success, got an error")
	if res != expectedResult {
		t.Errorf("Expected Expect to return %v, got %v", expectedResult, res)
	}
}

// TestMatch 测试 Match 方法正确返回结果或错误
func TestMatch(t *testing.T) {
	expectedResult := "success"
	r := OK(expectedResult)
	res, err := r.Match()
	if err != nil || res != expectedResult {
		t.Errorf("Expected Match to return result %v with no error, got result `%v` with error `%v`", expectedResult, res, err)
	}

	expectedErr := errors.New("fail")
	r = Err[string](expectedErr)
	res, err = r.Match()
	if err != expectedErr || res != "" { // 假设零值为 "" (空字符串)
		t.Errorf("Expected Match to return zero value and error %v, got result `%v` with error `%v`", expectedErr, res, err)
	}
}

// TestAndThen 测试 AndThen 方法能够在 Result 为 Ok 时应用函数，并且正确传播错误
func TestAndThen(t *testing.T) {
	okResult := OK("success")
	secondOp := func(s string) Result[int] {
		return OK(len(s))
	}
	newResult := AndThen(okResult, secondOp)
	if newResult.Ok() {
		expectedLength := len("success")
		if res, _ := newResult.Match(); res != expectedLength {
			t.Errorf("Expected AndThen to return %v, got %v", expectedLength, res)
		}
	} else {
		t.Errorf("Expected AndThen to succeed when provided with an Ok Result, got error %v", newResult.Err())
	}

	errResult := Err[string](errors.New("fail"))
	newResult = AndThen(errResult, secondOp)
	if newResult.Ok() {
		t.Errorf("Expected AndThen to propagate error, got an Ok Result")
	}
}

// TestResultAdditionalMethods 测试额外的方法
func TestResultAdditionalMethods(t *testing.T) {
	t.Run("Ok方法", func(t *testing.T) {
		okResult := OK(42)
		errResult := Err[int](errors.New("test error"))
		
		assert.True(t, okResult.Ok())
		assert.False(t, errResult.Ok())
	})

	t.Run("Err方法", func(t *testing.T) {
		okResult := OK(42)
		errResult := Err[int](errors.New("test error"))
		
		assert.Nil(t, okResult.Err())
		assert.NotNil(t, errResult.Err())
	})

	t.Run("Unwrap方法-错误情况", func(t *testing.T) {
		errResult := Err[int](errors.New("test error"))
		
		assert.Panics(t, func() {
			errResult.Unwrap()
		})
	})

	t.Run("UnwrapErr方法-成功情况", func(t *testing.T) {
		okResult := OK(42)
		
		assert.Panics(t, func() {
			okResult.UnwrapErr()
		})
	})

	t.Run("UnwrapErr方法-错误情况", func(t *testing.T) {
		testErr := errors.New("test error")
		errResult := Err[int](testErr)
		
		err := errResult.UnwrapErr()
		assert.Equal(t, testErr, err)
	})

	t.Run("UnwrapOr方法-成功情况", func(t *testing.T) {
		okResult := OK(42)
		value := okResult.UnwrapOr(0)
		assert.Equal(t, 42, value)
	})

	t.Run("UnwrapOr方法-错误情况", func(t *testing.T) {
		errResult := Err[int](errors.New("test error"))
		value := errResult.UnwrapOr(99)
		assert.Equal(t, 99, value)
	})

	t.Run("Match方法-成功情况", func(t *testing.T) {
		okResult := OK(42)
		value, err := okResult.Match()
		assert.Equal(t, 42, value)
		assert.Nil(t, err)
	})

	t.Run("Match方法-错误情况", func(t *testing.T) {
		testErr := errors.New("test error")
		errResult := Err[int](testErr)
		value, err := errResult.Match()
		assert.Equal(t, 0, value)
		assert.Equal(t, testErr, err)
	})

	t.Run("Expect方法-成功情况", func(t *testing.T) {
		okResult := OK(42)
		value := okResult.Expect("should not panic")
		assert.Equal(t, 42, value)
	})

	t.Run("Expect方法-错误情况", func(t *testing.T) {
		errResult := Err[int](errors.New("test error"))
		assert.Panics(t, func() {
			errResult.Expect("custom message")
		})
	})
}

// TestResultComplexChaining 测试复杂的链式操作
func TestResultComplexChaining(t *testing.T) {
	t.Run("成功的复杂链", func(t *testing.T) {
		result := OK(10)
		
		// 测试基本的链式操作
		value1 := result.UnwrapOr(0)
		assert.Equal(t, 10, value1)
		
		// 测试Match方法
		value2, err := result.Match()
		assert.Equal(t, 10, value2)
		assert.Nil(t, err)
	})

	t.Run("中途失败的复杂链", func(t *testing.T) {
		errResult := Err[int](errors.New("middle error"))
		
		// 测试错误情况的链式操作
		value := errResult.UnwrapOr(99)
		assert.Equal(t, 99, value)
		
		// 测试Match方法
		value2, err := errResult.Match()
		assert.Equal(t, 0, value2)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "middle error")
	})
}

// TestResultEdgeCases 测试边界情况
func TestResultEdgeCases(t *testing.T) {
	t.Run("零值处理", func(t *testing.T) {
		zeroResult := OK(0)
		assert.True(t, zeroResult.Ok())
		assert.Equal(t, 0, zeroResult.Unwrap())
	})

	t.Run("空字符串处理", func(t *testing.T) {
		emptyResult := OK("")
		assert.True(t, emptyResult.Ok())
		assert.Equal(t, "", emptyResult.Unwrap())
	})

	t.Run("nil指针处理", func(t *testing.T) {
		var ptr *int
		nilResult := OK(ptr)
		assert.True(t, nilResult.Ok())
		assert.Nil(t, nilResult.Unwrap())
	})

	t.Run("大数值处理", func(t *testing.T) {
		bigNum := int64(9223372036854775807) // math.MaxInt64
		bigResult := OK(bigNum)
		assert.True(t, bigResult.Ok())
		assert.Equal(t, bigNum, bigResult.Unwrap())
	})
}

// TestResultConcurrency 测试并发安全性
func TestResultConcurrency(t *testing.T) {
	t.Run("并发读取", func(t *testing.T) {
		result := OK(42)
		
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.True(t, result.Ok())
				assert.Equal(t, 42, result.Unwrap())
			}()
		}
		wg.Wait()
	})

	t.Run("并发UnwrapOr操作", func(t *testing.T) {
		result := OK(10)
		
		var wg sync.WaitGroup
		results := make([]int, 100)
		
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				results[index] = result.UnwrapOr(index)
			}(i)
		}
		wg.Wait()
		
		for _, r := range results {
			assert.Equal(t, 10, r)
		}
	})
}
