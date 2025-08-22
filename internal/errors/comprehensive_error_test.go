package errors

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ========== 测试格式化方法 ==========

func TestErrorImpl_Format(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := NewErrorImpl("wrapper error", 500, cause, map[string][]string{
		"service": {"api"},
		"method":  {"POST"},
	}, []Frame{
		{File: "/path/to/file.go", Function: "TestFunction", Line: 42, Package: "github.com/test"},
	}, context.Background())

	t.Run("Format with %v", func(t *testing.T) {
		result := fmt.Sprintf("%v", err)
		assert.Equal(t, "wrapper error, root cause", result)
	})

	t.Run("Format with %+v", func(t *testing.T) {
		result := fmt.Sprintf("%+v", err)
		assert.Contains(t, result, "wrapper error")
		assert.Contains(t, result, "code: 500")
		assert.Contains(t, result, "tags: [")
		assert.Contains(t, result, "service=api")
		assert.Contains(t, result, "method=POST")
		assert.Contains(t, result, "stack:")
	})

	t.Run("Format with %s", func(t *testing.T) {
		result := fmt.Sprintf("%s", err)
		assert.Equal(t, "wrapper error, root cause", result)
	})

	t.Run("Format with %q", func(t *testing.T) {
		result := fmt.Sprintf("%q", err)
		assert.Equal(t, "\"wrapper error, root cause\"", result)
	})
}

func TestErrorImpl_Details(t *testing.T) {
	err := NewErrorImpl("test error", 404, nil, map[string][]string{
		"type": {"validation"},
	}, []Frame{
		{File: "/test.go", Function: "TestFunc", Line: 10, Package: "test"},
	}, nil)

	details := err.Details()
	assert.Contains(t, details, "test error")
	assert.Contains(t, details, "code: 404")
	assert.Contains(t, details, "tags: [type=validation]")
	assert.Contains(t, details, "stack:")
	assert.Contains(t, details, "/test.go:10 in TestFunc")
}

func TestErrorImpl_ToString(t *testing.T) {
	t.Run("With stack trace", func(t *testing.T) {
		err := NewErrorImpl("error1", 0, nil, nil, []Frame{
			{File: "/test1.go", Function: "Func1", Line: 1},
			{File: "/test2.go", Function: "Func2", Line: 2},
		}, nil)

		output := err.ToString(true, " | ")
		// Test that output contains the expected components (first frame only)
		assert.Contains(t, output, "error1")
		assert.Contains(t, output, "Func1")
		assert.Contains(t, output, "/test1.go:1")
		// ToString typically shows only the first frame, so don't assert on second frame
	})

	t.Run("Without stack trace", func(t *testing.T) {
		cause := fmt.Errorf("cause")
		err := NewErrorImpl("wrapper", 0, cause, nil, nil, nil)

		output := err.ToString(false, " -> ")
		assert.Equal(t, "wrapper -> cause", output)
	})

	t.Run("No cause, no trace", func(t *testing.T) {
		err := NewErrorImpl("simple", 0, nil, nil, nil, nil)
		output := err.ToString(false, " -> ")
		assert.Equal(t, "simple", output)
	})
}

// ========== 测试错误遍历方法 ==========

func TestErrorImpl_TraverseToRoot(t *testing.T) {
	// 创建错误链：err3 -> err2 -> err1
	err1 := fmt.Errorf("root")
	err2 := NewErrorImpl("middle", 0, err1, nil, nil, nil)
	err3 := NewErrorImpl("top", 0, err2, nil, nil, nil)

	t.Run("Complete traversal", func(t *testing.T) {
		var visited []string
		result := err3.TraverseToRoot(func(err error) error {
			visited = append(visited, err.Error())
			return nil // Continue
		})

		assert.Nil(t, result)
		assert.Len(t, visited, 3)
		assert.Contains(t, visited[0], "top")
		assert.Contains(t, visited[1], "middle")
		assert.Equal(t, "root", visited[2])
	})

	t.Run("Early termination", func(t *testing.T) {
		stopErr := fmt.Errorf("stop")
		var visited []string

		result := err3.TraverseToRoot(func(err error) error {
			visited = append(visited, err.Error())
			if err.Error() == "middle, root" {
				return stopErr
			}
			return nil
		})

		assert.Equal(t, stopErr, result)
		assert.Len(t, visited, 2) // Should stop at middle
	})
}

func TestErrorImpl_TraverseToSource(t *testing.T) {
	err1 := fmt.Errorf("source")
	err2 := NewErrorImpl("wrapper", 0, err1, nil, nil, nil)

	var results []struct {
		message  string
		isSource bool
	}

	result := err2.TraverseToSource(func(err error, isSource bool) error {
		results = append(results, struct {
			message  string
			isSource bool
		}{err.Error(), isSource})
		return nil
	})

	assert.Nil(t, result)
	assert.Len(t, results, 2)
	assert.False(t, results[0].isSource) // wrapper is not source
	assert.True(t, results[1].isSource)  // source is source
}

// ========== 测试 Root 方法 ==========

