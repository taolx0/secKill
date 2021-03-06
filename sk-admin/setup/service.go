package setup

import (
	"context"
	"flag"
	"fmt"
	kitPrometheus "github.com/go-kit/kit/metrics/prometheus"
	kitZipkin "github.com/go-kit/kit/tracing/zipkin"
	stdPrometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"os"
	"os/signal"
	register "secKill/pkg/discover"
	"secKill/sk-admin/endpoint"
	"secKill/sk-admin/plugins"
	"secKill/sk-admin/service"
	"secKill/sk-admin/transport"
	"secKill/user-service/config"
	"syscall"
	"time"
)

//初始化Http服务
func InitServer(serviceHost string, servicePort string) {
	log.Printf("initial service , port is %s\n", servicePort)

	flag.Parse()

	errChan := make(chan error)

	fieldKeys := []string{"method"}

	requestCount := kitPrometheus.NewCounterFrom(stdPrometheus.CounterOpts{
		Namespace: "Tommy",
		Subsystem: "user_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)

	requestLatency := kitPrometheus.NewSummaryFrom(stdPrometheus.SummaryOpts{
		Namespace: "Tommy",
		Subsystem: "user_service",
		Name:      "request_latency",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)
	rateBucket := rate.NewLimiter(rate.Every(time.Second*1), 100)

	var (
		activityService service.ActivityService
		productService  service.ProductService
		skAdminService  service.Service
	)

	skAdminService = service.SkAdminService{}
	activityService = service.ActivityServiceImpl{}
	productService = service.ProductServiceImpl{}

	// add logging middleware
	skAdminService = plugins.SkAdminLoggingMiddleware(config.Logger)(skAdminService)
	skAdminService = plugins.SkAdminMetrics(requestCount, requestLatency)(skAdminService)

	activityService = plugins.ActivityLoggingMiddleware(config.Logger)(activityService)
	activityService = plugins.ActivityMetrics(requestCount, requestLatency)(activityService)

	productService = plugins.ProductLoggingMiddleware(config.Logger)(productService)
	productService = plugins.ProductMetrics(requestCount, requestLatency)(productService)

	createActivityEnd := endpoint.MakeCreateActivityEndpoint(activityService)
	createActivityEnd = plugins.NewTokenBucketLimiterWithBuildIn(rateBucket)(createActivityEnd)
	createActivityEnd = kitZipkin.TraceEndpoint(config.ZipkinTracer, "create-activity")(createActivityEnd)

	GetActivityEnd := endpoint.MakeGetActivityEndpoint(activityService)
	GetActivityEnd = plugins.NewTokenBucketLimiterWithBuildIn(rateBucket)(GetActivityEnd)
	GetActivityEnd = kitZipkin.TraceEndpoint(config.ZipkinTracer, "get-activity")(GetActivityEnd)

	createProductEnd := endpoint.MakeCreateProductEndpoint(productService)
	createProductEnd = plugins.NewTokenBucketLimiterWithBuildIn(rateBucket)(createProductEnd)
	createProductEnd = kitZipkin.TraceEndpoint(config.ZipkinTracer, "create-product")(createProductEnd)

	GetProductEnd := endpoint.MakeGetProductEndpoint(productService)
	GetProductEnd = plugins.NewTokenBucketLimiterWithBuildIn(rateBucket)(GetProductEnd)
	GetProductEnd = kitZipkin.TraceEndpoint(config.ZipkinTracer, "get-product")(GetProductEnd)

	//创建健康检查的Endpoint
	healthEndpoint := endpoint.MakeHealthCheckEndpoint(skAdminService)
	healthEndpoint = kitZipkin.TraceEndpoint(config.ZipkinTracer, "health-endpoint")(healthEndpoint)

	endpoints := endpoint.SkAdminEndpoints{
		GetActivityEndpoint:    GetActivityEnd,
		CreateActivityEndpoint: createActivityEnd,
		CreateProductEndpoint:  createProductEnd,
		GetProductEndpoint:     GetProductEnd,
		HealthCheckEndpoint:    healthEndpoint,
	}
	ctx := context.Background()
	//创建http.Handler
	r := transport.MakeHttpHandler(ctx, endpoints, config.ZipkinTracer, config.Logger)

	//http server
	go func() {
		fmt.Println("Http Server start at port:" + servicePort)
		//启动前执行注册
		register.Register()
		handler := r
		errChan <- http.ListenAndServe(serviceHost+":"+servicePort, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	err := <-errChan
	//服务退出取消注册
	register.Deregister()
	log.Println("sk-admin service deregister")
	fmt.Println(err)
}
