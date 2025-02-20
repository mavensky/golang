// Package config common
// Created by chenguolin 2018-11-16
package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

// global conf
var conf *Config

// loadFrom load toml file
func loadFrom(filePath string) *Config {
	if conf != nil {
		return conf
	}

	_, err := os.Stat(filePath)
	if err != nil {
		panic(err)
	}

	conf = &Config{}
	_, err = toml.DecodeFile(filePath, conf)
	if err != nil {
		panic(err)
	}

	return conf
}

// GetConfig get Config
func GetConfig(filePath string) *Config {
	if filePath == "" {
		return nil
	}

	return loadFrom(filePath)
}

// GetDeployConf get DeployConf
func GetDeployConf() *DeployConf {
	if conf == nil {
		return nil
	}

	return conf.Deploy
}

// GetMysqlConf get MysqlConf
func GetMysqlConf() *MysqlConf {
	if conf == nil {
		return nil
	}

	return conf.Mysql
}

// GetRedisConf get RedisConf
func GetRedisConf() *RedisConf {
	if conf == nil {
		return nil
	}

	return conf.Redis
}

// GetMemcacheConf get MemcacheConf
func GetMemcacheConf() *MemcacheConf {
	if conf == nil {
		return nil
	}

	return conf.Memcache
}

// GetKafkaConf get KafkaConf
func GetKafkaConf() *KafkaConf {
	if conf == nil {
		return nil
	}

	return conf.Kafka
}
