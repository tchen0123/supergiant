package types

type App struct {
	Name string `json:"name"`
}

type Component struct {
	Name string `json:"name"`
	// TODO kinda weird,
	// you choose a container that has the deploy file, and then reference it as a command
	CustomDeployScript *CustomDeployScript `json:"custom_deploy_script"`

	// TODO these all seem to be a departure in terms of relations... I feel like
	// there is a more elegant solution to house the info, and delete it atomically
	CurrentReleaseID string `json:"current_release_id"`
	TargetReleaseID  string `json:"target_release_id"`
	// We should just store the DeployTaskID but actually should render the task
	// when showing it in HTTP.
	DeployTaskID string `json:"deploy_task_id"`
}

// TODO implement...
type CustomDeployScript struct {
	Image   string `json:"image"`
	Command string `json:"command"`
}

// Volume
//==============================================================================
type VolumeBlueprint struct {
	Name string `json:"name"`
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
	Volume string `json:"volume"`
	Path   string `json:"path"`
}

// Port
//==============================================================================
type Port struct {
	Protocol string `json:"protocol"`
	Number   int    `json:"number"`
	Public   bool   `json:"public"`
}

// ResourceAllocation
//==============================================================================
type ResourceAllocation struct {
	Min uint `json:"min"`
	Max uint `json:"max"`
}

// NOTE the word Blueprint is used for Volumes and Containers, since they are
// both "definitions" that create "instances" of the real thing

type Blueprint struct {
	Volumes                []*VolumeBlueprint    `json:"volumes"`
	Containers             []*ContainerBlueprint `json:"containers"`
	TerminationGracePeriod int                   `json:"termination_grace_period"`
}

// Release
//==============================================================================
type Release struct {
	ID            string     `json:"id"`
	InstanceCount int        `json:"instance_count"`
	Blueprint     *Blueprint `json:"blueprint"`
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
	ID int `json:"id"` // actually just the number (starting w/ 1) of the instance order in the release

	// BaseName is the name of the instance without the Release ID appended. It is
	// used for naming volumes, which move between releases.
	BaseName string `json:"base_name"`
	Name     string `json:"name"`

	Status InstanceStatus `json:"status"`
}

// Entrypoint
//==============================================================================
type Entrypoint struct {
	Domain  string `json:"domain"`  // e.g. blog.qbox.io
	Address string `json:"address"` // the ELB address

	// NOTE we actually don't need this -- we can always attach the policy, and enable per port
	// IPWhitelistEnabled bool   `json:"ip_whitelist_enabled"`
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
}

// ImageRepo
//==============================================================================
type ImageRepo struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}