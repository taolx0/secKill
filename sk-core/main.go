package main

import (
	"secKill/sk-core/setup"
)

func main() {
	//setup.InitZk()
	setup.InitEtcd()
	setup.InitRedis()
	setup.RunService()
}
