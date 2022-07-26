package lock

import (
	"time"
)

type Locker interface {
	Acquire(key string, opts ...LockerOpt) (Lock, error)
}

type Lock interface {
	Release() error
}

func DefaultOptions() *Options {
	return &Options{
		AcquireTimeout: 5 * time.Second,
		Try:            false,
		TTL:            0,
	}
}

type Options struct {
	// 获取锁超时时间, 超过此时间没有获取到锁返回 ErrDeadlineExceeded 错误
	AcquireTimeout time.Duration
	// 锁过期时间，如果为0则不会过期，需要手动释放锁
	TTL time.Duration
	// 尝试获取锁，如果没有获取到返回 ErrAlreadyLocked 错误
	Try bool
}

type LockerOpt func(op *Options)

// WithTry 尝试获取锁，如果没有获取到返回 ErrAlreadyLocked 错误
func WithTry() LockerOpt {
	return LockerOpt(func(op *Options) {
		op.Try = true
	})
}

// WithAcquireTimeout 获取锁超时时间, 超过此时间没有获取到锁返回 ErrDeadlineExceeded 错误
func WithAcquireTimeout(timeout time.Duration) LockerOpt {
	return LockerOpt(func(op *Options) {
		op.AcquireTimeout = timeout
	})
}

// WithLockTTL 锁过期时间，如果为0则不会过期，需要手动释放锁
func WithLockTTL(ttl time.Duration) LockerOpt {
	return LockerOpt(func(op *Options) {
		op.TTL = ttl
	})
}
