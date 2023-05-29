package discovery

import (
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc/resolver"
	"strings"
)

type Server struct {
	Name    string `json:"name"`
	Addr    string `json:"addr"`
	Version string `json:"version"` // 版本
	Weight  int64  `json:"weight"`  // 服务的权重
}

func BuildPrefix(server Server) string {
	if server.Version == "" {
		return fmt.Sprintf("/%s/", server.Name)
	}
	return fmt.Sprintf("/%s/%s/", server.Name, server.Version)
}

// BuildRegisterPath 找到唯一的注册路径
func BuildRegisterPath(server Server) string {
	return fmt.Sprintf("%s%s", BuildPrefix(server), server.Addr)
}

// ParseValue 将value 值反序列化到一个server实例中
func ParseValue(value []byte) (Server, error) {
	server := Server{}
	if err := json.Unmarshal(value, &server); err != nil {
		return server, err
	}
	return server, nil
}

// SplitPath 切割路径
func SplitPath(path string) (Server, error) {

	server := Server{}
	strs := strings.Split(path, "/")
	if len(strs) == 0 {
		return server, errors.New("invalid path")
	}
	server.Addr = strs[len(strs)-1]
	return server, nil
}

// Exist 判断路径是否存在
func Exist(l []resolver.Address, addr resolver.Address) bool {
	for i := range l {
		if l[i].Addr == addr.Addr {
			return true
		}
	}
	return false
}

func Remove(l []resolver.Address, addr resolver.Address) ([]resolver.Address, bool) {
	for i := range l {
		if l[i].Addr == addr.Addr {
			l[i] = l[len(l)-1]
			return l[:len(l)-1], true
		}
	}
	return nil, false
}

func BuildResolverUrl(app string) string {
	return schema + ":///" + app
}
