package core

import (
	"log"
	"path"
	"reflect"
	"time"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

type ReleaseCollection struct {
	core      *Core
	Component *ComponentResource
}

type ReleaseResource struct {
	collection *ReleaseCollection
	*common.Release

	// TODO these are shared between releases, it's kinda funky right now
	ExternalService *guber.Service `json:"-"`
	InternalService *guber.Service `json:"-"`

	imageRepos  []*ImageRepoResource
	entrypoints map[string]*EntrypointResource // a map for quick lookup
}

type ReleaseList struct {
	Items []*ReleaseResource `json:"items"`
}

// EtcdKey implements the Collection interface.
func (c *ReleaseCollection) EtcdKey(timestamp common.ID) string {
	key := path.Join("/releases", *c.Component.App().Name, *c.Component.Name)
	if timestamp != nil {
		key = path.Join(key, *timestamp)
	}
	return key
}

// InitializeResource implements the Collection interface.
func (c *ReleaseCollection) InitializeResource(r Resource) error {
	resource := r.(*ReleaseResource)
	resource.collection = c

	// TODO
	// We do this here because this is called when pulling from the DB. If it's
	// being pulled from the DB, it can be assumed to have services.
	// Still really sloppy, since there could be an error.
	svc, err := resource.getService(resource.ExternalServiceName())
	if err != nil {
		return err
	}
	resource.ExternalService = svc

	svc, err = resource.getService(resource.InternalServiceName())
	if err != nil {
		return err
	}
	resource.InternalService = svc

	repos, err := resource.getImageRepos()
	if err != nil {
		return err
	}
	resource.imageRepos = repos

	resource.entrypoints, err = resource.getEntrypoints()
	if err != nil {
		return err
	}

	return nil
}

// List returns an ReleaseList.
func (c *ReleaseCollection) List() (*ReleaseList, error) {
	list := new(ReleaseList)
	err := c.core.DB.List(c, list)
	return list, err
}

// New initializes an Release with a pointer to the Collection.
func (c *ReleaseCollection) New() *ReleaseResource {
	return &ReleaseResource{
		Release: &common.Release{
			Meta: common.NewMeta(),
		},
	}
}

// Create takes an Release and creates it in etcd.
func (c *ReleaseCollection) Create(r *ReleaseResource) (*ReleaseResource, error) {
	r.Timestamp = newReleaseTimestamp()
	if err := c.core.DB.Create(c, r.Timestamp, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Get takes an id and returns an ReleaseResource if it exists.
func (c *ReleaseCollection) Get(id common.ID) (*ReleaseResource, error) {
	r := c.New()
	if err := c.core.DB.Get(c, id, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Resource-level
//==============================================================================

// PersistableObject satisfies the Resource interface
func (r *ReleaseResource) PersistableObject() interface{} {
	return r.Release
}

// Save saves the Release in etcd through an update.
func (r *ReleaseResource) Save() error {
	return r.collection.core.DB.Update(r.collection, r.Timestamp, r)
}

// Delete removes all assets (volumes, pods, etc.) and deletes the Release in
// etcd.
func (r *ReleaseResource) Delete() error {
	if !r.Retired {
		if err := r.deleteServices(); err != nil {
			return err
		}
		if err := r.removeExternalPortsFromEntrypoint(); err != nil {
			return err
		}

		c := make(chan error)
		for _, instance := range r.Instances().List().Items {
			go func(instance *InstanceResource) {
				c <- instance.Delete()
			}(instance)
		}
		for i := 0; i < r.InstanceCount; i++ {
			if err := <-c; err != nil {
				return err
			}
		}
	}
	return r.collection.core.DB.Delete(r.collection, r.Timestamp)
}

func newReleaseTimestamp() common.ID {
	stamp := time.Now().Format("20060102150405")
	return &stamp
}

func (r *ReleaseResource) App() *AppResource {
	return r.Component().App()
}

func (r *ReleaseResource) Component() *ComponentResource {
	return r.collection.Component
}

func (r *ReleaseResource) Instances() *InstanceCollection {
	return &InstanceCollection{
		core:    r.collection.core,
		Release: r,
	}
}

func (r *ReleaseResource) IsStarted() bool {
	for _, instance := range r.Instances().List().Items {
		if !instance.IsStarted() {
			return false
		}
	}
	return true
}

func (r *ReleaseResource) IsStopped() bool {
	for _, instance := range r.Instances().List().Items {
		if !instance.IsStopped() {
			return false
		}
	}
	return true
}

func (r *ReleaseResource) imageRepoNames() (repoNames []string) { // TODO convert Image into Value object w/ repo, image, version
	for _, container := range r.Containers {
		repoNames = append(repoNames, ImageRepoName(container))
	}
	return uniqStrs(repoNames)
}

func (r *ReleaseResource) getEntrypoints() (map[string]*EntrypointResource, error) { // TODO convert Image into Value object w/ repo, image, version
	entrypoints := make(map[string]*EntrypointResource)
	for _, port := range r.containerPorts(true) {
		if port.EntrypointDomain == nil {
			continue
		}
		entrypoint, err := r.collection.core.Entrypoints().Get(port.EntrypointDomain)
		if err != nil {

			// TODO
			if isNotFoundError(err) {
				log.Printf("Entrypoint %s does not exist", *port.EntrypointDomain)
				continue
			}

			return nil, err
		}
		entrypoints[*port.EntrypointDomain] = entrypoint
	}
	return entrypoints, nil
}

func (r *ReleaseResource) containerPorts(public bool) (ports []*common.Port) {
	for _, container := range r.Containers {
		for _, port := range container.Ports {
			if port.Public == public {
				ports = append(ports, port)
			}
		}
	}
	return ports
}

// Operations-------------------------------------------------------------------
func (r *ReleaseResource) getService(name string) (*guber.Service, error) {
	return r.collection.core.K8S.Services(*r.App().Name).Get(name)
}

func (r *ReleaseResource) provisionService(name string, svcType string, svcPorts []*guber.ServicePort) (*guber.Service, error) {
	// doing this here so I don't have to repeat in both external and internal provision methods
	if len(svcPorts) == 0 {
		return nil, nil
	}

	if service, _ := r.getService(name); service != nil {
		return service, nil // already created
	}

	service := &guber.Service{
		Metadata: &guber.Metadata{
			Name: name,
		},
		Spec: &guber.ServiceSpec{
			Type: svcType,
			Selector: map[string]string{
				"service": *r.Component().Name,
			},
			Ports: svcPorts,
		},
	}
	log.Printf("Creating Service %s", name)
	return r.collection.core.K8S.Services(*r.App().Name).Create(service)
}

func (r *ReleaseResource) ExternalServiceName() string {
	return *r.Component().Name + "-public"
}

func (r *ReleaseResource) InternalServiceName() string {
	return *r.Component().Name
}

func (r *ReleaseResource) provisionExternalService() error {
	var ports []*guber.ServicePort
	for _, port := range r.containerPorts(true) {
		ports = append(ports, asKubeServicePort(port))
	}
	svc, err := r.provisionService(r.ExternalServiceName(), "NodePort", ports)
	if err != nil {
		return err
	}
	// TODO repeated in initialization
	r.ExternalService = svc
	return nil
}

func (r *ReleaseResource) provisionInternalService() error {
	var ports []*guber.ServicePort
	for _, port := range r.containerPorts(false) {
		ports = append(ports, asKubeServicePort(port))
	}
	svc, err := r.provisionService(r.InternalServiceName(), "ClusterIP", ports)
	if err != nil {
		return err
	}
	// TODO repeated in initialization
	r.InternalService = svc
	return nil
}

func (r *ReleaseResource) deleteServices() (err error) {
	log.Printf("Deleting Service %s", r.ExternalServiceName())
	if _, err = r.collection.core.K8S.Services(*r.App().Name).Delete(r.ExternalServiceName()); err != nil {
		return err
	}
	log.Printf("Deleting Service %s", r.InternalServiceName())
	if _, err = r.collection.core.K8S.Services(*r.App().Name).Delete(r.InternalServiceName()); err != nil {
		return err
	}
	return nil
}

// NOTE it seems weird here, but "Provision" == "CreateUnlessExists"
func (r *ReleaseResource) provisionSecrets() error {
	for _, repo := range r.imageRepos {
		if err := r.App().ProvisionSecret(repo); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReleaseResource) InternalPorts() (ports []*InternalPort) {
	if r.InternalService == nil {
		return ports
	}
	for _, port := range r.containerPorts(false) {
		ports = append(ports, NewInternalPort(port, r))
	}
	return ports
}

func (r *ReleaseResource) ExternalPorts() (ports []*ExternalPort) {
	if r.ExternalService == nil {
		return ports
	}
	for _, port := range r.containerPorts(true) {
		entrypoint, ok := r.entrypoints[*port.EntrypointDomain]

		if !ok {
			log.Printf("Entrypoint %s does not exist", *port.EntrypointDomain)
			continue
		}

		ports = append(ports, NewExternalPort(port, r, entrypoint))
	}
	return ports
}

func (r *ReleaseResource) addExternalPortsToEntrypoint() error {
	if r.ExternalService == nil {
		return nil
	}

	// NOTE we find from the service so that we don't try to add not-yet serviced
	// ports to the ELB
	ports := r.ExternalPorts()
	for _, svcPort := range r.ExternalService.Spec.Ports {
		for _, port := range ports {
			if port.Number == svcPort.Port {
				if err := port.AddToELB(); err != nil {
					return err
				}
				return nil
			}
		}
	}
	return nil
}

func (r *ReleaseResource) removeExternalPortsFromEntrypoint() error {
	for _, port := range r.ExternalPorts() {
		if err := port.RemoveFromELB(); err != nil {
			return err
		}
	}
	return nil
}

// AddNewPorts adds any new ports defined in containers to the existing
// Services. This is used as a part of the deployment process, and is used in
// conjunction with RemoveOldPorts.
// We use the config returned from the services themselves, as opposed to just
// updating the config, because auto-assigned ports need to be preserved.
func (newR *ReleaseResource) AddNewPorts(oldR *ReleaseResource) error {
	newRInternalPorts := newR.InternalPorts()
	oldRInternalPorts := oldR.InternalPorts()
	var newInternalPorts []*InternalPort
	for _, np := range newRInternalPorts {
		new := true
		for _, op := range oldRInternalPorts {
			if reflect.DeepEqual(*np.Port, *op.Port) {
				new = false
				break
			}
		}
		if new {
			newInternalPorts = append(newInternalPorts, np)
		}
	}

	newRExternalPorts := newR.ExternalPorts()
	oldRExternalPorts := oldR.ExternalPorts()
	var newExternalPorts []*ExternalPort
	for _, np := range newRExternalPorts {
		new := true
		for _, op := range oldRExternalPorts {
			if reflect.DeepEqual(*np.Port, *op.Port) {
				new = false
				break
			}
		}
		if new {
			newExternalPorts = append(newExternalPorts, np)
		}
	}

	if len(newInternalPorts) > 0 {
		svc := newR.InternalService
		log.Printf("Adding new ports to Service %s", svc.Metadata.Name)

		for _, port := range newInternalPorts {
			svc.Spec.Ports = append(svc.Spec.Ports, asKubeServicePort(port.Port))
		}

		svc, err := newR.collection.core.K8S.Services(svc.Metadata.Namespace).Update(svc.Metadata.Name, svc)
		if err != nil {
			return err
		}
		newR.InternalService = svc
	}

	if len(newExternalPorts) > 0 {
		svc := newR.ExternalService
		log.Printf("Adding new ports to Service %s", svc.Metadata.Name)

		for _, port := range newExternalPorts {
			svc.Spec.Ports = append(svc.Spec.Ports, asKubeServicePort(port.Port))
		}

		svc, err := newR.collection.core.K8S.Services(svc.Metadata.Namespace).Update(svc.Metadata.Name, svc)
		if err != nil {
			return err
		}
		newR.ExternalService = svc

		for _, port := range newExternalPorts {
			if err := port.AddToELB(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (newR *ReleaseResource) RemoveOldPorts(oldR *ReleaseResource) error {
	newRInternalPorts := newR.InternalPorts()
	oldRInternalPorts := oldR.InternalPorts()
	var oldInternalPorts []*InternalPort
	for _, op := range oldRInternalPorts {
		old := true
		for _, np := range newRInternalPorts {
			if reflect.DeepEqual(*np.Port, *op.Port) {
				old = false
				break
			}
		}
		if old {
			oldInternalPorts = append(oldInternalPorts, op)
		}
	}
	newRExternalPorts := newR.ExternalPorts()
	oldRExternalPorts := oldR.ExternalPorts()
	var oldExternalPorts []*ExternalPort
	for _, op := range oldRExternalPorts {
		old := true
		for _, np := range newRExternalPorts {
			if reflect.DeepEqual(*np.Port, *op.Port) {
				old = false
				break
			}
		}
		if old {
			oldExternalPorts = append(oldExternalPorts, op)
		}
	}

	if len(oldInternalPorts) > 0 {
		svc := newR.InternalService
		log.Printf("Removing old ports from Service %s", svc.Metadata.Name)

		for _, port := range oldInternalPorts {
			for i, svcPort := range svc.Spec.Ports {
				// remove ports from Service spec
				if svcPort.Port == port.Number {
					svc.Spec.Ports = append(svc.Spec.Ports[:i], svc.Spec.Ports[i+1:]...)
				}
			}
		}
		svc, err := newR.collection.core.K8S.Services(svc.Metadata.Namespace).Update(svc.Metadata.Name, svc)
		if err != nil {
			return err
		}
		newR.InternalService = svc
	}

	if len(oldExternalPorts) > 0 {
		svc := newR.ExternalService
		log.Printf("Removing old ports from Service %s", svc.Metadata.Name)

		for _, port := range oldExternalPorts {
			for i, svcPort := range svc.Spec.Ports {
				// remove ports from Service spec
				if svcPort.Port == port.Number {
					svc.Spec.Ports = append(svc.Spec.Ports[:i], svc.Spec.Ports[i+1:]...)
				}
			}
		}

		svc, err := newR.collection.core.K8S.Services(svc.Metadata.Namespace).Update(svc.Metadata.Name, svc)
		if err != nil {
			return err
		}
		newR.ExternalService = svc

		for _, port := range oldExternalPorts {
			if err := port.RemoveFromELB(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Provision creates needed assets for all instances. It does not actually
// start instances -- that is handled by deploy.go.
func (r *ReleaseResource) Provision() error {
	if err := r.provisionSecrets(); err != nil {
		return err
	}

	// Create Services
	if err := r.provisionInternalService(); err != nil {
		return err
	}
	if err := r.provisionExternalService(); err != nil {
		return err
	}

	if err := r.addExternalPortsToEntrypoint(); err != nil {
		return err
	}

	// Concurrently provision volumes
	// which is not actually concurrent... just sends all requests, and then
	// loops waiting, which prevents concurrently polling while waiting.
	for _, vol := range r.volumes() {
		if ok, err := vol.Exists(); err != nil {
			return err
		} else if !ok {
			if err = vol.Create(); err != nil {
				return err
			}
		}
	}
	for _, vol := range r.volumes() {
		if err := vol.WaitForAvailable(); err != nil {
			return err
		}
	}

	return nil
}

func (r *ReleaseResource) volumes() (vols []*AwsVolume) {
	for _, instance := range r.Instances().List().Items {
		vols = append(vols, instance.Volumes()...)
	}
	return vols
}

func (r *ReleaseResource) getImageRepos() (repos []*ImageRepoResource, err error) { // Not returning ImageRepoResource, since they are defined before hand
	for _, repoName := range r.imageRepoNames() {
		repo, err := r.collection.core.ImageRepos().Get(&repoName)
		if err != nil {

			// TODO this method is ambiguously named
			if isNotFoundError(err) {
				// if there is no repo, we can assume this is a public repo (though it
				// may not be) -- this represents a TODO on how to report errors from
				// Kubernetes
				continue
			}

			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

// TODO naming inconsistencies for kube definitions of resources
// ImagePullSecrets returns repo names defined for Kube pods
func (r *ReleaseResource) ImagePullSecrets() (pullSecrets []*guber.ImagePullSecret, err error) { // TODO don't need to return error here it seems
	for _, repo := range r.imageRepos {
		pullSecrets = append(pullSecrets, asKubeImagePullSecret(repo))
	}
	return pullSecrets, nil
}
