package client

import (
	"path"

	"github.com/supergiant/supergiant/common"
)

type Component common.Component

type ComponentCollection struct {
	client *Client

	App *AppResource
}

type ComponentResource struct {
	collection *ComponentCollection
	*Component
}

type ComponentList struct {
	Items []*ComponentResource
}

func (c *ComponentCollection) path() string {
	return path.Join("apps", *c.App.Name, "components")
}

func (r *ComponentResource) path() string {
	return path.Join(r.collection.path(), *r.Name)
}

// Collection-level
//==============================================================================
func (c *ComponentCollection) New(m *Component) *ComponentResource {
	return &ComponentResource{c, m}
}

func (c *ComponentCollection) List() (*ComponentList, error) {
	list := new(ComponentList)
	if err := c.client.Get(c.path(), list); err != nil {
		return nil, err
	}
	// see TODO in instance.go
	for _, component := range list.Items {
		component.collection = c
	}
	return list, nil
}

func (c *ComponentCollection) Create(m *Component) (*ComponentResource, error) {
	r := c.New(m)
	if err := c.client.Post(c.path(), m, r.Component); err != nil {
		return nil, err
	}
	return r, nil
}

func (c *ComponentCollection) Get(name common.ID) (*ComponentResource, error) {
	m := &Component{
		PersistableComponent: &common.PersistableComponent{ // TODO any way to not make this so weird?
			Name: name,
		},
	}
	r := c.New(m)
	if err := c.client.Get(r.path(), r.Component); err != nil {
		return nil, err
	}
	return r, nil
}

// Resource-level
//==============================================================================
// func (r *ComponentResource) Update(m *Component) (*ComponentResource, error) {
//   r.collection.client.
// }

func (r *ComponentResource) Delete() error {
	return r.collection.client.Delete(r.path())
}

// Relations
func (r *ComponentResource) Releases() *ReleaseCollection {
	return &ReleaseCollection{
		client:    r.collection.client,
		App:       r.collection.App,
		Component: r,
	}
}

func (r *ComponentResource) CurrentRelease() (*ReleaseResource, error) {
	return r.Releases().Get(r.CurrentReleaseTimestamp)
}

func (r *ComponentResource) TargetRelease() (*ReleaseResource, error) {
	return r.Releases().Get(r.TargetReleaseTimestamp)
}
