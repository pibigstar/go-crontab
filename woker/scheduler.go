package woker

import (
	"fmt"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"

	"go-crontab/common"
)

var GScheduler *scheduler

// 调度器
type scheduler struct {
	JobEventChan chan *common.JobEvent
	JobPlanTable map[string]*common.JobSchedulePlan
}

// 初始化调度器
func InitScheduler() {
	GScheduler = &scheduler{
		JobEventChan: make(chan *common.JobEvent, 1000),
		JobPlanTable: make(map[string]*common.JobSchedulePlan),
	}
	go GScheduler.ScheduleLoop()
}

func (s *scheduler) PushJobEvent(event *common.JobEvent) {
	s.JobEventChan <- event
}

// 调度
func (s *scheduler) ScheduleLoop() {

	sleep := s.TrySchedule()
	timer := time.NewTimer(sleep)

	for {
		select {
		case event := <-s.JobEventChan:
			s.HandleEvent(event)
		case <-timer.C:

		}
		sleep := s.TrySchedule()
		timer.Reset(sleep)
	}
}

// 处理事件
func (s *scheduler) HandleEvent(event *common.JobEvent) {
	switch event.EventType {
	case mvccpb.PUT:
		// 更新任务
		schedulePlan, err := common.BuildJobPlan(event)
		if err != nil {
			glog.Errorf("build job plan, err: %s", err.Error())
			return
		}
		s.JobPlanTable[event.Job.Name] = schedulePlan
	case mvccpb.DELETE:
		// 删除任务
		if _, ok := s.JobPlanTable[event.Job.Name]; ok {
			delete(s.JobPlanTable, event.Job.Name)
		}
	}
}

// 执行任务
func (s *scheduler) TrySchedule() time.Duration {
	var (
		now      = time.Now()
		nearTime time.Time
	)
	if len(s.JobPlanTable) == 0 {
		return time.Second * 1
	}

	for _, jobPlan := range s.JobPlanTable {
		// 任务执行时间到了，执行任务
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			fmt.Println("执行任务:", jobPlan.Job.Name)
			// 更新下次执行时间
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}

		// 获取下次执行时间最近的任务
		if nearTime.IsZero() || jobPlan.NextTime.Before(nearTime) {
			nearTime = jobPlan.NextTime
		}
	}
	sleep := nearTime.Sub(now)
	return sleep
}
