package main

import (
	"context"
	"flag"
	"fmt"
	kitZipkin "github.com/go-kit/kit/tracing/zipkin"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"net/http"
	"os"
	"os/signal"
	localConfig "secKill/oauth-service/config"
	"secKill/oauth-service/endpoint"
	"secKill/oauth-service/plugins"
	"secKill/oauth-service/service"
	"secKill/oauth-service/transport"
	"secKill/pb"
	"secKill/pkg/bootstrap"
	conf "secKill/pkg/config"
	register "secKill/pkg/discover"
	"secKill/pkg/mysql"
	"syscall"
	"time"
)

func main() {
	var (
		servicePort = flag.String("service.port", bootstrap.HttpConfig.Port, "service port")
		grpcAddr    = flag.String("grpc", bootstrap.RpcConfig.Port, "gRPC listen address.")
	)

	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	rateBucket := rate.NewLimiter(rate.Every(time.Second*1), 100)

	var tokenService service.TokenService
	var tokenGranter service.TokenGranter
	var tokenEnhancer service.TokenEnhancer
	var tokenStore service.TokenStore
	var userDetailsService service.UserDetailsService
	var clientDetailsService service.ClientDetailsService
	var srv service.Service

	// add logging middleware

	tokenEnhancer = service.NewJwtTokenEnhancer("secret")
	tokenStore = service.NewJwtTokenStore(tokenEnhancer.(*service.JwtTokenEnhancer))
	tokenService = service.NewTokenService(tokenStore, tokenEnhancer)
	userDetailsService = service.NewRemoteUserDetailService()
	clientDetailsService = service.NewMysqlClientDetailsService()
	srv = service.NewCommentService()

	tokenGranter = service.NewComposeTokenGranter(map[string]service.TokenGranter{
		"password":      service.NewUsernamePasswordTokenGranter("password", userDetailsService, tokenService),
		"refresh_token": service.NewRefreshGranter("refresh_token", userDetailsService, tokenService),
	})

	tokenEndpoint := endpoint.MakeTokenEndpoint(tokenGranter, clientDetailsService)
	tokenEndpoint = endpoint.MakeClientAuthorizationMiddleware(localConfig.Logger)(tokenEndpoint)
	tokenEndpoint = plugins.NewTokenBucketLimiterWithBuildIn(rateBucket)(tokenEndpoint)
	tokenEndpoint = kitZipkin.TraceEndpoint(localConfig.ZipkinTracer, "token-endpoint")(tokenEndpoint)
	//tokenEndpoint = plugins.ClientAuthorizationMiddleware(clientDetailsService)(tokenEndpoint)

	checkTokenEndpoint := endpoint.MakeCheckTokenEndpoint(tokenService)
	checkTokenEndpoint = endpoint.MakeClientAuthorizationMiddleware(localConfig.Logger)(checkTokenEndpoint)
	checkTokenEndpoint = plugins.NewTokenBucketLimiterWithBuildIn(rateBucket)(checkTokenEndpoint)
	checkTokenEndpoint = kitZipkin.TraceEndpoint(localConfig.ZipkinTracer, "check-endpoint")(checkTokenEndpoint)
	//tokenEndpoint = plugins.ClientAuthorizationMiddleware(clientDetailsService)(checkTokenEndpoint)

	gRPCCheckTokenEndpoint := endpoint.MakeCheckTokenEndpoint(tokenService)
	gRPCCheckTokenEndpoint = plugins.NewTokenBucketLimiterWithBuildIn(rateBucket)(gRPCCheckTokenEndpoint)
	gRPCCheckTokenEndpoint = kitZipkin.TraceEndpoint(localConfig.ZipkinTracer, "grpc-check-endpoint")(gRPCCheckTokenEndpoint)
	//tokenEndpoint = plugins.ClientAuthorizationMiddleware(clientDetailsService)(checkTokenEndpoint)

	//创建健康检查的Endpoint
	healthEndpoint := endpoint.MakeHealthCheckEndpoint(srv)
	healthEndpoint = kitZipkin.TraceEndpoint(localConfig.ZipkinTracer, "health-endpoint")(healthEndpoint)

	endpoints := endpoint.OAuth2Endpoints{
		TokenEndpoint:          tokenEndpoint,
		CheckTokenEndpoint:     checkTokenEndpoint,
		HealthCheckEndpoint:    healthEndpoint,
		GRPCCheckTokenEndpoint: gRPCCheckTokenEndpoint,
	}

	//创建http.Handler
	r := transport.MakeHttpHandler(ctx, endpoints, tokenService, clientDetailsService, localConfig.ZipkinTracer, localConfig.Logger)

	//http server
	go func() {
		fmt.Println("Http Server start at port:" + *servicePort)
		mysql.InitMysql(conf.MysqlConfig.Host, conf.MysqlConfig.Port, conf.MysqlConfig.User, conf.MysqlConfig.Pwd, conf.MysqlConfig.Db)
		//启动前执行注册
		register.Register()
		handler := r
		errChan <- http.ListenAndServe(":"+*servicePort, handler)
	}()

	//grpc server
	go func() {
		fmt.Println("grpc Server start at port:" + *grpcAddr)
		listener, err := net.Listen("tcp", ":"+*grpcAddr)
		if err != nil {
			errChan <- err
			return
		}
		serverTracer := kitZipkin.GRPCServerTrace(localConfig.ZipkinTracer, kitZipkin.Name("grpc-transport"))
		tr := localConfig.ZipkinTracer
		md := metadata.MD{}
		parentSpan := tr.StartSpan("test")

		_ = b3.InjectGRPC(&md)(parentSpan.Context())

		ctx := metadata.NewIncomingContext(context.Background(), md)
		handler := transport.NewGRPCServer(ctx, endpoints, serverTracer)
		gRPCServer := grpc.NewServer()
		pb.RegisterOAuthServiceServer(gRPCServer, handler)
		errChan <- gRPCServer.Serve(listener)
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
