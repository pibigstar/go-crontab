package worker

import (
	"context"
	"errors"
	"github.com/coreos/etcd/clientv3"
	"go-crontab/common"
	"runtime"
)

type jobLocker struct {
	JobName    string
	lease      clientv3.Lease
	cancelFunc context.CancelFunc
	leaseId    clientv3.LeaseID
	isLock     bool
}

func (j *jobLocker) TryLock() error {
	leaseGrantResp, err := j.lease.Grant(context.Background(), 5)
	if err != nil {
		return err
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	j.cancelFunc = cancelFunc
	j.leaseId = leaseGrantResp.ID
	defer cancelFunc()

	keepAliveResp, err := j.lease.KeepAlive(ctx, leaseGrantResp.ID)
	if err != nil {
		return err
	}
	// 续租
	go func() {
		for {
			select {
			case resp := <-keepAliveResp:
				if resp == nil {
					runtime.Goexit()
				}
			}
		}
	}()
	// 创建事务开始抢锁
	txn := GJobManager.cli.Txn(context.Background())

	lockKey := common.EtcdLockJobPrefix + j.JobName
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseGrantResp.ID))).
		Else(clientv3.OpGet(lockKey))
	txnResponse, err := txn.Commit()
	if err != nil {
		return err
	}

	if !txnResponse.Succeeded {
		return errors.New("failed to get lock")
	}
	j.isLock = true
	return nil
}

func (j *jobLocker) UnLock() {
	if j.isLock {
		j.cancelFunc()
		j.lease.Revoke(context.Background(), j.leaseId)
	}
}
