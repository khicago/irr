package errors

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBasicErrorFunctionality tests core error functionality
func TestBasicErrorFunctionality(t *testing.T) {
	ctx := context.Background()

	t.Run("Basic error creation", func(t *testing.T) {
		err := NewErrorImpl("test message", 404, nil, nil, nil, ctx)

		assert.Equal(t, "test message", err.Error())
		assert.Equal(t, "test message", err.String())
		assert.Equal(t, int64(404), err.Code())
		assert.True(t, err.HasCode())
		assert.False(t, err.HasStackTrace())
		assert.Empty(t, err.StackTrace())
		assert.Empty(t, err.Tags())
		assert.Equal(t, ctx, err.Context())
		assert.Nil(t, err.Unwrap())
	})

	t.Run("Error with cause", func(t *testing.T) {
		cause := fmt.Errorf("root cause")
		err := NewErrorImpl("wrapper", 500, cause, nil, nil, ctx)

		assert.Equal(t, cause, err.Unwrap())
		assert.Contains(t, err.Error(), "wrapper")
		assert.Contains(t, err.Error(), "root cause")
	})

	t.Run("Error with stack trace", func(t *testing.T) {
		trace := []Frame{{Function: "test.func", File: "test.go", Line: 42}}
		err := NewErrorImpl("traced error", 0, nil, nil, trace, ctx)

		assert.True(t, err.HasStackTrace())
		assert.Len(t, err.StackTrace(), 1)
		assert.Equal(t, trace[0], err.StackTrace()[0])
	})

	t.Run("Error with tags", func(t *testing.T) {
		tags := map[string][]string{"key": {"value1", "value2"}}
		err := NewErrorImpl("tagged", 0, nil, tags, nil, ctx)

		allTags := err.Tags()
		assert.Contains(t, allTags["key"], "value1")
		assert.Contains(t, allTags["key"], "value2")
	})
}

// TestTagOperations tests tag management
func TestTagOperations(t *testing.T) {
	err := NewErrorImpl("test", 0, nil, nil, nil, context.Background())

	t.Run("SetTag and GetTag", func(t *testing.T) {
		err.SetTag("module", "auth")
		assert.Equal(t, []string{"auth"}, err.GetTag("module"))

		err.SetTag("module", "user") // Add another value
		values := err.GetTag("module")
		assert.Len(t, values, 2)
		assert.Contains(t, values, "auth")
		assert.Contains(t, values, "user")

		assert.Empty(t, err.GetTag("nonexistent"))
	})

	t.Run("Tag fluent API", func(t *testing.T) {
		result := err.Tag("action", "login")
		assert.Equal(t, err, result) // Returns self for chaining
		assert.Contains(t, err.GetTag("action"), "login")
	})
}

// TestContextOperations tests context management
func TestContextOperations(t *testing.T) {
	originalCtx := context.WithValue(context.Background(), "user", "alice")
	err := NewErrorImpl("test", 0, nil, nil, nil, originalCtx)

	t.Run("Context retrieval", func(t *testing.T) {
		ctx := err.Context()
		assert.Equal(t, "alice", ctx.Value("user"))
	})

	t.Run("Context replacement", func(t *testing.T) {
		newCtx := context.WithValue(context.Background(), "user", "bob")
		result := err.WithContext(newCtx)

		assert.Equal(t, err, result) // Returns self
		assert.Equal(t, "bob", err.Context().Value("user"))
	})
}

// TestCodeOperations tests error code management
func TestCodeOperations(t *testing.T) {
	err := NewErrorImpl("test", 0, nil, nil, nil, context.Background())

	t.Run("Initial state", func(t *testing.T) {
		assert.False(t, err.HasCode())
		assert.Equal(t, int64(0), err.Code())
	})

	t.Run("Set code", func(t *testing.T) {
		result := err.SetCode(404)
		assert.Equal(t, err, result) // Returns self
		assert.True(t, err.HasCode())
		assert.Equal(t, int64(404), err.Code())
	})
}

