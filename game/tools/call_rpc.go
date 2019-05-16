package tools

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"fmt"
	"fish/common/api/thrift/gen-go/rpc"
	"net"
)

func GetRpcClient() (client *rpc.UserServiceClient, closeTransport func() error, err error) {
	var transport thrift.TTransport
	transport, err = thrift.NewTSocket(net.JoinHostPort("127.0.0.1", "8989"))
	if err != nil {
		fmt.Errorf("NewTSocket failed. err: [%v]\n", err)
		return
	}

	transport, err = thrift.NewTBufferedTransportFactory(8192).GetTransport(transport)
	if err != nil {
		fmt.Errorf("NewTransport failed. err: [%v]\n", err)
		return
	}
	closeTransport = transport.Close

	if err = transport.Open(); err != nil {
		fmt.Errorf("Transport.Open failed. err: [%v]\n", err)
		return
	}
	protocolFactory := thrift.NewTCompactProtocolFactory()
	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)
	client = rpc.NewUserServiceClient(thrift.NewTStandardClient(iprot, oprot))
	return
}
