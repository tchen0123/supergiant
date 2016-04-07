package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

type port struct {
	*common.Port
	release *ReleaseResource
}

func (p *port) name() string {
	return strconv.Itoa(p.Number)
}

type InternalPort struct {
	*port
}

func NewInternalPort(p *common.Port, r *ReleaseResource) *InternalPort {
	return &InternalPort{
		port: &port{p, r},
	}
}

// a method because this will change on Release
func (ip *InternalPort) service() *guber.Service {
	return ip.release.InternalService
}

func (ip *InternalPort) Address() *common.PortAddress {
	svcMeta := ip.service().Metadata
	host := fmt.Sprintf("%s.%s.svc.cluster.local", svcMeta.Name, svcMeta.Namespace)
	return &common.PortAddress{
		Port:    ip.name(),
		Address: fmt.Sprintf("%s://%s:%d", strings.ToLower(ip.Protocol), host, ip.Number),
	}
}

//==============================================================================
//==============================================================================
//==============================================================================
//==============================================================================
//==============================================================================
//==============================================================================
//==============================================================================
//==============================================================================

type ExternalPort struct {
	*port
	entrypoint *EntrypointResource
}

// NOTE we pass entrypoint here, instead of simply finding from the port
// definition because it prevents unnecessary multiple lookups on the Entrypoint
func NewExternalPort(p *common.Port, r *ReleaseResource, e *EntrypointResource) *ExternalPort {
	return &ExternalPort{
		port:       &port{p, r},
		entrypoint: e,
	}
}

// a method because this will change on Release
func (ep *ExternalPort) service() *guber.Service {
	return ep.release.ExternalService
}

func (ep *ExternalPort) nodePort() int {
	for _, port := range ep.service().Spec.Ports {
		if port.Port == ep.Number {
			return port.NodePort
		}
	}
	panic("Could not find NodePort")
}

func (ep *ExternalPort) elbPort() int {
	if ep.PreserveNumber {
		return ep.Number
	}
	return ep.nodePort()
}

func (ep *ExternalPort) Address() *common.PortAddress {
	// TODO
	//
	// Current it is assumed that all external ports have an entrypoint, which is
	// not technically true. We should return a random node IP if there is no
	// entrypoint.
	return &common.PortAddress{
		Port:    ep.name(),
		Address: fmt.Sprintf("%s://%s:%d", strings.ToLower(ep.Protocol), ep.entrypoint.Address, ep.elbPort()),
	}
}

// TODO like the comment above, this only applies when there is an EntrypointDomain
func (ep *ExternalPort) AddToELB() error {
	return ep.entrypoint.AddPort(ep.elbPort(), ep.nodePort())
}

func (ep *ExternalPort) RemoveFromELB() error {
	return ep.entrypoint.RemovePort(ep.elbPort())
}
