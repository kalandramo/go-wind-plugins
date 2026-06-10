package hptimer

// Option 是高精度定时器服务器的配置选项。
type Option func(*Server)

// WithGracefullyShutdown 设置是否启用优雅关闭模式。
func WithGracefullyShutdown(enable bool) Option {
	return func(s *Server) {
		s.gracefullyShutdown = enable
	}
}

// WithTimerObserver 设置定时器观察者，用于接收任务触发事件。
// 如果不设置，默认使用内置的回调机制（TimerTask.Callback）。
func WithTimerObserver(observer TimerObserver) Option {
	return func(s *Server) {
		s.timerObserver = observer
	}
}
