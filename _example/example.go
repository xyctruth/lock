package main

import (
	"fmt"
	"github.com/xyctruth/lock/drive/drive_etcd"
	etcdClient "go.etcd.io/etcd/client/v3"
	"os"
)

var (
	Endpoint string
	UserName string
	Password string
)

func init() {
	Endpoint = os.Getenv("Endpoint")
	if Endpoint == "" {
		panic("Endpoint not found")
	}
	UserName = os.Getenv("UserName")
	if UserName == "" {
		panic("UserName not found")
	}
	Password = os.Getenv("Password")
	if Password == "" {
		panic("Password not found")
	}

}

func main() {
	locker, _ := drive_etcd.NewEtcdLocker(
		etcdClient.Config{
			Endpoints: []string{Endpoint},
			Username:  UserName,
			Password:  Password,
		},
	)
	lock1, err := locker.Acquire("key")
	defer lock1.Release()
	if err != nil {
		fmt.Println(err)
	}
	// Do something
}
