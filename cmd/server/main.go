package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"stock-lens/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	// 启动编排集中在 internal/app，main 只负责进程信号和退出码。
	return app.Run(ctx, args)
}
