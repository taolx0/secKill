package plugins

import (
	"github.com/go-kit/kit/log"
	"secKill/sk-app/model"
	"secKill/sk-app/service"
	"time"
)

// loggingMiddleware Make a new type
// that contains Service interface and logger instance
type skAppLoggingMiddleware struct {
	service.Service
	logger log.Logger
}

// LoggingMiddleware make logging middleware
//func SkAppLoggingMiddleware(logger log.Logger) service.SerMiddleware {
//	return func(next service.Service) service.Service {
//		return skAppLoggingMiddleware{next, logger}
//	}
//}

func (mw skAppLoggingMiddleware) HealthCheck() (result bool) {

	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"function", "HealthCheck",
			"result", result,
			"took", time.Since(begin),
		)
	}(time.Now())

	result = mw.Service.HealthCheck()
	return
}

func (mw skAppLoggingMiddleware) SecInfo(productId int) map[string]interface{} {

	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"function", "Check",
			"took", time.Since(begin),
		)
	}(time.Now())

	ret := mw.Service.SecInfo(productId)
	return ret
}

func (mw skAppLoggingMiddleware) SecInfoList() ([]map[string]interface{}, int, error) {

	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"function", "Check",
			"took", time.Since(begin),
		)
	}(time.Now())

	data, num, err := mw.Service.SecInfoList()
	return data, num, err
}

func (mw skAppLoggingMiddleware) SecKill(req *model.SecRequest) (map[string]interface{}, int, error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"function", "Check",
			"took", time.Since(begin),
		)
	}(time.Now())

	result, num, err := mw.Service.SecKill(req)
	return result, num, err
}
