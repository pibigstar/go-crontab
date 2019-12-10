package main

import "go-crontab/master"

func main() {
	var (
		err error
	)
	// 初始化环境
	master.InitDev()

	// 启动APIServer，提供http服务
	if err = master.InitApiServer(); err !=nil {
		panic(err)
	}
}