package drive_etcd

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/xyctruth/lock"
	etcdClient "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type Locker struct {
	client *etcdClient.Client
	op     *lock.Options
}

func NewEtcdLocker(config etcdClient.Config, opts ...lock.LockerOpt) (lock.Locker, error) {
	client, err := etcdClient.New(config)
	if err != nil {
		return nil, err
	}

	op := lock.DefaultOptions()
	for _, opt := range opts {
		opt(op)
	}

	locker := &Locker{
		client: client,
		op:     op,
	}
	return locker, nil
}

func (locker *Locker) Acquire(key string, opts ...lock.LockerOpt) (lock.Lock, error) {
	return locker.acquire(key, opts...)
}

func (locker *Locker) acquire(key string, opts ...lock.LockerOpt) (lock.Lock, error) {
	op := &lock.Options{
		AcquireTimeout: locker.op.AcquireTimeout,
		TTL:            locker.op.TTL,
		Try:            locker.op.Try,
	}

	for _, opt := range opts {
		opt(op)
	}

	ttl := int(op.TTL / time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), op.AcquireTimeout)
	defer cancel()
	session, err := concurrency.NewSession(locker.client, concurrency.WithTTL(ttl), concurrency.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	key = addPrefix(key)
	mutex := concurrency.NewMutex(session, key)

	if op.Try {
		// 无堵塞 尝试获取锁
		tryLockErr := locker.tryLock(ctx, mutex)
		if tryLockErr == concurrency.ErrLocked {
			session.Close()
			return nil, lock.ErrAlreadyLocked
		}

		if tryLockErr != nil {
			session.Close()
			return nil, tryLockErr
		}
	} else {
		// 堵塞获取锁
		lockErr := locker.lock(ctx, mutex)

		if lockErr == context.DeadlineExceeded {
			session.Close()
			return nil, lock.ErrDeadlineExceeded
		}

		if lockErr != nil {
			session.Close()
			return nil, lockErr
		}
	}

	// 传入了 ttl 不自动续约, 如果没传 ttl 会自动续约
	if ttl > 0 {
		session.Orphan()
	}
	l := &Lock{mutex: mutex, Mutex: &sync.Mutex{}, session: session}

	return l, nil
}

func (locker *Locker) lock(ctx context.Context, mutex *concurrency.Mutex) error {
	return mutex.Lock(ctx)
}

func (locker *Locker) tryLock(ctx context.Context, mutex *concurrency.Mutex) error {
	return mutex.TryLock(ctx)
}

const (
	prefix = "/etcdClient-lock"
)

func addPrefix(key string) string {
	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}
	return prefix + key
}

type Lock struct {
	*sync.Mutex
	mutex   *concurrency.Mutex
	session *concurrency.Session
}

func (l *Lock) Release() error {
	if l == nil {
		return lock.ErrNotFoundLock
	}
	l.Lock()
	defer l.Unlock()

	defer l.session.Close()
	return l.mutex.Unlock(context.Background())
}
