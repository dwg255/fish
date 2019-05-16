package service

import (
	"context"
	"fish/account/common"
	"fish/common/api/thrift/gen-go/rpc"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
)
var (
	redisConf = common.AccountConf.RedisConf
)
type UserServer struct {
}

func (p *UserServer) CreateNewUser(ctx context.Context, nickName string, avatarAuto string, gold int64) (r *rpc.Result_, err error) {
	logs.Debug("CreateNewUser nickName: %v", nickName)
	re,err := redisConf.RedisPool.Exists("test").Result()
	if err != nil {
		err = fmt.Errorf("redis err: %v",err)
		return
	}
	if re == 1 {
		err = fmt.Errorf("key already exists!")
		return
	}
	return
}

func (p *UserServer) GetUserInfoById(ctx context.Context, userId int32) (r *rpc.Result_, err error) {
	result ,err := redisConf.RedisPool.HGetAll(redisConf.RedisKeyPrefix + strconv.Itoa(int(userId))).Result()
	if err != nil {
		return r,err
	}
	if len(result) == 0 {
		err = fmt.Errorf("")
	}

	return
}

func (p *UserServer) GetUserInfoByken(ctx context.Context, token string) (r *rpc.Result_, err error) {
	return
}

func (p *UserServer) ModifyGoldById(ctx context.Context, behavior string, userId int32, gold int64) (r *rpc.Result_, err error) {
	return
}

func (p *UserServer) ModifyGoldByToken(ctx context.Context, behavior string, token string, gold int64) (r *rpc.Result_, err error) {
	return
}
