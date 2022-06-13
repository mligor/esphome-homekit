package main

import (
	esphomehomekit "github.com/mligor/esphome-homekit"
)

func main() {

	svc := esphomehomekit.New()
	svc.Start()
}
