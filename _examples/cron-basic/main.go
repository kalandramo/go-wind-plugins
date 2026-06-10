// Package main demonstrates a cron timer server using go-wind.
//
// This example registers several cron jobs with different schedules:
//   - A job running every 5 seconds (second-level expression)
//   - A job running every 10 seconds
//   - A job running every minute
//   - A job using the @every descriptor
//
// Run:
//
//	go run ./_examples/cron-basic
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	cronServer "github.com/tx7do/go-wind-plugins/transport/cron"
)

func main() {
	srv := cronServer.NewServer()

	// Register jobs BEFORE Start (recommended).
	// Every 5 seconds (second-level expression).
	_, _ = srv.NewTimerJob("*/5 * * * * *", func() {
		fmt.Printf("[tick] 5-second job at %s\n", time.Now().Format("15:04:05"))
	})

	// Every 10 seconds.
	_, _ = srv.NewTimerJob("*/10 * * * * *", func() {
		fmt.Printf("[tick] 10-second job at %s\n", time.Now().Format("15:04:05"))
	})

	// Every minute.
	_, _ = srv.NewTimerJob("0 */1 * * * *", func() {
		fmt.Printf("[tick] minute job at %s\n", time.Now().Format("15:04:05"))
	})

	// Descriptor: every 3 seconds.
	_, _ = srv.NewTimerJob("@every 3s", func() {
		fmt.Printf("[tick] @every 3s job at %s\n", time.Now().Format("15:04:05"))
	})

	fmt.Printf("cron server started with %d jobs\n", srv.GetJobCount())
	fmt.Println("Press Ctrl+C to stop")

	// Graceful shutdown on SIGINT / SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := srv.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("server stopped")
}
