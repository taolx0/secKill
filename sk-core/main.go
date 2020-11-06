package main

import (
	"secKill/sk-core/setup"
)

func main() {
	setup.InitZk()
	setup.InitRedis()
	setup.RunService()
}
