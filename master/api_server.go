package master

import (
	"context"
	"encoding/json"
	"fmt"
	"go-crontab/master/common"
	"net"
	"net/http"
	"runtime"
	"time"
)

var GAPIServer *apiServer

type apiServer struct {
	httpServer *http.Server
}

func InitDev() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func InitApiServer() error {
	mux := &http.ServeMux{}
	mux.HandleFunc("/job/create", createJob)

	server := &http.Server{
		Handler:           mux,
		ReadTimeout:       time.Duration(GConfig.ReadTimeOut) * time.Millisecond,
		WriteTimeout:      time.Duration(GConfig.WriteTimeOut) * time.Millisecond,
	}
	GAPIServer = &apiServer{
		httpServer: server,
	}
	if listener, err := net.Listen("tcp", fmt.Sprintf(":%d", GConfig.APIPort)); err != nil {
		return err
	} else {
		go  server.Serve(listener)
	}
	return nil
}

func createJob(w http.ResponseWriter, r *http.Request) {
	var (
		job *common.Job
		old *common.Job
		err error
	)
	if r.Body != nil {
		if err = json.NewDecoder(r.Body).Decode(&job); err !=nil {
			goto ERR
		}
	}
	if old, err = GJobManager.SaveJob(context.Background(), job); err != nil {
		goto ERR
	}
	w.Write(common.BuildResponse(0, "Success", old))
	return

ERR:
	w.Write(common.BuildResponse(-1, err.Error(), nil))
	w.WriteHeader(http.StatusForbidden)
}