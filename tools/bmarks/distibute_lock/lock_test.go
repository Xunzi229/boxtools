package distibute_lock

import (
    "fmt"
    "github.com/go-redis/redis"
    "sync"
    "testing"
)

func TestLockTimeout(t *testing.T) {
    LockRedis = redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:6379",
    })
	//
	wg := sync.WaitGroup{}

	count := 0
	for i := 0; i < 15; i++ {
		wg.Add(1)
		go LockTimeout("user:1", map[string]interface{}{
			"expiration": 5,
		}, func() {
			count++
			wg.Done()
			return
		})
	}

	wg.Wait()

	fmt.Println(count)
}

