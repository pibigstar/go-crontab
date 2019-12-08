package test

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"testing"
	"time"
)

// 分布式锁
func TestLock(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		// 集群列表
		Endpoints:   []string{"106.54.212.69:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Error(err)
	}
	defer cli.Close()

	var (
		ctx            = context.TODO()
		cancelFunc     context.CancelFunc
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		txn            clientv3.Txn
		txnResp *clientv3.TxnResponse
		key            = "/job/lock/job1"

	)

	ctx, cancelFunc = context.WithCancel(ctx)
	// 1. 申请一个租约
	lease := clientv3.NewLease(cli)
	if leaseGrantResp, err = lease.Grant(ctx, 5); err !=nil {
		t.Error(err)
	}
	leaseId = leaseGrantResp.ID

	// 取消自动续约
	defer cancelFunc()
	// 将租约设置为失效
	defer lease.Revoke(ctx, leaseId)

	// 2. 自动续租
	go func() {
		for {
			if keepRespChan, err = lease.KeepAlive(ctx, leaseId); err !=nil {
				t.Error(err)
			}
			select {
			case resp := <-keepRespChan:
				if resp == nil{
					t.Log("租约已失效")
					goto END
				} else {
					t.Log("自动续租:", resp.ID)
				}
			}
		}
		END:
	}()

	// 3. 开启事务
	txn = cli.Txn(ctx)
	// 如果createRevision为0，那么证明这个key还没被设置过值
	txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key,"pibigstar", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(key))

	// 提交事务
	if txnResp, err = txn.Commit(); err !=nil {
		t.Error(err)
	}
	// 没有提交成功，进入了then方法
	if !txnResp.Succeeded {
		t.Log("强占锁失败，锁已被占用: ", txnResp.OpResponse().Get().Kvs[0].Value)
		return
	}

	// 3. 处理业务逻辑
	t.Log("开始处理业务.....")
	time.Sleep(time.Second * 5)
}
