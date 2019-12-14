package worker

import (
	"context"
	"go-crontab/common"
	"os/exec"
	"time"
)

var GExecutor *executor

type executor struct {
}

func (e *executor) ExecuteJob(ctx context.Context, jobPlan *common.JobSchedulePlan) {
	go func() {

		result := &common.JobExecuteResult{
			JobPlan:   jobPlan,
			StartTime: time.Now(),
		}
		// 执行任务
		cmd := exec.CommandContext(ctx, "bash.exe", "-c", jobPlan.Job.Command)
		// 输出执行结果
		output, err := cmd.CombinedOutput()

		result.OutPut = output
		result.Err = err
		result.EndTime = time.Now()
		// 将执行结果放入通道
		GScheduler.PushJobResult(result)
	}()
}
