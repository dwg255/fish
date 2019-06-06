package main

import (
	"encoding/json"
	"fish/account/common"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func conversionLogLevel(logLevel string) int {
	switch logLevel {
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	}
	return logs.LevelDebug
}

func initLogger() (err error) {
	config := make(map[string]interface{})
	config["filename"] = common.AccountConf.LogPath
	config["level"] = conversionLogLevel(common.AccountConf.LogLevel)

	configStr, err := json.Marshal(config)
	if err != nil {
		return
	}
	err = logs.SetLogger(logs.AdapterFile, string(configStr))
	return
}

func initMysql() (err error) {
	conf := common.AccountConf.MysqlConf
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", conf.MysqlUser, conf.MysqlPassword, conf.MysqlAddr, conf.MysqlDatabase)
	logs.Debug(dsn)
	database, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return
	}
	common.AccountConf.MysqlConf.Pool = database
	return
}

func initRedis() (err error) {
	//client := redis.NewClusterClient(&redis.ClusterOptions{
	//	Addrs: common.AccountConf.RedisConf.RedisAddrs,
	//})
	//_, err = client.Ping().Result()
	//if err != nil {
	//	return
	//}
	//common.AccountConf.RedisConf.RedisPool = client
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	common.AccountConf.RedisConf.RedisPool = client
	return
}

func initSec() (err error) {
	err = initLogger()
	if err != nil {
		return
	}

	err = initRedis()
	if err != nil {
		return
	}

	/*err = initMysql()
	if err != nil {
		return
	}*/
	return
}
