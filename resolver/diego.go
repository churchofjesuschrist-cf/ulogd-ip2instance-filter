// +build diego

package resolver

import (
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
	"fmt"
	"log"
	"net"
)

type diegoResolver struct {
	client garden.Client
}

var theDiegoResolver *diegoResolver = &diegoResolver{}

func init() {
	// TODO: use the ulogd stack configuration to locate the garden socket
	log.Printf("init() for diego")
	conn := connection.New("unix", "/var/vcap/data/garden/garden.sock")
	theDiegoResolver.client = client.New(conn)
}

func Get() Resolver {
	return theDiegoResolver
}

func (dr *diegoResolver) Resolve(ip net.IP) (*AppInstanceInfo, error) {
	// log.Printf("Resolve() for diego")

	ipStr := ip.String()

	props := make(garden.Properties)
	allContainers, err := dr.client.Containers(props)
	if nil != err {
		return nil, fmt.Errorf("getting containers from garden Client: %v", err)
	}

	for _, c := range allContainers {
		containerInfo, err := c.Info()
		if nil != err {
			log.Printf("getting Info from a Container: %v", err)
			continue
		}

		if ipStr == containerInfo.ContainerIP {
			// log.Printf("c.Handle() = %v", c.Handle())
			// log.Printf("c.Properties():")
			p, _ := c.Properties()
			// logProperties(p)

			return &AppInstanceInfo{
				ContainerIP:   ip,
				Guid:          p["network.app_id"],
				InstanceIndex: 0,
			}, nil
		}
	}

	return nil, fmt.Errorf("no container found matching IP=%v", ip)
}

func logProperties(allProps garden.Properties) {
	log.Printf("(container properties)")
	for k, v := range allProps {
		log.Printf("%v:%v\n", k, v)
	}
	log.Printf("(end container properties)")
}
