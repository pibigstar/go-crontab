package worker

import (
	"go-crontab/common"
	"math/rand"
	"os/exec"
	"time"
)

var GExecutor *executor

type executor struct {
}

func (e *executor) ExecuteJob(jobInfo *common.JobExecuteInfo) {
	go func() {
		result := &common.JobExecuteResult{
			Job:       jobInfo.Job,
			StartTime: time.Now(),
		}
		// 随机睡眠 0~1s,防止某个CPU一直抢占到锁
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		locker := GJobManager.CreateLocker(jobInfo.Job.Name)
		err := locker.TryLock()
		if err != nil {
			result.Err = err
			result.EndTime = time.Now()
			// 将执行结果放入通道
			GScheduler.PushJobResult(result)
			return
		}
		defer locker.UnLock()

		// 执行任务
		cmd := exec.CommandContext(jobInfo.Ctx, "bash.exe", "-c", jobInfo.Job.Command)
		// 输出执行结果
		output, err := cmd.CombinedOutput()

		result.OutPut = output
		result.Err = err
		result.EndTime = time.Now()
		// 将执行结果放入通道
		GScheduler.PushJobResult(result)
	}()
}
