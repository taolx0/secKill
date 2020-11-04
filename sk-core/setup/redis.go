package setup

import (
	"github.com/go-redis/redis"
	config "github.com/taolx0/secKill/pkg/config"
	"log"
)

//初始化redis
func InitRedis() {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host,
		Password: config.Redis.Password,
		DB:       config.Redis.Db,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Printf("Connect redis failed. Error : %v", err)
	}
	config.Redis.RedisConn = client
}
