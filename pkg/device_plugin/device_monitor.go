package device_plugin

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type DeviceMonitor struct {
	path    string
	devices map[string]*pluginapi.Device
	notify  chan struct{}
}

func NewDeviceMonitor(path string) *DeviceMonitor {
	return &DeviceMonitor{
		path:    path,
		devices: make(map[string]*pluginapi.Device),
		notify:  make(chan struct{}),
	}
}

// List
func (dm *DeviceMonitor) List() error {
	err := filepath.WalkDir(dm.path, func(path string, d os.DirEntry, err error) error {
		if d.IsDir() {
			klog.Infof("%s is dir, skip", path)
			return nil
		}
		dm.devices[d.Name()] = &pluginapi.Device{
			ID:     d.Name(),
			Health: pluginapi.Healthy,
		}
		return nil
	})
	return fmt.Errorf("failed to list devices: %v", err)
}

// Watch device change
func (dm *DeviceMonitor) Watch() error {
	klog.Infoln("Watching device change")

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to watch devices: %v", err)
	}
	defer w.Close()

	errChan := make(chan error)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				errChan <- fmt.Errorf("device monitor panic: %v", err)
			}
		}()
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					continue
				}
				if event.Op == fsnotify.Create {
					dev := path.Base(event.Name)
					dm.devices[dev] = &pluginapi.Device{
						ID:     dev,
						Health: pluginapi.Healthy,
					}
					dm.notify <- struct{}{}
					klog.Infof("find new device: %s", event.Name)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					dev := path.Base(event.Name)
					delete(dm.devices, dev)
					dm.notify <- struct{}{}
					klog.Infof("remove device: %s", event.Name)
				}
			case err, ok := <-w.Errors:
				if !ok {
					continue
				}
				klog.Errorf("failed to watch devices: %v", err)
			}
		}
	}()

	err = w.Add(dm.path)
	if err != nil {
		return fmt.Errorf("failed to watch devices: %v", err)
	}
	return <-errChan
}

// Devices transform
func (dm *DeviceMonitor) Devices() []*pluginapi.Device {
	devices := make([]*pluginapi.Device, 0, len(dm.devices))
	for _, v := range dm.devices {
		devices = append(devices, v)
	}
	return devices
}

// String
func String(devs []*pluginapi.Device) string {
	strs := make([]string, 0, len(devs))
	for _, d := range devs {
		strs = append(strs, d.ID)
	}
	return strings.Join(strs, ",")
}
