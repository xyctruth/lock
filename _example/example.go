package main

import (
	"fmt"
	"github.com/xyctruth/lock/drive/drive_etcd"
	etcdClient "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"os"
	"time"
)

var (
	DefaultEndpoint = "http://0.0.0.0:2379"
	DefaultUserName = ""
	DefaultPassword = ""
)

func init() {
	endpoint := os.Getenv("ETCD_ENDPOINT")
	if endpoint != "" {
		DefaultEndpoint = endpoint
	}
	userName := os.Getenv("ETCD_USERNAME")
	if userName != "" {
		DefaultUserName = userName
	}
	password := os.Getenv("ETCD_PASSWORD")
	if userName != "" {
		DefaultPassword = password
	}
}

func main() {
	locker, err := drive_etcd.NewEtcdLocker(
		etcdClient.Config{
			Endpoints:   []string{DefaultEndpoint},
			Username:    DefaultUserName,
			Password:    DefaultPassword,
			DialTimeout: 3 * time.Second,
			DialOptions: []grpc.DialOption{grpc.WithBlock()},
		},
	)
	if err != nil {
		panic(err)
	}

	lock1, err := locker.Acquire("key")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer lock1.Release()

	// Do something
}
