package drive_etcd

import (
	"github.com/stretchr/testify/require"
	"github.com/xyctruth/lock"
	etcdClient "go.etcd.io/etcd/client/v3"
	"os"
	"testing"
	"time"
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

func TestLock(t *testing.T) {
	locker, err := NewEtcdLocker(
		etcdClient.Config{
			Endpoints: []string{Endpoint},
			Username:  UserName,
			Password:  Password,
		},
	)
	key := "TestLockKey"

	lock1, err := locker.Acquire(key)
	require.NoError(t, err)
	defer func() {
		err = lock1.Release()
		require.NoError(t, err)
	}()

	_, err = locker.Acquire(key, lock.WithTry())
	require.Error(t, err, lock.ErrAlreadyLocked)

	_, err = locker.Acquire(key, lock.WithAcquireTimeout(5*time.Second))
	require.Error(t, err, lock.ErrDeadlineExceeded)
}

func TestLockByTTL(t *testing.T) {
	locker, err := NewEtcdLocker(
		etcdClient.Config{
			Endpoints: []string{Endpoint},
			Username:  UserName,
			Password:  Password,
		},
	)
	key := "TestLockByTTLKey"

	lock1, err := locker.Acquire(key, lock.WithLockTTL(time.Second))
	require.NoError(t, err)
	defer func() {
		err = lock1.Release()
		require.NoError(t, err)
	}()

	_, err = locker.Acquire(key, lock.WithTry())
	require.Error(t, err, lock.ErrAlreadyLocked)

	lock2, err := locker.Acquire(key, lock.WithAcquireTimeout(5*time.Second))
	require.NoError(t, err)

	defer func() {
		err = lock2.Release()
		require.NoError(t, err)
	}()

}
