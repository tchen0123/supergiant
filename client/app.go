package client

import (
	"path"

	"github.com/supergiant/supergiant/common"
)

type App common.App

type AppCollection struct {
	client *Client
}

type AppResource struct {
	collection *AppCollection
	*App
}

type AppList struct {
	Items []*AppResource
}

func (c *AppCollection) path() string {
	return "apps"
}

func (r *AppResource) path() string {
	return path.Join("apps", *r.Name)
}

// Collection-level
//==============================================================================
func (c *AppCollection) New(m *App) *AppResource {
	return &AppResource{c, m}
}

func (c *AppCollection) List() (*AppList, error) {
	list := new(AppList)
	if err := c.client.Get(c.path(), list); err != nil {
		return nil, err
	}
	// see TODO in instance.go
	for _, app := range list.Items {
		app.collection = c
	}
	return list, nil
}

func (c *AppCollection) Create(m *App) (*AppResource, error) {
	r := c.New(m)
	if err := c.client.Post(c.path(), m, r.App); err != nil {
		return nil, err
	}
	return r, nil
}

func (c *AppCollection) Get(name common.ID) (*AppResource, error) {
	m := &App{
		Name: name,
	}
	r := c.New(m)
	if err := c.client.Get(r.path(), r.App); err != nil {
		return nil, err
	}
	return r, nil
}

// Resource-level
//==============================================================================
// func (r *AppResource) Update(m *App) (*AppResource, error) {
//   r.collection.client.
// }

func (r *AppResource) Delete() error {
	return r.collection.client.Delete(r.path())
}

// Relations
func (r *AppResource) Components() *ComponentCollection {
	return &ComponentCollection{
		client: r.collection.client,
		App:    r,
	}
}