// TestFrameString tests Frame.String method
func TestFrameString(t *testing.T) {
	frame := Frame{
		Function: "main.authenticate",
		File:     "/app/auth.go",
		Line:     42,
	}

	result := frame.String()
	assert.Contains(t, result, "main.authenticate")
	assert.Contains(t, result, "/app/auth.go:42")
}

// TestFormatMethods tests formatting methods
func TestFormatMethods(t *testing.T) {
	err := NewErrorImpl("test error", 404, nil, nil, nil, context.Background())

	t.Run("Basic string formats", func(t *testing.T) {
		s := fmt.Sprintf("%s", err)
		v := fmt.Sprintf("%v", err)
		assert.Equal(t, s, v)
		assert.Equal(t, "test error", s)

		q := fmt.Sprintf("%q", err)
		assert.Equal(t, `"test error"`, q)
	})

	t.Run("ToString without trace", func(t *testing.T) {
		result := err.ToString(false, "\n")
		assert.Equal(t, "test error", result)
	})

	t.Run("Details format", func(t *testing.T) {
		details := err.Details()
		assert.Contains(t, details, "test error")
		assert.Contains(t, details, "code: 404")
	})
}

// TestTraversalMethods tests error traversal functionality
func TestTraversalMethods(t *testing.T) {
	root := fmt.Errorf("root error")
	middle := NewErrorImpl("middle", 0, root, nil, nil, context.Background())
	top := NewErrorImpl("top", 0, middle, nil, nil, context.Background())

	t.Run("Root finding", func(t *testing.T) {
		foundRoot := top.Root()
		assert.Equal(t, root, foundRoot)
	})

	t.Run("TraverseToRoot", func(t *testing.T) {
		var collected []string
		err := top.TraverseToRoot(func(err error) error {
			collected = append(collected, err.Error())
			return nil
		})

		assert.NoError(t, err)
		assert.Len(t, collected, 3)
		assert.Contains(t, collected[0], "top")
		assert.Contains(t, collected[1], "middle")
		assert.Equal(t, "root error", collected[2])
	})

	t.Run("TraverseToSource", func(t *testing.T) {
		var sources []error
		var wrappers []error

		top.TraverseToSource(func(err error, isSource bool) error {
			if isSource {
				sources = append(sources, err)
			} else {
				wrappers = append(wrappers, err)
			}
			return nil
		})

		assert.Len(t, sources, 1)  // root error
		assert.Len(t, wrappers, 2) // top and middle
	})
}

// TestTypeSafetyHelpers tests type assertion helpers through public interface
func TestTypeSafetyHelpers(t *testing.T) {
	t.Run("Error with stack trace", func(t *testing.T) {
		trace := []Frame{{Function: "test", File: "test.go", Line: 1}}
		err := NewErrorImpl("test", 0, nil, nil, trace, context.Background())

		provider, hasTrace := safeGetStackTraceProvider(err)
		assert.True(t, hasTrace)
		assert.NotNil(t, provider)
		assert.True(t, provider.HasStackTrace())
		assert.Len(t, provider.StackTrace(), 1)
	})

	t.Run("Standard error", func(t *testing.T) {
		err := fmt.Errorf("standard")
		provider, hasTrace := safeGetStackTraceProvider(err)
		assert.False(t, hasTrace)
		assert.Nil(t, provider)
	})
}

// ExampleNewErrorImpl demonstrates basic usage
func ExampleNewErrorImpl() {
	err := NewErrorImpl("operation failed", 500, nil, nil, nil, context.Background())

	fmt.Printf("Error: %s\n", err.Error())
	fmt.Printf("Code: %d\n", err.Code())
	fmt.Printf("Has code: %t\n", err.HasCode())

	// Output:
	// Error: operation failed
	// Code: 500
	// Has code: true
}

// ExampleErrorImpl_Tags demonstrates tag usage
func ExampleErrorImpl_Tags() {
	err := NewErrorImpl("auth failed", 401, nil, nil, nil, context.Background())

	err.SetTag("module", "authentication")
	err.SetTag("user", "alice")

	fmt.Printf("Module: %v\n", err.GetTag("module"))
	fmt.Printf("User: %v\n", err.GetTag("user"))

	// Output:
	// Module: [authentication]
	// User: [alice]
}
