package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type Worker struct {
	taskId int
	params string
	client *rpc.Client
}

func (worker *Worker) DoTask() {
	fmt.Println("Doing Task")

	time.Sleep(1000 * time.Millisecond)
	args := &CompleteTaskArgs{}
	var reply CompleteTaskReply
	worker.client.Call("Server.CompleteTask", args, &reply)
}

func (worker *Worker) SendTask(args *SendTaskArgs, reply *SendTaskReply) error {
	if worker.taskId != -1 {
		log.Fatal("already process task!")
	}

	worker.params = args.Params
	worker.taskId = args.TaskId
	go worker.DoTask()

	return nil
}

func (worker *Worker) RunHearbeat() {
	args := &HeartbeatArgs{WorkerId: 0}
	var reply HeartbeatReply
	for {
		fmt.Println("Sending hearbeat")
		worker.client.Call("Server.Heartbeat", args, &reply)
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	fmt.Println("Starting worker")

	client, err := rpc.DialHTTP("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	// register worker
	args := &RegisterWorkerArgs{Name: "worker", Ip: "localhost", Port: 7070}
	var reply RegisterWorkerReply

	err = client.Call("Server.RegisterWorker", args, &reply)
	if err != nil {
		log.Fatal("RegisterWorker error:", err)
	}

	worker := new(Worker)
	worker.client = client
	// heartbeat go routine
	go worker.RunHearbeat()

	// open upto requests from server
	rpc.Register(worker)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", "localhost:7070")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)
}
