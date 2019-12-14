package main

import (
	"os"
	"os/signal"
	"syscall"

	"go-crontab/woker"
)

func main() {
	var err error
	// 初始化环境
	woker.InitDev()

	// 初始化调度器
	woker.InitScheduler()

	// 初始化JobManager
	if err = woker.InitJobManager(); err != nil {
		panic(err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-interrupt
}
