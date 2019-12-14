package main

import (
	"go-crontab/master"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var err error
	// 初始化环境
	master.InitDev()

	// 初始化JobManager
	if err = master.InitJobManager(); err != nil {
		panic(err)
	}

	// 启动APIServer，提供http服务
	master.InitApiServer()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-interrupt
}
