package main

import (
	"github.com/taolx0/secKill/sk-core/setup"
)

func main() {
	setup.InitZk()
	setup.InitRedis()
	setup.RunService()
}
