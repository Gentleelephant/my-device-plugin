package main

import (
	"github.com/Gentleelephant/my-device-plugin/pkg/device_plugin"
	"github.com/Gentleelephant/my-device-plugin/pkg/utils"
	"k8s.io/klog/v2"
)

func main() {

	klog.Infof("device plugin starting")
	dp := device_plugin.NewGopherDevicePlugin()
	go dp.Run()

	if err := dp.Register(); err != nil {
		klog.Fatalf("failed to register device plugin: %v", err)
	}

	stop := make(chan struct{})
	err := utils.WatchKubelet(stop)
	if err != nil {
		klog.Errorf("failed to watch kubelet: %v", err)
	}
	<-stop
	klog.Infof("kubelet restart,exit")
}
