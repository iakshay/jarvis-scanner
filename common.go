package common

import "errors"
import "strings"
import "net"

type TaskType int
type PortScanType int
type PortStatus int
type IpStatus int
type TaskData interface{}

func (scantype PortScanType) String() string {
	switch scantype {
	case NormalScan:
		return "NormalScan"
	case SynScan:
		return "SynScan"
	case FinScan:
		return "FinScan"
	default:
		panic("invalid scan type")
	}
}

const (
	IpAlive IpStatus = 0
	IsDead  IpStatus = 1
)

const (
	PortOpen       PortStatus = 1 << 0
	PortClosed     PortStatus = 1 << 1
	PortFiltered   PortStatus = 1 << 2
	PortUnfiltered PortStatus = 1 << 3
)

func (portStatus PortStatus) String() string {
	var status []string

	if (portStatus & PortOpen) != 0 {
		status = append(status, "Open")
	}

	if (portStatus & PortClosed) != 0 {
		status = append(status, "Closed")
	}

	if (portStatus & PortFiltered) != 0 {
		status = append(status, "Filtered")
	}

	if (portStatus & PortUnfiltered) != 0 {
		status = append(status, "Unfiltered")
	}

	return strings.Join(status, "|")
}

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

type IpResult struct {
	Ip     net.IP
	Status IpStatus
}

type PortRange struct {
	Start uint16
	End   uint16
}

type PortResult struct {
	Port   uint16
	Status PortStatus
	Banner string
}

//
// IsAlive param
type IsAliveParam struct {
	IpRange IpRange
}

//
// IsAlive result
type IsAliveResult struct {
	Result []IpResult
}

//
// PortScan param
type PortScanParam struct {
	Type      PortScanType
	Ip        net.IP
	PortRange PortRange
}

//
// PortScan result
type PortScanResult struct {
	Result []PortResult
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
	/*if param.Type != NormalScan || param.Type != SynScan || param.Type != FinScan {
		return errors.New("Invalid port scan type")
	}*/
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
	Result TaskData
}

type CompleteTaskReply struct{}

type SendTaskArgs struct {
	TaskId int
	Param  TaskParam
}
type SendTaskReply struct{}
