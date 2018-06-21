package iface

import (
	"log"
	"github.com/songgao/water"
)

const (
	MTU	= "1300"
)

func New() (*water.Interface, error) {
	config := walter.Config{
		DeviceType: water.TUN,
	}
	iface, err := water.New(config)

	return iface, err
}
