package discovery

import (
	"context"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"time"
)

const schema = "etcd"

type Resolver struct {
	schema      string
	EtcdAddrs   []string
	DialTimeout int

	closech     chan struct{}
	watchCh     clientv3.WatchChan
	cli         *clientv3.Client
	keyPrefix   string
	srvAddrList []resolver.Address
	cc          resolver.ClientConn
	logger      *logrus.Logger
}

// NewResolver create a new resolver.Builder base on etcd
func NewResolver(etcdAddress []string, logger *logrus.Logger) *Resolver {
	return &Resolver{
		schema:      schema,
		EtcdAddrs:   etcdAddress,
		DialTimeout: 3,
		logger:      logger,
	}
}

// Schema returns the schema supprted by this reslover
func (r *Resolver) Scheme() string {
	return r.schema
}

func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.cc = cc
	r.keyPrefix = BuildPrefix(Server{Name: target.Endpoint(), Version: target.URL.Host})
	if _, err := r.start(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Resolver) start() (chan<- struct{}, error) {
	var err error
	r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	resolver.Register(r)
	r.closech = make(chan struct{})
	if err = r.sync(); err != nil {
		return nil, err
	}
	go r.watch()
	return r.closech, nil
}

// watch event update
func (r *Resolver) watch() {
	ticker := time.NewTicker(time.Minute)
	r.watchCh = r.cli.Watch(context.Background(), r.keyPrefix, clientv3.WithPrefix())

	for {
		select {
		case <-r.closech:
			return
		case res, ok := <-r.watchCh:
			if ok {
				r.update(res.Events)
			}
		case <-ticker.C:
			if err := r.sync(); err != nil {
				r.logger.Error("sync error", err)
			}
		}
	}

}

func (r *Resolver) update(events []*clientv3.Event) {
	for _, ev := range events {
		var info Server
		var err error
		switch ev.Type {
		case clientv3.EventTypePut:
			info, err = ParseValue(ev.Kv.Value)
			if err != nil {
				continue
			}
			addr := resolver.Address{Addr: info.Addr, Attributes: attributes.New("Weight", info.Weight)}
			if !Exist(r.srvAddrList, addr) {
				r.srvAddrList = append(r.srvAddrList, addr)
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrList})
			}
		case clientv3.EventTypeDelete:
			info, err = SplitPath(string(ev.Kv.Key))
			if err != nil {
				continue
			}
			addr := resolver.Address{Addr: info.Addr}
			if s, ok := Remove(r.srvAddrList, addr); ok {
				r.srvAddrList = s
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrList})
			}
		}
	}
}

// sync 同步获取所有地址的信息
func (r *Resolver) sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := r.cli.Get(ctx, r.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	r.srvAddrList = []resolver.Address{}
	for _, v := range res.Kvs {
		info, err := ParseValue(v.Value)
		if err != nil {
			continue
		}

		addr := resolver.Address{Addr: info.Addr, Attributes: attributes.New("Weight", info.Weight)}
		r.srvAddrList = append(r.srvAddrList, addr)
	}
	r.cc.UpdateState(resolver.State{
		Addresses: r.srvAddrList,
	})
	return nil
}

func (r *Resolver) Close() {
	r.closech <- struct{}{}
}

func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {}
