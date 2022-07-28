# Lock

Go distributed lock library

## ETCD Drive

```go
locker, _ := drive_etcd.NewEtcdLocker(
    etcdClient.Config{
        Endpoints: []string{"http://0.0.0.0:2379"},
    },
)
lock1, err := locker.Acquire("key")
if err != nil {
    fmt.Println(err)
    return
}
defer lock1.Release()

// Do something
```


获取锁超时时间, 超过此时间没有获取到锁返回 `ErrDeadlineExceeded` 错误
```go
locker.Acquire(key, lock.WithAcquireTimeout(5*time.Second))
```

非堵塞尝试获取锁，如果没有获取到锁返回 `ErrAlreadyLocked` 错误
```go
locker.Acquire(key, lock.WithTry(true))
```

锁过期时间，如果为0则会自动续约不会过期，需要手动释放锁
```go
locker.Acquire(key, lock.WithLockTTL(5*time.Second))
```
