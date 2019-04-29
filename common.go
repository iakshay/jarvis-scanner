package common

import "errors"
import "strings"
import "net"
import "log"
import "time"
import "encoding/json"

type TaskState int
type TaskType int
type WorkerState int
type JobType int
type PortScanType int
type PortStatus int
type IpStatus int
type TaskData interface{}
type JobSubmitData interface{}

const (
	HeartbeatInterval time.Duration = time.Second
)

const (
	Queued     TaskState = 0
	InProgress TaskState = 1
	Complete   TaskState = 2
)

const (
	Undetermined WorkerState = 0
	Unavailable  WorkerState = -1
	Available    WorkerState = 1
)

const (
	IpAlive IpStatus = 0
	IpDead  IpStatus = 1
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
	IsAliveJob  JobType = 0
	PortScanJob JobType = 1
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

func (ipRange *IpRange) Iterate() []net.IP {
	var ips []net.IP

	if ip := ipRange.Start.To4(); ip == nil {
		log.Fatal("expected IPv4")
	}

	start := ipRange.Start.To4()
	end := ipRange.End.To4()
	//log.Println("%d %d %d %d", start[0], start[1], start[2], start[3])
	for p1 := int(start[0]); p1 <= int(end[0]); p1++ {
		for p2 := int(start[1]); p2 <= int(end[1]); p2++ {
			for p3 := int(start[2]); p3 <= int(end[2]); p3++ {
				for p4 := int(start[3]); p4 <= int(end[3]); p4++ {
					ips = append(ips, net.IPv4(byte(p1), byte(p2), byte(p3), byte(p4)))
				}
			}
		}
	}

	return ips
}

type IpResult struct {
	Ip     net.IP
	Status IpStatus
}

type PortRange struct {
	Start uint16 `json:"Start"`
	End   uint16  `json:"End"`
}

type PortResult struct {
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
type IsAliveResult []IpResult

//
// PortScan param
type PortScanParam struct {
	Type      PortScanType
	Ip        net.IP
	PortRange PortRange
}

//
// PortScan result
type PortScanResult map[uint16]PortResult

//
// generic task param
type TaskParam struct {
	Type TaskType
	Data TaskData
}

type JobIsAliveParam struct {
	IpBlock string `json:"IpBlock"`
}

type JobPortScanParam struct {
	Type      PortScanType
	Ip        string
	PortRange PortRange `json:"PortRange"`
}

//
// generic job param
type JobSubmitParam struct {
	Type JobType `json:"Type,omitempty"`
	// should be JobIsAliveParam or JobPortScanParam
	Data json.RawMessage
}

type JobSubmitReply struct{}

// list
type JobListParam struct{}
type JobInfo struct {
	JobId int
	JobName  string // TaskType
	Data  JobSubmitData
}
type JobListReply struct {
	Jobs []JobInfo
}

// detail
type JobDetailParam struct {
	JobId int
}

type WorkerTaskData struct {
	TaskId int
	TaskState string
	WorkerId   int
	WorkerName string
	WorkerAddress string
	Data       TaskData
}

type JobDetailReply struct {
	JobId int
	Data  []WorkerTaskData
}

//
// validate PortScanParam
func (param *JobPortScanParam) Validate() error {
	if ip := net.ParseIP(param.Ip); ip == nil || ip.To4() == nil {
		return errors.New("Invalid IPv4 address")
	}
	if param.Type != NormalScan && param.Type != SynScan && param.Type != FinScan {
		return errors.New("Invalid port scan type")
	}
	if param.PortRange.Start < 0 || param.PortRange.End < 0 || param.PortRange.Start > param.PortRange.End {
		return errors.New("Invalid port port range")
	}
	return nil
}

//
// validate IsAliveParam
func (param *IsAliveParam) Validate() error {
	// validate ipBlock or single IP
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
	TaskId   int
	TaskType TaskType
	TaskData TaskData
}
type SendTaskReply struct{}
