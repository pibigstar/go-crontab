package worker

import (
	"context"
	"runtime"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gogf/gf/os/glog"

	"go-crontab/common"
)

var GJobManager *jobManager

type jobManager struct {
	cli     *clientv3.Client
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

func InitDev() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func InitJobManager() error {
	cli, err := clientv3.New(clientv3.Config{
		// 集群列表
		Endpoints:   []string{"106.54.212.69:2379"},
		DialTimeout: 5000 * time.Millisecond,
	})
	if err != nil {
		return err
	}
	watcher := clientv3.NewWatcher(cli)
	lease := clientv3.NewLease(cli)
	GJobManager = &jobManager{
		cli:     cli,
		lease:   lease,
		watcher: watcher,
	}

	response, err := cli.Get(context.TODO(), common.EtcdJobPrefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, kv := range response.Kvs {
		job, err := common.UnPackJob(kv.Value)
		if err != nil {
			glog.Printf("err: %s", err.Error())
			continue
		}
		jobEvent := &common.JobEvent{
			Job:       job,
			EventType: common.UpdateJob,
		}
		GScheduler.PushJobEvent(jobEvent)
	}

	revision := response.Header.Revision + 1
	// 监听 /cron/job/ 目录
	go GJobManager.WatchJobEvent(revision)

	// 监听 /cron/kill/ 目录
	go GJobManager.WatchKiller()

	return nil
}

// 监听任务列表
func (j *jobManager) WatchJobEvent(revision int64) {
	watchChan := j.watcher.Watch(context.TODO(), common.EtcdJobPrefix, clientv3.WithRev(revision), clientv3.WithPrefix())
	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			var (
				jobEvent = &common.JobEvent{}
				job      *common.Job
				err      error
			)
			switch event.Type {
			case mvccpb.PUT:
				job, err = common.UnPackJob(event.Kv.Value)
				if err != nil {
					glog.Errorf("unpack event value, err: %s", err.Error())
				}
				jobEvent.EventType = common.UpdateJob
			case mvccpb.DELETE:
				job = &common.Job{
					Name: string(event.Kv.Key),
				}
				jobEvent.EventType = common.DeleteJob
			}
			jobEvent.Job = job
			GScheduler.PushJobEvent(jobEvent)
		}
	}
}

// 监听任务强杀列表
func (j *jobManager) WatchKiller() {
	watchChan := j.watcher.Watch(context.TODO(), common.EtcdKillJobPrefix, clientv3.WithPrefix())
	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			var (
				jobEvent = &common.JobEvent{}
				job      *common.Job
			)
			switch event.Type {
			case mvccpb.PUT:
				jobName := strings.TrimPrefix(string(event.Kv.Key), common.EtcdKillJobPrefix)
				job = &common.Job{
					Name: jobName,
				}
			}
			jobEvent.Job = job
			jobEvent.EventType = common.KillJob
			GScheduler.PushJobEvent(jobEvent)
		}
	}
}

func (j *jobManager) CreateLocker(jobName string) *jobLocker {
	return &jobLocker{
		JobName: jobName,
		lease:   j.lease,
	}
}
