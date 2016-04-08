package core

import (
	"fmt"
	"time"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

func RunCustomDeployment(core *Core, component *ComponentResource) error {
	cd := component.CustomDeployScript
	name := fmt.Sprintf("supergiant-custom-deploy-%s-%s-%s", *component.App().Name, *component.Name, *component.TargetReleaseTimestamp)
	pod := &guber.Pod{
		Metadata: &guber.Metadata{
			Name: name,
		},
		Spec: &guber.PodSpec{
			Containers: []*guber.Container{
				&guber.Container{
					Name:    "container",
					Image:   cd.Image,
					Command: cd.Command,
				},
			},
		},
	}

	pod, err := core.K8S.Pods("default").Create(pod)
	if err != nil {
		return err
	}

	var timeout time.Duration
	if cd.Timeout == 0 {
		timeout = 30 * time.Minute
	} else {
		timeout = time.Duration(cd.Timeout) * time.Second
	}

	var log string

	common.WaitFor(name, timeout, time.Second*5, func() (bool, error) {
		pod, err := core.K8S.Pods("default").Get(name)
		if err != nil {
			return false, err
		} else if pod == nil {
			return true, nil // done
		}

		if latestLog, err := core.K8S.Pods("default").Log(name); err != nil {
			log = latestLog
		}

		return false, nil // pod still exists, keep going
	})

	fmt.Println(log)

	// Now we need to check to see if there were reported errors about the pod
	query := &guber.QueryParams{
		FieldSelector: "involvedObject.kind=Pod,involvedObject.name=" + name,
	}
	events, err := core.K8S.Events("default").Query(query)
	if err != nil {
		return err
	}

	for _, event := range events.Items {
		fmt.Println("EVENT: ", fmt.Sprintf("%#v", event))
	}

	return nil
}
