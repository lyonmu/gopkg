# pinyin

中文拼音转换工具，提供中文到拼音的各种格式转换及中文检测功能。基于 [go-pinyin](https://github.com/mozillazg/go-pinyin) 构建。

## 功能

- 自动过滤输入中的空格和 emoji
- 字母和数字原样保留，标点符号被过滤

## API

### `GetPinyinInitial(chinese string) string`

返回中文字符串的拼音首字母（大写）。字母和数字原样保留，标点符号和 emoji 被过滤。

| 输入 | 输出 |
|------|------|
| `"张三"` | `"ZS"` |
| `"张a3"` | `"ZA3"` |
| `"🎉张🔥三"` | `"ZS"` |

### `GetPinyinSpelling(chinese string) string`

返回中文字符串的完整拼音（全大写，无分隔符）。

| 输入 | 输出 |
|------|------|
| `"张三"` | `"ZHANGSAN"` |
| `"李四"` | `"LISI"` |

### `GetPinyinBigHump(chinese string) string`

返回中文字符串的拼音，大驼峰格式（每个拼音首字母大写）。

| 输入 | 输出 |
|------|------|
| `"张三"` | `"ZhangSan"` |

### `GetPinyinSmallHump(chinese string) string`

返回中文字符串的拼音，小驼峰格式（首字母小写，后续每个拼音首字母大写）。

| 输入 | 输出 |
|------|------|
| `"张三"` | `"zhangSan"` |
| `"你好世界"` | `"niHaoShiJie"` |

### `IsAllChinese(s string) bool`

判断字符串是否全部由中文字符组成。空格和 emoji 会在检测前被移除。

| 输入 | 输出 |
|------|------|
| `"张三"` | `true` |
| `"张 三"` | `true` |
| `"🎉张三"` | `true` |
| `"abc"` | `false` |
| `""` | `false` |

## 使用示例

```go
package main

import (
    "fmt"

    "github.com/lyonmu/gopkg/pinyin"
)

func main() {
    // 获取拼音首字母
    fmt.Println(pinyin.GetPinyinInitial("张三"))         // "ZS"
    fmt.Println(pinyin.GetPinyinInitial("张a3"))         // "ZA3"

    // 获取完整拼音
    fmt.Println(pinyin.GetPinyinSpelling("张三"))         // "ZHANGSAN"

    // 大驼峰格式
    fmt.Println(pinyin.GetPinyinBigHump("张三"))          // "ZhangSan"

    // 小驼峰格式
    fmt.Println(pinyin.GetPinyinSmallHump("你好世界"))    // "niHaoShiJie"

    // 检测是否全中文
    fmt.Println(pinyin.IsAllChinese("张三"))             // true
    fmt.Println(pinyin.IsAllChinese("abc"))              // false
}
```
