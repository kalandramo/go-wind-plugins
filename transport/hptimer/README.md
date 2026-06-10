# High Precision Timer (hptimer)

基于最小堆 + 单一 `time.Timer` 的高精度定时任务服务器，实现了 `go-wind/transport.Server` 接口，支持毫秒级定时精度，可以与 `go-wind` 应用框架无缝集成。

## 核心特性

- **毫秒级精度**：基于最小堆调度，远超传统 cron 的秒级精度
- **单次 / 循环 / Cron**：支持绝对时间触发、固定间隔循环、cron 表达式三种模式
- **极低资源消耗**：单一 goroutine + 单一 timer，任务数量不影响资源占用
- **高并发安全**：支持高并发 Add/Remove，互斥锁保护堆操作
- **任务优先级**：支持 PriorityHigh / PriorityMedium / PriorityLow 三级优先级
- **任务取消**：每个任务持有独立 context，可随时取消
- **观察者模式**：通过 `TimerObserver` 接口解耦任务触发事件
- **优雅关闭**：等待任务执行完毕后再退出
- **阻塞式生命周期**：`Start` 阻塞直到 context 取消，兼容 `go-wind` App

## 安装

```bash
go get github.com/tx7do/go-wind-plugins/transport/hptimer
```

## 快速开始

```go
package main

import (
    "context"
    "log"
    "time"

    hptimer "github.com/tx7do/go-wind-plugins/transport/hptimer"
)

func main() {
    srv := hptimer.NewServer()

    // 启动服务器（阻塞）
    go func() {
        ctx := context.Background()
        if err := srv.Start(ctx); err != nil {
            panic(err)
        }
    }()

    // 等待引擎启动
    time.Sleep(100 * time.Millisecond)

    // 单次任务：5 秒后触发
    srv.AddTask(hptimer.NewTimerTask("task1",
        time.Now().Add(5*time.Second),
        hptimer.WithCallback(func(ctx context.Context) error {
            log.Println("task1 fired!")
            return nil
        }),
    ))

    // 循环任务：每 2 秒执行一次
    srv.AddTask(hptimer.NewTimerTask("task2",
        time.Now().Add(2*time.Second),
        hptimer.WithInterval(2*time.Second),
        hptimer.WithCallback(func(ctx context.Context) error {
            log.Println("task2 interval fired!")
            return nil
        }),
    ))

    // Cron 任务：每分钟执行
    srv.AddTask(hptimer.NewTimerTask("task3",
        time.Time{}, // At 留空，由 Cron 计算
        hptimer.WithCron("*/1 * * * *"),
        hptimer.WithCallback(func(ctx context.Context) error {
            log.Println("task3 cron fired!")
            return nil
        }),
    ))

    select {}
}
```

## 配置选项

| 选项                               | 说明              | 默认值    |
|----------------------------------|-----------------|--------|
| `WithGracefullyShutdown(enable)` | 是否等待运行中任务完成后再退出 | `true` |
| `WithTimerObserver(observer)`    | 自定义任务触发观察者      | 内置回调机制 |

## 任务创建

使用 `NewTimerTask` 函数创建任务，通过可选参数配置：

```go
task := hptimer.NewTimerTask(
    "my-task",                          // 任务 ID（唯一标识）
    time.Now().Add(10*time.Second),     // 首次触发时间
    hptimer.WithInterval(5*time.Second), // 可选：循环间隔
    hptimer.WithCron("0 */5 * * * * *"), // 可选：cron 表达式（与 Interval 二选一）
    hptimer.WithCallback(func(ctx context.Context) error {
        log.Println("task triggered")
        return nil
    }),
    hptimer.WithData(myPayload),          // 可选：任意负载
    hptimer.WithPriority(hptimer.PriorityHigh), // 可选：优先级
    hptimer.WithContext(parentCtx),       // 可选：父 context
)
```

### 任务选项

| 选项                    | 说明                          |
|-----------------------|-----------------------------|
| `WithInterval(d)`     | 设置循环间隔（0 表示单次任务）            |
| `WithCron(expr)`      | 设置 cron 表达式（与 Interval 二选一） |
| `WithCallback(fn)`    | 设置任务回调函数                    |
| `WithData(v)`         | 设置任意负载                      |
| `WithPriority(p)`     | 设置优先级（High / Medium / Low）  |
| `WithContext(parent)` | 设置父 context                 |

## 任务管理

```go
// 添加任务
taskID := srv.AddTask(task)

// 移除任务
ok := srv.RemoveTask(taskID)
```

## 三种触发模式

| 模式      | 配置                               | 行为               |
|---------|----------------------------------|------------------|
| 单次触发    | `At` 非空，`Interval` = 0，`Cron` 为空 | 在指定时间触发一次        |
| 间隔循环    | `Interval` > 0                   | 每隔 Interval 重复触发 |
| Cron 循环 | `Cron` 非空                        | 按 cron 表达式重复触发   |

> `Interval` 和 `Cron` 同时设置时，`Interval` 优先。

## 与 Cron 对比

| 维度     | hptimer               | cron (robfig/cron) |
|--------|-----------------------|--------------------|
| 精度     | 毫秒级                   | 秒级                 |
| 高频任务   | 高效                    | 不适合                |
| 任务数量影响 | 几乎无                   | 任务多时调度压力增大         |
| 资源占用   | 1 timer + 1 goroutine | 随任务数线性增长           |
| 动态增删   | 高效                    | 频繁变更有锁竞争           |
| 触发模式   | 绝对时间 / 间隔 / Cron      | Cron 表达式           |
| 适用场景   | 高精度 / 高频 / 批量 / 动态    | 周期性 / 低频 / 业务型     |

## 基准测试

在 `Intel Core i7-14700HX` / `96G` / `Go 1.26` 环境下：

| Benchmark  | ns/op      | 说明            |
|------------|------------|---------------|
| SingleTask | ~1,579,091 | 单任务添加+触发      |
| BatchTasks | ~1,643,064 | 1000 任务批量添加触发 |

单任务和批量 1000 任务的调度全流程（添加、调度、触发、回调），平均每轮耗时约 1.5~1.6 ms。

## 参考资料

- [gorhill/cronexpr](https://github.com/gorhill/cronexpr)
- [container/heap - Go 标准库](https://pkg.go.dev/container/heap)
