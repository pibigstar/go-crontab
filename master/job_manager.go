package master

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/gogf/gf/os/glog"
	"go-crontab/master/common"
	"log"
	"time"
)
var GJobManager *jobManager

type jobManager struct {
	cli *clientv3.Client
	lease clientv3.Lease
}

func InitJobManager()  error {
	cli, err := clientv3.New(clientv3.Config{
		// 集群列表
		Endpoints:   []string{"106.54.212.69:2379"},
		DialTimeout: 5000 * time.Millisecond,
	})
	if err != nil {
		return err
	}
	lease := clientv3.NewLease(cli)
	GJobManager = &jobManager{
		cli: cli,
		lease: lease,
	}

	// 监听kill
	go func() {
		for {
			watch := GJobManager.cli.Watch(context.TODO(), common.EtcdKillJobPrefix, clientv3.WithPrefix())
			resp := <- watch
			glog.Printf("resp: %+v \n", resp)
		}
	}()

	return nil
}

func (j *jobManager) SaveJob(ctx context.Context, job *common.Job) (*common.Job,error) {
	bs, _ := json.Marshal(job)
	putResp, err := j.cli.Put(ctx, common.BuildJobName(job), string(bs), clientv3.WithPrevKV())
	if err !=nil {
		return nil, err
	}
	var old *common.Job
	// 更新操作，返回旧值
	if putResp.PrevKv != nil {
		 _ = json.Unmarshal(putResp.PrevKv.Value, &old)
	}
	return old, nil
}

func (j *jobManager) DeleteJob(ctx context.Context, job *common.Job) (*common.Job, error) {

	deleteResp, err := j.cli.Delete(ctx, common.BuildJobName(job), clientv3.WithPrevKV())
	if err !=nil {
		return nil, err
	}
	// 删除操作，返回旧值
	var old *common.Job
	if deleteResp.PrevKvs != nil {
		_ = json.Unmarshal(deleteResp.PrevKvs[0].Value, &old)
	}
	return old, nil
}

func (j *jobManager) ListJobs(ctx context.Context) ([]*common.Job, error) {
	getResp, err := j.cli.Get(ctx, common.EtcdJobPrefix, clientv3.WithPrefix())
	if err !=nil {
		return nil, err
	}
	var jobs []*common.Job
	for _, kv := range getResp.Kvs {
		job := &common.Job{}
		err := json.Unmarshal(kv.Value, &job)
		if err != nil {
			log.Print(err.Error(), string(kv.Value))
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (j *jobManager) SaveJobWithLease(ctx context.Context, job *common.Job) error {
	// 申请一个租约
	leaseResp, err := j.lease.Grant(ctx, 1)
	if err != nil {
		return err
	}

	_, err = j.cli.Put(ctx, common.BuildKillJobName(job), "kill", clientv3.WithLease(leaseResp.ID))
	if err !=nil {
		return  err
	}

	return nil
}