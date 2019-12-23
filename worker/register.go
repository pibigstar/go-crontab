package worker

import (
	"context"
	"errors"
	"go-crontab/common"
	"go.etcd.io/etcd/clientv3"
	"net"
	"time"
)

type Register struct {
	client *clientv3.Client
	lease  clientv3.Lease
	ip     string
}

func InitRegister() error {
	cli, err := clientv3.New(clientv3.Config{
		// 集群列表
		Endpoints:   []string{"106.54.212.69:2379"},
		DialTimeout: 5000 * time.Millisecond,
	})
	if err != nil {
		return err
	}
	lease := clientv3.NewLease(cli)

	ip, err := getIp()
	if err != nil {
		return err
	}
	register := &Register{
		client: cli,
		lease:  lease,
		ip:     ip,
	}

	go register.KeepAlive()

	return nil
}

func getIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		// 取第一个不为回环地址的 ipv4
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			ipv4 := ipNet.IP.To4()
			if ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}
	return "", errors.New("not found ip")
}

func (r *Register) KeepAlive() {
	var (
		keepAliveResp <-chan *clientv3.LeaseKeepAliveResponse
	)
	leaseResp, err := r.lease.Grant(context.Background(), 10)
	if err != nil {
		return
	}
	key := common.EtcdWorkerPrefix + r.ip
	for {
		ctx, cancel := context.WithCancel(context.Background())
		_, err := r.client.Put(ctx, key, "", clientv3.WithLease(leaseResp.ID))
		if err != nil {
			goto RETRY
		}

		keepAliveResp, err = r.lease.KeepAlive(ctx, leaseResp.ID)
		if err != nil {
			goto RETRY
		}
		for {
			select {
			case resp := <-keepAliveResp:
				if resp == nil {
					goto RETRY
				}
			}
		}
	RETRY:
		time.Sleep(1 * time.Second)
		if cancel != nil {
			cancel()
		}
	}
}
