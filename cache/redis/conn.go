package redis

import (
	"github.com/garyburd/redigo/redis"
	"time"
	"fmt"
)

var (
	pool *redis.Pool
	redisHost = "127.0.0.1:6379"
	redisPass = "123456"
)

//创建一个redis连接池
func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:100,
		MaxActive:100,
		IdleTimeout:300*time.Second,
		Dial: func() (redis.Conn, error) {
			//打开连接
			conn, err := redis.Dial("tcp", redisHost)
			if err != nil {
				fmt.Println(err.Error())
				return nil, err
			}

			//访问认证
			if _, err := conn.Do("AUTH", redisPass); err != nil {
				conn.Close()
				return nil, err
			}
			return conn, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

//初始化
func init()  {
	pool = newRedisPool()
}

//暴露连接池
func RedisPool() *redis.Pool {
	return pool
}

