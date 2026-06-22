# structx

struct 操作工具包，提供 struct 转 map、差异比较、属性赋值等功能，基于 Go 反射实现。

## 功能

- 支持指针类型输入（自动解引用）
- 支持嵌套 struct、slice、array、map 等复杂类型
- slice 和 array 统一转换为 `[]any`
- 输出 key 始终使用 Go 导出字段名，不读取 struct tag
- 递归深度最大为 10；超过限制的分支返回 `nil`
- 仅检测当前递归路径中的指针循环，循环分支返回 `nil`
- `time.Time` 等声明了字段但没有导出字段的值 struct 保留原始值；零字段 struct 转换为空 map

## API

### `StructToMap(v any) (map[string]any, error)`

将 struct 转换为 `map[string]any`。只包含可导出的字段，key 使用 Go 字段名，不读取 `json`、`mapstructure` 等 tag。

| 输入 | 输出 |
|------|------|
| `struct{A int; B string}{A: 1, B: "hi"}` | `map[string]any{"A": 1, "B": "hi"}` |
| `&struct{X int}{X: 42}` | `map[string]any{"X": 42}` |
| `struct{A [2]int}{A: [2]int{1, 2}}` | `map[string]any{"A": []any{1, 2}}` |
| `"string"` | `nil, error` |

嵌套 struct、slice、array 和 map 会递归转换。递归最多处理 10 层；超过限制或遇到当前递归路径上的指针循环时，对应分支为 `nil`。共享但不构成当前路径循环的指针会分别正常转换。`time.Time` 以及其他声明了字段但没有导出字段的值 struct 不展开，保留原始值；零字段 struct 转换为 `map[string]any{}`。

### `DiffStruct(dst, src any) (map[string]any, []string, error)`

比较两个同类型 struct，返回不同的字段 map 和字段名列表。

| dst | src | 返回 |
|-----|-----|------|
| `{Name:"a", Age:10}` | `{Name:"a", Age:20}` | `map{"Age":10}, ["Age"]` |
| `{Name:"a"}` | `{Name:"a"}` | `map{}, []` |

### `AssignNonZero(dst, src any) error`

将 src 中非零值字段赋值给 dst。dst 必须是指针。零值字段（如 `0`、`""`、`false`、`nil`）不会被赋值。

| dst | src | 结果 |
|-----|-----|------|
| `&{Name:"old"}` | `{Name:"new", Age:5}` | `&{Name:"new", Age:5}` |
| `&{Name:"keep"}` | `{}` | `&{Name:"keep"}`（不变） |

### `AssignOverwrite(dst, src any) error`

将 src 中所有字段赋值给 dst，包括零值字段。dst 必须是指针。

| dst | src | 结果 |
|-----|-----|------|
| `&{Name:"old", Age:10}` | `{Name:"new"}` | `&{Name:"new", Age:0}` |
| `&{Active:true}` | `{Active:false}` | `&{Active:false}` |

### `Assign(dst, src any) error`

已弃用：使用 `AssignNonZero` 获取明确语义。行为与 `AssignNonZero` 完全一致。

## 使用示例

```go
package main

import (
    "fmt"

    "github.com/lyonmu/gopkg/structx"
)

type User struct {
    Name string
    Age  int
}

func main() {
    // StructToMap
    u := User{Name: "Alice", Age: 30}
    m, _ := structx.StructToMap(u)
    fmt.Println(m) // map[Age:30 Name:Alice]

    // DiffStruct
    u1 := User{Name: "Alice", Age: 30}
    u2 := User{Name: "Alice", Age: 31}
    diff, fields, _ := structx.DiffStruct(u1, u2)
    fmt.Println(diff)   // map[Age:30]
    fmt.Println(fields) // [Age]

    // AssignNonZero
    dst := &User{Name: "old"}
    src := User{Name: "new", Age: 25}
    structx.AssignNonZero(dst, src)
    fmt.Println(dst) // &{new 25}
}
```
