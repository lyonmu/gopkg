// Package viper provides a generic configuration manager with type-safe
// loading, dynamic reloading, and lifecycle management built on top of
// github.com/spf13/viper.
//
// # Why not use viper.WatchConfig directly?
//
// viper's built-in WatchConfig (as of v1.21) does not support:
//   - Error notification callbacks to the caller (errors are only logged internally)
//   - Graceful shutdown of the watch goroutine (no Close API)
//   - Lifecycle management (no way to switch config files at runtime)
//   - Concurrent LoadConfig calls (no serialization)
//   - Automatic unmarshal into a generic config type T
//
// This package wraps viper with a session-based architecture that provides
// those capabilities while still leveraging viper's config parsing and
// file format support.
package viper
