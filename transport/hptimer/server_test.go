package hptimer

import (
	"context"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	// 创建高精度定时器服务
	srv := NewServer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(ctx); err != nil {
			t.Errorf("server start failed: %v", err)
		}
	}()

	// 等待引擎启动
	time.Sleep(50 * time.Millisecond)

	// 注册一个高精度定时任务，触发后通过 channel 通知
	triggerCh := make(chan struct{}, 1)
	taskID := "test_task"
	innerTaskID := srv.AddTask(&TimerTask{
		ID:       TimerTaskID(taskID),
		At:       time.Now().Add(50 * time.Millisecond),
		Interval: 0,
		Callback: func(ctx context.Context) error {
			triggerCh <- struct{}{}
			return nil
		},
	})
	if innerTaskID == "" {
		t.Fatalf("AddTask failed: %v", taskID)
	}

	// 验证任务是否按时触发
	select {
	case <-triggerCh:
		// success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("定时任务未按时触发")
	}

	// 删除任务（已触发应返回false）
	removed := srv.RemoveTask(TimerTaskID(taskID))
	if removed {
		t.Errorf("RemoveTask 应返回 false (已触发)")
	}

	// 优雅关闭
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestServerRepeatTask(t *testing.T) {
	srv := NewServer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = srv.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// 循环任务：每 50ms 触发一次
	triggerCount := 0
	done := make(chan struct{}, 1)

	srv.AddTask(&TimerTask{
		ID:       TimerTaskID("repeat_task"),
		At:       time.Now().Add(50 * time.Millisecond),
		Interval: 50 * time.Millisecond,
		Callback: func(ctx context.Context) error {
			triggerCount++
			if triggerCount >= 3 {
				select {
				case done <- struct{}{}:
				default:
				}
			}
			return nil
		},
	})

	select {
	case <-done:
		// success: triggered at least 3 times
	case <-time.After(2 * time.Second):
		t.Fatalf("循环任务触发次数不足，期望 >=3，实际 %d", triggerCount)
	}

	cancel()
	time.Sleep(50 * time.Millisecond)
}
