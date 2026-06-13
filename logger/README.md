# logger

基于 [zap](https://github.com/uber-go/zap) 和 [lumberjack](https://github.com/natefinch/lumberjack) 的日志工具，提供库友好的默认配置和可选文件轮转输出。

## 功能

- 默认只输出到控制台，不创建日志目录
- 显式启用文件输出后，按模块名写入并自动轮转
- 支持 `console` 和 `json` 两种输出格式
- 支持 `debug`、`info`、`warn`、`error` 四种级别
- 配置结构可通过 `mapstructure`、`yaml`、`json` 标签接入配置文件

## API

### `DefaultConfig() Config`

返回安全默认配置：

| 字段 | 默认值 |
|------|--------|
| `Module` | `"app"` |
| `Level` | `InfoLevel` |
| `Format` | `ConsoleFormat` |
| `Output.Console.Enabled` | `true` |
| `Output.File.Enabled` | `false` |

### `New(config Config) (*zap.Logger, error)`

根据配置创建 `*zap.Logger`。

| 行为 | 说明 |
|------|------|
| 空 `Module` | 自动补为 `"app"` |
| 空 `Level` | 自动补为 `InfoLevel` |
| 空 `Format` | 自动补为 `ConsoleFormat` |
| 无输出启用 | 自动回退到控制台输出 |
| 非法级别或格式 | 返回错误 |
| 启用文件但路径为空 | 返回错误 |

### `NewDefault() (*zap.Logger, error)`

使用 `DefaultConfig()` 创建 logger。

## 使用示例

### 默认控制台输出

```go
package main

import (
	"log"

	"github.com/lyonmu/gopkg/logger"
)

func main() {
	zapLogger, err := logger.NewDefault()
	if err != nil {
		log.Fatal(err)
	}
	defer zapLogger.Sync()

	zapLogger.Info("application started")
}
```

### 启用文件输出

```go
package main

import (
	"log"

	"github.com/lyonmu/gopkg/logger"
)

func main() {
	cfg := logger.DefaultConfig()
	cfg.Module = "api"
	cfg.Level = logger.DebugLevel
	cfg.Output.File.Enabled = true
	cfg.Output.File.Path = "./logs"

	zapLogger, err := logger.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer zapLogger.Sync()

	zapLogger.Info("file output enabled")
}
```

日志文件路径为：

```text
./logs/api/api.log
```

### JSON 输出

```go
cfg := logger.DefaultConfig()
cfg.Module = "worker"
cfg.Format = logger.JSONFormat

zapLogger, err := logger.New(cfg)
if err != nil {
	log.Fatal(err)
}
defer zapLogger.Sync()

zapLogger.Warn("job delayed")
```

JSON 格式会附带 `module` 字段，便于日志检索。

## 注意事项

- 默认配置不会写文件，适合作为工具库被其他项目引用。
- 文件输出需要显式设置 `Output.File.Enabled = true` 和 `Output.File.Path`。
- 文件轮转默认值为 `MaxSize: 10`、`MaxAge: 7`、`MaxBackups: 3`，单位分别为 MB、天、备份数量。
