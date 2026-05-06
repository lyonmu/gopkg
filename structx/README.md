# structx

struct 操作工具包，提供 struct 转 map、差异比较、属性赋值等功能，基于 Go 反射实现。

## 功能

- 支持指针类型输入（自动解引用）
- 支持嵌套 struct、slice、map 等复杂类型
- 递归深度限制为 10 层，防止循环引用

## API

### `StructToMap(v any) (map[string]any, error)`

将 struct 转换为 `map[string]any`，key 为字段名。

| 输入 | 输出 |
|------|------|
| `struct{A int; B string}{A: 1, B: "hi"}` | `map[string]any{"A": 1, "B": "hi"}` |
| `&struct{X int}{X: 42}` | `map[string]any{"X": 42}` |
| `"string"` | `nil, error` |

### `DiffStruct(dst, src any) (map[string]any, []string, error)`

比较两个同类型 struct，返回不同的字段 map 和字段名列表。

| dst | src | 返回 |
|-----|-----|------|
| `{Name:"a", Age:10}` | `{Name:"a", Age:20}` | `map{"Age":10}, ["Age"]` |
| `{Name:"a"}` | `{Name:"a"}` | `map{}, []` |

### `Assign(dst, src any) error`

将 src 中非零值的字段赋值给 dst。dst 必须是指针。

| dst | src | 结果 |
|-----|-----|------|
| `&{Name:"old"}` | `{Name:"new", Age:5}` | `&{Name:"new", Age:5}` |
| `&{Name:"keep"}` | `{}` | `&{Name:"keep"}`（不变） |

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

    // Assign
    dst := &User{Name: "old"}
    src := User{Name: "new", Age: 25}
    structx.Assign(dst, src)
    fmt.Println(dst) // &{new 25}
}
```
