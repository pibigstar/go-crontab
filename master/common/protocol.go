package common

import (
	"encoding/json"
	"fmt"
)

const (
	EtcdJobPrefix = "/cron/job/"
	EtcdKillJobPrefix = "/cron/kill/"
)

type Job struct {
	Name string `json:"name"`
	Command string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

func BuildJobName(job *Job) string {
	return fmt.Sprintf(EtcdJobPrefix + "%s",job.Name)
}
func BuildKillJobName(job *Job) string {
	return fmt.Sprintf(EtcdKillJobPrefix + "%s",job.Name)
}


type Response struct {
	Code int `json:"code"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

func BuildResponse(code int, msg string, data interface{}) []byte {
	resp := &Response{
		Code: code,
		Msg:  msg,
		Data: data,
	}
	bs, _ := json.Marshal(resp)
	return bs
}
