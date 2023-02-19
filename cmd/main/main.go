package main

import (
	"time"

	"github.com/kylestang/kayak/pkg/bridges"
)

func main() {
	group := bridges.NewFacebookGroup("", time.Minute*10)

	err := group.UpdateEntries()
	if err == nil {
		println("Success")
	} else {
		println(err.Error())
	}
}
