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
	lifecycleMu sync.Mutex
	mu          sync.RWMutex
	cfg         *T
	current     T
	closed      bool
	session     *watchSession
	watchCh     chan struct{}
	errCh       chan error
}

type watchSession struct {
	v              *viper.Viper
	watcher        *fsnotify.Watcher
	configFile     string
	realConfigFile string
	doneCh         chan struct{}
	closeOnce      sync.Once
}

func (session *watchSession) close() {
	session.closeOnce.Do(func() {
		close(session.doneCh)
		_ = session.watcher.Close()
	})
}

// NewConfigManager 创建一个新的泛型配置管理器。
func NewConfigManager[T any](cfg *T) *ConfigManager[T] {
	return &ConfigManager[T]{
		cfg:     cfg,
		watchCh: make(chan struct{}, defaultChannelSize),
		errCh:   make(chan error, defaultChannelSize),
	}
}

// LoadConfig 从指定路径加载配置文件并启用动态监听。
func (cm *ConfigManager[T]) LoadConfig(path string, filetype string) error {
	if cm.cfg == nil {
		return ErrNilConfig
	}

	cm.lifecycleMu.Lock()
	defer cm.lifecycleMu.Unlock()

	if cm.closed {
		return ErrAlreadyClosed
	}

	session, cfg, err := cm.newSession(path, filetype)
	if err != nil {
		return err
	}

	oldSession := cm.session
	cm.session = session
	cm.updateConfig(cfg)
	go cm.watchLoop(session)
	if oldSession != nil {
		oldSession.close()
	}

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
	cm.lifecycleMu.Lock()
	if cm.closed {
		cm.lifecycleMu.Unlock()
		return
	}

	cm.closed = true
	session := cm.session
	cm.session = nil
	cm.lifecycleMu.Unlock()

	if session != nil {
		session.close()
	}
}

func (cm *ConfigManager[T]) updateConfig(cfg T) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.current = cfg
	*cm.cfg = cfg
}

func (cm *ConfigManager[T]) newSession(path string, filetype string) (*watchSession, T, error) {
	var cfg T

	cleanConfigFile := filepath.Clean(path)
	configDir := filepath.Dir(cleanConfigFile)
	realConfigFile, _ := filepath.EvalSymlinks(cleanConfigFile)

	v := viper.New()
	v.SetConfigFile(cleanConfigFile)
	if filetype != "" {
		v.SetConfigType(filetype)
	}
	if err := v.ReadInConfig(); err != nil {
		return nil, cfg, fmt.Errorf("read config: %w", err)
	}
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, cfg, fmt.Errorf("unmarshal config: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, cfg, fmt.Errorf("create config watcher: %w", err)
	}
	if err := watcher.Add(configDir); err != nil {
		_ = watcher.Close()
		return nil, cfg, fmt.Errorf("watch config directory: %w", err)
	}

	return &watchSession{
		v:              v,
		watcher:        watcher,
		configFile:     cleanConfigFile,
		realConfigFile: realConfigFile,
		doneCh:         make(chan struct{}),
	}, cfg, nil
}

func (cm *ConfigManager[T]) watchLoop(session *watchSession) {
	for {
		select {
		case <-session.doneCh:
			return
		case event, ok := <-session.watcher.Events:
			if !ok {
				return
			}
			if session.shouldReload(event) {
				cm.reload(session)
			}
		case err, ok := <-session.watcher.Errors:
			if !ok {
				return
			}
			cm.notifyError(session, fmt.Errorf("watch config: %w", err))
		}
	}
}

func (session *watchSession) shouldReload(event fsnotify.Event) bool {
	currentConfigFile, _ := filepath.EvalSymlinks(session.configFile)
	configChanged := filepath.Clean(event.Name) == session.configFile &&
		(event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename))
	symlinkChanged := currentConfigFile != "" && currentConfigFile != session.realConfigFile

	if symlinkChanged {
		session.realConfigFile = currentConfigFile
	}

	return configChanged || symlinkChanged
}

func (cm *ConfigManager[T]) reload(session *watchSession) {
	var cfg T

	if err := session.v.ReadInConfig(); err != nil {
		cm.notifyError(session, fmt.Errorf("reload config: read config: %w", err))
		return
	}
	if err := session.v.Unmarshal(&cfg); err != nil {
		cm.notifyError(session, fmt.Errorf("reload config: unmarshal config: %w", err))
		return
	}

	cm.lifecycleMu.Lock()
	defer cm.lifecycleMu.Unlock()

	if cm.closed || cm.session != session {
		return
	}

	cm.updateConfig(cfg)
	cm.notifyWatch()
}

func (cm *ConfigManager[T]) notifyWatch() {
	select {
	case cm.watchCh <- struct{}{}:
	default:
	}
}

func (cm *ConfigManager[T]) notifyError(session *watchSession, err error) {
	cm.lifecycleMu.Lock()
	defer cm.lifecycleMu.Unlock()

	if cm.closed || cm.session != session {
		return
	}

	select {
	case cm.errCh <- err:
	default:
	}
}
