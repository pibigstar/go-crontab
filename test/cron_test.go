package test

import (
	"github.com/gorhill/cronexpr"
	"testing"
	"time"
)

func TestCron(t *testing.T) {
	var (
		expression *cronexpr.Expression
		err error
	)
	cron := "*/2 * * * * * *"
	if expression, err = cronexpr.Parse(cron); err != nil {
		t.Error(err)
	}
	now := time.Now()
	next := expression.Next(now)

	time.AfterFunc(next.Sub(now), func() {
		t.Log("执行时间:", time.Now())
	})

	time.Sleep(time.Second * 5)
}