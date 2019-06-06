package main

import (
	"fish/account/common"
	"fish/account/service"
	"fish/common/api/thrift/gen-go/rpc"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/astaxie/beego/logs"
)

func main() {
	err := initConf()
	if err != nil {
		logs.Error("init conf err:%v",err)
		return
	}
	logs.Debug("init conf success")
	service.InitAesTool()
	err = initSec()
	if err != nil {
		logs.Error("initSec err:%v",err)
		return
	}
	logs.Debug("init sec success")

	port := fmt.Sprintf(":%d",common.AccountConf.ThriftPort)
	transport, err := thrift.NewTServerSocket(port)
	if err != nil {
		panic(err)
	}
	handler := &service.UserServer{}
	processor := rpc.NewUserServiceProcessor(handler)

	transportFactory := thrift.NewTBufferedTransportFactory(8192)
	protocolFactory := thrift.NewTCompactProtocolFactory()

	server := thrift.NewTSimpleServer4(
		processor,
		transport,
		transportFactory,
		protocolFactory,
	)
	logs.Debug("account server %s",port)
	if err := server.Serve(); err != nil {
		panic(err)
	}
}