# go-consumergroup [![Build Status](https://travis-ci.org/meitu/go-consumergroup.svg?branch=master)](https://travis-ci.org/meitu/go-consumergroup) [![Go Report Card](https://goreportcard.com/badge/github.com/meitu/go-consumergroup)](https://goreportcard.com/report/github.com/meitu/go-consumergroup)

Go-consumergroup is a kafka consumer library written in golang with rebalance and chroot supports.

[Chinese Doc](./README.zh-CN.md)

## Requirements
* Apache Kafka 0.8.x, 0.9.x, 0.10.x

## Dependencies
* [go-zookeeper](https://github.com/samuel/go-zookeeper)
* [sarama](https://github.com/Shopify/sarama)
* [zk_wrapper](https://github.com/meitu/zk_wrapper)

## Getting started 

* API documentation and examples are available via [godoc](https://godoc.org/github.com/meitu/go-consumergroup).
* The example directory contains more elaborate [example](example/example.go) applications.

## User Defined Logger 

```
logger := logrus.New()
cg.SetLogger(logger)
```

## Run Tests

```shell
$ make test
```

***NOTE:*** `docker-compse` is required to run tests
