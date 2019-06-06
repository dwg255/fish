package main

import (
	"fish/account/common"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

func initConf() (err error) {
	conf, err := config.NewConfig("ini", "./common/conf/account.conf")
	if err != nil {
		fmt.Println("new config failed,err:", err)
		return
	}

	common.AccountConf.ThriftPort, err = conf.Int("account_port")
	if err != nil {
		return
	}
	common.AccountConf.AccountAesKey = conf.String("account_aes_key")
	if common.AccountConf.AccountAesKey == "" || len(common.AccountConf.AccountAesKey) < 16 {
		return fmt.Errorf("conf err: invalid account_aes_key :%v", common.AccountConf.AccountAesKey)
	}
	logs.Debug("account_aes_key :%v",common.AccountConf.AccountAesKey)

	common.AccountConf.LogPath = conf.String("log_path")
	if common.AccountConf.LogPath == "" {
		return fmt.Errorf("conf err: log_path is null")
	}

	common.AccountConf.LogLevel = conf.String("log_level")
	if common.AccountConf.LogLevel == "" {
		return fmt.Errorf("conf err: log_level is null")
	}

	//redis配置
	common.AccountConf.RedisConf.RedisAddrs = conf.Strings("redis_addrs")
	if len(common.AccountConf.RedisConf.RedisAddrs) == 0 {
		return fmt.Errorf("conf err: redis addr is null")
	}
	common.AccountConf.RedisConf.RedisKeyPrefix = conf.String("redis_key_prefix")
	if len(common.AccountConf.RedisConf.RedisKeyPrefix) == 0 {
		return fmt.Errorf("conf err: redis_key_prefix is null")
	}

	//mysql配置
	common.AccountConf.MysqlConf.MysqlAddr = conf.String("mysql_addr")
	if len(common.AccountConf.MysqlConf.MysqlAddr) == 0 {
		return fmt.Errorf("conf err: mysql_addr is null")
	}
	common.AccountConf.MysqlConf.MysqlUser = conf.String("mysql_user")
	if len(common.AccountConf.MysqlConf.MysqlUser) == 0 {
		return fmt.Errorf("conf err: mysql_user is null")
	}
	common.AccountConf.MysqlConf.MysqlPassword = conf.String("mysql_password")
	if len(common.AccountConf.MysqlConf.MysqlPassword) == 0 {
		return fmt.Errorf("conf err: mysql_password is null")
	}
	common.AccountConf.MysqlConf.MysqlDatabase = conf.String("mysql_db")
	if len(common.AccountConf.MysqlConf.MysqlDatabase) == 0 {
		return fmt.Errorf("conf err: mysql_db is null")
	}
	return
}
