package utils

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func WatchKubelet(stop chan<- struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create fsnotify watcher: %v", err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}
				klog.Infof("fsnotify event: [%s]  [%v]", event.Name, event.Op.String())
				if event.Name == pluginapi.KubeletSocket && event.Op == fsnotify.Create {
					klog.Warning("inotify event: kubelet socket created, restarting")
					stop <- struct{}{}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				klog.Errorf("fsnotify error: %v", err)
			}
		}

	}()

	err = watcher.Add(pluginapi.KubeletSocket)
	if err != nil {
		return fmt.Errorf("failed to add kubelet socket to fsnotify watcher: %v", err)
	}

	return nil

}
