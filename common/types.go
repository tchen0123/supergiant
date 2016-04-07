package common

// Tags is a meta field for holding unstructued key/val info.
type Tags map[string]string

// Meta is a set of fields on all Resources for holding metadata.
type Meta struct {
	Created *Timestamp `json:"created"`
	Updated *Timestamp `json:"updated"`
	Tags    Tags       `json:"tags"`
}

// TODO weird placement
func NewMeta() *Meta {
	return &Meta{
		Tags: make(Tags),
	}
}

type App struct {
	Name ID `json:"name"`
	*Meta
}

// see note in core/resource.go about PersistableObject()
type PersistableComponent struct {
	Name ID `json:"name"`
	// TODO kinda weird,
	// you choose a container that has the deploy file, and then reference it as a command
	CustomDeployScript *CustomDeployScript `json:"custom_deploy_script"`

	// TODO these all seem to be a departure in terms of relations... I feel like
	// there is a more elegant solution to house the info, and delete it atomically
	CurrentReleaseTimestamp ID `json:"current_release_id"`
	TargetReleaseTimestamp  ID `json:"target_release_id"`
	// We should just store the DeployTaskID but actually should render the task
	// when showing it in HTTP.
	DeployTaskID ID `json:"deploy_task_id"`

	*Meta
}

type Component struct {
	*PersistableComponent
	Addresses *ComponentAddresses `json:"addresses"`
}

type PortAddress struct {
	Port    string `json:"port"` // TODO really this should be the name of the port, which currently is the string of the number
	Address string `json:"address"`
}

type ComponentAddresses struct {
	External []*PortAddress `json:"external"`
	Internal []*PortAddress `json:"internal"`
}

type CustomDeployScript struct {
	Image   string   `json:"image"`
	Command []string `json:"command"`
	Timeout int      `json:"timeout"`
}

// Volume
//==============================================================================
type VolumeBlueprint struct {
	Name ID     `json:"name"`
	Type string `json:"type"`
	Size int    `json:"size"`
}

// Container
//==============================================================================
type ContainerBlueprint struct {
	Image  string              `json:"image"`
	Ports  []*Port             `json:"ports"`
	Env    []*EnvVar           `json:"env"`
	CPU    *ResourceAllocation `json:"cpu"`
	RAM    *ResourceAllocation `json:"ram"`
	Mounts []*Mount            `json:"mounts"`
}

// EnvVar
//==============================================================================
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"` // this may be templated, "something_{{ instance_id }}"
}

// Mount
//==============================================================================
type Mount struct {
	Volume ID     `json:"volume"` // TODO should be VolumeName
	Path   string `json:"path"`
}

// Port
//==============================================================================
type Port struct {
	Protocol string `json:"protocol"`
	Number   int    `json:"number"` // should be uint
	Public   bool   `json:"public"`

	// EntrypointDomain specifies which Entrypoint this Port is added to. Does not
	// apply when Public is false.
	EntrypointDomain ID `json:"entrypoint_domain,omitempty"`

	// PreserveNumber instructs the Entrypoint to set the actual Port number
	// specified as the external load balancer port. Does not apply when
	// EntrypointDomain is nil.
	//
	// NOTE we may want to replace this with something like ExternalNumber so that
	// it is more flexible / obvious.
	PreserveNumber bool `json:"preserve_number"`
}

// ResourceAllocation
//==============================================================================
type ResourceAllocation struct {
	Min uint `json:"min"`
	Max uint `json:"max"`
}

// Release
//==============================================================================
// NOTE the word Blueprint is used for Volumes and Containers, since they are
// both "definitions" that create "instances" of the real thing
type Release struct {
	// NOTE Timestamp here does not use the Timestamp type.
	Timestamp ID `json:"timestamp"`

	InstanceCount int `json:"instance_count"`

	// These attributes, when changed from last Release, indicate a restart is
	// needed (or just new instances through other means).
	Volumes                []*VolumeBlueprint    `json:"volumes"`
	Containers             []*ContainerBlueprint `json:"containers"`
	TerminationGracePeriod int                   `json:"termination_grace_period"`

	// Retired defines whether or not a Release still has active assets, like pods
	// or services. When retired is true, we skip attempting to delete assets.
	Retired bool `json:"retired"`

	*Meta
}

// Instance
//==============================================================================
type InstanceStatus string

const (
	InstanceStatusStopped InstanceStatus = "STOPPED"
	InstanceStatusStarted InstanceStatus = "STARTED"
)

// NOTE Instances are not stored in etcd, so the json tags here apply to HTTP
type Instance struct {
	ID ID `json:"id"` // actually just the number (starting w/ 1) of the instance order in the release

	// BaseName is the name of the instance without the Release ID appended. It is
	// used for naming volumes, which move between releases.
	BaseName string `json:"base_name"`
	Name     string `json:"name"`

	Status InstanceStatus `json:"status"`
}

// Entrypoint
//==============================================================================
type Entrypoint struct {
	Domain  ID     `json:"domain"`  // e.g. blog.qbox.io
	Address string `json:"address"` // the ELB address

	// NOTE we actually don't need this -- we can always attach the policy, and enable per port
	// IPWhitelistEnabled bool   `json:"ip_whitelist_enabled"`

	*Meta
}

// Task
//==============================================================================
type TaskType int

const (
	TaskTypeDeployComponent TaskType = iota
	TaskTypeDeleteComponent
	TaskTypeDeleteApp
	TaskTypeDeleteRelease
	TaskTypeStartInstance
	TaskTypeStopInstance
)

type Task struct {
	Type        TaskType `json:"type"`
	Data        []byte   `json:"data"`
	Status      string   `json:"status"`
	Attempts    int      `json:"attempts"`
	MaxAttempts int      `json:"max_attempts"` // this is static; config-level
	Error       string   `json:"error"`

	*Meta
}

// ImageRepo
//==============================================================================
type ImageRepo struct {
	Name ID     `json:"name"`
	Key  string `json:"key"`

	*Meta
}
