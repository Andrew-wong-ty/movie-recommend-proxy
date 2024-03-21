package proxy

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
)

var increaseOrCreate = `
	local N = tonumber(ARGV[1]) -- maximum requests
	local SEC = tonumber(ARGV[2]) -- the expiration time
	local exist = redis.call('EXISTS', KEYS[1])
	if exist == 1 then
		-- increment and check current count
		local count = redis.call('INCR', KEYS[1])
		if count > N then
			return 0
		else
			return 1
		end
	else
		-- create a new key and then set expiration time
		redis.call('SET', KEYS[1], 0)
		redis.call('EXPIRE', KEYS[1], SEC)
		return 1
	end
`

type RateLimiter struct {
	maxRequest   int
	expireTime   int
	redisClient  *redis.Client
	rateLimitKey string
}

func (rl *RateLimiter) CanProcess() bool {
	ctx := context.Background()
	result, err := rl.redisClient.Eval(ctx, increaseOrCreate,
		[]string{rl.rateLimitKey}, rl.maxRequest, rl.expireTime).Result()
	if err != nil {
		log.Fatal(err)
	}
	resInt := result.(int64)
	if resInt == 0 {
		return false
	} else {
		return true
	}
}

func NewRateLimiter(maxRequest, expireTime int,
	rateLimitKey, redisAddr, password string) *RateLimiter {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: password,
		DB:       0,
	})
	return &RateLimiter{
		maxRequest:   maxRequest,
		expireTime:   expireTime,
		redisClient:  redisClient,
		rateLimitKey: rateLimitKey,
	}
}
