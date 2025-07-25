package device_plugin

import (
	"context"
	"github.com/Gentleelephant/my-device-plugin/pkg/common"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"strings"
)

func (gdp *GopherDevicePlugin) GetDevicePluginOptions(_ context.Context, _ *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		PreStartRequired:                true,
		GetPreferredAllocationAvailable: true,
	}, nil
}

func (gdp *GopherDevicePlugin) ListAndWatch(_ *pluginapi.Empty, srv pluginapi.DevicePlugin_ListAndWatchServer) error {
	devs := gdp.dm.Devices()
	klog.Infof("find devices: [%v]", devs)

	// send devices
	err := srv.Send(&pluginapi.ListAndWatchResponse{
		Devices: devs,
	})
	if err != nil {
		klog.Errorf("failed to send device plugin list: %v", err)
	}

	for range gdp.dm.notify {
		devs = gdp.dm.Devices()
		klog.Infof("device update,new device list: [%v]", devs)
		// send devices
		err = srv.Send(&pluginapi.ListAndWatchResponse{
			Devices: devs,
		})
		if err != nil {
			klog.Errorf("failed to send device plugin list: %v", err)
		}
	}
	return nil
}

// GetPreferredAllocation
func (gdp *GopherDevicePlugin) GetPreferredAllocation(_ context.Context, req *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	klog.Infoln("[GetPreferredAllocation] is called")
	return &pluginapi.PreferredAllocationResponse{}, nil
}

// Allocate
func (gdp *GopherDevicePlugin) Allocate(_ context.Context, req *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	klog.Infoln("[Allocate] is called")
	ret := &pluginapi.AllocateResponse{}
	for _, r := range req.ContainerRequests {
		klog.Infof("request device: [%v]", strings.Join(r.GetDevicesIDs(), ","))
		klog.Infof("request container: [%v]", r.Size())

		mounts := make([]*pluginapi.Mount, 0)
		for _, s := range r.GetDevicesIDs() {
			mounts = append(mounts, &pluginapi.Mount{
				ContainerPath: common.DevicePath + s,
				HostPath:      common.DevicePath + s,
			})
		}

		response := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				"Gopher": strings.Join(r.DevicesIDs, ","),
			},
			Mounts: mounts,
		}
		ret.ContainerResponses = append(ret.ContainerResponses, &response)
	}
	return ret, nil
}

// PreStartContainer
func (gdp *GopherDevicePlugin) PreStartContainer(_ context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	klog.Infoln("[PreStartContainer] is called")
	return &pluginapi.PreStartContainerResponse{}, nil
}
