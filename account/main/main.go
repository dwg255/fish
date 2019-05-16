package main

import (
	"context"
	"crypto/tls"
	"fish/account/service"
	"fish/common/api/thrift/gen-go/rpc"
	"fmt"
	"time"

	//"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/astaxie/beego/logs"
)
func Set(key string, value interface{}, expiration time.Duration)  {

	return
}

func main() {
	err := initConf()
	if err != nil {
		logs.Error("init conf err:%v",err)
		return
	}
	logs.Debug("init conf success")

	err = initSec()
	if err != nil {
		logs.Error("initSec err:%v",err)
		return
	}
	logs.Debug("init sec success")

	port := fmt.Sprintf(":%d",8000)
	transport, err := thrift.NewTServerSocket(port)
	if err != nil {
		panic(err)
	}
	handler := &service.UserServer{}
	handler.CreateNewUser(context.Background(),"testnick","headimg",1000)
	handler.GetUserInfoById(context.Background(),1)
	processor := rpc.NewUserServiceProcessor(handler)

	transportFactory := thrift.NewTBufferedTransportFactory(8192)
	protocolFactory := thrift.NewTCompactProtocolFactory()

	server := thrift.NewTSimpleServer4(
		processor,
		transport,
		transportFactory,
		protocolFactory,
	)
	if err := server.Serve(); err != nil {
		panic(err)
	}
}
func runServer(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string, secure bool) error {
	var transport thrift.TServerTransport
	var err error
	if secure {
		cfg := new(tls.Config)
		if cert, err := tls.LoadX509KeyPair("server.crt", "server.key"); err == nil {
			cfg.Certificates = append(cfg.Certificates, cert)
		} else {
			return err
		}
		transport, err = thrift.NewTSSLServerSocket(addr, cfg)
	} else {
		transport, err = thrift.NewTServerSocket(addr)
	}

	if err != nil {
		return err
	}
	fmt.Printf("%T\n", transport)
	handler := &service.UserServer{}
	processor := rpc.NewUserServiceProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)

	fmt.Println("Starting the simple server... on ", addr)
	return server.Serve()
}