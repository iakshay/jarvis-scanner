package common

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
	Params string
}
type SendTaskReply struct{}
