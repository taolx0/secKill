package setup

import (
	"fmt"
	"os"
	"os/signal"
	register "secKill/pkg/discover"
	"secKill/sk-core/service/srv_redis"
	"syscall"
)

func RunService() {
	//启动处理线程
	srv_redis.RunProcess()
	errChan := make(chan error)
	//http server
	go func() {
		register.Register()
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	err := <-errChan
	//服务退出取消注册
	register.Deregister()
	fmt.Println(err)
}
