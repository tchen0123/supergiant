package core

import (
	"path"

	"github.com/supergiant/supergiant/common"
)

type ComponentCollection struct {
	core *Core
	App  *AppResource
}

type ComponentResource struct {
	collection *ComponentCollection
	*common.Component
}

type ComponentList struct {
	Items []*ComponentResource `json:"items"`
}

// EtcdKey implements the Collection interface.
func (c *ComponentCollection) EtcdKey(name common.ID) string {
	key := path.Join("/components", *c.App.Name)
	if name != nil {
		key = path.Join(key, *name)
	}
	return key
}

// InitializeResource implements the Collection interface.
func (c *ComponentCollection) InitializeResource(r Resource) error {
	resource := r.(*ComponentResource)
	resource.collection = c

	// TODO it seems wrong this is called here -- execessive to have to load the
	// current Release, Entrypoints, and Kube Services just to render a
	// Component.
	// However, it's rare a Component is loaded out of the context of its
	// Release. We will change this when we see issues.
	return resource.decorate()
}

// List returns an ComponentList.
func (c *ComponentCollection) List() (*ComponentList, error) {
	list := new(ComponentList)
	err := c.core.DB.List(c, list)
	return list, err
}

// New initializes an Component with a pointer to the Collection.
func (c *ComponentCollection) New() *ComponentResource {
	// Yes, this looks insane.
	return &ComponentResource{
		Component: &common.Component{
			PersistableComponent: &common.PersistableComponent{
				Meta: common.NewMeta(),
			},
		},
	}
}

// Create takes an Component and creates it in etcd.
func (c *ComponentCollection) Create(r *ComponentResource) (*ComponentResource, error) {
	if err := c.core.DB.Create(c, r.Name, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Get takes a name and returns an ComponentResource if it exists.
func (c *ComponentCollection) Get(name common.ID) (*ComponentResource, error) {
	r := c.New()
	if err := c.core.DB.Get(c, name, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Resource-level
//==============================================================================

func (r *ComponentResource) decorate() error {
	if r.CurrentReleaseTimestamp == nil {
		return nil
	}

	externalAddrs, err := r.externalAddresses()
	if err != nil {
		return err
	}
	internalAddrs, err := r.internalAddresses()
	if err != nil {
		return err
	}

	r.Addresses = &common.ComponentAddresses{
		External: externalAddrs,
		Internal: internalAddrs,
	}

	return nil
}

// PersistableObject satisfies the Resource interface
func (r *ComponentResource) PersistableObject() interface{} {
	return r.PersistableComponent
}

// Save saves the Component in etcd through an update.
func (r *ComponentResource) Save() error {
	return r.collection.core.DB.Update(r.collection, r.Name, r)
}

// Delete cascades delete calls to current and target releases, and deletes the
// Component in etcd.
//
// TODO this should somehow stop any ongoing tasks related to the Component.
func (r *ComponentResource) Delete() error {
	releases, err := r.Releases().List()
	if err != nil {
		return err
	}

	// NOTE we delete releases concurrently, because when target and current both
	// exist (i.e. a deploy was still running), then one may hang waiting on the
	// other to release an asset like volumes.

	c := make(chan error)
	for _, release := range releases.Items {
		go func(release *ReleaseResource) {
			c <- release.Delete()
		}(release)
	}
	for i := 0; i < len(releases.Items); i++ {
		if err := <-c; err != nil {
			return err
		}
	}

	return r.collection.core.DB.Delete(r.collection, r.Name)
}

func (r *ComponentResource) App() *AppResource {
	return r.collection.App
}

func (r *ComponentResource) Releases() *ReleaseCollection {
	return &ReleaseCollection{
		core:      r.collection.core,
		Component: r,
	}
}

func (r *ComponentResource) CurrentRelease() (*ReleaseResource, error) {
	return r.Releases().Get(r.CurrentReleaseTimestamp)
}

func (r *ComponentResource) TargetRelease() (*ReleaseResource, error) {
	return r.Releases().Get(r.TargetReleaseTimestamp)
}

func (r *ComponentResource) externalAddresses() (addrs []*common.PortAddress, err error) {
	release, err := r.CurrentRelease()
	if err != nil {
		return nil, err
	}
	for _, port := range release.ExternalPorts() {
		addrs = append(addrs, port.Address())
	}
	return addrs, nil
}

func (r *ComponentResource) internalAddresses() (addrs []*common.PortAddress, err error) {
	release, err := r.CurrentRelease()
	if err != nil {
		return nil, err
	}
	for _, port := range release.InternalPorts() {
		addrs = append(addrs, port.Address())
	}
	return addrs, nil
}
