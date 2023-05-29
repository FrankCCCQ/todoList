package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strings"
	"time"
)

type Register struct {
	EtcdAddrs   []string
	DialTimeout int
	closeCh     chan struct{}
	leasesID    clientv3.LeaseID
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse
	srvInfo     Server
	srvTTL      int64
	cli         *clientv3.Client
	logger      *logrus.Logger
}

// NewRegister 基于ETCD 创建一个register
func NewRegister(etcdAddrs []string, logger *logrus.Logger) *Register {
	return &Register{
		EtcdAddrs:   etcdAddrs,
		DialTimeout: 3,
		logger:      logger,
	}
}

// Register 初始化自己的实例
func (r *Register) Register(srvInfo Server, ttl int64) (chan<- struct{}, error) {
	var err error
	if strings.Split(srvInfo.Addr, ":")[0] == "" {
		return nil, errors.New("invalid ip address")
	}
	// 初始化
	if r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	}); err != nil {
		return nil, err
	}
	r.srvInfo = srvInfo
	r.srvTTL = ttl
	if err = r.register(); err != nil {
		return nil, err
	}
	r.closeCh = make(chan struct{})
	go r.keepAlive()
	return r.closeCh, nil
}

// register 新建etcd 自带的实例
func (r *Register) register() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.DialTimeout)*time.Second)
	defer cancel()

	// etcd 的租约
	leaseResp, err := r.cli.Grant(ctx, r.srvTTL)
	if err != nil {
		return err
	}
	r.leasesID = leaseResp.ID
	if r.keepAliveCh, err = r.cli.KeepAlive(context.Background(), r.leasesID); err != nil {
		return err
	}
	data, err := json.Marshal(r.srvInfo)
	if err != nil {
		return err
	}
	// 把注册的实例 put到etcd上
	_, err = r.cli.Put(context.Background(), BuildRegisterPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID))
	return err

}

func (r *Register) keepAlive() {
	ticker := time.NewTicker(time.Duration(r.srvTTL) * time.Second)
	for {
		select {
		case <-r.closeCh:
			// 服务注销
			if err := r.unregister(); err != nil {
				fmt.Println("unregister failed error", err)
			}
			// 吊销leaseID 租赁凭证
			if _, err := r.cli.Revoke(context.Background(), r.leasesID); err != nil {
				fmt.Println("revoke fail")
			}
		// 查看 租赁凭证是否alive
		case res := <-r.keepAliveCh:
			// 重新租赁
			if res == nil {
				if err := r.register(); err != nil {
					fmt.Println("register error")
				}
			}
			// 超时器，每一轮超时可能是由租赁凭证到期导致
		case <-ticker.C:
			if r.keepAliveCh == nil {
				if err := r.register(); err != nil {
					fmt.Println("register error")
				}
			}
		}
	}
}

// unregister 在etcd中把这个服务删掉
func (r *Register) unregister() error {
	_, err := r.cli.Delete(context.Background(), BuildRegisterPath(r.srvInfo))
	return err
}
