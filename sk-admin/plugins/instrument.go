package plugins

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"github.com/gohouse/gorose/v2"
	"golang.org/x/time/rate"
	"secKill/sk-admin/model"
	"secKill/sk-admin/service"
	"time"
)

var ErrLimitExceed = errors.New("rate limit exceed")

// NewTokenBucketLimiterWithJuju 使用juju/rateLimit创建限流中间件
//func NewTokenBucketLimiterWithJuju(bkt *ratelimit.Bucket) endpoint.Middleware {
//	return func(next endpoint.Endpoint) endpoint.Endpoint {
//		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
//			if bkt.TakeAvailable(1) == 0 {
//				return nil, ErrLimitExceed
//			}
//			return next(ctx, request)
//		}
//	}
//}

// NewTokenBucketLimiterWithBuildIn 使用x/time/rate创建限流中间件
func NewTokenBucketLimiterWithBuildIn(bkt *rate.Limiter) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if !bkt.Allow() {
				return nil, ErrLimitExceed
			}
			return next(ctx, request)
		}
	}
}

// metricMiddleware 定义监控中间件，嵌入Service
// 新增监控指标项：requestCount和requestLatency
type skAdminMetricMiddleware struct {
	service.Service
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
}

type activityMetricMiddleware struct {
	service.ActivityService
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
}

type productMetricMiddleware struct {
	service.ProductService
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
}

// Metrics 封装监控方法
func SkAdminMetrics(requestCount metrics.Counter, requestLatency metrics.Histogram) service.ServiceMiddleware {
	return func(next service.Service) service.Service {
		return skAdminMetricMiddleware{
			next,
			requestCount,
			requestLatency}
	}
}

// Metrics 封装监控方法
func ProductMetrics(requestCount metrics.Counter, requestLatency metrics.Histogram) service.ProductServiceMiddleware {
	return func(next service.ProductService) service.ProductService {
		return productMetricMiddleware{
			next,
			requestCount,
			requestLatency}
	}
}

// Metrics 封装监控方法
func ActivityMetrics(requestCount metrics.Counter, requestLatency metrics.Histogram) service.ActivityServiceMiddleware {
	return func(next service.ActivityService) service.ActivityService {
		return activityMetricMiddleware{
			next,
			requestCount,
			requestLatency}
	}
}

func (mw skAdminMetricMiddleware) HealthCheck() (result bool) {

	defer func(begin time.Time) {
		lvs := []string{"method", "HealthCheck"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	result = mw.Service.HealthCheck()
	return
}

func (mw productMetricMiddleware) CreateProduct(product *model.Product) error {

	defer func(begin time.Time) {
		lvs := []string{"method", "HealthCheck"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err := mw.ProductService.CreateProduct(product)
	return err
}

func (mw productMetricMiddleware) GetProductList() ([]gorose.Data, error) {

	defer func(begin time.Time) {
		lvs := []string{"method", "HealthCheck"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	data, err := mw.ProductService.GetProductList()
	return data, err
}

func (mw activityMetricMiddleware) GetActivityList() ([]gorose.Data, error) {

	defer func(begin time.Time) {
		lvs := []string{"method", "HealthCheck"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	result, err := mw.ActivityService.GetActivityList()
	return result, err
}

func (mw activityMetricMiddleware) CreateActivity(activity *model.Activity) error {

	defer func(begin time.Time) {
		lvs := []string{"method", "HealthCheck"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	err := mw.ActivityService.CreateActivity(activity)
	return err
}
