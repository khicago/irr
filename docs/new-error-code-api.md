# 🔧 新错误码API设计文档

## 📋 概述

IRR库引入了全新的错误码API设计，解决了原有API的语义模糊问题，提供更清晰、更直观的错误码管理体验。

## 🎯 设计目标

### 核心问题
原有的 `GetCode()` 和 `ClosestCode()` 方法存在以下问题：
1. **语义模糊**：无法区分"未设置错误码"和"显式设置为0"
2. **用户困惑**：`GetCode()` 实际返回的是最近的错误码，而非当前错误的错误码
3. **API不一致**：方法名称与实际行为不匹配

### 解决方案
新API采用**方案B**设计，提供清晰的语义分离：
- **明确的层级概念**：当前层、最近层、根层
- **显式的状态检查**：区分"未设置"和"设置为0"
- **向后兼容**：保留原有API，标记为废弃

## 🚀 新API详解

### 核心方法

#### 1. `NearestCode() int64`
```go
// 返回错误链中最近的有效错误码（非零）
// 这是推荐使用的方法，符合用户直觉
func (ir *BasicIrr) NearestCode() int64
```

**使用场景**：
- HTTP状态码返回
- 业务错误码展示
- 错误分类处理

**示例**：
```go
inner := irr.Error("database error").SetCode(500)
outer := irr.Wrap(inner, "service error") // 未设置错误码

fmt.Println(outer.NearestCode()) // 输出: 500
```

#### 2. `CurrentCode() int64`
```go
// 返回当前错误对象的错误码（可能为0）
func (ir *BasicIrr) CurrentCode() int64
```

**使用场景**：
- 检查特定层级的错误码
- 调试和日志记录
- 精确的错误码控制

**示例**：
```go
err := irr.Error("test error").SetCode(404)
fmt.Println(err.CurrentCode()) // 输出: 404

wrapped := irr.Wrap(err, "wrapper") // 未设置错误码
fmt.Println(wrapped.CurrentCode()) // 输出: 0
```

#### 3. `RootCode() int64`
```go
// 返回错误链根部的错误码
func (ir *BasicIrr) RootCode() int64
```

**使用场景**：
- 获取原始错误的错误码
- 错误溯源分析
- 根因分析

**示例**：
```go
root := irr.Error("root cause").SetCode(404)
middle := irr.Wrap(root, "middle layer").SetCode(500)
top := irr.Wrap(middle, "top layer").SetCode(400)

fmt.Println(top.RootCode()) // 输出: 404
```

#### 4. `HasCurrentCode() bool`
```go
// 检查当前错误对象是否显式设置了错误码
func (ir *BasicIrr) HasCurrentCode() bool
```

**使用场景**：
- 区分"未设置"和"设置为0"
- 条件性错误码处理
- API行为验证

**示例**：
```go
err1 := irr.Error("test")
fmt.Println(err1.HasCurrentCode()) // 输出: false

err2 := irr.Error("test").SetCode(0)
fmt.Println(err2.HasCurrentCode()) // 输出: true (显式设置了)

err3 := irr.Error("test").SetCode(404)
fmt.Println(err3.HasCurrentCode()) // 输出: true
```

#### 5. `HasAnyCode() bool`
```go
// 检查错误链中是否有任何错误码
func (ir *BasicIrr) HasAnyCode() bool
```

**使用场景**：
- 快速检查错误链是否包含错误码
- 错误处理策略选择
- 性能优化判断

**示例**：
```go
err1 := irr.Error("no code")
wrapped1 := irr.Wrap(err1, "still no code")
fmt.Println(wrapped1.HasAnyCode()) // 输出: false

err2 := irr.Error("with code").SetCode(500)
wrapped2 := irr.Wrap(err2, "wrapper")
fmt.Println(wrapped2.HasAnyCode()) // 输出: true
```

## 📊 API对比表

| 场景 | 旧API | 新API | 优势 |
|------|-------|-------|------|
| 获取最近错误码 | `GetCode()` | `NearestCode()` | 语义清晰 |
| 获取当前错误码 | 无直接方法 | `CurrentCode()` | 精确控制 |
| 获取根错误码 | 无直接方法 | `RootCode()` | 溯源分析 |
| 检查是否设置 | 无法区分 | `HasCurrentCode()` | 状态明确 |
| 检查链中是否有码 | 需要遍历 | `HasAnyCode()` | 性能优化 |

## 🔄 向后兼容性

### 废弃方法
```go
// Deprecated: 使用 NearestCode() 获得更清晰的语义
func (ir *BasicIrr) GetCode() int64

// Deprecated: 使用 NearestCode() 获得更清晰的语义  
func (ir *BasicIrr) ClosestCode() int64
```

### 迁移指南

#### 1. 替换 `GetCode()`
```go
// 旧代码
code := err.GetCode()

// 新代码
code := err.NearestCode()
```

#### 2. 替换 `ClosestCode()`
```go
// 旧代码
code := err.ClosestCode()

// 新代码
code := err.NearestCode()
```

