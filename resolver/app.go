package resolver

import (
	"net"
)

type AppInstanceInfo struct {
	ContainerIP   net.IP
	Guid          string
	InstanceIndex uint
}

type Resolver interface {
	Resolve(ip net.IP) (*AppInstanceInfo, error)
}

func (ai *AppInstanceInfo) String() string {
	return ai.Guid
}
