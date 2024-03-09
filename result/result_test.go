package result

import (
	"errors"
	"testing"
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
