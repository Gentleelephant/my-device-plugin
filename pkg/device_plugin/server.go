package device_plugin

import (
	"context"
	"fmt"
	"github.com/Gentleelephant/my-device-plugin/pkg/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"net"
	"os"
	"path"
	"syscall"
	"time"
)

type GopherDevicePlugin struct {
	server *grpc.Server
	stop   chan struct{}
	dm     *DeviceMonitor
}

func NewGopherDevicePlugin() *GopherDevicePlugin {
	return &GopherDevicePlugin{
		server: grpc.NewServer(grpc.EmptyServerOption{}),
		stop:   make(chan struct{}),
		dm:     NewDeviceMonitor(common.DevicePath),
	}
}

func (gdp *GopherDevicePlugin) Run() error {
	err := gdp.dm.List()
	if err != nil {
		klog.Errorf("failed to list devices: %v", err)
	}

	go func() {
		if err = gdp.dm.Watch(); err != nil {
			klog.Errorf("failed to watch devices: %v", err)
		}
	}()

	pluginapi.RegisterDevicePluginServer(gdp.server, gdp)
	// delete old unix socket
	socket := path.Join(pluginapi.DevicePluginPath, common.DeviceSocket)
	err = syscall.Unlink(socket)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete old socket: %v", err)
	}

	sock, err := net.Listen("unix", socket)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %v", err)
	}

	go gdp.server.Serve(sock)

	conn, err := connect(common.DeviceSocket, common.ConnectTimeout)
	if err != nil {
		return err
	}
	conn.Close()

	return nil
}

// connect
func connect(socket string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c, err := grpc.DialContext(ctx, socket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			if deadline, ok := ctx.Deadline(); ok {
				return net.DialTimeout("unix", addr, time.Until(deadline))
			}
			return net.DialTimeout("unix", addr, common.ConnectTimeout)
		}),
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}
