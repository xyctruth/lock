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

type Config struct {
	// Endpoints is a list of URLs.
	Endpoints []string `json:"endpoints"`

	// DialTimeout is the timeout for failing to establish a connection.
	DialTimeout time.Duration `json:"dial-timeout"`

	// Username is a user name for authentication.
	Username string `json:"username"`

	// Password is a password for authentication.
	Password string `json:"password"`
}

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
	session, err := concurrency.NewSession(locker.client, concurrency.WithTTL(ttl))
	if err != nil {
		return nil, err
	}

	key = addPrefix(key)
	mutex := concurrency.NewMutex(session, key)

	if op.Try {
		// 无堵塞 尝试获取锁，已上锁返回已上锁错误
		tryLockErr := locker.tryLock(mutex, op.AcquireTimeout)
		if tryLockErr == concurrency.ErrLocked {
			session.Close()
			return nil, lock.ErrAlreadyLocked
		}

		if tryLockErr != nil {
			session.Close()
			return nil, tryLockErr
		}
	} else {
		// 堵塞获取锁， 超过获取锁时间返回  超过了最后期限错误
		tryLockErr := locker.lock(mutex, op.AcquireTimeout)

		if tryLockErr == context.DeadlineExceeded {
			session.Close()
			return nil, lock.ErrDeadlineExceeded
		}

		if tryLockErr != nil {
			session.Close()
			return nil, tryLockErr
		}
	}

	// 不自动续约
	if ttl > 0 {
		session.Orphan()
	}
	l := &Lock{mutex: mutex, Mutex: &sync.Mutex{}, session: session}

	return l, nil
}

type Lock struct {
	*sync.Mutex
	mutex   *concurrency.Mutex
	session *concurrency.Session
}

func (l *Lock) Release() error {
	if l == nil {
		return lock.ErrNotFoundLocked
	}
	l.Lock()
	defer l.Unlock()

	defer l.session.Close()
	return l.mutex.Unlock(context.Background())
}

func (locker *Locker) lock(mutex *concurrency.Mutex, tryLockTimeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), tryLockTimeout)
	defer cancel()
	return mutex.Lock(ctx)
}

func (locker *Locker) tryLock(mutex *concurrency.Mutex, tryLockTimeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), tryLockTimeout)
	defer cancel()
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