func TestErrorImpl_Root(t *testing.T) {
	t.Run("Chain with multiple levels", func(t *testing.T) {
		err1 := fmt.Errorf("root")
		err2 := NewErrorImpl("middle", 0, err1, nil, nil, nil)
		err3 := NewErrorImpl("top", 0, err2, nil, nil, nil)

		root := err3.Root()
		assert.Equal(t, err1, root)
	})

	t.Run("Single error", func(t *testing.T) {
		err := NewErrorImpl("single", 0, nil, nil, nil, nil)
		root := err.Root()
		assert.Equal(t, err, root)
	})
}

// ========== 测试边界条件 ==========

func TestErrorImpl_EdgeCases(t *testing.T) {
	t.Run("Nil context handling", func(t *testing.T) {
		err := NewErrorImpl("test", 0, nil, nil, nil, nil)
		ctx := err.Context()
		assert.Equal(t, context.Background(), ctx)
	})

	t.Run("Empty tags handling", func(t *testing.T) {
		err := NewErrorImpl("test", 0, nil, nil, nil, nil)
		tags := err.Tags()
		assert.NotNil(t, tags)
		assert.Empty(t, tags)
	})

	t.Run("Nil tags modification", func(t *testing.T) {
		err := NewErrorImpl("test", 0, nil, nil, nil, nil)
		err.SetTag("key", "value")

		tags := err.Tags()
		assert.Equal(t, []string{"value"}, tags["key"])
	})

	t.Run("GetTag with nil tags", func(t *testing.T) {
		err := NewErrorImpl("test", 0, nil, nil, nil, nil)
		values := err.GetTag("nonexistent")
		assert.Nil(t, values)
	})

	t.Run("GetTag with nonexistent key", func(t *testing.T) {
		err := NewErrorImpl("test", 0, nil, map[string][]string{
			"existing": {"value"},
		}, nil, nil)

		values := err.GetTag("nonexistent")
		assert.Nil(t, values)

		existing := err.GetTag("existing")
		assert.Equal(t, []string{"value"}, existing)
	})

	t.Run("Empty trace handling", func(t *testing.T) {
		err := NewErrorImpl("test", 0, nil, nil, []Frame{}, nil)
		assert.False(t, err.HasStackTrace())
		assert.Empty(t, err.StackTrace())
	})
}

// ========== 测试并发安全性 ==========

func TestErrorImpl_ConcurrentAccess(t *testing.T) {
	err := NewErrorImpl("test", 0, nil, map[string][]string{
		"initial": {"value"},
	}, nil, nil)

	// 并发读取
	t.Run("Concurrent reads", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()

				// 多次读取应该是安全的
				for j := 0; j < 100; j++ {
					_ = err.Error()
					_ = err.Tags()
					_ = err.Code()
					_ = err.HasCode()
					_ = err.HasStackTrace()
					_ = err.Context()
				}
			}()
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	// 注意：tags的并发写入需要外部同步，这里只测试读取安全性
	t.Run("Concurrent tag reads", func(t *testing.T) {
		done := make(chan bool, 5)

		for i := 0; i < 5; i++ {
			go func() {
				defer func() { done <- true }()

				for j := 0; j < 100; j++ {
					tags := err.Tags()
					assert.NotNil(t, tags)

					values := err.GetTag("initial")
					assert.Equal(t, []string{"value"}, values)
				}
			}()
		}

		for i := 0; i < 5; i++ {
			<-done
		}
	})
}

// ========== 性能基准测试 ==========

func BenchmarkErrorImpl(b *testing.B) {
	err := NewErrorImpl("benchmark error", 500, nil, map[string][]string{
		"service": {"api"},
		"method":  {"GET"},
	}, []Frame{
		{File: "/bench.go", Function: "BenchFunc", Line: 1},
	}, context.Background())

	b.Run("Error()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = err.Error()
		}
	})

	b.Run("Tags()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = err.Tags()
		}
	})

	b.Run("GetTag()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = err.GetTag("service")
		}
	})

	b.Run("HasCode()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = err.HasCode()
		}
	})

	b.Run("Details()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = err.Details()
		}
	})
}

// ========== 示例测试 ==========

func ExampleNewErrorImpl_withAllFeatures() {
	err := NewErrorImpl(
		"database query failed",
		1001,
		fmt.Errorf("connection timeout"),
		map[string][]string{
			"table":    {"users"},
			"query_id": {"12345"},
		},
		[]Frame{
			{File: "/app/db.go", Function: "QueryUsers", Line: 42},
		},
		context.Background(),
	)

	fmt.Println("Error:", err.Error())
	fmt.Println("Code:", err.Code())
	fmt.Println("Has trace:", err.HasStackTrace())

	// Output:
	// Error: database query failed, connection timeout
	// Code: 1001
	// Has trace: true
}

func ExampleErrorImpl_tagManagement() {
	err := NewErrorImpl("validation failed", 400, nil, nil, nil, nil)

	// 设置标签
	err.SetTag("field", "email")
	err.SetTag("field", "username") // 同一key可以有多个值
	err.SetTag("severity", "high")

	// 获取所有标签
	tags := err.Tags()
	fmt.Printf("All tags: %+v\n", tags)

	// 获取特定标签
	fields := err.GetTag("field")
	fmt.Printf("Field errors: %v\n", fields)

	// Output:
	// All tags: map[field:[email username] severity:[high]]
	// Field errors: [email username]
}
