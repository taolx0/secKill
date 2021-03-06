package plugins

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/time/rate"
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
