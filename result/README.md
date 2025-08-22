# Result Package - 类型安全的错误处理

IRR Result包实现了标准的`Result<T,E>`模式，提供类型安全的函数式错误处理。遵循Rust、Scala、Haskell等语言的最佳实践。

## 🎯 设计原理

基于第一性原理和伟大项目的技术品位准则：
- **类型安全**: 编译期保证错误类型一致性
- **函数式**: 支持Map、AndThen、OrElse等操作
- **零开销**: 利用Go 1.23+泛型别名，真正的零运行时开销
- **语法糖**: `ResultIRR[T]` = `Result[T, IRR]`，完全透明的类型别名

## 📋 环境要求

**Go 1.23+** - 利用泛型别名特性实现零开销抽象

## 📝 核心类型

```go
// 标准Result<T,E>模式
type Result[T, E any] struct {
    value T
    err   E
    isOk  bool
}

// 🎯 Go 1.23+ 泛型别名 - 零运行时开销的语法糖
type ResultIRR[T any] = Result[T, irr.IRR]    // IRR错误类型
type ResultError[T any] = Result[T, error]    // 标准error接口
```

## 🚀 基本用法

### 1. 创建Result

```go
// 成功值
user := result.OkIRR(User{ID: 1, Name: "Alice"})
data := result.OkError([]byte("hello"))

// 错误值  
err := result.ErrIRR(irr.ErrorC(404, "user not found"))
err := result.ErrError(fmt.Errorf("network timeout"))
```

### 2. 检查和提取值

```go
func processUser(userResult result.ResultIRR[User]) {
    if userResult.IsOk() {
        user := userResult.Unwrap()
        fmt.Printf("User: %+v\n", user)
    } else {
        err := userResult.UnwrapErr()
        fmt.Printf("Error: %v\n", err)
    }
}

// 安全提取（推荐）
func safeProcess(userResult result.ResultIRR[User]) {
    if user, ok := userResult.Value(); ok {
        fmt.Printf("User: %+v\n", user)
    } else if err, ok := userResult.Error(); ok {
        fmt.Printf("Error: %v\n", err)
    }
}
```

### 3. 模式匹配（函数式风格）

```go
func handleUser(userResult result.ResultIRR[User]) {
    userResult.Match(
        func(user User) {
            fmt.Printf("Success: %+v\n", user)
        },
        func(err irr.IRR) {
            fmt.Printf("Error: %v\n", err)
        },
    )
}
```

## 🔗 函数式操作

### 1. Map操作 - 值转换

```go
// 将User转换为UserDTO
userDTOResult := result.Map(userResult, func(user User) UserDTO {
    return UserDTO{
        ID:   user.ID,
        Name: strings.ToUpper(user.Name),
    }
})
```

### 2. AndThen操作 - 链式处理

```go
func getUserByID(id int) result.ResultIRR[User] {
    if id <= 0 {
        return result.ErrIRR(irr.ErrorC(400, "invalid user id"))
    }
    return result.OkIRR(User{ID: id, Name: "Alice"})
}

func getProfile(user User) result.ResultIRR[Profile] {
    return result.OkIRR(Profile{UserID: user.ID, Bio: "Developer"})
}

// 链式操作，自动错误短路
profileResult := result.AndThen(getUserByID(123), getProfile)
```

### 3. OrElse操作 - 错误恢复

```go
// 从缓存获取用户，失败时从数据库获取
userResult := result.OrElse(getUserFromCache(123), func(cacheErr irr.IRR) result.ResultIRR[User] {
    return getUserFromDB(123)
})
```

## 🔄 与标准库集成

```go
// 包装标准库函数
func readFile(filename string) result.ResultError[[]byte] {
    return result.FromTuple(os.ReadFile(filename))
}

// 转换为标准库格式
data, err := result.ToTuple(readFile("config.json"))
if err != nil {
    log.Printf("Failed to read config: %v", err)
}
```

## 💡 实际应用示例

### 1. HTTP API处理

```go
func handleGetUser(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    
    result := result.AndThen(
        parseUserID(userID),
        func(id int) result.ResultIRR[User] {
            return getUserFromDB(id)
        },
    )
    
    result.Match(
        func(user User) {
            json.NewEncoder(w).Encode(user)
        },
        func(err irr.IRR) {
            if coder, ok := err.(interface{ Code() int64 }); ok {
                w.WriteHeader(int(coder.Code()))
            } else {
                w.WriteHeader(500)
            }
            json.NewEncoder(w).Encode(map[string]string{
                "error": err.Error(),
            })
        },
    )
}
```

### 2. 数据管道处理

```go
func processDataPipeline(input string) result.ResultIRR[ProcessedData] {
    return result.AndThen(
        validateInput(input),
        func(validInput string) result.ResultIRR[ParsedData] {
            return parseData(validInput)
        },
    ).AndThen(func(parsed ParsedData) result.ResultIRR[ProcessedData] {
        return transformData(parsed)
    })
}
```

### 3. 批量操作

```go
func processUsers(userIDs []int) result.ResultIRR[[]User] {
    results := make([]result.ResultIRR[User], len(userIDs))
    for i, id := range userIDs {
        results[i] = getUserByID(id)
    }
    
    // 收集所有成功结果，任一失败则整体失败
    return result.Collect(results...)
}
```

## 🏆 最佳实践

### 1. 优先使用语法糖
```go
// ✅ 推荐：使用语法糖
func getUser(id int) result.ResultIRR[User]

// ❌ 避免：冗长的完整类型
func getUser(id int) result.Result[User, irr.IRR]
```

### 2. 合理使用panic方法
```go
// ✅ 测试和调试场景
user := userResult.Expect("test user should exist")

// ❌ 生产环境避免
user := userResult.Unwrap() // 可能panic
```

### 3. 函数式风格优于命令式
```go
// ✅ 函数式风格
result := result.Map(getUserByID(123), func(user User) string {
    return user.Name
})

// ❌ 命令式风格
userResult := getUserByID(123)
if userResult.IsOk() {
    user := userResult.Unwrap()
    result = result.OkIRR(user.Name)
} else {
    result = result.ErrIRR(userResult.UnwrapErr())
}
```

## 🎯 与IRR错误系统集成

Result包与IRR错误系统深度集成，提供：
- **类型安全的错误传播**
- **结构化错误信息**
- **上下文感知的错误处理**
- **企业级错误码管理**

这使得Result不仅仅是错误处理工具，更是构建可靠系统的基础设施。