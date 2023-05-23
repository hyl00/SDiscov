package discov

import (
	"context"
	"sync"
	"time"

	"github.com/hyl00/SDiscov/etcdclient"

	"github.com/zeromicro/go-zero/core/logx"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	coolDownInterval = time.Second
	requestTimeout   = 3 * time.Second
)

var RequestTimeout = requestTimeout

type discover struct {
	client     *clientv3.Client
	serverList map[string]map[string]string
	lock       sync.Mutex
	ctx        context.Context
	Change     chan struct{}
}

type kv struct {
	key   string
	value string
}

func NewDiscover(ctx context.Context, endpoints []string, username, password string) (*discover, error) {
	client, err := etcdclient.NewEtcdClient(endpoints, username, password)
	if err != nil {
		return nil, err
	}
	d := &discover{
		client:     client,
		serverList: make(map[string]map[string]string),
		lock:       sync.Mutex{},
		ctx:        ctx,
		Change:     make(chan struct{}),
	}
	return d, err
}

func (d *discover) Load(key string) {
	var resp *clientv3.GetResponse
	for {
		var err error
		ctx, cancel := context.WithTimeout(d.ctx, RequestTimeout)
		resp, err = d.client.Get(ctx, key, clientv3.WithPrefix())
		cancel()
		if err == nil { //保证etcd连接不上或者出现故障时，内存中的服务地址不会被更新
			break
		}
		logx.Error(err)
		time.Sleep(coolDownInterval)
	}
	var kvs []kv
	for _, ev := range resp.Kvs {
		kvs = append(kvs, kv{
			key:   string(ev.Key),
			value: string(ev.Value),
		})
	}
	d.handleChanges(key, kvs)
	go d.watch(key)
}

func (d *discover) watch(key string) {
	watchChan := d.client.Watch(d.ctx, key, clientv3.WithPrefix())
	for watch := range watchChan {
		var kvs []kv
		for _, ev := range watch.Events {
			kvs = append(kvs, kv{
				key:   string(ev.Kv.Key),
				value: string(ev.Kv.Value),
			})
		}
		d.handleChanges(key, kvs)
	}
}

func (d *discover) GetAddrs(key string) []string {
	addrs := make([]string, 0)
	for _, v := range d.serverList[key] {
		addrs = append(addrs, v)
	}
	return addrs
}

func (d *discover) handleChanges(key string, kvs []kv) {
	d.lock.Lock()
	defer d.lock.Unlock()
	vals, ok := d.serverList[key]
	if !ok {
		vals = make(map[string]string)
	}
	for _, v := range kvs {
		vals[v.key] = v.value
	}
	d.serverList[key] = vals
	d.Change <- struct{}{}
}

func (d *discover) Stop() error {
	return d.client.Close()
}
