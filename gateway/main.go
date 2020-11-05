package main

import (
	"flag"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/go-kit/kit/log"
	"github.com/openzipkin/zipkin-go"
	zipKinHttpSvr "github.com/openzipkin/zipkin-go/middleware/http"
	zipKinHttp "github.com/openzipkin/zipkin-go/reporter/http"
	_ "github.com/taolx0/secKill/gateway/config"
	"github.com/taolx0/secKill/gateway/route"
	"github.com/taolx0/secKill/pkg/bootstrap"
	register "github.com/taolx0/secKill/pkg/discover"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// 创建环境变量
	var (
		zipkinURL = flag.String("zipkin.url", "http://114.67.98.210:9411/api/v2/spans", "Zipkin server url")
	)
	flag.Parse()

	//创建日志组件
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var zipkinTracer *zipkin.Tracer
	{
		var (
			err           error
			useNoopTracer = *zipkinURL == ""
			reporter      = zipKinHttp.NewReporter(*zipkinURL)
		)
		defer reporter.Close()
		zEP, _ := zipkin.NewEndpoint(bootstrap.HttpConfig.Host, bootstrap.HttpConfig.Port)
		zipkinTracer, err = zipkin.NewTracer(
			reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer),
		)
		if err != nil {
			_ = logger.Log("err", err)
			os.Exit(1)
		}
		if !useNoopTracer {
			_ = logger.Log("tracer", "Zipkin", "type", "Native", "URL", *zipkinURL)
		}
	}
	register.Register()

	tags := map[string]string{
		"component": "gateway_server",
	}

	hystrixRouter := route.Routes(zipkinTracer, "Circuit Breaker:Service unavailable", logger)

	handler := zipKinHttpSvr.NewServerMiddleware(
		zipkinTracer,
		zipKinHttpSvr.SpanName(bootstrap.DiscoverConfig.ServiceName),
		zipKinHttpSvr.TagResponseSize(true),
		zipKinHttpSvr.ServerTags(tags),
	)(hystrixRouter)

	err := make(chan error)

	//启用hystrix实时监控，监听端口为9010
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	go func() {
		err <- http.ListenAndServe(net.JoinHostPort("", "9010"), hystrixStreamHandler)
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		err <- fmt.Errorf("%s", <-c)
	}()

	//开始监听
	go func() {
		_ = logger.Log("transport", "HTTP", "addr", "9090")
		register.Register()
		err <- http.ListenAndServe(":9090", handler)
	}()

	// 开始运行，等待结束
	err2 := <-err
	//服务退出取消注册
	register.Deregister()
	_ = logger.Log("exit", err2)
}
