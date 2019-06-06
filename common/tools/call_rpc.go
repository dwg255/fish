package tools

import (
	"fish/common/api/thrift/gen-go/rpc"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"net"
)

func GetRpcClient(host,port string) (client *rpc.UserServiceClient, closeTransport func() error, err error) {
	var transport thrift.TTransport
	transport, err = thrift.NewTSocket(net.JoinHostPort(host,port))
	if err != nil {
		err = fmt.Errorf("NewTSocket failed. err: [%v]\n", err)
		return
	}

	transport, err = thrift.NewTBufferedTransportFactory(8192).GetTransport(transport)
	if err != nil {
		err = fmt.Errorf("NewTransport failed. err: [%v]\n", err)
		return
	}
	closeTransport = transport.Close

	if err = transport.Open(); err != nil {
		err = fmt.Errorf("Transport.Open failed. err: [%v]\n", err)
		return
	}
	protocolFactory := thrift.NewTCompactProtocolFactory()
	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)
	client = rpc.NewUserServiceClient(thrift.NewTStandardClient(iprot, oprot))
	return
}