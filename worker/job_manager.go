package worker

import (
	"context"
	"runtime"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gogf/gf/os/glog"

	"go-crontab/common"
)

var GJobManager *jobManager

type jobManager struct {
	cli     *clientv3.Client
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

	GJobManager = &jobManager{
		cli:     cli,
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
			EventType: mvccpb.PUT,
		}
		GScheduler.PushJobEvent(jobEvent)
	}

	revision := response.Header.Revision + 1
	// 监听
	go GJobManager.WatchJobEvent(revision)

	return nil
}

func (j *jobManager) WatchJobEvent(revision int64) {
	watchChan := j.watcher.Watch(context.TODO(), common.EtcdJobPrefix, clientv3.WithRev(revision), clientv3.WithPrefix())
	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			var (
				job *common.Job
				err error
			)
			switch event.Type {
			case mvccpb.PUT:
				job, err = common.UnPackJob(event.Kv.Value)
				if err != nil {
					glog.Errorf("unpack event value, err: %s", err.Error())
				}
			case mvccpb.DELETE:
				job = &common.Job{
					Name: string(event.Kv.Key),
				}
			}

			jobEvent := &common.JobEvent{
				EventType: event.Type,
				Job:       job,
			}
			GScheduler.PushJobEvent(jobEvent)
		}
	}
}
