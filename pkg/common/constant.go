package common

import "time"

const (
	ResourceName   = "gentleelephant.com/gopher"
	DevicePath     = "/etc/gophers"
	DeviceSocket   = "gopher.sock"
	ConnectTimeout = time.Second * 5
)
