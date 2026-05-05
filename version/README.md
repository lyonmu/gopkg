# version

版本信息管理工具包。支持通过编译时 `-ldflags` 注入版本信息，并提供多种格式化输出方法。

## 安装

```bash
go get github.com/lyonmu/gopkg/version
```

## 快速开始

### 1. 编译时注入

在 Makefile 中定义注入参数：

```makefile
VERSION   ?= $(shell git describe --tags --always --dirty)
BRANCH    ?= $(shell git rev-parse --abbrev-ref HEAD)
REVISION  ?= $(shell git rev-parse HEAD)
BUILDUSER ?= $(shell whoami)
BUILDDATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X github.com/lyonmu/gopkg/version.Version=$(VERSION) \
           -X github.com/lyonmu/gopkg/version.Branch=$(BRANCH) \
           -X github.com/lyonmu/gopkg/version.Revision=$(REVISION) \
           -X github.com/lyonmu/gopkg/version.BuildUser=$(BUILDUSER) \
           -X github.com/lyonmu/gopkg/version.BuildDate=$(BUILDDATE)

build:
	go build -ldflags "$(LDFLAGS)" -o myapp ./cmd/myapp
```

### 2. 代码中使用

```go
package main

import (
	"fmt"
	"log/slog"

	"github.com/lyonmu/gopkg/version"
)

func main() {
	// 简洁格式输出
	fmt.Println(version.Info())
	// (version=1.0.0, branch=main, revision=abc123)

	// 完整格式输出
	fmt.Println(version.Print("myapp"))
	// myapp, version 1.0.0 (branch: main, revision: abc123)
	//   build user:       developer
	//   build date:       2026-05-05
	//   go version:       go1.24.0
	//   platform:         linux/amd64
	//   tags:             netgo

	// 构建上下文
	fmt.Println(version.BuildContext())
	// (go=go1.24.0, platform=linux/amd64, user=developer, date=2026-05-05, tags=netgo)

	// 结构化日志
	logger := slog.Default()
	logger.Info("Starting server", version.Slog()...)

	// User-Agent 字符串
	ua := version.ComponentUserAgent("MyApp")
	fmt.Println(ua) // MyApp/1.0.0
}
```

## 编译时注入参数

通过 `-ldflags "-X"` 注入以下变量：

| 变量 | 说明 | 示例值 |
|------|------|--------|
| `Version` | 语义化版本号 | `1.0.0` |
| `Revision` | Git 提交哈希，留空则自动从 `debug.ReadBuildInfo()` 获取 | `abc123def` |
| `Branch` | Git 分支名 | `main` |
| `BuildUser` | 构建执行者 | `developer` |
| `BuildDate` | 构建日期 | `2026-05-05T10:30:00Z` |

`GoVersion`、`GoOS`、`GoArch` 在运行时自动获取，无需注入。

## API 参考

### 变量

```go
var (
	Version   string // 版本号
	Revision  string // Git revision
	Branch    string // Git 分支
	BuildUser string // 构建用户
	BuildDate string // 构建日期
	GoVersion string // Go 运行时版本（自动获取）
	GoOS      string // 操作系统（自动获取）
	GoArch    string // CPU 架构（自动获取）
)
```

### 函数

| 函数 | 说明 |
|------|------|
| `Print(program string) string` | 返回完整的版本信息，使用模板格式化 |
| `Info() string` | 返回简短版本信息 `(version=..., branch=..., revision=...)` |
| `BuildContext() string` | 返回构建上下文信息 |
| `Slog() []any` | 返回 key-value 对切片，用于结构化日志 |
| `GetRevision() string` | 获取 revision，优先使用注入值，否则返回运行时计算值 |
| `GetTags() string` | 返回编译时的 build tags |
| `ComponentUserAgent(component string) string` | 生成 `component/Version` 格式的 User-Agent 字符串 |
