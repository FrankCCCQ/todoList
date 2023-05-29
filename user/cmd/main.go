package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"
	"user/config"
	"user/discovery"
	"user/internal/handler"
	"user/internal/repository"
	"user/internal/service"
)

func main() {
	config.InitConfig()
	repository.InitDB()
	// grpc 服务器的开始 绑定服务到这个grpc 服务器上
	grpcAddress := viper.GetString("server.grpcAddress")

	// etcd 的地址
	etcdAddress := []string{viper.GetString("etcd.address")}

	// 服务注册
	etcdRegister := discovery.NewRegister(etcdAddress, logrus.New())
	userNode := discovery.Server{
		Name: viper.GetString("server.domain"),
		Addr: grpcAddress,
	}
	if _, err := etcdRegister.Register(userNode, 10); err != nil {
		panic(err)
	}

	// 作为grpc 的 server
	server := grpc.NewServer()
	defer server.Stop()

	// 绑定服务
	service.RegisterUserServiceServer(server, handler.NewUserService())
	// 监听端口
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		panic(err)
	}
	if err = server.Serve(lis); err != nil {
		panic(err)
	}
}
