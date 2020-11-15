package srv_redis

import (
	"crypto/md5"
	"fmt"
	"log"
	conf "secKill/pkg/config"
	"secKill/sk-core/config"
	"secKill/sk-core/service/srv_err"
	"secKill/sk-core/service/srv_user"
	"time"
)

func HandleUser() {
	log.Println("handle user running")
	for req := range config.SecLayerCtx.Read2HandleChan {
		log.Printf("begin process request : %v", req)
		resp, err := HandleSecKill(req)
		if err != nil {
			log.Printf("process request %v failed, err : %v", req, err)
			resp = &config.SecResult{
				Code: srv_err.ErrServiceBusy,
			}
		}
		fmt.Println("request handling ... ", resp)
		timer := time.NewTicker(time.Millisecond * time.Duration(conf.SecKill.SendToWriteChanTimeout))
		select {
		case config.SecLayerCtx.Handle2WriteChan <- resp:
		case <-timer.C:
			log.Printf("send to response chan timeout, resp : %v", resp)
			break
		}
	}
	return
}

func HandleSecKill(req *config.SecRequest) (res *config.SecResult, err error) {
	config.SecLayerCtx.RWSecProductLock.RLock()
	defer config.SecLayerCtx.RWSecProductLock.RUnlock()

	res = &config.SecResult{}
	res.ProductId = req.ProductId
	res.UserId = req.UserId

	product, ok := conf.SecKill.SecProductInfoMap[req.ProductId]
	if !ok {
		log.Printf("not found product : %v", req.ProductId)
		res.Code = srv_err.ErrNotFoundProduct
		return
	}

	if product.Status == srv_err.ProductStatusSoldOut {
		res.Code = srv_err.ErrSoldOut
		return
	}
	nowTime := time.Now().Unix()

	config.SecLayerCtx.HistoryMapLock.Lock()
	userHistory, ok := config.SecLayerCtx.HistoryMap[req.UserId]
	if !ok {
		userHistory = &srv_user.UserBuyHistory{
			History: make(map[int]int, 16),
		}
		config.SecLayerCtx.HistoryMap[req.UserId] = userHistory
	}
	historyCount := userHistory.GetProductBuyCount(req.ProductId)
	config.SecLayerCtx.HistoryMapLock.Unlock()

	if historyCount >= product.OnePersonBuyLimit {
		res.Code = srv_err.ErrAlreadyBuy
		return
	}

	curSoldCount := config.SecLayerCtx.ProductCountMgr.Count(req.ProductId)

	if curSoldCount >= product.Total {
		res.Code = srv_err.ErrSoldOut
		product.Status = srv_err.ProductStatusSoldOut
		return
	}

	//curRate := rand.Float64()
	curRate := 0.1
	fmt.Println(curRate, product.BuyRate)
	if curRate > product.BuyRate {
		res.Code = srv_err.ErrRetry
		return
	}

	userHistory.Add(req.ProductId, 1)
	config.SecLayerCtx.ProductCountMgr.Add(req.ProductId, 1)

	//用户Id、商品id、当前时间、密钥

	res.Code = srv_err.ErrSecKillSuccess
	tokenData := fmt.Sprintf("userId=%d&productId=%d&timestamp=%d&security=%s", req.UserId, req.ProductId, nowTime, conf.SecKill.TokenPassWd)
	res.Token = fmt.Sprintf("%x", md5.Sum([]byte(tokenData))) //MD5加密
	res.TokenTime = nowTime

	return
}
