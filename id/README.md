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

	// 生成唯一 ID
	id := gen.GenID()
	fmt.Println(id) // 例如: 13835058055282114561
}
```

## API

### `IDGenerator` 接口

```go
type IDGenerator interface {
	GenID() int64
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
- `error` — 创建失败时返回错误（如 machineId 函数报错或返回 0）

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

// 使用环境变量
gen, err := id.NewSonySnowFlake(func() (int, error) {
	mid := os.Getenv("MACHINE_ID")
	id, err := strconv.Atoi(mid)
	if err != nil {
		return 0, fmt.Errorf("invalid MACHINE_ID: %w", err)
	}
	return id, nil
})
```

### `GenID() int64`

生成下一个唯一 ID。ID 单调递增，在同一时间窗口内通过序列号保证唯一性。

```go
id1 := gen.GenID()
id2 := gen.GenID()
// id2 > id1
```

## 注意事项

- **机器 ID 唯一性**：集群中每个节点必须使用不同的机器 ID，否则可能生成重复 ID
- **机器 ID 范围**：默认 16 bits，取值范围 1-65535
- **时钟回拨**：Sonyflake 不支持时钟回拨，系统时间倒退时 `GenID()` 会返回 0
- **ID 大小**：返回的 int64 始终为正数，可安全用于数据库主键
