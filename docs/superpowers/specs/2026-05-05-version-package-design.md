# Version 包清理和完善 — 设计文档

## 背景

`version/` 目录包含从 Prometheus 项目移植的版本信息管理代码，当前存在以下问题：

1. 包含 Prometheus 特定函数（`PrometheusUserAgent()`）和注释引用（`promslog`），不适合作为通用工具包
2. `Print()` 函数每次调用都重新解析模板，存在性能浪费
3. 缺少测试文件和模块文档
4. `version/` 目录未加入 git 跟踪

**方案选择**：方案 A — 保守清理，不改变任何公开 API 签名。

## 代码变更

### 移除

- `PrometheusUserAgent()` 函数（第 98-100 行）
- `Slog()` 函数注释中的 `promslog` 示例引用

### 优化

- 将 `Print()` 中的模板解析移至包级变量或 `init()` 中预编译，避免每次调用重复解析
- 保留 `template.Must` 和当前 panic 行为 — 内置固定模板正确性在编译期确定

### 保留不动

- 所有公开变量：`Version`, `Revision`, `Branch`, `BuildUser`, `BuildDate`, `GoVersion`, `GoOS`, `GoArch`
- 所有公开函数签名：`Print`, `Info`, `BuildContext`, `Slog`, `GetRevision`, `GetTags`, `ComponentUserAgent`
- `computeRevision()` 和 `init()` 逻辑

## 测试设计

文件：`version/version_test.go`

使用 table-driven tests 覆盖：

| 函数 | 测试场景 |
|------|---------|
| `Info()` | 空值、有值情况下的格式化输出 |
| `BuildContext()` | 空值、有值情况下的格式化输出 |
| `GetRevision()` | `Revision` 为空时返回 `computedRevision`，不为空时返回 `Revision` |
| `GetTags()` | 返回 `computedTags` |
| `Print(program)` | 输出包含程序名、版本号、branch 等信息 |
| `Slog()` | 返回的 slice 包含预期的 key-value 对 |
| `ComponentUserAgent()` | 拼接 component + "/" + Version |

不测试 `computeRevision()`：它依赖 `debug.ReadBuildInfo()`，行为由运行时环境决定。间接通过 `GetRevision()` 覆盖。

## 文档

### version/README.md

内容：
- 功能简介：编译时版本信息注入工具
- 安装命令：`go get github.com/lyonmu/gopkg/version`
- 使用示例：
  - Makefile/GitHub Actions 中 `-ldflags` 注入方式
  - Go 代码中调用 `version.Info()`、`version.Print()` 示例
  - 输出示例
- API 说明：每个公开函数的简要说明
- 编译注入参数列表：`-X` 可用的变量名及说明

### 根 README.md

模块表格中添加一行：
```
| [version](version/README.md) | 版本信息管理，支持编译时注入 |
```

## 文件变更清单

| 文件 | 操作 |
|------|------|
| `version/version.go` | 修改（移除 Prometheus 代码，优化模板缓存） |
| `version/version_test.go` | 新增 |
| `version/README.md` | 新增 |
| `README.md` | 修改（添加 version 模块条目） |
