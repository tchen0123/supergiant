package core

import (
	"strconv"
	"strings"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

// kube_helpers.go is a collection of helper methods that convert a Supergiant
// resource definition into a Kubernetes resource defition.
// (and some other assorted things that should maybe be moved out...)

func kubeVolumeMounts(m *common.ContainerBlueprint) (volMounts []*guber.VolumeMount) {
	for _, mount := range m.Mounts {
		volMounts = append(volMounts, asKubeVolumeMount(mount))
	}
	return volMounts
}

func kubeContainerPorts(m *common.ContainerBlueprint) (cPorts []*guber.ContainerPort) {
	for _, port := range m.Ports {
		cPorts = append(cPorts, asKubeContainerPort(port))
	}
	return cPorts
}

func interpolatedEnvVars(m *common.ContainerBlueprint, instance *InstanceResource) (envVars []*guber.EnvVar) {
	for _, envVar := range m.Env {
		envVars = append(envVars, asKubeEnvVar(envVar, instance))
	}
	return envVars
}

func ImageRepoName(m *common.ContainerBlueprint) string {
	return strings.Split(m.Image, "/")[0]
}

func asKubeContainer(m *types.ContainerBlueprint, instance *InstanceResource) *guber.Container { // NOTE how instance must be passed here
	// TODO
	resources := new(guber.Resources)
	if m.RAM != nil {
		if m.RAM.Min != 0 {
			resources.Requests.Memory = types.BytesFromMiB(m.RAM.Min).ToKubeMebibytes()
		}
		if m.RAM.Max != 0 {
			resources.Limits.Memory = types.BytesFromMiB(m.RAM.Max).ToKubeMebibytes()
		}
	}
	if m.CPU != nil {
		if m.CPU.Min != 0 {
			resources.Requests.CPU = types.CoresFromMillicores(m.CPU.Min).ToKubeMillicores()
		}
		if m.CPU.Max != 0 {
			resources.Limits.CPU = types.CoresFromMillicores(m.CPU.Max).ToKubeMillicores()
		}
	}

	return &guber.Container{
		Name:         "container", // TODO this will fail with multiple containers ------------------------------------ TODO
		Image:        m.Image,
		Env:          interpolatedEnvVars(m, instance),
		Resources:    resources,
		VolumeMounts: kubeVolumeMounts(m),
		Ports:        kubeContainerPorts(m),

		// TODO this should be an option, enabled by default with volumes
		SecurityContext: &guber.SecurityContext{
			Privileged: true,
		},
	}
}

// EnvVar
//==============================================================================
func interpolatedValue(m *common.EnvVar, instance *InstanceResource) string {
	r := strings.NewReplacer(
		"{{ instance_id }}", *instance.ID,
		"{{ other_stuff }}", "TODO")
	return r.Replace(m.Value)
}

func asKubeEnvVar(m *common.EnvVar, instance *InstanceResource) *guber.EnvVar {
	return &guber.EnvVar{
		Name:  m.Name,
		Value: interpolatedValue(m, instance),
	}
}

// Volume
//==============================================================================
func asKubeVolume(m *AwsVolume) (*guber.Volume, error) {
	vol, err := m.awsVolume()
	if err != nil {
		return nil, err
	}

	return &guber.Volume{
		Name: *m.Blueprint.Name, // NOTE this is not the physical volume name
		AwsElasticBlockStore: &guber.AwsElasticBlockStore{
			VolumeID: *vol.VolumeId,
			FSType:   "ext4",
		},
	}, nil
}

// Mount
//==============================================================================
func asKubeVolumeMount(m *common.Mount) *guber.VolumeMount {
	return &guber.VolumeMount{
		Name:      *m.Volume,
		MountPath: m.Path,
	}
}

// Port
//==============================================================================
func portName(m *common.Port) string {
	return strconv.Itoa(m.Number)
}

func asKubeContainerPort(m *common.Port) *guber.ContainerPort {
	return &guber.ContainerPort{
		ContainerPort: m.Number,
	}
}

func asKubeServicePort(m *common.Port) *guber.ServicePort {
	return &guber.ServicePort{
		Name:     portName(m),
		Port:     m.Number,
		Protocol: "TCP", // this is default; only other option is UDP
	}
}

// ImageRepo
//==============================================================================
func asKubeImagePullSecret(m *ImageRepoResource) *guber.ImagePullSecret {
	return &guber.ImagePullSecret{
		Name: *m.Name,
	}
}

func asKubeSecret(m *ImageRepoResource) *guber.Secret {
	return &guber.Secret{
		Metadata: &guber.Metadata{
			Name: *m.Name,
		},
		Type: "kubernetes.io/dockercfg",
		Data: map[string]string{
			".dockercfg": m.Key,
		},
	}
}
