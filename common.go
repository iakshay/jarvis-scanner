package common

import "errors"

type TaskType int
type PortScanType int
type TaskData interface{}

const (
	IsAliveTask  TaskType = 0
	PortScanTask TaskType = 1
)

const (
	NormalScan PortScanType = 0
	SynScan    PortScanType = 1
	FinScan    PortScanType = 2
)

type PortRange struct {
	Start uint16
	End   uint16
}

type IsAliveParam struct {
	IpBlock string
}

type PortScanParam struct {
	SubType PortScanType
	IpBlock string
	Ports   PortRange
}

type TaskParam struct {
	Type TaskType
	Data TaskData
}

func (param *IsAliveParam) Validate() error {
	return nil
}

func (param *PortScanParam) Validate() error {
	return nil
}

func (param *TaskParam) Validate() error {
	switch param.Type {
	case IsAliveTask:
		t, ok := param.Data.(IsAliveParam)
		if !ok {
			return errors.New("Invalid param data")
		}
		return t.Validate()
	case PortScanTask:
		t, ok := param.Data.(PortScanParam)
		if !ok {
			return errors.New("Invalid param data")
		}
		return t.Validate()
	default:
		return errors.New("Invalid task type")
	}

	return nil
}

type HeartbeatArgs struct {
	WorkerId int
}

type HeartbeatReply struct{}

type RegisterWorkerArgs struct {
	Name    string
	Address string
}

type RegisterWorkerReply struct {
	WorkerId int
}

type CompleteTaskArgs struct {
	TaskId int
	Result string
}

type CompleteTaskReply struct{}

type SendTaskArgs struct {
	TaskId int
	Param  TaskParam
}
type SendTaskReply struct{}
