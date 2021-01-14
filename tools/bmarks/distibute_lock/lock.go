package distibute_lock


import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

var (
	LockRedis *redis.Client
)

type LockObject struct {
	expiration time.Duration
	cacheKey   string
	timeout    int64
	needWait   bool
	run        func()
}

func LockTimeout(key string, args map[string]interface{}, exec func()) {
	if LockRedis == nil {
		panic("not initialize redis client")
	}
	timeout := int64(5)
	if args["timeout"] != nil {
		timeout, _ = args["timeout"].(int64)
	}
	expiration := time.Duration(-1)
	if args["expiration"] != nil {
		e, _ := args["expiration"].(int)
		expiration = time.Duration(e)
	}
	l := LockObject{
		expiration: expiration,
		cacheKey:   key,
		timeout:    timeout,
		needWait:   false,
		run:        exec,
	}
	l.lock()
}

func (l *LockObject) lock() {
	expiration := time.Duration(-1)

	l.tryUntilTimeout(func() error {
		expiration = l.generateExpiration()
		if ok, err := LockRedis.SetNX(l.cacheKey, expiration, 0).Result(); ok && err == nil {
			return nil
		}
		if l.expiration != -1 {
			oldExp := int64(0)
			if v, err := LockRedis.Get(l.cacheKey).Result(); err == nil {
				if x, err1 := strconv.Atoi(v); err1 != nil {
					oldExp = int64(x)
				}
			}
			if oldExp < time.Now().UnixNano() {
				expiration = l.generateExpiration()

				r, _ := LockRedis.GetSet(l.cacheKey, expiration).Result()
				r1, _ := strconv.Atoi(r)
				oldExp = int64(r1)
				if oldExp < time.Now().UnixNano() {
					return nil
				}
			}
		}
		return errors.New("need continue")
	})

	defer func() {
		if l.expiration == -1 || expiration > time.Duration(time.Now().UnixNano()) {
			LockRedis.Del(l.cacheKey)
		}
	}()
	l.run()
}

func (l *LockObject) generateExpiration() time.Duration {
	if l.expiration == -1 {
		return 1
	}
	return time.Duration(time.Now().Add(l.expiration * time.Second).Add(1 * time.Second).UnixNano())
}

func (l *LockObject) tryUntilTimeout(yield func() error) {
	if l.timeout == 0 {
		if err := yield(); err == nil {
			return
		}
	} else {
		start := time.Now().UnixNano()
		for true {
			diff := (time.Now().UnixNano() - start) / int64(time.Second)
			if diff > l.timeout {
				break
			}
			if err := yield(); err == nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	msg := fmt.Sprintf("Timeout on lock %s exceeded %d sec", l.cacheKey, l.timeout)
	panic(msg)
}
