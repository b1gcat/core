# License 包

该包提供了一个简单的软件授权机制，可以根据编译时间和授权天数来限制软件的使用期限。

## 功能特点

- 基于编译时间和授权天数的软件授权
- 启动时自动检查授权是否有效
- 超过授权期限时自动退出程序
- 支持通过构建参数注入编译时间

## 使用方法

### 1. 导入包

```go
import (
    "github.com/b1gcat/core/license"
)
```

### 2. 设置授权天数

在程序启动时设置授权的天数：

```go
func main() {
    // 设置授权天数为30天
    license.SetExpiration(30)
    
    // 检查授权是否有效
    // 如果超过授权期限，程序会自动退出
    license.CheckLicense()
    
    // 正常的程序逻辑...
}
```

### 3. 编译程序

编译时需要注入编译时间戳：

```bash
go build -ldflags "-X github.com/b1gcat/core/license.buildTime=$(date +%s)" -o your_program
```

## 工作原理

1. **编译时间注入**：通过构建参数 `-ldflags` 将当前时间戳注入到 `buildTime` 变量中
2. **授权天数设置**：通过 `SetExpiration()` 函数设置软件的授权天数
3. **授权检查**：程序启动时调用 `CheckLicense()` 函数进行授权检查
   - 如果当前时间超过编译时间加上授权天数，程序会自动退出
   - 如果授权天数未设置或为0，则认为是永久有效

## 注意事项

1. **生产环境**：确保在生产环境中正确注入编译时间戳
2. **授权天数**：授权天数应该根据实际的授权协议进行设置
3. **退出行为**：超过授权期限时，程序会无提示自动退出

## 示例

### 示例程序

```go
package main

import (
    "github.com/b1gcat/core/license"
)

func main() {
    // 设置授权天数为30天
    license.SetExpiration(30)
    
    // 检查授权
    license.CheckLicense()
    
    // 正常的程序逻辑
    // ...
}
```

### 编译命令

```bash
go build -ldflags "-X github.com/b1gcat/core/license.buildTime=$(date +%s)" -o myapp
```

## 测试

运行测试：

```bash
go test -v
```

测试会验证授权检查的基本功能。
