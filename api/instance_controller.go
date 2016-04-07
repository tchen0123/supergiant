package api

import (
	"net/http"

	"github.com/supergiant/supergiant/api/task"
	"github.com/supergiant/supergiant/core"
	"github.com/supergiant/supergiant/common"
)

type InstanceController struct {
	core *core.Core
}

func (c *InstanceController) Index(w http.ResponseWriter, r *http.Request) {
	release, err := loadRelease(c.core, w, r)
	if err != nil {
		return
	}

	instances := release.Instances().List()

	body, err := marshalBody(w, instances)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *InstanceController) Show(w http.ResponseWriter, r *http.Request) {
	instance, err := loadInstance(c.core, w, r)
	if err != nil {
		return
	}

	body, err := marshalBody(w, instance)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *InstanceController) Start(w http.ResponseWriter, r *http.Request) {
	instance, err := loadInstance(c.core, w, r)
	if err != nil {
		return
	}

	msg := &task.StartInstanceMessage{
		AppName:          instance.App().Name,
		ComponentName:    instance.Component().Name,
		ReleaseTimestamp: instance.Release().Timestamp,
		ID:               instance.ID,
	}
	_, err = c.core.Tasks().Start(common.TaskTypeStartInstance, msg)
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, instance)
	if err != nil {
		return
	}
	renderWithStatusAccepted(w, body)
}

func (c *InstanceController) Stop(w http.ResponseWriter, r *http.Request) {
	instance, err := loadInstance(c.core, w, r)
	if err != nil {
		return
	}

	msg := &task.StopInstanceMessage{
		AppName:          instance.App().Name,
		ComponentName:    instance.Component().Name,
		ReleaseTimestamp: instance.Release().Timestamp,
		ID:               instance.ID,
	}
	_, err = c.core.Tasks().Start(common.TaskTypeStopInstance, msg)
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, instance)
	if err != nil {
		return
	}
	renderWithStatusAccepted(w, body)
}
