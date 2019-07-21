package redis

import (
	"errors"
	"fmt"
	redis "github.com/go-redis/redis"
	"github.com/nothollyhigh/kiss/log"
	"github.com/nothollyhigh/kiss/util"
	"strings"
	"sync"
	"time"
)

type ClusterConfig struct {
	Addrs    []string `json:"Addrs"`
	Password string   `json:"Password"`
	// Database           int      `json:"Database"`
	PoolSize           int  `json:"PoolSize"`
	DialTimeout        int  `json:"DialTimeout"`
	ReadTimeout        int  `json:"ReadTimeout"`
	WriteTimeout       int  `json:"WriteTimeout"`
	PoolTimeout        int  `json:"PoolTimeout"`
	IdleTimeout        int  `json:"IdleTimeout"`
	IdleCheckFrequency int  `json:"IdleCheckFrequency"`
	KeepaliveInterval  int  `json:"KeepaliveInterval"`
	MaxRedirects       int  `json:"MaxRedirects"`
	ReadOnly           bool `json:"ReadOnly"`
	RouteByLatency     bool `json:"RouteByLatency"`
	dialTimeout        time.Duration
	readTimeout        time.Duration
	writeTimeout       time.Duration
	poolTimeout        time.Duration
	idleTimeout        time.Duration
	idleCheckFrequency time.Duration
	keepaliveInterval  time.Duration
}

type RedisCluster struct {
	sync.RWMutex
	client  *redis.ClusterClient
	ticker  *time.Ticker
	scripts map[string]*redisScript
}

func (rds *RedisCluster) Client() *redis.ClusterClient {
	return rds.client
}

func (rds *RedisCluster) getScriptSha1(tag string) (string, error) {
	rds.RLock()
	defer rds.RUnlock()
	if script, ok := rds.scripts[tag]; ok {
		return script.sha1, script.err
	}
	return "", errors.New(fmt.Sprintf("invalid redis lua script tag: %s", tag))
}

// 应用层应在初始化阶段完成所有LoadScript操作, 初始化后不应再调用此方法
func (rds *RedisCluster) LoadScript(tag string, src string) error {
	rds.Lock()
	defer rds.Unlock()
	client := rds.Client()
	cmd := client.ScriptLoad(src)
	sha1, err := cmd.Result()
	rds.scripts[tag] = &redisScript{
		tag:  tag,
		src:  src,
		sha1: sha1,
		err:  err,
	}
	if err != nil {
		log.Error("redis ScriptLoad '%s' failed: %v", tag, err)
	}
	return err
}

func (rds *RedisCluster) reLoadScript(tag string) error {
	rds.Lock()
	defer rds.Unlock()

	var (
		err  error = errors.New("no such script")
		sha1       = ""
	)
	if script, ok := rds.scripts[tag]; ok {
		cmd := rds.Client().ScriptLoad(script.src)
		sha1, err = cmd.Result()
		script.sha1, script.err = sha1, err
	}

	if err != nil {
		log.Error("redis ReLoadScript '%s' failed: %v", tag, err)
	}

	return err
}

func (rds *RedisCluster) EvalSha(tag string, keys []string, args ...interface{}) (interface{}, error) {
	var ret interface{} = nil
	sha1, err := rds.getScriptSha1(tag)
	if err == nil {
		cmd := rds.Client().EvalSha(sha1, keys, args...)
		ret, err = cmd.Result()
		if err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT") {
			if err = rds.reLoadScript(tag); err == nil {
				if sha1, err = rds.getScriptSha1(tag); err == nil {
					cmd = rds.Client().EvalSha(sha1, keys, args...)
					ret, err = cmd.Result()
				}
			}
		}
	}

	return ret, err
}

func (rds *RedisCluster) Eval(src string, keys []string, args ...interface{}) (interface{}, error) {
	cmd := rds.Client().Eval(src, keys, args...)
	ret, err := cmd.Result()
	return ret, err
}

func (rds *RedisCluster) Close() error {
	return rds.client.Close()
}

func NewCluster(conf ClusterConfig) *RedisCluster {
	if conf.DialTimeout > 0 {
		conf.dialTimeout = time.Second * time.Duration(conf.DialTimeout)
	} else {
		conf.dialTimeout = time.Second * 3
	}
	if conf.ReadTimeout > 0 {
		conf.readTimeout = time.Second * time.Duration(conf.ReadTimeout)
	} else {
		conf.readTimeout = time.Second * 2
	}
	if conf.WriteTimeout > 0 {
		conf.writeTimeout = time.Second * time.Duration(conf.WriteTimeout)
	} else {
		conf.writeTimeout = time.Second * 2
	}
	if conf.PoolTimeout > 0 {
		conf.poolTimeout = time.Second * time.Duration(conf.PoolTimeout)
	} else {
		// conf.readTimeout = time.Second * 2
	}
	if conf.IdleTimeout > 0 {
		conf.idleTimeout = time.Second * time.Duration(conf.IdleTimeout)
	} else {
		// conf.idleTimeout  = time.Second * 2
	}
	if conf.IdleCheckFrequency > 0 {
		conf.idleCheckFrequency = time.Second * time.Duration(conf.IdleCheckFrequency)
	} else {
		// conf.idleCheckFrequency = time.Second * 300
	}
	if conf.KeepaliveInterval > 0 {
		conf.keepaliveInterval = time.Second * time.Duration(conf.KeepaliveInterval)
	} else {
		conf.keepaliveInterval = time.Second * 300
	}

	log.Info("db.NewRedisCluster Connect To Redis ...")

	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:              conf.Addrs,
		Password:           conf.Password,
		DialTimeout:        conf.dialTimeout,
		ReadTimeout:        conf.readTimeout,
		WriteTimeout:       conf.writeTimeout,
		PoolSize:           conf.PoolSize,
		PoolTimeout:        conf.poolTimeout,
		IdleTimeout:        conf.idleTimeout,
		IdleCheckFrequency: conf.idleCheckFrequency,
		MaxRedirects:       conf.MaxRedirects,
		ReadOnly:           conf.ReadOnly,
		RouteByLatency:     conf.RouteByLatency,
	})

	cmd := client.Ping()
	if cmd.Err() != nil {
		log.Fatal("db.NewRedisCluster client.Ping() Failed: %v", cmd.Err())
	}

	ticker := time.NewTicker(conf.keepaliveInterval)
	util.Go(func() {
		for {
			if _, ok := <-ticker.C; !ok {
				break
			}
			if err := client.Ping().Err(); err != nil {
				log.Debug("db.Redis Ping: %v", err)
			}
		}
	})

	log.Info("db.NewRedisCluster Connect To Redis Success")

	return &RedisCluster{
		client:  client,
		scripts: map[string]*redisScript{},
		ticker:  ticker,
	}
}
