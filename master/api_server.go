package master

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"

	"go-crontab/common"
)

func InitDev() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func InitApiServer() {
	s := g.Server()

	s.SetPort(GConfig.APIPort)
	s.SetLogPath("log")
	s.EnablePProf()
	s.SetReadTimeout(time.Duration(GConfig.ReadTimeOut) * time.Millisecond)
	s.SetWriteTimeout(time.Duration(GConfig.WriteTimeOut) * time.Millisecond)

	s.BindHandler("/job/create", createJob)
	s.BindHandler("/job/delete", deleteJob)
	s.BindHandler("/job/list", listJobs)
	s.BindHandler("/job/kill", killJob)

	go s.Run()
}

func createJob(r *ghttp.Request) {
	var (
		job *common.Job
		old *common.Job
		err error
	)
	if r.Body != nil {
		if err = json.NewDecoder(r.Body).Decode(&job); err != nil {
			goto ERR
		}
	}
	if old, err = GJobManager.SaveJob(context.Background(), job); err != nil {
		goto ERR
	}
	r.Response.Write(common.BuildResponse(0, "Success", old))
	return

ERR:
	r.Response.Write(common.BuildResponse(-1, err.Error(), nil))
	r.Response.WriteHeader(http.StatusForbidden)
}

func deleteJob(r *ghttp.Request) {
	var (
		job *common.Job
		old *common.Job
		err error
	)
	if r.Body != nil {
		if err = json.NewDecoder(r.Body).Decode(&job); err != nil {
			goto ERR
		}
	}
	if old, err = GJobManager.DeleteJob(context.Background(), job); err != nil {
		goto ERR
	}
	r.Response.Write(common.BuildResponse(0, "Success", old))
	return

ERR:
	r.Response.Write(common.BuildResponse(-1, err.Error(), nil))
	r.Response.WriteHeader(http.StatusForbidden)
}

func listJobs(r *ghttp.Request) {
	var (
		jobs []*common.Job
		err  error
	)
	jobs, err = GJobManager.ListJobs(context.Background())
	if err != nil {
		goto ERR
	}

	r.Response.Write(common.BuildResponse(0, "Success", jobs))
	return

ERR:
	r.Response.Write(common.BuildResponse(-1, err.Error(), nil))
	r.Response.WriteHeader(http.StatusForbidden)
}

func killJob(r *ghttp.Request) {
	var (
		job *common.Job
		err error
	)
	if r.Body != nil {
		if err = json.NewDecoder(r.Body).Decode(&job); err != nil {
			goto ERR
		}
	}
	err = GJobManager.SaveJobWithLease(context.Background(), job)
	if err != nil {
		goto ERR
	}
	r.Response.Write(common.BuildResponse(0, "Success", nil))
	return

ERR:
	r.Response.Write(common.BuildResponse(-1, err.Error(), nil))
	r.Response.WriteHeader(http.StatusForbidden)
}
