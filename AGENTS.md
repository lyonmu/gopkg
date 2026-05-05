# AGENTS.md

## Project

Go utility toolkit collecting reusable functions from development work. Module path: `github.com/lyonmu/gopkg`.

## Commands

- Run all tests: `go test ./...`
- Run single package tests: `go test ./<pkg>/...`
- Add a dependency: `go get <module>`
- Tidy dependencies: `go mod tidy`

## Project Structure

```
gopkg/
├── README.md              # 项目总览，模块列表及链接
├── AGENTS.md              # 项目指令文件（本文件）
├── go.mod
├── go.sum
├── LICENSE
├── <pkg>/                 # 子包目录（每个工具函数一个包）
│   ├── <pkg>.go          # 实现文件
│   ├── <pkg>_test.go     # 测试文件
│   └── README.md         # 该包的 API 文档和用法示例
└── docs/                  # 临时文档目录（gitignored，不提交）
```

## Structure Conventions

- **每个子包独立目录** — 工具函数集合放在项目根目录下的独立子目录中
- **根 `README.md` 只做高层说明** — 项目简介、模块表格（链接到子包 README）、安装命令、许可证
- **每个子包必须有 `README.md`** — 包含功能说明、API 文档、输入输出示例、用法代码片段
- **`docs/` 是 gitignored 的** — 不提交到仓库，用于临时文档或草稿


## Code Conventions

- **Go 1.24.0** 最小版本
- **错误处理** — 工具函数优先返回零值而非 panic

## Testing

- **Table-driven tests** — 所有测试使用 table-driven 风格，写在 `_test.go` 文件中
- **测试覆盖** — 新功能必须包含正常输入、边界情况的测试用例
- **验证顺序** — 代码变更完成后，运行 `go test ./...` 确认所有测试通过
- **单包测试** — 开发过程中使用 `go test ./<pkg>/...` 快速验证

## Linting

- **Go 格式化** — 代码需符合 `go fmt` 或 `gofumpt` 规范
- **Import 规范** — import 分组（标准库 → 第三方 → 本地），组间空行分隔

## Git Commit Format

采用 Conventional Commits 规范：

```
<type>[optional scope]: <description>
```

### Types

| Type | 用途 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(pinyin): 添加小驼峰拼音转换` |
| `fix` | Bug 修复 | `fix(pinyin): 修复 emoji 过滤不完整` |
| `docs` | 文档变更 | `docs: 更新 README 模块列表` |
| `style` | 代码格式（不影响逻辑） | `style: 统一 import 分组` |
| `refactor` | 重构（非新功能非 bug） | `refactor(pinyin): 提取 cleanInput 函数` |
| `test` | 测试相关 | `test(pinyin): 添加边界情况测试` |
| `chore` | 构建/工具/杂项 | `chore: 更新 go.mod` |

### Rules

- scope 使用子包名，全局变更可省略
- 中文描述，祈使句，首字母不大写，末尾不加句号
- 不超过 50 个字符

### Examples

```
feat(pinyin): 添加拼音首字母转换功能

fix(pinyin): 修复空字符串输入返回异常

docs: 添加 pinyin 模块 README

test(pinyin): 添加 emoji 混合输入测试

chore: 更新 go.mod 依赖版本
```
