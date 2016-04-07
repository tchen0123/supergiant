package client

import (
	"path"

	"github.com/supergiant/supergiant/common"
)

type Entrypoint common.Entrypoint

type EntrypointCollection struct {
	client *Client
}

type EntrypointResource struct {
	collection *EntrypointCollection
	*Entrypoint
}

type EntrypointList struct {
	Items []*EntrypointResource
}

func (c *EntrypointCollection) path() string {
	return path.Join("entrypoints")
}

func (r *EntrypointResource) path() string {
	return path.Join(r.collection.path(), *r.Domain)
}

// Collection-level
//==============================================================================
func (c *EntrypointCollection) New(m *Entrypoint) *EntrypointResource {
	return &EntrypointResource{c, m}
}

func (c *EntrypointCollection) List() (*EntrypointList, error) {
	list := new(EntrypointList)
	if err := c.client.Get(c.path(), list); err != nil {
		return nil, err
	}
	// see TODO in instance.go
	for _, release := range list.Items {
		release.collection = c
	}
	return list, nil
}

func (c *EntrypointCollection) Create(m *Entrypoint) (*EntrypointResource, error) {
	r := c.New(m)
	if err := c.client.Post(c.path(), m, r.Entrypoint); err != nil {
		return nil, err
	}
	return r, nil
}

func (c *EntrypointCollection) Get(domain common.ID) (*EntrypointResource, error) {
	m := &Entrypoint{
		Domain: domain,
	}
	r := c.New(m)
	if err := c.client.Get(r.path(), r.Entrypoint); err != nil {
		return nil, err
	}
	return r, nil
}

// Resource-level
//==============================================================================
func (r *EntrypointResource) Delete() error {
	return r.collection.client.Delete(r.path())
}
