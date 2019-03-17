package main

type HeartbeatArgs struct {
  WorkerId int
}

type HeartbeatReply struct{}

type RegisterWorkerArgs struct {
	Name string
	Ip   string
	Port int
}

type RegisterWorkerReply struct{}

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
