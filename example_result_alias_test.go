package irr_test

import (
	"testing"

	"github.com/khicago/irr"
	"github.com/khicago/irr/result"
)

// TestGenericAlias 验证Go 1.23+泛型别名功能
func TestGenericAlias(t *testing.T) {
	// 测试ResultIRR语法糖
	userResult := result.OkIRR(User{ID: 1, Name: "Alice"})

	// 验证类型别名完全透明
	if !userResult.IsOk() {
		t.Error("Expected Ok result")
	}

	user, ok := userResult.Value()
	if !ok || user.ID != 1 {
		t.Error("Failed to extract user value")
	}

	// 测试错误情况
	errResult := result.ErrIRR[User](irr.ErrorC(404, "user not found"))
	if errResult.IsOk() {
		t.Error("Expected Err result")
	}

	err, ok := errResult.Error()
	if !ok {
		t.Error("Failed to extract error")
	}

	// 验证IRR接口
	if coder, ok := err.(interface{ Code() int64 }); ok {
		if coder.Code() != 404 {
			t.Errorf("Expected code 404, got %d", coder.Code())
		}
	} else {
		t.Error("Expected IRR with Code method")
	}
}

// TestResultErrorAlias 验证ResultError别名
func TestResultErrorAlias(t *testing.T) {
	// 标准库兼容性测试
	dataResult := result.FromTuple([]byte("test"), nil)

	if !dataResult.IsOk() {
		t.Error("Expected Ok result from tuple")
	}

	data := dataResult.UnwrapOr([]byte("default"))
	if string(data) != "test" {
		t.Errorf("Expected 'test', got '%s'", string(data))
	}
}

type User struct {
	ID   int
	Name string
}
