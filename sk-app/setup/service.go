package setup

import (
	"context"
	"flag"
	"fmt"
	kitZipkin "github.com/go-kit/kit/tracing/zipkin"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	conf "secKill/pkg/config"
	register "secKill/pkg/discover"
	"secKill/sk-app/endpoint"
	"secKill/sk-app/plugins"
	"secKill/sk-app/service"
	"secKill/sk-app/transport"
	"syscall"
	"time"
)

//初始化Http服务
func InitServer(host string, servicePort string) {
	log.Println("host is:", host)
	log.Println("port is:", servicePort)

	flag.Parse()

	errChan := make(chan error)

	rateBucket := rate.NewLimiter(rate.Every(time.Second*1), 5000)

	var (
		skAppService service.Service
	)
	skAppService = service.SkAppService{}

	// add logging middleware
	//skAppService = plugins.SkAppLoggingMiddleware(config.Logger)(skAppService)
	//skAppService = plugins.SkAppMetrics(requestCount, requestLatency)(skAppService)

	healthCheckEnd := endpoint.MakeHealthCheckEndpoint(skAppService)
	healthCheckEnd = plugins.NewTokenBucketLimiterWithBuildIn(rateBucket)(healthCheckEnd)
	healthCheckEnd = kitZipkin.TraceEndpoint(conf.ZipkinTracer, "heath-check")(healthCheckEnd)

	GetSecInfoEnd := endpoint.MakeSecInfoEndpoint(skAppService)
	GetSecInfoEnd = plugins.NewTokenBucketLimiterWithBuildIn(rateBucket)(GetSecInfoEnd)
	GetSecInfoEnd = kitZipkin.TraceEndpoint(conf.ZipkinTracer, "sec-info")(GetSecInfoEnd)

	GetSecInfoListEnd := endpoint.MakeSecInfoListEndpoint(skAppService)
	GetSecInfoListEnd = plugins.NewTokenBucketLimiterWithBuildIn(rateBucket)(GetSecInfoListEnd)
	GetSecInfoListEnd = kitZipkin.TraceEndpoint(conf.ZipkinTracer, "sec-info-list")(GetSecInfoListEnd)

	//秒杀单独限流
	secRateBucket := rate.NewLimiter(rate.Every(time.Microsecond*100), 1000)

	SecKillEnd := endpoint.MakeSecKillEndpoint(skAppService)
	SecKillEnd = plugins.NewTokenBucketLimiterWithBuildIn(secRateBucket)(SecKillEnd)
	//SecKillEnd = kitZipkin.TraceEndpoint(conf.ZipkinTracer, "sec-kill")(SecKillEnd)

	testEnd := endpoint.MakeTestEndpoint(skAppService)
	testEnd = kitZipkin.TraceEndpoint(conf.ZipkinTracer, "test")(testEnd)

	endpoints := endpoint.SkAppEndpoints{
		SecKillEndpoint:        SecKillEnd,
		HeathCheckEndpoint:     healthCheckEnd,
		GetSecInfoEndpoint:     GetSecInfoEnd,
		GetSecInfoListEndpoint: GetSecInfoListEnd,
		TestEndpoint:           testEnd,
	}
	ctx := context.Background()
	//创建http.Handler
	r := transport.MakeHttpHandler(ctx, endpoints, conf.ZipkinTracer, conf.Logger)

	//http server
	go func() {
		fmt.Println("Http Server start at port:" + servicePort)
		//启动前执行注册
		register.Register()
		handler := r
		errChan <- http.ListenAndServe(host+":"+servicePort, handler)
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
