package common

import "testing"

func TestParam(t *testing.T) {
	param := TaskParam{
		Type: IsAliveTask,
		Data: TaskParam{},
	}

	if err := param.Validate(); err == nil {
		t.Error("Expected invalid param data")
	}

	param.Type = 50
	if err := param.Validate(); err == nil {
		t.Error("Expected invalid param type")
	}
	param.Type = IsAliveTask
	param.Data = IsAliveParam{}
	if err := param.Validate(); err == nil {
		t.Error("Expected invalid param data")
	}
}
