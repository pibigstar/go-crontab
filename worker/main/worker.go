package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-crontab/worker"
)

func main() {
	var err error
	// 初始化环境
	worker.InitDev()

	// 初始化调度器
	worker.InitScheduler()

	// 初始化JobManager
	if err = worker.InitJobManager(); err != nil {
		panic(err)
	}

	// 初始化LogManager
	if err = worker.InitLogManager(); err != nil {
		panic(err)
	}

	// 将节点注册到服务中心
	if err = worker.InitRegister(); err != nil {
		panic(err)
	}

	fmt.Println("worker start...")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-interrupt
}
