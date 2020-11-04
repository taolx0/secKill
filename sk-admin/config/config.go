package config

import (
	"github.com/go-kit/kit/log"
	"github.com/openzipkin/zipkin-go"
	zipKinHttp "github.com/openzipkin/zipkin-go/reporter/http"
	_ "github.com/openzipkin/zipkin-go/reporter/recorder"
	"github.com/spf13/viper"
	"github.com/taolx0/secKill/pkg/bootstrap"
	_ "github.com/taolx0/secKill/pkg/bootstrap"
	conf "github.com/taolx0/secKill/pkg/config"
	"os"
)

const (
	kConfigType = "CONFIG_TYPE"
)

var ZipkinTracer *zipkin.Tracer
var Logger log.Logger

func init() {
	Logger = log.NewLogfmtLogger(os.Stderr)
	Logger = log.With(Logger, "ts", log.DefaultTimestampUTC)
	Logger = log.With(Logger, "caller", log.DefaultCaller)
	viper.AutomaticEnv()
	initDefault()

	if err := conf.LoadRemoteConfig(); err != nil {
		_ = Logger.Log("Fail to load remote config", err)
	}

	if err := conf.Sub("mysql", &conf.MysqlConfig); err != nil {
		_ = Logger.Log("Fail to parse mysql", err)
	}
	if err := conf.Sub("trace", &conf.TraceConfig); err != nil {
		_ = Logger.Log("Fail to parse trace", err)
	}
	zipkinUrl := "http://" + conf.TraceConfig.Host + ":" + conf.TraceConfig.Port + conf.TraceConfig.Url
	_ = Logger.Log("zipkin url", zipkinUrl)
}

func initDefault() {
	viper.SetDefault(kConfigType, "yaml")
}

func initTracer(zipkinURL string) {
	var (
		err           error
		useNoopTracer = zipkinURL == ""
		reporter      = zipKinHttp.NewReporter(zipkinURL)
	)
	//defer reporter.Close()
	zEP, _ := zipkin.NewEndpoint(bootstrap.DiscoverConfig.ServiceName, bootstrap.HttpConfig.Port)
	ZipkinTracer, err = zipkin.NewTracer(
		reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer),
	)
	if err != nil {
		_ = Logger.Log("err", err)
		os.Exit(1)
	}
	if !useNoopTracer {
		_ = Logger.Log("tracer", "Zipkin", "type", "Native", "URL", zipkinURL)
	}
}
