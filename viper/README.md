# viper

类型安全的配置管理工具，基于 [spf13/viper](https://github.com/spf13/viper) 封装。

## 功能

- 调用方传入配置结构体指针，工具库负责加载并写回
- 支持配置文件动态监听
- 配置成功更新后通过 `Watch()` 通知调用方
- 配置读取或解析失败后通过 `Errors()` 通知调用方，并保留旧配置
- 使用独立 `viper.Viper` 实例，避免污染全局配置状态

## API

### `NewConfigManager[T any](cfg *T) *ConfigManager[T]`

创建配置管理器。`cfg` 必须是非 nil 指针，首次加载和热重载成功都会写回该指针。

### `LoadConfig(path, filetype string) error`

从指定配置文件加载配置，并启动文件监听。

重复调用时，新配置成功加载后会替换当前配置和监听会话；旧监听器会被停止，并在 `LoadConfig` 返回前退出。新配置加载失败时，现有配置和监听会话保持不变。并发的加载与关闭操作由管理器协调。

| 参数 | 说明 |
|------|------|
| `path` | 配置文件路径 |
| `filetype` | 配置文件类型，例如 `yaml`、`json`、`toml` |

### `GetConfig() T`

线程安全地返回当前配置副本。

### `Watch() <-chan struct{}`

返回配置成功更新的状态通知通道。通道容量为 1；消费方尚未读取通知时，后续成功更新通知会被合并。

### `Errors() <-chan error`

返回动态监听期间的错误状态通知通道。通道容量为 1；消费方尚未读取错误时，后续错误通知可能被合并。它不是完整的错误事件日志。热重载失败时不会覆盖当前有效配置。

### `Close()`

停止当前配置文件监听并等待监听 goroutine 退出。重复调用安全；关闭后再次调用 `LoadConfig` 会返回 `ErrAlreadyClosed`。`Watch()` 和 `Errors()` 通道不会被关闭。

## 使用示例

```go
package main

import (
	"fmt"
	"log"

	config "github.com/lyonmu/gopkg/viper"
)

type AppConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
}

func main() {
	cfg := AppConfig{}
	cm := config.NewConfigManager(&cfg)
	defer cm.Close()

	if err := cm.LoadConfig("config.yaml", "yaml"); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case <-cm.Watch():
				fmt.Println("config reloaded:", cm.GetConfig())
			case err := <-cm.Errors():
				log.Println("config reload failed:", err)
			}
		}
	}()

	fmt.Println(cm.GetConfig())
}
```

## 注意事项

- `Watch()` 和 `Errors()` 是容量为 1 的合并状态通知，不是完整的更新或错误事件日志。
- 重复调用 `LoadConfig()` 只保留最新成功加载的配置和监听会话，旧监听器会在该调用返回前停止。
- `LoadConfig()` 与 `Close()` 的生命周期操作会被协调；关闭后不能重新加载。
- 热重载失败不会覆盖当前有效配置。
- 并发读取应使用 `GetConfig()`；直接读取传入的配置变量时，调用方需要自行保证并发安全。
