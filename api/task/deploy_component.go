package task

import (
	"encoding/json"
	"fmt"

	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core"
	"github.com/supergiant/supergiant/deploy"
)

type DeployComponentMessage struct {
	AppName       common.ID
	ComponentName common.ID
}

type DeployComponent struct {
	core *core.Core
}

func (j DeployComponent) Perform(data []byte) error {
	msg := new(DeployComponentMessage)
	if err := json.Unmarshal(data, msg); err != nil {
		return err
	}

	app, err := j.core.Apps().Get(msg.AppName)
	if err != nil {
		return err
	}
	component, err := app.Components().Get(msg.ComponentName)
	if err != nil {
		return err
	}

	var currentRelease *core.ReleaseResource
	if component.CurrentReleaseTimestamp != nil {
		currentRelease, err = component.CurrentRelease()
		if err != nil {
			return err
		}
	}
	// There should always be a target release at this point
	targetRelease, err := component.TargetRelease()
	if err != nil {
		return err
	}

	// This sets up all the necessary dependencies (the only thing needed past the
	// first release is volumes for new instances)
	if err := targetRelease.Provision(); err != nil {
		return err
	}

	if currentRelease != nil {
		targetRelease.AddNewPorts(currentRelease)
	}

	if customDeploy := component.CustomDeployScript; customDeploy != nil {
		if err := core.RunCustomDeployment(j.core, component); err != nil {
			return err
		}
	} else {
		// This goes to the deploy/ folder which uses the client package.
		if err := deploy.Deploy(app.Name, component.Name); err != nil {
			return err
		}
	}

	// Make sure old release (current) has been fully stopped, and the new release
	// (target) has been fully started.
	// It doesn't matter on the first deploy, though.
	if currentRelease != nil {
		if !currentRelease.IsStopped() {
			return fmt.Errorf("Current Release for Component %s:%s is not completely stopped.", *app.Name, *component.Name)
		}
	}
	if !targetRelease.IsStarted() {
		return fmt.Errorf("Target Release for Component %s:%s is not completely started.", *app.Name, *component.Name)
	}

	// TODO really sloppy
	// Stopping instances doesn't remove volumes. So, user-defined deploys, when
	// removing instances, can't control the volumes, which need to be deleted.
	if currentRelease != nil && targetRelease.InstanceCount < currentRelease.InstanceCount {
		instancesRemoving := currentRelease.InstanceCount - targetRelease.InstanceCount
		instances := currentRelease.Instances().List().Items
		for _, instance := range instances[len(instances)-instancesRemoving:] { // TODO test that this works correctly
			instance.DeleteVolumes()
		}
	}

	if currentRelease != nil {
		targetRelease.RemoveOldPorts(currentRelease)

		currentRelease.Retired = true
		currentRelease.Save()
	}

	// If we're all good, we set target to current, and remove target.
	// Also, set the deploy task ID to nil.
	// TODO we should use *string so we can just set to nil
	component.CurrentReleaseTimestamp = component.TargetReleaseTimestamp
	component.TargetReleaseTimestamp = nil
	component.DeployTaskID = nil
	return component.Save()
}