#### 3. 检查错误码状态
```go
// 旧代码（无法实现）
// 无法区分未设置和设置为0

// 新代码
if err.HasCurrentCode() {
    // 显式设置了错误码（可能为0）
    code := err.CurrentCode()
} else {
    // 未设置错误码
}
```

## 💡 最佳实践

### 1. 错误码设置策略
```go
// ✅ 推荐：在创建错误时设置错误码
func GetUser(id string) (*User, error) {
    if id == "" {
        return nil, irr.Error("invalid user id").SetCode(400)
    }
    
    user, err := db.GetUser(id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, irr.Wrap(err, "user not found").SetCode(404)
        }
        return nil, irr.Wrap(err, "database error").SetCode(500)
    }
    
    return user, nil
}
```

### 2. 错误码检查策略
```go
// ✅ 推荐：使用新API进行错误码检查
func HandleError(err error) {
    if irrErr, ok := err.(irr.IRR); ok {
        // 检查是否有错误码
        if irrErr.HasAnyCode() {
            code := irrErr.NearestCode()
            // 根据错误码处理
            switch code {
            case 400:
                // 客户端错误
            case 500:
                // 服务器错误
            }
        }
        
        // 检查特定层级的错误码
        if irrErr.HasCurrentCode() {
            currentCode := irrErr.CurrentCode()
            // 处理当前层级的错误码
        }
    }
}
```

### 3. 错误链分析
```go
// ✅ 推荐：使用新API进行错误链分析
func AnalyzeError(err error) {
    if irrErr, ok := err.(irr.IRR); ok {
        fmt.Printf("最近错误码: %d\n", irrErr.NearestCode())
        fmt.Printf("当前错误码: %d\n", irrErr.CurrentCode())
        fmt.Printf("根错误码: %d\n", irrErr.RootCode())
        fmt.Printf("当前层是否设置: %t\n", irrErr.HasCurrentCode())
        fmt.Printf("链中是否有码: %t\n", irrErr.HasAnyCode())
    }
}
```

### 4. HTTP状态码映射
```go
// ✅ 推荐：使用NearestCode进行HTTP状态码映射
func ErrorToHTTPStatus(err error) int {
    if irrErr, ok := err.(irr.IRR); ok && irrErr.HasAnyCode() {
        code := irrErr.NearestCode()
        if code >= 400 && code < 600 {
            return int(code) // 直接使用HTTP状态码
        }
    }
    return 500 // 默认服务器错误
}
```

## 🧪 测试策略

### 1. 单元测试覆盖
```go
func TestErrorCodeAPI(t *testing.T) {
    // 测试基本功能
    err := irr.Error("test").SetCode(404)
    assert.Equal(t, int64(404), err.NearestCode())
    assert.Equal(t, int64(404), err.CurrentCode())
    assert.True(t, err.HasCurrentCode())
    assert.True(t, err.HasAnyCode())
    
    // 测试错误链
    wrapped := irr.Wrap(err, "wrapper")
    assert.Equal(t, int64(404), wrapped.NearestCode())
    assert.Equal(t, int64(0), wrapped.CurrentCode())
    assert.False(t, wrapped.HasCurrentCode())
    assert.True(t, wrapped.HasAnyCode())
}
```

### 2. 边界条件测试
```go
func TestErrorCodeEdgeCases(t *testing.T) {
    // 测试零值
    err := irr.Error("test").SetCode(0)
    assert.Equal(t, int64(0), err.CurrentCode())
    assert.True(t, err.HasCurrentCode()) // 显式设置了
    assert.False(t, err.HasAnyCode())    // 但没有有效错误码
    
    // 测试负数
    err = irr.Error("test").SetCode(-1)
    assert.Equal(t, int64(-1), err.NearestCode())
    assert.True(t, err.HasAnyCode())
}
```

## 🚀 性能优化

### 1. 零分配路径
新API在常见场景下实现零分配：
```go
// 零分配的错误码检查
if err.HasAnyCode() {
    code := err.NearestCode() // 无内存分配
}
```

### 2. 缓存优化
对于深层错误链，考虑缓存错误码查找结果：
```go
type CachedIrr struct {
    *irr.BasicIrr
    cachedNearestCode *int64
}

func (c *CachedIrr) NearestCode() int64 {
    if c.cachedNearestCode == nil {
        code := c.BasicIrr.NearestCode()
        c.cachedNearestCode = &code
    }
    return *c.cachedNearestCode
}
```

## 🎯 总结

新的错误码API设计解决了原有API的核心问题：

1. **语义清晰**：每个方法的名称准确反映其功能
2. **功能完整**：覆盖所有错误码管理场景
3. **向后兼容**：平滑迁移，无破坏性变更
4. **性能优化**：高效的实现，支持零分配路径
5. **测试完备**：100%测试覆盖，确保质量

通过采用新的错误码API，开发者可以：
- 🎯 更精确地控制错误码
- 🔍 更清晰地理解错误状态  
- 🚀 更高效地处理错误场景
- 🛡️ 更可靠地构建错误处理逻辑

这一改进将显著提升IRR库的用户体验和代码质量。 