package main

import (
	"fish/hall/common"
	"fmt"
	"github.com/astaxie/beego/config"
)

func initConf() (err error) {
	conf, err := config.NewConfig("ini", "./common/conf/hall.conf")
	if err != nil {
		fmt.Println("new hall config failed,err:", err)
		return
	}

	common.HallConf.HallHost = conf.String("hall_host")
	if common.HallConf.HallHost == "" {
		return fmt.Errorf("conf err: hall_host is null")
	}

	common.HallConf.HallPort, err = conf.Int("hall_port")
	if err != nil {
		return
	}

	common.HallConf.HallSecret = conf.String("hall_secret")
	if common.HallConf.HallSecret == "" {
		return fmt.Errorf("conf err: hall_secret is null")
	}

	common.HallConf.LogPath = conf.String("log_path")
	if common.HallConf.LogPath == "" {
		return fmt.Errorf("conf err: log_path is null")
	}

	common.HallConf.LogLevel = conf.String("log_level")
	if common.HallConf.LogLevel == "" {
		return fmt.Errorf("conf err: log_level is null")
	}

	common.HallConf.Version = conf.String("version")
	if common.HallConf.Version == "" {
		return fmt.Errorf("conf err: version is null")
	}

	accountConf, err := config.NewConfig("ini", "./common/conf/account.conf")
	if err != nil {
		fmt.Println("new account config failed,err:", err)
		return
	}
	common.HallConf.AccountHost = accountConf.String("account_host")
	if common.HallConf.AccountHost == "" {
		return fmt.Errorf("conf err: account_host is null")
	}

	common.HallConf.AccountPort, err = accountConf.Int("account_port")
	if err != nil {
		return
	}

	return
}
