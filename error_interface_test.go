package irr

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrorImpl_ComprehensiveCoverage 测试所有未覆盖的方法和路径
func TestErrorImpl_ComprehensiveCoverage(t *testing.T) {
	t.Run("Frame String method", func(t *testing.T) {
		frame := Frame{
			File:     "/path/to/file.go",
			Function: "TestFunction",
			Line:     42,
			Package:  "github.com/test/package",
		}
		expected := "/path/to/file.go:42 in TestFunction"
		assert.Equal(t, expected, frame.String())
	})

	t.Run("Error with cause formatting", func(t *testing.T) {
		cause := fmt.Errorf("original error")
		err := Wraps(cause, "wrapped message").Build()

		// Error() method should include cause
		assert.Equal(t, "wrapped message, original error", err.Error())

		// String() method should also include cause
		assert.Equal(t, "wrapped message, original error", err.String())
	})

	t.Run("Format method with different verbs", func(t *testing.T) {
		err := New("test error").Code(500).Tag("type", "api").Build()

		// Test %v verb
		result := fmt.Sprintf("%v", err)
		assert.Equal(t, "test error", result)

		// Test %+v verb (should show details)
		detailResult := fmt.Sprintf("%+v", err)
		assert.Contains(t, detailResult, "test error")
		assert.Contains(t, detailResult, "code: 500")
		assert.Contains(t, detailResult, "tags: [type=api]")

		// Test %s verb
		stringResult := fmt.Sprintf("%s", err)
		assert.Equal(t, "test error", stringResult)

		// Test %q verb (quoted)
		quotedResult := fmt.Sprintf("%q", err)
		assert.Equal(t, "\"test error\"", quotedResult)
	})

	t.Run("Context methods", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "test_key", "test_value")
		err := New("test error").Build()

		// Test Context() with nil context
		defaultCtx := err.Context()
		assert.Equal(t, context.Background(), defaultCtx)

		// Test WithContext()
		enrichedErr := err.WithContext(ctx)
		assert.Equal(t, ctx, enrichedErr.Context())
		assert.Equal(t, "test_value", enrichedErr.Context().Value("test_key"))
	})

	t.Run("Tag methods comprehensive", func(t *testing.T) {
		err := New("test error").Build()

		// Test SetTag method (void)
		err.SetTag("key1", "value1")
		err.SetTag("key1", "value2") // Same key, different value
		err.SetTag("key2", "value3")

		// Test Tags() method - should return copies
		tags := err.Tags()
		assert.Len(t, tags["key1"], 2)
		assert.Contains(t, tags["key1"], "value1")
		assert.Contains(t, tags["key1"], "value2")
		assert.Equal(t, []string{"value3"}, tags["key2"])

		// Modify returned map should not affect original
		tags["key1"][0] = "modified"
		originalTags := err.Tags()
		assert.Equal(t, "value1", originalTags["key1"][0]) // Should not be modified

		// Test GetTag method
		key1Values := err.GetTag("key1")
		assert.Len(t, key1Values, 2)
		assert.Contains(t, key1Values, "value1")
		assert.Contains(t, key1Values, "value2")

		// Test GetTag with non-existent key
		nonExistent := err.GetTag("non_existent")
		assert.Nil(t, nonExistent)

		// Test Tag method (fluent)
		fluentErr := err.Tag("key3", "value4")
		assert.Equal(t, err, fluentErr) // Should return same instance
		assert.Equal(t, []string{"value4"}, err.GetTag("key3"))
	})

	t.Run("Tag methods with nil tags", func(t *testing.T) {
		// Create error with nil tags map
		err := newErrorImpl("test", 0, nil, nil, nil, nil)

		// Tags() on nil map should return empty map
		tags := err.Tags()
		assert.NotNil(t, tags)
		assert.Empty(t, tags)

		// GetTag on nil map should return nil
		values := err.GetTag("any_key")
		assert.Nil(t, values)

		// SetTag should initialize map
		err.SetTag("new_key", "new_value")
		assert.Equal(t, []string{"new_value"}, err.GetTag("new_key"))
	})

	t.Run("Unwrap method", func(t *testing.T) {
		cause := fmt.Errorf("root cause")
		err := Wraps(cause, "wrapper").Build()

		unwrapped := err.Unwrap()
		assert.Equal(t, cause, unwrapped)

		// Test Unwrap on error without cause
		simpleErr := Error("simple error")
		assert.Nil(t, simpleErr.Unwrap())
	})

	t.Run("Root method", func(t *testing.T) {
		// Create error chain: err3 -> err2 -> err1
		err1 := fmt.Errorf("root error")
		err2 := Wraps(err1, "middle error").Build()
		err3 := Wraps(err2, "top error").Build()

		root := err3.Root()
		assert.Equal(t, err1, root)

		// Test Root on simple error
		simpleErr := Error("simple")
		assert.Equal(t, simpleErr, simpleErr.Root())
	})

	t.Run("TraverseToRoot method", func(t *testing.T) {
		err1 := fmt.Errorf("root")
		err2 := Wraps(err1, "middle").Build()
		err3 := Wraps(err2, "top").Build()

		var visited []string
		result := err3.TraverseToRoot(func(err error) error {
			visited = append(visited, err.Error())
			return nil // Continue traversal
		})

		assert.Nil(t, result)
		assert.Len(t, visited, 3)
		assert.Equal(t, "top, middle, root", visited[0])
		assert.Contains(t, visited[1], "middle")
		assert.Equal(t, "root", visited[2])

		// Test early termination
		stopErr := fmt.Errorf("stop here")
		result = err3.TraverseToRoot(func(err error) error {
			if strings.Contains(err.Error(), "middle") {
				return stopErr
			}
			return nil
		})
		assert.Equal(t, stopErr, result)
	})

	t.Run("TraverseToSource method", func(t *testing.T) {
		err1 := fmt.Errorf("source")
		err2 := Wraps(err1, "wrapper").Build()

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
	})

	t.Run("Details method with all components", func(t *testing.T) {
		err := New("test error").
			Code(404).
			Tag("service", "api").
			Tag("method", "GET").
			Trace().
			Build()

		details := err.Details()
		assert.Contains(t, details, "test error")
		assert.Contains(t, details, "code: 404")
		assert.Contains(t, details, "tags: [")
		assert.Contains(t, details, "service=api")
		assert.Contains(t, details, "method=GET")
		assert.Contains(t, details, "stack:")
	})

	t.Run("ToString with trace", func(t *testing.T) {
		err := Trace("test error with trace")

		// ToString with trace enabled
		traceOutput := err.ToString(true, " | ")

		// For a single Trace() error, there should be one stack trace line
		// Format: message + shortFunc + frame details
		assert.Contains(t, traceOutput, "test error with trace")
		assert.Contains(t, traceOutput, "TestErrorImpl_ComprehensiveCoverage") // Function name should be included
		assert.Contains(t, traceOutput, "error_interface_test.go")             // File should be included

		// Test with Track() to verify multiple stack traces work
		wrappedErr := Track(err, "wrapped error")
		wrappedOutput := wrappedErr.ToString(true, " | ")
		wrappedParts := strings.Split(wrappedOutput, " | ")

		// Should have 2 parts: Track call + Trace call
		assert.Equal(t, 2, len(wrappedParts))
		assert.Contains(t, wrappedParts[0], "wrapped error")
		assert.Contains(t, wrappedParts[1], "test error with trace")
	})

	t.Run("ToString without trace", func(t *testing.T) {
		cause := fmt.Errorf("original")
		err := Wraps(cause, "wrapper").Build()

		// ToString without trace should use custom separator for cause
		output := err.ToString(false, " --> ")
		expected := "wrapper --> original"
		assert.Equal(t, expected, output)
	})

	t.Run("HasCode and HasStackTrace", func(t *testing.T) {
		// Test error without code
		errNoCode := Error("no code")
		assert.False(t, errNoCode.HasCode())
		assert.Equal(t, int64(0), errNoCode.Code())

		// Test error with code
		errWithCode := ErrorC(500, "with code")
		assert.True(t, errWithCode.HasCode())
		assert.Equal(t, int64(500), errWithCode.Code())

		// Test error without trace
		errNoTrace := Error("no trace")
		assert.False(t, errNoTrace.HasStackTrace())
		assert.Empty(t, errNoTrace.StackTrace())

		// Test error with trace
		errWithTrace := Trace("with trace")
		assert.True(t, errWithTrace.HasStackTrace())
		assert.NotEmpty(t, errWithTrace.StackTrace())
	})
}
