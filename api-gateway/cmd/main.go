package main

import (
	"api-gateway/config"
	"api-gateway/discovery"
	service "api-gateway/internal/service"
	"api-gateway/pkg/util"
	"api-gateway/routers"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config.InitConfig()

	//etcdAddress := []string{viper.GetString("etcd.address")}
	//etcdRegister := discovery.NewResolver(etcdAddress, logrus.New())
	//resolver.Register(etcdRegister)
	go startListen()
	// 这个只是句法块 没有特殊含义
	{
		osSignal := make(chan os.Signal, 1)
		signal.Notify(osSignal, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
		s := <-osSignal
		fmt.Println("exit", s)
	}
	fmt.Println("gateway listen on :4000")
}

func startListen() {
	//etcd 注册
	etcdAddress := []string{viper.GetString("etcd.address")}
	etcdRegister := discovery.NewResolver(etcdAddress, logrus.New())
	// grpc 注册
	resolver.Register(etcdRegister)
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)

	//服务名
	userServiceName := viper.GetString("domain.user")
	taskServiceName := viper.GetString("domain.task")

	// grpc 连接
	connUser, err := RPCConnect(ctx, userServiceName, etcdRegister)
	if err != nil {
		return
	}
	fmt.Println("connUser", connUser)
	// 通过 grpc 拉去 server的服务
	userService := service.NewUserServiceClient(connUser)

	connTask, err := RPCConnect(ctx, taskServiceName, etcdRegister)
	if err != nil {
		return
	}
	taskService := service.NewTaskServiceClient(connTask)

	// TODO: 熔断

	ginRouter := routers.NewRouter(userService, taskService)

	//
	//opts := []grpc.DialOption{
	//	grpc.WithInsecure(),
	//}
	//userConn, _ := grpc.Dial("127.0.0.1:10001", opts...)
	//userService := service.NewUserServiceClient(userConn)
	//taskConn, _ := grpc.Dial("127.0.0.1:10002", opts...)
	//taskService := service.NewTaskServiceClient(taskConn)
	//ginRouter := routers.NewRouter(userService, taskService)
	server := &http.Server{
		Addr:           viper.GetString("server.port"),
		Handler:        ginRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("server", server)
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("绑定失败, 可能端口被占用", err)
	}
	go func() {
		//TODO 关闭
		util.GracefullyShutdown(server)
	}()
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("gateway启动失败, err: ", err)
	}
}

func RPCConnect(ctx context.Context, serviceName string, etcdRegister *discovery.Resolver) (conn *grpc.ClientConn, err error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	addr := fmt.Sprintf("%s:///%s", etcdRegister.Scheme(), serviceName)
	fmt.Printf("etcdRegister.Scheme :%s : addr: %s\n", etcdRegister.Scheme(), addr)
	conn, err = grpc.DialContext(ctx, addr, opts...)
	return
}
