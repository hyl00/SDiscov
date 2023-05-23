package register

import (
	"context"
	"fmt"

	"github.com/hyl00/SDiscov/etcdclient"

	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const timeToLive int64 = 10

// TimeToLive is seconds to live in etcd.
var TimeToLive = timeToLive

type (
	publisher struct {
		client        *clientv3.Client
		leaseID       clientv3.LeaseID
		keepaliveChan <-chan *clientv3.LeaseKeepAliveResponse
		key           string
		value         string
		ctx           context.Context
		stop          chan struct{}
	}
)

func NewPublisher(ctx context.Context, endpoints []string, username, password, key, value string) (*publisher, error) {
	client, err := etcdclient.NewEtcdClient(endpoints, username, password)
	if err != nil {
		return nil, err
	}
	p := &publisher{
		client: client,
		key:    key,
		value:  value,
		ctx:    ctx,
		stop:   make(chan struct{}),
	}
	return p, err
}

func (p *publisher) Start() error {
	err := p.register()
	if err != nil {
		return err
	}
	return p.keepaliveAsync()
}

//register put key、value to etcd with 10s timeout
func (p *publisher) register() error {
	resp, err := p.client.Grant(p.ctx, timeToLive)
	if err != nil {
		return err
	}
	p.leaseID = resp.ID
	key := makeEtcdKey(p.key, int64(p.leaseID))
	_, err = p.client.Put(p.ctx, key, p.value, clientv3.WithLease(p.leaseID))
	return err
}

func (p *publisher) keepaliveAsync() error {
	ch, err := p.client.Lease.KeepAlive(p.ctx, p.leaseID)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case _, ok := <-ch:
				if !ok { //超期续约重新注册
					_, err = p.client.Revoke(p.ctx, p.leaseID)
					if err != nil {
						logx.Errorf("etcd publisher revoke: %s", err.Error())
					}
					err = p.register()
					if err != nil {
						logx.Errorf("etcd publisher register: %s", err.Error())
					}
				}
			case <-p.stop:
				p.free()
				return
			}
		}
	}()
	return nil
}

func (p *publisher) Stop() {

	p.stop <- struct{}{}

}

func (p *publisher) free() {
	err := p.client.Close()
	if err != nil {
		logx.Errorf("etcd publisher close: %s", err.Error())
	}
	close(p.stop)
	_, err = p.client.Revoke(p.ctx, p.leaseID)
	if err != nil {
		logx.Errorf("etcd publisher revoke: %s", err.Error())
	}
}

func makeEtcdKey(key string, id int64) string {
	return fmt.Sprintf("%s/%d", key, id)
}
