# golang
golang 相关代码仓库，课余时间学习总结与归纳所写，欢迎各路大神欢迎Fork、Star和提各种反馈Bug, 谢谢~

# project
1. go-log: golang logger 基础库
2. go-mysql: golang mysql 基础库
3. go-common: golang 基础代码 公共库
4. go-cron: golang 定时任务
5. go-processor: golang 处理机
6. go-http: golang HTTP server
7. go-healthcheck: golang HTTP health pprof check
8. go-lrucache: golang lru cache
9. go-prometheus: golang prometheus监控封装
10. go-kafka: golang kafka使用封装
11. telegram-bot: telegram bot 机器人

# dependency
1. update: dep ensure -update -v

# postscript
由于项目是托管在本地Gitlab仓库下，所以项目里默认路径为**gitlab.local.com/golang/xxxx**。  
所以项目Clone下来之后如果要使用需要做个全局替换，例如替换成**yyyy**

1. Linux: sed -i 's/gitlab\.local\.com\/golang\/xxxx/yyyy/g' `find . -name "*.go" | grep -v vendor`
2. Mac: sed -i "" 's/gitlab\.local\.com\/golang\/xxxx/yyyy/g' `find . -name "*.go" | grep -v vendor`

# Gitlab
1. [Gitlab 安装使用](https://chenguolin.github.io/2018/12/18/Git-Gitlab-%E5%AE%89%E8%A3%85%E4%BD%BF%E7%94%A8/)
2. [Gitlab CI和CD配置](https://chenguolin.github.io/2018/12/24/Git-Gitlab-CI%E5%92%8CCD%E9%85%8D%E7%BD%AE/)
