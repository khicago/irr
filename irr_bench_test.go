package irr

import (
	"errors"
	"fmt"
	"testing"
)

func BenchmarkError(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Error("test error %d", i)
	}
}

func BenchmarkErrorC(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ErrorC(1001, "test error %d", i)
	}
}

func BenchmarkWrap(b *testing.B) {
	baseErr := errors.New("base error")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Wrap(baseErr, "wrap error %d", i)
	}
}

func BenchmarkTrace(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Trace("test error %d", i)
	}
}

func BenchmarkTrack(b *testing.B) {
	baseErr := errors.New("base error")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Track(baseErr, "track error %d", i)
	}
}

func BenchmarkToString(b *testing.B) {
	err := Track(Error("inner error"), "outer error")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = err.ToString(false, ", ")
	}
}

func BenchmarkToStringWithTrace(b *testing.B) {
	err := Track(Trace("inner error"), "outer error")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = err.ToString(true, ", ")
	}
}

func BenchmarkTraverseToSource(b *testing.B) {
	// 创建深层嵌套的错误链
	err := Error("base error")
	for i := 0; i < 10; i++ {
		err = Wrap(err, "level %d", i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = err.TraverseToSource(func(e error, isSource bool) error {
			return nil
		})
	}
}

func BenchmarkTraverseToRoot(b *testing.B) {
	// 创建深层嵌套的错误链
	err := Error("base error")
	for i := 0; i < 10; i++ {
		err = Wrap(err, "level %d", i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = err.TraverseToRoot(func(e error) error {
			return nil
		})
	}
}

func BenchmarkErrorsIs(b *testing.B) {
	baseErr := Error("base error")
	wrappedErr := Wrap(baseErr, "wrapped")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = errors.Is(wrappedErr, baseErr)
	}
}

func BenchmarkSetGetCode(b *testing.B) {
	err := Error("test error")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err.SetCode(int64(i))
		_ = err.GetCode()
	}
}

func BenchmarkSetGetTag(b *testing.B) {
	err := Error("test error")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err.SetTag("key", fmt.Sprintf("value%d", i))
		_ = err.GetTag("key")
	}
}

// 对比标准库性能
func BenchmarkStdError(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fmt.Errorf("test error %d", i)
	}
}

func BenchmarkStdWrap(b *testing.B) {
	baseErr := errors.New("base error")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fmt.Errorf("wrap error %d: %w", i, baseErr)
	}
}
