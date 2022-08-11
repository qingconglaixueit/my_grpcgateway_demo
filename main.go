package main

import (
	"github.com/golang/glog"
	"mytest/server"
)

func main() {
	defer glog.Flush()
	go server.RunGrpcSvr()
	server.RunGrpcGwWithSwagger()

}
