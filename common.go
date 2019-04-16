package common

import "errors"
import "net"

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

type IpRange struct {
	Start net.IP
	End   net.IP
}

type PortRange struct {
	Start uint16
	End   uint16
}

//
// IsAlive param
type IsAliveParam struct {
	IpRange IpRange
}

//
// PortScan param
type PortScanParam struct {
	Type      PortScanType
	Ip        net.IP
	PortRange PortRange
}

//
// generic task param
type TaskParam struct {
	Type TaskType
	Data TaskData
}

//
// validate IsAliveParam
func (param *IsAliveParam) Validate() error {
	return nil
}

//
// validate PortScanParam
func (param *PortScanParam) Validate() error {
	if param.Type != NormalScan || param.Type != SynScan || param.Type != FinScan {
		return errors.New("Invalid port scan type")
	}
	if param.PortRange.Start < 0 || param.PortRange.End < 0 || param.PortRange.Start > param.PortRange.End {
		return errors.New("Invalid port port range")
	}
	return nil
}

//
// validate TaskParam
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
