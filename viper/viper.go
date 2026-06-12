package viper

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const defaultChannelSize = 1

var (
	ErrNilConfig     = errors.New("config pointer cannot be nil")
	ErrAlreadyClosed = errors.New("config manager is closed")
)

// ConfigManager 是一个泛型配置管理器，支持类型安全的配置加载和动态监听。
type ConfigManager[T any] struct {
	mu      sync.RWMutex
	cfg     *T
	current T

	v       *viper.Viper
	watcher *fsnotify.Watcher

	configFile     string
	realConfigFile string

	watchCh chan struct{}
	errCh   chan error
	doneCh  chan struct{}

	closeOnce sync.Once
	closed    bool
}

// NewConfigManager 创建一个新的泛型配置管理器。
func NewConfigManager[T any](cfg *T) *ConfigManager[T] {
	return &ConfigManager[T]{
		cfg:     cfg,
		v:       viper.New(),
		watchCh: make(chan struct{}, defaultChannelSize),
		errCh:   make(chan error, defaultChannelSize),
		doneCh:  make(chan struct{}),
	}
}

// LoadConfig 从指定路径加载配置文件并启用动态监听。
func (cm *ConfigManager[T]) LoadConfig(path string, filetype string) error {
	if cm.cfg == nil {
		return ErrNilConfig
	}
	if cm.isClosed() {
		return ErrAlreadyClosed
	}

	cm.v.SetConfigFile(path)
	if filetype != "" {
		cm.v.SetConfigType(filetype)
	}

	cfg, err := cm.readConfig()
	if err != nil {
		return err
	}
	if err := cm.startWatcher(path); err != nil {
		return err
	}

	cm.updateConfig(cfg)

	return nil
}

// GetConfig 获取当前配置的副本，线程安全。
func (cm *ConfigManager[T]) GetConfig() T {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.current
}

// Watch 返回一个通道，用于监听配置成功更新事件。
func (cm *ConfigManager[T]) Watch() <-chan struct{} {
	return cm.watchCh
}

// Errors 返回一个通道，用于监听配置动态加载期间发生的错误。
func (cm *ConfigManager[T]) Errors() <-chan error {
	return cm.errCh
}

// Close 停止监听配置变化。重复调用是安全的。
func (cm *ConfigManager[T]) Close() {
	cm.closeOnce.Do(func() {
		cm.mu.Lock()
		cm.closed = true
		watcher := cm.watcher
		cm.mu.Unlock()

		close(cm.doneCh)
		if watcher != nil {
			_ = watcher.Close()
		}
	})
}

func (cm *ConfigManager[T]) readConfig() (T, error) {
	var cfg T

	if err := cm.v.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := cm.v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

func (cm *ConfigManager[T]) updateConfig(cfg T) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.current = cfg
	*cm.cfg = cfg
}

func (cm *ConfigManager[T]) startWatcher(path string) error {
	cleanConfigFile := filepath.Clean(path)
	configDir := filepath.Dir(cleanConfigFile)
	realConfigFile, _ := filepath.EvalSymlinks(cleanConfigFile)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create config watcher: %w", err)
	}
	if err := watcher.Add(configDir); err != nil {
		_ = watcher.Close()
		return fmt.Errorf("watch config directory: %w", err)
	}

	cm.mu.Lock()
	if cm.watcher != nil {
		oldWatcher := cm.watcher
		_ = oldWatcher.Close()
	}
	cm.watcher = watcher
	cm.configFile = cleanConfigFile
	cm.realConfigFile = realConfigFile
	cm.mu.Unlock()

	go cm.watchLoop(watcher)

	return nil
}

func (cm *ConfigManager[T]) watchLoop(watcher *fsnotify.Watcher) {
	for {
		select {
		case <-cm.doneCh:
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if cm.shouldReload(event) {
				cm.reload()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			cm.notifyError(fmt.Errorf("watch config: %w", err))
		}
	}
}

func (cm *ConfigManager[T]) shouldReload(event fsnotify.Event) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	currentConfigFile, _ := filepath.EvalSymlinks(cm.configFile)
	configChanged := filepath.Clean(event.Name) == cm.configFile &&
		(event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename))
	symlinkChanged := currentConfigFile != "" && currentConfigFile != cm.realConfigFile

	if symlinkChanged {
		cm.realConfigFile = currentConfigFile
	}

	return configChanged || symlinkChanged
}

func (cm *ConfigManager[T]) reload() {
	cfg, err := cm.readConfig()
	if err != nil {
		cm.notifyError(fmt.Errorf("reload config: %w", err))
		return
	}

	cm.updateConfig(cfg)
	cm.notifyWatch()
}

func (cm *ConfigManager[T]) notifyWatch() {
	if cm.isClosed() {
		return
	}

	select {
	case cm.watchCh <- struct{}{}:
	default:
	}
}

func (cm *ConfigManager[T]) notifyError(err error) {
	if cm.isClosed() {
		return
	}

	select {
	case cm.errCh <- err:
	default:
	}
}

func (cm *ConfigManager[T]) isClosed() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.closed
}
