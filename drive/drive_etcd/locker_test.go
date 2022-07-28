package drive_etcd

import (
	"github.com/stretchr/testify/require"
	"github.com/xyctruth/lock"
	etcdClient "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"os"
	"testing"
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

func TestInvalidEndpointNewClient(t *testing.T) {
	_, err := NewEtcdLocker(
		etcdClient.Config{
			Endpoints:   []string{"1.1.1.1:2379"},
			DialTimeout: 1 * time.Second,
			DialOptions: []grpc.DialOption{grpc.WithBlock()},
		},
	)
	require.Error(t, err, lock.ErrDeadlineExceeded)
}

func TestInvalidEndpointLock(t *testing.T) {
	locker, err := NewEtcdLocker(
		etcdClient.Config{
			Endpoints: []string{"1.1.1.1:2379"},
		},
	)
	require.NoError(t, err)
	key := "TestInvalidEndpointLock"

	_, err = locker.Acquire(key, lock.WithAcquireTimeout(time.Second))
	require.Error(t, err, lock.ErrDeadlineExceeded)
}

func TestLockByGlobalTTL(t *testing.T) {
	locker, err := NewEtcdLocker(
		etcdClient.Config{
			Endpoints: []string{DefaultEndpoint},
			Username:  DefaultUserName,
			Password:  DefaultPassword,
		},
		lock.WithLockTTL(time.Second*2),
	)
	key := "TestLockByGlobalTTL"
	lock1, err := locker.Acquire(key)
	require.NoError(t, err)
	defer func() {
		err = lock1.Release()
		require.NoError(t, err)
	}()

	_, err = locker.Acquire(key, lock.WithTry(true))
	require.Error(t, err, lock.ErrAlreadyLocked)

	lock2, err := locker.Acquire(key, lock.WithAcquireTimeout(3*time.Second))
	require.NoError(t, err)

	defer func() {
		err = lock2.Release()
		require.NoError(t, err)
	}()
}

func TestLockByTTL(t *testing.T) {
	locker, err := NewEtcdLocker(
		etcdClient.Config{
			Endpoints: []string{DefaultEndpoint},
			Username:  DefaultUserName,
			Password:  DefaultPassword,
		},
	)
	key := "TestLockByTTL"

	lock1, err := locker.Acquire(key, lock.WithLockTTL(time.Second*2))
	require.NoError(t, err)
	defer func() {
		err = lock1.Release()
		require.NoError(t, err)
	}()

	_, err = locker.Acquire(key, lock.WithTry(true))
	require.Error(t, err, lock.ErrAlreadyLocked)

	lock2, err := locker.Acquire(key, lock.WithAcquireTimeout(3*time.Second))
	require.NoError(t, err)

	defer func() {
		err = lock2.Release()
		require.NoError(t, err)
	}()
}

func TestLock(t *testing.T) {
	locker, err := NewEtcdLocker(
		etcdClient.Config{
			Endpoints: []string{DefaultEndpoint},
			Username:  DefaultUserName,
			Password:  DefaultPassword,
		},
	)
	key := "TestLock"

	lock1, err := locker.Acquire(key)
	require.NoError(t, err)
	defer func() {
		err = lock1.Release()
		require.NoError(t, err)
	}()

	_, err = locker.Acquire(key, lock.WithTry(true))
	require.Error(t, err, lock.ErrAlreadyLocked)

	_, err = locker.Acquire(key, lock.WithAcquireTimeout(3*time.Second))
	require.Error(t, err, lock.ErrDeadlineExceeded)
}
