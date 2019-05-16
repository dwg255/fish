package common

import (
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
)

var (
	AccountConf = &AccountServiceConf{
		RedisConf: &RedisConf{},
		MysqlConf: &MysqlConf{},
	}
)

type AccountServiceConf struct {
	ThriftPort int
	LogPath    string
	LogLevel   string
	RedisConf  *RedisConf
	MysqlConf  *MysqlConf
}

type RedisConf struct {
	RedisAddrs     []string
	RedisKeyPrefix string
	//RedisPool      *redis.ClusterClient
	RedisPool *redis.Client
}

type MysqlConf struct {
	MysqlAddr     string
	MysqlUser     string
	MysqlPassword string
	MysqlDatabase string
	Pool          *sqlx.DB
}
