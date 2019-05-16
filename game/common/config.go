package common

import (
	"github.com/garyburd/redigo/redis"
	etcd_client "github.com/coreos/etcd/clientv3"
	"sync"
	"github.com/jmoiron/sqlx"
)

type GameConf struct {
	HttpPort    int
	RedisConf   RedisConf
	RedisKey    RedisKey
	EtcdConf    EtcdConf
	LogPath     string
	LogLevel    string
	AppSecret   string
	PumpingRate int //抽水率，百分之几
	MysqlConf   MysqlConf
	//RwBlackLock sync.RWMutex
}

type MysqlConf struct {
	MysqlAddr     string
	MysqlUser     string
	MysqlPassword string
	MysqlDatabase string
	Pool          *sqlx.DB
}

type RedisConf struct {
	RedisAddr        string
	RedisMaxIdle     int
	RedisMaxActive   int
	RedisIdleTimeout int
	RedisPool        *redis.Pool
}

type RedisKey struct {
	RedisKeyUserStake  string
	RedisKeyInvestBase string
}

type EtcdConf struct {
	EtcdAddr string
	Timeout  int
	//EtcdSecKeyPrefix  string
	//EtcdSecProductKey string
	EtcdClient *etcd_client.Client
}

type CanStakeChip struct {
	GoldMin int64
	GoldMax int64
	Chips   [3]int
}

type GameManage struct {
	Periods        int    //设置期数id  自增
	GameStatus     int    //游戏状态
	Timer          int    //当局秒数
	StakeCountdown int    //押注倒计时
	SendTime       int    //本局已发送押注的次数
	TimesId        string //场次id  唯一字符串 期数是循环的，timesId作为唯一标识
	Lock           sync.RWMutex
}

type InvestConf struct {
	Id          int           `json:"id"`
	Name        string        `json:"name"`
	ParentId    int           `json:"parent_id"`
	Rate        float32       `json:"rate"`
	Weight      int           `json:"weight"`
	Icon        string        `json:"icon"`
	ParentArr   []*InvestConf `json:"-"`
	WinPosition []int         `json:"-"`
}
