# version

版本信息管理工具包。支持通过编译时 `-ldflags` 注入版本信息，并自动从 `debug.ReadBuildInfo()` 检测 VCS 信息。

## 安装

```bash
go get github.com/lyonmu/gopkg/version
```

## 快速开始

### 1. 编译时注入

在 Makefile 中定义注入参数：

```makefile
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse HEAD)
BRANCH  ?= $(shell git rev-parse --abbrev-ref HEAD)

LDFLAGS := -X github.com/lyonmu/gopkg/version.Version=$(VERSION) \
           -X github.com/lyonmu/gopkg/version.Commit=$(COMMIT) \
           -X github.com/lyonmu/gopkg/version.Branch=$(BRANCH)

build:
	go build -ldflags "$(LDFLAGS)" -o myapp ./cmd/myapp
```

> `Commit` 可不注入，`GetCommit()`、`Info()` 和 `Slog()` 会自动从
> `debug.ReadBuildInfo()` 的 VCS 信息中获取。`Version` 和 `Branch` 没有自动检测逻辑。

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
	// (branch=main, commit=abc123)

	// 完整格式输出
	fmt.Println(version.Print("myapp"))
	// myapp, (branch: main, commit: abc123)
	// program version:	v1.2.3
	// go version:	go1.24.0
	// platform:	linux/amd64
	// tags:	netgo

	// 构建上下文
	fmt.Println(version.BuildContext())
	// (go=go1.24.0, platform=linux/amd64, tags=netgo)

	// 结构化日志
	logger := slog.Default()
	logger.Info("Starting server", version.Slog()...)

	// 自动检测 commit 和 tags
	fmt.Println(version.GetCommit()) // abc123def 或 abc123def-modified
	fmt.Println(version.GetTags())   // netgo 或 unknown
}
```

## 编译时注入参数

通过 `-ldflags "-X"` 注入以下变量：

| 变量 | 说明 | 示例值 |
|------|------|--------|
| `Version` | 程序版本号，由构建系统注入 | `v1.2.3` |
| `Commit` | Git 提交哈希；`GetCommit()` 在该值为空时使用自动检测结果 | `abc123def` |
| `Branch` | Git 分支名 | `main` |

`GoVersion`、`GoOS`、`GoArch` 在运行时自动获取，无需注入。

## 自动检测机制

当未通过 `-ldflags` 注入 `Commit` 时，`GetCommit()` 会自动从
`debug.ReadBuildInfo()` 中读取 VCS 信息：

- **vcs.revision** — Git 提交哈希
- **vcs.modified** — 如果有未提交的修改，会在 commit 后追加 `-modified` 后缀
- **-tags** — 编译时使用的 build tags

如果构建信息不可用，自动检测的 commit 和 build tags 均为 `unknown`。
`Commit` 非空时始终优先于自动检测结果。

> `Print()` 的 `commit` 字段直接读取公开变量 `Commit`，不会回退到自动检测结果。
> 如需在完整输出中显示提交号，请在编译时注入 `Commit`。

```bash
# 直接 go build（无 -ldflags）
go build -o myapp ./cmd/myapp

./myapp --version
# myapp, (branch: , commit: )
# program version:
# go version:	go1.24.0
# platform:	linux/amd64
# tags:	unknown
```

## Print 输出格式

`Print(program)` 使用预编译模板输出多行版本信息：

```
{{program}}, (branch: {{branch}}, commit: {{commit}})
program version:	{{version}}
go version:	{{goVersion}}
platform:	{{platform}}
tags:	{{tags}}
```

## API 参考

### 变量

```go
var (
	Version   string // 程序版本号，可通过 -ldflags 注入
	Commit    string // Git commit，可通过 -ldflags 注入
	Branch    string // Git 分支，可通过 -ldflags 注入
	GoVersion string // Go 运行时版本（自动获取，默认 runtime.Version()）
	GoOS      string // 操作系统（自动获取，默认 runtime.GOOS）
	GoArch    string // CPU 架构（自动获取，默认 runtime.GOARCH）
)
```

### 函数

| 函数 | 说明 |
|------|------|
| `Print(program string) string` | 返回包含程序版本、分支、提交号和构建上下文的完整信息 |
| `Info() string` | 返回简短版本信息 `(branch=..., commit=...)` |
| `BuildContext() string` | 返回构建上下文 `(go=..., platform=..., tags=...)` |
| `Slog() []any` | 返回 5 对 key-value，用于结构化日志 |
| `GetCommit() string` | 获取 commit，优先使用注入值，否则返回自动检测结果（可能带 `-modified` 后缀） |
| `GetTags() string` | 返回编译时的 build tags |

### Slog 返回的键值对

```go
[]any{
	"commit",    GetCommit(),
	"branch",    Branch,
	"goversion", GoVersion,
	"goos",      GoOS,
	"goarch",    GoArch,
}
```

用法示例：

```go
logger.Info("server starting", version.Slog()...)
// 输出: {"level":"info","msg":"server starting","commit":"abc123","branch":"main","goversion":"go1.24.0","goos":"linux","goarch":"amd64"}
```
