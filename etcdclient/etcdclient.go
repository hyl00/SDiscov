package etcdclient

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func NewEtcdClient(endpoints []string, username, password string) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:            endpoints,
		AutoSyncInterval:     0,
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    0,
		DialKeepAliveTimeout: 0,
		MaxCallSendMsgSize:   0,
		MaxCallRecvMsgSize:   0,
		TLS:                  nil,
		Username:             username,
		Password:             password,
		RejectOldCluster:     false,
		DialOptions:          nil,
		Context:              nil,
		Logger:               nil,
		LogConfig:            nil,
		PermitWithoutStream:  false,
	})
}
