package master

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"go-crontab/master/common"
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
	return nil
}

func (j *jobManager) SaveJob(ctx context.Context, job *common.Job) (old *common.Job,err error) {
	var (
		putResponse *clientv3.PutResponse
	)
	bs, _ := json.Marshal(job)
	if putResponse, err = j.cli.Put(ctx, common.BuildJobName(job), string(bs), clientv3.WithPrevKV()); err !=nil {
		return
	}
	// 更新操作，返回旧值
	if putResponse.PrevKv != nil {
		 _ = json.Unmarshal(putResponse.PrevKv.Value, &old)
	}
	return
}