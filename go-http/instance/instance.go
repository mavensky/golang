// Package instance global instance
// Created by chenguolin 2018-11-16
package instance

import (
	"fmt"
	"os"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-redis/redis"

	"gitlab.local.com/golang/go-common/time"
	"gitlab.local.com/golang/go-http/config"
	"gitlab.local.com/golang/go-mysql"
)

var (
	mysqlClient *mysql.Mysql
	redisClient *redis.ClusterClient
	mcClient    *memcache.Client
	kafkaConf   *config.KafkaConf
)

// PkgInitFunc pkg init function
type PkgInitFunc func()

// pkgInitFuncs all init functions
var pkgInitFuncs = make(map[string]PkgInitFunc, 0)

// AddInitFunc add InitFunc 2 pkgInitFuncs
// same name InitFunc will be override
func AddInitFunc(name string, f PkgInitFunc) {
	pkgInitFuncs[name] = f
}

// AppInit init application
func AppInit(filePath string) {
	fmt.Println(fmt.Sprintf("%s AppInit start ...", time.GetCurrentTime()))
	// 获取api模块配置
	conf := config.GetConfig(filePath)
	if conf == nil {
		panic("AppInit GetConfig is nil")
	}

	var err error

	// new instance
	mysqlClient, err = newMysqlClient(conf.Mysql)
	if err != nil {
		panic(fmt.Sprintf("AppInit newMysqlClient error:%s", err))
	}
	redisClient, err = newRedisClient(conf.Redis)
	if err != nil {
		panic(fmt.Sprintf("AppInit newRedisClient error:%s", err))
	}
	mcClient, err = newMcClient(conf.Memcache)
	if err != nil {
		panic(fmt.Sprintf("AppInit newMcClient error:%s", err))
	}
	kafkaConf = conf.Kafka

	// pkg下相关的service执行Init函数
	for name, f := range pkgInitFuncs {
		f()
		fmt.Println(fmt.Sprintf("run package:%s Init function ok ~", name))
	}
	fmt.Println(fmt.Sprintf("%s AppInit successful ~", time.GetCurrentTime()))
}

// AppInitTest test init application
func AppInitTest() {
	// TODO 用户需要自行修改本地配置文件路径
	filePath := os.Getenv("GOPATH") + "/src/gitlab.local.com/golang/go-http/config/conf/config-pre.toml"
	AppInit(filePath)
}

// GetMysqlClient get mysqlClient
func GetMysqlClient() *mysql.Mysql {
	return mysqlClient
}

// GetRedisClient get redisClient
func GetRedisClient() *redis.ClusterClient {
	return redisClient
}

// GetMcClient get mcClient
func GetMcClient() *memcache.Client {
	return mcClient
}

// GetKafkaConf get KafkaConf
func GetKafkaConf() *config.KafkaConf {
	return kafkaConf
}
