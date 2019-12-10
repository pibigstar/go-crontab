package master

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"
)

type APIServer struct {
	httpServer *http.Server
}

func InitDev() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func InitApiServer() (err error){
	mux := &http.ServeMux{}
	mux.HandleFunc("/job/create", createJob)

	config := parseConfig()
	server := &http.Server{
		Handler:           mux,
		ReadTimeout:       time.Duration(config.ReadTimeOut) * time.Millisecond,
		WriteTimeout:      time.Duration(config.WriteTimeOut) * time.Millisecond,
	}
	var listener net.Listener
	if listener, err = net.Listen("tcp", fmt.Sprintf(":%d", config.APIPort)); err != nil {
		return err
	}
	if err = server.Serve(listener); err != nil {
		return err
	}
	return
}

func createJob(w http.ResponseWriter, r *http.Request) {
	fmt.Println("创建任务")
}