package main

import (
	"fmt"
	common "github.com/iakshay/jarvis-scanner"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

type Worker struct {
	Id     int
	taskId int
	params string
	client *rpc.Client
}

func (worker *Worker) DoTask() {
	fmt.Println("Doing Task")

	time.Sleep(1000 * time.Millisecond)
	args := &common.CompleteTaskArgs{}
	var reply common.CompleteTaskReply
	worker.client.Call("Server.CompleteTask", args, &reply)
}

func (worker *Worker) SendTask(args *common.SendTaskArgs, reply *common.SendTaskReply) error {
	if worker.taskId != -1 {
		log.Fatal("already process task!")
	}

	worker.params = args.Params
	worker.taskId = args.TaskId
	go worker.DoTask()

	return nil
}

func (worker *Worker) RunHearbeat() {
	args := &common.HeartbeatArgs{WorkerId: worker.Id}
	var reply common.HeartbeatReply
	for {
		fmt.Println("Sending hearbeat")
		worker.client.Call("Server.Heartbeat", args, &reply)
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	fmt.Println("Starting worker")
	var wg sync.WaitGroup
	wg.Add(1)
	client, err := rpc.DialHTTP("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	worker := new(Worker)
	worker.client = client
	worker.taskId = -1

	// open upto requests from server
	rpc.Register(worker)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", "localhost:7070")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)

	// register worker
	args := &common.RegisterWorkerArgs{Name: "worker", Address: "localhost:7070"}
	var reply common.RegisterWorkerReply

	err = client.Call("Server.RegisterWorker", args, &reply)
	if err != nil {
		log.Fatal("RegisterWorker error:", err)
	}
	worker.Id = reply.WorkerId

	// heartbeat go routine
	go worker.RunHearbeat()

	wg.Wait()
}
