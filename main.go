package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/spider-docker/hypervisor"
)

func init() {
	/*
	**To configure glog
	 */
	flag.Set("logtostderr", "true")
	flag.Set("log_dir", ".")
	flag.Set("V", "3")
	flag.Parse()
}

func main() {
	glog.Infoln("start run spidermain")
	hypervisor.Main()
}
