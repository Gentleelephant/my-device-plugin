package device_plugin

import (
	"context"
	"github.com/Gentleelephant/my-device-plugin/pkg/common"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"path"
)

// Register
func (gdp *GopherDevicePlugin) Register() error {
	conn, err := connect(pluginapi.KubeletSocket, common.ConnectTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	request := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(common.DeviceSocket),
		ResourceName: common.ResourceName,
	}
	_, err = client.Register(context.Background(), request)
	if err != nil {
		klog.Errorf("failed to register device plugin: %v", err)
	}
	return nil
}
