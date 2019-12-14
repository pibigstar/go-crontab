package test

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func TestCmd(t *testing.T) {
	var (
		cmd    *exec.Cmd
		output []byte
		err    error
	)
	cmd = exec.Command("bash.exe", "-c", "ls -l")

	if output, err = cmd.CombinedOutput(); err != nil {
		t.Error(err)
	}
	t.Log(string(output))
}

type result struct {
	err    error
	output []byte
}

func TestCmdWithContext(t *testing.T) {
	var (
		cmd        *exec.Cmd
		output     []byte
		err        error
		cancel     context.CancelFunc
		ctx        context.Context
		resultChan chan *result
	)

	resultChan = make(chan *result, 1)

	ctx, cancel = context.WithCancel(context.TODO())
	go func() {
		cmd = exec.CommandContext(ctx, "bash.exe", "-c", "sleep 2;ls -l")
		output, err = cmd.CombinedOutput()
		resultChan <- &result{
			err:    err,
			output: output,
		}
	}()

	time.Sleep(time.Second * 1)
	// 取消任务
	cancel()

	result := <-resultChan

	t.Log(result.err, string(result.output))
}
