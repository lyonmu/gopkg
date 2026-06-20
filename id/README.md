# id

分布式唯一 ID 生成器，基于 [Sonyflake](https://github.com/sony/sonyflake) 算法实现。

## 安装

```bash
go get github.com/lyonmu/gopkg/id
```

## 快速开始

```go
package main

import (
	"fmt"

	"github.com/lyonmu/gopkg/id"
)

func main() {
	// 创建 ID 生成器，传入机器 ID 获取函数
	gen, err := id.NewSonySnowFlake(func() (int, error) {
		return 1, nil // 每台机器应使用不同的 ID
	})
	if err != nil {
		panic(err)
	}

	// 生成唯一 ID；生成失败时必须处理错误
	generatedID, err := gen.GenID()
	if err != nil {
		panic(err)
	}
	fmt.Println(generatedID)
}
```

## API

### `IDGenerator` 接口

```go
type IDGenerator interface {
	GenID() (int64, error)
}
```

所有 ID 生成器都实现此接口，方便 mock 和替换。

### `NewSonySnowFlake(machineId func() (int, error)) (IDGenerator, error)`

创建基于 Sonyflake 算法的 ID 生成器。

| 参数 | 说明 |
|------|------|
| `machineId` | 返回当前节点机器 ID 的函数，应确保集群内唯一 |

**返回值：**
- `IDGenerator` — 生成器实例
- `error` — 创建失败时返回错误（如 machineId 函数报错或机器 ID 为 0）

**默认配置：**

| 配置项 | 默认值 |
|--------|--------|
| 序列号位数 | 8 bits |
| 机器 ID 位数 | 16 bits |
| 时间粒度 | 10ms |
| 起始时间 | 2025-01-01 00:00:00 UTC |
| 机器 ID 校验 | 不允许为 0 |

**使用示例：**

```go
// 使用固定机器 ID
gen, err := id.NewSonySnowFlake(func() (int, error) {
	return 42, nil
})
if err != nil {
	return err
}

// 使用环境变量
gen, err = id.NewSonySnowFlake(func() (int, error) {
	mid := os.Getenv("MACHINE_ID")
	id, err := strconv.Atoi(mid)
	if err != nil {
		return 0, fmt.Errorf("invalid MACHINE_ID: %w", err)
	}
	return id, nil
})
if err != nil {
	return err
}
```

同一个机器 ID 只能由一个运行中的生成器实例使用，不能在多个进程或实例中并行复用。

### `GenID() (int64, error)`

生成下一个 ID。调用方必须检查返回的错误；发生时钟异常或底层生成失败时，不能使用同时返回的 ID。

```go
id1, err := gen.GenID()
if err != nil {
	return err
}

id2, err := gen.GenID()
if err != nil {
	return err
}
```

## 注意事项

- **机器 ID 唯一性**：集群中每个运行中的生成器必须使用不同的机器 ID；相同机器 ID 不能并行使用，否则可能生成重复 ID
- **机器 ID 范围**：默认 16 bits，取值范围 1-65535
- **固定时间基准**：默认起始时间固定为 `2025-01-01 00:00:00 UTC`，可降低同一机器 ID 在进程重启后生成重复 ID 的风险
- **快速重启边界**：默认时间粒度为 10ms；同一机器 ID 在同一个 10ms 时间窗口内快速重启时，固定时间基准不能绝对保证不重复
- **错误处理**：Sonyflake 不支持时钟回拨等异常情况，必须检查 `GenID()` 返回的错误，不应使用失败调用返回的 ID
