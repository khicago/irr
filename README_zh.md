# IRR - Go语言结构化错误处理库

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Build Status](https://travis-ci.org/khicago/irr.svg?branch=master)](https://travis-ci.org/khicago/irr)
[![codecov](https://codecov.io/gh/khicago/irr/branch/master/graph/badge.svg)](https://codecov.io/gh/khicago/irr)
[![Go Report Card](https://goreportcard.com/badge/github.com/khicago/irr)](https://goreportcard.com/report/github.com/khicago/irr)
[![GoDoc](https://godoc.org/github.com/khicago/irr?status.svg)](https://godoc.org/github.com/khicago/irr)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

IRR是一个Go错误处理库，在错误传播链中保留上下文信息。它解决了错误消息在应用程序层级间传播时丢失关键调试信息的常见问题。

**[English Version](README.md)**

## 问题描述

传统的Go错误处理经常导致信息丢失：

```go
// 传统方式 - 每一层都丢失上下文
func processUser(id string) error {
    user, err := getUserFromDB(id)
    if err != nil {
        return fmt.Errorf("failed to get user: %w", err)
    }
    
    if err := validateUser(user); err != nil {
        return fmt.Errorf("validation failed: %w", err) 
    }
    
    return nil
}
// 结果: "validation failed: failed to get user: connection timeout"
// 缺失: 何时、何处、什么上下文、错误分类
```

## 解决方案

IRR提供带有上下文保留的结构化错误处理：

```go
// IRR方式 - 保留结构化上下文
func processUser(ctx context.Context, id string) error {
    user, err := getUserFromDB(id)
    if err != nil {
        return irr.Wraps(err, "failed to get user for id=%s", id).
            Tag("operation", "user_fetch").
            Tag("user_id", id).
            WithEnricher(&enrichers.TimeoutEnricher{}).
            BuildWithContext(ctx)
    }
    
    if err := validateUser(user); err != nil {
        return irr.Wraps(err, "user validation failed").
            Tag("operation", "validation").
            Tag("user_id", id).
            BuildWithContext(ctx)
    }
    
    return nil
}
// 结果: 包含标签、上下文和调试信息的结构化错误
```

## 主要特性

### Builder模式API
用于构造带有可选属性错误的流式接口：
```go
err := irr.New("operation failed").
    Code(1001).
    Tag("service", "user").
    Trace().
    Build()
```

### 上下文增强
可插拔的增强器从Go上下文中提取信息：
```go
import "github.com/khicago/irr/enrichers"

err := irr.New("timeout occurred").
    WithEnricher(enrichers.NewRequestInfoEnricher()). // 提取trace_id, user_id
    WithEnricher(&enrichers.TimeoutEnricher{}).       // 检测超时
    BuildWithContext(ctx)
```

### 企业错误代码
使用IRC（IRR代码）系统的结构化错误分类：
```go
const (
    ErrSystemDatabase    irc.Code = 1001  // 系统错误 (1000-1999)
    ErrBusinessValidation irc.Code = 2001  // 业务错误 (2000-2999)
    ErrAPINotFound       irc.Code = 3002  // API错误 (3000-3999)
)

return ErrBusinessValidation.Error("email is required")
```

### 可选监控集成
在应用程序边界进行非侵入式指标收集：
```go
// 在处理层
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    err := h.businessLogic(r.Context())
    if err != nil {
        h.analyzer.Analyze(err)  // 可选的指标收集
        h.logger.Error(err.Details())
        http.Error(w, "Internal server error", 500)
    }
}
```

## 安装

```bash
go get github.com/khicago/irr
```

## 基础用法

### 创建错误

```go
// 简单错误
err := irr.New("user not found").Build()

// 带元数据的错误
err := irr.New("invalid input: %s", input).
    Code(400).
    Tag("field", "email").
    Tag("service", "user").
    Build()

// 带堆栈跟踪的错误
err := irr.New("critical failure").
    Trace().
    Build()
```

### 包装错误

```go
user, err := db.GetUser(id)
if err != nil {
    return irr.Wraps(err, "failed to fetch user").
        Tag("user_id", id).
        Tag("operation", "get_user").
        BuildWithContext(ctx)
}
```

### 上下文感知错误

```go
// 自动超时检测
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

err := irr.New("operation failed").
    WithEnricher(&enrichers.TimeoutEnricher{}).
    BuildWithContext(ctx)

// 如果上下文超时：
// err.Tags()["timeout"] = "true"
// err.Tags()["deadline_exceeded"] = "2024-01-01T10:05:00Z"
```

### 企业错误代码

```go
import "github.com/khicago/irr/codes"

// 定义错误分类
const (
    ErrUserNotFound     codes.Code = 2001
    ErrInvalidInput     codes.Code = 2002
    ErrDatabaseTimeout  codes.Code = 1001
)

// 创建分类错误
func GetUser(id string) (*User, error) {
    user, err := db.Query(id)
    if err != nil {
        return nil, ErrDatabaseTimeout.Track(err, "query failed for user=%s", id)
    }
    
    if user == nil {
        return nil, ErrUserNotFound.Error("user id=%s not found", id)
    }
    
    return user, nil
}
```

### Result类型（可选）

受Rust启发的函数式编程风格错误处理：

```go
import "github.com/khicago/irr/result"

func safeDivision(a, b int) result.Result[int] {
    if b == 0 {
        return result.Err[int](irr.New("division by zero").Build())
    }
    return result.Ok(a / b)
}

// 使用
res := safeDivision(10, 2)
if res.IsOk() {
    fmt.Printf("Result: %d", res.Unwrap())
} else {
    fmt.Printf("Error: %v", res.UnwrapErr())
}
```

## 高级用法

### 自定义增强器

```go
type DatabaseEnricher struct{}

func (d *DatabaseEnricher) Enrich(ctx context.Context, err irr.IRR) irr.IRR {
    if dbName := ctx.Value("db-name"); dbName != nil {
        err.SetTag("database", dbName.(string))
    }
    return err
}

// 使用
err := irr.New("query failed").
    WithEnricher(&DatabaseEnricher{}).
    BuildWithContext(ctx)
```

### 处理层错误处理

```go
type ErrorHandler struct {
    logger   *zap.Logger
    analyzer *metrics.ErrorAnalyzer
}

func (h *ErrorHandler) HandleError(err error) {
    if irrErr, ok := err.(irr.IRR); ok {
        // 提取结构化信息
        code := irrErr.Code()
        tags := irrErr.Tags()
        
        // 结构化日志
        h.logger.Error("operation failed",
            zap.Error(err),
            zap.Int64("code", code),
            zap.Any("tags", tags),
        )
        
        // 可选指标（非阻塞）
        go h.analyzer.Analyze(irrErr)
    }
}
```

### 跨服务错误传播

```go
// 服务A
err := irr.New("service unavailable").
    Tag("service", "user-service").
    Tag("trace_id", ctx.Value("trace_id")).
    Code(5001).
    BuildWithContext(ctx)

// 序列化传输
errorJSON := err.MarshalJSON()

// 服务B - 重建错误上下文
reconstructedErr := irr.UnmarshalFromJSON(errorJSON)

// 链接新上下文
chainedErr := irr.Wraps(reconstructedErr, "downstream service failed").
    Tag("service", "order-service").
    BuildWithContext(ctx)
```

## 性能

IRR为生产使用而设计，开销最小：

```
BenchmarkError-8           2000000    750 ns/op    112 B/op    2 allocs/op
BenchmarkWrap-8            1500000    950 ns/op    144 B/op    3 allocs/op  
BenchmarkTrace-8           1000000   1200 ns/op    256 B/op    4 allocs/op
BenchmarkTrack-8            800000   1450 ns/op    288 B/op    5 allocs/op

// 与标准库比较：
BenchmarkStdError-8        3000000    420 ns/op     64 B/op    1 allocs/op
BenchmarkStdWrap-8         2000000    680 ns/op     96 B/op    2 allocs/op
```

**性能特性：**
- 跟踪对象内存池化
- 带缓存的延迟字符串构建
- 简单情况的零分配快速路径
- 指标的高效原子操作

## 质量保证

**测试覆盖率：**
```
Package                              Coverage
github.com/khicago/irr               88.7%
github.com/khicago/irr/internal/errors  93.3%
github.com/khicago/irr/enrichers     100.0%
github.com/khicago/irr/reporter      86.9%
github.com/khicago/irr/codes         71.4%
github.com/khicago/irr/metrics       70.7%
```

**质量标准：**
- 语义化测试命名以提高可维护性
- 边界情况和并发访问测试
- 性能回归检测的基准测试
- 完整的API契约验证

## 与其他库的比较

| 特性 | IRR | pkg/errors | std errors | zerolog |
|------|-----|------------|------------|---------|
| 结构化上下文 | 是 | 否 | 否 | 有限 |
| 错误代码 | 是(codes) | 否 | 否 | 否 |
| 上下文集成 | 原生 | 否 | 否 | 否 |
| 监控就绪 | 是 | 否 | 否 | 部分 |
| Result类型 | 可选 | 否 | 否 | 否 |
| 性能 | 优化 | 良好 | 优秀 | 良好 |

## 文档

- [API文档](https://godoc.org/github.com/khicago/irr)
- [架构设计](./design.0.2.md) - 深入设计分析和逐层分解
- [IRC企业指南](./docs/irc-enterprise-practices.md) - 生产模式和最佳实践
- [示例](./examples/) - 实际使用示例

## 迁移指南

### 从标准库迁移

```go
// 之前
return fmt.Errorf("operation failed: %w", err)

// 之后
return irr.Wraps(err, "operation failed").
    Tag("operation", "user_fetch").
    BuildWithContext(ctx)
```

### 从pkg/errors迁移

```go
// 之前
return errors.Wrap(err, "operation failed")

// 之后  
return irr.Wraps(err, "operation failed").
    Trace().
    BuildWithContext(ctx)
```

## 贡献

1. Fork仓库
2. 创建特性分支：`git checkout -b feature/amazing-feature`
3. 运行测试：`go test -v ./...`
4. 运行基准测试：`go test -bench=. -benchmem`
5. 提交pull request

**开发环境设置：**
```bash
git clone https://github.com/khicago/irr.git
cd irr
go mod tidy
go test -v ./...
```

## 许可证

本项目基于MIT许可证 - 查看[LICENSE](LICENSE)文件获取详情。

## 致谢

- 受Rust结构化错误处理模式启发
- 为Go社区对更好错误上下文保留的需求而构建
- 感谢所有贡献者和早期采用者

---

**被全球公司和项目用于生产错误处理。**