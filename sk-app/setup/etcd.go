package setup

import (
	"context"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"log"
	conf "secKill/pkg/config"
	"time"
)

func InitEtcd() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"}, // conf.Etcd.Host
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Printf("Connect etcd failed. Error : %v", err)
	}
	conf.Etcd.EtcdSecProductKey = "product"
	conf.Etcd.EtcdConn = client
	loadProductFromEtcd(conf.Etcd.EtcdSecProductKey)
}

func loadProductFromEtcd(key string) {
	log.Println("start get from etcd success")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	rsp, err := conf.Etcd.EtcdConn.Get(ctx, key)
	if err != nil {
		log.Printf("get [%s] from etcd failed, err : %v", key, err)
	}
	log.Printf("get from etcd success, rsp : %v", rsp)

	var secProductInfo []*conf.SecProductInfoConf
	for k, v := range rsp.Kvs {
		log.Printf("key = [%v], value = [%v]", k, v)
		err := json.Unmarshal(v.Value, &secProductInfo)
		if err != nil {
			log.Printf("Unmsharl second product info failed, err : %v", err)
		}
		log.Printf("second info conf is [%v]", secProductInfo)
	}
	updateSecProductInfo(secProductInfo)
}

func updateSecProductInfo(secProductInfo []*conf.SecProductInfoConf) {
	tmp := make(map[int]*conf.SecProductInfoConf, 1024)
	for _, v := range secProductInfo {
		log.Printf("updateSecProductInfo %v", v)
		tmp[v.ProductId] = v
	}
	conf.SecKill.RWBlackLock.Lock()
	conf.SecKill.SecProductInfoMap = tmp
	conf.SecKill.RWBlackLock.Unlock()
}
