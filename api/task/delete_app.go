package task

import (
	"encoding/json"

	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/core"
)

type DeleteAppMessage struct {
	AppName common.ID
}

// DeleteApp implements task.Performable interface
type DeleteApp struct {
	core *core.Core
}

func (j DeleteApp) Perform(data []byte) error {
	message := new(DeleteAppMessage)
	if err := json.Unmarshal(data, message); err != nil {
		return err
	}

	app, err := j.core.Apps().Get(message.AppName)
	if err != nil {
		return err
	}

	return app.Delete()
}
