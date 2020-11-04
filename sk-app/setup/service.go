package setup

import (
	"context"
	"flag"
	"fmt"
	//kitPrometheus "github.com/go-kit/kit/metrics/prometheus"
	kitZipkin "github.com/go-kit/kit/tracing/zipkin"
	localConfig "github.com/longjoy/micro-go-book/ch13-seckill/pkg/config"
	register "github.com/longjoy/micro-go-book/ch13-seckill/pkg/discover"
	"github.com/longjoy/micro-go-book/ch13-seckill/sk-app/endpoint"
	"github.com/longjoy/micro-go-book/ch13-seckill/sk-app/plugins"
	"github.com/longjoy/micro-go-book/ch13-seckill/sk-app/service"
	"github.com/longjoy/micro-go-book/ch13-seckill/sk-app/transport"
	//stdPrometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//初始化Http服务
func InitServer(host string, servicePort string) {
	log.Println("host is:", host)
	log.Println("port is:", servicePort)

	flag.Parse()

	errChan := make(chan error)

	//fieldKeys := []string{"method"}

	//requestCount := kitPrometheus.NewCounterFrom(stdPrometheus.CounterOpts{
	//	Namespace: "aoho",
	//	Subsystem: "sk_app",
	//	Name:      "request_count",
	//	Help:      "Number of requests received.",
	//}, fieldKeys)
	//
	//requestLatency := kitPrometheus.NewSummaryFrom(stdPrometheus.SummaryOpts{
	//	Namespace: "aoho",
	//	Subsystem: "sk_app",
	//	Name:      "request_latency",
	//	Help:      "Total duration of requests in microseconds.",
	//}, fieldKeys)

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
	healthCheckEnd = kitZipkin.TraceEndpoint(localConfig.ZipkinTracer, "heath-check")(healthCheckEnd)

	GetSecInfoEnd := endpoint.MakeSecInfoEndpoint(skAppService)
	GetSecInfoEnd = plugins.NewTokenBucketLimiterWithBuildIn(rateBucket)(GetSecInfoEnd)
	GetSecInfoEnd = kitZipkin.TraceEndpoint(localConfig.ZipkinTracer, "sec-info")(GetSecInfoEnd)

	GetSecInfoListEnd := endpoint.MakeSecInfoListEndpoint(skAppService)
	GetSecInfoListEnd = plugins.NewTokenBucketLimiterWithBuildIn(rateBucket)(GetSecInfoListEnd)
	GetSecInfoListEnd = kitZipkin.TraceEndpoint(localConfig.ZipkinTracer, "sec-info-list")(GetSecInfoListEnd)

	//秒杀单独限流
	secRateBucket := rate.NewLimiter(rate.Every(time.Microsecond*100), 1000)

	SecKillEnd := endpoint.MakeSecKillEndpoint(skAppService)
	SecKillEnd = plugins.NewTokenBucketLimiterWithBuildIn(secRateBucket)(SecKillEnd)
	//SecKillEnd = kitZipkin.TraceEndpoint(localConfig.ZipkinTracer, "sec-kill")(SecKillEnd)

	testEnd := endpoint.MakeTestEndpoint(skAppService)
	testEnd = kitZipkin.TraceEndpoint(localConfig.ZipkinTracer, "test")(testEnd)

	endpoints := endpoint.SkAppEndpoints{
		SecKillEndpoint:        SecKillEnd,
		HeathCheckEndpoint:     healthCheckEnd,
		GetSecInfoEndpoint:     GetSecInfoEnd,
		GetSecInfoListEndpoint: GetSecInfoListEnd,
		TestEndpoint:           testEnd,
	}
	ctx := context.Background()
	//创建http.Handler
	r := transport.MakeHttpHandler(ctx, endpoints, localConfig.ZipkinTracer, localConfig.Logger)

	//http server
	go func() {
		fmt.Println("Http Server start at port:" + servicePort)
		//启动前执行注册
		register.Register()
		handler := r
		errChan <- http.ListenAndServe(":"+servicePort, handler)
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
