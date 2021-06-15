package lib

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)
//初始化redis
func InitRedisPool(path string) (err error) {
	//普通的db方式
	ConfRedisMap := &RedisMapConf{}
	err = ParseConfig(path, ConfRedisMap)
	if err != nil {
		return err
	}
	if len(ConfRedisMap.List) == 0 {
		Log.TagInfo(NewTrace(),DLTagRedisEmpty, map[string]interface{}{})
	}
	RedisMapPool = map[string]*redis.Client{}

	for confName, cfg := range ConfRedisMap.List {
		if cfg.ReadTimeout == 0 {
			cfg.ReadTimeout = 100
		}
		if cfg.WriteTimeout == 0 {
			cfg.WriteTimeout = 100
		}
		rdb := redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password, // no password set
			DB:       cfg.Db,  // use default DB
			WriteTimeout: time.Duration(cfg.WriteTimeout)*time.Second,
			ReadTimeout: time.Duration(cfg.ReadTimeout)*time.Second,

		})

		RedisMapPool[confName] = rdb

	}

	return
}
//获取redis
func GetRedisPool(name string) (*redis.Client, error) {
	if db, ok := RedisMapPool[name]; ok {
		return db, nil
	}
	return nil, errors.New("get redis error")
}
//通过配置 执行redis
func RedisConfDo(trace *TraceContext, name string, args ...interface{}) (interface{}, error) {
	c, err := GetRedisPool(name)
	if err != nil {
		Log.TagError(trace, DLTagRedisFailed, map[string]interface{}{
			"err":  "GetRedisPool_error:"+name,
			"bind":   args,
		})
		return nil, err
	}
	startExecTime := time.Now()
	var ctx = context.Background()
	reply, err := c.Do(ctx, args...).Result()
	endExecTime := time.Now()
	if err != nil {
		Log.TagError(trace, DLTagRedisFailed, map[string]interface{}{
			"err":       err,
			"bind":      args,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	}
	return reply, err
}
//关闭数据库链接
func CloseRedis() error {
	for _, rdb := range RedisMapPool {
		rdb.Close()
	}
	return nil
}